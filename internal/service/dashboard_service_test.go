package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// statsRepo is a minimal repository.Querier used to drive Stats/RotateAPIKey/
// SetWebhook/ListSettlements. It embeds the interface so unimplemented methods
// panic loudly if a test path unexpectedly hits them.
type statsRepo struct {
	repository.Querier

	statsRow    repository.MerchantStatsRow
	disputes    int64
	payouts     []repository.Payout
	rotateHash  string
	webhookURL  *string
	webhookHash *string
}

func (r *statsRepo) MerchantStats(_ context.Context, _ repository.MerchantStatsParams) (repository.MerchantStatsRow, error) {
	return r.statsRow, nil
}

func (r *statsRepo) CountDisputesByMerchant(_ context.Context, _ repository.CountDisputesByMerchantParams) (int64, error) {
	return r.disputes, nil
}

func (r *statsRepo) ListPayoutsByMerchant(_ context.Context, _ repository.ListPayoutsByMerchantParams) ([]repository.Payout, error) {
	return r.payouts, nil
}

func (r *statsRepo) RotateMerchantAPIKey(_ context.Context, arg repository.RotateMerchantAPIKeyParams) (repository.Merchant, error) {
	r.rotateHash = arg.ApiKeyHash
	return repository.Merchant{ID: arg.ID, ApiKeyHash: arg.ApiKeyHash}, nil
}

func (r *statsRepo) SetMerchantWebhook(_ context.Context, arg repository.SetMerchantWebhookParams) (repository.Merchant, error) {
	r.webhookURL = arg.WebhookUrl
	r.webhookHash = arg.WebhookSecretHash
	return repository.Merchant{ID: arg.ID, WebhookUrl: arg.WebhookUrl, WebhookSecretHash: arg.WebhookSecretHash}, nil
}

// TestMerchantStatsAggregation is table-driven over the KPI ratio derivation:
// success rate, refund ratio and chargeback ratio are all guarded against a
// zero denominator and computed from the by-status counts.
func TestMerchantStatsAggregation(t *testing.T) {
	tests := []struct {
		name            string
		row             repository.MerchantStatsRow
		disputes        int64
		wantVolume      int64
		wantSuccess     float64
		wantRefundRatio float64
		wantCBRatio     float64
	}{
		{
			name: "typical mix",
			row: repository.MerchantStatsRow{
				Count: 100, VolumeMinor: 500000,
				AuthorizedCount: 90, CapturedCount: 80, RefundedCount: 10, FailedCount: 10,
			},
			disputes:        5,
			wantVolume:      500000,
			wantSuccess:     0.9,
			wantRefundRatio: 0.1,
			wantCBRatio:     0.05,
		},
		{
			name:            "no payments -> zero ratios, no div-by-zero",
			row:             repository.MerchantStatsRow{},
			disputes:        0,
			wantVolume:      0,
			wantSuccess:     0,
			wantRefundRatio: 0,
			wantCBRatio:     0,
		},
		{
			name: "all captured",
			row: repository.MerchantStatsRow{
				Count: 4, VolumeMinor: 4000,
				AuthorizedCount: 4, CapturedCount: 4, RefundedCount: 0, FailedCount: 0,
			},
			disputes:        0,
			wantVolume:      4000,
			wantSuccess:     1,
			wantRefundRatio: 0,
			wantCBRatio:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &statsRepo{statsRow: tc.row, disputes: tc.disputes}
			svc := NewMerchantService(repo, zerolog.Nop())
			from := time.Now().Add(-24 * time.Hour)
			to := time.Now()

			got, err := svc.Stats(context.Background(), uuid.New(), from, to)
			if err != nil {
				t.Fatalf("Stats: %v", err)
			}
			if got.VolumeMinor != tc.wantVolume {
				t.Errorf("volumeMinor=%d want %d", got.VolumeMinor, tc.wantVolume)
			}
			if got.SuccessRate != tc.wantSuccess {
				t.Errorf("successRate=%v want %v", got.SuccessRate, tc.wantSuccess)
			}
			if got.RefundRatio != tc.wantRefundRatio {
				t.Errorf("refundRatio=%v want %v", got.RefundRatio, tc.wantRefundRatio)
			}
			if got.ChargebackRatio != tc.wantCBRatio {
				t.Errorf("chargebackRatio=%v want %v", got.ChargebackRatio, tc.wantCBRatio)
			}
			// by-status must mirror the row.
			if got.ByStatus.Authorized != tc.row.AuthorizedCount ||
				got.ByStatus.Captured != tc.row.CapturedCount ||
				got.ByStatus.Refunded != tc.row.RefundedCount ||
				got.ByStatus.Failed != tc.row.FailedCount {
				t.Errorf("byStatus mismatch: %+v vs row %+v", got.ByStatus, tc.row)
			}
		})
	}
}

// TestRotateAPIKeyReturnsFreshKeyAndPersistsHashOnly asserts the new key is
// returned once, has the sk_live_ prefix, and only its hash reaches the repo
// (the raw key is never persisted).
func TestRotateAPIKeyReturnsFreshKeyAndPersistsHashOnly(t *testing.T) {
	repo := &statsRepo{}
	svc := NewMerchantService(repo, zerolog.Nop())

	got, err := svc.RotateAPIKey(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("RotateAPIKey: %v", err)
	}
	if len(got.APIKey) < 16 || got.APIKey[:8] != "sk_live_" {
		t.Fatalf("rotated key looks wrong: %q", got.APIKey)
	}
	// The repo must have received the HASH, never the raw key.
	if repo.rotateHash == "" {
		t.Fatal("no hash persisted")
	}
	if repo.rotateHash == got.APIKey {
		t.Fatal("raw key was persisted instead of its hash")
	}
	if len(repo.rotateHash) != 64 {
		t.Fatalf("persisted value is not a sha256 hex hash: len=%d", len(repo.rotateHash))
	}
}

// TestSetWebhookRotatesSecretAndStoresHashOnly asserts the signing secret is
// returned once and only its hash is persisted alongside the URL.
func TestSetWebhookRotatesSecretAndStoresHashOnly(t *testing.T) {
	repo := &statsRepo{}
	svc := NewMerchantService(repo, zerolog.Nop())

	cfg, err := svc.SetWebhook(context.Background(), uuid.New(), "https://merchant.example.com/hook")
	if err != nil {
		t.Fatalf("SetWebhook: %v", err)
	}
	if cfg.SigningSecret == "" || cfg.SigningSecret[:6] != "whsec_" {
		t.Fatalf("signing secret looks wrong: %q", cfg.SigningSecret)
	}
	if cfg.WebhookURL != "https://merchant.example.com/hook" {
		t.Fatalf("url mismatch: %q", cfg.WebhookURL)
	}
	if repo.webhookURL == nil || *repo.webhookURL != cfg.WebhookURL {
		t.Fatalf("url not persisted: %v", repo.webhookURL)
	}
	if repo.webhookHash == nil || *repo.webhookHash == cfg.SigningSecret {
		t.Fatal("raw signing secret was persisted instead of its hash")
	}
	if len(*repo.webhookHash) != 64 {
		t.Fatalf("persisted secret is not a sha256 hex hash: len=%d", len(*repo.webhookHash))
	}
}
