# Phase 1 — Merchant Human Auth (OAuth + Session) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ให้ร้านค้าล็อกอินเข้า dashboard ด้วย Google (OIDC) หรือ dev-login (sandbox) แล้วได้ session cookie ที่ใช้ยืนยันตัวตนกับ API — เป็นฐานของ dashboard ทุก phase ถัดไป

**Architecture:** Go backend เป็นเจ้าของ identity/OAuth ทั้งหมด: ออก `pc_session` (HMAC-signed token, stdlib) เป็น httpOnly cookie. Next.js app ใหม่ (`web-app/`) proxy ทุก `/api/*` → Go `/v1/*` ผ่าน rewrites (browser เห็น origin เดียว → cookie first-party, ไม่มี CORS). Google callback ชี้กลับ origin ของ Next แล้ว proxy เข้า Go.

**Tech Stack:** Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 · zerolog (backend, **ไม่มี dependency ใหม่** — JWT/HMAC และ OAuth ใช้ stdlib) · Next.js App Router + TypeScript + Tailwind (frontend)

## Global Constraints

- Module path: `github.com/yourco/payment-gateway`
- **ห้ามเพิ่ม Go dependency ใหม่** — session token ใช้ `crypto/hmac`+`crypto/sha256`; OAuth ใช้ `net/http`+`encoding/json`
- ทุก response ใช้ envelope กลาง: `domain.Success` / `domain.Created` / `domain.Error(c, status, code, msg)`
- Query DB ผ่าน sqlc เท่านั้น — เขียน `.sql` ใน `internal/repository/queries/` แล้วรัน `make sqlc` (ห้ามแก้ไฟล์ `*.sql.go` ที่ generate ด้วยมือ)
- ทุก migration ต้องมีคู่ `.up.sql` + `.down.sql` (test `migrations_test.go` บังคับ)
- Secret ไม่ log/persist แบบดิบ; session token เซ็นด้วย `JWT_SECRET` (มี config อยู่แล้ว)
- Service test: fake repo แบบ embed `repository.Querier` (nil) แล้ว override เฉพาะ method ที่ใช้ (ตาม `merchant_service_test.go`)
- Handler test: fake service interface + `app.Test(httptest.NewRequest(...))` + inject locals ผ่าน middleware ปลอม (ตาม `merchant_handler_test.go`)
- รันทั้งชุด: `go build ./... && go test ./...`

---

### Task 1: Config — OAuth + session settings

**Files:**
- Modify: `internal/config/config.go`
- Modify: `.env.example`
- Test: `internal/config/config_auth_test.go` (create)

**Interfaces:**
- Produces: `Config.GoogleClientID`, `Config.GoogleClientSecret`, `Config.OAuthRedirectBase`, `Config.SessionTTLHours int` (default 168), method `Config.SessionTTL() time.Duration`

- [ ] **Step 1: Write the failing test**

Create `internal/config/config_auth_test.go`:
```go
package config

import (
	"testing"
	"time"
)

func TestSessionTTLDefault(t *testing.T) {
	c := &Config{SessionTTLHours: 0}
	if got := c.SessionTTL(); got != 168*time.Hour {
		t.Fatalf("SessionTTL()=%v want 168h (default)", got)
	}
}

func TestSessionTTLFromHours(t *testing.T) {
	c := &Config{SessionTTLHours: 24}
	if got := c.SessionTTL(); got != 24*time.Hour {
		t.Fatalf("SessionTTL()=%v want 24h", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/config/ -run TestSessionTTL -v`
Expected: FAIL — `c.SessionTTL undefined`

- [ ] **Step 3: Write minimal implementation**

In `internal/config/config.go`, add fields to the `Config` struct (after the `JWTSecret`/`IdempotencyTTLs` block):
```go
	// Google OIDC credentials for merchant dashboard login. Empty disables the
	// Google login button; dev-login (sandbox) still works.
	GoogleClientID     string `mapstructure:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `mapstructure:"GOOGLE_CLIENT_SECRET"`

	// OAuthRedirectBase is the public origin of the Next.js app (e.g.
	// https://app.paycore.example). The Google redirect_uri is built as
	// <base>/api/auth/google/callback so the browser stays on one origin.
	OAuthRedirectBase string `mapstructure:"OAUTH_REDIRECT_BASE"`

	// SessionTTLHours is the lifetime of the pc_session cookie. Default 168 (7d).
	SessionTTLHours int `mapstructure:"SESSION_TTL_HOURS"`
```

Add the keys to `envKeys` (append to the `JWT_SECRET`... line group):
```go
	"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "OAUTH_REDIRECT_BASE", "SESSION_TTL_HOURS",
```

Add a default in `Load()` near the other `SetDefault` calls:
```go
	v.SetDefault("SESSION_TTL_HOURS", 168)
```

Add the helper method near `IsProd()` (needs `time` — already imported? config.go does NOT import time; add it):
```go
// SessionTTL returns the pc_session cookie lifetime. Falls back to 7 days when
// SESSION_TTL_HOURS is unset or non-positive.
func (c *Config) SessionTTL() time.Duration {
	h := c.SessionTTLHours
	if h <= 0 {
		h = 168
	}
	return time.Duration(h) * time.Hour
}
```

Add `"time"` to the import block of `config.go`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/config/ -run TestSessionTTL -v`
Expected: PASS

- [ ] **Step 5: Update `.env.example`**

Append to `.env.example`:
```
# --- Merchant dashboard login (Phase 1) ---
# Google OIDC. Leave blank to disable the Google button (dev-login still works in sandbox).
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
# Public origin of the Next.js app; Google redirect_uri = <base>/api/auth/google/callback
OAUTH_REDIRECT_BASE=http://localhost:3000
# pc_session cookie lifetime in hours (default 168 = 7 days)
SESSION_TTL_HOURS=168
```

- [ ] **Step 6: Commit**

```bash
git add internal/config/config.go internal/config/config_auth_test.go .env.example
git commit -m "feat(config): add Google OIDC + session TTL settings"
```

---

### Task 2: `pkg/session` — HMAC-signed session tokens (stdlib)

