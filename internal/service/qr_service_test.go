package service

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/yourco/payment-gateway/internal/external"
	"github.com/yourco/payment-gateway/internal/pkg/promptpay"
	"github.com/yourco/payment-gateway/internal/repository"
)

// ---- fakeQRRepo: in-memory Querier for QR service unit tests --------------

type fakeQRRepo struct {
	repository.Querier

	mu       sync.Mutex
	rows     map[uuid.UUID]repository.QrPayment
	webhooks []repository.CreateWebhookEventParams
}

func newFakeQRRepo() *fakeQRRepo {
	return &fakeQRRepo{rows: map[uuid.UUID]repository.QrPayment{}}
}

func (f *fakeQRRepo) CreateQRPayment(_ context.Context, arg repository.CreateQRPaymentParams) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Enforce the (correlation_ref) unique index like Postgres so idempotency of
	// the reference is exercised.
	if arg.CorrelationRef != nil {
		for _, r := range f.rows {
			if r.CorrelationRef != nil && *r.CorrelationRef == *arg.CorrelationRef {
				return repository.QrPayment{}, errors.New("duplicate correlation_ref")
			}
		}
	}
	row := repository.QrPayment{
		ID:             arg.ID,
		MerchantID:     arg.MerchantID,
		Method:         arg.Method,
		AmountMinor:    arg.AmountMinor,
		Currency:       arg.Currency,
		Status:         arg.Status,
		Reference:      arg.Reference,
		QrPayload:      arg.QrPayload,
		QrImageUrl:     arg.QrImageUrl,
		ProviderRef:    arg.ProviderRef,
		CorrelationRef: arg.CorrelationRef,
		ExpiresAt:      arg.ExpiresAt,
		CreatedAt:      pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
	f.rows[key(arg.ID)] = row
	return row, nil
}

func (f *fakeQRRepo) GetQRByProviderRef(_ context.Context, providerRef *string) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if providerRef == nil {
		return repository.QrPayment{}, pgx.ErrNoRows
	}
	for _, r := range f.rows {
		if r.ProviderRef != nil && *r.ProviderRef == *providerRef {
			return r, nil
		}
	}
	return repository.QrPayment{}, pgx.ErrNoRows
}

func (f *fakeQRRepo) GetQRByCorrelationRef(_ context.Context, correlationRef *string) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if correlationRef == nil {
		return repository.QrPayment{}, pgx.ErrNoRows
	}
	for _, r := range f.rows {
		if r.CorrelationRef != nil && *r.CorrelationRef == *correlationRef {
			return r, nil
		}
	}
	return repository.QrPayment{}, pgx.ErrNoRows
}

func (f *fakeQRRepo) GetQRPaymentForMerchant(_ context.Context, arg repository.GetQRPaymentForMerchantParams) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[key(arg.ID)]
	if !ok || r.MerchantID != arg.MerchantID {
		return repository.QrPayment{}, pgx.ErrNoRows
	}
	return r, nil
}

func (f *fakeQRRepo) UpdateQRStatus(_ context.Context, arg repository.UpdateQRStatusParams) (repository.QrPayment, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[key(arg.ID)]
	if !ok {
		return repository.QrPayment{}, pgx.ErrNoRows
	}
	r.Status = arg.Status
	r.PayerBank = arg.PayerBank
	r.ProviderRef = arg.ProviderRef
	r.PaidAt = arg.PaidAt
	f.rows[key(arg.ID)] = r
	return r, nil
}

func (f *fakeQRRepo) CreateWebhookEvent(_ context.Context, arg repository.CreateWebhookEventParams) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.webhooks = append(f.webhooks, arg)
	return nil
}

func (f *fakeQRRepo) webhookCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.webhooks)
}

// ---- test wiring ----------------------------------------------------------

const qrSecret = "qr-webhook-test-secret-at-least-32-bytes!"

