package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/repository"
)

// AuthService resolves dashboard human identities. LoginWithOAuth is
// login-or-provision: a first-time provider identity creates a merchant (with a
// generated sandbox API key) and its owner user; a returning identity is reused.
type AuthService interface {
	LoginWithOAuth(ctx context.Context, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.MerchantUser, error)
}

type authService struct {
	repo repository.Querier
	log  zerolog.Logger
}

// NewAuthService wires the auth service.
func NewAuthService(repo repository.Querier, log zerolog.Logger) AuthService {
	return &authService{repo: repo, log: log.With().Str("service", "auth").Logger()}
}

func (s *authService) LoginWithOAuth(ctx context.Context, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error) {
	if s.repo == nil {
		return nil, errors.New("auth repository not configured")
	}
	// Returning identity: reuse.
	row, err := s.repo.GetMerchantUserByOAuth(ctx, repository.GetMerchantUserByOAuthParams{
		OauthProvider: provider,
		OauthSubject:  id.Subject,
	})
	if err == nil {
		_ = s.repo.TouchMerchantUserLogin(ctx, row.ID)
		return toDomainMerchantUser(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// First-time identity: provision merchant + owner user.
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	merchantID := uuid.New()
	name := id.Name
	if name == "" {
		name = id.Email
	}
	if _, err := s.repo.CreateMerchant(ctx, repository.CreateMerchantParams{
		ID:                 toPgUUID(merchantID),
		Name:               name,
		Mcc:                nil,
		SettlementCurrency: "THB",
		ApiKeyHash:         middleware.HashAPIKey(rawKey),
	}); err != nil {
		return nil, err
	}
	created, err := s.repo.CreateMerchantUser(ctx, repository.CreateMerchantUserParams{
		ID:            toPgUUID(uuid.New()),
		MerchantID:    toPgUUID(merchantID),
		Email:         id.Email,
		Name:          name,
		AvatarUrl:     id.Picture,
		OauthProvider: provider,
		OauthSubject:  id.Subject,
	})
	if err != nil {
		return nil, err
	}
	_ = s.repo.TouchMerchantUserLogin(ctx, created.ID)
	return toDomainMerchantUser(created), nil
}

func (s *authService) GetUser(ctx context.Context, id uuid.UUID) (*domain.MerchantUser, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	row, err := s.repo.GetMerchantUserByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return toDomainMerchantUser(row), nil
}

func toDomainMerchantUser(r repository.MerchantUser) *domain.MerchantUser {
	u := &domain.MerchantUser{
		ID:         pgUUIDToUUID(r.ID),
		MerchantID: pgUUIDToUUID(r.MerchantID),
		Email:      r.Email,
		Name:       r.Name,
		AvatarURL:  r.AvatarUrl,
		Provider:   r.OauthProvider,
		Role:       r.Role,
	}
	u.CreatedAt = r.CreatedAt.Time
	u.UpdatedAt = r.UpdatedAt.Time
	return u
}
