package middleware

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/yourco/payment-gateway/internal/domain"
)

// Locals keys set by the auth middleware for downstream handlers.
const (
	LocalMerchant   = "merchant"    // *domain.Merchant
	LocalMerchantID = "merchant_id" // uuid.UUID
)

// MerchantResolver resolves an authenticated merchant from the presented API
// key hash. The service/repository layer implements this; the middleware stays
// decoupled from sqlc-generated types.
type MerchantResolver interface {
	// ResolveByAPIKeyHash returns the active merchant whose stored api_key_hash
	// matches the given hex-encoded SHA-256 hash, or ErrUnauthorized.
	ResolveByAPIKeyHash(ctx context.Context, apiKeyHash string) (*domain.Merchant, error)
}

// HashAPIKey returns the hex-encoded SHA-256 of a raw API key. The same function
// is used at onboarding (to store the hash) and at auth (to compare), so the two
// always agree. The raw key is never logged or persisted.
func HashAPIKey(rawKey string) string {
	sum := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(sum[:])
}

// APIKeyAuth authenticates a merchant by API key. The key is taken from the
// Authorization header (Bearer <key> or a bare token) or the X-API-Key header,
// hashed, and matched against merchants.api_key_hash. On success the resolved
// merchant and its id are stored in ctx locals for handlers to scope queries;
// on any failure it returns 401 without leaking whether the merchant exists.
//
// SECURITY: the raw API key is never logged.
func APIKeyAuth(resolver MerchantResolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		rawKey := extractAPIKey(c)
		if rawKey == "" {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "missing API key")
		}

		merchant, err := resolver.ResolveByAPIKeyHash(c.UserContext(), HashAPIKey(rawKey))
		if err != nil || merchant == nil {
			if err != nil && !errors.Is(err, domain.ErrUnauthorized) {
				// A real backend error (DB down) — surface as unauthorized to the
				// caller but keep the cause out of the response.
				return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid API key")
			}
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid API key")
		}

		c.Locals(LocalMerchant, merchant)
		c.Locals(LocalMerchantID, merchant.ID)
		return c.Next()
	}
}

// AdminAuth authorizes the platform operator console. The presented X-Admin-Key
// header is compared against the configured admin key with a constant-time
// compare (crypto/subtle) so the check is not a timing oracle. When the admin
// key is unset (empty config) the whole admin surface is disabled — every
// request is rejected with 401 rather than silently allowing access.
//
// SECURITY: the admin key is never logged; a mismatch and an unset key are
// indistinguishable to the caller.
func AdminAuth(adminKey string) fiber.Handler {
	want := strings.TrimSpace(adminKey)
	return func(c *fiber.Ctx) error {
		got := strings.TrimSpace(c.Get("X-Admin-Key"))
		// Reject when the admin surface is disabled (no key configured) or the
		// presented key is empty, without running the compare against "".
		if want == "" || got == "" {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "admin access denied")
		}
		if subtle.ConstantTimeCompare([]byte(got), []byte(want)) != 1 {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "admin access denied")
		}
		return c.Next()
	}
}

// extractAPIKey pulls the raw API key from the request. Preference order:
// Authorization: Bearer <key>, then a bare Authorization value, then X-API-Key.
func extractAPIKey(c *fiber.Ctx) string {
	if h := strings.TrimSpace(c.Get("Authorization")); h != "" {
		if len(h) > 7 && strings.EqualFold(h[:7], "Bearer ") {
			return strings.TrimSpace(h[7:])
		}
		return h
	}
	return strings.TrimSpace(c.Get("X-API-Key"))
}

// MerchantIDFromCtx returns the authenticated merchant id set by APIKeyAuth.
func MerchantIDFromCtx(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalMerchantID).(uuid.UUID)
	return id, ok
}

// MerchantFromCtx returns the authenticated merchant set by APIKeyAuth.
func MerchantFromCtx(c *fiber.Ctx) (*domain.Merchant, bool) {
	m, ok := c.Locals(LocalMerchant).(*domain.Merchant)
	return m, ok
}
