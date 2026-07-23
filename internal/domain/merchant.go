package domain

import (
	"time"

	"github.com/google/uuid"
)

// CreateMerchantRequest onboards a new merchant. The API key is generated
// server-side, returned exactly once in the create response, and only its hash
// is persisted.
type CreateMerchantRequest struct {
	Name               string `json:"name" validate:"required,max=200"`
	MCC                string `json:"mcc" validate:"omitempty,len=4"`
	SettlementCurrency string `json:"settlement_currency" validate:"required,len=3"`
}

// Merchant is the merchant profile returned by the API. The API key hash is
// never exposed.
type Merchant struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Status             string    `json:"status"`
	MCC                string    `json:"mcc,omitempty"`
	SettlementCurrency string    `json:"settlement_currency"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// MerchantCredential is returned once on onboarding: it carries the merchant
// profile plus the plaintext API key. The key is not retrievable afterwards.
type MerchantCredential struct {
	Merchant
	APIKey string `json:"api_key"` // shown once; store it securely
}
