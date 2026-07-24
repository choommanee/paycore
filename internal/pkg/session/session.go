// Package session issues and verifies compact, HMAC-SHA256-signed session
// tokens for the merchant dashboard. Format: base64url(payloadJSON).base64url(mac)
// where mac = HMAC-SHA256(secret, base64url(payloadJSON)). No external deps.
package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned for a malformed token or a signature mismatch.
	ErrInvalidToken = errors.New("session: invalid token")
	// ErrExpired is returned when the token's exp is in the past.
	ErrExpired = errors.New("session: token expired")
)

// Claims is the authenticated identity carried by a session token.
type Claims struct {
	UserID     uuid.UUID `json:"uid"`
	MerchantID uuid.UUID `json:"mid"`
	Email      string    `json:"email"`
}

// payload is the wire form: claims plus issued/expiry unix seconds.
type payload struct {
	Claims
	IssuedAt int64 `json:"iat"`
	Expires  int64 `json:"exp"`
}

// Manager signs and verifies tokens with a fixed secret and TTL.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager returns a Manager. ttl is the token lifetime; a non-positive ttl
// yields tokens that are already expired (used in tests).
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

var enc = base64.RawURLEncoding

// Issue returns a signed token for the claims, expiring after the manager TTL.
func (m *Manager) Issue(c Claims) (string, error) {
	now := time.Now().UTC()
	body, err := json.Marshal(payload{
		Claims:   c,
		IssuedAt: now.Unix(),
		Expires:  now.Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	b64 := enc.EncodeToString(body)
	return b64 + "." + enc.EncodeToString(m.mac(b64)), nil
}

// Verify checks the signature and expiry and returns the claims.
func (m *Manager) Verify(token string) (*Claims, error) {
	b64, sig, ok := strings.Cut(token, ".")
	if !ok {
		return nil, ErrInvalidToken
	}
	gotMAC, err := enc.DecodeString(sig)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(gotMAC, m.mac(b64)) {
		return nil, ErrInvalidToken
	}
	body, err := enc.DecodeString(b64)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().UTC().Unix() >= p.Expires {
		return nil, ErrExpired
	}
	c := p.Claims
	return &c, nil
}

func (m *Manager) mac(b64 string) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write([]byte(b64))
	return h.Sum(nil)
}
