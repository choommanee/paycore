package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentLinksMigrationReversible(t *testing.T) {
	up := mustRead(t, "000009_payment_links.up.sql")
	down := mustRead(t, "000009_payment_links.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE PAYMENT_LINKS") {
		t.Fatalf("up does not create payment_links: %s", up)
	}
	if !strings.Contains(strings.ToLower(up), "public_id") || !strings.Contains(upU, "UNIQUE") {
		t.Fatalf("up must have a unique public_id index: %s", up)
	}
	if !strings.Contains(upU, "AMOUNT_MINOR") || !strings.Contains(upU, "CHECK") {
		t.Fatalf("up must CHECK amount_minor > 0: %s", up)
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "payment_links") {
		t.Fatalf("down must drop payment_links: %s", down)
	}
}
