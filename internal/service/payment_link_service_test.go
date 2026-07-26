package service

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

type fakeLinkRepo struct {
	repository.Querier
	mu    sync.Mutex
	byID  map[uuid.UUID]repository.PaymentLink
	slugs map[string]bool
}

func newFakeLinkRepo() *fakeLinkRepo {
	return &fakeLinkRepo{byID: map[uuid.UUID]repository.PaymentLink{}, slugs: map[string]bool{}}
}

func (f *fakeLinkRepo) CreatePaymentLink(_ context.Context, a repository.CreatePaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.slugs[a.PublicID] = true
	pl := repository.PaymentLink{
		ID: a.ID, MerchantID: a.MerchantID, PublicID: a.PublicID, Title: a.Title,
		Description: a.Description, AmountMinor: a.AmountMinor, Currency: a.Currency,
		AllowedMethods: a.AllowedMethods, LinkType: a.LinkType, Status: a.Status,
		Reference: a.Reference, ImageUrl: a.ImageUrl,
	}
	f.byID[uuid.UUID(a.ID.Bytes)] = pl
	return pl, nil
}

func (f *fakeLinkRepo) GetPaymentLink(_ context.Context, a repository.GetPaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	pl, ok := f.byID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(pl.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return pl, nil
}

func newLinkSvc(repo repository.Querier) PaymentLinkService {
	return NewPaymentLinkService(repo, "https://pay.example", zerolog.Nop())
}

func TestCreatePaymentLinkGeneratesSlugAndURL(t *testing.T) {
	repo := newFakeLinkRepo()
	svc := newLinkSvc(repo)
	mid := uuid.New()
	uid := uuid.New()
	got, err := svc.Create(context.Background(), mid, &uid, domain.CreatePaymentLinkRequest{
		Title: "Coffee", AmountMinor: 5000,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.PublicID == "" {
		t.Fatal("expected a generated public_id slug")
	}
	if got.Currency != "THB" {
		t.Fatalf("currency default = %q want THB", got.Currency)
	}
	if got.LinkType != "single_use" {
		t.Fatalf("link_type default = %q want single_use", got.LinkType)
	}
	if got.Status != "active" {
		t.Fatalf("status = %q want active", got.Status)
	}
	if !strings.HasSuffix(got.URL, "/pay/"+got.PublicID) || !strings.HasPrefix(got.URL, "https://pay.example") {
		t.Fatalf("url = %q want https://pay.example/pay/<slug>", got.URL)
	}
}

func TestGetPaymentLinkOtherMerchantIsNotFound(t *testing.T) {
	repo := newFakeLinkRepo()
	svc := newLinkSvc(repo)
	owner := uuid.New()
	created, _ := svc.Create(context.Background(), owner, nil, domain.CreatePaymentLinkRequest{Title: "X", AmountMinor: 100})

	// A different merchant must NOT be able to read it.
	_, err := svc.Get(context.Background(), uuid.New(), created.ID)
	if err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("cross-merchant Get err=%v want ErrPaymentLinkNotFound", err)
	}
	// The owner can.
	got, err := svc.Get(context.Background(), owner, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("owner Get failed: %v", err)
	}
}
