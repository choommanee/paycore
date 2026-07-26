package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// MerchantAuth authenticates a merchant from EITHER a dashboard session cookie
// (pc_session) OR an API key, resolving the same merchant context so a route can
// serve both the cookie-based dashboard and API-key clients. The session cookie
// is tried first (dashboard is the common caller); a valid cookie also sets
// LocalUserID. On neither credential it returns 401 without leaking which was
// missing.
func MerchantAuth(mgr *session.Manager, resolver MerchantResolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Session cookie.
		if raw := c.Cookies(SessionCookieName); raw != "" {
			if claims, err := mgr.Verify(raw); err == nil {
				c.Locals(LocalMerchantID, claims.MerchantID)
				c.Locals(LocalUserID, claims.UserID)
				return c.Next()
			}
		}
		// 2. API key.
		if rawKey := extractAPIKey(c); rawKey != "" {
			merchant, err := resolver.ResolveByAPIKeyHash(c.UserContext(), HashAPIKey(rawKey))
			if err == nil && merchant != nil {
				c.Locals(LocalMerchant, merchant)
				c.Locals(LocalMerchantID, merchant.ID)
				return c.Next()
			}
			if err != nil && !errors.Is(err, domain.ErrUnauthorized) {
				return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			}
		}
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}
}
