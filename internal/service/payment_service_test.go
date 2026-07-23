package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/repository"
)

// ---- fakeRepo: in-memory Querier for service unit tests -------------------

// fakeRepo implements just enough of repository.Querier to drive the payment
// service without a database. Unimplemented methods are inherited from the
// embedded nil Querier and will panic if unexpectedly called (surfacing a test
// gap rather than silently passing).
type fakeRepo struct {
	repository.Querier

	mu       sync.Mutex
	payments map[uuid.UUID]repository.Payment
	byIdem   map[string]uuid.UUID
	ledger   []repository.InsertLedgerEntryParams
	audits   []repository.WriteAuditLogParams
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		payments: map[uuid.UUID]repository.Payment{},
		byIdem:   map[string]uuid.UUID{},
	}
}

func key(p pgtype.UUID) uuid.UUID { return uuid.UUID(p.Bytes) }

func (f *fakeRepo) CreatePayment(_ context.Context, arg repository.CreatePaymentParams) (repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Enforce the (merchant_id, idempotency_key) unique constraint like Postgres,
	// so tests exercise the concurrent-duplicate reservation path.
	if arg.IdempotencyKey != "" {
		if _, exists := f.byIdem[arg.IdempotencyKey]; exists {
			return repository.Payment{}, &pgconn.PgError{Code: "23505"}
		}
	}
	p := repository.Payment{
		ID:             arg.ID,
		MerchantID:     arg.MerchantID,
		AmountMinor:    arg.AmountMinor,
		Currency:       arg.Currency,
		Status:         arg.Status,
		CardBrand:      arg.CardBrand,
		CardLast4:      arg.CardLast4,
		Reference:      arg.Reference,
		IdempotencyKey: arg.IdempotencyKey,
	}
	f.payments[key(arg.ID)] = p
	if arg.IdempotencyKey != "" {
		f.byIdem[arg.IdempotencyKey] = key(arg.ID)
	}
	return p, nil
}

func (f *fakeRepo) UpdatePaymentStatus(_ context.Context, arg repository.UpdatePaymentStatusParams) (repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.payments[key(arg.ID)]
	if !ok {
		return repository.Payment{}, pgx.ErrNoRows
	}
	p.Status = arg.Status
	p.CapturedAmountMinor = arg.CapturedAmountMinor
	p.RefundedAmountMinor = arg.RefundedAmountMinor
	p.AcquirerRef = arg.AcquirerRef
	p.AuthCode = arg.AuthCode
	f.payments[key(arg.ID)] = p
	return p, nil
}

func (f *fakeRepo) GetPayment(_ context.Context, id pgtype.UUID) (repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.payments[key(id)]
	if !ok {
		return repository.Payment{}, pgx.ErrNoRows
	}
	return p, nil
}

func (f *fakeRepo) GetPaymentForMerchant(_ context.Context, arg repository.GetPaymentForMerchantParams) (repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p, ok := f.payments[key(arg.ID)]
	if !ok || p.MerchantID != arg.MerchantID {
		// Scoped miss (wrong merchant) is indistinguishable from not-found.
		return repository.Payment{}, pgx.ErrNoRows
	}
	return p, nil
}

func (f *fakeRepo) GetByIdempotencyKey(_ context.Context, arg repository.GetByIdempotencyKeyParams) (repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	id, ok := f.byIdem[arg.IdempotencyKey]
	if !ok {
		return repository.Payment{}, pgx.ErrNoRows
	}
	return f.payments[id], nil
}

func (f *fakeRepo) InsertLedgerEntry(_ context.Context, arg repository.InsertLedgerEntryParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ledger = append(f.ledger, arg)
	return nil
}

func (f *fakeRepo) WriteAuditLog(_ context.Context, arg repository.WriteAuditLogParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.audits = append(f.audits, arg)
	return nil
}

func (f *fakeRepo) ListPaymentsByMerchant(_ context.Context, arg repository.ListPaymentsByMerchantParams) ([]repository.Payment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []repository.Payment
	for _, p := range f.payments {
		if p.MerchantID == arg.MerchantID {
			out = append(out, p)
		}
	}
	return out, nil
}

func (f *fakeRepo) auditActions() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]string, len(f.audits))
	for i, a := range f.audits {
		out[i] = a.Action
	}
	return out
}

// stubVault returns a fixed PAN regardless of token so the acquirer path runs.
type stubVault struct{ pan string }

func (s stubVault) Detokenize(_ context.Context, _ string) (string, error) {
	if s.pan == "" {
		return "", errors.New("no pan")
	}
	return s.pan, nil
}

