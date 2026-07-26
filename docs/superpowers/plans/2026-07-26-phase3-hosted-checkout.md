# Phase 3 — Hosted Checkout (card + PromptPay) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** turn a Phase 2 payment link into a working public hosted-checkout page at `/pay/[publicId]` that takes real payments end-to-end via **card (sandbox)** and **PromptPay QR** — creating a short-lived, token-authenticated `checkout_sessions` row that bridges to the existing `PaymentService` and `QRService`.

**Architecture:** A new `checkout_sessions` table (migration `000010`) holds one payment attempt, authenticated by an opaque token given to the browser and stored as a SHA-256 hash (exactly like `api_key_hash`). A new `CheckoutService` resolves a session by token hash and drives it: PromptPay → mints a `qr_payments` row via `QRService` and returns the EMVCo payload for the page to render + poll; card → tokenizes the PAN with the `crypto.Vault` (sandbox only) and charges through `PaymentService`. Three **unauthenticated** public routes under `/v1/checkout/*` are the only new API surface — the session token in the URL path is the credential. The Next.js `/pay/[publicId]` page is public (no cookie), creates the session client-side, shows merchant + amount + a card|promptpay selector, renders the PromptPay QR with the existing vanilla `qrcode.min.js` asset (no npm dep), and polls to completion.

**Tech Stack:** Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 · shopspring/decimal · Next.js 14 App Router + TypeScript + Tailwind. **No new Go or npm dependency.**

## Global Constraints

- Module `github.com/yourco/payment-gateway`. **NO new Go dependency; NO new npm dependency** (if a QR renderer seems needed, reuse the existing `web/assets/qrcode.min.js` vanilla lib — see Task 8).
- Money is stored as **integer minor units** (สตางค์, `amount_minor BIGINT`) in `checkout_sessions`, but **converted to `decimal.Decimal` major units** with `money.FromMinor(minor, currency)` before calling `PaymentService.Create` / `QRService.Create` (both take `decimal.Decimal`). Never pass minor units to those services.
- **Session token = opaque random, stored hashed.** Generate a random token, hand the raw token to the browser exactly once (on create), and persist only `session_token_hash = middleware.HashAPIKey(token)` (hex SHA-256). Look sessions up by that hash. Never store or log the raw token.
- **Session scoping (anti-IDOR):** a session token grants access to **its own session only**. The server derives merchant context from the resolved session row's `merchant_id` — never from the request body or params. Sessions are looked up solely by token hash.
- **Card PAN path is SANDBOX-gated.** Receiving a raw PAN on our server is sandbox-only (like the sandbox payer sim). When `SANDBOX_MODE=false`, the card pay path returns a clear "method not available" error and never accepts card data. PromptPay works in all modes. Document this in code comments.
- Public checkout endpoints are **UNAUTHENTICATED** (no `pc_session` cookie, no API key). Session creation (`POST /v1/checkout/sessions`) is **rate-limited per client IP**.
- Every response uses the central envelope `domain.Success` / `domain.Created` / `domain.Error`.
- DB access via sqlc only (`.sql` files in `internal/repository/queries/`, run `make sqlc`; never hand-edit `*.sql.go`).
- Every migration has a paired up/down (`migrations_test.go` enforces it). Next migration = `000010`.
- TDD: write the failing test first, watch it fail, implement minimally, watch it pass, commit. Complete code in every step. Frequent commits.
- Service tests: fake repo embeds `repository.Querier` (nil) and overrides only the methods used (pattern from `payment_link_service_test.go`, `merchant_service_test.go`). Narrow consumer interfaces (`Charger`, `QRIssuer`, `Tokenizer`) are faked directly.
- Handler tests: fake the service interface + `app.Test`; public routes need **no** auth-local injection (pattern from `payment_link_handler_test.go`, minus `withMerchant`).
- Reuse existing helpers (do NOT redefine): `toPgUUID`, `pgUUIDToUUID`, `pgTimestamptz`, `strPtr`, `ptrStr` (service pkg); `middleware.HashAPIKey`, `middleware.clientIP` (unexported, same pkg); `paginate`/`validationErrorResponse` (handler pkg); `money.FromMinor` (`internal/pkg/money`).
- Run set: `go build ./... && go test ./...`; frontend `cd web-app && npm run build`.

---

### Task 1: Migration + sqlc — `checkout_sessions`

**Files:**
- Create: `migrations/000010_checkout_sessions.up.sql`, `migrations/000010_checkout_sessions.down.sql`
- Create: `internal/repository/queries/checkout_session.sql`
- Generate: `internal/repository/checkout_session.sql.go`, `models.go`, `querier.go` (via `make sqlc`)
- Test: `migrations/migrations_checkout_test.go`

**Interfaces:**
- Produces (sqlc-generated): `repository.CheckoutSession` model; querier methods `CreateCheckoutSession`, `GetCheckoutSessionByTokenHash`, `UpdateCheckoutSession` + their `*Params` structs.

- [ ] **Step 1: Write the failing migration test**

Create `migrations/migrations_checkout_test.go`:
```go
package migrations_test

import (
	"strings"
	"testing"
)

func TestCheckoutSessionsMigrationReversible(t *testing.T) {
	up := mustRead(t, "000010_checkout_sessions.up.sql")
	down := mustRead(t, "000010_checkout_sessions.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE CHECKOUT_SESSIONS") {
		t.Fatalf("up does not create checkout_sessions: %s", up)
	}
	// The session token is stored hashed and must be uniquely addressable.
	if !strings.Contains(strings.ToLower(up), "session_token_hash") || !strings.Contains(upU, "UNIQUE") {
		t.Fatalf("up must have a unique session_token_hash: %s", up)
	}
	if !strings.Contains(upU, "AMOUNT_MINOR") || !strings.Contains(upU, "CHECK") {
		t.Fatalf("up must CHECK amount_minor > 0: %s", up)
	}
	// Nullable bridges to payments / qr_payments / payment_links.
	for _, fk := range []string{"payment_link_id", "payment_id", "qr_payment_id"} {
		if !strings.Contains(strings.ToLower(up), fk) {
			t.Fatalf("up must have column %s: %s", fk, up)
		}
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "checkout_sessions") {
		t.Fatalf("down must drop checkout_sessions: %s", down)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./migrations/ -run TestCheckoutSessionsMigrationReversible -v`
Expected: FAIL — file not found (open `000010_checkout_sessions.up.sql`: no such file).

- [ ] **Step 3: Write the migrations**

Create `migrations/000010_checkout_sessions.up.sql`:
```sql
-- A single hosted-checkout payment attempt. Created from a payment_link (or
-- directly via API); the opaque session token is given to the browser and only
-- its SHA-256 hash is stored here (like api_key_hash). The token is the sole
-- credential for the public /v1/checkout routes; merchant context is derived
-- from this row, never from the request. Short-lived (expires_at ~30 min).
CREATE TABLE checkout_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    payment_link_id     UUID REFERENCES payment_links(id) ON DELETE SET NULL,
    session_token_hash  TEXT NOT NULL,
    amount_minor        BIGINT NOT NULL CHECK (amount_minor > 0),
    currency            TEXT NOT NULL DEFAULT 'THB',
    status              TEXT NOT NULL DEFAULT 'open',  -- open | processing | requires_action | paid | failed | expired
    selected_method     TEXT NOT NULL DEFAULT '',
    payment_id          UUID REFERENCES payments(id) ON DELETE SET NULL,
    qr_payment_id       UUID REFERENCES qr_payments(id) ON DELETE SET NULL,
    customer_email      TEXT NOT NULL DEFAULT '',
    return_url          TEXT NOT NULL DEFAULT '',
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The token hash is the lookup key for every public checkout request.
CREATE UNIQUE INDEX checkout_sessions_token_hash_idx ON checkout_sessions (session_token_hash);
-- A merchant's recent sessions (ops / debugging), newest first.
CREATE INDEX checkout_sessions_merchant_idx ON checkout_sessions (merchant_id, created_at DESC);
```

Create `migrations/000010_checkout_sessions.down.sql`:
```sql
DROP INDEX IF EXISTS checkout_sessions_merchant_idx;
DROP INDEX IF EXISTS checkout_sessions_token_hash_idx;
DROP TABLE IF EXISTS checkout_sessions;
```

- [ ] **Step 4: Run the migration test to verify it passes**

Run: `go test ./migrations/ -v`
Expected: PASS (new test + existing `TestEveryMigrationHasUpAndDown` pairing test).

- [ ] **Step 5: Write the sqlc queries**

Create `internal/repository/queries/checkout_session.sql`:
```sql
-- internal/repository/queries/checkout_session.sql
-- Hosted-checkout sessions. A session is ALWAYS resolved by its token hash (the
-- token is the credential); the merchant scope comes from the resolved row, not
-- from the caller. Updates are by primary key (already resolved from the hash).

-- name: CreateCheckoutSession :one
INSERT INTO checkout_sessions (
    id, merchant_id, payment_link_id, session_token_hash, amount_minor, currency,
    status, selected_method, payment_id, qr_payment_id, customer_email, return_url, expires_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
RETURNING *;

-- name: GetCheckoutSessionByTokenHash :one
SELECT * FROM checkout_sessions WHERE session_token_hash = $1;

-- name: UpdateCheckoutSession :one
UPDATE checkout_sessions
SET status = $2, selected_method = $3, payment_id = $4, qr_payment_id = $5,
    customer_email = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;
```

- [ ] **Step 6: Generate + build**

Run: `make sqlc && go build ./...`
Expected: no errors. Open `internal/repository/models.go` and confirm the `CheckoutSession` struct exists. **Record the exact generated field names for later tasks** — sqlc lowercases/camelizes: `payment_link_id` → `PaymentLinkID`, `qr_payment_id` → `QrPaymentID`, `payment_id` → `PaymentID`, `return_url` → `ReturnUrl`, `session_token_hash` → `SessionTokenHash`, `selected_method` → `SelectedMethod`, `customer_email` → `CustomerEmail`. Nullable FKs (`PaymentLinkID`, `PaymentID`, `QrPaymentID`) are `pgtype.UUID`; `ExpiresAt`/`CreatedAt`/`UpdatedAt` are `pgtype.Timestamptz`; the `NOT NULL DEFAULT ''` text columns are plain `string`. If any generated name differs, use the generated name in Tasks 3–5.

- [ ] **Step 7: Commit**

```bash
git add migrations/000010_checkout_sessions.up.sql migrations/000010_checkout_sessions.down.sql \
        migrations/migrations_checkout_test.go internal/repository/queries/checkout_session.sql \
        internal/repository/
git commit -m "feat(db): checkout_sessions table + sqlc queries"
```

---

### Task 2: Domain types, method registry helper, errors + HTTP mapping

**Files:**
- Create: `internal/domain/checkout.go`
- Test: `internal/domain/checkout_test.go`
- Modify: `internal/domain/errors.go` (3 new sentinels)
- Modify: `internal/middleware/middleware.go` (3 new error→HTTP cases)

