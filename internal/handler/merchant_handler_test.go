package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// fakeMerchantSvc implements service.MerchantService for handler tests. Get
// returns a fixed merchant for any id so a leak would surface as a 200 rather
// than a not-found from the repo.
type fakeMerchantSvc struct{}

func (f *fakeMerchantSvc) Onboard(_ context.Context, req domain.CreateMerchantRequest) (*domain.MerchantCredential, error) {
	return &domain.MerchantCredential{
		Merchant: domain.Merchant{ID: uuid.New(), Name: req.Name, SettlementCurrency: req.SettlementCurrency},
		APIKey:   "sk_live_test",
	}, nil
}

func (f *fakeMerchantSvc) Get(_ context.Context, id uuid.UUID) (*domain.Merchant, error) {
	return &domain.Merchant{ID: id, Name: "Acme", Status: "active", SettlementCurrency: "THB"}, nil
}

func (f *fakeMerchantSvc) ResolveByAPIKeyHash(_ context.Context, _ string) (*domain.Merchant, error) {
	return nil, domain.ErrUnauthorized
}

func (f *fakeMerchantSvc) Profile(_ context.Context, id uuid.UUID) (*domain.MerchantProfile, error) {
	return &domain.MerchantProfile{ID: id, Name: "Acme", Status: "active", SettlementCurrency: "THB"}, nil
}

func (f *fakeMerchantSvc) Stats(_ context.Context, _ uuid.UUID, from, to time.Time) (*domain.MerchantStats, error) {
	return &domain.MerchantStats{From: from, To: to}, nil
}

func (f *fakeMerchantSvc) ListSettlements(_ context.Context, _ uuid.UUID, _ int) ([]*domain.Settlement, error) {
	return nil, nil
}

func (f *fakeMerchantSvc) RotateAPIKey(_ context.Context, _ uuid.UUID) (*domain.RotatedKey, error) {
	return &domain.RotatedKey{APIKey: "sk_live_rotated"}, nil
}

func (f *fakeMerchantSvc) SetWebhook(_ context.Context, _ uuid.UUID, url string) (*domain.WebhookConfig, error) {
	return &domain.WebhookConfig{WebhookURL: url, SigningSecret: "whsec_test"}, nil
}

// withMerchant injects an authenticated merchant id into locals, emulating what
// APIKeyAuth does after a successful key resolution.
func withMerchant(mid uuid.UUID) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(middleware.LocalMerchantID, mid)
		return c.Next()
	}
}

// TestMerchantGetOwnershipReturns404ForOtherMerchant asserts the BOLA fix:
// merchant A (authenticated) requesting merchant B's id gets 404, not the
// profile, and not a 403 (which would leak existence).
func TestMerchantGetOwnershipReturns404ForOtherMerchant(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})

	merchantA := uuid.New()
	merchantB := uuid.New()
	app.Get("/v1/merchants/:id", withMerchant(merchantA), h.Get)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/merchants/"+merchantB.String(), nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("cross-merchant GET -> %d, want 404", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if env.Code != "MERCHANT_NOT_FOUND" {
		t.Fatalf("code=%q want MERCHANT_NOT_FOUND", env.Code)
	}
	if strings.Contains(string(body), "Acme") {
		t.Fatalf("cross-merchant GET leaked the other merchant's profile: %s", body)
	}
}

// TestMerchantGetOwnSucceeds confirms the fix does not break the legitimate case.
func TestMerchantGetOwnSucceeds(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})

	me := uuid.New()
	app.Get("/v1/merchants/:id", withMerchant(me), h.Get)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/merchants/"+me.String(), nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("own GET -> %d, want 200", resp.StatusCode)
	}
}

// TestMerchantGetUnauthenticatedReturns404 asserts that without an authenticated
// merchant in context the handler still returns 404 (not a panic / 500 / leak).
func TestMerchantGetUnauthenticatedReturns404(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/merchants/:id", h.Get) // no auth middleware

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/merchants/"+uuid.NewString(), nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("unauthenticated GET -> %d, want 404", resp.StatusCode)
	}
}

// TestMerchantOnboardValidationEnvelope asserts the shared validation helper is
// applied: a missing required field yields 400 VALIDATION_ERROR with a
// machine-parseable data.errors array (JSON field names, not Go struct paths).
func TestMerchantOnboardValidationEnvelope(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/merchants", h.Onboard)

	// Missing name + missing settlement_currency.
	req := httptest.NewRequest(fiber.MethodPost, "/v1/merchants", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d want 400", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Code string `json:"code"`
		Data struct {
			Errors []FieldError `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if env.Code != "VALIDATION_ERROR" {
		t.Fatalf("code=%q want VALIDATION_ERROR", env.Code)
	}
	if len(env.Data.Errors) == 0 {
		t.Fatalf("expected field errors, got none: %s", body)
	}
	// Must NOT leak the raw validator message with Go struct/tag internals.
	if strings.Contains(string(body), "Error:Field validation") ||
		strings.Contains(string(body), "CreateMerchantRequest.") {
		t.Fatalf("validation response leaked internal validator prose: %s", body)
	}
	// Field names must be the wire (snake_case) names.
	sawSettlement := false
	for _, fe := range env.Data.Errors {
		if fe.Field == "settlement_currency" {
			sawSettlement = true
		}
		if fe.Code == "" || fe.Message == "" {
			t.Fatalf("field error missing code/message: %+v", fe)
		}
	}
	if !sawSettlement {
		t.Fatalf("expected settlement_currency in field errors: %+v", env.Data.Errors)
	}
}
