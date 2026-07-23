// Package migrations embeds the SQL migration files so they can be applied
// in-process at boot. The distroless runtime image ships no shell or migrate
// binary, so the server runs migrations itself (see internal/db.RunMigrations)
// using this embedded filesystem as the golang-migrate iofs source.
package migrations

import "embed"

// FS holds every up/down migration. golang-migrate's iofs source driver reads
// the *.up.sql / *.down.sql files directly from it, so no files ship alongside
// the binary.
//
//go:embed *.sql
var FS embed.FS
