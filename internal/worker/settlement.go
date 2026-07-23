package worker

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// SettlementWorker aggregates captured (net of refunded) payments per
// merchant/currency over a rolling window into a single payout row per run. A
// flat fee in basis points is deducted to produce the net payout amount.
type SettlementWorker struct {
	repo     repository.Querier
	interval time.Duration
	window   time.Duration
	feeBps   int64 // fee in basis points (e.g. 250 = 2.50%)
	log      zerolog.Logger
}

// NewSettlementWorker wires the settlement worker. A typical production cadence
// is daily; here interval/window are configurable for tests and local runs.
func NewSettlementWorker(repo repository.Querier, interval, window time.Duration, feeBps int64, log zerolog.Logger) *SettlementWorker {
	return &SettlementWorker{repo: repo, interval: interval, window: window, feeBps: feeBps, log: log}
}

func (w *SettlementWorker) Name() string            { return "settlement" }
func (w *SettlementWorker) Interval() time.Duration { return w.interval }

// Run aggregates the current window and writes one payout per merchant/currency
// that had captured volume.
func (w *SettlementWorker) Run(ctx context.Context) error {
	if w.repo == nil {
		return nil
	}
	end := time.Now().UTC()
	start := end.Add(-w.window)

	rows, err := w.repo.AggregateCapturedForSettlement(ctx, repository.AggregateCapturedForSettlementParams{
		UpdatedAt:   pgtype.Timestamptz{Time: start, Valid: true},
		UpdatedAt_2: pgtype.Timestamptz{Time: end, Valid: true},
	})
	if err != nil {
		return err
	}

	for _, r := range rows {
		gross := r.GrossMinor
		refunded := r.RefundedMinor
		settleable := gross - refunded
		if settleable <= 0 {
			continue
		}
		fee := settleable * w.feeBps / 10000
		net := settleable - fee

		// payment_count is a positive aggregate; clamp to int32 to avoid an
		// overflow wrap in the (unreachable in practice) event of >2.1B payments
		// in a single settlement window.
		paymentCount := r.PaymentCount
		if paymentCount > math.MaxInt32 {
			paymentCount = math.MaxInt32
		}

		if _, err := w.repo.CreatePayout(ctx, repository.CreatePayoutParams{
			ID:            pgtype.UUID{Bytes: uuid.New(), Valid: true},
			MerchantID:    r.MerchantID,
			Currency:      r.Currency,
			GrossMinor:    gross,
			RefundedMinor: refunded,
			FeeMinor:      fee,
			NetMinor:      net,
			PaymentCount:  int32(paymentCount), // #nosec G115 -- clamped to math.MaxInt32 above
			Status:        "pending",
			PeriodStart:   pgtype.Timestamptz{Time: start, Valid: true},
			PeriodEnd:     pgtype.Timestamptz{Time: end, Valid: true},
		}); err != nil {
			return err
		}
		w.log.Info().
			Str("merchant_id", uuid.UUID(r.MerchantID.Bytes).String()).
			Str("currency", r.Currency).
			Int64("net_minor", net).
			Int64("fee_minor", fee).
			Msg("payout created")
	}
	return nil
}
