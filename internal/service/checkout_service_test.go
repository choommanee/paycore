package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	moneypkg "github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

type decimalDecimal = decimal.Decimal

func moneyFromMinor(minor int64, currency string) (decimal.Decimal, error) {
	return moneypkg.FromMinor(minor, currency)
}

// ---- fakes (shared by Tasks 3–5) -------------------------------------------

type fakeCheckoutRepo struct {
	repository.Querier
	mu             sync.Mutex
	linksByPublic  map[string]repository.PaymentLink
	linksByID      map[uuid.UUID]repository.PaymentLink
	sessByHash     map[string]repository.CheckoutSession
	sessByID       map[uuid.UUID]repository.CheckoutSession
	merchantNames  map[uuid.UUID]string
	linkStatusSets []string // records UpdatePaymentLinkStatus calls
}

func newFakeCheckoutRepo() *fakeCheckoutRepo {
	return &fakeCheckoutRepo{
		linksByPublic: map[string]repository.PaymentLink{},
		linksByID:     map[uuid.UUID]repository.PaymentLink{},
		sessByHash:    map[string]repository.CheckoutSession{},
		sessByID:      map[uuid.UUID]repository.CheckoutSession{},
		merchantNames: map[uuid.UUID]string{},
	}
}

func (f *fakeCheckoutRepo) putLink(l repository.PaymentLink) {
	f.linksByPublic[l.PublicID] = l
	f.linksByID[uuid.UUID(l.ID.Bytes)] = l
}

func (f *fakeCheckoutRepo) GetPaymentLinkByPublicID(_ context.Context, publicID string) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByPublic[publicID]
	if !ok {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return l, nil
}

func (f *fakeCheckoutRepo) GetPaymentLink(_ context.Context, a repository.GetPaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(l.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return l, nil
}

func (f *fakeCheckoutRepo) GetMerchant(_ context.Context, id pgtype.UUID) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name, ok := f.merchantNames[uuid.UUID(id.Bytes)]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return repository.Merchant{ID: id, Name: name, Status: "active"}, nil
}

func (f *fakeCheckoutRepo) CreateCheckoutSession(_ context.Context, a repository.CreateCheckoutSessionParams) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs := repository.CheckoutSession{
		ID: a.ID, MerchantID: a.MerchantID, PaymentLinkID: a.PaymentLinkID,
		SessionTokenHash: a.SessionTokenHash, AmountMinor: a.AmountMinor, Currency: a.Currency,
		Status: a.Status, SelectedMethod: a.SelectedMethod, PaymentID: a.PaymentID,
		QrPaymentID: a.QrPaymentID, CustomerEmail: a.CustomerEmail, ReturnUrl: a.ReturnUrl,
		ExpiresAt: a.ExpiresAt,
	}
	f.sessByHash[a.SessionTokenHash] = cs
	f.sessByID[uuid.UUID(a.ID.Bytes)] = cs
	return cs, nil
}

func (f *fakeCheckoutRepo) GetCheckoutSessionByTokenHash(_ context.Context, hash string) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs, ok := f.sessByHash[hash]
	if !ok {
		return repository.CheckoutSession{}, pgx.ErrNoRows
	}
	return cs, nil
}

func (f *fakeCheckoutRepo) UpdateCheckoutSession(_ context.Context, a repository.UpdateCheckoutSessionParams) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs := f.sessByID[uuid.UUID(a.ID.Bytes)]
	cs.Status = a.Status
	cs.SelectedMethod = a.SelectedMethod
	cs.PaymentID = a.PaymentID
	cs.QrPaymentID = a.QrPaymentID
	cs.CustomerEmail = a.CustomerEmail
	f.sessByID[uuid.UUID(a.ID.Bytes)] = cs
	f.sessByHash[cs.SessionTokenHash] = cs
	return cs, nil
}

func (f *fakeCheckoutRepo) UpdatePaymentLinkStatus(_ context.Context, a repository.UpdatePaymentLinkStatusParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.linkStatusSets = append(f.linkStatusSets, a.Status)
	l := f.linksByID[uuid.UUID(a.ID.Bytes)]
	l.Status = a.Status
	f.linksByID[uuid.UUID(a.ID.Bytes)] = l
	return l, nil
}

