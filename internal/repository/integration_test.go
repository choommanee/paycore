package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/yourco/payment-gateway/internal/db"
	"github.com/yourco/payment-gateway/internal/repository"
)

// requireDB returns a live pool or skips the test when DATABASE_URL is unset, so
// `go test ./...` stays green without a database while still exercising the full
// repository when one is provided (e.g. in CI with a Postgres service).
func requireDB(t *testing.T) (*pgxPoolWrapper, func()) {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration test")
	}
	ctx := context.Background()
	pool, err := db.NewPool(ctx, db.DefaultPoolConfig(dsn))
	if err != nil {
		t.Fatalf("connect database: %v", err)
	}
	return &pgxPoolWrapper{pool}, pool.Close
}

// pgxPoolWrapper adapts the pool for repository.New (which accepts any DBTX).
type pgxPoolWrapper struct {
	pool interface {
		repository.DBTX
		Close()
	}
}

func toPgUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: id, Valid: true} }

// TestPaymentRepoRoundTrip exercises merchant + payment + status update + audit
// log writes against a real database. Skipped when DATABASE_URL is unset.
func TestPaymentRepoRoundTrip(t *testing.T) {
	w, cleanup := requireDB(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	repo := repository.New(w.pool)

	// Create a merchant.
	merchantID := uuid.New()
	if _, err := repo.CreateMerchant(ctx, repository.CreateMerchantParams{
		ID:                 toPgUUID(merchantID),
		Name:               "itest-merchant",
		SettlementCurrency: "THB",
		ApiKeyHash:         "itest-hash-" + merchantID.String(),
	}); err != nil {
		t.Fatalf("CreateMerchant: %v", err)
	}

	// Create a payment.
	paymentID := uuid.New()
	idem := "itest-" + paymentID.String()
	created, err := repo.CreatePayment(ctx, repository.CreatePaymentParams{
		ID:             toPgUUID(paymentID),
		MerchantID:     toPgUUID(merchantID),
		AmountMinor:    12345,
		Currency:       "THB",
		Status:         "authorized",
		IdempotencyKey: idem,
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if created.AmountMinor != 12345 {
		t.Fatalf("amount round-trip mismatch: %d", created.AmountMinor)
	}

	// Idempotency lookup returns the same row.
	got, err := repo.GetByIdempotencyKey(ctx, repository.GetByIdempotencyKeyParams{
		MerchantID:     toPgUUID(merchantID),
		IdempotencyKey: idem,
	})
	if err != nil {
		t.Fatalf("GetByIdempotencyKey: %v", err)
	}
	if got.ID != created.ID {
		t.Fatalf("idempotency lookup returned different payment")
	}

	// Capture (status update) + ledger entry + audit log.
	if _, err := repo.UpdatePaymentStatus(ctx, repository.UpdatePaymentStatusParams{
		ID:                  toPgUUID(paymentID),
		Status:              "captured",
		CapturedAmountMinor: 12345,
	}); err != nil {
		t.Fatalf("UpdatePaymentStatus: %v", err)
	}
	if err := repo.InsertLedgerEntry(ctx, repository.InsertLedgerEntryParams{
		ID:          toPgUUID(uuid.New()),
		PaymentID:   toPgUUID(paymentID),
		EntryType:   "capture",
		AmountMinor: 12345,
		Currency:    "THB",
	}); err != nil {
		t.Fatalf("InsertLedgerEntry: %v", err)
	}
	if err := repo.WriteAuditLog(ctx, repository.WriteAuditLogParams{
		ID:         toPgUUID(uuid.New()),
		Actor:      merchantID.String(),
		Action:     "payment.captured",
		EntityType: "payment",
		EntityID:   paymentID.String(),
		Metadata:   []byte(`{"captured_amount_minor":12345}`),
	}); err != nil {
		t.Fatalf("WriteAuditLog: %v", err)
	}

	// Confirm the persisted status.
	after, err := repo.GetPayment(ctx, toPgUUID(paymentID))
	if err != nil {
		t.Fatalf("GetPayment: %v", err)
	}
	if after.Status != "captured" || after.CapturedAmountMinor != 12345 {
		t.Fatalf("status not persisted: %+v", after)
	}
}
