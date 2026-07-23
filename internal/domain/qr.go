package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// QRMethod is the rail a QR payment is settled over.
type QRMethod string

const (
	QRPromptPayDynamic QRMethod = "promptpay_dynamic" // one-time QR, amount embedded, webhook-confirmed
	QRPromptPayStatic  QRMethod = "promptpay_static"  // reusable QR, amount entered by payer
	QRCardScheme       QRMethod = "card_scheme"       // Visa/Mastercard QR over the acquiring rail
	QRCrossBorder      QRMethod = "cross_border"      // foreign wallet scanning a Thai QR (e.g. PromptPay x PayNow)
)

// QRStatus is the async lifecycle of a QR payment. Unlike cards there is no
// authorize/capture split — the payer pushes funds and the bank/PSP confirms.
type QRStatus string

const (
	QRAwaitingPayment QRStatus = "awaiting_payment" // QR shown, waiting for payer
	QRPaid            QRStatus = "paid"             // confirmed by bank/PSP webhook
	QRExpired         QRStatus = "expired"          // not paid before expiry
	QRFailed          QRStatus = "failed"           //
)

// CreateQRRequest is the merchant-facing request to mint a QR.
type CreateQRRequest struct {
	MerchantID uuid.UUID       `json:"merchant_id" validate:"required"`
	Method     QRMethod        `json:"method" validate:"required"`
	Amount     decimal.Decimal `json:"amount" validate:"required"`
	Currency   string          `json:"currency" validate:"required,len=3"`
	Reference  string          `json:"reference" validate:"max=64"`
	ExpiresIn  int             `json:"expires_in_seconds" validate:"omitempty,min=30,max=3600"` // default 900
}

// QRPayment is the persisted QR transaction.
type QRPayment struct {
	ID             uuid.UUID       `json:"id"`
	MerchantID     uuid.UUID       `json:"merchant_id"`
	Method         QRMethod        `json:"method"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
	Status         QRStatus        `json:"status"`
	Reference      string          `json:"reference,omitempty"`
	QRPayload      string          `json:"qr_payload"`                // EMVCo string to render as a QR image
	QRImageURL     string          `json:"qr_image_url,omitempty"`    // optional pre-rendered PNG/SVG
	PayerBank      string          `json:"payer_bank,omitempty"`      // filled on confirmation
	ProviderRef    string          `json:"provider_ref,omitempty"`    // bank/PSP transaction id
	CorrelationRef string          `json:"correlation_ref,omitempty"` // stable key an inbound webhook is matched against
	ExpiresAt      time.Time       `json:"expires_at"`
	PaidAt         *time.Time      `json:"paid_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// SandboxQRView is the PUBLIC, secret-free display projection of a QR payment
// returned to the sandbox payer simulator (the "Sandbox Bank" app). It carries
// only what the simulator needs to render a payment-confirmation screen — never
// the qr_payload, provider handles, or any secret. Amount is exposed as both a
// major-unit decimal (for display) and integer minor units (satang).
type SandboxQRView struct {
	ID           uuid.UUID       `json:"id"`
	Amount       decimal.Decimal `json:"amount"`       // major units, e.g. "1290.00"
	AmountMinor  int64           `json:"amount_minor"` // integer minor units (satang)
	Currency     string          `json:"currency"`
	MerchantName string          `json:"merchant_name"`
	Method       QRMethod        `json:"method"`
	Status       QRStatus        `json:"status"`
	Reference    string          `json:"reference,omitempty"`
}

// SandboxQRListItem is one row of the sandbox "incoming requests" list: a recent
// pending QR payment the simulator can tap to pay. Secret-free.
type SandboxQRListItem struct {
	ID           uuid.UUID       `json:"id"`
	Amount       decimal.Decimal `json:"amount"`
	AmountMinor  int64           `json:"amount_minor"`
	Currency     string          `json:"currency"`
	MerchantName string          `json:"merchant_name"`
	Method       QRMethod        `json:"method"`
	Status       QRStatus        `json:"status"`
	CreatedAt    time.Time       `json:"created_at"`
}

// QRWebhook is the inbound notification from the bank/PSP that a QR was paid.
// The signature is verified before the body is trusted (see QRProvider).
type QRWebhook struct {
	Reference   string          `json:"reference"`    // matches QRPayment.Reference
	ProviderRef string          `json:"provider_ref"` // bank/PSP txn id (idempotency)
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	PayerBank   string          `json:"payer_bank"`
	Status      string          `json:"status"` // "paid" | "failed"
	PaidAt      time.Time       `json:"paid_at"`
}