// ConsumePaymentLinkIfActive mirrors the real sqlc query: it atomically flips
// an active link to 'paid' and returns pgx.ErrNoRows if the link is missing,
// not owned by merchantID, or not currently 'active' (already consumed,
// disabled, etc). A successful flip is recorded into linkStatusSets alongside
// UpdatePaymentLinkStatus calls so existing "was the link closed" assertions
// keep working regardless of which query performed the transition.
func (f *fakeCheckoutRepo) ConsumePaymentLinkIfActive(_ context.Context, a repository.ConsumePaymentLinkIfActiveParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(l.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) || l.Status != "active" {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	l.Status = "paid"
	f.linksByID[uuid.UUID(a.ID.Bytes)] = l
	f.linkStatusSets = append(f.linkStatusSets, "paid")
	return l, nil
}

// ReleasePaymentLinkReservation mirrors the real sqlc query: it reverts a link
// to 'active' ONLY if it is currently 'paid' (the reserved state). If the status
// is anything else (e.g. a concurrent Disable set it 'disabled'), it affects no
// row and returns pgx.ErrNoRows, leaving that status untouched. A successful
// release records "active" in linkStatusSets.
func (f *fakeCheckoutRepo) ReleasePaymentLinkReservation(_ context.Context, a repository.ReleasePaymentLinkReservationParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(l.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) || l.Status != "paid" {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	l.Status = "active"
	f.linksByID[uuid.UUID(a.ID.Bytes)] = l
	f.linkStatusSets = append(f.linkStatusSets, "active")
	return l, nil
}

type fakeCharger struct {
	createFn func(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error)

	mu          sync.Mutex
	createCalls int // counts Create invocations; used to assert the charger fires exactly once
}

func (f *fakeCharger) Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
	f.mu.Lock()
	f.createCalls++
	f.mu.Unlock()
	if f.createFn != nil {
		return f.createFn(ctx, idemKey, req)
	}
	return &domain.Payment{ID: uuid.New(), Status: domain.StatusCaptured}, nil
}

func (f *fakeCharger) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.createCalls
}

type fakeQR struct {
	createFn func(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error)
	getFn    func(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error)

	mu       sync.Mutex
	getCalls int // counts Get invocations; used to assert on provider-poll counts
}

func (f *fakeQR) Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
	if f.createFn != nil {
		return f.createFn(ctx, req)
	}
	return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "EMVCO-PAYLOAD"}, nil
}

func (f *fakeQR) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error) {
	f.mu.Lock()
	f.getCalls++
	f.mu.Unlock()
	if f.getFn != nil {
		return f.getFn(ctx, merchantID, id)
	}
	return &domain.QRPayment{ID: id, Status: domain.QRAwaitingPayment, QRPayload: "EMVCO-PAYLOAD"}, nil
}

func (f *fakeQR) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.getCalls
}

type fakeVault struct {
	tokenizeFn func(ctx context.Context, pan string) (string, string, error)
}

func (f *fakeVault) Tokenize(ctx context.Context, pan string) (string, string, error) {
	if f.tokenizeFn != nil {
		return f.tokenizeFn(ctx, pan)
	}
	if len(pan) < 4 {
		return "", "", pgx.ErrNoRows
	}
	return "tokv1.fake", pan[len(pan)-4:], nil
}

// newCheckoutSvc builds the service with fakes. sandbox defaults to true so the
// card path is exercised; individual tests can rebuild with sandbox=false.
func newCheckoutSvc(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, sandbox bool) CheckoutService {
	return NewCheckoutService(repo, charger, qr, vault, sandbox, zerolog.Nop())
}

// mkLink seeds an active link and returns it.
func mkLink(f *fakeCheckoutRepo, merchant uuid.UUID, publicID string, amountMinor int64, methods []string) repository.PaymentLink {
	l := repository.PaymentLink{
		ID:             toPgUUID(uuid.New()),
		MerchantID:     toPgUUID(merchant),
		PublicID:       publicID,
		Title:          "Coffee",
		Description:    "Latte",
		AmountMinor:    amountMinor,
		Currency:       "THB",
		AllowedMethods: methods,
		LinkType:       "single_use",
		Status:         "active",
	}
	f.putLink(l)
	f.merchantNames[merchant] = "Acme Cafe"
	return l
}

