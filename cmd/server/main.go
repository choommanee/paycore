package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/yourco/payment-gateway/internal/config"
	"github.com/yourco/payment-gateway/internal/db"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/handler"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/crypto"
	"github.com/yourco/payment-gateway/internal/pkg/metrics"
	"github.com/yourco/payment-gateway/internal/pkg/promptpay"
	"github.com/yourco/payment-gateway/internal/repository"
	"github.com/yourco/payment-gateway/internal/router"
	"github.com/yourco/payment-gateway/internal/service"
	"github.com/yourco/payment-gateway/internal/worker"
	"github.com/yourco/payment-gateway/migrations"
)

// @title       Payment Gateway API
// @version     0.1.0
// @description Full acquiring payment gateway (authorize / capture / refund / 3DS / webhook)
// @BasePath    /v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	// Structured JSON logging (never log PAN / CVV / full track data)
	zerolog.TimeFieldFormat = time.RFC3339
	logger := log.Output(os.Stdout).Level(cfg.LogLevel())

	ctx := context.Background()

	// Auto-migrate on boot (gated by MIGRATE_ON_BOOT). The distroless runtime
	// image has no shell or migrate binary, so migrations run in-process from the
	// embedded FS before serving. A migration failure is fatal — a half-migrated
	// schema must never start taking money.
	if cfg.MigrateOnBoot {
		logger.Info().Msg("MIGRATE_ON_BOOT=true; applying database migrations")
		if err := db.RunMigrations(ctx, cfg.DatabaseURL, migrations.FS, logger); err != nil {
			log.Fatal().Err(err).Msg("run migrations")
		}
	}

	// Open the Postgres connection pool. In production a missing/unreachable DB
	// is fatal; in dev we log and continue so the process still boots for local
	// scaffolding (readiness will report 503 until the DB is up).
	poolCfg := db.DefaultPoolConfig(cfg.DatabaseURL).WithSizing(cfg.DBMaxConns, cfg.DBMinConns)
	pool, err := db.NewPool(ctx, poolCfg)
	if err != nil {
		if cfg.IsProd() {
			log.Fatal().Err(err).Msg("connect database")
		}
		logger.Warn().Err(err).Msg("database unavailable; continuing (dev only)")
	}
	if pool != nil {
		defer pool.Close()
	}

	// sqlc-generated data access. repository.New accepts any DBTX; a *pgxpool.Pool
	// satisfies it. Nil when the pool failed to open in dev.
	var repo repository.Querier
	if pool != nil {
		repo = repository.New(pool)
	}

	// Wire dependencies (repository -> service -> handler).
	// The pool doubles as the transaction manager (it can Begin a pgx.Tx) so
	// money-moving operations are written atomically.
	var txm service.TxManager
	if pool != nil {
		txm = pool
	}

	// External rails: mock implementations for sandbox/dev. Swapping in the real
	// acquirer / 3DS / QR provider is just another impl behind these interfaces.
	acquirer := external.NewMockAcquirer()
	threeDS := external.NewMockThreeDS("")
	qrProvider := external.NewMockQRProvider(cfg.QRWebhookSecret, cfg.QRProviderBaseURL)

	// Card-data vault. Backed by a local dev KMS here; production wires an
	// HSM/KMS-backed KMS. The vault detokenizes single-use tokens to a PAN inside
	// PCI scope and never logs or persists card data.
	kms, err := crypto.NewDevLocalKMS(cfg.KMSKeyID)
	if err != nil {
		log.Fatal().Err(err).Msg("init card vault kms")
	}
	vault := crypto.NewVault(kms)

	// Prometheus instruments (payments-by-status counter, handler latency
	// histogram) served at GET /metrics.
	metricsSet := metrics.New()

	// Async audit outbox: the authorize hot path enqueues audit rows here so they
	// are persisted off the synchronous payment transaction (write-amplification
	// reduction). Only started when a DB pool is available.
	var auditWorker *worker.AuditWorker
	auditCtx, stopAudit := context.WithCancel(context.Background())
	if repo != nil {
		auditWorker = worker.NewAuditWorker(repo, 4096, logger)
		auditWorker.Start(auditCtx)
	}

	svc := service.NewPaymentService(repo, txm, vault, acquirer, threeDS, logger)
	// Attach the payment status counter when the concrete service supports it.
	if mw, ok := svc.(interface {
		WithMetrics(service.StatusObserver) service.PaymentService
	}); ok {
		svc = mw.WithMetrics(metricsSet)
	}
	// Attach the CRes verification secret for the browser 3DS return, reusing the
	// webhook signing secret as the shared HMAC key.
	if mw, ok := svc.(interface {
		WithThreeDSSecret(string) service.PaymentService
	}); ok {
		svc = mw.WithThreeDSSecret(cfg.WebhookSigningSecret)
	}
	// Attach the async audit sink when the outbox worker is running.
	if auditWorker != nil {
		if mw, ok := svc.(interface {
			WithAuditSink(service.AuditSink) service.PaymentService
		}); ok {
			svc = mw.WithAuditSink(auditWorker)
		}
	}

	// QR payments. promptpay.Target resolves per-merchant in production; here it
	// is constructed from config so locally-built PromptPay QR has a target.
	qrTarget := promptpay.Target{
		MobileNo:   cfg.PromptPayMobileNo,
		NationalID: cfg.PromptPayNationalID,
		EWalletID:  cfg.PromptPayEWalletID,
	}
	qrSvc := service.NewQRService(repo, txm, qrProvider, qrTarget, logger)

	// Merchant onboarding/lookup + API-key resolution, chargeback/disputes, and
	// the platform (admin) read models.
	merchantSvc := service.NewMerchantService(repo, logger)
	disputeSvc := service.NewDisputeService(repo, logger)
	adminSvc := service.NewAdminService(repo, logger)

	// Health handler pings the pool for readiness. Passing the typed pool (which
	// may be nil) directly would make the Pinger interface non-nil with a nil
	// value; guard so readiness correctly reports "unconfigured" when there is
	// no pool.
	var pinger handler.Pinger
	if pool != nil {
		pinger = pool
	}
	h := handler.New(svc, qrSvc, merchantSvc, disputeSvc, adminSvc, pinger, logger)

	// Sandbox payer simulator ("Sandbox Bank"). Only wired when SANDBOX_MODE=true
	// so a real deploy never exposes an unauthenticated route that marks payments
	// paid. The service signs the bank confirmation with QR_WEBHOOK_SECRET
	// server-side and replays it through the normal ConfirmFromWebhook path.
	if cfg.SandboxMode && repo != nil {
		sandboxSvc := service.NewSandboxService(repo, qrSvc, cfg.QRWebhookSecret, logger)
		h = h.WithSandbox(sandboxSvc, logger)
		logger.Warn().Msg("SANDBOX_MODE=true: public /v1/sandbox payer-simulator endpoints are ENABLED (must be off in production)")
	}

	// Behind a trusted TLS-terminating upstream (Railway/Fly/LB) the socket peer
	// is the platform's proxy — and its source IP rotates, which breaks per-IP
	// rate limiting (each request looks like a new client) and pollutes audit
	// logs. When the operator has acknowledged such an upstream
	// (TLS_TERMINATED_UPSTREAM), read the real caller from X-Forwarded-For so
	// c.IP() reflects the client. We do NOT enable this for direct/local serving,
	// where honoring a client-supplied XFF would let it spoof its own IP.
	proxyHeader := ""
	if cfg.TLSTerminatedUpstream {
		proxyHeader = fiber.HeaderXForwardedFor
	}

	app := fiber.New(fiber.Config{
		AppName:               "payment-gateway",
		DisableStartupMessage: true,
		ReadTimeout:           15 * time.Second,
		WriteTimeout:          15 * time.Second,
		BodyLimit:             1 * 1024 * 1024, // 1MB
		ProxyHeader:           proxyHeader,
		ErrorHandler:          middleware.ErrorHandler(logger),
	})

	middleware.Register(app, cfg, logger)
	// Record handler latency into the Prometheus histogram.
	app.Use(middleware.Metrics(metricsSet))
	// Merchant API-key auth applied to the payment / QR route groups, plus a
	// per-merchant rate limiter mounted after auth on those groups only.
	auth := middleware.APIKeyAuth(merchantSvc)
	adminAuth := middleware.AdminAuth(cfg.AdminAPIKey)
	rateLimit := middleware.RateLimiter(cfg.RateLimitPerSec)
	// Dedicated IP-keyed limiter for the public self-service signup endpoint.
	signupLimit := middleware.SignupRateLimiter(cfg.SignupRateLimitPerHour)

	// /metrics leaks business/volume telemetry, so it is NOT exposed on the public
	// money API by default. It is bound on a separate internal listener when
	// METRICS_ADDR is set; only mounted on the public listener when METRICS_PUBLIC
	// is explicitly enabled (dev convenience).
	var publicMetrics fiber.Handler
	if cfg.MetricsPublic {
		publicMetrics = middleware.MetricsHandler(metricsSet)
	}
	router.Setup(app, h, auth, adminAuth, rateLimit, signupLimit, publicMetrics, cfg.WebDir, cfg.SandboxMode)

	// Separate internal metrics listener (not the public money API).
	var metricsApp *fiber.App
	if cfg.MetricsAddr != "" {
		metricsApp = fiber.New(fiber.Config{DisableStartupMessage: true})
		metricsApp.Get("/metrics", middleware.MetricsHandler(metricsSet))
		go func() {
			if lerr := metricsApp.Listen(cfg.MetricsAddr); lerr != nil && !errors.Is(lerr, fiber.ErrServiceUnavailable) {
				logger.Error().Err(lerr).Msg("metrics listener")
			}
		}()
		logger.Info().Str("metrics_addr", cfg.MetricsAddr).Msg("internal metrics listener started")
	}

	// Background ops workers. They share a context cancelled on shutdown so the
	// loops exit cleanly. Workers are best-effort and no-op when the DB pool is
	// unavailable (repo == nil in dev scaffolding).
	workerCtx, stopWorkers := context.WithCancel(context.Background())
	if repo != nil {
		reporter := external.NewMockSettlementReporter() // swap for a real acquirer report feed
		worker.StartAll(workerCtx, logger,
			worker.NewSettlementWorker(repo, time.Hour, 24*time.Hour, cfg.SettlementFeeBps, logger),
			worker.NewReconciliationWorker(repo, reporter, time.Hour, 24*time.Hour, logger),
			worker.NewWebhookWorker(repo, cfg.WebhookSigningSecret, cfg.WebhookDefaultURL, 15*time.Second, int32(cfg.WebhookMaxAttempts), logger), // #nosec G115 -- WebhookMaxAttempts is clamped to [1,100] in config.validate
		)
	}

	go func() {
		// Serve over TLS when a cert/key pair is configured; otherwise plain HTTP
		// (TLS terminated upstream at a load balancer / mesh).
		var lerr error
		if cfg.TLSEnabled() {
			lerr = app.ListenTLS(cfg.HTTPAddr, cfg.TLSCertFile, cfg.TLSKeyFile)
		} else {
			lerr = app.Listen(cfg.HTTPAddr)
		}
		if lerr != nil && !errors.Is(lerr, fiber.ErrServiceUnavailable) {
			log.Fatal().Err(lerr).Msg("server listen")
		}
	}()
	// Warn loudly when serving plaintext in prod behind an acknowledged upstream:
	// config.validate already refused to start with neither TLS nor the ack flag.
	if cfg.IsProd() && !cfg.TLSEnabled() {
		logger.Warn().Msg("serving plaintext HTTP in production (TLS_TERMINATED_UPSTREAM=true): TLS MUST be terminated by a trusted upstream")
	}
	logger.Info().
		Str("addr", cfg.HTTPAddr).
		Bool("tls", cfg.TLSEnabled()).
		Msg("payment-gateway started")

	// Graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("shutting down...")
	stopWorkers() // signal background workers to exit their loops
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
	}
	if metricsApp != nil {
		if err := metricsApp.ShutdownWithContext(ctx); err != nil {
			logger.Error().Err(err).Msg("metrics listener shutdown failed")
		}
	}
	// Drain the async audit outbox so no buffered audit row is lost.
	if auditWorker != nil {
		stopAudit()
		auditWorker.Wait()
	} else {
		stopAudit()
	}
}
