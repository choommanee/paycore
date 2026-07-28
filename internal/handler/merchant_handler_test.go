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

func (f *fakeMerchantSvc) StatsSeries(_ context.Context, _ uuid.UUID, days int) (*domain.StatsSeries, error) {
	return &domain.StatsSeries{Days: days}, nil
}

func (f *fakeMerchantSvc) ListSettlements(_ context.Context, _ uuid.UUID, _ int) ([]*domain.Settlement, error) {
	return nil, nil
}

func (f *fakeMerchantSvc) ListTransactions(_ context.Context, _ uuid.UUID, _, _ int32) ([]*domain.Transaction, error) {
	return []*domain.Transaction{
		{ID: uuid.New(), Source: "card", Method: "card", AmountMinor: 12900, Currency: "THB", Status: "captured", Reference: "REF-1", CreatedAt: time.Now().UTC()},
		{ID: uuid.New(), Source: "promptpay", Method: "promptpay_dynamic", AmountMinor: 5000, Currency: "THB", Status: "paid", CreatedAt: time.Now().UTC()},
	}, nil
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

// spyMerchantSvc wraps fakeMerchantSvc and records the `days` argument
// StatsSeries was called with, so a test can assert the handler's clamping
// logic reaches the service correctly.
type spyMerchantSvc struct {
	fakeMerchantSvc
	gotDays int
}

func (s *spyMerchantSvc) StatsSeries(ctx context.Context, id uuid.UUID, days int) (*domain.StatsSeries, error) {
	s.gotDays = days
	return s.fakeMerchantSvc.StatsSeries(ctx, id, days)
}

// TestStatsSeriesReturnsEnvelope asserts GET /v1/stats/series returns 200 with
// the success envelope wrapping the domain.StatsSeries shape (days, series,
// totals, trend).
func TestStatsSeriesReturnsEnvelope(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/stats/series", withMerchant(uuid.New()), h.StatsSeries)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/stats/series", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Success bool `json:"success"`
		Data    struct {
			Days   int                       `json:"days"`
			Series []domain.StatsSeriesPoint `json:"series"`
			Totals domain.StatsSeriesTotals  `json:"totals"`
			Trend  domain.StatsSeriesTrend   `json:"trend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if !env.Success {
		t.Fatalf("success=false: %s", body)
	}
	if env.Data.Days != 30 {
		t.Fatalf("data.days=%d want 30 (default)", env.Data.Days)
	}
}

// TestStatsSeriesUnauthenticatedReturns401 mirrors the other /v1/stats-style
// handlers: no authenticated merchant in context -> 401, not a panic.
func TestStatsSeriesUnauthenticatedReturns401(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/stats/series", h.StatsSeries) // no auth middleware

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/stats/series", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", resp.StatusCode)
	}
}

// TestStatsSeriesDaysClamping asserts the handler clamps ?days to [1, 90] and
// defaults to 30 when absent/invalid, before it ever reaches the service.
func TestStatsSeriesDaysClamping(t *testing.T) {
	cases := []struct {
		name  string
		query string
		want  int
	}{
		{"default", "", 30},
		{"in range", "?days=7", 7},
		{"below min clamps to 1", "?days=0", 1},
		{"negative clamps to 1", "?days=-5", 1},
		{"above max clamps to 90", "?days=365", 90},
		{"non-numeric falls back to default", "?days=abc", 30},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &spyMerchantSvc{}
			h := NewMerchantHandler(svc, zerolog.Nop())
			app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
			app.Get("/v1/stats/series", withMerchant(uuid.New()), h.StatsSeries)

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/stats/series"+tc.query, nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("status=%d want 200", resp.StatusCode)
			}
			if svc.gotDays != tc.want {
				t.Fatalf("days passed to service=%d want %d", svc.gotDays, tc.want)
			}
		})
	}
}

// ---- ListTransactions -------------------------------------------------

// spyTransactionsSvc wraps fakeMerchantSvc and records the limit/offset
// ListTransactions was called with, so a test can assert the handler's
// pagination parsing (the shared `paginate` helper) reaches the service
// correctly.
type spyTransactionsSvc struct {
	fakeMerchantSvc
	gotLimit, gotOffset int32
}

func (s *spyTransactionsSvc) ListTransactions(ctx context.Context, id uuid.UUID, limit, offset int32) ([]*domain.Transaction, error) {
	s.gotLimit, s.gotOffset = limit, offset
	return s.fakeMerchantSvc.ListTransactions(ctx, id, limit, offset)
}

// TestListTransactionsReturnsEnvelope asserts GET /v1/transactions returns 200
// with the success envelope wrapping a list of domain.Transaction, and that
// the wire JSON uses the snake_case field names from the BA spec
// (amount_minor, created_at) rather than Go field names.
func TestListTransactionsReturnsEnvelope(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/transactions", withMerchant(uuid.New()), h.ListTransactions)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/transactions", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Success bool                 `json:"success"`
		Data    []domain.Transaction `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if !env.Success {
		t.Fatalf("success=false: %s", body)
	}
	if len(env.Data) != 2 {
		t.Fatalf("len(data)=%d want 2: %s", len(env.Data), body)
	}
	if env.Data[0].Source != "card" || env.Data[0].AmountMinor != 12900 {
		t.Fatalf("data[0]=%+v want source=card amount_minor=12900", env.Data[0])
	}
	if !strings.Contains(string(body), `"amount_minor"`) || !strings.Contains(string(body), `"created_at"`) {
		t.Fatalf("response missing expected snake_case wire fields: %s", body)
	}
}

// TestListTransactionsUnauthenticatedReturns401 mirrors the other
// /v1/stats-style handlers: no authenticated merchant in context -> 401.
func TestListTransactionsUnauthenticatedReturns401(t *testing.T) {
	h := NewMerchantHandler(&fakeMerchantSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/transactions", h.ListTransactions) // no auth middleware

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/transactions", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status=%d want 401", resp.StatusCode)
	}
}

// TestListTransactionsPagination asserts ?limit/?offset are parsed via the
// shared paginate helper (default 50, clamped to <=200) and reach the service
// untouched.
func TestListTransactionsPagination(t *testing.T) {
	cases := []struct {
		name       string
		query      string
		wantLimit  int32
		wantOffset int32
	}{
		{"defaults", "", 50, 0},
		{"explicit", "?limit=20&offset=40", 20, 40},
		{"limit above max clamps to default", "?limit=500", 50, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := &spyTransactionsSvc{}
			h := NewMerchantHandler(svc, zerolog.Nop())
			app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
			app.Get("/v1/transactions", withMerchant(uuid.New()), h.ListTransactions)

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/transactions"+tc.query, nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != fiber.StatusOK {
				t.Fatalf("status=%d want 200", resp.StatusCode)
			}
			if svc.gotLimit != tc.wantLimit || svc.gotOffset != tc.wantOffset {
				t.Fatalf("limit/offset passed to service=%d/%d want %d/%d", svc.gotLimit, svc.gotOffset, tc.wantLimit, tc.wantOffset)
			}
		})
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