// ---- Task 3 tests ----------------------------------------------------------

func TestCreateFromLinkReturnsTokenAndDisplay(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)

	view, err := svc.CreateFromLink(context.Background(), "pl_abc")
	if err != nil {
		t.Fatalf("CreateFromLink: %v", err)
	}
	if !strings.HasPrefix(view.Token, "cs_") {
		t.Fatalf("token = %q want cs_ prefix", view.Token)
	}
	if view.Status != string(domain.CheckoutOpen) {
		t.Fatalf("status = %q want open", view.Status)
	}
	if view.AmountMinor != 5000 || view.Currency != "THB" {
		t.Fatalf("amount/currency = %d/%s", view.AmountMinor, view.Currency)
	}
	if view.MerchantName != "Acme Cafe" || view.Title != "Coffee" {
		t.Fatalf("display = %q / %q", view.MerchantName, view.Title)
	}
	if len(view.AllowedMethods) != 2 {
		t.Fatalf("methods = %v want [card promptpay]", view.AllowedMethods)
	}
	if !view.Sandbox {
		t.Fatal("sandbox flag should be true")
	}
	// The RAW token must never be stored — only its hash keys the row.
	if _, ok := repo.sessByHash[middleware.HashAPIKey(view.Token)]; !ok {
		t.Fatal("session not stored under token hash")
	}
	if _, ok := repo.sessByHash[view.Token]; ok {
		t.Fatal("raw token must not be a storage key")
	}
}

func TestCreateFromLinkInactiveLinkIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_dead", 100, nil)
	l.Status = "disabled"
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)

	_, err := svc.CreateFromLink(context.Background(), "pl_dead")
	if err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("disabled link err = %v want ErrPaymentLinkNotFound", err)
	}
}

func TestCreateFromLinkUnknownIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.CreateFromLink(context.Background(), "pl_nope"); err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("unknown link err = %v want ErrPaymentLinkNotFound", err)
	}
}

var _ = time.Now // retained; time is used by later tasks' tests

// ---- Task 4 tests -----------------------------------------------------------

func openSession(t *testing.T, repo *fakeCheckoutRepo, svc CheckoutService) string {
	t.Helper()
	view, err := svc.CreateFromLink(context.Background(), "pl_abc")
	if err != nil {
		t.Fatalf("CreateFromLink: %v", err)
	}
	return view.Token
}

func TestPayPromptPayMintsQRAndRequiresAction(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	qr := &fakeQR{createFn: func(_ context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
		if req.Method != domain.QRPromptPayDynamic {
			t.Fatalf("method = %v want promptpay_dynamic", req.Method)
		}
		if !req.Amount.Equal(decimalFromMinor(t, 5000, "THB")) {
			t.Fatalf("amount not converted to major units: %s", req.Amount)
		}
		return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "PP-EMV"}, nil
	}}
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.QRPayload != "PP-EMV" {
		t.Fatalf("qr_payload = %q want PP-EMV", view.QRPayload)
	}
	if view.SelectedMethod != "promptpay" {
		t.Fatalf("selected_method = %q", view.SelectedMethod)
	}
}

func TestPayCardCapturedMarksPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	charger := &fakeCharger{createFn: func(_ context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
		if !strings.HasPrefix(idemKey, "cs_") {
			t.Fatalf("idem key = %q want cs_<session>", idemKey)
		}
		if !req.Capture {
			t.Fatal("card charge must Capture=true")
		}
		if req.PaymentToken != "tokv1.fake" {
			t.Fatalf("payment token = %q want vault token", req.PaymentToken)
		}
		return &domain.Payment{ID: uuid.New(), Status: domain.StatusCaptured}, nil
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutPaid) {
		t.Fatalf("status = %q want paid", view.Status)
	}
	// single_use link must be closed as paid.
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid]", repo.linkStatusSets)
	}
}