func newService(t *testing.T, repo *fakeRepo, pan string) PaymentService {
	t.Helper()
	return NewPaymentService(
		repo,
		nil, // no TxManager -> runs directly against repo
		stubVault{pan: pan},
		external.NewMockAcquirer(),
		external.NewMockThreeDS(""),
		zerolog.Nop(),
	)
}

func baseReq(pan string) domain.CreatePaymentRequest {
	return domain.CreatePaymentRequest{
		MerchantID:   uuid.New(),
		Amount:       decimal.RequireFromString("100.00"),
		Currency:     "THB",
		PaymentToken: "tok_" + pan,
	}
}

// ---- assertTransition (state machine) ------------------------------------

func TestAssertTransition(t *testing.T) {
	valid := []struct{ from, to domain.PaymentStatus }{
		{domain.StatusRequiresAction, domain.StatusAuthorized},
		{domain.StatusRequiresAction, domain.StatusFailed},
		{domain.StatusAuthorized, domain.StatusCaptured},
		{domain.StatusAuthorized, domain.StatusVoided},
		{domain.StatusCaptured, domain.StatusPartialRefunded},
		{domain.StatusCaptured, domain.StatusRefunded},
		{domain.StatusPartialRefunded, domain.StatusPartialRefunded},
		{domain.StatusPartialRefunded, domain.StatusRefunded},
	}
	for _, tc := range valid {
		if err := assertTransition(tc.from, tc.to); err != nil {
			t.Errorf("expected %s->%s valid, got %v", tc.from, tc.to, err)
		}
	}

	invalid := []struct{ from, to domain.PaymentStatus }{
		{domain.StatusAuthorized, domain.StatusRefunded},
		{domain.StatusCaptured, domain.StatusVoided},
		{domain.StatusCaptured, domain.StatusAuthorized},
		{domain.StatusVoided, domain.StatusCaptured},
		{domain.StatusRefunded, domain.StatusCaptured},
		{domain.StatusFailed, domain.StatusAuthorized},
		{domain.StatusRequiresAction, domain.StatusCaptured},
		{domain.StatusAuthorized, domain.StatusAuthorized},
	}
	for _, tc := range invalid {
		if err := assertTransition(tc.from, tc.to); !errors.Is(err, domain.ErrInvalidState) {
			t.Errorf("expected %s->%s invalid, got %v", tc.from, tc.to, err)
		}
	}
}

// ---- Create: authorize-only, decline, 3DS, idempotency -------------------