**Interfaces:**
- Produces:
  - `domain.CheckoutStatus` + constants `CheckoutOpen`, `CheckoutProcessing`, `CheckoutRequiresAction`, `CheckoutPaid`, `CheckoutFailed`, `CheckoutExpired`
  - `domain.CheckoutSupportedMethods []string` (= `{"card","promptpay"}`) and `domain.DisplayMethods(allowed []string) []string`
  - `domain.CheckoutSessionRequest{ Link string }`
  - `domain.CardInput{ Number string; ExpMonth, ExpYear int; CVV, HolderName string }`
  - `domain.CheckoutPayRequest{ Method string; Card *CardInput; CustomerEmail string }`
  - `domain.CheckoutSessionView{ ... }` (public projection; see code)
  - `domain.ErrCheckoutSessionNotFound`, `domain.ErrCheckoutSessionExpired`, `domain.ErrCheckoutMethodUnavailable`

- [ ] **Step 1: Write the failing test**

Create `internal/domain/checkout_test.go`:
```go
package domain

import (
	"reflect"
	"testing"
)

func TestDisplayMethodsEmptyMeansAllSupported(t *testing.T) {
	got := DisplayMethods(nil)
	if !reflect.DeepEqual(got, []string{"card", "promptpay"}) {
		t.Fatalf("empty allowed = %v want [card promptpay]", got)
	}
}

func TestDisplayMethodsIntersectsAndKeepsSupportedOrder(t *testing.T) {
	// Link allows a Phase-4 wallet + promptpay + card, in a different order.
	got := DisplayMethods([]string{"truemoney", "promptpay", "card"})
	// Only supported methods survive, in CheckoutSupportedMethods order.
	if !reflect.DeepEqual(got, []string{"card", "promptpay"}) {
		t.Fatalf("got %v want [card promptpay]", got)
	}
}

func TestDisplayMethodsCardOnly(t *testing.T) {
	got := DisplayMethods([]string{"card"})
	if !reflect.DeepEqual(got, []string{"card"}) {
		t.Fatalf("got %v want [card]", got)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/ -run TestDisplayMethods -v`
Expected: FAIL — `DisplayMethods` undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/domain/checkout.go`:
```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// CheckoutStatus is the lifecycle of a hosted-checkout session.
type CheckoutStatus string

const (
	CheckoutOpen           CheckoutStatus = "open"             // created, no method chosen
	CheckoutProcessing     CheckoutStatus = "processing"      // payment in flight
	CheckoutRequiresAction CheckoutStatus = "requires_action" // awaiting 3DS redirect or QR scan
	CheckoutPaid           CheckoutStatus = "paid"
	CheckoutFailed         CheckoutStatus = "failed"
	CheckoutExpired        CheckoutStatus = "expired"
)

// CheckoutSupportedMethods is the set of methods the hosted checkout can actually
// process in Phase 3: card (sandbox-gated, raw PAN) and promptpay (all modes).
// Wallet / redirect methods (Beam parity) arrive in Phase 4.
var CheckoutSupportedMethods = []string{"card", "promptpay"}

// DisplayMethods intersects a link's allowed methods with the methods the
// checkout can process, preserving CheckoutSupportedMethods order. An empty
// allowed list means "every method the merchant enabled", so it returns all
// supported methods.
func DisplayMethods(allowed []string) []string {
	if len(allowed) == 0 {
		out := make([]string, len(CheckoutSupportedMethods))
		copy(out, CheckoutSupportedMethods)
		return out
	}
	set := make(map[string]bool, len(allowed))
	for _, m := range allowed {
		set[m] = true
	}
	out := make([]string, 0, len(CheckoutSupportedMethods))
	for _, m := range CheckoutSupportedMethods {
		if set[m] {
			out = append(out, m)
		}
	}
	return out
}

// CheckoutSessionRequest creates a session from a payment link's public id.
type CheckoutSessionRequest struct {
	Link string `json:"link" validate:"required"`
}

// CardInput is a raw card entry. SANDBOX ONLY — a real deployment tokenizes card
// data in the browser via hosted fields and never sends a PAN to this server.
type CardInput struct {
	Number     string `json:"number" validate:"required,min=12,max=19"`
	ExpMonth   int    `json:"exp_month" validate:"required,min=1,max=12"`
	ExpYear    int    `json:"exp_year" validate:"required,min=2024,max=2100"`
	CVV        string `json:"cvv" validate:"required,min=3,max=4"`
	HolderName string `json:"holder_name" validate:"omitempty,max=200"`
}

// CheckoutPayRequest selects a method and carries its data. Card is required only
// when Method == "card" (enforced in the service, not the validator).
type CheckoutPayRequest struct {
	Method        string     `json:"method" validate:"required,oneof=card promptpay"`
	Card          *CardInput `json:"card" validate:"omitempty"`
	CustomerEmail string     `json:"customer_email" validate:"omitempty,email"`
}

// CheckoutSessionView is the PUBLIC, secret-free projection returned to the
// hosted checkout page. It NEVER carries session_token_hash, payment internals,
// or card data. Token is populated ONLY on session creation (the browser holds it
// thereafter); it is omitted on subsequent reads. QRPayload / NextActionURL are
// set only when the chosen method produced them.
type CheckoutSessionView struct {
	Token          string    `json:"session_token,omitempty"`
	ID             uuid.UUID `json:"id"`
	Status         string    `json:"status"`
	AmountMinor    int64     `json:"amount_minor"`
	Currency       string    `json:"currency"`
	MerchantName   string    `json:"merchant_name"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	ImageURL       string    `json:"image_url,omitempty"`
	AllowedMethods []string  `json:"allowed_methods"`
	SelectedMethod string    `json:"selected_method,omitempty"`
	QRPayload      string    `json:"qr_payload,omitempty"`      // PromptPay EMVCo string
	NextActionURL  string    `json:"next_action_url,omitempty"` // card 3DS redirect
	ReturnURL      string    `json:"return_url,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
	Sandbox        bool      `json:"sandbox"` // card form is sandbox-only; drives the UI label
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/ -run TestDisplayMethods -v`
Expected: PASS (all three).

- [ ] **Step 5: Add the three domain errors**

In `internal/domain/errors.go`, inside the existing `var ( ... )` block (after `ErrPaymentLinkNotFound` on the last line), add:
```go
	ErrCheckoutSessionNotFound   = errors.New("checkout session not found")
	ErrCheckoutSessionExpired    = errors.New("checkout session expired")
	ErrCheckoutMethodUnavailable = errors.New("payment method not available")
```

- [ ] **Step 6: Map the three errors to HTTP**

In `internal/middleware/middleware.go`, in the `ErrorHandler` `switch`, add three cases next to the `ErrPaymentLinkNotFound` case (around line 195):
```go
		case errors.Is(err, domain.ErrCheckoutSessionNotFound):
			return domain.Error(c, fiber.StatusNotFound, "CHECKOUT_SESSION_NOT_FOUND", err.Error())
		case errors.Is(err, domain.ErrCheckoutSessionExpired):
			return domain.Error(c, fiber.StatusGone, "CHECKOUT_SESSION_EXPIRED", err.Error())
		case errors.Is(err, domain.ErrCheckoutMethodUnavailable):
			return domain.Error(c, fiber.StatusUnprocessableEntity, "CHECKOUT_METHOD_UNAVAILABLE", err.Error())
```
Without these, a not-found/expired/unsupported-method outcome would fall through to a generic 500.

- [ ] **Step 7: Build + regression**

Run: `go build ./... && go test ./internal/domain/... ./internal/middleware/...`
Expected: PASS.

- [ ] **Step 8: Commit**

```bash
git add internal/domain/checkout.go internal/domain/checkout_test.go \
        internal/domain/errors.go internal/middleware/middleware.go
git commit -m "feat(domain): checkout session types, method registry, errors"
```

---

### Task 3: `CheckoutService.CreateFromLink` (session token + display)

**Files:**
- Create: `internal/service/checkout_service.go`
- Test: `internal/service/checkout_service_test.go`
- Reuse existing package helpers: `toPgUUID`, `pgUUIDToUUID`, `pgTimestamptz`

**Interfaces:**
- Consumes: `repository.Querier` (`GetPaymentLinkByPublicID`, `GetPaymentLink`, `GetMerchant`, `CreateCheckoutSession`); `middleware.HashAPIKey`; `money.FromMinor`.
- Produces:
  - `service.Charger interface { Create(ctx, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) }` (satisfied by `PaymentService`)
  - `service.QRIssuer interface { Create(ctx, req domain.CreateQRRequest) (*domain.QRPayment, error); Get(ctx, merchantID, id uuid.UUID) (*domain.QRPayment, error) }` (satisfied by `QRService`)
  - `service.Tokenizer interface { Tokenize(ctx, pan string) (token, last4 string, err error) }` (satisfied by `*crypto.Vault`)
  - `service.CheckoutService interface { CreateFromLink(ctx, publicID string) (*domain.CheckoutSessionView, error); Get(ctx, token string) (*domain.CheckoutSessionView, error); Pay(ctx, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) }`
  - `service.NewCheckoutService(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, sandbox bool, log zerolog.Logger) CheckoutService`

> This task implements the type, constructor, `CreateFromLink`, and the private `buildView` helper. `Get` and `Pay` are added in Tasks 4–5; to keep the file compiling, stub them to `return nil, domain.ErrNotImplemented` now and replace them there. The test file's fakes (`fakeCheckoutRepo`, `fakeCharger`, `fakeQR`, `fakeVault`, `newCheckoutSvc`) are created here and REUSED by Tasks 4–5 — do not redefine them.

- [ ] **Step 1: Write the failing test**

Create `internal/service/checkout_service_test.go`:
```go
package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/repository"
)

// ---- fakes (shared by Tasks 3–5) -------------------------------------------

type fakeCheckoutRepo struct {
	repository.Querier
	mu             sync.Mutex
	linksByPublic  map[string]repository.PaymentLink
	linksByID      map[uuid.UUID]repository.PaymentLink
	sessByHash     map[string]repository.CheckoutSession
	sessByID       map[uuid.UUID]repository.CheckoutSession
	merchantNames  map[uuid.UUID]string
	linkStatusSets []string // records UpdatePaymentLinkStatus calls
}

func newFakeCheckoutRepo() *fakeCheckoutRepo {
	return &fakeCheckoutRepo{
		linksByPublic: map[string]repository.PaymentLink{},
		linksByID:     map[uuid.UUID]repository.PaymentLink{},
		sessByHash:    map[string]repository.CheckoutSession{},
		sessByID:      map[uuid.UUID]repository.CheckoutSession{},
		merchantNames: map[uuid.UUID]string{},
	}
}

func (f *fakeCheckoutRepo) putLink(l repository.PaymentLink) {
	f.linksByPublic[l.PublicID] = l
	f.linksByID[uuid.UUID(l.ID.Bytes)] = l
}

func (f *fakeCheckoutRepo) GetPaymentLinkByPublicID(_ context.Context, publicID string) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByPublic[publicID]
	if !ok {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return l, nil
}