func TestPayCardRequiresActionOn3DS(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		return &domain.Payment{ID: uuid.New(), Status: domain.StatusRequiresAction, NextActionURL: "https://acs/challenge"}, nil
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000003220", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.NextActionURL != "https://acs/challenge" {
		t.Fatalf("next_action_url = %q", view.NextActionURL)
	}
	// NOTE: with reserve-before-charge, a single_use link's reservation happens
	// BEFORE the charge call returns, so it is already consumed ('paid') here
	// even though the *session* is still requires_action pending 3DS. This is
	// intentional: the reservation must not be released while a charge attempt
	// is in flight / awaiting customer action (err == nil from the charger), or
	// a concurrent second attempt could race in and double-charge exactly as
	// this fix is meant to prevent. It is only released on an actual
	// decline/error (see TestPayCardDeclineReleasesSingleUseLink).
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid] (reserved, held during 3DS pending)", repo.linkStatusSets)
	}
}

func TestPayCardDeclinedMarksFailedAndSurfacesError(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		return nil, domain.ErrCardDeclined
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000000002", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != domain.ErrCardDeclined {
		t.Fatalf("err = %v want ErrCardDeclined", err)
	}
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	if row.Status != string(domain.CheckoutFailed) {
		t.Fatalf("session status = %q want failed", row.Status)
	}
}

func TestPayCardBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false) // sandbox OFF
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (card is sandbox-only)", err)
	}
}

func TestPayMethodNotAllowedByLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"}) // card NOT allowed
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable", err)
	}
}

func TestPayUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	_, err := svc.Pay(context.Background(), "cs_nope", domain.CheckoutPayRequest{Method: "promptpay"})
	if err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}

// decimalFromMinor is a test helper mirroring money.FromMinor for assertions.
func decimalFromMinor(t *testing.T, minor int64, currency string) decimalDecimal {
	t.Helper()
	d, err := moneyFromMinor(minor, currency)
	if err != nil {
		t.Fatalf("FromMinor: %v", err)
	}
	return d
}

// ---- Task 5 tests -----------------------------------------------------------

func TestGetPromptPaySyncsToPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"})
	// QR is awaiting at pay time, then reports paid on the next Get.
	var calls int
	qr := &fakeQR{
		createFn: func(_ context.Context, _ domain.CreateQRRequest) (*domain.QRPayment, error) {
			return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "PP-EMV"}, nil
		},
		getFn: func(_ context.Context, _, id uuid.UUID) (*domain.QRPayment, error) {
			calls++
			st := domain.QRAwaitingPayment
			if calls >= 2 {
				st = domain.QRPaid
			}
			return &domain.QRPayment{ID: id, Status: st, QRPayload: "PP-EMV"}, nil
		},
	}
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"}); err != nil {
		t.Fatalf("Pay: %v", err)
	}

	// First Get: QR still awaiting -> requires_action, payload re-surfaced.
	v1, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get1: %v", err)
	}
	if v1.Status != string(domain.CheckoutRequiresAction) || v1.QRPayload != "PP-EMV" {
		t.Fatalf("v1 = %q / %q", v1.Status, v1.QRPayload)
	}
	// Second Get: QR paid -> session paid, link closed.
	v2, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get2: %v", err)
	}
	if v2.Status != string(domain.CheckoutPaid) {
		t.Fatalf("v2 status = %q want paid", v2.Status)
	}
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status = %v want [paid]", repo.linkStatusSets)
	}
}

// TestGetPromptPayPollsProviderAtMostOnce guards against the double-poll
// regression: for a still-requires_action promptpay session, Get must fetch
// the QR from the provider at most once (syncStatus's fetch reused for the
// re-surfaced payload), not twice.
func TestGetPromptPayPollsProviderAtMostOnce(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"})
	qr := &fakeQR{}
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"}); err != nil {
		t.Fatalf("Pay: %v", err)
	}
	before := qr.callCount()

	v, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v.Status != string(domain.CheckoutRequiresAction) || v.QRPayload != "EMVCO-PAYLOAD" {
		t.Fatalf("v = %q / %q, want requires_action / EMVCO-PAYLOAD", v.Status, v.QRPayload)
	}
	if got := qr.callCount() - before; got != 1 {
		t.Fatalf("qr.Get called %d times during Get(), want at most 1", got)
	}
}

