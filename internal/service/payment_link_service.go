package service

import (
	"context"
	"crypto/rand"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// PaymentLinkService creates and manages shareable payment links. Every read and
// update is scoped by merchant id (from auth context) to prevent cross-merchant
// access (IDOR).
type PaymentLinkService interface {
	Create(ctx context.Context, merchantID uuid.UUID, createdBy *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error)
	List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.PaymentLink, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error)
	Disable(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error)
}

type paymentLinkService struct {
	repo    repository.Querier
	baseURL string
	log     zerolog.Logger
}

// NewPaymentLinkService wires the service. publicBaseURL is the origin the shared
// checkout URL is built on (e.g. the Next app origin); the URL is
// <baseURL>/pay/<public_id>.
func NewPaymentLinkService(repo repository.Querier, publicBaseURL string, log zerolog.Logger) PaymentLinkService {
	return &paymentLinkService{repo: repo, baseURL: strings.TrimRight(publicBaseURL, "/"), log: log.With().Str("service", "payment_link").Logger()}
}

func (s *paymentLinkService) Create(ctx context.Context, merchantID uuid.UUID, createdBy *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, errors.New("payment link repository not configured")
	}
	slug, err := generatePublicID()
	if err != nil {
		return nil, err
	}
	currency := req.Currency
	if currency == "" {
		currency = "THB"
	}
	linkType := req.LinkType
	if linkType == "" {
		linkType = "single_use"
	}
	methods := req.AllowedMethods
	if methods == nil {
		methods = []string{}
	}
	var createdByPg pgtype.UUID
	if createdBy != nil {
		createdByPg = toPgUUID(*createdBy)
	}
	var expiresPg pgtype.Timestamptz
	if req.ExpiresAt != nil {
		expiresPg = pgtype.Timestamptz{Time: *req.ExpiresAt, Valid: true}
	}
	row, err := s.repo.CreatePaymentLink(ctx, repository.CreatePaymentLinkParams{
		ID:             toPgUUID(uuid.New()),
		MerchantID:     toPgUUID(merchantID),
		PublicID:       slug,
		Title:          req.Title,
		Description:    req.Description,
		AmountMinor:    req.AmountMinor,
		Currency:       currency,
		AllowedMethods: methods,
		LinkType:       linkType,
		Status:         "active",
		Reference:      req.Reference,
		ImageUrl:       req.ImageURL,
		ExpiresAt:      expiresPg,
		CreatedBy:      createdByPg,
	})
	if err != nil {
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListPaymentLinksByMerchant(ctx, repository.ListPaymentLinksByMerchantParams{
		MerchantID: toPgUUID(merchantID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.PaymentLink, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.toDomain(r))
	}
	return out, nil
}

func (s *paymentLinkService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, domain.ErrPaymentLinkNotFound
	}
	row, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) Disable(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, domain.ErrPaymentLinkNotFound
	}
	row, err := s.repo.UpdatePaymentLinkStatus(ctx, repository.UpdatePaymentLinkStatusParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
		Status:     "disabled",
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) toDomain(r repository.PaymentLink) *domain.PaymentLink {
	pl := &domain.PaymentLink{
		ID:             pgUUIDToUUID(r.ID),
		MerchantID:     pgUUIDToUUID(r.MerchantID),
		PublicID:       r.PublicID,
		Title:          r.Title,
		Description:    r.Description,
		AmountMinor:    r.AmountMinor,
		Currency:       r.Currency,
		AllowedMethods: r.AllowedMethods,
		LinkType:       r.LinkType,
		Status:         r.Status,
		Reference:      r.Reference,
		ImageURL:       r.ImageUrl,
		URL:            s.baseURL + "/pay/" + r.PublicID,
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}
	if r.ExpiresAt.Valid {
		t := r.ExpiresAt.Time
		pl.ExpiresAt = &t
	}
	if pl.AllowedMethods == nil {
		pl.AllowedMethods = []string{}
	}
	return pl
}

// generatePublicID returns a URL-safe base62-ish slug (hex is fine and avoids an
// alphabet dependency) with enough entropy to be unguessable.
func generatePublicID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const hexdigits = "0123456789abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, len(buf))
	for i, b := range buf {
		out[i] = hexdigits[int(b)%len(hexdigits)]
	}
	return "pl_" + string(out), nil
}
