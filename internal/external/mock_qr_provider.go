package external

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// MockQRProvider is a deterministic implementation of QRProvider for sandbox and
// test use. Generate mints a fake EMVCo-style payload and a pre-rendered image
// URL; VerifyWebhook checks an HMAC-SHA256 signature over the raw webhook body
// using a shared secret (config: QR_WEBHOOK_SECRET).
//
// In production the payload would come from the bank/ITMX rail and the webhook
// secret would be rotated via a secrets manager.
type MockQRProvider struct {
	// WebhookSecret is the shared secret used to sign & verify inbound webhooks.
	WebhookSecret string
	// BaseURL is the origin used to build the simulated QR image URL.
	BaseURL string
}

// NewMockQRProvider returns a MockQRProvider. An empty baseURL falls back to a
// local sandbox default.
func NewMockQRProvider(webhookSecret, baseURL string) *MockQRProvider {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://sandbox.qr.example.com"
	}
	return &MockQRProvider{WebhookSecret: webhookSecret, BaseURL: baseURL}
}

// compile-time assertion.
var _ QRProvider = (*MockQRProvider)(nil)

// Generate registers the expected payment and returns a deterministic payload,
// image URL and provider reference. The reference echoes back the merchant
// reference for reconciliation while the ProviderRef is a fresh uuid.
func (p *MockQRProvider) Generate(_ context.Context, req QRGenRequest) (*QRGenResult, error) {
	providerRef := uuid.NewString()
	// A stand-in payload. Locally-built PromptPay QR would be produced by the
	// promptpay package; here we synthesize a stable string for the sandbox.
	payload := fmt.Sprintf(
		"MOCKQR|method=%s|amount=%s|currency=%s|ref=%s|pref=%s",
		req.Method, req.Amount.String(), req.Currency, req.Reference, providerRef,
	)
	imageURL := fmt.Sprintf("%s/qr/%s.png", strings.TrimRight(p.BaseURL, "/"), providerRef)

	return &QRGenResult{
		Payload:     payload,
		ImageURL:    imageURL,
		ProviderRef: providerRef,
	}, nil
}

// VerifyWebhook validates that signature is a valid hex-encoded HMAC-SHA256 of
// body computed with the shared secret. Comparison is constant-time.
func (p *MockQRProvider) VerifyWebhook(_ context.Context, signature string, body []byte) error {
	if strings.TrimSpace(p.WebhookSecret) == "" {
		return fmt.Errorf("qr webhook: no shared secret configured")
	}
	expected := ComputeQRWebhookSignature(p.WebhookSecret, body)

	// Accept an optional "sha256=" prefix for compatibility with common tooling.
	got := strings.TrimSpace(signature)
	got = strings.TrimPrefix(got, "sha256=")

	gotBytes, err := hex.DecodeString(got)
	if err != nil {
		return fmt.Errorf("qr webhook: signature is not valid hex")
	}
	expBytes, _ := hex.DecodeString(expected)
	if !hmac.Equal(gotBytes, expBytes) {
		return fmt.Errorf("qr webhook: signature mismatch")
	}
	return nil
}

// ComputeQRWebhookSignature returns the hex-encoded HMAC-SHA256 of body using
// secret. Exposed so tests and frontend simulations can produce a valid
// signature that VerifyWebhook will accept.
func ComputeQRWebhookSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}
