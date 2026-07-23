package middleware

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/config"
	"github.com/yourco/payment-gateway/internal/domain"
)

// testConfig returns a minimal dev config (CORS disabled, not prod) for
// middleware tests.
func testConfig() *config.Config {
	return &config.Config{Env: "development"}
}

// declaredDomainErrors are every sentinel error a service may return. Each must
// map to a NON-500 status in ErrorHandler so genuine client outcomes (declines,
// insufficient funds, 3DS challenges, idempotency reuse) never surface as a
// generic 500 that breaks merchant retry/alerting.
var declaredDomainErrors = []error{
	domain.ErrInvalidRequest,
	domain.ErrPaymentNotFound,
	domain.ErrDuplicateRequest,
	domain.ErrIdempotencyKeyReuse,
	domain.ErrCardDeclined,
	domain.ErrInsufficientFund,
	domain.ErrThreeDSRequired,
	domain.ErrRefundExceeds,
	domain.ErrInvalidState,
	domain.ErrUnauthorized,
	domain.ErrMerchantNotFound,
	domain.ErrDisputeNotFound,
	domain.ErrForbidden,
}

// TestErrorHandlerNoDomainErrorFallsThroughTo500 asserts that no declared domain
// error is handled by the default (500) branch.
func TestErrorHandlerNoDomainErrorFallsThroughTo500(t *testing.T) {
	for _, derr := range declaredDomainErrors {
		derr := derr
		t.Run(derr.Error(), func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
			app.Get("/boom", func(_ *fiber.Ctx) error { return derr })

			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/boom", nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode == fiber.StatusInternalServerError {
				body, _ := io.ReadAll(resp.Body)
				t.Fatalf("domain error %q fell through to 500: %s", derr, body)
			}
			if resp.StatusCode < 400 || resp.StatusCode >= 500 {
				t.Fatalf("domain error %q -> status %d, want a 4xx client status", derr, resp.StatusCode)
			}
		})
	}
}

// TestErrorHandlerSpecificStatuses locks the important client-outcome mappings.
func TestErrorHandlerSpecificStatuses(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{domain.ErrInsufficientFund, fiber.StatusPaymentRequired},
		{domain.ErrThreeDSRequired, fiber.StatusConflict},
		{domain.ErrRefundExceeds, fiber.StatusUnprocessableEntity},
		{domain.ErrIdempotencyKeyReuse, fiber.StatusConflict},
		{domain.ErrInvalidState, fiber.StatusConflict},
		{domain.ErrInvalidRequest, fiber.StatusBadRequest},
		{domain.ErrPaymentNotFound, fiber.StatusNotFound},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.err.Error(), func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
			app.Get("/e", func(_ *fiber.Ctx) error { return tc.err })
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/e", nil))
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.want {
				t.Fatalf("%q -> %d, want %d", tc.err, resp.StatusCode, tc.want)
			}
		})
	}
}

// TestSecurityHeadersPresent verifies helmet emits the expected headers.
func TestSecurityHeadersPresent(t *testing.T) {
	app := fiber.New()
	// Reuse the same helmet config Register installs.
	Register(app, testConfig(), zerolog.Nop())
	app.Get("/ok", func(c *fiber.Ctx) error { return c.SendString("ok") })

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/ok", nil))
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	if got := resp.Header.Get("X-Content-Type-Options"); got != "nosniff" {
		t.Fatalf("X-Content-Type-Options=%q want nosniff", got)
	}
	if got := resp.Header.Get("X-Frame-Options"); got != "DENY" {
		t.Fatalf("X-Frame-Options=%q want DENY", got)
	}
}
