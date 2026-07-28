package worker

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func refHMAC(secret, msg string) string {
	m := hmac.New(sha256.New, []byte(secret))
	m.Write([]byte(msg))
	return hex.EncodeToString(m.Sum(nil))
}

// TestSignedHeaders verifies the standards-grade webhook signature scheme:
// a legacy body-only HMAC for backward compatibility, plus a timestamped,
// replay-resistant v1 signature that binds the timestamp into the MAC.
func TestSignedHeaders(t *testing.T) {
	secret := "whsec_test_secret"
	w := &WebhookWorker{signingKey: []byte(secret)}
	body := []byte(`{"event":"payment.captured","id":"pay_1","amount":12900}`)
	var ts int64 = 1700000000

	h := w.signedHeaders(body, ts)

	// Legacy body-only signature (kept for backward compatibility).
	wantLegacy := "sha256=" + refHMAC(secret, string(body))
	if h["X-Signature"] != wantLegacy {
		t.Fatalf("X-Signature = %q want %q", h["X-Signature"], wantLegacy)
	}

	// Timestamp header.
	if h["X-PayCore-Timestamp"] != "1700000000" {
		t.Fatalf("X-PayCore-Timestamp = %q want 1700000000", h["X-PayCore-Timestamp"])
	}

	// v1 binds the timestamp: HMAC over "<ts>.<body>".
	wantV1 := refHMAC(secret, "1700000000."+string(body))
	wantSig := "t=1700000000,v1=" + wantV1
	if h["X-PayCore-Signature"] != wantSig {
		t.Fatalf("X-PayCore-Signature = %q want %q", h["X-PayCore-Signature"], wantSig)
	}
}

// TestSignedHeadersTimestampBinding proves replay resistance: the same body at a
// different timestamp produces a different v1 signature, so a captured delivery
// cannot be replayed with a fresh timestamp without the secret.
func TestSignedHeadersTimestampBinding(t *testing.T) {
	w := &WebhookWorker{signingKey: []byte("whsec_x")}
	body := []byte(`{"id":"evt_1"}`)

	a := w.signedHeaders(body, 1700000000)["X-PayCore-Signature"]
	b := w.signedHeaders(body, 1700000001)["X-PayCore-Signature"]
	if a == b {
		t.Fatal("v1 signature must change when the timestamp changes (replay binding)")
	}
	// The legacy body-only signature, by contrast, is timestamp-independent.
	la := w.signedHeaders(body, 1700000000)["X-Signature"]
	lb := w.signedHeaders(body, 1700000001)["X-Signature"]
	if la != lb {
		t.Fatal("legacy body-only signature must not depend on the timestamp")
	}
}
