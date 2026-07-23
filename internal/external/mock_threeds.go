package external

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// MockThreeDS is a deterministic implementation of the ThreeDS interface for
// sandbox and test use. It never contacts a real ACS/Directory Server.
//
// Behaviour is driven by the PAN's last4 so tests can force a challenge or a
// frictionless success:
//
//	last4 == "0119" -> a challenge is required; a challengeURL is returned and
//	                   the cryptogram is empty (caller must complete the
//	                   challenge, then finalize via Cryptogram()).
//	last4 == "0000" -> authentication fails (simulated card not enrolled).
//	otherwise       -> frictionless success; a deterministic cryptogram is
//	                   returned immediately.
//
// The returned challengeURL points at BaseURL so a frontend simulation can
// render a fake ACS page.
//
// SECURITY: never log the PAN. Only last4 and derived, non-reversible values
// are emitted.
type MockThreeDS struct {
	// BaseURL is the origin used to build the simulated challenge URL. When
	// empty a sensible default is used.
	BaseURL string
}

// NewMockThreeDS returns a MockThreeDS. An empty baseURL falls back to a local
// sandbox default.
func NewMockThreeDS(baseURL string) *MockThreeDS {
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://sandbox.3ds.example.com"
	}
	return &MockThreeDS{BaseURL: baseURL}
}

// compile-time assertion.
var _ ThreeDS = (*MockThreeDS)(nil)

// Authenticate implements the ThreeDS interface. See the type doc for the
// deterministic rules keyed off the PAN last4.
func (m *MockThreeDS) Authenticate(_ context.Context, pan string, amount decimal.Decimal, currency, returnURL string) (challengeURL string, cryptogram string, err error) {
	l4 := last4(pan)
	switch l4 {
	case "0000":
		return "", "", fmt.Errorf("3ds: authentication failed (card not enrolled)")
	case "0119":
		// Challenge flow: hand back a URL the shopper is redirected to. No
		// cryptogram yet — it is minted once the challenge succeeds.
		return m.challengeURL(l4, amount, currency, returnURL), "", nil
	default:
		// Frictionless success — cryptogram issued immediately.
		return "", m.Cryptogram(pan, amount, currency), nil
	}
}

// Cryptogram derives a deterministic, non-reversible CAVV/AAV-style value from
// the transaction inputs. Real providers return an issuer-signed cryptogram;
// this stand-in lets tests assert a stable value and lets a frontend simulation
// complete a challenge. The full PAN is hashed and never returned or logged.
func (m *MockThreeDS) Cryptogram(pan string, amount decimal.Decimal, currency string) string {
	seed := fmt.Sprintf("%s|%s|%s", pan, amount.String(), strings.ToUpper(currency))
	sum := sha256.Sum256([]byte(seed))
	// A CAVV is 20 bytes; expose the first 20 bytes hex-encoded.
	return hex.EncodeToString(sum[:20])
}

// challengeURL builds a simulated ACS challenge URL. It intentionally omits the
// PAN; only the last4 and a return URL are included.
func (m *MockThreeDS) challengeURL(l4 string, amount decimal.Decimal, currency, returnURL string) string {
	return fmt.Sprintf(
		"%s/challenge?last4=%s&amount=%s&currency=%s&return_url=%s",
		strings.TrimRight(m.BaseURL, "/"),
		l4,
		amount.String(),
		strings.ToUpper(currency),
		returnURL,
	)
}
