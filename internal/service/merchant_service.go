package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/repository"
)

// MerchantService onboards and looks up merchants and resolves API-key auth. It
// also backs the self-service Merchant Dashboard (/v1/me and /v1/stats etc.),
// where every method is scoped to the authenticated merchant id resolved from
// the API key — never a value trusted from the request body.
type MerchantService interface {
	Onboard(ctx context.Context, req domain.CreateMerchantRequest) (*domain.MerchantCredential, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Merchant, error)
	// ResolveByAPIKeyHash satisfies middleware.MerchantResolver.
	ResolveByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Merchant, error)

	// Profile returns the self-view for GET /v1/me (profile + webhook URL).
	Profile(ctx context.Context, id uuid.UUID) (*domain.MerchantProfile, error)
	// Stats aggregates KPIs from payments over [from, to] for a merchant.
	Stats(ctx context.Context, id uuid.UUID, from, to time.Time) (*domain.MerchantStats, error)
	// StatsSeries returns a zero-filled daily series for the last `days` UTC
	// calendar days (clamped to [1, 90]) plus totals and the trend vs the
	// immediately preceding window of equal length, for the dashboard sparkline.
	StatsSeries(ctx context.Context, id uuid.UUID, days int) (*domain.StatsSeries, error)
	// ListSettlements returns the merchant's payout rows, most recent first.
	ListSettlements(ctx context.Context, id uuid.UUID, limit int) ([]*domain.Settlement, error)
	// RotateAPIKey issues a new API key (returned once) and invalidates the old.
	RotateAPIKey(ctx context.Context, id uuid.UUID) (*domain.RotatedKey, error)
	// SetWebhook sets the merchant webhook URL and rotates its signing secret,
	// returning the raw secret exactly once (only its hash is persisted).
	SetWebhook(ctx context.Context, id uuid.UUID, url string) (*domain.WebhookConfig, error)
}

type merchantService struct {
	repo repository.Querier
	log  zerolog.Logger
}

// NewMerchantService wires the merchant service.
func NewMerchantService(repo repository.Querier, log zerolog.Logger) MerchantService {
	return &merchantService{repo: repo, log: log.With().Str("service", "merchant").Logger()}
}

// Onboard creates a merchant, generating an API key that is returned exactly
// once. Only the SHA-256 hash of the key is persisted; the raw key is never
// stored or logged.
func (s *merchantService) Onboard(ctx context.Context, req domain.CreateMerchantRequest) (*domain.MerchantCredential, error) {
	if s.repo == nil {
		return nil, errors.New("merchant repository not configured")
	}
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	row, err := s.repo.CreateMerchant(ctx, repository.CreateMerchantParams{
		ID:                 toPgUUID(uuid.New()),
		Name:               req.Name,
		Mcc:                strPtr(req.MCC),
		SettlementCurrency: req.SettlementCurrency,
		ApiKeyHash:         middleware.HashAPIKey(rawKey),
	})
	if err != nil {
		return nil, err
	}

	m := toDomainMerchant(row)
	return &domain.MerchantCredential{Merchant: *m, APIKey: rawKey}, nil
}

// Get returns a merchant by id.
func (s *merchantService) Get(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	row, err := s.repo.GetMerchant(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return toDomainMerchant(row), nil
}

// ResolveByAPIKeyHash resolves the active merchant for an API-key hash, or
// ErrUnauthorized when there is no match.
func (s *merchantService) ResolveByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Merchant, error) {
	if s.repo == nil {
		return nil, domain.ErrUnauthorized
	}
	row, err := s.repo.GetMerchantByAPIKeyHash(ctx, apiKeyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}
	return toDomainMerchant(row), nil
}