func (f *fakeCheckoutRepo) GetPaymentLink(_ context.Context, a repository.GetPaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	l, ok := f.linksByID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(l.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return l, nil
}

func (f *fakeCheckoutRepo) GetMerchant(_ context.Context, id pgtype.UUID) (repository.Merchant, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	name, ok := f.merchantNames[uuid.UUID(id.Bytes)]
	if !ok {
		return repository.Merchant{}, pgx.ErrNoRows
	}
	return repository.Merchant{ID: id, Name: name, Status: "active"}, nil
}

func (f *fakeCheckoutRepo) CreateCheckoutSession(_ context.Context, a repository.CreateCheckoutSessionParams) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs := repository.CheckoutSession{
		ID: a.ID, MerchantID: a.MerchantID, PaymentLinkID: a.PaymentLinkID,
		SessionTokenHash: a.SessionTokenHash, AmountMinor: a.AmountMinor, Currency: a.Currency,
		Status: a.Status, SelectedMethod: a.SelectedMethod, PaymentID: a.PaymentID,
		QrPaymentID: a.QrPaymentID, CustomerEmail: a.CustomerEmail, ReturnUrl: a.ReturnUrl,
		ExpiresAt: a.ExpiresAt,
	}
	f.sessByHash[a.SessionTokenHash] = cs
	f.sessByID[uuid.UUID(a.ID.Bytes)] = cs
	return cs, nil
}

func (f *fakeCheckoutRepo) GetCheckoutSessionByTokenHash(_ context.Context, hash string) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs, ok := f.sessByHash[hash]
	if !ok {
		return repository.CheckoutSession{}, pgx.ErrNoRows
	}
	return cs, nil
}

func (f *fakeCheckoutRepo) UpdateCheckoutSession(_ context.Context, a repository.UpdateCheckoutSessionParams) (repository.CheckoutSession, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cs := f.sessByID[uuid.UUID(a.ID.Bytes)]
	cs.Status = a.Status
	cs.SelectedMethod = a.SelectedMethod
	cs.PaymentID = a.PaymentID
	cs.QrPaymentID = a.QrPaymentID
	cs.CustomerEmail = a.CustomerEmail
	f.sessByID[uuid.UUID(a.ID.Bytes)] = cs
	f.sessByHash[cs.SessionTokenHash] = cs
	return cs, nil
}

func (f *fakeCheckoutRepo) UpdatePaymentLinkStatus(_ context.Context, a repository.UpdatePaymentLinkStatusParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.linkStatusSets = append(f.linkStatusSets, a.Status)
	l := f.linksByID[uuid.UUID(a.ID.Bytes)]
	l.Status = a.Status
	f.linksByID[uuid.UUID(a.ID.Bytes)] = l
	return l, nil
}

type fakeCharger struct {
	createFn func(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error)
}

func (f *fakeCharger) Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
	if f.createFn != nil {
		return f.createFn(ctx, idemKey, req)
	}
	return &domain.Payment{ID: uuid.New(), Status: domain.StatusCaptured}, nil
}

type fakeQR struct {
	createFn func(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error)
	getFn    func(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error)
}

func (f *fakeQR) Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
	if f.createFn != nil {
		return f.createFn(ctx, req)
	}
	return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "EMVCO-PAYLOAD"}, nil
}

func (f *fakeQR) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error) {
	if f.getFn != nil {
		return f.getFn(ctx, merchantID, id)
	}
	return &domain.QRPayment{ID: id, Status: domain.QRAwaitingPayment, QRPayload: "EMVCO-PAYLOAD"}, nil
}

type fakeVault struct {
	tokenizeFn func(ctx context.Context, pan string) (string, string, error)
}

func (f *fakeVault) Tokenize(ctx context.Context, pan string) (string, string, error) {
	if f.tokenizeFn != nil {
		return f.tokenizeFn(ctx, pan)
	}
	if len(pan) < 4 {
		return "", "", pgx.ErrNoRows
	}
	return "tokv1.fake", pan[len(pan)-4:], nil
}

// newCheckoutSvc builds the service with fakes. sandbox defaults to true so the
// card path is exercised; individual tests can rebuild with sandbox=false.
func newCheckoutSvc(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, sandbox bool) CheckoutService {
	return NewCheckoutService(repo, charger, qr, vault, sandbox, zerolog.Nop())
}

// mkLink seeds an active link and returns it.
func mkLink(f *fakeCheckoutRepo, merchant uuid.UUID, publicID string, amountMinor int64, methods []string) repository.PaymentLink {
	l := repository.PaymentLink{
		ID:             toPgUUID(uuid.New()),
		MerchantID:     toPgUUID(merchant),
		PublicID:       publicID,
		Title:          "Coffee",
		Description:    "Latte",
		AmountMinor:    amountMinor,
		Currency:       "THB",
		AllowedMethods: methods,
		LinkType:       "single_use",
		Status:         "active",
	}
	f.putLink(l)
	f.merchantNames[merchant] = "Acme Cafe"
	return l
}

// ---- Task 3 tests ----------------------------------------------------------

func TestCreateFromLinkReturnsTokenAndDisplay(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)

	view, err := svc.CreateFromLink(context.Background(), "pl_abc")
	if err != nil {
		t.Fatalf("CreateFromLink: %v", err)
	}
	if !strings.HasPrefix(view.Token, "cs_") {
		t.Fatalf("token = %q want cs_ prefix", view.Token)
	}
	if view.Status != string(domain.CheckoutOpen) {
		t.Fatalf("status = %q want open", view.Status)
	}
	if view.AmountMinor != 5000 || view.Currency != "THB" {
		t.Fatalf("amount/currency = %d/%s", view.AmountMinor, view.Currency)
	}
	if view.MerchantName != "Acme Cafe" || view.Title != "Coffee" {
		t.Fatalf("display = %q / %q", view.MerchantName, view.Title)
	}
	if len(view.AllowedMethods) != 2 {
		t.Fatalf("methods = %v want [card promptpay]", view.AllowedMethods)
	}
	if !view.Sandbox {
		t.Fatal("sandbox flag should be true")
	}
	// The RAW token must never be stored — only its hash keys the row.
	if _, ok := repo.sessByHash[middleware.HashAPIKey(view.Token)]; !ok {
		t.Fatal("session not stored under token hash")
	}
	if _, ok := repo.sessByHash[view.Token]; ok {
		t.Fatal("raw token must not be a storage key")
	}
}

func TestCreateFromLinkInactiveLinkIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_dead", 100, nil)
	l.Status = "disabled"
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)

	_, err := svc.CreateFromLink(context.Background(), "pl_dead")
	if err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("disabled link err = %v want ErrPaymentLinkNotFound", err)
	}
}

func TestCreateFromLinkUnknownIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.CreateFromLink(context.Background(), "pl_nope"); err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("unknown link err = %v want ErrPaymentLinkNotFound", err)
	}
}

var _ = time.Now // retained; time is used by later tasks' tests
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/service/ -run TestCreateFromLink -v`
Expected: FAIL — `NewCheckoutService` undefined.

- [ ] **Step 3: Write the implementation**

Create `internal/service/checkout_service.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/money"
	"github.com/yourco/payment-gateway/internal/repository"
)

// checkoutSessionTTL is the short lifetime of a hosted-checkout session.
const checkoutSessionTTL = 30 * time.Minute

// Charger is the slice of PaymentService the checkout needs (card charge).
// *paymentService satisfies it.
type Charger interface {
	Create(ctx context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error)
}

// QRIssuer is the slice of QRService the checkout needs (PromptPay mint + poll).
// *qrService satisfies it.
type QRIssuer interface {
	Create(ctx context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.QRPayment, error)
}

// Tokenizer turns a raw PAN into a vault token. *crypto.Vault satisfies it. Used
// ONLY on the sandbox card path; the token is handed to the payment service,
// which detokenizes it inside PCI scope.
type Tokenizer interface {
	Tokenize(ctx context.Context, pan string) (token string, last4 string, err error)
}

// CheckoutService drives the public hosted-checkout page. It turns a payment link
// into a short-lived session authenticated by an opaque token (stored hashed),
// initiates card / PromptPay payments via the existing services, and reports
// status for polling. The session token is the ONLY credential; every operation
// is scoped to the session it resolves — merchant context comes from the row,
// never the caller.
type CheckoutService interface {
	CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error)
	Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error)
	Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error)
}

type checkoutService struct {
	repo    repository.Querier
	charger Charger
	qr      QRIssuer
	vault   Tokenizer
	sandbox bool
	log     zerolog.Logger
}

// NewCheckoutService wires the checkout service. sandbox gates the raw-PAN card
// path: when false, card pay is refused (real hosted-fields tokenization is out
// of scope). vault/charger/qr may be nil in scaffolding but are required for the
// pay paths.
func NewCheckoutService(repo repository.Querier, charger Charger, qr QRIssuer, vault Tokenizer, sandbox bool, log zerolog.Logger) CheckoutService {
	return &checkoutService{
		repo:    repo,
		charger: charger,
		qr:      qr,
		vault:   vault,
		sandbox: sandbox,
		log:     log.With().Str("service", "checkout").Logger(),
	}
}

// CreateFromLink opens a session for an active, unexpired payment link. The raw
// token is returned once (on the view); only its hash is persisted.
func (s *checkoutService) CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error) {
	if s.repo == nil {
		return nil, domain.ErrCheckoutSessionNotFound
	}
	link, err := s.repo.GetPaymentLinkByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	// A link must be active and unexpired to open a checkout. We return the same
	// 404 for missing / disabled / expired so the endpoint is not an existence or
	// state oracle.
	if link.Status != "active" {
		return nil, domain.ErrPaymentLinkNotFound
	}
	if link.ExpiresAt.Valid && time.Now().After(link.ExpiresAt.Time) {
		return nil, domain.ErrPaymentLinkNotFound
	}

	token, err := generateSessionToken()
	if err != nil {
		return nil, err
	}
	row, err := s.repo.CreateCheckoutSession(ctx, repository.CreateCheckoutSessionParams{
		ID:               toPgUUID(uuid.New()),
		MerchantID:       link.MerchantID,
		PaymentLinkID:    link.ID,
		SessionTokenHash: middleware.HashAPIKey(token),
		AmountMinor:      link.AmountMinor,
		Currency:         link.Currency,
		Status:           string(domain.CheckoutOpen),
		SelectedMethod:   "",
		CustomerEmail:    "",
		ReturnUrl:        "",
		ExpiresAt:        pgTimestamptz(time.Now().Add(checkoutSessionTTL)),
		// PaymentID / QrPaymentID left as zero pgtype.UUID (NULL) until a method runs.
	})
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, row, &link)
	view.Token = token
	return view, nil
}

// Get and Pay are implemented in Tasks 4–5. Stubbed so the file compiles now.
func (s *checkoutService) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	return nil, domain.ErrNotImplemented
}

func (s *checkoutService) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	return nil, domain.ErrNotImplemented
}

// buildView renders the public, secret-free projection. link may be passed in to
// avoid a re-read; when nil and the session references one, it is loaded (scoped
// by merchant id). Merchant name is looked up best-effort.
func (s *checkoutService) buildView(ctx context.Context, row repository.CheckoutSession, link *repository.PaymentLink) *domain.CheckoutSessionView {
	if link == nil && row.PaymentLinkID.Valid {
		if l, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		}); err == nil {
			link = &l
		}
	}
	name := ""
	if m, err := s.repo.GetMerchant(ctx, row.MerchantID); err == nil {
		name = m.Name
	}
	view := &domain.CheckoutSessionView{
		ID:             pgUUIDToUUID(row.ID),
		Status:         row.Status,
		AmountMinor:    row.AmountMinor,
		Currency:       row.Currency,
		MerchantName:   name,
		SelectedMethod: row.SelectedMethod,
		ReturnURL:      row.ReturnUrl,
		Sandbox:        s.sandbox,
	}
	if row.ExpiresAt.Valid {
		view.ExpiresAt = row.ExpiresAt.Time
	}
	var allowed []string
	if link != nil {
		view.Title = link.Title
		view.Description = link.Description
		view.ImageURL = link.ImageUrl
		allowed = link.AllowedMethods
	}
	view.AllowedMethods = domain.DisplayMethods(allowed)
	return view
}

