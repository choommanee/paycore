package worker

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// fakeWebhookRepo implements the slice of repository.Querier the delivery worker
// touches. It serves one pending event and records the delivered id.
type fakeWebhookRepo struct {
	repository.Querier

	mu        sync.Mutex
	pending   []repository.WebhookEvent
	secrets   map[uuid.UUID][]byte // merchant_id -> webhook_secret_enc
	delivered []pgtype.UUID
}

func (f *fakeWebhookRepo) ListPendingWebhooks(_ context.Context, _ int32) ([]repository.WebhookEvent, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := f.pending
	f.pending = nil // deliver-once
	return out, nil
}

func (f *fakeWebhookRepo) GetMerchantWebhookSigningKey(_ context.Context, id pgtype.UUID) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	mid, err := uuid.FromBytes(id.Bytes[:])
	if err != nil {
		return nil, pgx.ErrNoRows
	}
	enc, ok := f.secrets[mid]
	if !ok {
		return nil, pgx.ErrNoRows
	}
	return enc, nil
}

func (f *fakeWebhookRepo) MarkWebhookDelivered(_ context.Context, id pgtype.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delivered = append(f.delivered, id)
	return nil
}

// fakeOpener is an identity "decryptor": the stored blob is the raw secret. This
// keeps the worker test independent of the concrete crypto implementation while
// still exercising the per-merchant key path.
type fakeOpener struct{}

func (fakeOpener) Open(_ context.Context, ciphertext []byte) ([]byte, error) {
	return ciphertext, nil
}

func pgUUID(u uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: u, Valid: true} }

func verifyV1(t *testing.T, secret, sigHeader string, ts string, body []byte) {
	t.Helper()
	want := "t=" + ts + ",v1=" + hmacHexStr(secret, ts+"."+string(body))
	if sigHeader != want {
		t.Fatalf("X-PayCore-Signature = %q, not signed with the expected secret (%q)", sigHeader, want)
	}
}

func hmacHexStr(secret, msg string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}

// TestDeliverSignsWithMerchantSecret: a merchant with its own webhook secret has
// its delivery signed with THAT secret (verifiable with the whsec_ it was
// issued), not the global key.
func TestDeliverSignsWithMerchantSecret(t *testing.T) {
	merchantSecret := "whsec_merchant_specific_secret"
	body := []byte(`{"event":"payment.paid","id":"evt_1"}`)

	var gotSig, gotTs string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-PayCore-Signature")
		gotTs = r.Header.Get("X-PayCore-Timestamp")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mid := uuid.New()
	url := srv.URL
	repo := &fakeWebhookRepo{
		secrets: map[uuid.UUID][]byte{mid: []byte(merchantSecret)},
		pending: []repository.WebhookEvent{{
			ID:         pgUUID(uuid.New()),
			MerchantID: pgUUID(mid),
			EventType:  "payment.paid",
			Payload:    body,
			TargetUrl:  &url,
		}},
	}
	w := NewWebhookWorker(repo, fakeOpener{}, "global_secret", "", time.Second, 5, zerolog.Nop())

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(repo.delivered) != 1 {
		t.Fatalf("expected 1 delivered event, got %d", len(repo.delivered))
	}
	// The signature must verify with the MERCHANT'S secret, and must NOT match the
	// global secret.
	verifyV1(t, merchantSecret, gotSig, gotTs, body)
	globalWant := "t=" + gotTs + ",v1=" + hmacHexStr("global_secret", gotTs+"."+string(body))
	if gotSig == globalWant {
		t.Fatal("delivery was signed with the global secret, not the merchant's")
	}
}

// TestDeliverFallsBackToGlobalSecret: a merchant with no stored secret is signed
// with the global key (backward compat / default endpoint).
func TestDeliverFallsBackToGlobalSecret(t *testing.T) {
	body := []byte(`{"event":"payment.paid","id":"evt_2"}`)

	var gotSig, gotTs string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotSig = r.Header.Get("X-PayCore-Signature")
		gotTs = r.Header.Get("X-PayCore-Timestamp")
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	mid := uuid.New()
	url := srv.URL
	repo := &fakeWebhookRepo{
		secrets: map[uuid.UUID][]byte{}, // none for this merchant
		pending: []repository.WebhookEvent{{
			ID:         pgUUID(uuid.New()),
			MerchantID: pgUUID(mid),
			EventType:  "payment.paid",
			Payload:    body,
			TargetUrl:  &url,
		}},
	}
	w := NewWebhookWorker(repo, fakeOpener{}, "global_secret", "", time.Second, 5, zerolog.Nop())

	if err := w.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	verifyV1(t, "global_secret", gotSig, gotTs, body)
}
