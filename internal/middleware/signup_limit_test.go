package middleware

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
)

// TestSignupRateLimiterBlocksAfterMax asserts the public onboarding limiter
// permits exactly `max` requests from a client IP and then returns 429 in the
// standard error envelope (code RATE_LIMITED). This mirrors how the limiter is
// mounted on POST /v1/merchants in router.Setup.
func TestSignupRateLimiterBlocksAfterMax(t *testing.T) {
	const max = 5
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	// Mount ONLY on the signup route, exactly as the router does.
	app.Post("/v1/merchants", SignupRateLimiter(max), func(c *fiber.Ctx) error {
		return domain.Created(c, fiber.Map{"ok": true})
	})

	// The first `max` requests succeed (201); the next is limited (429).
	for i := 1; i <= max; i++ {
		req := httptest.NewRequest(fiber.MethodPost, "/v1/merchants", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("request %d: status=%d want 201", i, resp.StatusCode)
		}
		resp.Body.Close()
	}

	// The (max+1)-th request from the same IP must be rate limited.
	req := httptest.NewRequest(fiber.MethodPost, "/v1/merchants", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("over-limit request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("over-limit status=%d want 429", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	if err := json.Unmarshal(body, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v (%s)", err, body)
	}
	if env.Success {
		t.Fatalf("429 envelope success=true, want false: %s", body)
	}
	if env.Code != "RATE_LIMITED" {
		t.Fatalf("code=%q want RATE_LIMITED: %s", env.Code, body)
	}
}

// TestSignupRateLimiterDefaultsWhenNonPositive asserts a non-positive limit
// falls back to the safe default (5) rather than blocking everything or nothing.
func TestSignupRateLimiterDefaultsWhenNonPositive(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Post("/v1/merchants", SignupRateLimiter(0), func(c *fiber.Ctx) error {
		return domain.Created(c, fiber.Map{"ok": true})
	})

	// 5 succeed, 6th blocked (default fallback = 5).
	for i := 1; i <= 5; i++ {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/v1/merchants", nil))
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("request %d: status=%d want 201", i, resp.StatusCode)
		}
		resp.Body.Close()
	}
	resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, "/v1/merchants", nil))
	if err != nil {
		t.Fatalf("over-limit request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("default-limit over-limit status=%d want 429", resp.StatusCode)
	}
}