func newQRService(repo repository.Querier) (QRService, *external.MockQRProvider) {
	provider := external.NewMockQRProvider(qrSecret, "")
	target := promptpay.Target{MobileNo: "0812345678"}
	return NewQRService(repo, nil, provider, target, zerolog.Nop()), provider
}

func promptpayReq(ref string) domain.CreateQRRequest {
	return domain.CreateQRRequest{
		MerchantID: uuid.New(),
		Method:     domain.QRPromptPayDynamic,
		Amount:     decimal.RequireFromString("100.00"),
		Currency:   "THB",
		Reference:  ref,
	}
}

// signedWebhook marshals a QRWebhook and returns the body + a valid signature.
func signedWebhook(t *testing.T, wh domain.QRWebhook) (body []byte, sig string) {
	t.Helper()
	body, err := json.Marshal(wh)
	if err != nil {
		t.Fatalf("marshal webhook: %v", err)
	}
	sig = external.ComputeQRWebhookSignature(qrSecret, body)
	return body, sig
}

// ---- Create embeds a correlation ref for PromptPay ------------------------

func TestPromptPayCreateEmbedsCorrelationRef(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)

	qp, err := svc.Create(context.Background(), promptpayReq("ORDER-123"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// correlation_ref must be a UNIQUE generated token (it has a unique index),
	// not the reusable merchant reference — otherwise two QR with the same
	// merchant reference would collide.
	if !strings.HasPrefix(qp.CorrelationRef, "PP") || qp.CorrelationRef == "ORDER-123" {
		t.Fatalf("correlation_ref=%q want a unique PP<...> token", qp.CorrelationRef)
	}
	// The EMVCo tag-62 payload must carry the correlation_ref so the bank echoes it.
	if !strings.Contains(qp.QRPayload, qp.CorrelationRef) {
		t.Fatalf("payload does not embed the correlation_ref %q: %s", qp.CorrelationRef, qp.QRPayload)
	}
}

func TestPromptPayCreateGeneratesRefWhenMerchantRefEmpty(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)

	qp, err := svc.Create(context.Background(), promptpayReq(""))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if qp.CorrelationRef == "" {
		t.Fatal("expected a generated correlation_ref when merchant ref is empty")
	}
	if !strings.Contains(qp.QRPayload, qp.CorrelationRef) {
		t.Fatalf("payload must embed the generated correlation_ref %q: %s", qp.CorrelationRef, qp.QRPayload)
	}
}

// ---- PromptPay create -> webhook correlate -> paid ------------------------

func TestPromptPayWebhookCorrelatesToPaid(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)

	qp, err := svc.Create(context.Background(), promptpayReq("PPREF-9"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	// The bank echoes our embedded correlation_ref back as `reference`, plus its
	// own txn id as provider_ref.
	body, sig := signedWebhook(t, domain.QRWebhook{
		Reference:   qp.CorrelationRef,
		ProviderRef: "BANK-TXN-0001",
		Amount:      decimal.RequireFromString("100.00"),
		Currency:    "THB",
		PayerBank:   "SCB",
		Status:      string(domain.QRPaid),
		PaidAt:      time.Now(),
	})

	confirmed, err := svc.ConfirmFromWebhook(context.Background(), sig, body)
	if err != nil {
		t.Fatalf("ConfirmFromWebhook: %v", err)
	}
	if confirmed.ID != qp.ID {
		t.Fatalf("correlated to wrong QR: %s != %s", confirmed.ID, qp.ID)
	}
	if confirmed.Status != domain.QRPaid {
		t.Fatalf("status=%s want paid", confirmed.Status)
	}
	if repo.webhookCount() != 1 {
		t.Fatalf("expected exactly one enqueued merchant webhook, got %d", repo.webhookCount())
	}

	// Idempotent replay: same confirmation -> no second merchant webhook.
	if _, err := svc.ConfirmFromWebhook(context.Background(), sig, body); err != nil {
		t.Fatalf("replay ConfirmFromWebhook: %v", err)
	}
	if repo.webhookCount() != 1 {
		t.Fatalf("replay enqueued a duplicate webhook: count=%d", repo.webhookCount())
	}
}

// ---- unknown / invalid reference is rejected ------------------------------

func TestPromptPayWebhookUnknownReferenceRejected(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)

	if _, err := svc.Create(context.Background(), promptpayReq("KNOWN")); err != nil {
		t.Fatalf("Create: %v", err)
	}

	body, sig := signedWebhook(t, domain.QRWebhook{
		Reference:   "DOES-NOT-EXIST",
		ProviderRef: "BANK-TXN-X",
		Amount:      decimal.RequireFromString("100.00"),
		Currency:    "THB",
		Status:      string(domain.QRPaid),
	})
	if _, err := svc.ConfirmFromWebhook(context.Background(), sig, body); !errors.Is(err, domain.ErrPaymentNotFound) {
		t.Fatalf("err=%v want ErrPaymentNotFound", err)
	}
}

