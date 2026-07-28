package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MerchantProfile is the self-view returned by GET /v1/me. It is the merchant
// profile plus its configured webhook endpoint. Secrets (api_key_hash,
// webhook_secret_hash) are never included.
type MerchantProfile struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	MCC                string    `json:"mcc,omitempty"`
	SettlementCurrency string    `json:"settlement_currency"`
	Status             string    `json:"status"`
	WebhookURL         string    `json:"webhook_url,omitempty"`
}

// StatsByStatus carries the per-status payment counts for the KPI tiles.
type StatsByStatus struct {
	Authorized int64 `json:"authorized"`
	Captured   int64 `json:"captured"`
	Refunded   int64 `json:"refunded"`
	Failed     int64 `json:"failed"`
}

// MerchantStats is the KPI rollup returned by GET /v1/stats over a window.
// VolumeMinor is captured (net-of-refund) money in minor units (satang). Ratios
// are fractions in [0,1]; the frontend renders them as percentages.
type MerchantStats struct {
	Count           int64         `json:"count"`
	VolumeMinor     int64         `json:"volumeMinor"`
	ByStatus        StatsByStatus `json:"byStatus"`
	SuccessRate     float64       `json:"successRate"`
	RefundRatio     float64       `json:"refundRatio"`
	ChargebackRatio float64       `json:"chargebackRatio"`
	From            time.Time     `json:"from"`
	To              time.Time     `json:"to"`
}

// StatsSeriesPoint is one day of the daily series returned by
// GET /v1/stats/series. VolumeMinor and Count are zero for calendar days with
// no payments (the service zero-fills gaps left by the DB, which only returns
// days that have rows). Date is UTC, formatted YYYY-MM-DD.
type StatsSeriesPoint struct {
	Date        string `json:"date"`
	VolumeMinor int64  `json:"volume_minor"`
	Count       int64  `json:"count"`
}

// StatsSeriesTotals sums the series over the requested window.
type StatsSeriesTotals struct {
	VolumeMinor int64 `json:"volume_minor"`
	Count       int64 `json:"count"`
}

// StatsSeriesTrend compares the requested window against the immediately
// preceding window of equal length. Values are fractions (e.g. 0.124 = +12.4%);
// both are 0 when the previous window has no activity to compare against.
type StatsSeriesTrend struct {
	VolumePct float64 `json:"volume_pct"`
	CountPct  float64 `json:"count_pct"`
}

// StatsSeries is the payload for GET /v1/stats/series: a zero-filled daily
// series for the dashboard sparkline plus window totals and the trend vs the
// prior equal-length window.
type StatsSeries struct {
	Days   int                `json:"days"`
	Series []StatsSeriesPoint `json:"series"`
	Totals StatsSeriesTotals  `json:"totals"`
	Trend  StatsSeriesTrend   `json:"trend"`
}

// Settlement is one payout row returned by GET /v1/settlements. Amounts are
// minor units. It mirrors the settlement worker's payout output.
type Settlement struct {
	ID            uuid.UUID `json:"id"`
	Currency      string    `json:"currency"`
	GrossMinor    int64     `json:"gross_minor"`
	RefundedMinor int64     `json:"refunded_minor"`
	FeeMinor      int64     `json:"fee_minor"`
	NetMinor      int64     `json:"net_minor"`
	PaymentCount  int32     `json:"payment_count"`
	Status        string    `json:"status"`
	PeriodStart   time.Time `json:"period_start"`
	PeriodEnd     time.Time `json:"period_end"`
	CreatedAt     time.Time `json:"created_at"`
}

// RotatedKey is returned exactly once by POST /v1/me/rotate-key. The old key is
// invalidated the moment this responds.
type RotatedKey struct {
	APIKey string `json:"api_key"` // shown once; store it securely
}

// SetWebhookRequest is the body of PUT /v1/me/webhook.
type SetWebhookRequest struct {
	URL string `json:"url" validate:"required,url,max=2048"`
}

// WebhookConfig is returned by PUT /v1/me/webhook. The signing secret is
// returned exactly once; it is persisted only as a hash plus an
// envelope-encrypted blob, and the outbound worker signs this merchant's
// deliveries with it, so the merchant verifies deliveries against this value.
type WebhookConfig struct {
	WebhookURL    string `json:"webhook_url"`
	SigningSecret string `json:"signing_secret"` // shown once; store it securely
}

// PlatformStats is the platform-wide KRI rollup returned by GET
// /v1/admin/stats. VolumeMinor is captured (net-of-refund) minor units.
type PlatformStats struct {
	MerchantCount   int64         `json:"merchantCount"`
	Count           int64         `json:"count"`
	VolumeMinor     int64         `json:"volumeMinor"`
	ByStatus        StatsByStatus `json:"byStatus"`
	SuccessRate     float64       `json:"successRate"`
	RefundRatio     float64       `json:"refundRatio"`
	ChargebackRatio float64       `json:"chargebackRatio"`
}

// Transaction is one row of the unified merchant activity feed returned by GET
// /v1/transactions: a UNION of card payments, PromptPay (QR) payments, and
// paid wallet checkout sessions, newest first. Source identifies which table
// the row came from ("card"|"promptpay"|"wallet"); Method is the specific
// payment method (e.g. "card", "promptpay_dynamic", "truemoney"). Amounts are
// minor units (no decimal conversion — all three source tables already store
// amount_minor BIGINT).
type Transaction struct {
	ID          uuid.UUID `json:"id"`
	Source      string    `json:"source"`
	Method      string    `json:"method"`
	AmountMinor int64     `json:"amount_minor"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Reference   string    `json:"reference"`
	CreatedAt   time.Time `json:"created_at"`
}

// AuditEntry is one immutable audit-log row returned by GET
// /v1/admin/audit-log.
type AuditEntry struct {
	ID         uuid.UUID       `json:"id"`
	Actor      string          `json:"actor"`
	Action     string          `json:"action"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}
