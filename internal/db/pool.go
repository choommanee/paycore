package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig tunes the pgx connection pool. Defaults are sane for a service that
// fans out short, transactional queries (auth/capture/ledger writes).
type PoolConfig struct {
	// DatabaseURL is the pgx connection string (cfg.DATABASE_URL).
	DatabaseURL string
	// MaxConns caps the pool size. Keep this well under the server's
	// max_connections, leaving headroom for migrations/admin sessions.
	MaxConns int32
	// MinConns keeps a few warm connections to avoid cold-start latency.
	MinConns int32
	// MaxConnLifetime recycles connections to survive failovers / rolling PG upgrades.
	MaxConnLifetime time.Duration
	// MaxConnIdleTime closes idle connections back down to MinConns.
	MaxConnIdleTime time.Duration
}

// DefaultPoolConfig returns pool settings for the given DSN. MaxConns/MinConns
// are sized for a service that fans out ~5-6 roundtrips plus a tx per authorize;
// 50 conns keeps headroom under a typical PG max_connections while removing the
// pool as a throughput ceiling, and 10 warm conns avoid cold-start latency
// spikes on the first burst. Both are overridable via WithSizing (config-driven).
func DefaultPoolConfig(databaseURL string) PoolConfig {
	return PoolConfig{
		DatabaseURL:     databaseURL,
		MaxConns:        50,
		MinConns:        10,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	}
}

// WithSizing overrides MaxConns/MinConns when the values are positive, letting
// the pool be tuned per environment from config. Non-positive values leave the
// defaults in place.
func (pc PoolConfig) WithSizing(maxConns, minConns int32) PoolConfig {
	if maxConns > 0 {
		pc.MaxConns = maxConns
	}
	if minConns > 0 {
		pc.MinConns = minConns
	}
	return pc
}

// NewPool opens a *pgxpool.Pool with the given settings and verifies
// connectivity with a Ping. The caller owns the pool and must Close it on
// shutdown. The DSN may contain credentials, so it is never logged here.
func NewPool(ctx context.Context, pc PoolConfig) (*pgxpool.Pool, error) {
	if pc.DatabaseURL == "" {
		return nil, fmt.Errorf("db: DATABASE_URL is empty")
	}

	cfg, err := pgxpool.ParseConfig(pc.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("db: parse config: %w", err)
	}

	if pc.MaxConns > 0 {
		cfg.MaxConns = pc.MaxConns
	}
	if pc.MinConns > 0 {
		cfg.MinConns = pc.MinConns
	}
	if pc.MaxConnLifetime > 0 {
		cfg.MaxConnLifetime = pc.MaxConnLifetime
	}
	if pc.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = pc.MaxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db: open pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping: %w", err)
	}

	return pool, nil
}
