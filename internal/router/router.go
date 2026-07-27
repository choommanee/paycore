package router

import (
	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/handler"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// Setup registers all routes. The auth argument is the API-key auth middleware
// applied to merchant-scoped route groups; health probes, the QR webhook and
// merchant onboarding stay open. sessionAuth is the dashboard human-login
// session middleware (pc_session cookie) gating GET /v1/auth/me; it has no
// effect unless h.Auth is wired. merchantAuth is the combined session-cookie-OR-
// API-key middleware gating the merchant dashboard routes (/v1/me, /v1/stats,
// /v1/settlements, /v1/disputes), /v1/payment-links (no effect unless
// h.PaymentLink is wired), and the read/refund /v1/payments routes. adminAuth
// gates the /v1/admin operator console
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
// without a signed bank webhook; it also gates POST /v1/auth/dev-login.
// checkoutLimit, when non-nil, is an IP-keyed limiter mounted ONLY on the public
// POST /v1/checkout/sessions session-creation endpoint (nil disables it); the
// other two checkout routes are unauthenticated but unlimited.
func Setup(app *fiber.App, h *handler.Handlers, auth, sessionAuth, merchantAuth, adminAuth, rateLimit, signupLimit, checkoutLimit fiber.Handler, metrics fiber.Handler, webDir string, sandbox bool) {
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

	// Merchant Dashboard (self-service). Retrofitted to merchantAuth so BOTH the
	// cookie dashboard and API-key clients reach these merchant-scoped routes.
	// merchantAuth sets LocalMerchantID identically to auth, so the handlers are
	// unchanged. Everything resolves the merchant from the auth context; the
	// request body is never trusted for identity.
	v1.Get("/me", merchantAuth, h.Merchant.Me)
	v1.Post("/me/rotate-key", merchantAuth, h.Merchant.RotateKey)
	v1.Put("/me/webhook", merchantAuth, h.Merchant.SetWebhook)
	v1.Get("/stats", merchantAuth, h.Merchant.Stats)
	v1.Get("/settlements", merchantAuth, h.Merchant.Settlements)
	v1.Get("/disputes", merchantAuth, h.Dispute.ListByMerchant)

	// Dashboard human auth (Google OIDC + session cookie). Public except /me.
	if h.Auth != nil {
		v1.Get("/auth/google/start", h.Auth.GoogleStart)
		v1.Get("/auth/google/callback", h.Auth.GoogleCallback)
		v1.Post("/auth/logout", h.Auth.Logout)
		v1.Get("/auth/me", sessionAuth, h.Auth.Me)
		if sandbox {
			v1.Post("/auth/dev-login", h.Auth.DevLogin)
		}
	}

	// Payment links (dashboard + API). merchantAuth accepts a session cookie OR
	// an API key. Reads/updates are merchant-scoped in the service (IDOR-safe).
	if h.PaymentLink != nil {
		links := v1.Group("/payment-links", merchantAuth)
		links.Post("/", h.PaymentLink.Create)
		links.Get("/", h.PaymentLink.List)
		links.Get("/:id", h.PaymentLink.Get)
		links.Patch("/:id", h.PaymentLink.Disable)
	}

	// Public hosted checkout (unauthenticated). The opaque session token in the
	// URL path is the credential; the service scopes every op to that session.
	// Session creation is IP-rate-limited (checkoutLimit) to blunt abuse.
	if h.Checkout != nil {
		checkout := v1.Group("/checkout")
		if checkoutLimit != nil {
			checkout.Post("/sessions", checkoutLimit, h.Checkout.Create)
		} else {
			checkout.Post("/sessions", h.Checkout.Create)
		}
		checkout.Get("/sessions/:token", h.Checkout.Get)
		checkout.Post("/sessions/:token/pay", h.Checkout.Pay)
		// Wallet approve/decline simulator. Mounted ONLY when SANDBOX_MODE=true so
		// production has no way to mark a session paid without a real PSP callback
		// (route absent -> 404). Public: the session token is the credential.
		if sandbox {
			checkout.Post("/sessions/:token/confirm-mock", h.Checkout.ConfirmMock)
		}
	}

	// Payments — read + refund routes accept a session cookie OR an API key so
	// the cookie dashboard can list/inspect/refund. The rate limiter is mounted
	// AFTER auth so it keys per merchant (merchantAuth sets LocalMerchantID just
	// like auth). Refund is money-moving: it keeps its idempotency-key gate and
	// relies on the pc_session cookie being SameSite=Lax for CSRF protection.
	readPayments := v1.Group("/payments", merchantAuth)
	if rateLimit != nil {
		readPayments.Use(rateLimit)
	}
	readPayments.Get("/", h.Payment.List)
	readPayments.Get("/:id", h.Payment.Get)
	readPayments.Post("/:id/refund", middleware.RequireIdempotencyKey(), h.Payment.Refund)
	readPayments.Get("/:id/disputes", h.Dispute.ListByPayment)

	// Money-moving / lifecycle routes stay API-key-only (auth). A cookie-only
	// caller is rejected here even though it can reach the read routes above.
	writePayments := v1.Group("/payments", auth)
	if rateLimit != nil {
		writePayments.Use(rateLimit)
	}
	writePayments.Post("/", middleware.RequireIdempotencyKey(), h.Payment.Create)
	writePayments.Post("/:id/capture", middleware.RequireIdempotencyKey(), h.Payment.Capture)
	writePayments.Post("/:id/void", h.Payment.Void)
	writePayments.Post("/:id/3ds/return", h.Payment.ThreeDSReturn)

	// Chargebacks / disputes: opening one is a write (API key); listing is a read.
	writePayments.Post("/:id/disputes", h.Dispute.Open)

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
