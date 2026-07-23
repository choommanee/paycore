package external

import (
	"context"
	"strings"
	"testing"

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
