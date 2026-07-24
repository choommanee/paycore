package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// SessionCookieName is the dashboard session cookie.
const SessionCookieName = "pc_session"

// LocalUserID is the ctx local holding the authenticated dashboard user id.
const LocalUserID = "user_id"

// SessionAuth authenticates a dashboard user from the pc_session cookie. On
// success it sets LocalMerchantID and LocalUserID (both uuid.UUID) so downstream
// handlers scope queries exactly as the API-key path does; on any failure it
// returns 401 without leaking why.
func SessionAuth(mgr *session.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Cookies(SessionCookieName)
		if raw == "" {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		}
		claims, err := mgr.Verify(raw)
		if err != nil {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		}
		c.Locals(LocalMerchantID, claims.MerchantID)
		c.Locals(LocalUserID, claims.UserID)
		return c.Next()
	}
}

// UserIDFromCtx returns the authenticated dashboard user id set by SessionAuth.
func UserIDFromCtx(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalUserID).(uuid.UUID)
	return id, ok
}

// SetSessionCookie writes the pc_session cookie. secure should be true in prod.
func SetSessionCookie(c *fiber.Ctx, token string, ttl time.Duration, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

// ClearSessionCookie expires the pc_session cookie (logout).
func ClearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}