func TestCreateAuthorizeOnly(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p, err := svc.Create(context.Background(), "idem-1", baseReq("4111111111111111"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Status != domain.StatusAuthorized {
		t.Fatalf("status=%s want authorized", p.Status)
	}
	if p.AcquirerRef == "" || p.AuthCode == "" {
		t.Fatal("authorized payment must carry acquirer ref + auth code")
	}
	// Audit + ledger written.
	if got := repo.auditActions(); len(got) == 0 || got[len(got)-1] != "payment.authorized" {
		t.Fatalf("audit actions=%v want ...payment.authorized", got)
	}
	if len(repo.ledger) != 1 || repo.ledger[0].EntryType != "authorize" {
		t.Fatalf("expected one authorize ledger entry, got %+v", repo.ledger)
	}
}

func TestCreateChargeCaptures(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	req := baseReq("4111111111111111")
	req.Capture = true
	p, err := svc.Create(context.Background(), "idem-charge", req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Status != domain.StatusCaptured {
		t.Fatalf("status=%s want captured", p.Status)
	}
	actions := repo.auditActions()
	if !contains(actions, "payment.authorized") || !contains(actions, "payment.captured") {
		t.Fatalf("expected authorize+capture audit, got %v", actions)
	}
}

// TestCreate3DSAuthFailDeclined: for last4 "0000" the mock 3DS provider fails
// authentication (card not enrolled), which the service maps to a card decline
// before it ever reaches the acquirer.
func TestCreate3DSAuthFailDeclined(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4000000000000000")
	_, err := svc.Create(context.Background(), "idem-3dsfail", baseReq("4000000000000000"))
	if !errors.Is(err, domain.ErrCardDeclined) {
		t.Fatalf("err=%v want ErrCardDeclined", err)
	}
}

// TestCreateInsufficientFundsDirect exercises the acquirer decline path by
// wiring a service with no 3DS provider, so authorization reaches the acquirer
// which declines a last4 "0000" PAN for insufficient funds. The failed attempt
// is persisted and audited.
func TestCreateInsufficientFundsDirect(t *testing.T) {
	repo := newFakeRepo()
	svc := NewPaymentService(
		repo, nil, stubVault{pan: "4000000000000000"},
		external.NewMockAcquirer(),
		nil, // no 3DS -> go straight to the acquirer
		zerolog.Nop(),
	)
	_, err := svc.Create(context.Background(), "idem-decline", baseReq("4000000000000000"))
	if !errors.Is(err, domain.ErrInsufficientFund) {
		t.Fatalf("err=%v want ErrInsufficientFund", err)
	}
	if !contains(repo.auditActions(), "payment.failed") {
		t.Fatalf("expected payment.failed audit, got %v", repo.auditActions())
	}
}

func TestCreate3DSRequiresAction(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4000000000000119") // last4 0119 -> 3ds challenge
	p, err := svc.Create(context.Background(), "idem-3ds", baseReq("4000000000000119"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p.Status != domain.StatusRequiresAction {
		t.Fatalf("status=%s want requires_action", p.Status)
	}
	if p.NextActionURL == "" {
		t.Fatal("requires_action payment must carry next_action_url")
	}
}

// TestCreateIdempotentReplay verifies a replay with the same key returns the
// stored payment without a second charge (dedup logic).
func TestCreateIdempotentReplay(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	req := baseReq("4111111111111111")

	p1, err := svc.Create(context.Background(), "dedup-key", req)
	if err != nil {
		t.Fatal(err)
	}
	auditsAfterFirst := len(repo.auditActions())

	p2, err := svc.Create(context.Background(), "dedup-key", req)
	if err != nil {
		t.Fatal(err)
	}
	if p1.ID != p2.ID {
		t.Fatalf("replay returned different payment id: %s != %s", p1.ID, p2.ID)
	}
	if got := len(repo.auditActions()); got != auditsAfterFirst {
		t.Fatalf("replay wrote new audit rows: before=%d after=%d", auditsAfterFirst, got)
	}
}

// countingAcquirer wraps MockAcquirer and counts Authorize calls so a test can
// assert the card network is contacted at most once per idempotency key.
type countingAcquirer struct {
	external.Acquirer
	mu    sync.Mutex
	auths int
}

func (c *countingAcquirer) Authorize(ctx context.Context, req external.AuthRequest) (*external.AuthResult, error) {
	c.mu.Lock()
	c.auths++
	c.mu.Unlock()
	return c.Acquirer.Authorize(ctx, req)
}

// TestCreateConcurrentDuplicateChargesOnce verifies that two concurrent Create
// calls sharing an idempotency key result in exactly one acquirer authorization
// (no double charge) and one persisted payment. The reservation insert makes the
// loser fail on the unique constraint before it reaches the acquirer.
func TestCreateConcurrentDuplicateChargesOnce(t *testing.T) {
	repo := newFakeRepo()
	acq := &countingAcquirer{Acquirer: external.NewMockAcquirer()}
	svc := NewPaymentService(
		repo, nil, stubVault{pan: "4111111111111111"},
		acq,
		nil, // no 3DS -> straight to acquirer
		zerolog.Nop(),
	)
	req := baseReq("4111111111111111")

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = svc.Create(context.Background(), "race-key", req)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("Create[%d] returned error: %v", i, err)
		}
	}
	acq.mu.Lock()
	auths := acq.auths
	acq.mu.Unlock()
	if auths != 1 {
		t.Fatalf("acquirer Authorize called %d times, want exactly 1 (double charge!)", auths)
	}
	if n := len(repo.payments); n != 1 {
		t.Fatalf("persisted %d payment rows, want 1", n)
	}
}

// ---- Capture / Void / Refund flows + invalid transitions -----------------

func authorize(t *testing.T, repo *fakeRepo, svc PaymentService, idem string) *domain.Payment {
	t.Helper()
	req := baseReq("4111111111111111")
	p, err := svc.Create(context.Background(), idem, req)
	if err != nil {
		t.Fatalf("authorize setup: %v", err)
	}
	return p
}

func TestCaptureThenRefundFull(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "cap-1")

	cap, err := svc.Capture(context.Background(), p.MerchantID, p.ID, domain.CaptureRequest{Amount: decimal.RequireFromString("100.00")})
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if cap.Status != domain.StatusCaptured {
		t.Fatalf("status=%s want captured", cap.Status)
	}

	ref, err := svc.Refund(context.Background(), p.MerchantID, p.ID, domain.RefundRequest{Amount: decimal.RequireFromString("100.00")})
	if err != nil {
		t.Fatalf("Refund: %v", err)
	}
	if ref.Status != domain.StatusRefunded {
		t.Fatalf("status=%s want refunded", ref.Status)
	}
}

func TestPartialRefund(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "pref-1")

	if _, err := svc.Capture(context.Background(), p.MerchantID, p.ID, domain.CaptureRequest{Amount: decimal.RequireFromString("100.00")}); err != nil {
		t.Fatal(err)
	}
	r1, err := svc.Refund(context.Background(), p.MerchantID, p.ID, domain.RefundRequest{Amount: decimal.RequireFromString("40.00")})
	if err != nil {
		t.Fatal(err)
	}
	if r1.Status != domain.StatusPartialRefunded {
		t.Fatalf("status=%s want partial_refunded", r1.Status)
	}
	// Remaining 60 refunds to fully refunded.
	r2, err := svc.Refund(context.Background(), p.MerchantID, p.ID, domain.RefundRequest{Amount: decimal.RequireFromString("60.00")})
	if err != nil {
		t.Fatal(err)
	}
	if r2.Status != domain.StatusRefunded {
		t.Fatalf("status=%s want refunded", r2.Status)
	}
}

func TestRefundExceedsCaptured(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "rex-1")
	if _, err := svc.Capture(context.Background(), p.MerchantID, p.ID, domain.CaptureRequest{Amount: decimal.RequireFromString("100.00")}); err != nil {
		t.Fatal(err)
	}
	_, err := svc.Refund(context.Background(), p.MerchantID, p.ID, domain.RefundRequest{Amount: decimal.RequireFromString("150.00")})
	if !errors.Is(err, domain.ErrRefundExceeds) {
		t.Fatalf("err=%v want ErrRefundExceeds", err)
	}
}