// generateSessionToken returns a cryptographically random, URL-safe opaque token
// with a recognizable prefix. Only its hash is persisted.
func generateSessionToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "cs_" + hex.EncodeToString(buf), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/ -run TestCreateFromLink -v`
Expected: PASS (all three). Then `go build ./... && go test ./internal/service/...` for regressions.

- [ ] **Step 5: Commit**

```bash
git add internal/service/checkout_service.go internal/service/checkout_service_test.go
git commit -m "feat(service): checkout CreateFromLink (token+hash, display view)"
```

---

### Task 4: `CheckoutService.Pay` — PromptPay + sandbox-gated card

**Files:**
- Modify: `internal/service/checkout_service.go` (replace `Pay` stub; add helpers)
- Modify: `internal/service/checkout_service_test.go` (append Task 4 tests; reuse Task 3 fakes)

**Interfaces:**
- Consumes: `Charger.Create`, `QRIssuer.Create`, `Tokenizer.Tokenize`, `repository.UpdateCheckoutSession`, `repository.GetPaymentLink`; `money.FromMinor`.
- Produces: `Pay` transitions an `open` session → `requires_action` (PromptPay QR minted / card 3DS) or `paid` (card captured) or `failed` (decline); returns the updated view (with `QRPayload` / `NextActionURL` when relevant).

- [ ] **Step 1: Write the failing tests**

Append to `internal/service/checkout_service_test.go`:
```go
func openSession(t *testing.T, repo *fakeCheckoutRepo, svc CheckoutService) string {
	t.Helper()
	view, err := svc.CreateFromLink(context.Background(), "pl_abc")
	if err != nil {
		t.Fatalf("CreateFromLink: %v", err)
	}
	return view.Token
}

func TestPayPromptPayMintsQRAndRequiresAction(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	qr := &fakeQR{createFn: func(_ context.Context, req domain.CreateQRRequest) (*domain.QRPayment, error) {
		if req.Method != domain.QRPromptPayDynamic {
			t.Fatalf("method = %v want promptpay_dynamic", req.Method)
		}
		if !req.Amount.Equal(decimalFromMinor(t, 5000, "THB")) {
			t.Fatalf("amount not converted to major units: %s", req.Amount)
		}
		return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "PP-EMV"}, nil
	}}
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.QRPayload != "PP-EMV" {
		t.Fatalf("qr_payload = %q want PP-EMV", view.QRPayload)
	}
	if view.SelectedMethod != "promptpay" {
		t.Fatalf("selected_method = %q", view.SelectedMethod)
	}
}

func TestPayCardCapturedMarksPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "promptpay"})
	charger := &fakeCharger{createFn: func(_ context.Context, idemKey string, req domain.CreatePaymentRequest) (*domain.Payment, error) {
		if !strings.HasPrefix(idemKey, "cs_") {
			t.Fatalf("idem key = %q want cs_<session>", idemKey)
		}
		if !req.Capture {
			t.Fatal("card charge must Capture=true")
		}
		if req.PaymentToken != "tokv1.fake" {
			t.Fatalf("payment token = %q want vault token", req.PaymentToken)
		}
		return &domain.Payment{ID: uuid.New(), Status: domain.StatusCaptured}, nil
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutPaid) {
		t.Fatalf("status = %q want paid", view.Status)
	}
	// single_use link must be closed as paid.
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid]", repo.linkStatusSets)
	}
}

func TestPayCardRequiresActionOn3DS(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		return &domain.Payment{ID: uuid.New(), Status: domain.StatusRequiresAction, NextActionURL: "https://acs/challenge"}, nil
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000003220", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.NextActionURL != "https://acs/challenge" {
		t.Fatalf("next_action_url = %q", view.NextActionURL)
	}
	if len(repo.linkStatusSets) != 0 {
		t.Fatal("link must NOT be marked paid on 3DS pending")
	}
}

func TestPayCardDeclinedMarksFailedAndSurfacesError(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	charger := &fakeCharger{createFn: func(_ context.Context, _ string, _ domain.CreatePaymentRequest) (*domain.Payment, error) {
		return nil, domain.ErrCardDeclined
	}}
	svc := newCheckoutSvc(repo, charger, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4000000000000002", ExpMonth: 1, ExpYear: 2031, CVV: "999"},
	})
	if err != domain.ErrCardDeclined {
		t.Fatalf("err = %v want ErrCardDeclined", err)
	}
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	if row.Status != string(domain.CheckoutFailed) {
		t.Fatalf("session status = %q want failed", row.Status)
	}
}

func TestPayCardBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false) // sandbox OFF
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (card is sandbox-only)", err)
	}
}

func TestPayMethodNotAllowedByLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"}) // card NOT allowed
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{
		Method: "card",
		Card:   &domain.CardInput{Number: "4111111111111111", ExpMonth: 12, ExpYear: 2030, CVV: "123"},
	})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable", err)
	}
}

func TestPayUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	_, err := svc.Pay(context.Background(), "cs_nope", domain.CheckoutPayRequest{Method: "promptpay"})
	if err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}

// decimalFromMinor is a test helper mirroring money.FromMinor for assertions.
func decimalFromMinor(t *testing.T, minor int64, currency string) decimalDecimal {
	t.Helper()
	d, err := moneyFromMinor(minor, currency)
	if err != nil {
		t.Fatalf("FromMinor: %v", err)
	}
	return d
}
```

> The two aliases at the bottom keep the test import list tidy: add these to the imports/vars of the test file. At the top of the file's import block add `moneypkg "github.com/yourco/payment-gateway/internal/pkg/money"` and `"github.com/shopspring/decimal"`, then define near the fakes:
> ```go
> type decimalDecimal = decimal.Decimal
> func moneyFromMinor(minor int64, currency string) (decimal.Decimal, error) { return moneypkg.FromMinor(minor, currency) }
> ```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/service/ -run TestPay -v`
Expected: FAIL — `Pay` returns `ErrNotImplemented` (assertions fail / wrong error).

- [ ] **Step 3: Replace the `Pay` stub + add helpers**

In `internal/service/checkout_service.go`, delete the `Pay` stub and add the following (keep the `Get` stub for Task 5):
```go
// Pay initiates payment for an open session with the selected method. It is
// idempotent: a session that is no longer open returns its current view (a
// double-submit is a no-op), and an expired session is swept to expired.
func (s *checkoutService) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		_, _ = s.transition(ctx, row, domain.CheckoutExpired)
		return nil, domain.ErrCheckoutSessionExpired
	}
	// Only an open session can be paid; anything else returns its current view
	// (idempotent double-submit).
	if domain.CheckoutStatus(row.Status) != domain.CheckoutOpen {
		return s.Get(ctx, token)
	}
	// The method must be one this link + this phase actually supports.
	allowed := domain.DisplayMethods(s.linkAllowedMethods(ctx, row))
	if !contains(allowed, req.Method) {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	switch req.Method {
	case "promptpay":
		return s.payPromptPay(ctx, row, req)
	case "card":
		return s.payCard(ctx, row, req)
	default:
		return nil, domain.ErrCheckoutMethodUnavailable
	}
}

func (s *checkoutService) payPromptPay(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	if s.qr == nil {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	qr, err := s.qr.Create(ctx, domain.CreateQRRequest{
		MerchantID: pgUUIDToUUID(row.MerchantID),
		Method:     domain.QRPromptPayDynamic,
		Amount:     amount,
		Currency:   row.Currency,
	})
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(domain.CheckoutRequiresAction),
		SelectedMethod: "promptpay",
		PaymentID:      row.PaymentID,
		QrPaymentID:    toPgUUID(qr.ID),
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, updated, nil)
	view.QRPayload = qr.QRPayload
	return view, nil
}

func (s *checkoutService) payCard(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	// Receiving a raw PAN is sandbox-only. In production, real hosted-fields
	// tokenization is out of scope, so refuse rather than accept card data.
	if !s.sandbox {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	if req.Card == nil {
		return nil, domain.ErrInvalidRequest
	}
	if s.vault == nil || s.charger == nil {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	// Tokenize the PAN so the PaymentService receives a vault token (never a PAN).
	tok, _, err := s.vault.Tokenize(ctx, req.Card.Number)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	amount, err := money.FromMinor(row.AmountMinor, row.Currency)
	if err != nil {
		return nil, domain.ErrInvalidRequest
	}
	sessionID := pgUUIDToUUID(row.ID)
	p, err := s.charger.Create(ctx, "cs_"+sessionID.String(), domain.CreatePaymentRequest{
		MerchantID:   pgUUIDToUUID(row.MerchantID),
		Amount:       amount,
		Currency:     row.Currency,
		PaymentToken: tok,
		Capture:      true,
		ReturnURL:    row.ReturnUrl,
	})
	if err != nil {
		// A decline / insufficient funds marks the session failed and surfaces the
		// error (mapped to 402 by the error handler) so the page can show it.
		_, _ = s.transition(ctx, row, domain.CheckoutFailed)
		return nil, err
	}
	next := domain.CheckoutPaid
	if p.Status == domain.StatusRequiresAction {
		next = domain.CheckoutRequiresAction
	}
	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(next),
		SelectedMethod: "card",
		PaymentID:      toPgUUID(p.ID),
		QrPaymentID:    row.QrPaymentID,
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		return nil, err
	}
	if next == domain.CheckoutPaid {
		s.markLinkPaid(ctx, updated)
	}
	view := s.buildView(ctx, updated, nil)
	view.NextActionURL = p.NextActionURL
	return view, nil
}

// ---- shared helpers (also used by Get in Task 5) ---------------------------

func (s *checkoutService) loadByToken(ctx context.Context, token string) (repository.CheckoutSession, error) {
	if s.repo == nil {
		return repository.CheckoutSession{}, domain.ErrCheckoutSessionNotFound
	}
	row, err := s.repo.GetCheckoutSessionByTokenHash(ctx, middleware.HashAPIKey(token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.CheckoutSession{}, domain.ErrCheckoutSessionNotFound
		}
		return repository.CheckoutSession{}, err
	}
	return row, nil
}

// transition updates only the status, preserving the row's other fields.
func (s *checkoutService) transition(ctx context.Context, row repository.CheckoutSession, status domain.CheckoutStatus) (repository.CheckoutSession, error) {
	return s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(status),
		SelectedMethod: row.SelectedMethod,
		PaymentID:      row.PaymentID,
		QrPaymentID:    row.QrPaymentID,
		CustomerEmail:  row.CustomerEmail,
	})
}

// markLinkPaid closes a single_use link once its session is paid. Best-effort:
// a failure is logged, not fatal to the payment.
func (s *checkoutService) markLinkPaid(ctx context.Context, row repository.CheckoutSession) {
	if !row.PaymentLinkID.Valid {
		return
	}
	link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	})
	if err != nil || link.LinkType != "single_use" {
		return
	}
	if _, err := s.repo.UpdatePaymentLinkStatus(ctx, repository.UpdatePaymentLinkStatusParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID, Status: "paid",
	}); err != nil {
		s.log.Warn().Err(err).Msg("mark link paid failed")
	}
}

// linkAllowedMethods returns the session link's allowed methods (nil if none).
func (s *checkoutService) linkAllowedMethods(ctx context.Context, row repository.CheckoutSession) []string {
	if !row.PaymentLinkID.Valid {
		return nil
	}
	link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID: row.PaymentLinkID, MerchantID: row.MerchantID,
	})
	if err != nil {
		return nil
	}
	return link.AllowedMethods
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func strFallback(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/ -run 'TestPay|TestCreateFromLink' -v`
Expected: PASS. Then `go build ./... && go test ./internal/service/...`.

