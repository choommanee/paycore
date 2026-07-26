package domain

import (
	"time"

	"github.com/google/uuid"
)

// CheckoutStatus is the lifecycle of a hosted-checkout session.
type CheckoutStatus string

const (
	CheckoutOpen           CheckoutStatus = "open"            // created, no method chosen
	CheckoutProcessing     CheckoutStatus = "processing"      // payment in flight
	CheckoutRequiresAction CheckoutStatus = "requires_action" // awaiting 3DS redirect or QR scan
	CheckoutPaid           CheckoutStatus = "paid"
	CheckoutFailed         CheckoutStatus = "failed"
	CheckoutExpired        CheckoutStatus = "expired"
)

// CheckoutSupportedMethods is the set of methods the hosted checkout can actually
// process in Phase 3: card (sandbox-gated, raw PAN) and promptpay (all modes).
// Wallet / redirect methods (Beam parity) arrive in Phase 4.
var CheckoutSupportedMethods = []string{"card", "promptpay"}

// DisplayMethods intersects a link's allowed methods with the methods the
// checkout can process, preserving CheckoutSupportedMethods order. An empty
// allowed list means "every method the merchant enabled", so it returns all
// supported methods.
func DisplayMethods(allowed []string) []string {
	if len(allowed) == 0 {
		out := make([]string, len(CheckoutSupportedMethods))
		copy(out, CheckoutSupportedMethods)
		return out
	}
	set := make(map[string]bool, len(allowed))
	for _, m := range allowed {
		set[m] = true
	}
	out := make([]string, 0, len(CheckoutSupportedMethods))
	for _, m := range CheckoutSupportedMethods {
		if set[m] {
			out = append(out, m)
		}
	}
	return out
}

// CheckoutSessionRequest creates a session from a payment link's public id.
type CheckoutSessionRequest struct {
	Link string `json:"link" validate:"required"`
}

// CardInput is a raw card entry. SANDBOX ONLY — a real deployment tokenizes card
// data in the browser via hosted fields and never sends a PAN to this server.
type CardInput struct {
	Number     string `json:"number" validate:"required,min=12,max=19"`
	ExpMonth   int    `json:"exp_month" validate:"required,min=1,max=12"`
	ExpYear    int    `json:"exp_year" validate:"required,min=2024,max=2100"`
	CVV        string `json:"cvv" validate:"required,min=3,max=4"`
	HolderName string `json:"holder_name" validate:"omitempty,max=200"`
}

// CheckoutPayRequest selects a method and carries its data. Card is required only
// when Method == "card" (enforced in the service, not the validator).
type CheckoutPayRequest struct {
	Method        string     `json:"method" validate:"required,oneof=card promptpay"`
	Card          *CardInput `json:"card" validate:"omitempty"`
	CustomerEmail string     `json:"customer_email" validate:"omitempty,email"`
}

// CheckoutSessionView is the PUBLIC, secret-free projection returned to the
// hosted checkout page. It NEVER carries session_token_hash, payment internals,
// or card data. Token is populated ONLY on session creation (the browser holds it
// thereafter); it is omitted on subsequent reads. QRPayload / NextActionURL are
// set only when the chosen method produced them.
type CheckoutSessionView struct {
	Token          string    `json:"session_token,omitempty"`
	ID             uuid.UUID `json:"id"`
	Status         string    `json:"status"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	MerchantName   string    `json:"merchant_name"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	ImageURL       string    `json:"image_url,omitempty"`
	AllowedMethods []string  `json:"allowed_methods"`
	SelectedMethod string    `json:"selected_method,omitempty"`
	QRPayload      string    `json:"qr_payload,omitempty"`      // PromptPay EMVCo string
	NextActionURL  string    `json:"next_action_url,omitempty"` // card 3DS redirect
	ReturnURL      string    `json:"return_url,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	Sandbox        bool      `json:"sandbox"` // card form is sandbox-only; drives the UI label
}
