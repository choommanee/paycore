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

// ---- fakeSandboxSvc: scriptable service.SandboxService for handler tests -----

type fakeSandboxSvc struct {
	view *domain.SandboxQRView
	qp   *domain.QRPayment
	list []domain.SandboxQRListItem
	err  error

	lastID     uuid.UUID
	lastStatus string // "paid" | "failed" from Pay/Decline
}

func (f *fakeSandboxSvc) View(_ context.Context, id uuid.UUID) (*domain.SandboxQRView, error) {
	f.lastID = id
	return f.view, f.err
}
func (f *fakeSandboxSvc) Pay(_ context.Context, id uuid.UUID) (*domain.QRPayment, error) {
	f.lastID, f.lastStatus = id, "paid"
	return f.qp, f.err
}
func (f *fakeSandboxSvc) Decline(_ context.Context, id uuid.UUID) (*domain.QRPayment, error) {
	f.lastID, f.lastStatus = id, "failed"
	return f.qp, f.err
}
func (f *fakeSandboxSvc) ListPending(_ context.Context, _ string, _ int32) ([]domain.SandboxQRListItem, error) {
	return f.list, f.err
}

func sandboxView() *domain.SandboxQRView {
	return &domain.SandboxQRView{
		ID:           uuid.New(),
		Amount:       decimal.RequireFromString("1290.00"),
		AmountMinor:  129000,
		Currency:     "THB",
		MerchantName: "Acme Store",
		Method:       domain.QRPromptPayDynamic,
		Status:       domain.QRAwaitingPayment,
		Reference:    "INV-1",
	}
}

func paidQR(status domain.QRStatus) *domain.QRPayment {
	return &domain.QRPayment{
		ID:       uuid.New(),
		Method:   domain.QRPromptPayDynamic,
		Amount:   decimal.RequireFromString("1290.00"),
		Currency: "THB",
		Status:   status,
	}
}

// sandboxApp mounts the sandbox routes with the SAME gate router.Setup applies
// (mount ONLY when sandbox is true AND the handler is wired). Mirrored here
// rather than calling router.Setup to avoid an import cycle (router imports
// handler). When sandbox is false the routes must be completely absent (404).
func sandboxApp(svc *fakeSandboxSvc, sandbox bool) *fiber.App {
	var sh *SandboxHandler
	if svc != nil {
		sh = NewSandboxHandler(svc, zerolog.Nop())
	}
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	v1 := app.Group("/v1")
	if sandbox && sh != nil {
		sb := v1.Group("/sandbox")
		sb.Get("/qr-payments", sh.List)
		sb.Get("/qr-payments/:id", sh.View)
		sb.Post("/qr-payments/:id/pay", sh.Pay)
		sb.Post("/qr-payments/:id/decline", sh.Decline)
	}
	return app
}

func TestSandboxView(t *testing.T) {
	view := sandboxView()
	svc := &fakeSandboxSvc{view: view}
	app := sandboxApp(svc, true)
	status, env := getReq(t, app, "/v1/sandbox/qr-payments/"+view.ID.String())
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
	data, _ := env.Data.(map[string]any)
	// decimal marshals "1290.00" as "1290" (trailing zeros dropped); the exact
	// value is asserted via amount_minor which is the integer-satang source.
	if data["amount"] != "1290" {
		t.Fatalf("amount echoed=%v want 1290", data["amount"])
	}
	if data["amount_minor"] != float64(129000) {
		t.Fatalf("amount_minor=%v want 129000", data["amount_minor"])
	}
	if data["merchant_name"] != "Acme Store" {
		t.Fatalf("merchant_name=%v want Acme Store", data["merchant_name"])
	}
}

func TestSandboxViewUnknownID404(t *testing.T) {
	svc := &fakeSandboxSvc{err: domain.ErrPaymentNotFound}
	app := sandboxApp(svc, true)
	status, env := getReq(t, app, "/v1/sandbox/qr-payments/"+uuid.NewString())
	if status != fiber.StatusNotFound || env.Code != "PAYMENT_NOT_FOUND" {
		t.Fatalf("status=%d code=%q want 404/PAYMENT_NOT_FOUND", status, env.Code)
	}
}

func TestSandboxViewInvalidID(t *testing.T) {
	app := sandboxApp(&fakeSandboxSvc{view: sandboxView()}, true)
	status, env := getReq(t, app, "/v1/sandbox/qr-payments/not-a-uuid")
	if status != fiber.StatusBadRequest || env.Code != "INVALID_ID" {
		t.Fatalf("status=%d code=%q want 400/INVALID_ID", status, env.Code)
	}
}

func TestSandboxPayAndDecline(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		retQR    *domain.QRPayment
		wantStat string // svc.lastStatus
		wantQR   string // returned status in envelope
	}{
		{"pay->paid", "/pay", paidQR(domain.QRPaid), "paid", "paid"},
		{"decline->failed", "/decline", paidQR(domain.QRFailed), "failed", "failed"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &fakeSandboxSvc{qp: tc.retQR}
			app := sandboxApp(svc, true)
			id := uuid.New()
			status, env := postJSON(t, app, "/v1/sandbox/qr-payments/"+id.String()+tc.path, ``)
			if status != fiber.StatusOK || !env.Success {
				t.Fatalf("status=%d env=%+v want 200/success", status, env)
			}
			if svc.lastStatus != tc.wantStat {
				t.Fatalf("service invoked with status=%q want %q", svc.lastStatus, tc.wantStat)
			}
			if svc.lastID != id {
				t.Fatalf("service id=%s want %s", svc.lastID, id)
			}
			data, _ := env.Data.(map[string]any)
			if data["status"] != tc.wantQR {
				t.Fatalf("returned status=%v want %v", data["status"], tc.wantQR)
			}
		})
	}
}

// TestSandboxPayIdempotentAlreadyPaid: an already-paid QR still returns paid.
func TestSandboxPayIdempotentAlreadyPaid(t *testing.T) {
	svc := &fakeSandboxSvc{qp: paidQR(domain.QRPaid)}
	app := sandboxApp(svc, true)
	status, env := postJSON(t, app, "/v1/sandbox/qr-payments/"+uuid.NewString()+"/pay", ``)
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
	data, _ := env.Data.(map[string]any)
	if data["status"] != "paid" {
		t.Fatalf("status=%v want paid", data["status"])
	}
}

// TestSandboxGateOff: with SANDBOX_MODE off, every sandbox route must be absent
// (404), even when a handler is wired.
func TestSandboxGateOff(t *testing.T) {
	svc := &fakeSandboxSvc{view: sandboxView(), qp: paidQR(domain.QRPaid)}
	app := sandboxApp(svc, false)
	id := uuid.NewString()
	paths := []struct {
		method, path string
	}{
		{fiber.MethodGet, "/v1/sandbox/qr-payments"},
		{fiber.MethodGet, "/v1/sandbox/qr-payments/" + id},
		{fiber.MethodPost, "/v1/sandbox/qr-payments/" + id + "/pay"},
		{fiber.MethodPost, "/v1/sandbox/qr-payments/" + id + "/decline"},
	}
	for _, p := range paths {
		var st int
		if p.method == fiber.MethodGet {
			st, _ = getReq(t, app, p.path)
		} else {
			st, _ = postJSON(t, app, p.path, ``)
		}
		if st != fiber.StatusNotFound {
			t.Fatalf("%s %s: status=%d want 404 (sandbox gate off)", p.method, p.path, st)
		}
	}
}