- [ ] **Step 5: Commit**

```bash
git add internal/service/checkout_service.go internal/service/checkout_service_test.go
git commit -m "feat(service): checkout Pay (promptpay mint + sandbox card charge)"
```

---

### Task 5: `CheckoutService.Get` — status poll, QR sync, expiry

**Files:**
- Modify: `internal/service/checkout_service.go` (replace `Get` stub; add `syncStatus`)
- Modify: `internal/service/checkout_service_test.go` (append Task 5 tests; reuse fakes)

**Interfaces:**
- Consumes: `QRIssuer.Get`, `repository.UpdateCheckoutSession`, `repository.GetPaymentLink`, `repository.UpdatePaymentLinkStatus`.
- Produces: `Get` returns the current view, first syncing: expired sweep, and for a PromptPay session whose QR is `paid`/`expired`/`failed` it transitions the session accordingly (and closes a single_use link on paid). It re-surfaces the QR payload so a reloaded page can re-render.

- [ ] **Step 1: Write the failing tests**

Append to `internal/service/checkout_service_test.go`:
```go
func TestGetPromptPaySyncsToPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"})
	// QR is awaiting at pay time, then reports paid on the next Get.
	var calls int
	qr := &fakeQR{
		createFn: func(_ context.Context, _ domain.CreateQRRequest) (*domain.QRPayment, error) {
			return &domain.QRPayment{ID: uuid.New(), Status: domain.QRAwaitingPayment, QRPayload: "PP-EMV"}, nil
		},
		getFn: func(_ context.Context, _, id uuid.UUID) (*domain.QRPayment, error) {
			calls++
			st := domain.QRAwaitingPayment
			if calls >= 2 {
				st = domain.QRPaid
			}
			return &domain.QRPayment{ID: id, Status: st, QRPayload: "PP-EMV"}, nil
		},
	}
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"}); err != nil {
		t.Fatalf("Pay: %v", err)
	}

	// First Get: QR still awaiting -> requires_action, payload re-surfaced.
	v1, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get1: %v", err)
	}
	if v1.Status != string(domain.CheckoutRequiresAction) || v1.QRPayload != "PP-EMV" {
		t.Fatalf("v1 = %q / %q", v1.Status, v1.QRPayload)
	}
	// Second Get: QR paid -> session paid, link closed.
	v2, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get2: %v", err)
	}
	if v2.Status != string(domain.CheckoutPaid) {
		t.Fatalf("v2 status = %q want paid", v2.Status)
	}
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status = %v want [paid]", repo.linkStatusSets)
	}
}

func TestGetSweepsExpiredSession(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, nil)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	// Force the stored session to be already expired.
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	row.ExpiresAt = pgTimestamptz(time.Now().Add(-time.Minute))
	repo.sessByHash[row.SessionTokenHash] = row
	repo.sessByID[uuid.UUID(row.ID.Bytes)] = row

	v, err := svc.Get(context.Background(), tok)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if v.Status != string(domain.CheckoutExpired) {
		t.Fatalf("status = %q want expired", v.Status)
	}
}

func TestGetUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.Get(context.Background(), "cs_nope"); err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/service/ -run TestGet -v`
Expected: FAIL — `Get` returns `ErrNotImplemented`.

- [ ] **Step 3: Replace the `Get` stub + add `syncStatus`**

In `internal/service/checkout_service.go`, delete the `Get` stub and add:
```go
// Get returns the current session view. It first syncs live state: an expired
// session is swept to expired; a PromptPay session reflects its QR payment's
// confirmed/expired/failed state (and closes a single_use link on paid).
func (s *checkoutService) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	row, err = s.syncStatus(ctx, row)
	if err != nil {
		return nil, err
	}
	view := s.buildView(ctx, row, nil)
	// Re-surface the PromptPay payload so a reloaded page can re-render the QR.
	if row.SelectedMethod == "promptpay" && row.QrPaymentID.Valid &&
		row.Status == string(domain.CheckoutRequiresAction) && s.qr != nil {
		if qr, gerr := s.qr.Get(ctx, pgUUIDToUUID(row.MerchantID), pgUUIDToUUID(row.QrPaymentID)); gerr == nil {
			view.QRPayload = qr.QRPayload
		}
	}
	return view, nil
}

// syncStatus advances a non-terminal session based on live state. Terminal
// states (paid/failed/expired) are returned unchanged.
func (s *checkoutService) syncStatus(ctx context.Context, row repository.CheckoutSession) (repository.CheckoutSession, error) {
	status := domain.CheckoutStatus(row.Status)
	if status == domain.CheckoutPaid || status == domain.CheckoutFailed || status == domain.CheckoutExpired {
		return row, nil
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		return s.transition(ctx, row, domain.CheckoutExpired)
	}
	if row.SelectedMethod == "promptpay" && row.QrPaymentID.Valid &&
		status == domain.CheckoutRequiresAction && s.qr != nil {
		qr, err := s.qr.Get(ctx, pgUUIDToUUID(row.MerchantID), pgUUIDToUUID(row.QrPaymentID))
		if err != nil {
			return row, nil // transient read error: report no change, poll again
		}
		switch qr.Status {
		case domain.QRPaid:
			updated, terr := s.transition(ctx, row, domain.CheckoutPaid)
			if terr != nil {
				return row, terr
			}
			s.markLinkPaid(ctx, updated)
			return updated, nil
		case domain.QRExpired:
			return s.transition(ctx, row, domain.CheckoutExpired)
		case domain.QRFailed:
			return s.transition(ctx, row, domain.CheckoutFailed)
		}
	}
	return row, nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/ -run 'TestGet|TestPay|TestCreateFromLink' -v`
Expected: PASS. Then `go build ./... && go test ./internal/service/...`.

- [ ] **Step 5: Commit**

```bash
git add internal/service/checkout_service.go internal/service/checkout_service_test.go
git commit -m "feat(service): checkout Get with QR poll sync + expiry sweep"
```

---

### Task 6: Handler + IP rate limiter + config + routes + main wiring

**Files:**
- Create: `internal/handler/checkout_handler.go`
- Test: `internal/handler/checkout_handler_test.go`
- Modify: `internal/middleware/middleware.go` (add `CheckoutRateLimiter`)
- Test: `internal/middleware/checkout_limit_test.go`
- Modify: `internal/config/config.go` (+ `CheckoutRateLimitPerMin`), `.env.example`
- Modify: `internal/handler/handler.go` (add `Checkout *CheckoutHandler` + `WithCheckout`)
- Modify: `internal/router/router.go` (add `checkoutLimit fiber.Handler` arg; mount `/checkout`)
- Modify: `cmd/server/main.go` (build service + handler + limiter; update `router.Setup` call)

**Interfaces:**
- Consumes: `service.CheckoutService`, `middleware.clientIP`, `validationErrorResponse`.
- Produces: `handler.NewCheckoutHandler(svc service.CheckoutService, log zerolog.Logger) *CheckoutHandler` with methods `Create`, `Get`, `Pay`; `middleware.CheckoutRateLimiter(perMin int) fiber.Handler`; `Config.CheckoutRateLimitPerMin int` (default 30).

- [ ] **Step 1: Write the failing handler test**

Create `internal/handler/checkout_handler_test.go`:
```go
package handler

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
)

type fakeCheckoutSvc struct {
	createFn func(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error)
	getFn    func(ctx context.Context, token string) (*domain.CheckoutSessionView, error)
	payFn    func(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error)
}

func (f *fakeCheckoutSvc) CreateFromLink(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error) {
	return f.createFn(ctx, publicID)
}
func (f *fakeCheckoutSvc) Get(ctx context.Context, token string) (*domain.CheckoutSessionView, error) {
	return f.getFn(ctx, token)
}
func (f *fakeCheckoutSvc) Pay(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	return f.payFn(ctx, token, req)
}

func newCheckoutApp(svc *fakeCheckoutSvc) *fiber.App {
	h := NewCheckoutHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/checkout/sessions", h.Create)
	app.Get("/v1/checkout/sessions/:token", h.Get)
	app.Post("/v1/checkout/sessions/:token/pay", h.Pay)
	return app
}

func TestCheckoutCreateReturnsTokenView(t *testing.T) {
	svc := &fakeCheckoutSvc{createFn: func(_ context.Context, publicID string) (*domain.CheckoutSessionView, error) {
		if publicID != "pl_abc" {
			t.Fatalf("public id = %q", publicID)
		}
		return &domain.CheckoutSessionView{ID: uuid.New(), Token: "cs_tok", Status: "open", AmountMinor: 5000, Currency: "THB", AllowedMethods: []string{"card", "promptpay"}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions", strings.NewReader(`{"link":"pl_abc"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 201 (%s)", resp.StatusCode, b)
	}
	var env domain.APIResponse
	body, _ := io.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &env)
	data, _ := env.Data.(map[string]any)
	if data["session_token"] != "cs_tok" {
		t.Fatalf("session_token = %v want cs_tok", data["session_token"])
	}
}

func TestCheckoutCreateMissingLinkIs400(t *testing.T) {
	app := newCheckoutApp(&fakeCheckoutSvc{})
	req := httptest.NewRequest("POST", "/v1/checkout/sessions", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("status = %d want 400", resp.StatusCode)
	}
}

