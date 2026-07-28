package router

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/handler"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// --- fakes: minimal but complete implementations of the two services the
// retrofitted routes touch. Every other handler is left nil in Handlers; its
// routes register (method values bind a nil receiver fine) but are never called.

type fakeMerchantSvc struct{ mid uuid.UUID }

func (f fakeMerchantSvc) Onboard(context.Context, domain.CreateMerchantRequest) (*domain.MerchantCredential, error) {
	return nil, nil
}
func (f fakeMerchantSvc) Get(context.Context, uuid.UUID) (*domain.Merchant, error) { return nil, nil }
func (f fakeMerchantSvc) ResolveByAPIKeyHash(context.Context, string) (*domain.Merchant, error) {
	return nil, domain.ErrUnauthorized // no API key path exercised here
}
func (f fakeMerchantSvc) Profile(context.Context, uuid.UUID) (*domain.MerchantProfile, error) {
	return &domain.MerchantProfile{ID: f.mid, Name: "Acme", SettlementCurrency: "THB", Status: "active"}, nil
}
func (f fakeMerchantSvc) Stats(context.Context, uuid.UUID, time.Time, time.Time) (*domain.MerchantStats, error) {
	return &domain.MerchantStats{Count: 1, VolumeMinor: 10000}, nil
}
func (f fakeMerchantSvc) StatsSeries(context.Context, uuid.UUID, int) (*domain.StatsSeries, error) {
	return &domain.StatsSeries{Days: 30}, nil
}
func (f fakeMerchantSvc) ListSettlements(context.Context, uuid.UUID, int) ([]*domain.Settlement, error) {
	return []*domain.Settlement{}, nil
}
func (f fakeMerchantSvc) RotateAPIKey(context.Context, uuid.UUID) (*domain.RotatedKey, error) {
	return &domain.RotatedKey{APIKey: "sk_new"}, nil
}
func (f fakeMerchantSvc) SetWebhook(context.Context, uuid.UUID, string) (*domain.WebhookConfig, error) {
	return &domain.WebhookConfig{WebhookURL: "https://x", SigningSecret: "whsec"}, nil
}

type fakePaymentSvc struct{ mid uuid.UUID }

func (f fakePaymentSvc) okPayment() *domain.Payment {
	return &domain.Payment{ID: uuid.New(), MerchantID: f.mid, Amount: decimal.RequireFromString("100.00"), Currency: "THB", Status: domain.StatusCaptured}
}
func (f fakePaymentSvc) Create(context.Context, string, domain.CreatePaymentRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Capture(context.Context, uuid.UUID, uuid.UUID, domain.CaptureRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Void(context.Context, uuid.UUID, uuid.UUID) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Refund(context.Context, uuid.UUID, uuid.UUID, domain.RefundRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Get(context.Context, uuid.UUID, uuid.UUID) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) List(context.Context, uuid.UUID, int32, int32) ([]*domain.Payment, error) {
	return []*domain.Payment{f.okPayment()}, nil
}
func (f fakePaymentSvc) HandleThreeDSResult(context.Context, uuid.UUID, uuid.UUID, bool) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) VerifyThreeDSResult(uuid.UUID, string, string) bool { return true }

// buildApp wires the real router with real middleware and the fakes above.
func buildApp(t *testing.T, mid uuid.UUID) (*fiber.App, *session.Manager) {
	t.Helper()
	log := zerolog.Nop()
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	msvc := fakeMerchantSvc{mid: mid}
	psvc := fakePaymentSvc{mid: mid}

	h := &handler.Handlers{
		Merchant: handler.NewMerchantHandler(msvc, log),
		Payment:  handler.NewPaymentHandler(psvc, log),
		Health:   handler.NewHealthHandler(nil),
	}
	auth := middleware.APIKeyAuth(msvc)
	sessionAuth := middleware.SessionAuth(mgr)
	merchantAuth := middleware.MerchantAuth(mgr, msvc)
	adminAuth := middleware.AdminAuth("")

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(log)})
	Setup(app, h, auth, sessionAuth, merchantAuth, adminAuth, nil, nil, nil, nil, "", false)
	return app, mgr
}

func TestRetrofit_CookieReachesDashboardRoutes(t *testing.T) {
	mid := uuid.New()
	app, mgr := buildApp(t, mid)
	tok, err := mgr.Issue(session.Claims{UserID: uuid.New(), MerchantID: mid, Email: "a@b.co"})
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	get := func(path string) int {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
		resp, _ := app.Test(req)
		return resp.StatusCode
	}
	for _, p := range []string{"/v1/me", "/v1/stats", "/v1/stats/series", "/v1/settlements", "/v1/payments", "/v1/payments/" + uuid.NewString()} {
		if code := get(p); code != 200 {
			t.Fatalf("GET %s via cookie -> %d want 200", p, code)
		}
	}

	// Refund via cookie: SameSite=Lax + Idempotency-Key gate; expect 200. A JSON
	// body is required because Fiber's BodyParser rejects requests with no
	// Content-Type (422/400) regardless of auth, so an empty/no body would fail
	// this request for a reason unrelated to the auth retrofit under test.
	rreq := httptest.NewRequest("POST", "/v1/payments/"+uuid.NewString()+"/refund", strings.NewReader(`{"amount":"10.00"}`))
	rreq.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
	rreq.Header.Set("Idempotency-Key", "idem-1")
	rreq.Header.Set("Content-Type", "application/json")
	rresp, _ := app.Test(rreq)
	if rresp.StatusCode != 200 {
		t.Fatalf("POST refund via cookie -> %d want 200", rresp.StatusCode)
	}
}

func TestRetrofit_CreateStaysApiKeyOnly(t *testing.T) {
	mid := uuid.New()
	app, mgr := buildApp(t, mid)
	tok, _ := mgr.Issue(session.Claims{UserID: uuid.New(), MerchantID: mid, Email: "a@b.co"})

	// A cookie-only caller must NOT be able to create a payment (Create stays on
	// API-key-only auth, whose resolver returns ErrUnauthorized for no key).
	req := httptest.NewRequest("POST", "/v1/payments", nil)
	req.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
	req.Header.Set("Idempotency-Key", "idem-2")
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Fatalf("POST /payments via cookie -> %d want 401", resp.StatusCode)
	}
}

func TestRetrofit_NoCredentialsRejected(t *testing.T) {
	app, _ := buildApp(t, uuid.New())
	resp, _ := app.Test(httptest.NewRequest("GET", "/v1/me", nil))
	if resp.StatusCode != 401 {
		t.Fatalf("GET /me no creds -> %d want 401", resp.StatusCode)
	}
}