func TestGetSweepsExpiredSession(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	// Force the stored session to be already expired.
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	row.ExpiresAt = pgTimestamptz(time.Now().Add(-time.Minute))
	repo.sessByHash[row.SessionTokenHash] = row
	repo.sessByID[uuid.UUID(row.ID.Bytes)] = row

	v, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v.Status != string(domain.CheckoutExpired) {
		t.Fatalf("status = %q want expired", v.Status)
	}
}

func TestGetUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.Get(context.Background(), "cs_nope"); err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}

// ---- single_use double-pay regression tests --------------------------------

// TestPayCardSingleUseLinkCannotBePaidTwice is the core regression test for the
// money-integrity bug: two checkout sessions opened against the same
// single_use link must not both be able to charge a card. The first Pay
// reserves the link (active -> paid) before charging and succeeds; the second
// Pay's reservation attempt finds the link already 'paid' and must refuse
// WITHOUT ever calling the charger.
func TestPayCardSingleUseLinkCannotBePaidTwice(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	charger := &fakeCharger{}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)

	// Two sessions from the same (still-active) single_use link — e.g. two
	// concurrent checkout tabs.
	tok1 := openSession(t, repo, svc)
	tok2 := openSession(t, repo, svc)

	view1, err := svc.Pay(context.Background(), tok1, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != nil {
		t.Fatalf("first Pay: %v", err)
	}
	if view1.Status != string(domain.CheckoutPaid) {
		t.Fatalf("first session status = %q want paid", view1.Status)
	}

	_, err = svc.Pay(context.Background(), tok2, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("second Pay err = %v want ErrCheckoutMethodUnavailable", err)
	}
	row2, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok2))
	if row2.Status != string(domain.CheckoutFailed) {
		t.Fatalf("second session status = %q want failed", row2.Status)
	}
	// The critical assertion: the card processor must have been hit exactly
	// once, not twice — this is what prevents the double charge.
	if got := charger.callCount(); got != 1 {
		t.Fatalf("charger.Create called %d times, want exactly 1 (no double charge)", got)
	}
}

// TestPayCardDeclineReleasesSingleUseLink verifies that when a reserved
// single_use link's charge is declined/errors (no money moved), the
// reservation is released back to 'active' so the link remains usable for a
// legitimate retry.
func TestPayCardDeclineReleasesSingleUseLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	firstCall := true
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		if firstCall {
			firstCall = false
			return nil, domain.ErrCardDeclined
		}
		return &domain.Payment{ID: uuid.New(), Status: domain.StatusCaptured}, nil
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok1 := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok1, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000000002", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != domain.ErrCardDeclined {
		t.Fatalf("err = %v want ErrCardDeclined", err)
	}
	row1, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok1))
	if row1.Status != string(domain.CheckoutFailed) {
		t.Fatalf("session status = %q want failed", row1.Status)
	}
	// The link must have been released back to 'active', not left stuck 'paid'
	// with no money having moved.
	link, err := repo.GetPaymentLink(context.Background(), repository.GetPaymentLinkParams{
		ID: row1.PaymentLinkID, MerchantID: row1.MerchantID,
	})
	if err != nil {
		t.Fatalf("GetPaymentLink: %v", err)
	}
	if link.Status != "active" {
		t.Fatalf("link status = %q want active (released after decline)", link.Status)
	}
	if len(repo.linkStatusSets) != 2 || repo.linkStatusSets[0] != "paid" || repo.linkStatusSets[1] != "active" {
		t.Fatalf("link status sets = %v want [paid active] (reserve then release)", repo.linkStatusSets)
	}

	// A subsequent pay against a fresh session for the same (now-active again)
	// link must succeed.
	tok2 := openSession(t, repo, svc)
	view2, err := svc.Pay(context.Background(), tok2, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != nil {
		t.Fatalf("retry Pay: %v", err)
	}
	if view2.Status != string(domain.CheckoutPaid) {
		t.Fatalf("retry status = %q want paid", view2.Status)
	}
}