func TestCheckoutPayForwardsToken(t *testing.T) {
	var gotToken string
	svc := &fakeCheckoutSvc{payFn: func(_ context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
		gotToken = token
		return &domain.CheckoutSessionView{Status: "requires_action", QRPayload: "PP", SelectedMethod: req.Method, AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/pay", strings.NewReader(`{"method":"promptpay"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 200 (%s)", resp.StatusCode, b)
	}
	if gotToken != "cs_tok" {
		t.Fatalf("token = %q want cs_tok", gotToken)
	}
}

func TestCheckoutGetReturnsView(t *testing.T) {
	svc := &fakeCheckoutSvc{getFn: func(_ context.Context, token string) (*domain.CheckoutSessionView, error) {
		return &domain.CheckoutSessionView{ID: uuid.New(), Status: "paid", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)
	resp, _ := app.Test(httptest.NewRequest("GET", "/v1/checkout/sessions/cs_tok", nil))
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestCheckout -v`
Expected: FAIL — `NewCheckoutHandler` undefined.

- [ ] **Step 3: Write the handler**

Create `internal/handler/checkout_handler.go`:
```go
package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/service"
)

// CheckoutHandler serves the PUBLIC hosted-checkout endpoints. There is NO auth
// middleware on these routes: the opaque session token in the URL path is the
// credential, and the service scopes everything to the resolved session.
type CheckoutHandler struct {
	svc      service.CheckoutService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewCheckoutHandler(svc service.CheckoutService, log zerolog.Logger) *CheckoutHandler {
	return &CheckoutHandler{svc: svc, validate: validator.New(), log: log}
}

// Create opens a checkout session from a payment link's public id.
// @Router /v1/checkout/sessions [post]
func (h *CheckoutHandler) Create(c *fiber.Ctx) error {
	var req domain.CheckoutSessionRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	view, err := h.svc.CreateFromLink(c.Context(), req.Link)
	if err != nil {
		return err
	}
	return domain.Created(c, view)
}

// Get returns the current session state (page load / polling).
// @Router /v1/checkout/sessions/{token} [get]
func (h *CheckoutHandler) Get(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	view, err := h.svc.Get(c.Context(), token)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}

// Pay initiates payment with the selected method.
// @Router /v1/checkout/sessions/{token}/pay [post]
func (h *CheckoutHandler) Pay(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	var req domain.CheckoutPayRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	view, err := h.svc.Pay(c.Context(), token, req)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}
```

- [ ] **Step 4: Run handler test to verify it passes**

Run: `go test ./internal/handler/ -run TestCheckout -v`
Expected: PASS.

- [ ] **Step 5: Write the rate-limiter test + implementation**

Create `internal/middleware/checkout_limit_test.go`:
```go
package middleware

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func TestCheckoutRateLimiterBlocksOverLimit(t *testing.T) {
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler(zerolog.Nop())})
	app.Post("/c", CheckoutRateLimiter(1), func(c *fiber.Ctx) error { return c.SendStatus(200) })

	r1, _ := app.Test(httptest.NewRequest("POST", "/c", nil))
	if r1.StatusCode != 200 {
		t.Fatalf("first = %d want 200", r1.StatusCode)
	}
	r2, _ := app.Test(httptest.NewRequest("POST", "/c", nil))
	if r2.StatusCode != fiber.StatusTooManyRequests {
		t.Fatalf("second = %d want 429", r2.StatusCode)
	}
}
```

In `internal/middleware/middleware.go`, add after `SignupRateLimiter`:
```go
// CheckoutRateLimiter limits public hosted-checkout session creation per client
// IP (the endpoint is unauthenticated). It mirrors SignupRateLimiter but on a
// per-minute window suited to interactive checkout. perMin <= 0 falls back to 30.
func CheckoutRateLimiter(perMin int) fiber.Handler {
	if perMin <= 0 {
		perMin = 30
	}
	return limiter.New(limiter.Config{
		Max:        perMin,
		Expiration: time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return "checkout:" + clientIP(c)
		},
		LimitReached: func(c *fiber.Ctx) error {
			return domain.Error(c, fiber.StatusTooManyRequests, "RATE_LIMITED", "too many checkout attempts from this address; please try again shortly")
		},
	})
}
```

Run: `go test ./internal/middleware/ -run TestCheckoutRateLimiter -v` → Expected: PASS.

- [ ] **Step 6: Add config `CheckoutRateLimitPerMin`**

In `internal/config/config.go`: add a field near `SignupRateLimitPerHour`:
```go
	// CheckoutRateLimitPerMin caps how many public hosted-checkout session
	// creations (POST /v1/checkout/sessions) a single client IP may make per
	// minute. The endpoint is public/unauthenticated. Default 30.
	CheckoutRateLimitPerMin int `mapstructure:"CHECKOUT_RATE_LIMIT_PER_MIN"`
```
Add `"CHECKOUT_RATE_LIMIT_PER_MIN"` to `envKeys` (next to `"SIGNUP_RATE_LIMIT_PER_HOUR"`). In `Load()`, add a default next to the others:
```go
	v.SetDefault("CHECKOUT_RATE_LIMIT_PER_MIN", 30)
```
Append to `.env.example`:
```
# Max public checkout-session creations per client IP per minute (default 30).
CHECKOUT_RATE_LIMIT_PER_MIN=30
```

- [ ] **Step 7: Wire aggregate + router + main**

`internal/handler/handler.go`: add `Checkout *CheckoutHandler` to `Handlers` (next to `PaymentLink`) and a builder:
```go
// WithCheckout attaches the public hosted-checkout handler.
func (h *Handlers) WithCheckout(cx *CheckoutHandler) *Handlers {
	h.Checkout = cx
	return h
}
```

`internal/router/router.go`: add a `checkoutLimit fiber.Handler` parameter to `Setup` (place it right after `signupLimit`), extend the doc comment, and mount the group after the `/payment-links` block:
```go
	// Public hosted checkout (unauthenticated). The opaque session token in the
	// URL path is the credential; the service scopes every op to that session.
	// Session creation is IP-rate-limited (checkoutLimit) to blunt abuse.
	if h.Checkout != nil {
		checkout := v1.Group("/checkout")
		if checkoutLimit != nil {
			checkout.Post("/sessions", checkoutLimit, h.Checkout.Create)
		} else {
			checkout.Post("/sessions", h.Checkout.Create)
		}
		checkout.Get("/sessions/:token", h.Checkout.Get)
		checkout.Post("/sessions/:token/pay", h.Checkout.Pay)
	}
```
Update the signature line:
```go
func Setup(app *fiber.App, h *handler.Handlers, auth, sessionAuth, merchantAuth, adminAuth, rateLimit, signupLimit, checkoutLimit fiber.Handler, metrics fiber.Handler, webDir string, sandbox bool) {
```

`cmd/server/main.go`: after the `linkSvc` / `WithPaymentLinks` block (around line 200), add:
```go
	// Public hosted checkout. The payment service (svc) satisfies Charger, the QR
	// service satisfies QRIssuer, and the card vault satisfies Tokenizer. The card
	// (raw PAN) path is gated on SANDBOX_MODE inside the service.
	checkoutSvc := service.NewCheckoutService(repo, svc, qrSvc, vault, cfg.SandboxMode, logger)
	h = h.WithCheckout(handler.NewCheckoutHandler(checkoutSvc, logger))
```
Near where `signupLimit` is built (around line 233), add:
```go
	checkoutLimit := middleware.CheckoutRateLimiter(cfg.CheckoutRateLimitPerMin)
```
Update the `router.Setup(...)` call to pass `checkoutLimit` immediately after `signupLimit`:
```go
	router.Setup(app, h, auth, sessionAuth, merchantAuth, adminAuth, rateLimit, signupLimit, checkoutLimit, publicMetrics, cfg.WebDir, cfg.SandboxMode)
```

> Note: `svc` is typed `service.PaymentService`, which has a `Create(ctx, idemKey, req)` method — it satisfies `service.Charger`. `qrSvc` is `service.QRService` with `Create`/`Get` — it satisfies `service.QRIssuer`. `vault` is `*crypto.Vault` with `Tokenize` — it satisfies `service.Tokenizer`. No new construction needed.

- [ ] **Step 8: Build + full suite**

Run: `go build ./... && go test ./...`
Expected: PASS everywhere (fix the `router.Setup` call-site arg count until green — it is the single call site).

- [ ] **Step 9: Commit**

```bash
git add internal/handler/checkout_handler.go internal/handler/checkout_handler_test.go \
        internal/handler/handler.go internal/middleware/middleware.go \
        internal/middleware/checkout_limit_test.go internal/router/router.go \
        internal/config/config.go .env.example cmd/server/main.go
git commit -m "feat(api): public /v1/checkout routes + IP rate limit + wiring"
```

---

### Task 7: Frontend `/pay/[publicId]` — session create, method selector, card form

**Files:**
- Create: `web-app/app/pay/[publicId]/page.tsx` (server component shell + QR script)
- Create: `web-app/components/CheckoutClient.tsx` (client: create session, selector, card)
- Reuse: `web-app/lib/format.ts` (`formatMoney`, created in Phase 2)

**Interfaces:**
- Consumes: `POST /api/checkout/sessions`, `POST /api/checkout/sessions/:token/pay`, `GET /api/checkout/sessions/:token` via the Next `/api/*` proxy (client-side fetch; no cookie needed).

> The Next `/api/*` rewrite (`web-app/next.config.js`) proxies to `${BACKEND_URL}/v1/*`, so the browser calls `/api/checkout/...` same-origin. This page is PUBLIC — it does NOT use `serverGet`/cookies. The root layout (`web-app/app/layout.tsx`) is a bare shell with no auth guard, so `/pay/*` is reachable without login (verify: it only renders `<body>{children}</body>`).

- [ ] **Step 1: Page shell (server component) + QR script**

Create `web-app/app/pay/[publicId]/page.tsx`:
```tsx
import Script from "next/script";
import CheckoutClient from "@/components/CheckoutClient";

// Public hosted checkout. No auth / cookies. The vanilla qrcode.min.js asset
// (copied into /public in Task 8) exposes a global `QRCode`; loaded here once.
export default function PayPage({ params }: { params: { publicId: string } }) {
  return (
    <main className="min-h-screen bg-paycore-bg text-paycore-text flex items-start justify-center p-4">
      <Script src="/qrcode.min.js" strategy="afterInteractive" />
      <CheckoutClient publicId={params.publicId} />
    </main>
  );
}
```

- [ ] **Step 2: Checkout client component (create + selector + card)**

Create `web-app/components/CheckoutClient.tsx`:
```tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { formatMoney } from "@/lib/format";

type CheckoutView = {
  id: string;
  status: string;
  amount_minor: number;
  currency: string;
  merchant_name: string;
  title: string;
  description?: string;
  image_url?: string;
  allowed_methods: string[];
  selected_method?: string;
  qr_payload?: string;
  next_action_url?: string;
  return_url?: string;
  expires_at: string;
  sandbox: boolean;
  session_token?: string;
};

const METHOD_LABEL: Record<string, string> = {
  card: "บัตรเครดิต/เดบิต",
  promptpay: "PromptPay",
};

export default function CheckoutClient({ publicId }: { publicId: string }) {
  const [view, setView] = useState<CheckoutView | null>(null);
  const [token, setToken] = useState("");
  const [method, setMethod] = useState("");
  const [err, setErr] = useState("");
  const [busy, setBusy] = useState(false);

  // Card form state (sandbox only).
  const [cardNumber, setCardNumber] = useState("");
  const [expMonth, setExpMonth] = useState("");
  const [expYear, setExpYear] = useState("");
  const [cvv, setCvv] = useState("");

  const createdRef = useRef(false);

  // Create the session once on mount.
  useEffect(() => {
    if (createdRef.current) return;
    createdRef.current = true;
    (async () => {
      const res = await fetch("/api/checkout/sessions", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ link: publicId }),
      });
      const env = await res.json().catch(() => null);
      if (!res.ok) {
        setErr(env?.message ?? "ไม่พบลิงก์ชำระเงินนี้");
        return;
      }
      const v: CheckoutView = env.data;
      setView(v);
      setToken(v.session_token ?? "");
      if (v.allowed_methods.length === 1) setMethod(v.allowed_methods[0]);
    })();
  }, [publicId]);

  async function payCard(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        method: "card",
        card: {
          number: cardNumber.replace(/\s+/g, ""),
          exp_month: Number(expMonth),
          exp_year: Number(expYear),
          cvv,
        },
      }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "ชำระเงินไม่สำเร็จ");
      return;
    }
    const v: CheckoutView = env.data;
    setView(v);
    if (v.next_action_url) window.location.href = v.next_action_url; // 3DS redirect
  }

  if (err && !view) {
    return <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10">{err}</div>;
  }
  if (!view) {
    return <div className="text-paycore-muted mt-10">กำลังโหลด…</div>;
  }

  // Terminal / QR states are rendered by the PromptPay + status component (Task 8).
  if (view.status === "paid" || view.status === "expired" || view.status === "failed" || view.selected_method === "promptpay") {
    return <CheckoutStatusView token={token} initial={view} />;
  }

  return (
    <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10 space-y-5">
      <header>
        <p className="text-paycore-muted text-sm">{view.merchant_name}</p>
        <h1 className="text-xl font-semibold">{view.title}</h1>
        <p className="text-2xl font-bold mt-1">{formatMoney(view.amount_minor, view.currency)}</p>
        {view.description && <p className="text-paycore-muted text-sm mt-1">{view.description}</p>}
      </header>

      <div>
        <p className="text-sm text-paycore-muted mb-2">เลือกวิธีชำระเงิน</p>
        <div className="flex flex-wrap gap-2">
          {view.allowed_methods.map((m) => (
            <button
              key={m}
              onClick={() => setMethod(m)}
              className={`rounded-full px-4 py-2 text-sm border ${
                method === m ? "bg-paycore-primary border-paycore-primary text-white" : "border-white/15 text-paycore-muted"
              }`}
            >
              {METHOD_LABEL[m] ?? m}
            </button>
          ))}
        </div>
      </div>

      {method === "promptpay" && (
        <PayPromptPayButton token={token} onPaid={setView} setErr={setErr} />
      )}

      {method === "card" && view.sandbox && (
        <form onSubmit={payCard} className="space-y-3">
          <p className="text-xs rounded-lg bg-yellow-500/10 text-yellow-300 px-3 py-2">
            โหมดทดสอบ (Sandbox) — ใช้บัตรทดสอบเท่านั้น เช่น 4111 1111 1111 1111
          </p>
          <input value={cardNumber} onChange={(e) => setCardNumber(e.target.value)} inputMode="numeric"
            placeholder="หมายเลขบัตร" className="w-full rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
          <div className="flex gap-2">
            <input value={expMonth} onChange={(e) => setExpMonth(e.target.value)} inputMode="numeric"
              placeholder="MM" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
            <input value={expYear} onChange={(e) => setExpYear(e.target.value)} inputMode="numeric"
              placeholder="YYYY" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
            <input value={cvv} onChange={(e) => setCvv(e.target.value)} inputMode="numeric"
              placeholder="CVV" className="w-1/3 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" />
          </div>
          {err && <p className="text-red-400 text-sm">{err}</p>}
          <button disabled={busy} className="w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
            {busy ? "กำลังชำระเงิน…" : `ชำระ ${formatMoney(view.amount_minor, view.currency)}`}
          </button>
        </form>
      )}

      {method === "card" && !view.sandbox && (
        <p className="text-sm text-paycore-muted">การชำระด้วยบัตรยังไม่พร้อมใช้งานบนระบบนี้</p>
      )}
    </div>
  );
}

// Placeholder components implemented fully in Task 8; declared here so this file
// compiles independently. Task 8 REPLACES these with the QR + polling versions.
function PayPromptPayButton(_: { token: string; onPaid: (v: CheckoutView) => void; setErr: (s: string) => void }) {
  return null;
}
function CheckoutStatusView(_: { token: string; initial: CheckoutView }) {
  return null;
}
```

- [ ] **Step 3: Build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/pay/[publicId]` compiles (dynamic route). The placeholder `PayPromptPayButton`/`CheckoutStatusView` render nothing yet — replaced in Task 8.

- [ ] **Step 4: Commit**

```bash
git add web-app/app/pay/ web-app/components/CheckoutClient.tsx
git commit -m "feat(web): hosted checkout page — session create, method selector, card form"
```

---

### Task 8: Frontend — PromptPay QR render (vanilla asset) + polling + success/return

**Files:**
- Create: `web-app/public/qrcode.min.js` (copied from `web/assets/qrcode.min.js` — NO npm dep)
- Modify: `web-app/components/CheckoutClient.tsx` (replace the two placeholder components with real ones)

**Interfaces:**
- Consumes: `POST /api/checkout/sessions/:token/pay` (promptpay), `GET /api/checkout/sessions/:token` (poll), the global `window.QRCode` from the copied asset.

> The existing static site ships a dependency-free QR renderer at `web/assets/qrcode.min.js` (global `QRCode`, used as `new QRCode(el, { text, width, height, correctLevel: QRCode.CorrectLevel.M })`). We copy it into the Next `public/` dir and load it via `<Script>` (Task 7). This avoids adding any npm QR dependency, honoring the Global Constraints.

- [ ] **Step 1: Copy the QR asset into Next public/**

Run:
```bash
mkdir -p web-app/public
cp web/assets/qrcode.min.js web-app/public/qrcode.min.js
```
Expected: `web-app/public/qrcode.min.js` exists (~20KB). It is committed as a static asset (no build step transforms it).

- [ ] **Step 2: Add the global QRCode type + replace the placeholders**

At the TOP of `web-app/components/CheckoutClient.tsx`, just under the `"use client";` line, add a minimal ambient type for the global:
```tsx
declare global {
  interface Window {
    QRCode?: new (el: HTMLElement, opts: { text: string; width: number; height: number; correctLevel: number }) => unknown;
  }
}
```

Then REPLACE the two placeholder functions at the bottom of the file with these real implementations:
```tsx
// waitForQRCode resolves once the vanilla qrcode.min.js global is available
// (loaded via <Script strategy="afterInteractive">). Gives up after ~5s.
function waitForQRCode(): Promise<NonNullable<Window["QRCode"]>> {
  return new Promise((resolve, reject) => {
    const started = Date.now();
    const tick = () => {
      if (typeof window !== "undefined" && window.QRCode) return resolve(window.QRCode);
      if (Date.now() - started > 5000) return reject(new Error("QR library not loaded"));
      setTimeout(tick, 50);
    };
    tick();
  });
}

// PayPromptPayButton initiates the PromptPay charge, then hands off to the status
// view (which renders the QR + polls).
function PayPromptPayButton({ token, onPaid, setErr }: { token: string; onPaid: (v: CheckoutView) => void; setErr: (s: string) => void }) {
  const [busy, setBusy] = useState(false);
  async function start() {
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ method: "promptpay" }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "สร้าง QR ไม่สำเร็จ");
      return;
    }
    onPaid(env.data as CheckoutView); // status becomes requires_action -> status view takes over
  }
  return (
    <button onClick={start} disabled={busy} className="w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
      {busy ? "กำลังสร้าง QR…" : "สร้าง QR PromptPay"}
    </button>
  );
}

// CheckoutStatusView renders the QR (for PromptPay) and polls session status
// until a terminal state, then shows success + optional return_url.
function CheckoutStatusView({ token, initial }: { token: string; initial: CheckoutView }) {
  const [view, setView] = useState<CheckoutView>(initial);
  const qrBox = useRef<HTMLDivElement>(null);

  // Render the QR whenever a payload is present and not yet paid.
  useEffect(() => {
    if (!view.qr_payload || view.status === "paid") return;
    let cancelled = false;
    (async () => {
      try {
        const QR = await waitForQRCode();
        if (cancelled || !qrBox.current) return;
        qrBox.current.innerHTML = "";
        new QR(qrBox.current, { text: view.qr_payload!, width: 220, height: 220, correctLevel: 1 });
      } catch {
        /* leave the payload text visible as a fallback */
      }
    })();
    return () => { cancelled = true; };
  }, [view.qr_payload, view.status]);

  // Poll status until terminal.
  useEffect(() => {
    if (view.status === "paid" || view.status === "expired" || view.status === "failed") return;
    const id = setInterval(async () => {
      const res = await fetch(`/api/checkout/sessions/${token}`, { cache: "no-store" });
      if (!res.ok) return;
      const env = await res.json();
      setView(env.data as CheckoutView);
    }, 3000);
    return () => clearInterval(id);
  }, [token, view.status]);

  if (view.status === "paid") {
    return (
      <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-8 mt-10 text-center space-y-4">
        <div className="text-4xl">✓</div>
        <h1 className="text-xl font-semibold">ชำระเงินสำเร็จ</h1>
        <p className="text-paycore-muted">{formatMoney(view.amount_minor, view.currency)}</p>
        {view.return_url && (
          <a href={view.return_url} className="inline-block rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2">
            กลับไปที่ร้านค้า
          </a>
        )}
      </div>
    );
  }
  if (view.status === "expired" || view.status === "failed") {
    return (
      <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-8 mt-10 text-center space-y-3">
        <div className="text-4xl">⚠️</div>
        <h1 className="text-lg font-semibold">
          {view.status === "expired" ? "หมดเวลาชำระเงิน" : "ชำระเงินไม่สำเร็จ"}
        </h1>
        <p className="text-paycore-muted text-sm">โปรดลองอีกครั้ง</p>
      </div>
    );
  }

  // PromptPay awaiting: show QR + amount.
  return (
    <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10 text-center space-y-4">
      <p className="text-paycore-muted text-sm">{view.merchant_name}</p>
      <h1 className="text-lg font-semibold">สแกนเพื่อชำระด้วย PromptPay</h1>
      <p className="text-2xl font-bold">{formatMoney(view.amount_minor, view.currency)}</p>
      <div ref={qrBox} className="mx-auto bg-white p-3 rounded-lg inline-block" style={{ minHeight: 220, minWidth: 220 }} />
      <p className="text-paycore-muted text-xs">รอการยืนยันการชำระเงิน…</p>
    </div>
  );
}
```

> `correctLevel: 1` is `QRCode.CorrectLevel.M` in the vanilla lib (the enum is `{ L:1, M:0, Q:3, H:2 }` in davidshimjs/qrcodejs — verify against `window.QRCode.CorrectLevel.M` at runtime; if the QR fails to scan, switch to reading `QR.CorrectLevel.M` via the resolved global instead of the literal). Passing the numeric literal avoids referencing the enum before the type is known to TypeScript.

- [ ] **Step 3: Build**

Run: `cd web-app && npm run build`
Expected: build succeeds; the placeholders are gone; `/pay/[publicId]` compiles with the QR + polling views.

- [ ] **Step 4: Commit**

```bash
git add web-app/public/qrcode.min.js web-app/components/CheckoutClient.tsx
git commit -m "feat(web): PromptPay QR render (vanilla asset) + status polling + success"
```

---

### Task 9: End-to-end smoke (sandbox) + docs note

**Files:**
- Modify: `README.md` (short "Hosted checkout (Phase 3)" note — routes + sandbox card caveat)

**Interfaces:** none (integration verification of Tasks 1–8).

- [ ] **Step 1: Run the full automated suite**

Run: `go build ./... && go test ./...` and `cd web-app && npm run build`
Expected: all PASS / build succeeds. This is the real gate; the manual smoke below is best-effort.

- [ ] **Step 2: Manual backend smoke — PromptPay (sandbox)**

Start the backend (sandbox; needs `DATABASE_URL` / `.env`; PromptPay target set):
```bash
SANDBOX_MODE=true MIGRATE_ON_BOOT=true JWT_SECRET=dev-secret-dev-secret-dev-secret-xx \
  PUBLIC_BASE_URL=http://localhost:3000 PROMPTPAY_MOBILE_NO=0812345678 \
  QR_WEBHOOK_SECRET=dev-secret-dev-secret-dev-secret-xx make run
