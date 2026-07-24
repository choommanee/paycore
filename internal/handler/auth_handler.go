package handler

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/oauth"
	"github.com/yourco/payment-gateway/internal/pkg/session"
	"github.com/yourco/payment-gateway/internal/service"
)

// oauthStateCookie holds the anti-CSRF state value between start and callback.
const oauthStateCookie = "pc_oauth_state"

// AuthConfig carries the request-independent settings the handler needs.
type AuthConfig struct {
	Secure            bool          // Set Secure on cookies (prod)
	SessionTTL        time.Duration // pc_session lifetime
	Sandbox           bool          // enables dev-login
	PostLoginRedirect string        // where the callback 302s (e.g. "/")
}

// AuthHandler serves dashboard login: Google OIDC, dev-login (sandbox), logout,
// and the /auth/me self-view.
type AuthHandler struct {
	svc      service.AuthService
	sessions *session.Manager
	provider oauth.OAuthProvider
	cfg      AuthConfig
	log      zerolog.Logger
}

// NewAuthHandler wires the auth handler.
func NewAuthHandler(svc service.AuthService, mgr *session.Manager, provider oauth.OAuthProvider, cfg AuthConfig, log zerolog.Logger) *AuthHandler {
	if cfg.PostLoginRedirect == "" {
		cfg.PostLoginRedirect = "/"
	}
	return &AuthHandler{svc: svc, sessions: mgr, provider: provider, cfg: cfg, log: log}
}

// GoogleStart sets a random state cookie and redirects to Google's consent page.
func (h *AuthHandler) GoogleStart(c *fiber.Ctx) error {
	if h.provider == nil {
		return domain.Error(c, fiber.StatusServiceUnavailable, "OAUTH_DISABLED", "google login is not configured")
	}
	state, err := randToken()
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name: oauthStateCookie, Value: state, Path: "/", MaxAge: 600,
		HTTPOnly: true, Secure: h.cfg.Secure, SameSite: fiber.CookieSameSiteLaxMode,
	})
	return c.Redirect(h.provider.AuthCodeURL(state), fiber.StatusFound)
}

// GoogleCallback verifies state, exchanges the code, provisions/looks up the
// user, sets pc_session, and redirects into the dashboard.
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	if h.provider == nil {
		return domain.Error(c, fiber.StatusServiceUnavailable, "OAUTH_DISABLED", "google login is not configured")
	}
	state := c.Query("state")
	want := c.Cookies(oauthStateCookie)
	if state == "" || want == "" || state != want {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_STATE", "oauth state mismatch")
	}
	code := c.Query("code")
	if code == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_REQUEST", "missing code")
	}
	id, err := h.provider.Exchange(c.UserContext(), code)
	if err != nil {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "oauth exchange failed")
	}
	return h.issueSessionAndRedirect(c, "google", id)
}

// DevLogin issues a session for a fixed sandbox merchant. Mounted only when
// SANDBOX_MODE is on (router guard); it also self-guards here.
func (h *AuthHandler) DevLogin(c *fiber.Ctx) error {
	if !h.cfg.Sandbox {
		return domain.Error(c, fiber.StatusNotFound, "NOT_FOUND", "not found")
	}
	id := domain.OAuthIdentity{Subject: "dev-user", Email: "dev@paycore.local", Name: "Dev Merchant"}
	user, err := h.svc.LoginWithOAuth(c.UserContext(), "dev", id)
	if err != nil {
		return err
	}
	if err := h.setSession(c, user); err != nil {
		return err
	}
	return domain.Success(c, fiber.Map{"ok": true})
}

// Logout clears the session cookie.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	middleware.ClearSessionCookie(c)
	return domain.Success(c, fiber.Map{"ok": true})
}

// Me returns the authenticated dashboard user (session-auth required).
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	uid, ok := middleware.UserIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}
	user, err := h.svc.GetUser(c.UserContext(), uid)
	if err != nil {
		return err
	}
	return domain.Success(c, domain.AuthMe{
		UserID:       user.ID,
		MerchantID:   user.MerchantID,
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		MerchantName: user.Name, // merchant display name; refined when profile join lands
	})
}

func (h *AuthHandler) issueSessionAndRedirect(c *fiber.Ctx, provider string, id domain.OAuthIdentity) error {
	user, err := h.svc.LoginWithOAuth(c.UserContext(), provider, id)
	if err != nil {
		return err
	}
	if err := h.setSession(c, user); err != nil {
		return err
	}
	// clear the one-shot state cookie
	c.Cookie(&fiber.Cookie{Name: oauthStateCookie, Value: "", Path: "/", MaxAge: -1})
	return c.Redirect(h.cfg.PostLoginRedirect, fiber.StatusFound)
}

func (h *AuthHandler) setSession(c *fiber.Ctx, user *domain.MerchantUser) error {
	tok, err := h.sessions.Issue(session.Claims{
		UserID: user.ID, MerchantID: user.MerchantID, Email: user.Email,
	})
	if err != nil {
		return err
	}
	middleware.SetSessionCookie(c, tok, h.cfg.SessionTTL, h.cfg.Secure)
	return nil
}

func randToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
