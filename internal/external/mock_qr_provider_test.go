package external

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestQRGenerate(t *testing.T) {
	p := NewMockQRProvider("secret", "")
	res, err := p.Generate(context.Background(), QRGenRequest{
		Method:    "promptpay",
		Amount:    decimal.RequireFromString("42.00"),
		Currency:  "THB",
		Reference: "ORDER-9",
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.ProviderRef == "" {
		t.Fatal("missing provider ref")
	}
	if !strings.Contains(res.Payload, "ORDER-9") {
		t.Fatalf("payload missing reference: %s", res.Payload)
	}
	if !strings.HasPrefix(res.ImageURL, "https://sandbox.qr.example.com/qr/") {
		t.Fatalf("unexpected image url: %s", res.ImageURL)
	}
}

func TestVerifyWebhookValidSignature(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"provider_ref":"abc","status":"paid"}`)

	sig := ComputeQRWebhookSignature(secret, body)
	if err := p.VerifyWebhook(context.Background(), sig, body); err != nil {
		t.Fatalf("valid signature rejected: %v", err)
	}
	// With the common sha256= prefix.
	if err := p.VerifyWebhook(context.Background(), "sha256="+sig, body); err != nil {
		t.Fatalf("prefixed signature rejected: %v", err)
	}
}

func TestVerifyWebhookRejects(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"status":"paid"}`)
	good := ComputeQRWebhookSignature(secret, body)

	// Tampered body.
	if err := p.VerifyWebhook(context.Background(), good, []byte(`{"status":"failed"}`)); err == nil {
		t.Fatal("expected mismatch for tampered body")
	}
	// Wrong secret signature.
	wrong := ComputeQRWebhookSignature("other-secret", body)
	if err := p.VerifyWebhook(context.Background(), wrong, body); err == nil {
		t.Fatal("expected mismatch for wrong-secret signature")
	}
	// Non-hex signature.
	if err := p.VerifyWebhook(context.Background(), "zzzz", body); err == nil {
		t.Fatal("expected error for non-hex signature")
	}
}

func TestVerifyWebhookNoSecret(t *testing.T) {
	p := NewMockQRProvider("", "")
	if err := p.VerifyWebhook(context.Background(), "deadbeef", []byte("x")); err == nil {
		t.Fatal("expected error when no shared secret configured")
	}
}

// ---- Fix 1: timestamped (v1) inbound signature + replay protection ---------

// TestVerifyWebhookTimestampedValid: a correctly-timestamped v1 signature
// (t=<now>,v1=<hex HMAC(secret,"<t>.<body>")>) verifies.
func TestVerifyWebhookTimestampedValid(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"provider_ref":"abc","status":"paid"}`)

	sig := ComputeQRWebhookSignatureV1(secret, time.Now().Unix(), body)
	if err := p.VerifyWebhook(context.Background(), sig, body); err != nil {
		t.Fatalf("valid timestamped signature rejected: %v", err)
	}
}

// TestVerifyWebhookTimestampedReplayRejected: a v1 signature that is perfectly
// valid but whose timestamp is older than the tolerance window is rejected —
// this is the replay-protection guarantee.
func TestVerifyWebhookTimestampedReplayRejected(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"provider_ref":"abc","status":"paid"}`)

	old := time.Now().Add(-10 * time.Minute).Unix() // > 5m tolerance
	sig := ComputeQRWebhookSignatureV1(secret, old, body)
	if err := p.VerifyWebhook(context.Background(), sig, body); err == nil {
		t.Fatal("expected an old-timestamp v1 signature to be rejected (replay)")
	}

	// A far-future timestamp (clock skew / forgery) is equally rejected.
	future := time.Now().Add(10 * time.Minute).Unix()
	if err := p.VerifyWebhook(context.Background(), ComputeQRWebhookSignatureV1(secret, future, body), body); err == nil {
		t.Fatal("expected a far-future-timestamp v1 signature to be rejected")
	}
}

// TestVerifyWebhookTimestampedTamperRejected: tampering with either the body or
// the bound timestamp breaks the MAC and is rejected.
func TestVerifyWebhookTimestampedTamperRejected(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"status":"paid"}`)
	now := time.Now().Unix()
	sig := ComputeQRWebhookSignatureV1(secret, now, body)

	// Tampered body: same signature over a different body must fail.
	if err := p.VerifyWebhook(context.Background(), sig, []byte(`{"status":"failed"}`)); err == nil {
		t.Fatal("expected mismatch for tampered body under v1")
	}

	// Tampered timestamp: swap t to a fresh (in-window) value while keeping the
	// original v1 hex. The MAC binds t, so it must fail (not silently pass the
	// window check). sig is "t=<now>,v1=<hex>"; keep the v1 part, change t.
	v1Part := strings.SplitN(sig, ",", 2)[1]
	forged := "t=" + strconv.FormatInt(now+1, 10) + "," + v1Part
	if err := p.VerifyWebhook(context.Background(), forged, body); err == nil {
		t.Fatal("expected mismatch when the bound timestamp is altered")
	}
}

// TestVerifyWebhookLegacyStillVerifies: the legacy body-only form (bare hex and
// sha256=<hex>) is still accepted for backward compatibility.
func TestVerifyWebhookLegacyStillVerifies(t *testing.T) {
	const secret = "shared-secret"
	p := NewMockQRProvider(secret, "")
	body := []byte(`{"provider_ref":"abc","status":"paid"}`)

	legacy := ComputeQRWebhookSignature(secret, body)
	if err := p.VerifyWebhook(context.Background(), legacy, body); err != nil {
		t.Fatalf("legacy body-only signature rejected: %v", err)
	}
	if err := p.VerifyWebhook(context.Background(), "sha256="+legacy, body); err != nil {
		t.Fatalf("legacy sha256= signature rejected: %v", err)
	}
}