```
In another shell:
```bash
# 1. Dashboard login (dev) + create a link allowing both methods.
curl -s -X POST localhost:8080/v1/auth/dev-login -c /tmp/pc.cookies >/dev/null
LINK=$(curl -s -X POST localhost:8080/v1/payment-links -b /tmp/pc.cookies \
  -H 'Content-Type: application/json' \
  -d '{"title":"Coffee","amount_minor":5000,"allowed_methods":["card","promptpay"]}')
PUBLIC_ID=$(echo "$LINK" | sed -E 's/.*"public_id":"([^"]+)".*/\1/')

# 2. Public: create a checkout session (no auth).
SESS=$(curl -s -X POST localhost:8080/v1/checkout/sessions \
  -H 'Content-Type: application/json' -d "{\"link\":\"$PUBLIC_ID\"}")
TOKEN=$(echo "$SESS" | sed -E 's/.*"session_token":"([^"]+)".*/\1/')

# 3. Pay with PromptPay -> requires_action + qr_payload.
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/pay" \
  -H 'Content-Type: application/json' -d '{"method":"promptpay"}'

# 4. Confirm as the bank via the sandbox simulator, then poll to paid.
QRID=$(curl -s "localhost:8080/v1/sandbox/qr-payments?status=awaiting_payment" | sed -E 's/.*"id":"([^"]+)".*/\1/')
curl -s -X POST "localhost:8080/v1/sandbox/qr-payments/$QRID/pay" >/dev/null
curl -s "localhost:8080/v1/checkout/sessions/$TOKEN"   # expect "status":"paid"
```
Expected: step 3 returns `status":"requires_action"` with a non-empty `qr_payload`; step 4's final GET returns `status":"paid"`. (If starting the server is impractical, skip and say so — automated tests are the gate.)

- [ ] **Step 3: Manual backend smoke — card (sandbox)**

```bash
# New session on a fresh link, then charge a test card.
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/pay" \
  -H 'Content-Type: application/json' \
  -d '{"method":"card","card":{"number":"4111111111111111","exp_month":12,"exp_year":2030,"cvv":"123"}}'
