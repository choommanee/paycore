package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// secretOpener decrypts a merchant's envelope-encrypted webhook signing secret.
// Satisfied by crypto.SecretBox. Kept as a narrow interface so the worker does
// not depend on the crypto package directly.
type secretOpener interface {
	Open(ctx context.Context, ciphertext []byte) ([]byte, error)
}

// WebhookWorker delivers pending webhook_events to merchant endpoints. Each POST
// is signed with an HMAC-SHA256 of the raw JSON body so merchants can verify
// authenticity. Each delivery is signed with the TARGET MERCHANT'S own signing
// secret (decrypted from webhook_secret_enc) when one is set, so a merchant can
// verify with the whsec_ returned by SetWebhook (Stripe-style per-tenant key
// isolation); deliveries for merchants with no secret fall back to the global
// signingKey. Delivery is retried with exponential backoff up to maxAttempts,
// after which the event is marked failed.
type WebhookWorker struct {
	repo        repository.Querier
	client      *http.Client
	signingKey  []byte        // global fallback secret (default endpoint / no per-merchant secret)
	secrets     secretOpener  // decrypts per-merchant secrets; nil => always use signingKey
	defaultURL  string        // fallback endpoint when the event has no target_url
	interval    time.Duration
	batchSize   int32
	maxAttempts int32
	baseBackoff time.Duration
	log         zerolog.Logger
}

// NewWebhookWorker wires the outbound webhook delivery worker. signingKey is the
// global fallback HMAC secret; secrets decrypts per-merchant signing secrets
// (may be nil in dev/scaffolding to always use signingKey); defaultURL is used
// when an event row carries no target_url.
func NewWebhookWorker(repo repository.Querier, secrets secretOpener, signingKey, defaultURL string, interval time.Duration, maxAttempts int32, log zerolog.Logger) *WebhookWorker {
	return &WebhookWorker{
		repo:        repo,
		client:      &http.Client{Timeout: 10 * time.Second},
		signingKey:  []byte(signingKey),
		secrets:     secrets,
		defaultURL:  defaultURL,
		interval:    interval,
		batchSize:   50,
		maxAttempts: maxAttempts,
		baseBackoff: 30 * time.Second,
		log:         log,
	}
}

func (w *WebhookWorker) Name() string            { return "webhook-delivery" }
func (w *WebhookWorker) Interval() time.Duration { return w.interval }

// Run delivers a batch of due webhook events.
func (w *WebhookWorker) Run(ctx context.Context) error {
	if w.repo == nil {
		return nil
	}
	events, err := w.repo.ListPendingWebhooks(ctx, w.batchSize)
	if err != nil {
		return err
	}
	for _, ev := range events {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		w.deliver(ctx, ev)
	}
	return nil
}

// deliver attempts one POST for a single event and records the outcome
// (delivered / retry-with-backoff / failed after cap).
func (w *WebhookWorker) deliver(ctx context.Context, ev repository.WebhookEvent) {
	id := ev.ID
	url := w.defaultURL
	if ev.TargetUrl != nil && *ev.TargetUrl != "" {
		url = *ev.TargetUrl
	}

	// No endpoint configured: fail fast rather than spin.
	if url == "" {
		msg := "no webhook target url configured"
		_ = w.repo.MarkWebhookFailed(ctx, repository.MarkWebhookFailedParams{ID: id, LastError: &msg})
		return
	}

	err := w.post(ctx, url, ev.Payload, w.signingKeyFor(ctx, ev.MerchantID))
	if err == nil {
		if derr := w.repo.MarkWebhookDelivered(ctx, id); derr != nil {
			w.log.Error().Err(derr).Msg("mark webhook delivered")
		}
		return
	}

	// attempts is the count BEFORE this attempt; +1 accounts for the try we just
	// made. When the cap is reached the event is parked as failed.
	nextAttemptCount := ev.Attempts + 1
	errStr := err.Error()
	if nextAttemptCount >= w.maxAttempts {
		if ferr := w.repo.MarkWebhookFailed(ctx, repository.MarkWebhookFailedParams{ID: id, LastError: &errStr}); ferr != nil {
			w.log.Error().Err(ferr).Msg("mark webhook failed")
		}
		w.log.Warn().Str("event_id", uuidStr(id)).Int32("attempts", nextAttemptCount).Msg("webhook delivery gave up")
		return
	}

	backoff := w.backoffFor(nextAttemptCount)
	next := pgtype.Timestamptz{Time: time.Now().UTC().Add(backoff), Valid: true}
	if rerr := w.repo.MarkWebhookRetry(ctx, repository.MarkWebhookRetryParams{
		ID:            id,
		NextAttemptAt: next,
		LastError:     &errStr,
	}); rerr != nil {
		w.log.Error().Err(rerr).Msg("mark webhook retry")
	}
}

