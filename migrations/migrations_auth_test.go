package migrations_test

import (
	"strings"
	"testing"
)

// TestMerchantUsersMigrationReversible asserts the merchant_users up migration
// creates the table + unique oauth index and the down migration drops the table.
func TestMerchantUsersMigrationReversible(t *testing.T) {
	up := mustRead(t, "000008_merchant_users.up.sql")
	down := mustRead(t, "000008_merchant_users.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE MERCHANT_USERS") {
		t.Fatalf("up does not create merchant_users: %s", up)
	}
	if !strings.Contains(upU, "UNIQUE") || !strings.Contains(strings.ToLower(up), "oauth_subject") {
		t.Fatalf("up must enforce a unique (oauth_provider, oauth_subject): %s", up)
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "merchant_users") {
		t.Fatalf("down must drop merchant_users: %s", down)
	}
}
