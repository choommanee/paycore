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

// CheckoutSupportedMethods is the set of methods the hosted checkout can process.
// Order is the display order on the payment page. card (sandbox raw PAN) and
// promptpay (all modes) come from Phase 3; the six wallet / redirect methods are
// Phase 4 MOCK adapters (sandbox-gated) that complete via /confirm-mock.
var CheckoutSupportedMethods = []string{
	"card", "promptpay",
	"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment",
}

// CheckoutWalletMethods are the Phase 4 e-wallet / redirect methods. They share a
// single MOCK adapter (service.payWallet): no card data, no PaymentService /
// QRService call, no money moved — the session goes to requires_action and is
// completed by the sandbox-only /confirm-mock endpoint. In production (sandbox
// off) they are refused (ErrCheckoutMethodUnavailable); real PSP wiring is out of
// scope for this phase.
var CheckoutWalletMethods = []string{
	"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment",
}

// IsWalletMethod reports whether m is a Phase 4 wallet / redirect method handled
// by the mock wallet adapter.
func IsWalletMethod(m string) bool {
	for _, w := range CheckoutWalletMethods {
		if w == m {
			return true
		}
	}
	return false
}

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
// when Method == "card" (enforced in the service). Wallet methods carry no data.
type CheckoutPayRequest struct {
	Method        string     `json:"method" validate:"required,oneof=card promptpay mobile_banking truemoney shopeepay alipay wechat card_installment"`
	Card          *CardInput `json:"card" validate:"omitempty"`
	CustomerEmail string     `json:"customer_email" validate:"omitempty,email"`
}

// CheckoutConfirmMockRequest is the body of the sandbox-only confirm-mock endpoint
// that simulates a wallet approve (true) or decline (false).
type CheckoutConfirmMockRequest struct {
	Approve bool `json:"approve"`
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
