package middleware

import (
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/config"
	"github.com/yourco/payment-gateway/internal/domain"
)

// Register wires cross-cutting middleware in the correct order. Rate limiting is
// intentionally NOT applied here: it is mounted per authenticated route group
// (see router.Setup + RateLimiter) so health/metrics/webhook stay unlimited and
// the limit is keyed per merchant rather than shared app-wide.
func Register(app *fiber.App, cfg *config.Config, log zerolog.Logger) {
	app.Use(recover.New(recover.Config{EnableStackTrace: !cfg.IsProd()}))
	app.Use(requestID())
	app.Use(requestLogger(log, cfg.IsProd()))

	// Security headers (nosniff, DENY framing, HSTS, referrer policy). Cheap and
	// on every response.
	app.Use(helmet.New(helmet.Config{
		XSSProtection:         "0",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ReferrerPolicy:        "no-referrer",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: false,
	}))

	// CORS locked to an explicit allowlist from config. Never '*' on an
	// authenticated money API; an empty allowlist disables CORS entirely.
	if origins := cfg.CORSAllowOriginList(); origins != "" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     "GET,POST,PUT",
			AllowHeaders:     "Content-Type, Authorization, Idempotency-Key, X-API-Key, X-Admin-Key, X-Signature, X-Request-ID",
			AllowCredentials: false,
		}))
	}
}

// RateLimiter returns a per-route-group limiter keyed by the authenticated
// merchant id (falling back to client IP before auth resolves). Mount it AFTER
// the auth middleware on the payments / QR groups. The ceiling is driven from
// config (requests per second) so load tests can run unthrottled.
func RateLimiter(perSec int) fiber.Handler {
	if perSec <= 0 {
		perSec = 600
	}
	return limiter.New(limiter.Config{
		Max:        perSec,
		Expiration: time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			if id, ok := MerchantIDFromCtx(c); ok {
				return "m:" + id.String()
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.Error(c, fiber.StatusTooManyRequests, "RATE_LIMITED", "rate limit exceeded")
		},
	})
}

// SignupRateLimiter returns a limiter for the public, unauthenticated merchant
// onboarding endpoint (POST /v1/merchants), keyed by client IP. Self-service
// sandbox signup is open to the internet, so it gets its own dedicated limiter
// (separate from the per-merchant money-route limiter) to blunt abuse: at most
// perHour signups per IP per rolling hour. Exceeding it returns 429 in the
// standard error envelope. perHour <= 0 falls back to 5.
// clientIP resolves the real caller's IP for IP-keyed rate limiting, working
// correctly behind a reverse proxy. Fiber's c.IP() (even with ProxyHeader) can
// return the platform's rotating internal hop, which breaks per-IP counting. We
// prefer Envoy's X-Envoy-External-Address (a single, proxy-determined client
// address that a client cannot forge by prepending X-Forwarded-For — Railway
// fronts services with Envoy), then the left-most X-Forwarded-For entry, and
// finally the socket peer for direct/local serving.
func clientIP(c *fiber.Ctx) string {
	if envoy := strings.TrimSpace(c.Get("X-Envoy-External-Address")); envoy != "" {
		return envoy
	}
	if xff := strings.TrimSpace(c.Get(fiber.HeaderXForwardedFor)); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return xff
	}
	return c.IP()
}

func SignupRateLimiter(perHour int) fiber.Handler {
	if perHour <= 0 {
		perHour = 5
	}
	return limiter.New(limiter.Config{
		Max:        perHour,
		Expiration: time.Hour,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "signup:" + clientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.Error(c, fiber.StatusTooManyRequests, "RATE_LIMITED", "too many signups from this address; please try again later")
		},
	})
}

func requestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Locals("request_id", id)
		c.Set("X-Request-ID", id)
		return c.Next()
	}
}

// logSampleEvery logs 1 in N successful (2xx/3xx) requests on the hot path. A
// deterministic atomic counter is used (not a PRNG) so it is allocation-free and
// contention-cheap. Non-2xx responses are always logged so errors are never
// dropped; sampling only thins the successful requests that would otherwise cap
// throughput via serialized stdout I/O.
const logSampleEvery = 20 // ~5%

