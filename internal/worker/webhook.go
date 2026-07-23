package worker

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// WebhookWorker delivers pending webhook_events to merchant endpoints. Each POST
// is signed with an HMAC-SHA256 of the raw JSON body in the X-Signature header so
// merchants can verify authenticity. Delivery is retried with exponential
// backoff up to maxAttempts, after which the event is marked failed.
type WebhookWorker struct {
	repo        repository.Querier
	client      *http.Client
	signingKey  []byte
	defaultURL  string        // fallback endpoint when the event has no target_url
	interval    time.Duration
	batchSize   int32
	maxAttempts int32
	baseBackoff time.Duration
	log         zerolog.Logger
}

// NewWebhookWorker wires the outbound webhook delivery worker. signingKey is the
// shared HMAC secret; defaultURL is used when an event row carries no target_url.
func NewWebhookWorker(repo repository.Querier, signingKey, defaultURL string, interval time.Duration, maxAttempts int32, log zerolog.Logger) *WebhookWorker {
	return &WebhookWorker{
		repo:        repo,
		client:      &http.Client{Timeout: 10 * time.Second},
		signingKey:  []byte(signingKey),
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

	err := w.post(ctx, url, ev.Payload)
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

// post signs and delivers the payload. A 2xx is success; anything else is an
// error that triggers retry/backoff.
func (w *WebhookWorker) post(ctx context.Context, url string, payload []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", w.sign(payload))

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

// sign returns the hex-encoded HMAC-SHA256 of the payload using the signing key.
func (w *WebhookWorker) sign(payload []byte) string {
	mac := hmac.New(sha256.New, w.signingKey)
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
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