```
Expected: `status":"paid"` (mock acquirer approves + captures). Repeat with `SANDBOX_MODE=false` on a card-only link → expect HTTP 422 `CHECKOUT_METHOD_UNAVAILABLE` (card entry not available in prod).

- [ ] **Step 4: Manual full-stack browser smoke (optional)**

```bash
# backend as above; frontend:
cd web-app && BACKEND_URL=http://localhost:8080 npm run dev
```
Browser: open `http://localhost:3000/pay/<PUBLIC_ID>` → merchant/amount render → choose PromptPay → QR renders → (confirm via sandbox as in Step 2 step 4) → page flips to "ชำระเงินสำเร็จ". Then a fresh link → card → test PAN → success.

- [ ] **Step 5: README note**

Add a short subsection to `README.md` under the existing merchant/checkout docs:
```markdown
### Hosted checkout (Phase 3)

Public, unauthenticated endpoints drive the `/pay/[publicId]` page. The opaque
session token (returned once on create, stored only as a SHA-256 hash) is the
credential:

- `POST /v1/checkout/sessions` `{ "link": "<public_id>" }` → `session_token` + display (IP rate-limited)
- `GET  /v1/checkout/sessions/:token` → status + display (poll)
- `POST /v1/checkout/sessions/:token/pay` `{ "method": "card"|"promptpay", ... }`

PromptPay works in all modes (confirmed via the QR webhook / sandbox simulator).
Card entry accepts a raw PAN ONLY when `SANDBOX_MODE=true`; in production it
returns `CHECKOUT_METHOD_UNAVAILABLE` (real hosted-fields tokenization is out of
scope). Money is stored in `checkout_sessions.amount_minor` (satang) and converted
to decimal major units before calling the payment / QR services.
```

- [ ] **Step 6: Commit**

```bash
git add README.md
git commit -m "docs: hosted checkout (Phase 3) endpoints + sandbox card caveat"
```

---

## Self-Review

**Spec coverage (design spec §3.3 checkout_sessions, §5.3 public checkout routes, §6 flow, §7 /pay page):**
- `checkout_sessions` table (all §3.3 columns: id, merchant_id, payment_link_id nullable, session_token_hash unique, amount_minor, currency, status, selected_method, payment_id nullable, qr_payment_id nullable, customer_email, return_url, expires_at, timestamps) → Task 1 ✓
- `POST /v1/checkout/sessions` (create from `{link}`, returns token + display, IP rate-limited) → Tasks 3 (service) + 6 (handler/route/limiter) ✓
- `GET /v1/checkout/sessions/:token` (status + display, poll) → Tasks 5 + 6 ✓
- `POST /v1/checkout/sessions/:token/pay` (card→next_action/paid; promptpay→qr payload+poll) → Tasks 4 + 6 ✓
- QR still confirmed via `/v1/webhooks/qr` + sandbox pay endpoint (unchanged; Get syncs from the qr_payment row) → Task 5 + Task 9 smoke ✓
- §6 flow (link → session token held in page memory → method selector from registry ∩ allowed_methods → pay → paid → success + return_url; single_use link → paid) → `DisplayMethods` (Task 2), `markLinkPaid` (Task 4/5), frontend (Tasks 7–8) ✓
- §7 `/pay/[publicId]` public page (merchant+amount+item, method selector card|promptpay, card form sandbox, promptpay QR render + poll, success) → Tasks 7–8 ✓
- Method registry (card, promptpay for Phase 3; wallets deferred to Phase 4) → `domain.CheckoutSupportedMethods` (Task 2) ✓
- `POST /v1/checkout/sessions/:token/confirm-mock` (§5.3) is a **Phase 4 (wallet)** endpoint — no card/promptpay flow needs it (PromptPay confirms via the existing QR webhook / sandbox). Intentionally out of Phase 3 scope; noted here.

**Global-constraint coverage:**
- No new Go dep (uses existing pgx/decimal/validator/fiber); no new npm dep (QR via copied vanilla asset) → Tasks 3–8 ✓
- Money integer minor units in table, converted via `money.FromMinor` before service calls → Tasks 3–5 (`payPromptPay`/`payCard` convert; asserted in `TestPayPromptPayMintsQRAndRequiresAction`) ✓
- Token opaque random, stored hashed (`middleware.HashAPIKey`), never stored/returned raw except once on create → Task 3 (`generateSessionToken`, `TestCreateFromLinkReturnsTokenAndDisplay` asserts raw token is NOT a storage key) ✓
- Session scoping: lookup only by token hash; merchant from row; updates by resolved id → Tasks 3–5 ✓
- Card PAN sandbox-gated → Task 4 (`payCard` returns `ErrCheckoutMethodUnavailable` when `!sandbox`; `TestPayCardBlockedOutsideSandbox`) ✓
- Public endpoints unauthenticated; create IP-rate-limited → Task 6 (no auth middleware on `/checkout`; `CheckoutRateLimiter` on `POST /sessions`) ✓
- Paired up/down migration + `migrations_test` → Task 1 ✓; sqlc only → Task 1 ✓

**Placeholder scan:** No `TODO`/`fill in`/"add error handling" placeholders. The frontend `PayPromptPayButton`/`CheckoutStatusView` are declared as compiling stubs in Task 7 **and explicitly replaced with complete code in Task 8** (called out in both tasks). The service `Get`/`Pay` stubs in Task 3 are explicitly replaced in Tasks 4/5. The `var _ = time.Now` guard in the test file is real (keeps `time` imported for later-task tests). All code steps contain complete code.

**Type consistency:**
- `CheckoutService` (`CreateFromLink`/`Get`/`Pay`) defined in Task 3, consumed by the handler + tests in Task 6, and by main wiring in Task 6.
- Consumer interfaces `Charger` (`Create(ctx, idemKey, req)`), `QRIssuer` (`Create`/`Get`), `Tokenizer` (`Tokenize`) defined in Task 3 — verified against the real signatures: `PaymentService.Create(ctx, idemKey string, req domain.CreatePaymentRequest)`, `QRService.Create(ctx, domain.CreateQRRequest)` / `QRService.Get(ctx, merchantID, id uuid.UUID)`, `*crypto.Vault.Tokenize(ctx, pan) (string,string,error)`. main.go passes `svc`, `qrSvc`, `vault` unchanged.
- `domain.CheckoutSessionView`/`CheckoutPayRequest`/`CardInput`/`CheckoutSessionRequest`/`DisplayMethods`/status constants defined Task 2, used Tasks 3–8.
- sqlc names (`CheckoutSession`, `CreateCheckoutSessionParams`, `UpdateCheckoutSessionParams` with `QrPaymentID`/`PaymentID`/`ReturnUrl`/`SessionTokenHash`/`SelectedMethod`/`CustomerEmail`, nullable `pgtype.UUID` FKs) are asserted in Task 1 Step 6 — **verify against generated code after `make sqlc`** and adjust Tasks 3–5 if any name differs (e.g. `QrPaymentID` casing).
- `money.FromMinor(minor int64, currency string) (decimal.Decimal, error)` — verified in `internal/pkg/money/money.go`.
- New errors `ErrCheckoutSessionNotFound`(404)/`ErrCheckoutSessionExpired`(410)/`ErrCheckoutMethodUnavailable`(422) defined Task 2, mapped in the same `switch` as `ErrPaymentLinkNotFound`, returned by Tasks 3–5.
- Frontend `formatMoney` (Phase 2 `web-app/lib/format.ts`) reused; the `/api/*` proxy (`web-app/next.config.js`) reused.

**Cross-task notes for execution:**
1. `router.Setup` signature grows by one arg (`checkoutLimit`, after `signupLimit`) — single call site in `cmd/server/main.go`.
2. `internal/service/checkout_service_test.go` is authored across Tasks 3–5: Task 3 defines all fakes (`fakeCheckoutRepo`, `fakeCharger`, `fakeQR`, `fakeVault`, `newCheckoutSvc`, `mkLink`); Tasks 4–5 append test functions only. Do not redefine the fakes.
3. `service.Get`/`service.Pay` are stubbed in Task 3 so the package compiles; they MUST be replaced in Tasks 4/5 (the plan deletes the stubs there). Running `go test ./internal/service/...` between Tasks 3 and 4 will pass Task 3's tests while `Get`/`Pay` still return `ErrNotImplemented` — that is expected until Tasks 4–5.
4. The QR enum literal (`correctLevel: 1`) is validated at runtime in Task 8's note; adjust if the copied lib's `CorrectLevel.M` differs.
