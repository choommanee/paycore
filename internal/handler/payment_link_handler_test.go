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

type fakeLinkSvc struct{ lastMerchant uuid.UUID }

func (f *fakeLinkSvc) Create(_ context.Context, merchantID uuid.UUID, _ *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	f.lastMerchant = merchantID
	return &domain.PaymentLink{ID: uuid.New(), MerchantID: merchantID, PublicID: "pl_abc", Title: req.Title, AmountMinor: req.AmountMinor, Currency: "THB", Status: "active", URL: "https://pay.example/pay/pl_abc", AllowedMethods: []string{}}, nil
}
func (f *fakeLinkSvc) List(_ context.Context, merchantID uuid.UUID, _, _ int32) ([]*domain.PaymentLink, error) {
	return []*domain.PaymentLink{{ID: uuid.New(), MerchantID: merchantID, Title: "A", AllowedMethods: []string{}}}, nil
}
func (f *fakeLinkSvc) Get(_ context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{ID: id, MerchantID: merchantID, Title: "A", AllowedMethods: []string{}}, nil
}
func (f *fakeLinkSvc) Disable(_ context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{ID: id, MerchantID: merchantID, Status: "disabled", AllowedMethods: []string{}}, nil
}

func TestCreateLinkUsesAuthedMerchantNotBody(t *testing.T) {
	svc := &fakeLinkSvc{}
	h := NewPaymentLinkHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	authed := uuid.New()
	app.Post("/v1/payment-links", withMerchant(authed), h.Create)

	// body tries to inject a different merchant_id — must be ignored.
	body := `{"title":"Coffee","amount_minor":5000,"merchant_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest("POST", "/v1/payment-links", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d want 201 (%s)", resp.StatusCode, b)
	}
	if svc.lastMerchant != authed {
		t.Fatalf("service got merchant %v, want authed %v (body must not override)", svc.lastMerchant, authed)
	}
}

func TestCreateLinkValidationRejectsZeroAmount(t *testing.T) {
	h := NewPaymentLinkHandler(&fakeLinkSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/payment-links", withMerchant(uuid.New()), h.Create)

	req := httptest.NewRequest("POST", "/v1/payment-links", strings.NewReader(`{"title":"X","amount_minor":0}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("zero amount -> %d want 400", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	_ = json.Unmarshal(body, &env)
	if env.Code != "VALIDATION_ERROR" {
		t.Fatalf("code=%q want VALIDATION_ERROR", env.Code)
	}
}