func TestPromptPayWebhookBadSignatureRejected(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)
	if _, err := svc.Create(context.Background(), promptpayReq("SIG")); err != nil {
		t.Fatalf("Create: %v", err)
	}
	body, _ := signedWebhook(t, domain.QRWebhook{Reference: "SIG", Status: string(domain.QRPaid)})
	if _, err := svc.ConfirmFromWebhook(context.Background(), "deadbeef", body); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest (bad signature)", err)
	}
}

func TestPromptPayWebhookNoHandleRejected(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)
	// Neither provider_ref nor reference -> nothing to correlate against.
	body, sig := signedWebhook(t, domain.QRWebhook{Amount: decimal.RequireFromString("100.00"), Currency: "THB", Status: string(domain.QRPaid)})
	if _, err := svc.ConfirmFromWebhook(context.Background(), sig, body); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest (no correlation handle)", err)
	}
}

// ---- amount mismatch is rejected ------------------------------------------

func TestPromptPayWebhookAmountMismatchRejected(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)
	qp, err := svc.Create(context.Background(), promptpayReq("AMT"))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	body, sig := signedWebhook(t, domain.QRWebhook{
		Reference: qp.CorrelationRef, // correlate to the minted QR, then mismatch on amount
		Amount:    decimal.RequireFromString("999.00"), // != minted 100.00
		Currency:  "THB",
		Status:    string(domain.QRPaid),
	})
	if _, err := svc.ConfirmFromWebhook(context.Background(), sig, body); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("err=%v want ErrInvalidRequest (amount mismatch)", err)
	}
}

// ---- card-scheme rail still correlates via provider_ref -------------------

func TestCardSchemeWebhookCorrelatesViaProviderRef(t *testing.T) {
	repo := newFakeQRRepo()
	svc, _ := newQRService(repo)

	req := domain.CreateQRRequest{
		MerchantID: uuid.New(),
		Method:     domain.QRCardScheme,
		Amount:     decimal.RequireFromString("50.00"),
		Currency:   "THB",
		Reference:  "CARD-1",
	}
	qp, err := svc.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if qp.ProviderRef == "" {
		t.Fatal("card-scheme QR must carry a provider_ref")
	}

	body, sig := signedWebhook(t, domain.QRWebhook{
		ProviderRef: qp.ProviderRef,
		Amount:      decimal.RequireFromString("50.00"),
		Currency:    "THB",
		Status:      string(domain.QRPaid),
	})
	confirmed, err := svc.ConfirmFromWebhook(context.Background(), sig, body)
	if err != nil {
		t.Fatalf("ConfirmFromWebhook: %v", err)
	}
	if confirmed.Status != domain.QRPaid || confirmed.ID != qp.ID {
		t.Fatalf("card-scheme correlation failed: %+v", confirmed)
	}
}
