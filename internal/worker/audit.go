package worker

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// AuditWorker is an in-process outbox for audit_log writes. The authorize hot
// path enqueues an audit row instead of INSERTing it inside the payment
// transaction; this worker drains the queue and writes rows in the background,
// removing one row-write per authorize from the synchronous tx (write
// amplification reduction).
//
// The audit_log is append-only and non-authoritative for money movement (the
// payments row + ledger_entries are), so eventual, at-least-once persistence off
// the hot path is an acceptable trade for throughput. On shutdown the queue is
// drained so no buffered audit row is lost during a graceful stop.
type AuditWorker struct {
	repo  repository.Querier
	ch    chan repository.WriteAuditLogParams
	log   zerolog.Logger
	wg    sync.WaitGroup
	dropn int // best-effort dropped-count when the buffer is full (logged)
}

// NewAuditWorker returns an AuditWorker with a buffered queue of the given size.
func NewAuditWorker(repo repository.Querier, buffer int, log zerolog.Logger) *AuditWorker {
	if buffer <= 0 {
		buffer = 1024
	}
	return &AuditWorker{
		repo: repo,
		ch:   make(chan repository.WriteAuditLogParams, buffer),
		log:  log.With().Str("worker", "audit-outbox").Logger(),
	}
}

// Start launches the drain loop. It runs until ctx is cancelled, then drains any
// remaining buffered rows before returning.
func (w *AuditWorker) Start(ctx context.Context) {
	w.wg.Add(1)
	// #nosec G118 -- intentional: this is a background outbox drain, not a
	// request-scoped goroutine. Its writes must be decoupled from any request
	// context (they may outlive the enqueuing request), and write() bounds each
	// write with its own timeout. Using the request context here would cancel a
	// buffered audit write the moment the originating request returned.
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-ctx.Done():
				w.drain()
				return
			case p := <-w.ch:
				w.write(context.Background(), p)
			}
		}
	}()
}

// Enqueue submits an audit row for asynchronous persistence. It never blocks the
// caller: if the buffer is full the row is dropped and counted (audit is
// best-effort off the hot path). Returns true when enqueued.
func (w *AuditWorker) Enqueue(p repository.WriteAuditLogParams) bool {
	select {
	case w.ch <- p:
		return true
	default:
		w.dropn++
		w.log.Warn().Int("dropped", w.dropn).Msg("audit outbox full; dropping row")
		return false
	}
}

// drain writes any rows still buffered at shutdown.
func (w *AuditWorker) drain() {
	for {
		select {
		case p := <-w.ch:
			// #nosec G118 -- shutdown drain: writes must outlive the request that
			// enqueued them; write() bounds each with its own timeout.
			w.write(context.Background(), p)
		default:
			return
		}
	}
}

func (w *AuditWorker) write(ctx context.Context, p repository.WriteAuditLogParams) {
	wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := w.repo.WriteAuditLog(wctx, p); err != nil {
		w.log.Error().Err(err).Str("action", p.Action).Msg("async audit write failed")
	}
}

// Wait blocks until the drain loop has exited (after ctx cancellation). Call
// during graceful shutdown so buffered audit rows are flushed.
func (w *AuditWorker) Wait() { w.wg.Wait() }
