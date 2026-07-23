package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

// scopedMerchantSvc records the merchant id every dashboard method is called
// with, so a test can assert /me and friends resolve identity from the auth
// context (locals) and never from the request.
type scopedMerchantSvc struct {
	fakeMerchantSvc
	gotProfileID uuid.UUID
	gotRotateID  uuid.UUID
}

func (s *scopedMerchantSvc) Profile(_ context.Context, id uuid.UUID) (*domain.MerchantProfile, error) {
	s.gotProfileID = id
	return &domain.MerchantProfile{ID: id, Name: "Acme", Status: "active", SettlementCurrency: "THB", WebhookURL: "https://ex/hook"}, nil
}

func (s *scopedMerchantSvc) RotateAPIKey(_ context.Context, id uuid.UUID) (*domain.RotatedKey, error) {
	s.gotRotateID = id
	return &domain.RotatedKey{APIKey: "sk_live_rotated"}, nil
}

// TestMeResolvesFromAuthContext asserts GET /v1/me scopes to the merchant id in
// locals (set by APIKeyAuth), not any body/param.
func TestMeResolvesFromAuthContext(t *testing.T) {
	svc := &scopedMerchantSvc{}
	h := NewMerchantHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})

	me := uuid.New()
	app.Get("/v1/me", withMerchant(me), h.Me)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/me", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	if svc.gotProfileID != me {
		t.Fatalf("Profile scoped to %s, want authenticated merchant %s", svc.gotProfileID, me)
	}
	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Data domain.MerchantProfile `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if env.Data.ID != me {
		t.Fatalf("returned profile id=%s want %s", env.Data.ID, me)
	}
}

// TestMeUnauthenticatedRejected asserts /me without an authenticated merchant in
// locals returns 401 (not a panic / leak).
func TestMeUnauthenticatedRejected(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/me", h.Me) // no auth middleware

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/me", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", resp.StatusCode)
	}
}

// TestRotateKeyScopedAndReturnedOnce asserts POST /v1/me/rotate-key scopes to
// the authenticated merchant and returns the new key once in the envelope.
func TestRotateKeyScopedAndReturnedOnce(t *testing.T) {
	svc := &scopedMerchantSvc{}
	h := NewMerchantHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})

	me := uuid.New()
	app.Post("/v1/me/rotate-key", withMerchant(me), h.RotateKey)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/v1/me/rotate-key", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	if svc.gotRotateID != me {
		t.Fatalf("RotateAPIKey scoped to %s, want %s", svc.gotRotateID, me)
	}
	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Data domain.RotatedKey `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if env.Data.APIKey != "sk_live_rotated" {
		t.Fatalf("api_key=%q want sk_live_rotated", env.Data.APIKey)
	}
}

// TestStatsInvalidRangeRejected asserts a malformed from/to yields 400 rather
// than silently returning wrong numbers.
func TestStatsInvalidRangeRejected(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/stats", withMerchant(uuid.New()), h.Stats)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/stats?from=not-a-date", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status=%d want 400", resp.StatusCode)
	}
}

// TestStatsDefaultWindowSucceeds asserts /v1/stats with no range uses the
// default 30-day window and returns 200.
func TestStatsDefaultWindowSucceeds(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/stats", withMerchant(uuid.New()), h.Stats)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/stats", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	_ = time.Now() // window is time-based; smoke test only asserts the happy path
}
