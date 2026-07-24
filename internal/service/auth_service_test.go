package service

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// fakeAuthRepo implements the auth slice of repository.Querier.
type fakeAuthRepo struct {
	repository.Querier
	mu         sync.Mutex
	usersByOA  map[string]repository.MerchantUser // "provider|subject" -> user
	usersByID  map[uuid.UUID]repository.MerchantUser
	merchants  map[uuid.UUID]bool
	createCnt  int
	touchedIDs []uuid.UUID
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		usersByOA: map[string]repository.MerchantUser{},
		usersByID: map[uuid.UUID]repository.MerchantUser{},
		merchants: map[uuid.UUID]bool{},
	}
}

func (f *fakeAuthRepo) GetMerchantUserByOAuth(_ context.Context, arg repository.GetMerchantUserByOAuthParams) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.usersByOA[arg.OauthProvider+"|"+arg.OauthSubject]
	if !ok {
		return repository.MerchantUser{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeAuthRepo) CreateMerchant(_ context.Context, arg repository.CreateMerchantParams) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.merchants[uuid.UUID(arg.ID.Bytes)] = true
	return repository.Merchant{ID: arg.ID, Name: arg.Name, Status: "active", SettlementCurrency: arg.SettlementCurrency}, nil
}

func (f *fakeAuthRepo) CreateMerchantUser(_ context.Context, arg repository.CreateMerchantUserParams) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createCnt++
	u := repository.MerchantUser{
		ID:            arg.ID,
		MerchantID:    arg.MerchantID,
		Email:         arg.Email,
		Name:          arg.Name,
		AvatarUrl:     arg.AvatarUrl,
		OauthProvider: arg.OauthProvider,
		OauthSubject:  arg.OauthSubject,
		Role:          "owner",
	}
	f.usersByOA[arg.OauthProvider+"|"+arg.OauthSubject] = u
	f.usersByID[uuid.UUID(arg.ID.Bytes)] = u
	return u, nil
}

func (f *fakeAuthRepo) TouchMerchantUserLogin(_ context.Context, id pgtype.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.touchedIDs = append(f.touchedIDs, uuid.UUID(id.Bytes))
	return nil
}

func (f *fakeAuthRepo) GetMerchantUserByID(_ context.Context, id pgtype.UUID) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.usersByID[uuid.UUID(id.Bytes)]
	if !ok {
		return repository.MerchantUser{}, pgx.ErrNoRows
	}
	return u, nil
}

func TestLoginWithOAuthProvisionsThenReuses(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, zerolog.Nop())
	id := domain.OAuthIdentity{Subject: "google-123", Email: "shop@x.co", Name: "Shop"}

	first, err := svc.LoginWithOAuth(context.Background(), "google", id)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	if first.Email != "shop@x.co" || first.MerchantID == uuid.Nil {
		t.Fatalf("unexpected user: %+v", first)
	}
	if repo.createCnt != 1 || len(repo.merchants) != 1 {
		t.Fatalf("first login should provision 1 merchant+user, got users=%d merchants=%d", repo.createCnt, len(repo.merchants))
	}

	second, err := svc.LoginWithOAuth(context.Background(), "google", id)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("second login created a new user: %v != %v", second.ID, first.ID)
	}
	if repo.createCnt != 1 {
		t.Fatalf("second login must NOT re-provision, createCnt=%d", repo.createCnt)
	}
	if len(repo.touchedIDs) != 2 {
		t.Fatalf("both logins should touch last_login_at, got %d", len(repo.touchedIDs))
	}
}