// requestLogger logs method/path/status/latency. It deliberately never logs the
// request body — bodies may contain cardholder-data-adjacent fields. On the hot
// path (prod) successful (2xx/3xx) requests are sampled; errors are always logged.
func requestLogger(log zerolog.Logger, prod bool) fiber.Handler {
	var counter atomic.Uint64
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		status := c.Response().StatusCode()

		// Always log >=400; sample the rest in prod to keep stdout I/O off the hot path.
		if prod && status < 400 && counter.Add(1)%logSampleEvery != 0 {
			return err
		}

		reqID, _ := c.Locals("request_id").(string)
		ev := log.Info()
		if status >= 500 {
			ev = log.Error()
		} else if status >= 400 {
			ev = log.Warn()
		}
		ev.
			Str("request_id", reqID).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", time.Since(start)).
			Msg("request")
		return err
	}
}

// ErrorHandler maps domain errors to stable HTTP responses. Every declared
// domain error must have a case here so genuine client outcomes (declines,
// insufficient funds, 3DS challenges) never surface as a generic 500.
func ErrorHandler(log zerolog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Fiber's own router errors (an unmatched route -> 404, wrong method ->
		// 405) carry an explicit HTTP status. Honor it instead of collapsing them
		// to a generic 500 so a disabled route (e.g. the sandbox endpoints when
		// SANDBOX_MODE is off) correctly returns 404, not 500.
		var fe *fiber.Error
		if errors.As(err, &fe) {
			code := "ERROR"
			switch fe.Code {
			case fiber.StatusNotFound:
				code = "NOT_FOUND"
			case fiber.StatusMethodNotAllowed:
				code = "METHOD_NOT_ALLOWED"
			}
			return domain.Error(c, fe.Code, code, fe.Message)
		}
		switch {
		case errors.Is(err, domain.ErrPaymentNotFound):
			return domain.Error(c, fiber.StatusNotFound, "PAYMENT_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrMerchantNotFound):
			return domain.Error(c, fiber.StatusNotFound, "MERCHANT_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrPaymentLinkNotFound):
			return domain.Error(c, fiber.StatusNotFound, "PAYMENT_LINK_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrDisputeNotFound):
			return domain.Error(c, fiber.StatusNotFound, "DISPUTE_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrUnauthorized):
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		case errors.Is(err, domain.ErrForbidden):
			return domain.Error(c, fiber.StatusForbidden, "FORBIDDEN", err.Error())
		case errors.Is(err, domain.ErrDuplicateRequest):
			return domain.Error(c, fiber.StatusConflict, "DUPLICATE_REQUEST", err.Error())
		case errors.Is(err, domain.ErrIdempotencyKeyReuse):
			return domain.Error(c, fiber.StatusConflict, "IDEMPOTENCY_KEY_REUSED", err.Error())
		case errors.Is(err, domain.ErrThreeDSRequired):
			return domain.Error(c, fiber.StatusConflict, "THREE_DS_REQUIRED", err.Error())
		case errors.Is(err, domain.ErrInsufficientFund):
			return domain.Error(c, fiber.StatusPaymentRequired, "INSUFFICIENT_FUNDS", err.Error())
		case errors.Is(err, domain.ErrRefundExceeds):
			return domain.Error(c, fiber.StatusUnprocessableEntity, "REFUND_EXCEEDS_CAPTURED", err.Error())
		case errors.Is(err, domain.ErrCardDeclined):
			return domain.Error(c, fiber.StatusPaymentRequired, "CARD_DECLINED", err.Error())
		case errors.Is(err, domain.ErrInvalidState):
			// A state-transition conflict (e.g. voiding a captured payment) is a
			// 409, not a 400: the request was well-formed but conflicts with the
			// resource's current state.
			return domain.Error(c, fiber.StatusConflict, "INVALID_STATE", err.Error())
		case errors.Is(err, domain.ErrInvalidRequest):
			return domain.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", err.Error())
		default:
			log.Error().Err(err).Str("path", c.Path()).Msg("unhandled error")
			return domain.Error(c, fiber.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
	}
}