func TestVoidAuthorized(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "void-1")
	v, err := svc.Void(context.Background(), p.MerchantID, p.ID)
	if err != nil {
		t.Fatalf("Void: %v", err)
	}
	if v.Status != domain.StatusVoided {
		t.Fatalf("status=%s want voided", v.Status)
	}
}

// TestVoidAfterCaptureInvalid confirms an invalid transition is rejected.
func TestVoidAfterCaptureInvalid(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "void-2")
	if _, err := svc.Capture(context.Background(), p.MerchantID, p.ID, domain.CaptureRequest{Amount: decimal.RequireFromString("100.00")}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Void(context.Background(), p.MerchantID, p.ID); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("err=%v want ErrInvalidState", err)
	}
}

func TestCaptureOverAmountRejected(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "over-1")
	_, err := svc.Capture(context.Background(), p.MerchantID, p.ID, domain.CaptureRequest{Amount: decimal.RequireFromString("200.00")})
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest", err)
	}
}

func TestGetNotFound(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	if _, err := svc.Get(context.Background(), uuid.New(), uuid.New()); !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Fatalf("err=%v want ErrPaymentNotFound", err)
	}
}

func TestCreateInvalidAmount(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	req := baseReq("4111111111111111")
	req.Amount = decimal.RequireFromString("0")
	if _, err := svc.Create(context.Background(), "z", req); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest", err)
	}
}

// ---- metrics observer hook -----------------------------------------------

type countingObserver struct {
	mu     sync.Mutex
	counts map[string]int
}

func (c *countingObserver) ObservePaymentStatus(status string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.counts == nil {
		c.counts = map[string]int{}
	}
	c.counts[status]++
}

func TestObservePaymentStatusHook(t *testing.T) {
	repo := newFakeRepo()
	base := newService(t, repo, "4111111111111111")
	obs := &countingObserver{}
	svc := base.(interface {
		WithMetrics(StatusObserver) PaymentService
	}).WithMetrics(obs)

	if _, err := svc.Create(context.Background(), "m-1", baseReq("4111111111111111")); err != nil {
		t.Fatal(err)
	}
	obs.mu.Lock()
	defer obs.mu.Unlock()
	if obs.counts[string(domain.StatusAuthorized)] != 1 {
		t.Fatalf("expected 1 authorized observation, got %v", obs.counts)
	}
}

func TestList(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	req := baseReq("4111111111111111")
	// Two payments for the same merchant.
	if _, err := svc.Create(context.Background(), "l-1", req); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(context.Background(), "l-2", req); err != nil {
		t.Fatal(err)
	}
	got, err := svc.List(context.Background(), req.MerchantID, 10, 0)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("List returned %d payments, want 2", len(got))
	}
}

