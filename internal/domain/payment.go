package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// PaymentStatus follows the classic auth/capture lifecycle.
type PaymentStatus string

const (
	StatusRequiresAction  PaymentStatus = "requires_action"  // 3DS challenge pending
	StatusAuthorized      PaymentStatus = "authorized"       // funds held, not captured
	StatusCaptured        PaymentStatus = "captured"         // funds captured
	StatusPartialRefunded PaymentStatus = "partial_refunded" //
	StatusRefunded        PaymentStatus = "refunded"         //
	StatusVoided          PaymentStatus = "voided"           // auth reversed before capture
	StatusFailed          PaymentStatus = "failed"           //
)

// CreatePaymentRequest is the merchant-facing charge/authorize request.
// NOTE: raw PAN never reaches this struct in production — the merchant sends a
// single-use payment token created client-side against the tokenization vault,
// keeping the gateway API layer outside PCI-DSS cardholder-data scope.
type CreatePaymentRequest struct {
	MerchantID   uuid.UUID       `json:"merchant_id" validate:"required"`
	Amount       decimal.Decimal `json:"amount" validate:"required"`
	Currency     string          `json:"currency" validate:"required,len=3"`
	PaymentToken string          `json:"payment_token" validate:"required"` // vault token, not PAN
	Capture      bool            `json:"capture"`                           // true = charge, false = authorize only
	Reference    string          `json:"reference" validate:"max=64"`
	ReturnURL    string          `json:"return_url" validate:"omitempty,url"` // for 3DS redirect
	Metadata     map[string]any  `json:"metadata,omitempty"`
}

type CaptureRequest struct {
	Amount decimal.Decimal `json:"amount" validate:"required"` // supports partial capture
}

type RefundRequest struct {
	Amount decimal.Decimal `json:"amount" validate:"required"`
	Reason string          `json:"reason" validate:"max=140"`
}

// Payment is the persisted aggregate root.
type Payment struct {
	ID              uuid.UUID       `json:"id"`
	MerchantID      uuid.UUID       `json:"merchant_id"`
	Amount          decimal.Decimal `json:"amount"`
	CapturedAmount  decimal.Decimal `json:"captured_amount"`
	RefundedAmount  decimal.Decimal `json:"refunded_amount"`
	Currency        string          `json:"currency"`
	Status          PaymentStatus   `json:"status"`
	CardBrand       string          `json:"card_brand,omitempty"` // e.g. visa
	CardLast4       string          `json:"card_last4,omitempty"` // last 4 only — never full PAN
	AcquirerRef     string          `json:"acquirer_ref,omitempty"`
	AuthCode        string          `json:"auth_code,omitempty"`
	NextActionURL   string          `json:"next_action_url,omitempty"` // 3DS redirect
	Reference       string          `json:"reference,omitempty"`
	IdempotencyKey  string          `json:"-"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
