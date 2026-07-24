package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthIdentity is the normalized profile returned by an OAuth/OIDC provider
// after a successful code exchange.
type OAuthIdentity struct {
	Subject string `json:"subject"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// MerchantUser is a human who logs into the dashboard, scoped to one merchant.
type MerchantUser struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Provider   string    `json:"provider"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AuthMe is the GET /v1/auth/me response: the logged-in user plus enough
// merchant context for the dashboard header.
type AuthMe struct {
	UserID       uuid.UUID `json:"user_id"`
	MerchantID   uuid.UUID `json:"merchant_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	MerchantName string    `json:"merchant_name"`
}
