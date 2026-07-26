package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func TestCheckoutRateLimiterBlocksOverLimit(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Post("/c", CheckoutRateLimiter(1), func(c *fiber.Ctx) error { return c.SendStatus(200) })

	r1, _ := app.Test(httptest.NewRequest("POST", "/c", nil))
	if r1.StatusCode != 200 {
		t.Fatalf("first = %d want 200", r1.StatusCode)
	}
	r2, _ := app.Test(httptest.NewRequest("POST", "/c", nil))
	if r2.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("second = %d want 429", r2.StatusCode)
	}
}
