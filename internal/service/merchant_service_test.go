package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/repository"
)

// fakeMerchantRepo implements the merchant slice of repository.Querier.
type fakeMerchantRepo struct {
	repository.Querier

	mu        sync.Mutex
	byID      map[uuid.UUID]repository.Merchant
	byHash    map[string]repository.Merchant
	createErr error
}

func newFakeMerchantRepo() *fakeMerchantRepo {
	return &fakeMerchantRepo{
		byID:   map[uuid.UUID]repository.Merchant{},
		byHash: map[string]repository.Merchant{},
	}
}

func (f *fakeMerchantRepo) CreateMerchant(_ context.Context, arg repository.CreateMerchantParams) (repository.Merchant, error) {
	if f.createErr != nil {
		return repository.Merchant{}, f.createErr
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	m := repository.Merchant{
		ID:                 arg.ID,
		Name:               arg.Name,
		Status:             "active",
		ApiKeyHash:         arg.ApiKeyHash,
		Mcc:                arg.Mcc,
		SettlementCurrency: arg.SettlementCurrency,
	}
	f.byID[key(arg.ID)] = m
	f.byHash[arg.ApiKeyHash] = m
	return m, nil
}

func (f *fakeMerchantRepo) GetMerchant(_ context.Context, id pgtype.UUID) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byID[key(id)]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return m, nil
}

func (f *fakeMerchantRepo) GetMerchantByAPIKeyHash(_ context.Context, hash string) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	m, ok := f.byHash[hash]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return m, nil
}

// ---- Onboard: key generation + hash-only persistence ----------------------

func TestOnboardIssuesKeyAndPersistsHashOnly(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())

	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{
		Name: "Acme", SettlementCurrency: "THB",
	})
	if err != nil {
		t.Fatalf("Onboard: %v", err)
	}
	if !strings.HasPrefix(cred.APIKey, "sk_live_") {
		t.Fatalf("api key has wrong prefix: %q", cred.APIKey)
	}
	// The stored hash must be the SHA-256 of the issued key, and the raw key must
	// NOT be persisted anywhere.
	stored, ok := repo.byHash[middleware.HashAPIKey(cred.APIKey)]
	if !ok {
		t.Fatal("stored merchant is not keyed by the hash of the issued API key")
	}
	if stored.ApiKeyHash == cred.APIKey {
		t.Fatal("raw API key was persisted instead of its hash")
	}
	if stored.ApiKeyHash == "" || len(stored.ApiKeyHash) != 64 {
		t.Fatalf("stored hash looks wrong: %q", stored.ApiKeyHash)
	}
}

func TestOnboardNoRepoErrors(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "X", SettlementCurrency: "THB"}); err == nil {
		t.Fatal("expected error when repo is not configured")
	}
}

func TestOnboardPropagatesRepoError(t *testing.T) {
	repo := newFakeMerchantRepo()
	repo.createErr = errors.New("db down")
	svc := NewMerchantService(repo, zerolog.Nop())
	if _, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "X", SettlementCurrency: "THB"}); err == nil {
		t.Fatal("expected the repo error to propagate")
	}
}

// ---- Get ------------------------------------------------------------------

func TestMerchantGetRoundTrip(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := svc.Get(context.Background(), cred.Merchant.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Name != "Acme" || got.SettlementCurrency != "THB" {
		t.Fatalf("round-trip mismatch: %+v", got)
	}
}

func TestMerchantGetNotFound(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	if _, err := svc.Get(context.Background(), uuid.New()); !errors.Is(err, domain.ErrMerchantNotFound) {
		t.Fatalf("err=%v want ErrMerchantNotFound", err)
	}
}

func TestMerchantGetNoRepo(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.Get(context.Background(), uuid.New()); !errors.Is(err, domain.ErrMerchantNotFound) {
		t.Fatalf("err=%v want ErrMerchantNotFound", err)
	}
}

// ---- ResolveByAPIKeyHash (auth resolution) --------------------------------

func TestResolveByAPIKeyHashSuccessAndMiss(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	cred, err := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "Acme", SettlementCurrency: "THB"})
	if err != nil {
		t.Fatal(err)
	}

	// Correct hash resolves the merchant.
	m, err := svc.ResolveByAPIKeyHash(context.Background(), middleware.HashAPIKey(cred.APIKey))
	if err != nil {
		t.Fatalf("resolve valid: %v", err)
	}
	if m.ID != cred.Merchant.ID {
		t.Fatalf("resolved wrong merchant: %s != %s", m.ID, cred.Merchant.ID)
	}

	// An unknown hash is unauthorized (never leaks existence).
	if _, err := svc.ResolveByAPIKeyHash(context.Background(), middleware.HashAPIKey("sk_live_bogus")); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("err=%v want ErrUnauthorized", err)
	}
}

func TestResolveByAPIKeyHashNoRepo(t *testing.T) {
	svc := NewMerchantService(nil, zerolog.Nop())
	if _, err := svc.ResolveByAPIKeyHash(context.Background(), "x"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("err=%v want ErrUnauthorized", err)
	}
}

// Two onboardings must never mint the same API key (crypto-random).
func TestGenerateAPIKeyUnique(t *testing.T) {
	repo := newFakeMerchantRepo()
	svc := NewMerchantService(repo, zerolog.Nop())
	c1, _ := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "A", SettlementCurrency: "THB"})
	c2, _ := svc.Onboard(context.Background(), domain.CreateMerchantRequest{Name: "B", SettlementCurrency: "THB"})
	if c1.APIKey == c2.APIKey {
		t.Fatal("two onboardings minted the same API key")
	}
}
