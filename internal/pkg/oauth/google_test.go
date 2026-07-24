package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGoogleExchange verifies the code->identity exchange against a stubbed
// token + userinfo endpoint (no real network).
func TestGoogleExchange(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/token"):
			_ = r.ParseForm()
			if r.FormValue("code") != "good-code" {
				http.Error(w, "bad code", http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{"access_token": "at-123", "token_type": "Bearer"})
		case strings.HasSuffix(r.URL.Path, "/userinfo"):
			if r.Header.Get("Authorization") != "Bearer at-123" {
				http.Error(w, "no token", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(map[string]any{
				"sub": "g-1", "email": "u@x.co", "name": "You", "picture": "http://p/x.png",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	g := NewGoogle("cid", "csecret", "http://app/api/auth/google/callback")
	g.TokenURL = srv.URL + "/token"
	g.UserInfoURL = srv.URL + "/userinfo"

	id, err := g.Exchange(context.Background(), "good-code")
	if err != nil {
		t.Fatalf("Exchange: %v", err)
	}
	if id.Subject != "g-1" || id.Email != "u@x.co" || id.Name != "You" {
		t.Fatalf("identity mismatch: %+v", id)
	}

	if _, err := g.Exchange(context.Background(), "bad-code"); err == nil {
		t.Fatal("expected error for bad code")
	}
}

func TestGoogleAuthCodeURL(t *testing.T) {
	g := NewGoogle("cid", "csecret", "http://app/cb")
	u := g.AuthCodeURL("state-xyz")
	for _, want := range []string{"client_id=cid", "state=state-xyz", "response_type=code", "scope="} {
		if !strings.Contains(u, want) {
			t.Fatalf("auth url missing %q: %s", want, u)
		}
	}
}
