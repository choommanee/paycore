// Package oauth implements the minimal OAuth2/OIDC code exchange the dashboard
// needs, using only the standard library (no golang.org/x/oauth2 dependency).
package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/yourco/payment-gateway/internal/domain"
)

// OAuthProvider is the surface the auth handler depends on.
type OAuthProvider interface {
	AuthCodeURL(state string) string
	Exchange(ctx context.Context, code string) (domain.OAuthIdentity, error)
}

// Google implements OAuthProvider against Google's OIDC endpoints. TokenURL and
// UserInfoURL default to the real endpoints and are overridable in tests.
type Google struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
	UserInfoURL  string
	client       *http.Client
}

// NewGoogle returns a Google provider with default OIDC endpoints.
func NewGoogle(clientID, clientSecret, redirectURL string) *Google {
	return &Google{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:     "https://oauth2.googleapis.com/token",
		UserInfoURL:  "https://openidconnect.googleapis.com/v1/userinfo",
		client:       &http.Client{Timeout: 10 * time.Second},
	}
}

func (g *Google) AuthCodeURL(state string) string {
	q := url.Values{}
	q.Set("client_id", g.ClientID)
	q.Set("redirect_uri", g.RedirectURL)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	return g.AuthURL + "?" + q.Encode()
}

func (g *Google) Exchange(ctx context.Context, code string) (domain.OAuthIdentity, error) {
	var zero domain.OAuthIdentity

	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", g.ClientID)
	form.Set("client_secret", g.ClientSecret)
	form.Set("redirect_uri", g.RedirectURL)
	form.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := g.client.Do(req)
	if err != nil {
		return zero, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("oauth: token exchange failed: %s", resp.Status)
	}
	var tok struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return zero, err
	}
	if tok.AccessToken == "" {
		return zero, fmt.Errorf("oauth: empty access token")
	}

	uReq, err := http.NewRequestWithContext(ctx, http.MethodGet, g.UserInfoURL, nil)
	if err != nil {
		return zero, err
	}
	uReq.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	uResp, err := g.client.Do(uReq)
	if err != nil {
		return zero, err
	}
	defer uResp.Body.Close()
	if uResp.StatusCode != http.StatusOK {
		return zero, fmt.Errorf("oauth: userinfo failed: %s", uResp.Status)
	}
	var info struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(uResp.Body).Decode(&info); err != nil {
		return zero, err
	}
	if info.Sub == "" {
		return zero, fmt.Errorf("oauth: userinfo missing subject")
	}
	return domain.OAuthIdentity{Subject: info.Sub, Email: info.Email, Name: info.Name, Picture: info.Picture}, nil
}
