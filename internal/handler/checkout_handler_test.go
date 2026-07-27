package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

type fakeCheckoutSvc struct {
	createFn      func(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error)
	getFn         func(ctx context.Context, token string) (*domain.CheckoutSessionView, error)
	payFn         func(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error)
	confirmMockFn func(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error)
}

func (f *fakeCheckoutSvc) CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error) {
	return f.createFn(ctx, publicID)
}
func (f *fakeCheckoutSvc) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	return f.getFn(ctx, token)
}
func (f *fakeCheckoutSvc) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	return f.payFn(ctx, token, req)
}
func (f *fakeCheckoutSvc) ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
	return f.confirmMockFn(ctx, token, approve)
}

func newCheckoutApp(svc *fakeCheckoutSvc) *fiber.App {
	h := NewCheckoutHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/checkout/sessions", h.Create)
	app.Get("/v1/checkout/sessions/:token", h.Get)
	app.Post("/v1/checkout/sessions/:token/pay", h.Pay)
	app.Post("/v1/checkout/sessions/:token/confirm-mock", h.ConfirmMock)
	return app
}

func TestCheckoutCreateReturnsTokenView(t *testing.T) {
	svc := &fakeCheckoutSvc{createFn: func(_ context.Context, publicID string) (*domain.CheckoutSessionView, error) {
		if publicID != "pl_abc" {
			t.Fatalf("public id = %q", publicID)
		}
		return &domain.CheckoutSessionView{ID: uuid.New(), Token: "cs_tok", Status: "open", AmountMinor: 5000, Currency: "THB", AllowedMethods: []string{"card", "promptpay"}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions", strings.NewReader(`{"link":"pl_abc"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 201 (%s)", resp.StatusCode, b)
	}
	var env domain.APIResponse
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &env)
	data, _ := env.Data.(map[string]any)
	if data["session_token"] != "cs_tok" {
		t.Fatalf("session_token = %v want cs_tok", data["session_token"])
	}
}

func TestCheckoutCreateMissingLinkIs400(t *testing.T) {
	app := newCheckoutApp(&fakeCheckoutSvc{})
	req := httptest.NewRequest("POST", "/v1/checkout/sessions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d want 400", resp.StatusCode)
	}
}

func TestCheckoutPayForwardsToken(t *testing.T) {
	var gotToken string
	svc := &fakeCheckoutSvc{payFn: func(_ context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
		gotToken = token
		return &domain.CheckoutSessionView{Status: "requires_action", QRPayload: "PP", SelectedMethod: req.Method, AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/pay", strings.NewReader(`{"method":"promptpay"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 200 (%s)", resp.StatusCode, b)
	}
	if gotToken != "cs_tok" {
		t.Fatalf("token = %q want cs_tok", gotToken)
	}
}

func TestCheckoutConfirmMockApprove(t *testing.T) {
	var gotToken string
	var gotApprove bool
	svc := &fakeCheckoutSvc{confirmMockFn: func(_ context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
		gotToken, gotApprove = token, approve
		return &domain.CheckoutSessionView{ID: uuid.New(), Status: "paid", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/confirm-mock", strings.NewReader(`{"approve":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 200 (%s)", resp.StatusCode, b)
	}
	if gotToken != "cs_tok" || !gotApprove {
		t.Fatalf("token/approve = %q/%v want cs_tok/true", gotToken, gotApprove)
	}
}

func TestCheckoutConfirmMockDecline(t *testing.T) {
	var gotApprove = true
	svc := &fakeCheckoutSvc{confirmMockFn: func(_ context.Context, _ string, approve bool) (*domain.CheckoutSessionView, error) {
		gotApprove = approve
		return &domain.CheckoutSessionView{Status: "failed", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/confirm-mock", strings.NewReader(`{"approve":false}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
	if gotApprove {
		t.Fatalf("approve = true want false")
	}
}

func TestCheckoutGetReturnsView(t *testing.T) {
	svc := &fakeCheckoutSvc{getFn: func(_ context.Context, token string) (*domain.CheckoutSessionView, error) {
		return &domain.CheckoutSessionView{ID: uuid.New(), Status: "paid", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)
	resp, _ := app.Test(httptest.NewRequest("GET", "/v1/checkout/sessions/cs_tok", nil))
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
}