// Profile returns the merchant self-view for GET /v1/me: profile fields plus the
// configured webhook URL. Secrets are never exposed.
func (s *merchantService) Profile(ctx context.Context, id uuid.UUID) (*domain.MerchantProfile, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	row, err := s.repo.GetMerchant(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return &domain.MerchantProfile{
		ID:                 pgUUIDToUUID(row.ID),
		Name:               row.Name,
		MCC:                ptrStr(row.Mcc),
		SettlementCurrency: row.SettlementCurrency,
		Status:             row.Status,
		WebhookURL:         ptrStr(row.WebhookUrl),
	}, nil
}

// Stats aggregates a merchant's payment KPIs over the inclusive [from, to]
// window. Volume is captured (net-of-refund) minor units; success/refund ratios
// are fractions in [0,1]. The chargeback ratio counts disputes over the same
// window against total payment count.
func (s *merchantService) Stats(ctx context.Context, id uuid.UUID, from, to time.Time) (*domain.MerchantStats, error) {
	if s.repo == nil {
		return &domain.MerchantStats{From: from, To: to}, nil
	}
	row, err := s.repo.MerchantStats(ctx, repository.MerchantStatsParams{
		MerchantID:  toPgUUID(id),
		CreatedAt:   pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	disputes, err := s.repo.CountDisputesByMerchant(ctx, repository.CountDisputesByMerchantParams{
		MerchantID:  toPgUUID(id),
		CreatedAt:   pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &domain.MerchantStats{
		Count:       row.Count,
		VolumeMinor: row.VolumeMinor,
		ByStatus: domain.StatsByStatus{
			Authorized: row.AuthorizedCount,
			Captured:   row.CapturedCount,
			Refunded:   row.RefundedCount,
			Failed:     row.FailedCount,
		},
		SuccessRate:     ratio(row.AuthorizedCount, row.Count),
		RefundRatio:     ratio(row.RefundedCount, row.Count),
		ChargebackRatio: ratio(disputes, row.Count),
		From:            from,
		To:              to,
	}, nil
}

// StatsSeries returns a zero-filled daily series for the last `days` UTC
// calendar days (ending today, inclusive) plus totals and the trend vs the
// immediately preceding window of equal length. Volume uses the same captured
// (net-of-refund) definition as Stats. Trend fractions are 0 when the previous
// window has no activity to compare against (divide-by-zero guard).
func (s *merchantService) StatsSeries(ctx context.Context, id uuid.UUID, days int) (*domain.StatsSeries, error) {
	days = clampDays(days)
	if s.repo == nil {
		return &domain.StatsSeries{Days: days}, nil
	}

	// Half-open window ending at the start of tomorrow (UTC) so "today" is
	// fully included; from is `days` calendar days back. The previous window is
	// the equal-length window immediately before it.
	to := truncateToUTCDay(time.Now().UTC()).AddDate(0, 0, 1)
	from := to.AddDate(0, 0, -days)
	prevTo := from
	prevFrom := prevTo.AddDate(0, 0, -days)

	rows, err := s.repo.StatsSeriesByDay(ctx, repository.StatsSeriesByDayParams{
		MerchantID:  toPgUUID(id),
		CreatedAt:   pgtype.Timestamptz{Time: from, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: to, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	prevRows, err := s.repo.StatsSeriesByDay(ctx, repository.StatsSeriesByDayParams{
		MerchantID:  toPgUUID(id),
		CreatedAt:   pgtype.Timestamptz{Time: prevFrom, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: prevTo, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	series, totals := zeroFillSeries(from, days, rows)
	prevTotals := sumSeriesRows(prevRows)

	return &domain.StatsSeries{
		Days:   days,
		Series: series,
		Totals: totals,
		Trend: domain.StatsSeriesTrend{
			VolumePct: pctChange(totals.VolumeMinor, prevTotals.VolumeMinor),
			CountPct:  pctChange(totals.Count, prevTotals.Count),
		},
	}, nil
}

// ListSettlements returns the merchant's payout rows, most recent first.
func (s *merchantService) ListSettlements(ctx context.Context, id uuid.UUID, limit int) ([]*domain.Settlement, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListPayoutsByMerchant(ctx, repository.ListPayoutsByMerchantParams{
		MerchantID: toPgUUID(id),
		Limit:      clampLimit(limit),
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Settlement, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainSettlement(r))
	}
	return out, nil
}

// RotateAPIKey generates a fresh API key, persists only its hash (invalidating
// the previous key), and returns the raw key exactly once.
func (s *merchantService) RotateAPIKey(ctx context.Context, id uuid.UUID) (*domain.RotatedKey, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	if _, err := s.repo.RotateMerchantAPIKey(ctx, repository.RotateMerchantAPIKeyParams{
		ID:         toPgUUID(id),
		ApiKeyHash: middleware.HashAPIKey(rawKey),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return &domain.RotatedKey{APIKey: rawKey}, nil
}

// SetWebhook stores the merchant's webhook URL and rotates its delivery signing
// secret. Only the SHA-256 hash of the secret is persisted; the raw secret is
// returned exactly once for the merchant to verify inbound deliveries against.
func (s *merchantService) SetWebhook(ctx context.Context, id uuid.UUID, url string) (*domain.WebhookConfig, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	secret, err := generateSigningSecret()
	if err != nil {
		return nil, err
	}
	secretHash := middleware.HashAPIKey(secret)
	if _, err := s.repo.SetMerchantWebhook(ctx, repository.SetMerchantWebhookParams{
		ID:                toPgUUID(id),
		WebhookUrl:        strPtr(url),
		WebhookSecretHash: strPtr(secretHash),
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return &domain.WebhookConfig{WebhookURL: url, SigningSecret: secret}, nil
}

// generateAPIKey returns a cryptographically random, URL-safe API key with a
// recognizable prefix.
func generateAPIKey() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "sk_live_" + hex.EncodeToString(buf), nil
}

// generateSigningSecret returns a cryptographically random webhook signing
// secret with a recognizable prefix.
func generateSigningSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "whsec_" + hex.EncodeToString(buf), nil
}

// ratio returns num/den as a fraction in [0,1], guarding division by zero.
func ratio(num, den int64) float64 {
	if den <= 0 {
		return 0
	}
	return float64(num) / float64(den)
}

// clampLimit bounds a caller-supplied list limit into a sane range.
func clampLimit(limit int) int32 {
	if limit <= 0 {
		return 50
	}
	if limit > 500 {
		return 500
	}
	return int32(limit)
}

// clampDays bounds a caller-supplied /v1/stats/series window into [1, 90],
// defaulting to 30 when unset or non-positive.
func clampDays(days int) int {
	if days <= 0 {
		return 30
	}
	if days > 90 {
		return 90
	}
	return days
}

// truncateToUTCDay zeroes the time-of-day component, in UTC.
func truncateToUTCDay(t time.Time) time.Time {
	t = t.UTC()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// pctChange returns (cur-prev)/prev as a fraction, guarding division by zero
// (returns 0 when prev is 0, matching the dashboard's "no prior data" case).
func pctChange(cur, prev int64) float64 {
	if prev == 0 {
		return 0
	}
	return float64(cur-prev) / float64(prev)
}

// zeroFillSeries builds one point per UTC calendar day in [from, from+days),
// ascending, filling days absent from rows with zero volume/count. It also
// returns the totals summed across the full window.
func zeroFillSeries(from time.Time, days int, rows []repository.StatsSeriesByDayRow) ([]domain.StatsSeriesPoint, domain.StatsSeriesTotals) {
	byDay := make(map[string]repository.StatsSeriesByDayRow, len(rows))
	for _, r := range rows {
		byDay[r.Day.Time.Format("2006-01-02")] = r
	}

	start := truncateToUTCDay(from)
	series := make([]domain.StatsSeriesPoint, days)
	var totals domain.StatsSeriesTotals
	for i := 0; i < days; i++ {
		key := start.AddDate(0, 0, i).Format("2006-01-02")
		point := domain.StatsSeriesPoint{Date: key}
		if r, ok := byDay[key]; ok {
			point.VolumeMinor = r.VolumeMinor
			point.Count = r.Count
		}
		series[i] = point
		totals.VolumeMinor += point.VolumeMinor
		totals.Count += point.Count
	}
	return series, totals
}

// sumSeriesRows totals raw (non-zero-filled) series rows, used for the
// previous-window comparison where per-day granularity isn't needed.
func sumSeriesRows(rows []repository.StatsSeriesByDayRow) domain.StatsSeriesTotals {
	var totals domain.StatsSeriesTotals
	for _, r := range rows {
		totals.VolumeMinor += r.VolumeMinor
		totals.Count += r.Count
	}
	return totals
}

func toDomainSettlement(r repository.Payout) *domain.Settlement {
	return &domain.Settlement{
		ID:            pgUUIDToUUID(r.ID),
		Currency:      r.Currency,
		GrossMinor:    r.GrossMinor,
		RefundedMinor: r.RefundedMinor,
		FeeMinor:      r.FeeMinor,
		NetMinor:      r.NetMinor,
		PaymentCount:  r.PaymentCount,
		Status:        r.Status,
		PeriodStart:   r.PeriodStart.Time,
		PeriodEnd:     r.PeriodEnd.Time,
		CreatedAt:     r.CreatedAt.Time,
	}
}

func toDomainMerchant(r repository.Merchant) *domain.Merchant {
	return &domain.Merchant{
		ID:                 pgUUIDToUUID(r.ID),
		Name:               r.Name,
		Status:             r.Status,
		MCC:                ptrStr(r.Mcc),
		SettlementCurrency: r.SettlementCurrency,
		CreatedAt:          r.CreatedAt.Time,
		UpdatedAt:          r.UpdatedAt.Time,
	}
}
