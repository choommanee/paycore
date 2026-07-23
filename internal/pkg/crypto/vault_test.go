package crypto

import (
	"context"
	"strings"
	"testing"
)

func newTestVault(t *testing.T) *Vault {
	t.Helper()
	kms, err := NewDevLocalKMS("alias/test")
	if err != nil {
		t.Fatalf("NewDevLocalKMS: %v", err)
	}
	return NewVault(kms)
}

func TestTokenizeDetokenizeRoundTrip(t *testing.T) {
	v := newTestVault(t)
	ctx := context.Background()

	pans := []string{
		"4111111111111111", // visa 16
		"5555555555554444", // mastercard 16
		"378282246310005",  // amex 15
		"4000000000000119", // 3ds test pan
	}
	for _, pan := range pans {
		token, last4, err := v.Tokenize(ctx, pan)
		if err != nil {
			t.Fatalf("Tokenize(%s): %v", pan, err)
		}
		if last4 != pan[len(pan)-4:] {
			t.Fatalf("last4=%s want %s", last4, pan[len(pan)-4:])
		}
		if !strings.HasPrefix(token, tokenPrefix+".") {
			t.Fatalf("token missing version prefix: %s", token)
		}
		if strings.Contains(token, pan) {
			t.Fatalf("token leaks PAN plaintext")
		}
		got, err := v.Detokenize(ctx, token)
		if err != nil {
			t.Fatalf("Detokenize: %v", err)
		}
		if got != pan {
			t.Fatalf("round-trip mismatch: got %s want %s", got, pan)
		}
	}
}

// TestTokenizeUnique verifies fresh per-record DEK / nonce make each token
// distinct even for the same PAN, while both still detokenize correctly.
func TestTokenizeUnique(t *testing.T) {
	v := newTestVault(t)
	ctx := context.Background()
	const pan = "4111111111111111"

	t1, _, err := v.Tokenize(ctx, pan)
	if err != nil {
		t.Fatal(err)
	}
	t2, _, err := v.Tokenize(ctx, pan)
	if err != nil {
		t.Fatal(err)
	}
	if t1 == t2 {
		t.Fatal("expected distinct tokens for repeated tokenization")
	}
	for _, tk := range []string{t1, t2} {
		got, err := v.Detokenize(ctx, tk)
		if err != nil || got != pan {
			t.Fatalf("detokenize failed: %s err=%v", got, err)
		}
	}
}

func TestTokenizeInvalidPANLength(t *testing.T) {
	v := newTestVault(t)
	ctx := context.Background()
	for _, bad := range []string{"", "123", "12345678901", strings.Repeat("9", 20)} {
		if _, _, err := v.Tokenize(ctx, bad); err == nil {
			t.Fatalf("expected error for PAN length %d", len(bad))
		}
	}
}

func TestDetokenizeBadToken(t *testing.T) {
	v := newTestVault(t)
	ctx := context.Background()
	cases := []string{
		"",
		"not-a-token",
		"tokv1.!!!notbase64",
		"tokv2." + "AAAA",       // wrong version
		tokenPrefix + ".AA",     // truncated buffer
	}
	for _, tk := range cases {
		if _, err := v.Detokenize(ctx, tk); err == nil {
			t.Fatalf("expected error detokenizing %q", tk)
		}
	}
}

// TestDetokenizeWrongKMS ensures a token minted under one master key cannot be
// detokenized under a different key (AEAD auth failure).
func TestDetokenizeWrongKMS(t *testing.T) {
	ctx := context.Background()
	k1, _ := NewLocalKMS("k1", []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"))
	k2, _ := NewLocalKMS("k2", []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"))
	v1 := NewVault(k1)
	v2 := NewVault(k2)

	token, _, err := v1.Tokenize(ctx, "4111111111111111")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := v2.Detokenize(ctx, token); err == nil {
		t.Fatal("expected detokenize failure under wrong KMS key")
	}
}

func TestLocalKMSWrapUnwrap(t *testing.T) {
	ctx := context.Background()
	k, err := NewLocalKMS("k", []byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if k.KeyID() != "k" {
		t.Fatalf("KeyID=%s", k.KeyID())
	}
	dek := []byte("this-is-a-32-byte-data-enc-key!!")
	wrapped, err := k.WrapDEK(ctx, dek)
	if err != nil {
		t.Fatal(err)
	}
	if string(wrapped) == string(dek) {
		t.Fatal("wrapped DEK equals plaintext")
	}
	un, err := k.UnwrapDEK(ctx, wrapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(un) != string(dek) {
		t.Fatal("unwrap mismatch")
	}
}

func TestNewLocalKMSBadKeyLen(t *testing.T) {
	if _, err := NewLocalKMS("k", []byte("short")); err == nil {
		t.Fatal("expected error for non-32-byte key")
	}
}
