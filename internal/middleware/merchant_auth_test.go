package middleware

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// mergedAuthStubResolver is a MerchantResolver used only by the MerchantAuth
// tests below. Named distinctly from the package's other stubResolver (in
// auth_test.go, which has a different field shape) to avoid a redeclaration.
type mergedAuthStubResolver struct {
	merchant *domain.Merchant
}

func (s mergedAuthStubResolver) ResolveByAPIKeyHash(_ context.Context, hash string) (*domain.Merchant, error) {
	if s.merchant != nil && hash == HashAPIKey("good-key") {
		return s.merchant, nil
	}
	return nil, domain.ErrUnauthorized
}

func mkApp(mgr *session.Manager, r MerchantResolver) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Get("/x", MerchantAuth(mgr, r), func(c *fiber.Ctx) error {
		if _, ok := MerchantIDFromCtx(c); !ok {
			return fiber.NewError(500, "no merchant local")
		}
		return c.SendStatus(200)
	})
	return app
}

func TestMerchantAuthViaSessionCookie(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	mid := uuid.New()
	tok, _ := mgr.Issue(session.Claims{UserID: uuid.New(), MerchantID: mid, Email: "a@b.co"})
	app := mkApp(mgr, mergedAuthStubResolver{})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Cookie", SessionCookieName+"="+tok)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("session cookie -> %d want 200", resp.StatusCode)
	}
}

func TestMerchantAuthViaAPIKey(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	m := &domain.Merchant{ID: uuid.New(), Status: "active"}
	app := mkApp(mgr, mergedAuthStubResolver{merchant: m})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer good-key")
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("api key -> %d want 200", resp.StatusCode)
	}
}

func TestMerchantAuthRejectsNeither(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	app := mkApp(mgr, mergedAuthStubResolver{})
	resp, _ := app.Test(httptest.NewRequest("GET", "/x", nil))
	if resp.StatusCode != 401 {
		t.Fatalf("no creds -> %d want 401", resp.StatusCode)
	}
}
