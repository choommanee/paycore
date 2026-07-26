package domain

import (
	"time"

	"github.com/google/uuid"
)

// ValidPaymentMethods is the full registry of method slugs a link may allow.
// The backend only truly processes card + promptpay today (Phase 3); the rest
// are Beam-parity slugs backed by mock adapters (Phase 4).
var ValidPaymentMethods = []string{
	"card", "promptpay", "mobile_banking", "truemoney",
	"shopeepay", "alipay", "wechat", "card_installment",
}

// IsValidMethod reports whether m is a known payment-method slug.
func IsValidMethod(m string) bool {
	for _, v := range ValidPaymentMethods {
		if v == m {
			return true
		}
	}
	return false
}

// CreatePaymentLinkRequest is the dashboard/API payload to create a link. The
// merchant id and created_by come from auth context, never the body.
type CreatePaymentLinkRequest struct {
	Title          string     `json:"title" validate:"required,max=200"`
	Description    string     `json:"description" validate:"omitempty,max=2000"`
	AmountMinor    int64      `json:"amount_minor" validate:"required,gt=0"`
	Currency       string     `json:"currency" validate:"omitempty,len=3,alpha"`
	AllowedMethods []string   `json:"allowed_methods" validate:"omitempty,dive,oneof=card promptpay mobile_banking truemoney shopeepay alipay wechat card_installment"`
	LinkType       string     `json:"link_type" validate:"omitempty,oneof=single_use reusable"`
	Reference      string     `json:"reference" validate:"omitempty,max=200"`
	ImageURL       string     `json:"image_url" validate:"omitempty,http_url,max=500"`
	ExpiresAt      *time.Time `json:"expires_at" validate:"omitempty"`
}

// PaymentLink is the API representation of a payment link. URL is the computed
// public checkout URL the merchant shares.
type PaymentLink struct {
	ID             uuid.UUID  `json:"id"`
	MerchantID     uuid.UUID  `json:"merchant_id"`
	PublicID       string     `json:"public_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	AmountMinor    int64      `json:"amount_minor"`
	Currency       string     `json:"currency"`
	AllowedMethods []string   `json:"allowed_methods"`
	LinkType       string     `json:"link_type"`
	Status         string     `json:"status"`
	Reference      string     `json:"reference,omitempty"`
	ImageURL       string     `json:"image_url,omitempty"`
	URL            string     `json:"url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
