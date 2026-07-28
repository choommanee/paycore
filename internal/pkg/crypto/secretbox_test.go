package crypto

import (
	"bytes"
	"context"
	"testing"
)

func newTestSecretBox(t *testing.T) *SecretBox {
	t.Helper()
	kms, err := NewDevLocalKMS("dev-key")
	if err != nil {
		t.Fatalf("kms: %v", err)
	}
	return NewSecretBox(kms)
}

// TestSecretBoxRoundTrip: a sealed secret opens back to the exact plaintext, and
// the sealed blob does not contain the plaintext in the clear.
func TestSecretBoxRoundTrip(t *testing.T) {
	sb := newTestSecretBox(t)
	ctx := context.Background()

	secret := []byte("whsec_deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")
	blob, err := sb.Seal(ctx, secret)
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if bytes.Contains(blob, secret) {
		t.Fatal("sealed blob leaks the plaintext secret")
	}
	got, err := sb.Open(ctx, blob)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if !bytes.Equal(got, secret) {
		t.Fatalf("round-trip mismatch: got %q want %q", got, secret)
	}
}

// TestSecretBoxSealIsNondeterministic: two seals of the same plaintext differ
// (fresh DEK + nonce), so ciphertext equality can't be used as an oracle.
func TestSecretBoxSealIsNondeterministic(t *testing.T) {
	sb := newTestSecretBox(t)
	ctx := context.Background()
	a, _ := sb.Seal(ctx, []byte("same"))
	b, _ := sb.Seal(ctx, []byte("same"))
	if bytes.Equal(a, b) {
		t.Fatal("two seals of the same plaintext must differ")
	}
}

// TestSecretBoxOpenRejectsTampered: a flipped ciphertext byte fails GCM auth.
func TestSecretBoxOpenRejectsTampered(t *testing.T) {
	sb := newTestSecretBox(t)
	ctx := context.Background()
	blob, err := sb.Seal(ctx, []byte("secret-value"))
	if err != nil {
		t.Fatalf("Seal: %v", err)
	}
	blob[len(blob)-1] ^= 0xFF
	if _, err := sb.Open(ctx, blob); err == nil {
		t.Fatal("expected Open to reject a tampered blob")
	}
}

// TestSecretBoxOpenRejectsGarbage: undersized / wrong-version blobs are rejected.
func TestSecretBoxOpenRejectsGarbage(t *testing.T) {
	sb := newTestSecretBox(t)
	ctx := context.Background()
	for _, b := range [][]byte{nil, {0x00}, {0x01}, {0x01, 0x00, 0x10}} {
		if _, err := sb.Open(ctx, b); err == nil {
			t.Fatalf("expected error for garbage blob %v", b)
		}
	}
}
