package worker

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/repository"
)

// ReconciliationWorker compares the gateway's own ledger totals for a window
// against the acquirer's settlement report. Any per-merchant/currency delta is
// recorded in recon_mismatches for ops follow-up. A zero delta is a clean match
// and is not persisted.
type ReconciliationWorker struct {
	repo     repository.Querier
	reporter external.SettlementReporter
	interval time.Duration
	window   time.Duration
	log      zerolog.Logger
}

// NewReconciliationWorker wires the reconciliation worker.
func NewReconciliationWorker(repo repository.Querier, reporter external.SettlementReporter, interval, window time.Duration, log zerolog.Logger) *ReconciliationWorker {
	return &ReconciliationWorker{repo: repo, reporter: reporter, interval: interval, window: window, log: log}
}

func (w *ReconciliationWorker) Name() string            { return "reconciliation" }
func (w *ReconciliationWorker) Interval() time.Duration { return w.interval }

// key uniquely identifies a merchant/currency bucket.
type reconKey struct {
	merchant string
	currency string
}

// Run pulls ledger totals and the acquirer report for the window and flags any
// mismatch. A missing side counts as zero, so acquirer-only or ledger-only rows
// are also surfaced.
func (w *ReconciliationWorker) Run(ctx context.Context) error {
	if w.repo == nil || w.reporter == nil {
		return nil
	}
	end := time.Now().UTC()
	start := end.Add(-w.window)

	ledgerRows, err := w.repo.LedgerCapturedTotals(ctx, repository.LedgerCapturedTotalsParams{
		CreatedAt:   pgtype.Timestamptz{Time: start, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return err
	}

	report, err := w.reporter.SettlementReport(ctx, end)
	if err != nil {
		return err
	}

	// Index both sides by (merchant, currency).
	ledger := make(map[reconKey]int64, len(ledgerRows))
	merchantUUID := make(map[reconKey]pgtype.UUID, len(ledgerRows))
	for _, r := range ledgerRows {
		mid := uuid.UUID(r.MerchantID.Bytes).String()
		k := reconKey{merchant: mid, currency: r.Currency}
		ledger[k] = r.NetMinor
		merchantUUID[k] = r.MerchantID
	}
	acquirer := make(map[reconKey]int64, len(report))
	for _, l := range report {
		k := reconKey{merchant: l.MerchantID, currency: l.Currency}
		acquirer[k] += l.AmountMinor
		if _, ok := merchantUUID[k]; !ok {
			if id, err := uuid.Parse(l.MerchantID); err == nil {
				merchantUUID[k] = pgtype.UUID{Bytes: id, Valid: true}
			}
		}
	}

	// Union of keys.
	seen := make(map[reconKey]bool)
	keys := make([]reconKey, 0, len(ledger)+len(acquirer))
	for k := range ledger {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	for k := range acquirer {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}

	reportDate := pgtype.Date{Time: end.Truncate(24 * time.Hour), Valid: true}
	for _, k := range keys {
		l := ledger[k]
		a := acquirer[k]
		delta := l - a
		if delta == 0 {
			continue // clean match
		}
		if err := w.repo.CreateReconMismatch(ctx, repository.CreateReconMismatchParams{
			ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
			MerchantID:    merchantUUID[k],
			Currency:      k.currency,
			LedgerMinor:   l,
			AcquirerMinor: a,
			DeltaMinor:    delta,
			ReportDate:    reportDate,
		}); err != nil {
			return err
		}
		w.log.Warn().
			Str("merchant_id", k.merchant).
			Str("currency", k.currency).
			Int64("ledger_minor", l).
			Int64("acquirer_minor", a).
			Int64("delta_minor", delta).
			Msg("reconciliation mismatch flagged")
	}
	return nil
}
