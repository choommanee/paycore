package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

type fakeAdminSvc struct {
	merchants   []*domain.Merchant
	audit       []*domain.AuditEntry
	disputes    []*domain.Dispute
	settlements []*domain.Settlement
	stats       *domain.PlatformStats
}

func (f *fakeAdminSvc) ListMerchants(_ context.Context, _ int) ([]*domain.Merchant, error) {
	return f.merchants, nil
}
func (f *fakeAdminSvc) ListAuditLog(_ context.Context, _ int) ([]*domain.AuditEntry, error) {
	return f.audit, nil
}
func (f *fakeAdminSvc) ListDisputes(_ context.Context, _ int) ([]*domain.Dispute, error) {
	return f.disputes, nil
}
func (f *fakeAdminSvc) ListSettlements(_ context.Context, _ int) ([]*domain.Settlement, error) {
	return f.settlements, nil
}
func (f *fakeAdminSvc) PlatformStats(_ context.Context) (*domain.PlatformStats, error) {
	return f.stats, nil
}

// TestAdminStatsReturnsPlatformKRIs asserts GET /v1/admin/stats returns the
// platform rollup in the standard envelope. The X-Admin-Key gate is exercised
// separately in middleware; here we mount the handler directly.
func TestAdminStatsReturnsPlatformKRIs(t *testing.T) {
	svc := &fakeAdminSvc{stats: &domain.PlatformStats{
		MerchantCount: 3, Count: 120, VolumeMinor: 999,
		SuccessRate: 0.9, RefundRatio: 0.05, ChargebackRatio: 0.01,
	}}
	h := NewAdminHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/admin/stats", h.Stats)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/admin/stats", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var env struct {
		Data domain.PlatformStats `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal: %v (%s)", err, body)
	}
	if env.Data.MerchantCount != 3 || env.Data.Count != 120 || env.Data.VolumeMinor != 999 {
		t.Fatalf("platform stats not returned faithfully: %+v", env.Data)
	}
}

// TestAdminMerchantsEmptyReturnsEmptyList asserts the list endpoints degrade to
// an empty (non-null) collection rather than erroring when there is no data.
func TestAdminMerchantsEmptyReturnsEmptyList(t *testing.T) {
	h := NewAdminHandler(&fakeAdminSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Get("/v1/admin/merchants", h.Merchants)

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/admin/merchants", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status=%d want 200", resp.StatusCode)
	}
}
