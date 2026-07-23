package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// ---- fakePaymentSvc: scriptable service.PaymentService for handler tests ----

type fakePaymentSvc struct {
	payment *domain.Payment
	err     error

	// captured call args (for assertions).
	lastMerchant uuid.UUID
	lastID       uuid.UUID
	lastIdem     string
}

func (f *fakePaymentSvc) Create(_ context.Context, idem string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
	f.lastIdem = idem
	f.lastMerchant = req.MerchantID
	return f.payment, f.err
}
func (f *fakePaymentSvc) Capture(_ context.Context, mid, id uuid.UUID, _ domain.CaptureRequest) (*domain.Payment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.payment, f.err
}
func (f *fakePaymentSvc) Void(_ context.Context, mid, id uuid.UUID) (*domain.Payment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.payment, f.err
}
func (f *fakePaymentSvc) Refund(_ context.Context, mid, id uuid.UUID, _ domain.RefundRequest) (*domain.Payment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.payment, f.err
}
func (f *fakePaymentSvc) Get(_ context.Context, mid, id uuid.UUID) (*domain.Payment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.payment, f.err
}
func (f *fakePaymentSvc) List(_ context.Context, mid uuid.UUID, _, _ int32) ([]*domain.Payment, error) {
	f.lastMerchant = mid
	if f.err != nil {
		return nil, f.err
	}
	return []*domain.Payment{f.payment}, nil
}
func (f *fakePaymentSvc) HandleThreeDSResult(_ context.Context, mid, id uuid.UUID, _ bool) (*domain.Payment, error) {
	f.lastMerchant, f.lastID = mid, id
	return f.payment, f.err
}
func (f *fakePaymentSvc) VerifyThreeDSResult(_ uuid.UUID, _, _ string) bool { return true }

func okPayment() *domain.Payment {
	return &domain.Payment{
		ID:       uuid.New(),
		Amount:   decimal.RequireFromString("100.00"),
		Currency: "THB",
		Status:   domain.StatusAuthorized,
	}
}

// paymentApp mounts the payment routes with an injected authenticated merchant
// and a shared central ErrorHandler so error mapping is exercised end to end.
func paymentApp(svc *fakePaymentSvc, mid uuid.UUID, authenticated bool) *fiber.App {
	h := NewPaymentHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	inject := func(c *fiber.Ctx) error {
		if authenticated {
			c.Locals(middleware.LocalMerchantID, mid)
		}
		c.Locals("idempotency_key", "idem-test")
		return c.Next()
	}
	g := app.Group("/v1/payments", inject)
	g.Post("/", h.Create)
	g.Get("/", h.List)
	g.Get("/:id", h.Get)
	g.Post("/:id/capture", h.Capture)
	g.Post("/:id/refund", h.Refund)
	g.Post("/:id/void", h.Void)
	return app
}

func postJSON(t *testing.T, app *fiber.App, path, body string) (int, domain.APIResponse) {
	t.Helper()
	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return do(t, app, req)
}

func getReq(t *testing.T, app *fiber.App, path string) (int, domain.APIResponse) {
	t.Helper()
	return do(t, app, httptest.NewRequest(fiber.MethodGet, path, nil))
}

// postReqWithHeader builds a JSON POST with one extra header (e.g. X-Signature).
func postReqWithHeader(path, body, hk, hv string) *http.Request {
	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(hk, hv)
	return req
}

func do(t *testing.T, app *fiber.App, req *http.Request) (int, domain.APIResponse) {
	t.Helper()
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	_ = json.Unmarshal(body, &env)
	return resp.StatusCode, env
}

// ---- Create --------------------------------------------------------------

func TestPaymentCreateSuccess(t *testing.T) {
	mid := uuid.New()
	svc := &fakePaymentSvc{payment: okPayment()}
	app := paymentApp(svc, mid, true)

	status, env := postJSON(t, app, "/v1/payments",
		`{"amount":"100.00","currency":"THB","payment_token":"tok_x"}`)
	if status != fiber.StatusCreated {
		t.Fatalf("status=%d want 201", status)
	}
	if !env.Success || env.Code != "CREATED" {
		t.Fatalf("envelope=%+v want success/CREATED", env)
	}
	// merchant id must come from auth context, never the body.
	if svc.lastMerchant != mid {
		t.Fatalf("merchant scoped to %s, want %s", svc.lastMerchant, mid)
	}
	if svc.lastIdem != "idem-test" {
		t.Fatalf("idempotency key not propagated: %q", svc.lastIdem)
	}
}

func TestPaymentCreateInvalidBody(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments", `{not-json`)
	if status != fiber.StatusBadRequest || env.Code != "INVALID_BODY" {
		t.Fatalf("status=%d code=%q want 400/INVALID_BODY", status, env.Code)
	}
}

func TestPaymentCreateValidationError(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), true)
	// Missing payment_token + currency.
	status, env := postJSON(t, app, "/v1/payments", `{"amount":"100.00"}`)
	if status != fiber.StatusBadRequest || env.Code != "VALIDATION_ERROR" {
		t.Fatalf("status=%d code=%q want 400/VALIDATION_ERROR", status, env.Code)
	}
}

