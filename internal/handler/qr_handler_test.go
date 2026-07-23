package handler

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// ---- fakeQRSvc: scriptable service.QRService for handler tests ------------

type fakeQRSvc struct {
	qp  *domain.QRPayment
	err error

	lastMerchant uuid.UUID
	lastID       uuid.UUID
	lastSig      string
}

func (f *fakeQRSvc) Create(_ context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
	f.lastMerchant = req.MerchantID
	return f.qp, f.err
}
func (f *fakeQRSvc) Get(_ context.Context, mid, id uuid.UUID) (*domain.QRPayment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.qp, f.err
}
func (f *fakeQRSvc) ConfirmFromWebhook(_ context.Context, sig string, _ []byte) (*domain.QRPayment, error) {
	f.lastSig = sig
	return f.qp, f.err
}

func okQR() *domain.QRPayment {
	return &domain.QRPayment{
		ID:        uuid.New(),
		Method:    domain.QRPromptPayDynamic,
		Amount:    decimal.RequireFromString("100.00"),
		Currency:  "THB",
		Status:    domain.QRAwaitingPayment,
		QRPayload: "0002010102...",
	}
}

func qrApp(svc *fakeQRSvc, mid uuid.UUID, authenticated bool) *fiber.App {
	h := NewQRHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	inject := func(c *fiber.Ctx) error {
		if authenticated {
			c.Locals(middleware.LocalMerchantID, mid)
		}
		return c.Next()
	}
	g := app.Group("/v1/qr-payments", inject)
	g.Post("/", h.Create)
	g.Get("/:id", h.Get)
	app.Post("/v1/webhooks/qr", h.Webhook) // no auth on the callback
	return app
}

func TestQRCreateSuccessScopesMerchant(t *testing.T) {
	mid := uuid.New()
	svc := &fakeQRSvc{qp: okQR()}
	app := qrApp(svc, mid, true)
	status, env := postJSON(t, app, "/v1/qr-payments",
		`{"method":"promptpay_dynamic","amount":"100.00","currency":"THB","reference":"R1"}`)
	if status != fiber.StatusCreated || !env.Success {
		t.Fatalf("status=%d env=%+v want 201/success", status, env)
	}
	if svc.lastMerchant != mid {
		t.Fatalf("merchant scoped to %s, want %s (must come from auth, not body)", svc.lastMerchant, mid)
	}
}

func TestQRCreateValidationError(t *testing.T) {
	app := qrApp(&fakeQRSvc{qp: okQR()}, uuid.New(), true)
	// missing method + currency.
	status, env := postJSON(t, app, "/v1/qr-payments", `{"amount":"100.00"}`)
	if status != fiber.StatusBadRequest || env.Code != "VALIDATION_ERROR" {
		t.Fatalf("status=%d code=%q want 400/VALIDATION_ERROR", status, env.Code)
	}
}

func TestQRCreateInvalidBody(t *testing.T) {
	app := qrApp(&fakeQRSvc{qp: okQR()}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/qr-payments", `{bad`)
	if status != fiber.StatusBadRequest || env.Code != "INVALID_BODY" {
		t.Fatalf("status=%d code=%q want 400/INVALID_BODY", status, env.Code)
	}
}

func TestQRGetSuccess(t *testing.T) {
	id := uuid.New()
	app := qrApp(&fakeQRSvc{qp: okQR()}, uuid.New(), true)
	status, env := getReq(t, app, "/v1/qr-payments/"+id.String())
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
}

func TestQRGetUnauthenticated(t *testing.T) {
	id := uuid.New()
	app := qrApp(&fakeQRSvc{qp: okQR()}, uuid.New(), false)
	status, env := getReq(t, app, "/v1/qr-payments/"+id.String())
	if status != fiber.StatusUnauthorized || env.Code != "UNAUTHORIZED" {
		t.Fatalf("status=%d code=%q want 401/UNAUTHORIZED", status, env.Code)
	}
}

func TestQRGetInvalidID(t *testing.T) {
	app := qrApp(&fakeQRSvc{qp: okQR()}, uuid.New(), true)
	status, env := getReq(t, app, "/v1/qr-payments/not-a-uuid")
	if status != fiber.StatusBadRequest || env.Code != "INVALID_ID" {
		t.Fatalf("status=%d code=%q want 400/INVALID_ID", status, env.Code)
	}
}

func TestQRGetNotFoundMaps404(t *testing.T) {
	id := uuid.New()
	app := qrApp(&fakeQRSvc{err: domain.ErrPaymentNotFound}, uuid.New(), true)
	status, env := getReq(t, app, "/v1/qr-payments/"+id.String())
	if status != fiber.StatusNotFound || env.Code != "PAYMENT_NOT_FOUND" {
		t.Fatalf("status=%d code=%q want 404/PAYMENT_NOT_FOUND", status, env.Code)
	}
}

// ---- Webhook -------------------------------------------------------------

func TestQRWebhookSuccessPassesSignature(t *testing.T) {
	svc := &fakeQRSvc{qp: okQR()}
	app := qrApp(svc, uuid.New(), true)

	req := postReqWithHeader("/v1/webhooks/qr", `{"reference":"R1","status":"paid"}`, "X-Signature", "abc123")
	status, env := do(t, app, req)
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
	if svc.lastSig != "abc123" {
		t.Fatalf("signature header not forwarded: %q", svc.lastSig)
	}
}

// A bad-signature/unknown webhook maps to the domain error's status without
// leaking internals.
func TestQRWebhookRejectedMaps400(t *testing.T) {
	app := qrApp(&fakeQRSvc{err: domain.ErrInvalidRequest}, uuid.New(), true)
	req := postReqWithHeader("/v1/webhooks/qr", `{"reference":"R1"}`, "X-Signature", "bad")
	status, env := do(t, app, req)
	if status != fiber.StatusBadRequest || env.Code != "INVALID_REQUEST" {
		t.Fatalf("status=%d code=%q want 400/INVALID_REQUEST", status, env.Code)
	}
}

func TestQRWebhookNotFoundMaps404(t *testing.T) {
	app := qrApp(&fakeQRSvc{err: domain.ErrPaymentNotFound}, uuid.New(), true)
	req := postReqWithHeader("/v1/webhooks/qr", `{"provider_ref":"nope"}`, "X-Signature", "sig")
	status, env := do(t, app, req)
	if status != fiber.StatusNotFound || env.Code != "PAYMENT_NOT_FOUND" {
		t.Fatalf("status=%d code=%q want 404/PAYMENT_NOT_FOUND", status, env.Code)
	}
}