// signingKeyFor resolves the HMAC key used to sign a merchant's delivery. It
// decrypts the merchant's own webhook secret (webhook_secret_enc) when present,
// so the merchant can verify with the whsec_ they were issued; otherwise it
// falls back to the global signing key (default endpoint / merchant with no
// secret set). A decrypt failure is logged (never the secret itself) and the
// global key is used so a single corrupt row does not wedge delivery.
func (w *WebhookWorker) signingKeyFor(ctx context.Context, merchantID pgtype.UUID) []byte {
	if w.secrets == nil || !merchantID.Valid {
		return w.signingKey
	}
	enc, err := w.repo.GetMerchantWebhookSigningKey(ctx, merchantID)
	if err != nil || len(enc) == 0 {
		return w.signingKey // no per-merchant secret (or lookup failed): global key
	}
	secret, err := w.secrets.Open(ctx, enc)
	if err != nil {
		w.log.Error().Str("merchant_id", uuidStr(merchantID)).Msg("decrypt merchant webhook secret failed; using global key")
		return w.signingKey
	}
	return secret
}

// post signs and delivers the payload with the given HMAC key. A 2xx is success;
// anything else is an error that triggers retry/backoff.
func (w *WebhookWorker) post(ctx context.Context, url string, payload, key []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range signedHeaders(key, payload, time.Now().Unix()) {
		req.Header.Set(k, v)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
	}
	return nil
}

// signedHeaders returns the HMAC-SHA256 signature headers merchants use to verify
// a delivery is authentic and not replayed:
//
//	X-Signature         : sha256=<hex HMAC(secret, body)>                 (legacy, body only)
//	X-PayCore-Timestamp : <unix seconds>
//	X-PayCore-Signature : t=<unix>,v1=<hex HMAC(secret, "<unix>.<body>")> (replay-resistant)
//
// The v1 signature binds the timestamp into the MAC (the Stripe scheme): a
// receiver recomputes HMAC over "<t>.<rawBody>", compares in constant time, and
// rejects deliveries whose t is outside a tolerance window — so a captured
// payload cannot be replayed later without the secret. The legacy X-Signature is
// kept so existing receivers keep working.
func signedHeaders(key, payload []byte, ts int64) map[string]string {
	tsStr := strconv.FormatInt(ts, 10)
	signed := make([]byte, 0, len(tsStr)+1+len(payload))
	signed = append(signed, tsStr...)
	signed = append(signed, '.')
	signed = append(signed, payload...)
	return map[string]string{
		"X-Signature":         "sha256=" + hmacHex(key, payload),
		"X-PayCore-Timestamp": tsStr,
		"X-PayCore-Signature": "t=" + tsStr + ",v1=" + hmacHex(key, signed),
	}
}

// hmacHex is the hex-encoded HMAC-SHA256 of b under key.
func hmacHex(key, b []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(b)
	return hex.EncodeToString(mac.Sum(nil))
}

// backoffFor returns an exponential backoff (base * 2^(attempt-1)) capped at 1h.
func (w *WebhookWorker) backoffFor(attempt int32) time.Duration {
	d := w.baseBackoff
	for i := int32(1); i < attempt; i++ {
		d *= 2
		if d >= time.Hour {
			return time.Hour
		}
	}
	if d > time.Hour {
		return time.Hour
	}
	return d
}

func uuidStr(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
