package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/repository"
)

// captureRepo records WriteAuditLog calls; all other Querier methods are
// inherited from the embedded nil interface and panic if unexpectedly used.
type captureRepo struct {
	repository.Querier
	mu   sync.Mutex
	rows []repository.WriteAuditLogParams
}

func (r *captureRepo) WriteAuditLog(_ context.Context, p repository.WriteAuditLogParams) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.rows = append(r.rows, p)
	return nil
}

func (r *captureRepo) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.rows)
}

// TestAuditWorkerDrainsOnShutdown verifies enqueued rows are persisted and that
// buffered rows are flushed when the context is cancelled.
func TestAuditWorkerDrainsOnShutdown(t *testing.T) {
	repo := &captureRepo{}
	w := NewAuditWorker(repo, 16, zerolog.Nop())
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	for i := 0; i < 5; i++ {
		if !w.Enqueue(repository.WriteAuditLogParams{
			ID:     pgtype.UUID{Valid: true},
			Action: "payment.authorized",
		}) {
			t.Fatalf("enqueue %d dropped unexpectedly", i)
		}
	}

	// Give the drain loop a moment, then stop and wait for the final drain.
	deadline := time.Now().Add(2 * time.Second)
	for repo.count() < 5 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	w.Wait()

	if got := repo.count(); got != 5 {
		t.Fatalf("persisted %d audit rows, want 5", got)
	}
}
