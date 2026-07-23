package external

import (
	"context"

	"github.com/shopspring/decimal"
)

// AuthRequest is what the gateway sends to the upstream acquirer / card switch.
// In a real ISO 8583 integration these map to specific bitmap fields.
type AuthRequest struct {
	PAN         string // decrypted inside PCI scope only, never logged
	ExpiryMonth int
	ExpiryYear  int
	Amount      decimal.Decimal
	Currency    string
	Merchant    string
	ThreeDSCryptogram string // CAVV/AAV from 3DS
}

type AuthResult struct {
	Approved    bool
	AuthCode    string
	AcquirerRef string
	DeclineCode string
	CardBrand   string
	Last4       string
}

// Acquirer abstracts the upstream authorization network so the gateway can
// support multiple acquirers (bank switch, card network) behind one interface.
type Acquirer interface {
	Authorize(ctx context.Context, req AuthRequest) (*AuthResult, error)
	Capture(ctx context.Context, acquirerRef string, amount decimal.Decimal) (*AuthResult, error)
	Refund(ctx context.Context, acquirerRef string, amount decimal.Decimal) (*AuthResult, error)
	Void(ctx context.Context, acquirerRef string) (*AuthResult, error)
}

// ThreeDS abstracts a 3-D Secure 2.x authentication provider (ACS/DS).
type ThreeDS interface {
	Authenticate(ctx context.Context, pan string, amount decimal.Decimal, currency, returnURL string) (challengeURL string, cryptogram string, err error)
}