// TestPayCardReusableLinkChargesWithoutConsuming verifies reusable links skip
// the reserve/consume dance entirely and can be paid repeatedly.
func TestPayCardReusableLinkChargesWithoutConsuming(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_reuse", 5000, []string{"card"})
	l.LinkType = "reusable"
	repo.putLink(l)
	charger := &fakeCharger{}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)

	for i := 0; i < 2; i++ {
		view, err := svc.CreateFromLink(context.Background(), "pl_reuse")
		if err != nil {
			t.Fatalf("CreateFromLink #%d: %v", i, err)
		}
		pview, err := svc.Pay(context.Background(), view.Token, domain.CheckoutPayRequest{
			Method: "card",
			Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
		})
		if err != nil {
			t.Fatalf("Pay #%d: %v", i, err)
		}
		if pview.Status != string(domain.CheckoutPaid) {
			t.Fatalf("Pay #%d status = %q want paid", i, pview.Status)
		}
	}
	if got := charger.callCount(); got != 2 {
		t.Fatalf("charger called %d times, want 2 (reusable link pays repeatedly)", got)
	}
	if len(repo.linkStatusSets) != 0 {
		t.Fatalf("reusable link must never reserve/consume/close, link status sets = %v", repo.linkStatusSets)
	}
}

// TestReleaseDoesNotReviveDisabledLink verifies the conditional release: if a
// reserved single_use link is disabled by the merchant mid-charge and the charge
// then declines, releasing the reservation must NOT flip the link back to
// 'active' (which would silently undo the disable). The release is a compare-and-
// swap that only reverts a link still in the reserved 'paid' state.
func TestReleaseDoesNotReviveDisabledLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	link := mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	// Mid-charge, simulate a concurrent merchant Disable flipping the (reserved,
	// now 'paid') link to 'disabled', then decline the charge.
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		_, _ = repo.UpdatePaymentLinkStatus(context.Background(), repository.UpdatePaymentLinkStatusParams{
			ID: link.ID, MerchantID: link.MerchantID, Status: "disabled",
		})
		return nil, domain.ErrCardDeclined
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000000002", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != domain.ErrCardDeclined {
		t.Fatalf("err = %v want ErrCardDeclined", err)
	}
	got, err := repo.GetPaymentLink(context.Background(), repository.GetPaymentLinkParams{ID: link.ID, MerchantID: link.MerchantID})
	if err != nil {
		t.Fatalf("GetPaymentLink: %v", err)
	}
	if got.Status != "disabled" {
		t.Fatalf("link status = %q want disabled (release must NOT revive a link disabled mid-flight)", got.Status)
	}
}

// TestPayCardFailsClosedWhenLinkMissing verifies that if a session references a
// payment link that can no longer be loaded, payCard fails closed (returns an
// error, charges nothing) rather than proceeding to charge without the reservation
// gate — a defense-in-depth guard against ever double-charging an unverifiable link.
func TestPayCardFailsClosedWhenLinkMissing(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	link := mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	charger := &fakeCharger{}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	// Simulate the link becoming unloadable after the session was created; the
	// session still references it, so payment must refuse rather than charge.
	repo.mu.Lock()
	delete(repo.linksByID, uuid.UUID(link.ID.Bytes))
	repo.mu.Unlock()

	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	}); err == nil {
		t.Fatal("expected an error when the referenced link cannot be loaded (fail closed)")
	}
	if got := charger.callCount(); got != 0 {
		t.Fatalf("charger called %d times, want 0 (must not charge when link is unverifiable)", got)
	}
}

// ---- Task: payWallet -------------------------------------------------------

func TestPayWalletRequiresActionAndReservesSingleUse(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "truemoney"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "truemoney"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.SelectedMethod != "truemoney" {
		t.Fatalf("selected_method = %q want truemoney", view.SelectedMethod)
	}
	// Wallet mock moves no money: no QR payload, no 3DS redirect URL.
	if view.QRPayload != "" || view.NextActionURL != "" {
		t.Fatalf("wallet must not set qr_payload/next_action_url: %q / %q", view.QRPayload, view.NextActionURL)
	}
	// single_use link reserved (flipped to paid) exactly like the card path.
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid] (reserved)", repo.linkStatusSets)
	}
}

func TestPayWalletBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"alipay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false) // sandbox OFF
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "alipay"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (wallet is sandbox-only)", err)
	}
	// Refused before any reservation.
	if len(repo.linkStatusSets) != 0 {
		t.Fatalf("no link reservation expected, got %v", repo.linkStatusSets)
	}
}

func TestPayWalletMethodNotAllowedByLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"}) // wallet NOT allowed
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "shopeepay"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable", err)
	}
}

func TestPayWalletReusableLinkNotReserved(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_abc", 5000, []string{"wechat"})
	l.LinkType = "reusable"
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "wechat"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if len(repo.linkStatusSets) != 0 {
		t.Fatalf("reusable link must not reserve, got %v", repo.linkStatusSets)
	}
}

func TestPayWalletSingleUseAlreadyConsumedRefuses(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_abc", 5000, []string{"truemoney"})
	l.Status = "paid" // already consumed by a prior session
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	// Open a session directly against the (now paid) link's stored row via a fresh
	// active link, then simulate the race: re-mark it paid before paying.
	// Simpler: open while active, then flip to paid, then pay.
	l.Status = "active"
	repo.putLink(l)
	tok := openSession(t, repo, svc)
	l.Status = "paid"
	repo.putLink(l)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "truemoney"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (link already consumed)", err)
	}
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	if row.Status != string(domain.CheckoutFailed) {
		t.Fatalf("session status = %q want failed", row.Status)
	}
}

// ---- Task: ConfirmMock -----------------------------------------------------

// walletRequiresAction opens a session and drives it to requires_action via a
// wallet method, returning the token.
func walletRequiresAction(t *testing.T, repo *fakeCheckoutRepo, svc CheckoutService, method string) string {
	t.Helper()
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: method}); err != nil {
		t.Fatalf("Pay(%s): %v", method, err)
	}
	return tok
}

func TestConfirmMockApproveMarksPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"truemoney"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, svc, "truemoney")

	view, err := svc.ConfirmMock(context.Background(), tok, true)
	if err != nil {
		t.Fatalf("ConfirmMock approve: %v", err)
	}
	if view.Status != string(domain.CheckoutPaid) {
		t.Fatalf("status = %q want paid", view.Status)
	}
	// Reserved at pay time -> [paid]; approve's markLinkPaid is an idempotent
	// no-op (link already paid), so the ledger stays [paid].
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid]", repo.linkStatusSets)
	}
}

func TestConfirmMockDeclineMarksFailedAndReleasesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"shopeepay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, svc, "shopeepay")

	view, err := svc.ConfirmMock(context.Background(), tok, false)
	if err != nil {
		t.Fatalf("ConfirmMock decline: %v", err)
	}
	if view.Status != string(domain.CheckoutFailed) {
		t.Fatalf("status = %q want failed", view.Status)
	}
	// Reserve then release -> [paid active].
	if len(repo.linkStatusSets) != 2 || repo.linkStatusSets[0] != "paid" || repo.linkStatusSets[1] != "active" {
		t.Fatalf("link status sets = %v want [paid active]", repo.linkStatusSets)
	}
}

func TestConfirmMockBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"alipay"})
	// Reach requires_action while sandbox is on, then confirm with a sandbox-off
	// service pointed at the same repo.
	on := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, on, "alipay")
	off := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false)

	if _, err := off.ConfirmMock(context.Background(), tok, true); err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (confirm-mock is sandbox-only)", err)
	}
}

func TestConfirmMockNonWalletSessionIsNoop(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"})
	qr := &fakeQR{} // default: awaiting_payment
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"}); err != nil {
		t.Fatalf("Pay promptpay: %v", err)
	}
	// promptpay is requires_action but NOT a wallet method: confirm-mock must not
	// flip it; it returns the current (requires_action) view unchanged.
	view, err := svc.ConfirmMock(context.Background(), tok, true)
	if err != nil {
		t.Fatalf("ConfirmMock: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action (unchanged)", view.Status)
	}
}

func TestConfirmMockUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.ConfirmMock(context.Background(), "cs_nope", true); err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}
