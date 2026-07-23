package middleware

import (
	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/domain"
)

// RequireIdempotencyKey enforces the Idempotency-Key header on all money-moving
// POST endpoints. The service layer persists (merchant_id, key) -> response so a
// retried request returns the original result instead of double-charging.
func RequireIdempotencyKey() fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("Idempotency-Key")
		if key == "" {
			return domain.Error(c, fiber.StatusBadRequest, "MISSING_IDEMPOTENCY_KEY",
				"Idempotency-Key header is required for this endpoint")
		}
		c.Locals("idempotency_key", key)
		return c.Next()
	}
}