**Files:**
- Create: `internal/pkg/session/session.go`
- Test: `internal/pkg/session/session_test.go`

**Interfaces:**
- Produces:
  - `session.Claims{ UserID uuid.UUID; MerchantID uuid.UUID; Email string }`
  - `session.NewManager(secret string, ttl time.Duration) *session.Manager`
  - `(*Manager) Issue(c Claims) (string, error)`
  - `(*Manager) Verify(token string) (*Claims, error)`
  - sentinel errors `session.ErrInvalidToken`, `session.ErrExpired`

- [ ] **Step 1: Write the failing test**

Create `internal/pkg/session/session_test.go`:
```go
package session

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func newClaims() Claims {
	return Claims{UserID: uuid.New(), MerchantID: uuid.New(), Email: "a@b.co"}
}

func TestIssueVerifyRoundTrip(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", time.Hour)
	c := newClaims()
	tok, err := m.Issue(c)
	if err != nil {
		t.Fatalf("Issue: %v", err)
	}
	got, err := m.Verify(tok)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got.UserID != c.UserID || got.MerchantID != c.MerchantID || got.Email != c.Email {
		t.Fatalf("claims mismatch: got %+v want %+v", got, c)
	}
}

func TestVerifyRejectsTamper(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", time.Hour)
	tok, _ := m.Issue(newClaims())
	// flip the last character of the payload
	bad := tok[:len(tok)-1] + "X"
	if _, err := m.Verify(bad); err == nil {
		t.Fatal("Verify accepted a tampered token")
	}
}

func TestVerifyRejectsWrongSecret(t *testing.T) {
	tok, _ := NewManager("secret-one-secret-one-secret-one-xx", time.Hour).Issue(newClaims())
	if _, err := NewManager("secret-two-secret-two-secret-two-xx", time.Hour).Verify(tok); err == nil {
		t.Fatal("Verify accepted a token signed with a different secret")
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	m := NewManager("test-secret-at-least-32-bytes-long!!", -time.Minute) // already expired
	tok, _ := m.Issue(newClaims())
	if _, err := m.Verify(tok); err != ErrExpired {
		t.Fatalf("Verify err=%v want ErrExpired", err)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/pkg/session/ -v`
Expected: FAIL — package/`NewManager` undefined

- [ ] **Step 3: Write minimal implementation**

Create `internal/pkg/session/session.go`:
```go
// Package session issues and verifies compact, HMAC-SHA256-signed session
// tokens for the merchant dashboard. Format: base64url(payloadJSON).base64url(mac)
// where mac = HMAC-SHA256(secret, base64url(payloadJSON)). No external deps.
package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidToken is returned for a malformed token or a signature mismatch.
	ErrInvalidToken = errors.New("session: invalid token")
	// ErrExpired is returned when the token's exp is in the past.
	ErrExpired = errors.New("session: token expired")
)

// Claims is the authenticated identity carried by a session token.
type Claims struct {
	UserID     uuid.UUID `json:"uid"`
	MerchantID uuid.UUID `json:"mid"`
	Email      string    `json:"email"`
}

// payload is the wire form: claims plus issued/expiry unix seconds.
type payload struct {
	Claims
	IssuedAt int64 `json:"iat"`
	Expires  int64 `json:"exp"`
}

// Manager signs and verifies tokens with a fixed secret and TTL.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager returns a Manager. ttl is the token lifetime; a non-positive ttl
// yields tokens that are already expired (used in tests).
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

var enc = base64.RawURLEncoding

// Issue returns a signed token for the claims, expiring after the manager TTL.
func (m *Manager) Issue(c Claims) (string, error) {
	now := time.Now().UTC()
	body, err := json.Marshal(payload{
		Claims:   c,
		IssuedAt: now.Unix(),
		Expires:  now.Add(m.ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	b64 := enc.EncodeToString(body)
	return b64 + "." + enc.EncodeToString(m.mac(b64)), nil
}

// Verify checks the signature and expiry and returns the claims.
func (m *Manager) Verify(token string) (*Claims, error) {
	b64, sig, ok := strings.Cut(token, ".")
	if !ok {
		return nil, ErrInvalidToken
	}
	gotMAC, err := enc.DecodeString(sig)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if !hmac.Equal(gotMAC, m.mac(b64)) {
		return nil, ErrInvalidToken
	}
	body, err := enc.DecodeString(b64)
	if err != nil {
		return nil, ErrInvalidToken
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, ErrInvalidToken
	}
	if time.Now().UTC().Unix() >= p.Expires {
		return nil, ErrExpired
	}
	c := p.Claims
	return &c, nil
}

func (m *Manager) mac(b64 string) []byte {
	h := hmac.New(sha256.New, m.secret)
	h.Write([]byte(b64))
	return h.Sum(nil)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/pkg/session/ -v`
Expected: PASS (all four tests)

- [ ] **Step 5: Commit**

```bash
git add internal/pkg/session/
git commit -m "feat(session): HMAC-signed session token manager (stdlib)"
```

---

### Task 3: Migration + sqlc — `merchant_users`

**Files:**
- Create: `migrations/000008_merchant_users.up.sql`
- Create: `migrations/000008_merchant_users.down.sql`
- Create: `internal/repository/queries/merchant_user.sql`
- Generate: `internal/repository/merchant_user.sql.go`, `internal/repository/models.go`, `internal/repository/querier.go` (via `make sqlc`)
- Test: `migrations/migrations_auth_test.go`

**Interfaces:**
- Produces (sqlc-generated): `repository.MerchantUser` model; querier methods `GetMerchantUserByOAuth`, `GetMerchantUserByID`, `CreateMerchantUser`, `TouchMerchantUserLogin` with their `*Params` structs.

- [ ] **Step 1: Write the failing test**

