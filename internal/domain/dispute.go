package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// DisputeStatus is the chargeback/dispute state machine:
//
//	opened -> under_review -> won | lost | accepted
type DisputeStatus string

const (
	DisputeOpened      DisputeStatus = "opened"
	DisputeUnderReview DisputeStatus = "under_review"
	DisputeWon         DisputeStatus = "won"
	DisputeLost        DisputeStatus = "lost"
	DisputeAccepted    DisputeStatus = "accepted"
)

// OpenDisputeRequest opens a dispute against a payment. Amount defaults to the
// payment amount when omitted (full chargeback).
type OpenDisputeRequest struct {
	Amount     decimal.Decimal `json:"amount" validate:"omitempty"`
	Reason     string          `json:"reason" validate:"max=280"`
	NetworkRef string          `json:"network_ref" validate:"max=64"`
}

// Dispute is the persisted dispute aggregate.
type Dispute struct {
	ID          uuid.UUID       `json:"id"`
	PaymentID   uuid.UUID       `json:"payment_id"`
	MerchantID  uuid.UUID       `json:"merchant_id"`
	Amount      decimal.Decimal `json:"amount"`
	Currency    string          `json:"currency"`
	Reason      string          `json:"reason,omitempty"`
	Status      DisputeStatus   `json:"status"`
	NetworkRef  string          `json:"network_ref,omitempty"`
	OpenedAt    time.Time       `json:"opened_at"`
	ResolvedAt  *time.Time      `json:"resolved_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
