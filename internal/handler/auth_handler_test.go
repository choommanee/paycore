package handler

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

type fakeAuthSvc struct{ user domain.MerchantUser }

func (f *fakeAuthSvc) LoginWithOAuth(_ context.Context, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error) {
	f.user = domain.MerchantUser{ID: uuid.New(), MerchantID: uuid.New(), Email: id.Email, Name: id.Name, Provider: provider, Role: "owner"}
	return &f.user, nil
}
func (f *fakeAuthSvc) GetUser(_ context.Context, _ uuid.UUID) (*domain.MerchantUser, error) {
	return &f.user, nil
}

type stubProvider struct{}

func (stubProvider) AuthCodeURL(state string) string {
	return "https://accounts.example/auth?state=" + state
}
func (stubProvider) Exchange(_ context.Context, code string) (domain.OAuthIdentity, error) {
	return domain.OAuthIdentity{Subject: "sub-" + code, Email: "u@x.co", Name: "U"}, nil
}

func newAuthHandler() (*AuthHandler, *session.Manager) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	h := NewAuthHandler(&fakeAuthSvc{}, mgr, stubProvider{}, AuthConfig{
		Secure: false, SessionTTL: time.Hour, Sandbox: true, PostLoginRedirect: "/",
	}, zerolog.Nop())
	return h, mgr
}

// A successful Google callback sets a valid pc_session cookie and redirects.
func TestGoogleCallbackSetsSessionCookie(t *testing.T) {
	h, mgr := newAuthHandler()
	app := fiber.New()
	app.Get("/v1/auth/google/start", h.GoogleStart)
	app.Get("/v1/auth/google/callback", h.GoogleCallback)

	// start -> capture state cookie
	startResp, _ := app.Test(httptest.NewRequest(fiber.MethodGet, "/v1/auth/google/start", nil))
	var stateCookie string
	for _, c := range startResp.Cookies() {
		if c.Name == oauthStateCookie {
			stateCookie = c.Value
		}
	}
	if stateCookie == "" {
		t.Fatal("start did not set an oauth state cookie")
	}

	// callback with matching state
	req := httptest.NewRequest(fiber.MethodGet, "/v1/auth/google/callback?code=abc&state="+stateCookie, nil)
	req.Header.Set("Cookie", oauthStateCookie+"="+stateCookie)
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusFound {
		t.Fatalf("callback -> %d want 302", resp.StatusCode)
	}
	var sess string
	for _, c := range resp.Cookies() {
		if c.Name == middleware.SessionCookieName {
			sess = c.Value
		}
	}
	if sess == "" {
		t.Fatal("callback did not set pc_session")
	}
	if _, err := mgr.Verify(sess); err != nil {
		t.Fatalf("pc_session not verifiable: %v", err)
	}
}

// The callback rejects a state mismatch (CSRF guard) with 400.
func TestGoogleCallbackRejectsStateMismatch(t *testing.T) {
	h, _ := newAuthHandler()
	app := fiber.New()
	app.Get("/v1/auth/google/callback", h.GoogleCallback)

	req := httptest.NewRequest(fiber.MethodGet, "/v1/auth/google/callback?code=abc&state=attacker", nil)
	req.Header.Set("Cookie", oauthStateCookie+"=real-state")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("state mismatch -> %d want 400", resp.StatusCode)
	}
}

// dev-login issues a session directly (sandbox only).
func TestDevLoginIssuesSession(t *testing.T) {
	h, mgr := newAuthHandler()
	app := fiber.New()
	app.Post("/v1/auth/dev-login", h.DevLogin)

	resp, _ := app.Test(httptest.NewRequest(fiber.MethodPost, "/v1/auth/dev-login", nil))
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("dev-login -> %d want 200", resp.StatusCode)
	}
	var sess string
	for _, c := range resp.Cookies() {
		if c.Name == middleware.SessionCookieName {
			sess = c.Value
		}
	}
	if sess == "" {
		t.Fatal("dev-login did not set pc_session")
	}
	if _, err := mgr.Verify(sess); err != nil {
		t.Fatalf("session invalid: %v", err)
	}
}