Create `migrations/migrations_auth_test.go`:
```go
package migrations_test

import (
	"strings"
	"testing"
)

// TestMerchantUsersMigrationReversible asserts the merchant_users up migration
// creates the table + unique oauth index and the down migration drops the table.
func TestMerchantUsersMigrationReversible(t *testing.T) {
	up := mustRead(t, "000008_merchant_users.up.sql")
	down := mustRead(t, "000008_merchant_users.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE MERCHANT_USERS") {
		t.Fatalf("up does not create merchant_users: %s", up)
	}
	if !strings.Contains(upU, "UNIQUE") || !strings.Contains(strings.ToLower(up), "oauth_subject") {
		t.Fatalf("up must enforce a unique (oauth_provider, oauth_subject): %s", up)
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "merchant_users") {
		t.Fatalf("down must drop merchant_users: %s", down)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./migrations/ -run TestMerchantUsersMigrationReversible -v`
Expected: FAIL — `read 000008_merchant_users.up.sql: ... no such file`

- [ ] **Step 3: Write the migrations**

Create `migrations/000008_merchant_users.up.sql`:
```sql
-- Human identities that log into the merchant dashboard. Distinct from the
-- API-key credential on merchants: one merchant may later have several users.
-- Login is via OAuth (Google) or 'dev' (sandbox only); we store the provider
-- subject, never a password.
CREATE TABLE merchant_users (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id    UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    email          TEXT NOT NULL,
    name           TEXT NOT NULL DEFAULT '',
    avatar_url     TEXT NOT NULL DEFAULT '',
    oauth_provider TEXT NOT NULL,            -- 'google' | 'dev'
    oauth_subject  TEXT NOT NULL,            -- provider 'sub'
    role           TEXT NOT NULL DEFAULT 'owner',
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A provider identity maps to exactly one user (idempotent login/signup).
CREATE UNIQUE INDEX merchant_users_oauth_idx
    ON merchant_users (oauth_provider, oauth_subject);

-- Fast lookup of all users for a merchant.
CREATE INDEX merchant_users_merchant_idx ON merchant_users (merchant_id);
```

Create `migrations/000008_merchant_users.down.sql`:
```sql
DROP INDEX IF EXISTS merchant_users_merchant_idx;
DROP INDEX IF EXISTS merchant_users_oauth_idx;
DROP TABLE IF EXISTS merchant_users;
```

- [ ] **Step 4: Run the migration test to verify it passes**

Run: `go test ./migrations/ -v`
Expected: PASS (new test + existing pairing test)

- [ ] **Step 5: Write the sqlc queries**

Create `internal/repository/queries/merchant_user.sql`:
```sql
-- internal/repository/queries/merchant_user.sql
-- Dashboard human identities. Login resolves a user by (provider, subject); a
-- first-time identity is created together with its merchant by the service layer.

-- name: GetMerchantUserByOAuth :one
SELECT * FROM merchant_users
WHERE oauth_provider = $1 AND oauth_subject = $2;

-- name: GetMerchantUserByID :one
SELECT * FROM merchant_users WHERE id = $1;

-- name: CreateMerchantUser :one
INSERT INTO merchant_users (
    id, merchant_id, email, name, avatar_url, oauth_provider, oauth_subject
) VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: TouchMerchantUserLogin :exec
UPDATE merchant_users SET last_login_at = NOW(), updated_at = NOW()
WHERE id = $1;
```

- [ ] **Step 6: Generate + build**

Run: `make sqlc && go build ./...`
Expected: no errors; `internal/repository/merchant_user.sql.go` created, `repository.MerchantUser` present in `models.go`.

- [ ] **Step 7: Commit**

```bash
git add migrations/000008_merchant_users.up.sql migrations/000008_merchant_users.down.sql \
        migrations/migrations_auth_test.go internal/repository/queries/merchant_user.sql \
        internal/repository/
git commit -m "feat(db): merchant_users table + sqlc queries"
```

---

### Task 4: Domain types for auth

**Files:**
- Create: `internal/domain/auth.go`
- Test: covered by Task 5/6 (pure structs — no standalone test)

**Interfaces:**
- Produces:
  - `domain.OAuthIdentity{ Subject, Email, Name, Picture string }`
  - `domain.MerchantUser{ ID, MerchantID uuid.UUID; Email, Name, AvatarURL, Provider, Role string; CreatedAt, UpdatedAt time.Time }`
  - `domain.AuthMe{ UserID, MerchantID uuid.UUID; Email, Name, AvatarURL, MerchantName string }`

- [ ] **Step 1: Write the type file**

Create `internal/domain/auth.go`:
```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// OAuthIdentity is the normalized profile returned by an OAuth/OIDC provider
// after a successful code exchange.
type OAuthIdentity struct {
	Subject string `json:"subject"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// MerchantUser is a human who logs into the dashboard, scoped to one merchant.
type MerchantUser struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	AvatarURL  string    `json:"avatar_url,omitempty"`
	Provider   string    `json:"provider"`
	Role       string    `json:"role"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AuthMe is the GET /v1/auth/me response: the logged-in user plus enough
// merchant context for the dashboard header.
type AuthMe struct {
	UserID       uuid.UUID `json:"user_id"`
	MerchantID   uuid.UUID `json:"merchant_id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	AvatarURL    string    `json:"avatar_url,omitempty"`
	MerchantName string    `json:"merchant_name"`
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add internal/domain/auth.go
git commit -m "feat(domain): auth identity + session view types"
```

---

### Task 5: `AuthService` — login-or-provision + user lookup

**Files:**
- Create: `internal/service/auth_service.go`
- Test: `internal/service/auth_service_test.go`
- Reuse helpers already in `internal/service` package: `generateAPIKey()`, `toPgUUID()`, `pgUUIDToUUID()`, `strPtr()`, `ptrStr()`

**Interfaces:**
- Consumes: `repository.Querier` (methods `GetMerchantUserByOAuth`, `CreateMerchant`, `CreateMerchantUser`, `TouchMerchantUserLogin`, `GetMerchantUserByID`, `GetMerchant`), `middleware.HashAPIKey`
- Produces:
  - `service.AuthService interface { LoginWithOAuth(ctx, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error); GetUser(ctx, id uuid.UUID) (*domain.MerchantUser, error) }`
  - `service.NewAuthService(repo repository.Querier, log zerolog.Logger) AuthService`

- [ ] **Step 1: Write the failing test**

Create `internal/service/auth_service_test.go`:
```go
package service

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// fakeAuthRepo implements the auth slice of repository.Querier.
type fakeAuthRepo struct {
	repository.Querier
	mu         sync.Mutex
	usersByOA  map[string]repository.MerchantUser // "provider|subject" -> user
	usersByID  map[uuid.UUID]repository.MerchantUser
	merchants  map[uuid.UUID]bool
	createCnt  int
	touchedIDs []uuid.UUID
}

func newFakeAuthRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		usersByOA: map[string]repository.MerchantUser{},
		usersByID: map[uuid.UUID]repository.MerchantUser{},
		merchants: map[uuid.UUID]bool{},
	}
}

func (f *fakeAuthRepo) GetMerchantUserByOAuth(_ context.Context, arg repository.GetMerchantUserByOAuthParams) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.usersByOA[arg.OauthProvider+"|"+arg.OauthSubject]
	if !ok {
		return repository.MerchantUser{}, pgx.ErrNoRows
	}
	return u, nil
}

func (f *fakeAuthRepo) CreateMerchant(_ context.Context, arg repository.CreateMerchantParams) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.merchants[uuid.UUID(arg.ID.Bytes)] = true
	return repository.Merchant{ID: arg.ID, Name: arg.Name, Status: "active", SettlementCurrency: arg.SettlementCurrency}, nil
}

func (f *fakeAuthRepo) CreateMerchantUser(_ context.Context, arg repository.CreateMerchantUserParams) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.createCnt++
	u := repository.MerchantUser{
		ID:            arg.ID,
		MerchantID:    arg.MerchantID,
		Email:         arg.Email,
		Name:          arg.Name,
		AvatarUrl:     arg.AvatarUrl,
		OauthProvider: arg.OauthProvider,
		OauthSubject:  arg.OauthSubject,
		Role:          "owner",
	}
	f.usersByOA[arg.OauthProvider+"|"+arg.OauthSubject] = u
	f.usersByID[uuid.UUID(arg.ID.Bytes)] = u
	return u, nil
}

func (f *fakeAuthRepo) TouchMerchantUserLogin(_ context.Context, id pgtype.UUID) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.touchedIDs = append(f.touchedIDs, uuid.UUID(id.Bytes))
	return nil
}

func (f *fakeAuthRepo) GetMerchantUserByID(_ context.Context, id pgtype.UUID) (repository.MerchantUser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	u, ok := f.usersByID[uuid.UUID(id.Bytes)]
	if !ok {
		return repository.MerchantUser{}, pgx.ErrNoRows
	}
	return u, nil
}

