package db

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/rs/zerolog"

	// Register migrate's pgx/v5 database driver under the "pgx5://" scheme; it
	// self-registers in init(). This is separate from the pgxpool the app serves
	// traffic with — migrations run once at boot on their own connection.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
)

// RunMigrations applies all pending up migrations from fsys against dsn, in
// process, using the embedded iofs source and the pgx/v5 database driver. It is
// idempotent: migrate.ErrNoChange (nothing to apply) is treated as success. Any
// other error is returned so the caller can fail fast — a half-migrated schema
// must never start serving. The applied schema version is logged.
//
// dsn is the standard Postgres URL (DATABASE_URL). The migrations table
// (schema_migrations) is created automatically on first run.
func RunMigrations(ctx context.Context, dsn string, fsys fs.FS, log zerolog.Logger) error {
	if dsn == "" {
		return fmt.Errorf("db: DATABASE_URL is empty; cannot run migrations")
	}

	src, err := iofs.New(fsys, ".")
	if err != nil {
		return fmt.Errorf("db: open embedded migrations: %w", err)
	}
	defer func() { _ = src.Close() }()

	// pgx/v5 database driver reads a plain sql.DB opened over the DSN. The
	// "pgx5://" scheme is what the driver registers; rewrite the URL scheme so
	// migrate.NewWithSourceInstance's database opener selects it.
	dbURL := "pgx5://" + trimScheme(dsn)

	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		return fmt.Errorf("db: init migrator: %w", err)
	}
	// Best-effort close of the underlying source/database handles.
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("db: run migrations: %w", err)
	}

	version, dirty, verr := m.Version()
	if verr != nil && !errors.Is(verr, migrate.ErrNilVersion) {
		return fmt.Errorf("db: read schema version: %w", verr)
	}
	log.Info().Uint("schema_version", version).Bool("dirty", dirty).Msg("migrations applied")
	return nil
}



// trimScheme strips a leading postgres:// or postgresql:// scheme from a DSN so
// it can be re-prefixed with the migrate driver's own scheme. A DSN without a
// recognised scheme is returned unchanged.
func trimScheme(dsn string) string {
	for _, p := range []string{"postgres://", "postgresql://"} {
		if len(dsn) >= len(p) && dsn[:len(p)] == p {
			return dsn[len(p):]
		}
	}
	return dsn
}
