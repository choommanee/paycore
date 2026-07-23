package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
)

// TestAdminAuth is table-driven over the X-Admin-Key gate: the correct key
// passes, a wrong/empty key is rejected 401, and — critically — an unset admin
// key disables the surface entirely (every request 401s even with a header).
func TestAdminAuth(t *testing.T) {
	const configured = "super-secret-admin-key-0123456789"

	tests := []struct {
		name       string
		adminKey   string // configured value
		header     string // presented X-Admin-Key
		setHeader  bool
		wantStatus int
	}{
		{name: "correct key passes", adminKey: configured, header: configured, setHeader: true, wantStatus: fiber.StatusOK},
		{name: "wrong key rejected", adminKey: configured, header: "nope", setHeader: true, wantStatus: fiber.StatusUnauthorized},
		{name: "missing header rejected", adminKey: configured, setHeader: false, wantStatus: fiber.StatusUnauthorized},
		{name: "empty header rejected", adminKey: configured, header: "", setHeader: true, wantStatus: fiber.StatusUnauthorized},
		{name: "unset admin key disables surface", adminKey: "", header: configured, setHeader: true, wantStatus: fiber.StatusUnauthorized},
		{name: "unset admin key rejects empty header too", adminKey: "", setHeader: false, wantStatus: fiber.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
			app.Get("/v1/admin/probe", AdminAuth(tc.adminKey), func(c *fiber.Ctx) error {
				return domain.Success(c, fiber.Map{"ok": true})
			})

			req := httptest.NewRequest(fiber.MethodGet, "/v1/admin/probe", nil)
			if tc.setHeader {
				req.Header.Set("X-Admin-Key", tc.header)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app.Test: %v", err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.wantStatus {
				t.Fatalf("status=%d want %d", resp.StatusCode, tc.wantStatus)
			}
		})
	}
}