func TestLoginWithOAuthProvisionsThenReuses(t *testing.T) {
	repo := newFakeAuthRepo()
	svc := NewAuthService(repo, zerolog.Nop())
	id := domain.OAuthIdentity{Subject: "google-123", Email: "shop@x.co", Name: "Shop"}

	first, err := svc.LoginWithOAuth(context.Background(), "google", id)
	if err != nil {
		t.Fatalf("first login: %v", err)
	}
	if first.Email != "shop@x.co" || first.MerchantID == uuid.Nil {
		t.Fatalf("unexpected user: %+v", first)
	}
	if repo.createCnt != 1 || len(repo.merchants) != 1 {
		t.Fatalf("first login should provision 1 merchant+user, got users=%d merchants=%d", repo.createCnt, len(repo.merchants))
	}

	second, err := svc.LoginWithOAuth(context.Background(), "google", id)
	if err != nil {
		t.Fatalf("second login: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("second login created a new user: %v != %v", second.ID, first.ID)
	}
	if repo.createCnt != 1 {
		t.Fatalf("second login must NOT re-provision, createCnt=%d", repo.createCnt)
	}
	if len(repo.touchedIDs) != 2 {
		t.Fatalf("both logins should touch last_login_at, got %d", len(repo.touchedIDs))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/service/ -run TestLoginWithOAuth -v`
Expected: FAIL — `NewAuthService undefined`

- [ ] **Step 3: Write minimal implementation**

Create `internal/service/auth_service.go`:
```go
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/repository"
)

// AuthService resolves dashboard human identities. LoginWithOAuth is
// login-or-provision: a first-time provider identity creates a merchant (with a
// generated sandbox API key) and its owner user; a returning identity is reused.
type AuthService interface {
	LoginWithOAuth(ctx context.Context, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error)
	GetUser(ctx context.Context, id uuid.UUID) (*domain.MerchantUser, error)
}

type authService struct {
	repo repository.Querier
	log  zerolog.Logger
}

// NewAuthService wires the auth service.
func NewAuthService(repo repository.Querier, log zerolog.Logger) AuthService {
	return &authService{repo: repo, log: log.With().Str("service", "auth").Logger()}
}

func (s *authService) LoginWithOAuth(ctx context.Context, provider string, id domain.OAuthIdentity) (*domain.MerchantUser, error) {
	if s.repo == nil {
		return nil, errors.New("auth repository not configured")
	}
	// Returning identity: reuse.
	row, err := s.repo.GetMerchantUserByOAuth(ctx, repository.GetMerchantUserByOAuthParams{
		OauthProvider: provider,
		OauthSubject:  id.Subject,
	})
	if err == nil {
		_ = s.repo.TouchMerchantUserLogin(ctx, row.ID)
		return toDomainMerchantUser(row), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	// First-time identity: provision merchant + owner user.
	rawKey, err := generateAPIKey()
	if err != nil {
		return nil, err
	}
	merchantID := uuid.New()
	name := id.Name
	if name == "" {
		name = id.Email
	}
	if _, err := s.repo.CreateMerchant(ctx, repository.CreateMerchantParams{
		ID:                 toPgUUID(merchantID),
		Name:               name,
		Mcc:                nil,
		SettlementCurrency: "THB",
		ApiKeyHash:         middleware.HashAPIKey(rawKey),
	}); err != nil {
		return nil, err
	}
	created, err := s.repo.CreateMerchantUser(ctx, repository.CreateMerchantUserParams{
		ID:            toPgUUID(uuid.New()),
		MerchantID:    toPgUUID(merchantID),
		Email:         id.Email,
		Name:          name,
		AvatarUrl:     id.Picture,
		OauthProvider: provider,
		OauthSubject:  id.Subject,
	})
	if err != nil {
		return nil, err
	}
	_ = s.repo.TouchMerchantUserLogin(ctx, created.ID)
	return toDomainMerchantUser(created), nil
}

func (s *authService) GetUser(ctx context.Context, id uuid.UUID) (*domain.MerchantUser, error) {
	if s.repo == nil {
		return nil, domain.ErrMerchantNotFound
	}
	row, err := s.repo.GetMerchantUserByID(ctx, toPgUUID(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMerchantNotFound
		}
		return nil, err
	}
	return toDomainMerchantUser(row), nil
}

func toDomainMerchantUser(r repository.MerchantUser) *domain.MerchantUser {
	u := &domain.MerchantUser{
		ID:            pgUUIDToUUID(r.ID),
		MerchantID:    pgUUIDToUUID(r.MerchantID),
		Email:         r.Email,
		Name:          r.Name,
		AvatarURL:     r.AvatarUrl,
		Provider:      r.OauthProvider,
		Role:          r.Role,
	}
	u.CreatedAt = r.CreatedAt.Time
	u.UpdatedAt = r.UpdatedAt.Time
	return u
}
```

> Note: if `make sqlc` emits nullable columns as pointers (`emit_pointers_for_null_types: true`), `AvatarUrl`/`Name` are `string` here because the migration declares them `NOT NULL DEFAULT ''`. If a field comes back `*string`, adjust the conversion with `ptrStr(...)`. Verify against the generated `repository.MerchantUser` struct after Task 3.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/service/ -run TestLoginWithOAuth -v`
Expected: PASS

- [ ] **Step 5: Run the full service package to catch regressions**

Run: `go test ./internal/service/...`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add internal/service/auth_service.go internal/service/auth_service_test.go
git commit -m "feat(service): OAuth login-or-provision + user lookup"
```

---

### Task 6: `SessionAuth` middleware + cookie helpers

**Files:**
- Create: `internal/middleware/session_auth.go`
- Test: `internal/middleware/session_auth_test.go`

**Interfaces:**
- Consumes: `session.Manager` (Task 2)
- Produces:
  - `const middleware.SessionCookieName = "pc_session"`
  - `const middleware.LocalUserID = "user_id"`
  - `middleware.SessionAuth(mgr *session.Manager) fiber.Handler` — verifies the `pc_session` cookie, sets `LocalMerchantID` (uuid.UUID) and `LocalUserID` (uuid.UUID) locals, else 401 `UNAUTHORIZED`
  - `middleware.UserIDFromCtx(c *fiber.Ctx) (uuid.UUID, bool)`
  - `middleware.SetSessionCookie(c *fiber.Ctx, token string, ttl time.Duration, secure bool)` and `middleware.ClearSessionCookie(c *fiber.Ctx)`

- [ ] **Step 1: Write the failing test**

Create `internal/middleware/session_auth_test.go`:
```go
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
	req.AddCookie(&fiber.Cookie{Name: SessionCookieName, Value: tok}.toStd()) // see helper below
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
```

> The `.toStd()` helper on the first test is pseudocode. Replace that cookie line with the standard-library form, which needs no helper:
> ```go
> req.Header.Set("Cookie", SessionCookieName+"="+tok)
> ```
> Use `req.Header.Set("Cookie", ...)` in all three tests for consistency.

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/ -run TestSessionAuth -v`
Expected: FAIL — `SessionAuth undefined`

- [ ] **Step 3: Write minimal implementation**

Create `internal/middleware/session_auth.go`:
```go
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// SessionCookieName is the dashboard session cookie.
const SessionCookieName = "pc_session"

// LocalUserID is the ctx local holding the authenticated dashboard user id.
const LocalUserID = "user_id"

// SessionAuth authenticates a dashboard user from the pc_session cookie. On
// success it sets LocalMerchantID and LocalUserID (both uuid.UUID) so downstream
// handlers scope queries exactly as the API-key path does; on any failure it
// returns 401 without leaking why.
func SessionAuth(mgr *session.Manager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		raw := c.Cookies(SessionCookieName)
		if raw == "" {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		}
		claims, err := mgr.Verify(raw)
		if err != nil {
			return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
		}
		c.Locals(LocalMerchantID, claims.MerchantID)
		c.Locals(LocalUserID, claims.UserID)
		return c.Next()
	}
}

// UserIDFromCtx returns the authenticated dashboard user id set by SessionAuth.
func UserIDFromCtx(c *fiber.Ctx) (uuid.UUID, bool) {
	id, ok := c.Locals(LocalUserID).(uuid.UUID)
	return id, ok
}

// SetSessionCookie writes the pc_session cookie. secure should be true in prod.
func SetSessionCookie(c *fiber.Ctx, token string, ttl time.Duration, secure bool) {
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(ttl.Seconds()),
		HTTPOnly: true,
		Secure:   secure,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}

// ClearSessionCookie expires the pc_session cookie (logout).
func ClearSessionCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteLaxMode,
	})
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/ -run TestSessionAuth -v`
Expected: PASS (all three)

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/session_auth.go internal/middleware/session_auth_test.go
git commit -m "feat(middleware): pc_session cookie auth + helpers"
```

---

### Task 7: OAuth provider (Google, stdlib) + `AuthHandler` + routes/wiring

**Files:**
- Create: `internal/pkg/oauth/google.go`
- Create: `internal/handler/auth_handler.go`
- Test: `internal/pkg/oauth/google_test.go`
- Test: `internal/handler/auth_handler_test.go`
- Modify: `internal/handler/handler.go` (add `Auth *AuthHandler`; extend `New(...)` signature)
- Modify: `internal/router/router.go` (`Setup(...)` gains `sessionAuth fiber.Handler`; mount auth routes)
- Modify: `cmd/server/main.go` (build session manager, google provider, auth service+handler, session middleware; update `handler.New` + `router.Setup` calls)

**Interfaces:**
- Consumes: `service.AuthService`, `session.Manager`, `config.Config`, `middleware.SetSessionCookie/ClearSessionCookie/SessionAuth`
- Produces:
  - `oauth.OAuthProvider interface { AuthCodeURL(state string) string; Exchange(ctx context.Context, code string) (domain.OAuthIdentity, error) }`
  - `oauth.NewGoogle(clientID, clientSecret, redirectURL string) *oauth.Google` (implements OAuthProvider; `Google.TokenURL` and `Google.UserInfoURL` fields default to Google endpoints, overridable in tests)
  - `handler.NewAuthHandler(svc service.AuthService, mgr *session.Manager, provider oauth.OAuthProvider, cfg handler.AuthConfig, log zerolog.Logger) *handler.AuthHandler`
  - `handler.AuthConfig{ Secure bool; SessionTTL time.Duration; Sandbox bool; PostLoginRedirect string }`
  - handler methods `GoogleStart`, `GoogleCallback`, `DevLogin`, `Logout`, `Me`

- [ ] **Step 1: Write the failing test for the Google provider**

Create `internal/pkg/oauth/google_test.go`:
```go
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
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/pkg/oauth/ -v`
Expected: FAIL — `NewGoogle undefined`

- [ ] **Step 3: Implement the Google provider**

Create `internal/pkg/oauth/google.go`:
```go
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
```

- [ ] **Step 4: Run provider test to verify it passes**

Run: `go test ./internal/pkg/oauth/ -v`
Expected: PASS

- [ ] **Step 5: Write the failing handler test**

Create `internal/handler/auth_handler_test.go`:
```go
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
	"github.com/yourco/payment-gateway/internal/pkg/oauth"
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

func (stubProvider) AuthCodeURL(state string) string { return "https://accounts.example/auth?state=" + state }
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
```

- [ ] **Step 6: Run handler test to verify it fails**

Run: `go test ./internal/handler/ -run 'TestGoogleCallback|TestDevLogin' -v`
Expected: FAIL — `NewAuthHandler undefined`

- [ ] **Step 7: Implement the auth handler**

Create `internal/handler/auth_handler.go`:
```go
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
```

- [ ] **Step 8: Run handler test to verify it passes**

Run: `go test ./internal/handler/ -run 'TestGoogleCallback|TestDevLogin' -v`
Expected: PASS

- [ ] **Step 9: Wire the handler into the aggregate + router + main**

In `internal/handler/handler.go`: add `Auth *AuthHandler` to the `Handlers` struct. Do NOT change the `New(...)` signature (it is broad); instead attach Auth via a `WithAuth` builder mirroring `WithSandbox`:
```go
// WithAuth attaches the dashboard auth handler (Google OIDC + session).
func (h *Handlers) WithAuth(a *AuthHandler) *Handlers {
	h.Auth = a
	return h
}
```

In `internal/router/router.go`: change `Setup` to accept a `sessionAuth fiber.Handler` argument (add it after `auth`), and mount the auth routes inside `v1` (before the payments group). Guard dev-login by the existing `sandbox` bool:
```go
	// Dashboard human auth (Google OIDC + session cookie). Public except /me.
	if h.Auth != nil {
		v1.Get("/auth/google/start", h.Auth.GoogleStart)
		v1.Get("/auth/google/callback", h.Auth.GoogleCallback)
		v1.Post("/auth/logout", h.Auth.Logout)
		v1.Get("/auth/me", sessionAuth, h.Auth.Me)
		if sandbox {
			v1.Post("/auth/dev-login", h.Auth.DevLogin)
		}
	}
```
Update the `Setup` doc comment to describe `sessionAuth`.

In `cmd/server/main.go`, after `merchantSvc`/`disputeSvc` are built and before `router.Setup`:
```go
	sessions := session.NewManager(cfg.JWTSecret, cfg.SessionTTL())
	authSvc := service.NewAuthService(repo, logger)
	var oauthProvider oauth.OAuthProvider
	if cfg.GoogleClientID != "" {
		oauthProvider = oauth.NewGoogle(
			cfg.GoogleClientID, cfg.GoogleClientSecret,
			strings.TrimRight(cfg.OAuthRedirectBase, "/")+"/api/auth/google/callback",
		)
	}
	authHandler := handler.NewAuthHandler(authSvc, sessions, oauthProvider, handler.AuthConfig{
		Secure:            cfg.IsProd(),
		SessionTTL:        cfg.SessionTTL(),
		Sandbox:           cfg.SandboxMode,
		PostLoginRedirect: "/",
	}, logger)
	sessionAuth := middleware.SessionAuth(sessions)
```
Attach the handler to the aggregate where `h` is built (`h := handler.New(...)` → `h = h.WithAuth(authHandler)`), and update the `router.Setup(...)` call to pass `sessionAuth` in its new position:
```go
	router.Setup(app, h, auth, sessionAuth, adminAuth, rateLimit, signupLimit, publicMetrics, cfg.WebDir, cfg.SandboxMode)
```
Add imports to `main.go`: `"strings"`, `".../internal/pkg/oauth"`, `".../internal/pkg/session"` (middleware/service already imported).

- [ ] **Step 10: Build + full test suite**

Run: `go build ./... && go test ./...`
Expected: PASS across all packages

- [ ] **Step 11: Manual smoke (sandbox dev-login, no Google needed)**

Run (two shells):
```bash
SANDBOX_MODE=true JWT_SECRET=dev-secret-dev-secret-dev-secret-xx make run
# other shell:
curl -i -X POST http://localhost:8080/v1/auth/dev-login -c /tmp/pc.cookies
curl -s http://localhost:8080/v1/auth/me -b /tmp/pc.cookies
```
Expected: dev-login returns 200 + a `pc_session` cookie; `/auth/me` returns the dev user JSON (not 401).

- [ ] **Step 12: Commit**

```bash
git add internal/pkg/oauth/ internal/handler/auth_handler.go internal/handler/auth_handler_test.go \
        internal/handler/handler.go internal/router/router.go cmd/server/main.go
git commit -m "feat(auth): Google OIDC + dev-login + session routes wired"
```

---

### Task 8: Next.js app scaffold (`web-app/`) with API proxy

**Files:**
- Create: `web-app/package.json`, `web-app/next.config.js`, `web-app/tsconfig.json`, `web-app/postcss.config.js`, `web-app/tailwind.config.ts`
- Create: `web-app/app/layout.tsx`, `web-app/app/globals.css`
- Create: `web-app/.env.example`, `web-app/.gitignore`, `web-app/README.md`

**Interfaces:**
- Produces: a Next.js app that proxies `GET/POST /api/*` → `${BACKEND_URL}/v1/*`, so the browser only ever calls the Next origin.

- [ ] **Step 1: Create the package + config files**

`web-app/package.json`:
```json
{
  "name": "paycore-web",
  "private": true,
  "scripts": {
    "dev": "next dev -p 3000",
    "build": "next build",
    "start": "next start -p 3000",
    "lint": "next lint"
  },
  "dependencies": {
    "next": "14.2.5",
    "react": "18.3.1",
    "react-dom": "18.3.1"
  },
  "devDependencies": {
    "@types/node": "20.14.10",
    "@types/react": "18.3.3",
    "autoprefixer": "10.4.19",
    "postcss": "8.4.39",
    "tailwindcss": "3.4.6",
    "typescript": "5.5.3"
  }
}
```

`web-app/next.config.js` — proxy `/api/*` to the Go backend's `/v1/*`:
```js
/** @type {import('next').NextConfig} */
const backend = process.env.BACKEND_URL || "http://localhost:8080";
const nextConfig = {
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${backend}/v1/:path*` }];
  },
};
module.exports = nextConfig;
```

`web-app/tsconfig.json`:
```json
{
  "compilerOptions": {
    "target": "ES2021",
    "lib": ["dom", "dom.iterable", "esnext"],
    "allowJs": false,
    "skipLibCheck": true,
    "strict": true,
    "noEmit": true,
    "esModuleInterop": true,
    "module": "esnext",
    "moduleResolution": "bundler",
    "resolveJsonModule": true,
    "isolatedModules": true,
    "jsx": "preserve",
    "incremental": true,
    "plugins": [{ "name": "next" }],
    "paths": { "@/*": ["./*"] }
  },
  "include": ["next-env.d.ts", "**/*.ts", "**/*.tsx", ".next/types/**/*.ts"],
  "exclude": ["node_modules"]
}
```

`web-app/postcss.config.js`:
```js
module.exports = { plugins: { tailwindcss: {}, autoprefixer: {} } };
```

`web-app/tailwind.config.ts` — seed PayCore brand tokens (refine against `web/assets/paycore.css` in a later task):
```ts
import type { Config } from "tailwindcss";

const config: Config = {
  content: ["./app/**/*.{ts,tsx}", "./components/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        paycore: {
          bg: "#0b0f1a",
          surface: "#141a2b",
          primary: "#3b82f6",
          primaryHover: "#2563eb",
          text: "#e5e9f0",
          muted: "#94a3b8",
        },
      },
      borderRadius: { xl2: "1rem" },
    },
  },
  plugins: [],
};
export default config;
```

- [ ] **Step 2: Create the app shell**

`web-app/app/globals.css`:
```css
@tailwind base;
@tailwind components;
@tailwind utilities;

body { @apply bg-paycore-bg text-paycore-text antialiased; }
```

`web-app/app/layout.tsx`:
```tsx
import "./globals.css";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "PayCore",
  description: "PayCore merchant dashboard & checkout",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="th">
      <body>{children}</body>
    </html>
  );
}
```

- [ ] **Step 3: Create env + ignore + README**

`web-app/.env.example`:
```
# Internal URL of the Go API (Railway private networking or localhost in dev)
BACKEND_URL=http://localhost:8080
```

`web-app/.gitignore`:
```
node_modules
.next
next-env.d.ts
.env
.env.local
```

`web-app/README.md`:
```md
# PayCore Web (Next.js)

Dashboard + hosted checkout. Proxies `/api/*` -> Go API `/v1/*` (see `next.config.js`),
so the browser only ever calls this origin (no CORS, first-party session cookie).

## Dev
    cp .env.example .env    # set BACKEND_URL if not localhost:8080
    npm install
    npm run dev             # http://localhost:3000

The Go API must run with SANDBOX_MODE=true to expose dev-login.
```

- [ ] **Step 4: Install + verify the scaffold builds**

Run:
```bash
cd web-app && npm install && npm run build
```
Expected: `npm run build` succeeds (compiles the empty shell; no route errors).

- [ ] **Step 5: Commit**

```bash
git add web-app/package.json web-app/next.config.js web-app/tsconfig.json \
        web-app/postcss.config.js web-app/tailwind.config.ts web-app/app/layout.tsx \
        web-app/app/globals.css web-app/.env.example web-app/.gitignore web-app/README.md
git commit -m "feat(web): Next.js scaffold with /api proxy to Go backend"
```

> Note: `package-lock.json` is generated by `npm install`. Commit it too if the repo convention keeps lockfiles (`git add web-app/package-lock.json`).

---

### Task 9: Login page + authenticated dashboard shell

**Files:**
- Create: `web-app/app/login/page.tsx`
- Create: `web-app/app/page.tsx` (dashboard home; redirects to `/login` when unauthenticated)
- Create: `web-app/lib/api.ts` (server-side fetch helper that forwards cookies)

**Interfaces:**
- Consumes: backend `GET /v1/auth/me`, `GET /v1/auth/google/start`, `POST /v1/auth/dev-login` (all via the `/api/*` proxy)

- [ ] **Step 1: Server-side API helper that forwards the session cookie**

`web-app/lib/api.ts`:
```ts
import { cookies, headers } from "next/headers";

// serverGet calls the backend through the same-origin /api proxy, forwarding the
// incoming request cookies so the pc_session cookie reaches the Go API.
export async function serverGet(path: string): Promise<Response> {
  const h = headers();
  const host = h.get("host");
  const proto = h.get("x-forwarded-proto") ?? "http";
  const cookieHeader = cookies().toString();
  return fetch(`${proto}://${host}/api${path}`, {
    headers: { cookie: cookieHeader },
    cache: "no-store",
  });
}
```

- [ ] **Step 2: Login page**

`web-app/app/login/page.tsx`:
```tsx
"use client";

import { useState } from "react";

export default function LoginPage() {
  const [busy, setBusy] = useState(false);

  async function devLogin() {
    setBusy(true);
    const res = await fetch("/api/auth/dev-login", { method: "POST" });
    if (res.ok) window.location.href = "/";
    else setBusy(false);
  }

  return (
    <main className="min-h-screen flex items-center justify-center px-4">
      <div className="w-full max-w-sm rounded-xl2 bg-paycore-surface p-8 shadow-xl">
        <h1 className="text-2xl font-semibold mb-1">PayCore</h1>
        <p className="text-paycore-muted mb-6 text-sm">เข้าสู่ระบบร้านค้า</p>

        <a
          href="/api/auth/google/start"
          className="block w-full text-center rounded-lg bg-white text-gray-900 font-medium py-2.5 mb-3 hover:bg-gray-100"
        >
          เข้าสู่ระบบด้วย Google
        </a>

        <button
          onClick={devLogin}
          disabled={busy}
          className="block w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium py-2.5 disabled:opacity-60"
        >
          {busy ? "กำลังเข้าสู่ระบบ…" : "Dev login (sandbox)"}
        </button>

        <p className="text-paycore-muted text-xs mt-4">
          ปุ่ม Google ใช้ได้เมื่อ backend ตั้งค่า GOOGLE_CLIENT_ID แล้ว
        </p>
      </div>
    </main>
  );
}
```

- [ ] **Step 3: Dashboard home that requires a session**

`web-app/app/page.tsx`:
```tsx
import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";

type AuthMe = {
  user_id: string;
  merchant_id: string;
  email: string;
  name: string;
  merchant_name: string;
};

export default async function DashboardHome() {
  const res = await serverGet("/auth/me");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`auth/me failed: ${res.status}`);
  const env = await res.json();
  const me: AuthMe = env.data;

  return (
    <main className="min-h-screen p-8">
      <header className="flex items-center justify-between mb-8">
        <h1 className="text-xl font-semibold">PayCore Dashboard</h1>
        <form action="/api/auth/logout" method="post">
          <button className="text-sm text-paycore-muted hover:text-paycore-text">ออกจากระบบ</button>
        </form>
      </header>
      <section className="rounded-xl2 bg-paycore-surface p-6">
        <p className="text-paycore-muted text-sm">เข้าสู่ระบบในชื่อ</p>
        <p className="text-lg font-medium">{me.name || me.email}</p>
        <p className="text-paycore-muted text-xs mt-1">merchant: {me.merchant_id}</p>
      </section>
    </main>
  );
}
```

> The logout `<form method="post">` posts to the proxied `/api/auth/logout`; the Go handler clears the cookie and returns JSON. A later phase can swap this for a fetch + client redirect; for Phase 1 the cookie clears and the next `/` load redirects to `/login`.

- [ ] **Step 4: Verify the build**

Run:
```bash
cd web-app && npm run build
```
Expected: build succeeds; `/login` and `/` compile.

- [ ] **Step 5: Manual end-to-end smoke (sandbox)**

```bash
# shell 1: backend
SANDBOX_MODE=true JWT_SECRET=dev-secret-dev-secret-dev-secret-xx make run
# shell 2: frontend
cd web-app && BACKEND_URL=http://localhost:8080 npm run dev
```
In a browser: open `http://localhost:3000` → redirected to `/login` → click **Dev login** → lands on the dashboard showing the dev user → **ออกจากระบบ** → reload `/` → back to `/login`.

- [ ] **Step 6: Commit**

```bash
git add web-app/app/login/page.tsx web-app/app/page.tsx web-app/lib/api.ts
git commit -m "feat(web): login page + session-guarded dashboard shell"
```

---

## Self-Review

**Spec coverage (Phase 1 slice of the design spec §3.1, §5.1, §7 login, §8):**
- merchant_users table → Task 3 ✓
- Google OIDC start/callback → Task 7 ✓
- dev-login (sandbox) → Task 7 ✓
- pc_session cookie (httpOnly/Secure/SameSite=Lax, JWT_SECRET) → Tasks 2, 6, 7 ✓
- sessionAuth middleware → Task 6 ✓
- /auth/me → Task 7 ✓
- OAuth state CSRF guard → Task 7 (state cookie compare) ✓
- Next.js `/api/*` proxy, no CORS, first-party cookie → Task 8 ✓
- `/login` + guarded dashboard shell → Task 9 ✓
- Config (Google/redirect/TTL) → Task 1 ✓
- PKCE: **deferred** — state CSRF is implemented; PKCE is a hardening follow-up noted for Phase 5 (confidential client with secret is acceptable for MVP).

**Placeholder scan:** none — every step has full code. The one pseudocode line (`.toStd()`) in Task 6's test is explicitly replaced by the note directly under it.

**Type consistency:** `session.Claims`/`NewManager`/`Issue`/`Verify` used identically in Tasks 2, 6, 7. `AuthService.LoginWithOAuth/GetUser` defined in Task 5, consumed in Task 7. `domain.OAuthIdentity`/`MerchantUser`/`AuthMe` defined in Task 4, used in Tasks 5, 7. `middleware.SessionCookieName`/`LocalUserID`/`SetSessionCookie`/`SessionAuth` defined in Task 6, used in Task 7. sqlc `*Params` names (`GetMerchantUserByOAuthParams`, `CreateMerchantUserParams`) assumed from the query names in Task 3 — **verify against generated code after `make sqlc`** (Task 5 step 3 note covers the nullable-field caveat).

**Known verification points (call out during execution):**
1. After `make sqlc`, confirm generated field names on `repository.MerchantUser` (`AvatarUrl` vs `AvatarURL`, `string` vs `*string`) and adjust `toDomainMerchantUser` if needed.
2. `router.Setup` is called from exactly one place (`cmd/server/main.go:215`) — the signature change is localized.
3. `handler.New` signature is left unchanged; Auth is attached via `WithAuth`, so no other call sites break.
