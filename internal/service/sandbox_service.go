package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

// SandboxService powers the public sandbox payer simulator (the "Sandbox Bank"
// app). It is ONLY wired when SANDBOX_MODE=true. It lets a browser mark a QR
// payment paid/failed WITHOUT a merchant API key by having the SERVER build and
// sign the bank confirmation internally, then replay it through the very same
// QRService.ConfirmFromWebhook path a real bank webhook takes. The webhook
// secret never leaves the server.
type SandboxService interface {
	// View returns the public, secret-free display projection of a QR payment for
	// the simulator's confirmation screen. Returns ErrPaymentNotFound when unknown.
	View(ctx context.Context, id uuid.UUID) (*domain.SandboxQRView, error)
	// Pay simulates a SUCCESSFUL bank payment: it signs a "paid" confirmation and
	// runs it through the real webhook-confirmation path (reusing all validation).
	// Idempotent — an already-paid QR is returned as paid.
	Pay(ctx context.Context, id uuid.UUID) (*domain.QRPayment, error)
	// Decline simulates a FAILED bank payment (transitions to failed).
	Decline(ctx context.Context, id uuid.UUID) (*domain.QRPayment, error)
	// ListPending returns recent QR payments in the given status for the
	// simulator's "incoming requests" list. status defaults to awaiting_payment.
	ListPending(ctx context.Context, status string, limit int32) ([]domain.SandboxQRListItem, error)
}

type sandboxService struct {
	repo repository.Querier
	qr   QRService
	// webhookSecret is the shared HMAC secret used to sign the simulated bank
	// confirmation. Held server-side only; never returned to the browser.
	webhookSecret string
	log           zerolog.Logger
}

// NewSandboxService constructs the sandbox simulator service. webhookSecret is
// the same QR_WEBHOOK_SECRET the QR provider verifies against, so a
// server-signed confirmation is accepted by ConfirmFromWebhook exactly like a
// real bank callback.
func NewSandboxService(repo repository.Querier, qr QRService, webhookSecret string, log zerolog.Logger) SandboxService {
	return &sandboxService{
		repo:          repo,
		qr:            qr,
		webhookSecret: webhookSecret,
		log:           log.With().Str("service", "sandbox").Logger(),
	}
}

func (s *sandboxService) View(ctx context.Context, id uuid.UUID) (*domain.SandboxQRView, error) {
	row, err := s.repo.GetQRPaymentWithMerchant(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	amount, _ := money.FromMinor(row.AmountMinor, row.Currency)
	return &domain.SandboxQRView{
		ID:           pgUUIDToUUID(row.ID),
		Amount:       amount,
		AmountMinor:  row.AmountMinor,
		Currency:     row.Currency,
		MerchantName: row.MerchantName,
		Method:       domain.QRMethod(row.Method),
		Status:       domain.QRStatus(row.Status),
		Reference:    row.Reference,
	}, nil
}

// Pay signs a "paid" confirmation for the QR and replays it through the real
// webhook path.
func (s *sandboxService) Pay(ctx context.Context, id uuid.UUID) (*domain.QRPayment, error) {
	return s.confirm(ctx, id, string(domain.QRPaid))
}

// Decline signs a "failed" confirmation for the QR and replays it.
func (s *sandboxService) Decline(ctx context.Context, id uuid.UUID) (*domain.QRPayment, error) {
	return s.confirm(ctx, id, "failed")
}

// confirm loads the QR, builds the correct domain.QRWebhook for status, signs it
// with the configured secret, and hands it to QRService.ConfirmFromWebhook so
// ALL existing signature / correlation / amount / state validation is reused.
func (s *sandboxService) confirm(ctx context.Context, id uuid.UUID, status string) (*domain.QRPayment, error) {
	row, err := s.repo.GetQRPayment(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}

	// Build the confirmation exactly as a bank webhook would: carry BOTH the
	// correlation_ref (PromptPay echoes the reference we embedded) AND the
	// provider_ref (card / cross-border rails) when present, so correlateWebhook
	// resolves this row regardless of method.
	wh := domain.QRWebhook{
		Reference:   ptrStr(row.CorrelationRef),
		ProviderRef: ptrStr(row.ProviderRef),
		Amount:      amount,
		Currency:    row.Currency,
		PayerBank:   "Sandbox Bank",
		Status:      status,
		PaidAt:      time.Now(),
	}
	body, err := json.Marshal(wh)
	if err != nil {
		return nil, err
	}
	// The SERVER signs the body with the shared secret; the browser never sees it.
	// Use the preferred timestamped (v1) form so the confirmation passes the
	// inbound replay/freshness check exactly like a real bank callback.
	sig := external.ComputeQRWebhookSignatureV1(s.webhookSecret, time.Now().Unix(), body)
	return s.qr.ConfirmFromWebhook(ctx, sig, body)
}

func (s *sandboxService) ListPending(ctx context.Context, status string, limit int32) ([]domain.SandboxQRListItem, error) {
	if status == "" {
		status = string(domain.QRAwaitingPayment)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := s.repo.ListPendingQRPayments(ctx, repository.ListPendingQRPaymentsParams{
		Status: status,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	out := make([]domain.SandboxQRListItem, 0, len(rows))
	for _, r := range rows {
		amount, _ := money.FromMinor(r.AmountMinor, r.Currency)
		item := domain.SandboxQRListItem{
			ID:           pgUUIDToUUID(r.ID),
			Amount:       amount,
			AmountMinor:  r.AmountMinor,
			Currency:     r.Currency,
			MerchantName: r.MerchantName,
			Method:       domain.QRMethod(r.Method),
			Status:       domain.QRStatus(r.Status),
		}
		if r.CreatedAt.Valid {
			item.CreatedAt = r.CreatedAt.Time
		}
		out = append(out, item)
	}
	return out, nil
}
