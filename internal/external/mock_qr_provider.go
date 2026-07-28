package external

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// webhookReplayTolerance is the maximum absolute clock skew allowed between the
// timestamp bound into a v1 inbound signature and server time. A captured
// callback whose timestamp is older (or newer) than this window is rejected as a
// replay. Mirrors the Stripe-style 5-minute default used by the outbound worker.
const webhookReplayTolerance = 5 * time.Minute

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

// VerifyWebhook validates an inbound webhook signature against the shared
// secret. Two forms are accepted:
//
//	v1 (preferred): "t=<unix>,v1=<hex HMAC(secret,"<t>.<rawBody>")>"
//	                — binds the timestamp into the MAC and is rejected when
//	                  |now - t| exceeds webhookReplayTolerance, so a captured
//	                  callback cannot be replayed later (mirrors the outbound
//	                  X-PayCore-Signature scheme).
//	legacy         : bare "<hex>" or "sha256=<hex>" over the body only — kept for
//	                  backward compatibility with existing/older senders.
//
// All HMAC comparisons are constant-time.
func (p *MockQRProvider) VerifyWebhook(_ context.Context, signature string, body []byte) error {
	if strings.TrimSpace(p.WebhookSecret) == "" {
		return fmt.Errorf("qr webhook: no shared secret configured")
	}
	sig := strings.TrimSpace(signature)

	// Preferred timestamped form takes precedence when the header is self-describing.
	if ts, v1, ok := parseTimestampedSig(sig); ok {
		return p.verifyTimestamped(ts, v1, body, time.Now())
	}
	return p.verifyLegacy(sig, body)
}

// verifyTimestamped checks the v1 scheme: it first rejects a timestamp outside
// the replay-tolerance window, then recomputes HMAC(secret, "<t>.<body>") and
// compares in constant time. Because t is bound into the MAC, a forged/altered
// timestamp fails the HMAC check.
func (p *MockQRProvider) verifyTimestamped(ts int64, v1 string, body []byte, now time.Time) error {
	skew := now.Unix() - ts
	if skew < 0 {
		skew = -skew
	}
	if skew > int64(webhookReplayTolerance/time.Second) {
		return fmt.Errorf("qr webhook: timestamp outside tolerance (possible replay)")
	}
	gotBytes, err := hex.DecodeString(v1)
	if err != nil {
		return fmt.Errorf("qr webhook: signature is not valid hex")
	}
	expBytes, _ := hex.DecodeString(computeQRWebhookV1(p.WebhookSecret, ts, body))
	if !hmac.Equal(gotBytes, expBytes) {
		return fmt.Errorf("qr webhook: signature mismatch")
	}
	return nil
}

// verifyLegacy checks the body-only form (bare hex or "sha256=<hex>").
func (p *MockQRProvider) verifyLegacy(sig string, body []byte) error {
	got := strings.TrimPrefix(sig, "sha256=")
	gotBytes, err := hex.DecodeString(got)
	if err != nil {
		return fmt.Errorf("qr webhook: signature is not valid hex")
	}
	expBytes, _ := hex.DecodeString(ComputeQRWebhookSignature(p.WebhookSecret, body))
	if !hmac.Equal(gotBytes, expBytes) {
		return fmt.Errorf("qr webhook: signature mismatch")
	}
	return nil
}

// parseTimestampedSig parses a "t=<unix>,v1=<hex>" header (keys in any order).
// It returns ok=false for anything that is not the timestamped form (bare hex,
// "sha256=<hex>", or a malformed header), so the caller falls back to legacy
// verification.
func parseTimestampedSig(sig string) (ts int64, v1 string, ok bool) {
	if !strings.Contains(sig, "=") {
		return 0, "", false
	}
	var tsStr string
	for _, part := range strings.Split(sig, ",") {
		k, v, found := strings.Cut(strings.TrimSpace(part), "=")
		if !found {
			return 0, "", false
		}
		switch k {
		case "t":
			tsStr = v
		case "v1":
			v1 = v
		}
	}
	if tsStr == "" || v1 == "" {
		return 0, "", false
	}
	ts, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return 0, "", false
	}
	return ts, v1, true
}

// ComputeQRWebhookSignature returns the hex-encoded HMAC-SHA256 of body using
// secret (the legacy body-only form). Exposed so tests and frontend simulations
// can produce a signature that VerifyWebhook still accepts for compatibility.
func ComputeQRWebhookSignature(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// ComputeQRWebhookSignatureV1 returns the preferred, replay-resistant inbound
// signature header value "t=<ts>,v1=<hex HMAC(secret,"<ts>.<body>")>". This is
// exactly the scheme the outbound worker emits in X-PayCore-Signature; the
// sandbox payer-simulator signs with it so its confirmations pass the timestamp
// check.
func ComputeQRWebhookSignatureV1(secret string, ts int64, body []byte) string {
	return "t=" + strconv.FormatInt(ts, 10) + ",v1=" + computeQRWebhookV1(secret, ts, body)
}

// computeQRWebhookV1 is the hex HMAC-SHA256 of "<ts>.<body>" under secret.
func computeQRWebhookV1(secret string, ts int64, body []byte) string {
	tsStr := strconv.FormatInt(ts, 10)
	signed := make([]byte, 0, len(tsStr)+1+len(body))
	signed = append(signed, tsStr...)
	signed = append(signed, '.')
	signed = append(signed, body...)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(signed)
	return hex.EncodeToString(mac.Sum(nil))
}
