// Package worker holds the background ops workers that run alongside the HTTP
// server: settlement (aggregating captured payments into payouts), reconciliation
// (comparing ledger totals against the acquirer settlement report) and outbound
// webhook delivery (signed, retried POSTs of pending webhook_events).
//
// Every worker is a Worker: a named goroutine loop with a fixed tick interval
// that stops cleanly when its context is cancelled. main.go starts them and
// cancels the shared context on shutdown.
package worker

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

// Worker is a long-running background loop.
type Worker interface {
	// Name identifies the worker in logs.
	Name() string
	// Interval is the tick period between runs.
	Interval() time.Duration
	// Run executes one unit of work. It should be idempotent and bounded; the
	// runner calls it on every tick and on shutdown-triggered final drain.
	Run(ctx context.Context) error
}

// Start launches w in its own goroutine, running immediately and then on every
// Interval tick until ctx is cancelled. Errors are logged (workers are
// best-effort and must not crash the process). The returned function is not
// needed for shutdown — cancelling ctx stops the loop — but Start blocks nothing.
func Start(ctx context.Context, w Worker, log zerolog.Logger) {
	l := log.With().Str("worker", w.Name()).Logger()
	go func() {
		l.Info().Dur("interval", w.Interval()).Msg("worker started")
		ticker := time.NewTicker(w.Interval())
		defer ticker.Stop()

		runOnce := func() {
			// Bound each run so a stuck DB call cannot wedge the loop.
			runCtx, cancel := context.WithTimeout(ctx, w.Interval()+30*time.Second)
			defer cancel()
			if err := w.Run(runCtx); err != nil {
				l.Error().Err(err).Msg("worker run failed")
			}
		}

		runOnce() // run immediately on startup
		for {
			select {
			case <-ctx.Done():
				l.Info().Msg("worker stopped")
				return
			case <-ticker.C:
				runOnce()
			}
		}
	}()
}

// StartAll launches every worker in ws.
func StartAll(ctx context.Context, log zerolog.Logger, ws ...Worker) {
	for _, w := range ws {
		Start(ctx, w, log)
	}
}