// A service-layer decline maps to 402 CARD_DECLINED via the central handler.
func TestPaymentCreateDeclineMaps402(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{err: domain.ErrCardDeclined}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments",
		`{"amount":"100.00","currency":"THB","payment_token":"tok_x"}`)
	if status != fiber.StatusPaymentRequired || env.Code != "CARD_DECLINED" {
		t.Fatalf("status=%d code=%q want 402/CARD_DECLINED", status, env.Code)
	}
}

func TestPaymentCreate3DSRequiredMaps409(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{err: domain.ErrThreeDSRequired}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments",
		`{"amount":"100.00","currency":"THB","payment_token":"tok_x"}`)
	if status != fiber.StatusConflict || env.Code != "THREE_DS_REQUIRED" {
		t.Fatalf("status=%d code=%q want 409/THREE_DS_REQUIRED", status, env.Code)
	}
}

// ---- Capture / Void / Refund ---------------------------------------------

func TestPaymentCaptureSuccess(t *testing.T) {
	mid, pid := uuid.New(), uuid.New()
	svc := &fakePaymentSvc{payment: okPayment()}
	app := paymentApp(svc, mid, true)
	status, env := postJSON(t, app, "/v1/payments/"+pid.String()+"/capture", `{"amount":"100.00"}`)
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
	if svc.lastID != pid || svc.lastMerchant != mid {
		t.Fatalf("capture not scoped: id=%s mid=%s", svc.lastID, svc.lastMerchant)
	}
}

func TestPaymentCaptureInvalidID(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments/not-a-uuid/capture", `{"amount":"1.00"}`)
	if status != fiber.StatusBadRequest || env.Code != "INVALID_ID" {
		t.Fatalf("status=%d code=%q want 400/INVALID_ID", status, env.Code)
	}
}

// Voiding a captured payment is an invalid transition -> 409 INVALID_STATE.
func TestPaymentVoidInvalidStateMaps409(t *testing.T) {
	pid := uuid.New()
	app := paymentApp(&fakePaymentSvc{err: domain.ErrInvalidState}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments/"+pid.String()+"/void", ``)
	if status != fiber.StatusConflict || env.Code != "INVALID_STATE" {
		t.Fatalf("status=%d code=%q want 409/INVALID_STATE", status, env.Code)
	}
}

func TestPaymentRefundExceedsMaps422(t *testing.T) {
	pid := uuid.New()
	app := paymentApp(&fakePaymentSvc{err: domain.ErrRefundExceeds}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments/"+pid.String()+"/refund", `{"amount":"9999.00"}`)
	if status != fiber.StatusUnprocessableEntity || env.Code != "REFUND_EXCEEDS_CAPTURED" {
		t.Fatalf("status=%d code=%q want 422/REFUND_EXCEEDS_CAPTURED", status, env.Code)
	}
}

func TestPaymentRefundNotFoundMaps404(t *testing.T) {
	pid := uuid.New()
	app := paymentApp(&fakePaymentSvc{err: domain.ErrPaymentNotFound}, uuid.New(), true)
	status, env := postJSON(t, app, "/v1/payments/"+pid.String()+"/refund", `{"amount":"1.00"}`)
	if status != fiber.StatusNotFound || env.Code != "PAYMENT_NOT_FOUND" {
		t.Fatalf("status=%d code=%q want 404/PAYMENT_NOT_FOUND", status, env.Code)
	}
}

// ---- Get / List ----------------------------------------------------------

func TestPaymentGetSuccess(t *testing.T) {
	pid := uuid.New()
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), true)
	status, env := getReq(t, app, "/v1/payments/"+pid.String())
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
}

func TestPaymentListSuccess(t *testing.T) {
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), true)
	status, env := getReq(t, app, "/v1/payments/?limit=10&offset=0")
	if status != fiber.StatusOK || !env.Success {
		t.Fatalf("status=%d env=%+v want 200/success", status, env)
	}
}

// ---- auth-context rejection paths ----------------------------------------

// Every merchant-scoped handler must reject a request with no authenticated
// merchant in context (401), not panic or leak.
func TestPaymentHandlersRejectUnauthenticated(t *testing.T) {
	pid := uuid.New()
	app := paymentApp(&fakePaymentSvc{payment: okPayment()}, uuid.New(), false /* not authenticated */)

	cases := []struct {
		method, path, body string
	}{
		{fiber.MethodGet, "/v1/payments/" + pid.String(), ""},
		{fiber.MethodGet, "/v1/payments/", ""},
		{fiber.MethodPost, "/v1/payments/" + pid.String() + "/capture", `{"amount":"1.00"}`},
		{fiber.MethodPost, "/v1/payments/" + pid.String() + "/refund", `{"amount":"1.00"}`},
		{fiber.MethodPost, "/v1/payments/" + pid.String() + "/void", ``},
	}
	for _, tc := range cases {
		var status int
		var env domain.APIResponse
		if tc.method == fiber.MethodGet {
			status, env = getReq(t, app, tc.path)
		} else {
			status, env = postJSON(t, app, tc.path, tc.body)
		}
		if status != fiber.StatusUnauthorized || env.Code != "UNAUTHORIZED" {
			t.Fatalf("%s %s -> %d/%q, want 401/UNAUTHORIZED", tc.method, tc.path, status, env.Code)
		}
	}
}
