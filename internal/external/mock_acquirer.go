package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// MockAcquirer is a deterministic, in-memory implementation of Acquirer for use
// in local development, sandbox environments and tests. It does NOT talk to any
// real card network. Behaviour is driven entirely by the test PAN / last4 so
// callers can reproduce approvals, declines and 3DS challenges without secrets.
//
// Test PAN conventions (matched on the last 4 digits):
//
//	last4 == "0000" -> declined (insufficient funds)
//	last4 == "0119" -> 3DS required (declined with decline code "3ds_required")
//	anything else   -> approved
//
// Decline codes are returned in DeclineCode; a successful auth carries an
// AuthCode and a freshly minted AcquirerRef (uuid) used to correlate follow-up
// Capture/Refund/Void calls.
//
// SECURITY: this type must never log the PAN or any card data.
type MockAcquirer struct{}

// NewMockAcquirer returns a ready-to-use MockAcquirer.
func NewMockAcquirer() *MockAcquirer { return &MockAcquirer{} }

// compile-time assertion.
var _ Acquirer = (*MockAcquirer)(nil)

// Decline codes surfaced by the mock. These loosely mirror the semantics of a
// real acquirer's response codes.
const (
	DeclineInsufficientFunds = "insufficient_funds"
	Decline3DSRequired       = "3ds_required"
)

// last4 returns the last four characters of the PAN, or the whole string when
// shorter. Only the last four digits are ever used, never the full PAN.
func last4(pan string) string {
	pan = strings.TrimSpace(pan)
	if len(pan) <= 4 {
		return pan
	}
	return pan[len(pan)-4:]
}

// brandFor infers a card brand from the PAN prefix. This is a coarse heuristic
// good enough for a sandbox; production relies on the acquirer's response.
func brandFor(pan string) string {
	pan = strings.TrimSpace(pan)
	switch {
	case strings.HasPrefix(pan, "4"):
		return "visa"
	case strings.HasPrefix(pan, "5"), strings.HasPrefix(pan, "2"):
		return "mastercard"
	case strings.HasPrefix(pan, "34"), strings.HasPrefix(pan, "37"):
		return "amex"
	default:
		return "unknown"
	}
}

// Authorize deterministically approves or declines based on the PAN's last4.
func (m *MockAcquirer) Authorize(_ context.Context, req AuthRequest) (*AuthResult, error) {
	l4 := last4(req.PAN)
	brand := brandFor(req.PAN)

	switch l4 {
	case "0000":
		return &AuthResult{
			Approved:    false,
			DeclineCode: DeclineInsufficientFunds,
			CardBrand:   brand,
			Last4:       l4,
		}, nil
	case "0119":
		// A 3DS-required response: the caller should route the shopper through
		// the ThreeDS provider and re-authorize with the resulting cryptogram.
		// If a cryptogram is already present we treat it as satisfied.
		if strings.TrimSpace(req.ThreeDSCryptogram) == "" {
			return &AuthResult{
				Approved:    false,
				DeclineCode: Decline3DSRequired,
				CardBrand:   brand,
				Last4:       l4,
			}, nil
		}
	}

	return &AuthResult{
		Approved:    true,
		AuthCode:    genAuthCode(),
		AcquirerRef: uuid.NewString(),
		CardBrand:   brand,
		Last4:       l4,
	}, nil
}

// Capture settles a previously authorized transaction. The mock always
// succeeds for a non-empty reference and echoes a new AuthCode.
func (m *MockAcquirer) Capture(_ context.Context, acquirerRef string, _ decimal.Decimal) (*AuthResult, error) {
	if strings.TrimSpace(acquirerRef) == "" {
		return nil, fmt.Errorf("mock acquirer: capture requires an acquirer ref")
	}
	return &AuthResult{
		Approved:    true,
		AuthCode:    genAuthCode(),
		AcquirerRef: acquirerRef,
	}, nil
}

// Refund returns funds against a captured transaction. Always succeeds for a
// non-empty reference in the mock.
func (m *MockAcquirer) Refund(_ context.Context, acquirerRef string, _ decimal.Decimal) (*AuthResult, error) {
	if strings.TrimSpace(acquirerRef) == "" {
		return nil, fmt.Errorf("mock acquirer: refund requires an acquirer ref")
	}
	return &AuthResult{
		Approved:    true,
		AuthCode:    genAuthCode(),
		AcquirerRef: acquirerRef,
	}, nil
}

// Void cancels an authorization that has not been captured. Always succeeds for
// a non-empty reference in the mock.
func (m *MockAcquirer) Void(_ context.Context, acquirerRef string) (*AuthResult, error) {
	if strings.TrimSpace(acquirerRef) == "" {
		return nil, fmt.Errorf("mock acquirer: void requires an acquirer ref")
	}
	return &AuthResult{
		Approved:    true,
		AuthCode:    genAuthCode(),
		AcquirerRef: acquirerRef,
	}, nil
}

// genAuthCode returns a short, deterministic-looking 6-char auth code derived
// from a uuid. Not card data; safe to log.
func genAuthCode() string {
	return strings.ToUpper(uuid.NewString()[:6])
}
