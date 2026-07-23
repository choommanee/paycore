package service

import (
	"context"

	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// AdminService backs the platform operator console behind X-Admin-Key. Its
// methods are intentionally NOT scoped to a single merchant — they read across
// the whole platform. Authorization is enforced by the AdminAuth middleware
// before any of these are reached.
type AdminService interface {
	ListMerchants(ctx context.Context, limit int) ([]*domain.Merchant, error)
	ListAuditLog(ctx context.Context, limit int) ([]*domain.AuditEntry, error)
	ListDisputes(ctx context.Context, limit int) ([]*domain.Dispute, error)
	ListSettlements(ctx context.Context, limit int) ([]*domain.Settlement, error)
	PlatformStats(ctx context.Context) (*domain.PlatformStats, error)
}

type adminService struct {
	repo repository.Querier
	log  zerolog.Logger
}

// NewAdminService wires the admin service.
func NewAdminService(repo repository.Querier, log zerolog.Logger) AdminService {
	return &adminService{repo: repo, log: log.With().Str("service", "admin").Logger()}
}

// ListMerchants returns all onboarded merchants, most recent first.
func (s *adminService) ListMerchants(ctx context.Context, limit int) ([]*domain.Merchant, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListAllMerchants(ctx, repository.ListAllMerchantsParams{
		Limit:  clampLimit(limit),
		Offset: 0,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Merchant, 0, len(rows))
	for i := range rows {
		out = append(out, toDomainMerchant(rows[i]))
	}
	return out, nil
}

// ListAuditLog returns the most recent audit-log rows.
func (s *adminService) ListAuditLog(ctx context.Context, limit int) ([]*domain.AuditEntry, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListAuditLog(ctx, clampLimit(limit))
	if err != nil {
		return nil, err
	}
	out := make([]*domain.AuditEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, &domain.AuditEntry{
			ID:         pgUUIDToUUID(r.ID),
			Actor:      r.Actor,
			Action:     r.Action,
			EntityType: r.EntityType,
			EntityID:   r.EntityID,
			Metadata:   r.Metadata,
			CreatedAt:  r.CreatedAt.Time,
		})
	}
	return out, nil
}

// ListDisputes returns all disputes across the platform, most recent first.
func (s *adminService) ListDisputes(ctx context.Context, limit int) ([]*domain.Dispute, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListAllDisputes(ctx, repository.ListAllDisputesParams{
		Limit:  clampLimit(limit),
		Offset: 0,
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

// ListSettlements returns platform-wide payout rows across all merchants, most
// recent first. Reuses toDomainSettlement from the merchant service (same pkg).
func (s *adminService) ListSettlements(ctx context.Context, limit int) ([]*domain.Settlement, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListAllPayouts(ctx, clampLimit(limit))
	if err != nil {
		return nil, err
	}
	out := make([]*domain.Settlement, 0, len(rows))
	for _, r := range rows {
		out = append(out, toDomainSettlement(r))
	}
	return out, nil
}

// PlatformStats returns platform-wide key risk indicators across all merchants.
func (s *adminService) PlatformStats(ctx context.Context) (*domain.PlatformStats, error) {
	if s.repo == nil {
		return &domain.PlatformStats{}, nil
	}
	row, err := s.repo.PlatformStats(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.PlatformStats{
		MerchantCount: row.MerchantCount,
		Count:         row.PaymentCount,
		VolumeMinor:   row.VolumeMinor,
		ByStatus: domain.StatsByStatus{
			Authorized: row.AuthorizedCount,
			Captured:   row.CapturedCount,
			Refunded:   row.RefundedCount,
			Failed:     row.FailedCount,
		},
		SuccessRate:     ratio(row.AuthorizedCount, row.PaymentCount),
		RefundRatio:     ratio(row.RefundedCount, row.PaymentCount),
		ChargebackRatio: ratio(row.DisputeCount, row.PaymentCount),
	}, nil
}
