package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/handler"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// Setup registers all routes. The auth argument is the API-key auth middleware
// applied to merchant-scoped route groups; health probes, the QR webhook and
// merchant onboarding stay open. adminAuth gates the /v1/admin operator console
// (X-Admin-Key, constant-time compare). rateLimit is a per-merchant limiter
// mounted on the money route groups only (nil disables it). metrics, when
// non-nil, is mounted at /metrics on the public listener — pass nil to keep it
// off the public listener (bind it to a separate internal listener instead).
// webDir, when non-empty, is served as static files so the dashboard / admin /
// checkout UIs run same-origin with the API. signupLimit, when non-nil, is an
// IP-keyed limiter mounted ONLY on the public POST /v1/merchants onboarding
// endpoint (nil disables it). sandbox, when true, mounts the PUBLIC sandbox
// payer-simulator endpoints (/v1/sandbox/...) — it MUST be false in production
// so those routes are completely absent (404) and no one can mark payments paid
// without a signed bank webhook.
func Setup(app *fiber.App, h *handler.Handlers, auth, adminAuth, rateLimit, signupLimit fiber.Handler, metrics fiber.Handler, webDir string, sandbox bool) {
	// Health / probes (no auth, never rate limited).
	app.Get("/healthz", h.Health.Live)
	app.Get("/readyz", h.Health.Ready)

	// Prometheus scrape endpoint. Only mounted on the public listener when
	// explicitly enabled; production binds it on a separate internal listener.
	if metrics != nil {
		app.Get("/metrics", metrics)
	}

	v1 := app.Group("/v1")

	// Merchant onboarding is unauthenticated (there is no API key yet); the
	// merchant lookup is protected by API-key auth. Because it is public and
	// self-service, POST /v1/merchants carries its own dedicated IP-keyed rate
	// limiter (mounted here only) so a single address cannot mass-create merchants.
	if signupLimit != nil {
		v1.Post("/merchants", signupLimit, h.Merchant.Onboard)
	} else {
		v1.Post("/merchants", h.Merchant.Onboard)
	}
	v1.Get("/merchants/:id", auth, h.Merchant.Get)

	// Merchant Dashboard (self-service). Everything resolves the merchant from
	// the API-key auth context; the request body is never trusted for identity.
	v1.Get("/me", auth, h.Merchant.Me)
	v1.Post("/me/rotate-key", auth, h.Merchant.RotateKey)
	v1.Put("/me/webhook", auth, h.Merchant.SetWebhook)
	v1.Get("/stats", auth, h.Merchant.Stats)
	v1.Get("/settlements", auth, h.Merchant.Settlements)
	v1.Get("/disputes", auth, h.Dispute.ListByMerchant)

	// Payments require merchant API-key auth. The resolved merchant scopes every
	// request. The rate limiter is mounted AFTER auth so it keys per merchant.
	payments := v1.Group("/payments", auth)
	if rateLimit != nil {
		payments.Use(rateLimit)
	}
	payments.Get("/", h.Payment.List)
	// Money-moving creation requires an idempotency key.
	payments.Post("/", middleware.RequireIdempotencyKey(), h.Payment.Create)
	payments.Get("/:id", h.Payment.Get)
	payments.Post("/:id/capture", middleware.RequireIdempotencyKey(), h.Payment.Capture)
	payments.Post("/:id/refund", middleware.RequireIdempotencyKey(), h.Payment.Refund)
	payments.Post("/:id/void", h.Payment.Void)
	payments.Post("/:id/3ds/return", h.Payment.ThreeDSReturn)

	// Chargebacks / disputes, scoped to the payment.
	payments.Post("/:id/disputes", h.Dispute.Open)
	payments.Get("/:id/disputes", h.Dispute.ListByPayment)

	// QR payments (PromptPay dynamic/static, card-scheme QR, cross-border).
	qr := v1.Group("/qr-payments", auth)
	if rateLimit != nil {
		qr.Use(rateLimit)
	}
	qr.Post("/", h.QR.Create)
	qr.Get("/:id", h.QR.Get) // merchant polls; or receives a webhook

	// Inbound bank/PSP confirmation callback (signature-verified, no API key,
	// never rate limited so legitimate bank retries are not dropped).
	v1.Post("/webhooks/qr", h.QR.Webhook)

	// Platform operator console. Gated by X-Admin-Key (constant-time compare);
	// when no admin key is configured every request here 401s.
	admin := v1.Group("/admin", adminAuth)
	admin.Get("/merchants", h.Admin.Merchants)
	admin.Get("/audit-log", h.Admin.AuditLog)
	admin.Get("/stats", h.Admin.Stats)
	admin.Get("/disputes", h.Admin.Disputes)
	admin.Get("/settlements", h.Admin.Settlements)

	// Sandbox payer simulator (the "Sandbox Bank" app). Mounted ONLY when
	// SANDBOX_MODE=true AND the handler is wired; every route is unauthenticated
	// (a browser with no merchant key) and the server signs the confirmation
	// internally. When sandbox is false these routes never register, so a real
	// deploy returns 404 and no one can mark a payment paid without a signed bank
	// webhook.
	if sandbox && h.Sandbox != nil {
		sb := v1.Group("/sandbox")
		sb.Get("/qr-payments", h.Sandbox.List)
		sb.Get("/qr-payments/:id", h.Sandbox.View)
		sb.Post("/qr-payments/:id/pay", h.Sandbox.Pay)
		sb.Post("/qr-payments/:id/decline", h.Sandbox.Decline)
	}

	// Static site so the dashboard / admin / checkout UIs run same-origin with
	// the API (no CORS). Mounted LAST so it never shadows an API route. Files are
	// served read-only; a missing file 404s rather than listing the directory.
	if webDir != "" {
		app.Static("/", webDir, fiber.Static{
			Browse: false,
		})
	}
}
