package migrations_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// migrationsDir is the directory containing this test (the migrations folder).
const migrationsDir = "."

// TestEveryMigrationHasUpAndDown asserts every *.up.sql has a matching *.down.sql
// (and vice versa) so a migration can always be rolled back. golang-migrate
// silently no-ops a missing direction, which is how irreversible migrations sneak
// in; this test fails fast instead.
func TestEveryMigrationHasUpAndDown(t *testing.T) {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}
	ups := map[string]bool{}
	downs := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		switch {
		case strings.HasSuffix(name, ".up.sql"):
			ups[strings.TrimSuffix(name, ".up.sql")] = true
		case strings.HasSuffix(name, ".down.sql"):
			downs[strings.TrimSuffix(name, ".down.sql")] = true
		}
	}
	if len(ups) == 0 {
		t.Fatal("no migrations found")
	}
	for base := range ups {
		if !downs[base] {
			t.Errorf("migration %q has an up but no down", base)
		}
	}
	for base := range downs {
		if !ups[base] {
			t.Errorf("migration %q has a down but no up", base)
		}
	}
}

// indexNameRe extracts index names created/dropped in a migration so up/down can
// be cross-checked.
var (
	createIdxRe = regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	dropIdxRe   = regexp.MustCompile(`(?i)DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?(\w+)`)
)

// TestWebhookOutboxIndexUpDownConsistent asserts the new outbox index migration
// creates and drops the same index name (a common copy-paste bug that leaves a
// down migration a no-op).
func TestWebhookOutboxIndexUpDownConsistent(t *testing.T) {
	up := mustRead(t, "000005_webhook_outbox_index.up.sql")
	down := mustRead(t, "000005_webhook_outbox_index.down.sql")

	upIdx := createIdxRe.FindStringSubmatch(up)
	if upIdx == nil {
		t.Fatalf("up migration does not CREATE INDEX: %s", up)
	}
	downIdx := dropIdxRe.FindStringSubmatch(down)
	if downIdx == nil {
		t.Fatalf("down migration does not DROP INDEX: %s", down)
	}
	if upIdx[1] != downIdx[1] {
		t.Fatalf("index name mismatch: up creates %q, down drops %q", upIdx[1], downIdx[1])
	}
	// The index must be the partial outbox index over still-pending rows.
	if !strings.Contains(strings.ToUpper(up), "WHERE DELIVERED = FALSE") {
		t.Fatalf("outbox index is not partial on pending rows: %s", up)
	}
}

func mustRead(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(migrationsDir, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(b)
}