// TestHandleThreeDSResultFailure drives a requires_action payment to failed when
// the shopper fails the challenge (ok=false).
func TestHandleThreeDSResultFailure(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4000000000000119")
	p, err := svc.Create(context.Background(), "tds-1", baseReq("4000000000000119"))
	if err != nil {
		t.Fatal(err)
	}
	if p.Status != domain.StatusRequiresAction {
		t.Fatalf("setup status=%s want requires_action", p.Status)
	}
	resolved, err := svc.HandleThreeDSResult(context.Background(), p.MerchantID, p.ID, false)
	if err != nil {
		t.Fatalf("HandleThreeDSResult: %v", err)
	}
	if resolved.Status != domain.StatusFailed {
		t.Fatalf("status=%s want failed", resolved.Status)
	}
	if !contains(repo.auditActions(), "payment.failed") {
		t.Fatalf("expected payment.failed audit, got %v", repo.auditActions())
	}
}

// TestHandleThreeDSResultSuccessNoToken: on success the service re-authorizes,
// but the single-use token is not stored on the operational row, so
// detokenizeFor3DS fails and the payment is declined (never fabricates a PAN).
func TestHandleThreeDSResultSuccessNoToken(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4000000000000119")
	p, err := svc.Create(context.Background(), "tds-2", baseReq("4000000000000119"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.HandleThreeDSResult(context.Background(), p.MerchantID, p.ID, true); !errors.Is(err, domain.ErrCardDeclined) {
		t.Fatalf("err=%v want ErrCardDeclined (no stored token)", err)
	}
}

// TestHandleThreeDSResultInvalidState rejects resolving a payment that is not in
// requires_action.
func TestHandleThreeDSResultInvalidState(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	p := authorize(t, repo, svc, "tds-3") // authorized, not requires_action
	if _, err := svc.HandleThreeDSResult(context.Background(), p.MerchantID, p.ID, true); !errors.Is(err, domain.ErrInvalidState) {
		t.Fatalf("err=%v want ErrInvalidState", err)
	}
}

// TestCreateIdempotencyKeyReuseRejected verifies that replaying an idempotency
// key with a DIFFERENT payload (different amount) is rejected with
// ErrIdempotencyKeyReuse instead of silently returning the stored payment.
func TestCreateIdempotencyKeyReuseRejected(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	req := baseReq("4111111111111111")

	if _, err := svc.Create(context.Background(), "reuse-key", req); err != nil {
		t.Fatalf("first Create: %v", err)
	}

	// Same key, different amount -> must be rejected, not silently deduped.
	req2 := req
	req2.Amount = decimal.RequireFromString("250.00")
	_, err := svc.Create(context.Background(), "reuse-key", req2)
	if !errors.Is(err, domain.ErrIdempotencyKeyReuse) {
		t.Fatalf("err=%v want ErrIdempotencyKeyReuse", err)
	}
}

// TestVerifyThreeDSResult confirms the CRes verification fails closed and only
// accepts a correctly-signed value with transStatus=Y.
func TestVerifyThreeDSResult(t *testing.T) {
	repo := newFakeRepo()
	base := newService(t, repo, "4111111111111111")
	svc := base.(interface {
		WithThreeDSSecret(string) PaymentService
	}).WithThreeDSSecret("test-3ds-secret-at-least-32-bytes-long!!")

	pid := uuid.New()

	// No CRes -> false even with transStatus=Y (the core fix: bare Y is not enough).
	if svc.VerifyThreeDSResult(pid, "", "Y") {
		t.Fatal("empty CRes must not verify")
	}
	// Wrong CRes -> false.
	if svc.VerifyThreeDSResult(pid, "deadbeef", "Y") {
		t.Fatal("wrong CRes must not verify")
	}

	// Correctly-signed CRes -> true only when transStatus=Y.
	mac := hmacSHA256Hex("test-3ds-secret-at-least-32-bytes-long!!", pid)
	if !svc.VerifyThreeDSResult(pid, mac, "Y") {
		t.Fatal("valid CRes with transStatus=Y must verify")
	}
	if svc.VerifyThreeDSResult(pid, mac, "N") {
		t.Fatal("valid CRes but transStatus!=Y must not verify")
	}
}

// TestVerifyThreeDSResultNoSecretFailsClosed verifies that with no secret
// configured the endpoint can never authorize (fails closed).
func TestVerifyThreeDSResultNoSecretFailsClosed(t *testing.T) {
	repo := newFakeRepo()
	svc := newService(t, repo, "4111111111111111")
	if svc.VerifyThreeDSResult(uuid.New(), "anything", "Y") {
		t.Fatal("with no secret configured verification must fail closed")
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}

// hmacSHA256Hex mirrors the service's CRes derivation for tests.
func hmacSHA256Hex(secret string, id uuid.UUID) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(id[:])
	return hex.EncodeToString(mac.Sum(nil))
}
