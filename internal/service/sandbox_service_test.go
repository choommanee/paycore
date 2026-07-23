package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// The sandbox service reuses the fakeQRRepo from qr_service_test.go and the real
// QRService (newQRService), proving the simulator drives the SAME
// ConfirmFromWebhook validation a bank webhook takes. These two repo methods are
// only needed by the sandbox path so they live here alongside the tests.

func (f *fakeQRRepo) GetQRPayment(_ context.Context, id pgtype.UUID) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[key(id)]
	if !ok {
		return repository.QrPayment{}, pgx.ErrNoRows
	}
	return r, nil
}

func (f *fakeQRRepo) GetQRPaymentWithMerchant(_ context.Context, id pgtype.UUID) (repository.GetQRPaymentWithMerchantRow, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[key(id)]
	if !ok {
		return repository.GetQRPaymentWithMerchantRow{}, pgx.ErrNoRows
	}
	return repository.GetQRPaymentWithMerchantRow{
		ID:             r.ID,
		MerchantID:     r.MerchantID,
		MerchantName:   "Acme Store",
		Method:         r.Method,
		AmountMinor:    r.AmountMinor,
		Currency:       r.Currency,
		Status:         r.Status,
		Reference:      r.Reference,
		ProviderRef:    r.ProviderRef,
		CorrelationRef: r.CorrelationRef,
		ExpiresAt:      r.ExpiresAt,
		PaidAt:         r.PaidAt,
		CreatedAt:      r.CreatedAt,
	}, nil
}

func newSandboxService(repo repository.Querier, qr QRService) SandboxService {
	return NewSandboxService(repo, qr, qrSecret, zerolog.Nop())
}

func TestSandboxViewReturnsDisplayInfo(t *testing.T) {
	repo := newFakeQRRepo()
	qr, _ := newQRService(repo)
	sb := newSandboxService(repo, qr)

	qp, err := qr.Create(context.Background(), promptpayReq("ORDER-1"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	view, err := sb.View(context.Background(), qp.ID)
	if err != nil {
		t.Fatalf("View: %v", err)
	}
	if view.MerchantName != "Acme Store" {
		t.Fatalf("merchant_name=%q want Acme Store", view.MerchantName)
	}
	if view.AmountMinor != 10000 || view.Amount.String() != "100" {
		t.Fatalf("amount=%s minor=%d want 100/10000", view.Amount, view.AmountMinor)
	}
	if view.Status != domain.QRAwaitingPayment {
		t.Fatalf("status=%s want awaiting_payment", view.Status)
	}
}

func TestSandboxViewUnknownNotFound(t *testing.T) {
	repo := newFakeQRRepo()
	qr, _ := newQRService(repo)
	sb := newSandboxService(repo, qr)
	_, err := sb.View(context.Background(), uuid.New())
	if err != domain.ErrPaymentNotFound {
		t.Fatalf("View unknown err=%v want ErrPaymentNotFound", err)
	}
}

func TestSandboxPayTransitionsToPaid(t *testing.T) {
	repo := newFakeQRRepo()
	qr, _ := newQRService(repo)
	sb := newSandboxService(repo, qr)

	qp, err := qr.Create(context.Background(), promptpayReq("ORDER-PAY"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	paid, err := sb.Pay(context.Background(), qp.ID)
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if paid.Status != domain.QRPaid {
		t.Fatalf("status=%s want paid", paid.Status)
	}
	if paid.PayerBank != "Sandbox Bank" {
		t.Fatalf("payer_bank=%q want Sandbox Bank", paid.PayerBank)
	}
	if repo.webhookCount() != 1 {
		t.Fatalf("outbound webhook count=%d want 1", repo.webhookCount())
	}
	// Idempotent: paying again returns paid without a second webhook.
	again, err := sb.Pay(context.Background(), qp.ID)
	if err != nil {
		t.Fatalf("Pay (idempotent): %v", err)
	}
	if again.Status != domain.QRPaid {
		t.Fatalf("idempotent status=%s want paid", again.Status)
	}
	if repo.webhookCount() != 1 {
		t.Fatalf("idempotent webhook count=%d want 1 (no duplicate)", repo.webhookCount())
	}
}

func TestSandboxDeclineTransitionsToFailed(t *testing.T) {
	repo := newFakeQRRepo()
	qr, _ := newQRService(repo)
	sb := newSandboxService(repo, qr)

	qp, err := qr.Create(context.Background(), promptpayReq("ORDER-FAIL"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	failed, err := sb.Decline(context.Background(), qp.ID)
	if err != nil {
		t.Fatalf("Decline: %v", err)
	}
	if failed.Status != domain.QRFailed {
		t.Fatalf("status=%s want failed", failed.Status)
	}
	// A failed confirmation must NOT enqueue an outbound payment.paid webhook.
	if repo.webhookCount() != 0 {
		t.Fatalf("webhook count=%d want 0 on decline", repo.webhookCount())
	}
}

func TestSandboxPayUnknownNotFound(t *testing.T) {
	repo := newFakeQRRepo()
	qr, _ := newQRService(repo)
	sb := newSandboxService(repo, qr)
	_, err := sb.Pay(context.Background(), uuid.New())
	if err != domain.ErrPaymentNotFound {
		t.Fatalf("Pay unknown err=%v want ErrPaymentNotFound", err)
	}
}
