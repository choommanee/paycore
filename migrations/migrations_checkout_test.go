package migrations_test

import (
	"strings"
	"testing"
)

func TestCheckoutSessionsMigrationReversible(t *testing.T) {
	up := mustRead(t, "000010_checkout_sessions.up.sql")
	down := mustRead(t, "000010_checkout_sessions.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE CHECKOUT_SESSIONS") {
		t.Fatalf("up does not create checkout_sessions: %s", up)
	}
	// The session token is stored hashed and must be uniquely addressable.
	if !strings.Contains(strings.ToLower(up), "session_token_hash") || !strings.Contains(upU, "UNIQUE") {
		t.Fatalf("up must have a unique session_token_hash: %s", up)
	}
	if !strings.Contains(upU, "AMOUNT_MINOR") || !strings.Contains(upU, "CHECK") {
		t.Fatalf("up must CHECK amount_minor > 0: %s", up)
	}
	// Nullable bridges to payments / qr_payments / payment_links.
	for _, fk := range []string{"payment_link_id", "payment_id", "qr_payment_id"} {
		if !strings.Contains(strings.ToLower(up), fk) {
			t.Fatalf("up must have column %s: %s", fk, up)
		}
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "checkout_sessions") {
		t.Fatalf("down must drop checkout_sessions: %s", down)
	}
}
