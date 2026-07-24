package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/pkg/session"
)

func TestSessionAuthAcceptsValidCookie(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	mid := uuid.New()
	uid := uuid.New()
	tok, _ := mgr.Issue(session.Claims{UserID: uid, MerchantID: mid, Email: "a@b.co"})

	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Get("/x", SessionAuth(mgr), func(c *fiber.Ctx) error {
		gotMid, _ := MerchantIDFromCtx(c)
		gotUid, _ := UserIDFromCtx(c)
		if gotMid != mid || gotUid != uid {
			return fiber.NewError(fiber.StatusInternalServerError, "locals not set")
		}
		return c.SendStatus(fiber.StatusOK)
	})

	req := httptest.NewRequest(fiber.MethodGet, "/x", nil)
	req.Header.Set("Cookie", SessionCookieName+"="+tok)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("valid cookie -> %d want 200", resp.StatusCode)
	}
}

func TestSessionAuthRejectsMissingCookie(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Get("/x", SessionAuth(mgr), func(c *fiber.Ctx) error { return c.SendStatus(200) })

	resp, _ := app.Test(httptest.NewRequest(fiber.MethodGet, "/x", nil))
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("no cookie -> %d want 401", resp.StatusCode)
	}
}

func TestSessionAuthRejectsBadCookie(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Get("/x", SessionAuth(mgr), func(c *fiber.Ctx) error { return c.SendStatus(200) })

	req := httptest.NewRequest(fiber.MethodGet, "/x", nil)
	req.Header.Set("Cookie", SessionCookieName+"=not-a-valid-token")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("bad cookie -> %d want 401", resp.StatusCode)
	}
}
