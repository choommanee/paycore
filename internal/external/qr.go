package external

import (
	"context"

	"github.com/shopspring/decimal"
)

// QRGenRequest asks a provider to register an expected payment and (for rails
// that mint server-side, e.g. a bank's dynamic-QR API) return a payload.
type QRGenRequest struct {
	Method    string          // domain.QRMethod
	Amount    decimal.Decimal
	Currency  string
	Reference string          // merchant reference, used to reconcile the webhook
	ExpirySec int
}

type QRGenResult struct {
	Payload    string // EMVCo QR string
	ImageURL   string // optional pre-rendered image
	ProviderRef string
}

// QRProvider abstracts a QR acquirer / bank rail (PromptPay via ITMX, a
// card-scheme QR gateway, or a cross-border switch). Swapping providers — or
// generating PromptPay locally vs. via a bank API — is just another impl.
type QRProvider interface {
	// Generate registers the expected payment upstream and returns the payload.
	// For locally-built PromptPay QR this may just echo a payload the service
	// already built, while still registering the reference for reconciliation.
	Generate(ctx context.Context, req QRGenRequest) (*QRGenResult, error)

	// VerifyWebhook validates the signature/HMAC of an inbound confirmation and
	// returns the raw verified body. Returns an error if the signature is bad.
	VerifyWebhook(ctx context.Context, signature string, body []byte) error
}
