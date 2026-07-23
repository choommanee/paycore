package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

// DisputeService manages chargebacks / disputes and their state machine.
type DisputeService interface {
	Open(ctx context.Context, merchantID, paymentID uuid.UUID, req domain.OpenDisputeRequest) (*domain.Dispute, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.Dispute, error)
	// ListByPayment is scoped to the authenticated merchant: the payment must
	// belong to that merchant, mirroring the ownership check in Open.
	ListByPayment(ctx context.Context, merchantID, paymentID uuid.UUID) ([]*domain.Dispute, error)
	// ListByMerchant returns all disputes for the authenticated merchant, most
	// recent first (GET /v1/disputes).
	ListByMerchant(ctx context.Context, merchantID uuid.UUID, limit int) ([]*domain.Dispute, error)
	Transition(ctx context.Context, id uuid.UUID, to domain.DisputeStatus, evidence []byte) (*domain.Dispute, error)
}

type disputeService struct {
	repo repository.Querier
	log  zerolog.Logger
}

// NewDisputeService wires the dispute service.
func NewDisputeService(repo repository.Querier, log zerolog.Logger) DisputeService {
	return &disputeService{repo: repo, log: log.With().Str("service", "dispute").Logger()}
}

// Open opens a dispute against a captured payment. The payment must belong to
// the authenticated merchant. When no amount is supplied the full payment amount
// is disputed.
func (s *disputeService) Open(ctx context.Context, merchantID, paymentID uuid.UUID, req domain.OpenDisputeRequest) (*domain.Dispute, error) {
	if s.repo == nil {
		return nil, errors.New("dispute repository not configured")
	}
	// Merchant-scoped fetch: a payment belonging to another merchant returns
	// ErrPaymentNotFound (404) rather than a 403 existence oracle.
	pay, err := s.repo.GetPaymentForMerchant(ctx, repository.GetPaymentForMerchantParams{
		ID:         toPgUUID(paymentID),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}

	amountMinor := pay.AmountMinor
	if req.Amount.IsPositive() {
		m, err := money.ToMinor(req.Amount, pay.Currency)
		if err != nil || m <= 0 || m > pay.AmountMinor {
			return nil, domain.ErrInvalidRequest
		}
		amountMinor = m
	}

	row, err := s.repo.CreateDispute(ctx, repository.CreateDisputeParams{
		ID:          toPgUUID(uuid.New()),
		PaymentID:   pay.ID,
		MerchantID:  pay.MerchantID,
		AmountMinor: amountMinor,
		Currency:    pay.Currency,
		Reason:      strPtr(req.Reason),
		Status:      string(domain.DisputeOpened),
		NetworkRef:  strPtr(req.NetworkRef),
	})
	if err != nil {
		return nil, err
	}
	return toDomainDispute(row), nil
}

// Get returns a dispute by id.
func (s *disputeService) Get(ctx context.Context, id uuid.UUID) (*domain.Dispute, error) {
	if s.repo == nil {
		return nil, domain.ErrDisputeNotFound
	}
	row, err := s.repo.GetDispute(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDisputeNotFound
		}
		return nil, err
	}
	return toDomainDispute(row), nil
}

// ListByPayment returns all disputes attached to a payment, most recent first.
// The payment must belong to the authenticated merchant; otherwise it returns
// ErrPaymentNotFound (404) rather than leaking another merchant's disputes.
func (s *disputeService) ListByPayment(ctx context.Context, merchantID, paymentID uuid.UUID) ([]*domain.Dispute, error) {
	if s.repo == nil {
		return nil, nil
	}
	// Ownership check: only the owning merchant may list a payment's disputes.
	pay, err := s.repo.GetPaymentForMerchant(ctx, repository.GetPaymentForMerchantParams{
		ID:         toPgUUID(paymentID),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentNotFound
		}
		return nil, err
	}
	rows, err := s.repo.ListDisputesByPayment(ctx, pay.ID)
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Dispute, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainDispute(r))
	}
	return out, nil
}

// ListByMerchant returns all disputes belonging to the authenticated merchant,
// most recent first. Scoped by merchant_id so no cross-tenant dispute leaks.
func (s *disputeService) ListByMerchant(ctx context.Context, merchantID uuid.UUID, limit int) ([]*domain.Dispute, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListDisputesByMerchant(ctx, repository.ListDisputesByMerchantParams{
		MerchantID: toPgUUID(merchantID),
		Limit:      clampLimit(limit),
		Offset:     0,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Dispute, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainDispute(r))
	}
	return out, nil
}

// Transition moves a dispute through its state machine:
//
//	opened -> under_review -> won | lost | accepted
//
// won/lost/accepted are terminal. Terminal transitions stamp resolved_at.
func (s *disputeService) Transition(ctx context.Context, id uuid.UUID, to domain.DisputeStatus, evidence []byte) (*domain.Dispute, error) {
	if s.repo == nil {
		return nil, domain.ErrDisputeNotFound
	}
	row, err := s.repo.GetDispute(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDisputeNotFound
		}
		return nil, err
	}
	if err := assertDisputeTransition(domain.DisputeStatus(row.Status), to); err != nil {
		return nil, err
	}

	var resolvedAt pgtype.Timestamptz
	if isDisputeTerminal(to) {
		resolvedAt = pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	}

	updated, err := s.repo.UpdateDisputeStatus(ctx, repository.UpdateDisputeStatusParams{
		ID:         row.ID,
		Status:     string(to),
		Evidence:   evidence, // nil keeps existing (COALESCE in SQL)
		ResolvedAt: resolvedAt,
	})
	if err != nil {
		return nil, err
	}
	return toDomainDispute(updated), nil
}

// assertDisputeTransition enforces the dispute state machine.
func assertDisputeTransition(from, to domain.DisputeStatus) error {
	allowed := map[domain.DisputeStatus]map[domain.DisputeStatus]bool{
		domain.DisputeOpened: {
			domain.DisputeUnderReview: true,
			domain.DisputeAccepted:    true,
		},
		domain.DisputeUnderReview: {
			domain.DisputeWon:      true,
			domain.DisputeLost:     true,
			domain.DisputeAccepted: true,
		},
	}
	if allowed[from][to] {
		return nil
	}
	return domain.ErrInvalidState
}

func isDisputeTerminal(s domain.DisputeStatus) bool {
	switch s {
	case domain.DisputeWon, domain.DisputeLost, domain.DisputeAccepted:
		return true
	default:
		return false
	}
}

func toDomainDispute(r repository.Dispute) *domain.Dispute {
	amount, _ := money.FromMinor(r.AmountMinor, r.Currency)
	d := &domain.Dispute{
		ID:         pgUUIDToUUID(r.ID),
		PaymentID:  pgUUIDToUUID(r.PaymentID),
		MerchantID: pgUUIDToUUID(r.MerchantID),
		Amount:     amount,
		Currency:   r.Currency,
		Reason:     ptrStr(r.Reason),
		Status:     domain.DisputeStatus(r.Status),
		NetworkRef: ptrStr(r.NetworkRef),
		OpenedAt:   r.OpenedAt.Time,
		CreatedAt:  r.CreatedAt.Time,
		UpdatedAt:  r.UpdatedAt.Time,
	}
	if r.ResolvedAt.Valid {
		t := r.ResolvedAt.Time
		d.ResolvedAt = &t
	}
	return d
}
