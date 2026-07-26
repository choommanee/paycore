# Phase 2 — Payment Links Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ให้ร้านค้า (ผ่าน dashboard cookie หรือ API key) สร้าง/ดู/ปิด **payment link** ที่มี public slug แชร์ได้ — เป็นวัตถุที่หน้า hosted checkout (Phase 3) จะเอาไป render

**Architecture:** เพิ่มตาราง `payment_links` + service/handler แบบ CRUD ตาม pattern เดิม. เพิ่ม **combined auth middleware** (`MerchantAuth`) ที่รับได้ทั้ง session cookie (dashboard) และ API key — payment-link routes อยู่หลัง middleware นี้ ทำให้ทั้ง dashboard และ API client ใช้งานได้. Frontend เพิ่มหน้า `/links` (list + create) และ `/links/[id]` (detail + copy URL + disable).

**Tech Stack:** Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 (arrays → `[]string` natively) · Next.js App Router + TypeScript + Tailwind. **ไม่มี dependency ใหม่** ทั้ง Go และ npm.

## Global Constraints

- Module `github.com/yourco/payment-gateway`. NO new Go/npm dependency.
- Money = integer minor units (สตางค์), `amount_minor > 0` (CHECK + validation).
- ทุก response ใช้ envelope กลาง `domain.Success/Created/Error`.
- DB access via sqlc เท่านั้น (`.sql` ใน `internal/repository/queries/`, รัน `make sqlc`; ห้ามแก้ `*.sql.go`).
- ทุก migration มีคู่ up/down (test `migrations_test.go` บังคับ). Migration ถัดไป = `000009`.
- **Tenant scoping (สำคัญ, กัน IDOR):** ทุก read/update ของ payment link ต้อง WHERE `merchant_id = <authed merchant>` ด้วยเสมอ — merchant id มาจาก auth context (cookie claims หรือ api-key), ไม่เชื่อ body/param.
- Service test: fake repo embed `repository.Querier` (nil) + override เฉพาะ method ที่ใช้ (pattern จาก `merchant_service_test.go`, `auth_service_test.go`).
- Handler test: fake service interface + `app.Test` + inject `middleware.LocalMerchantID` local (pattern จาก `merchant_handler_test.go`).
- Reuse helpers ที่มีอยู่ (อย่า redefine): `toPgUUID`, `pgUUIDToUUID`, `strPtr`, `ptrStr`, `clampLimit` (service pkg); `paginate`, `validationErrorResponse` (handler pkg); `middleware.MerchantIDFromCtx`, `middleware.UserIDFromCtx`, `middleware.SessionCookieName`, `middleware.HashAPIKey`, `middleware.extractAPIKey`(unexported — ดู Task 4).
- รันชุด: `go build ./... && go test ./...`; frontend `cd web-app && npm run build`.

---

### Task 1: Migration + sqlc — `payment_links`

**Files:**
- Create: `migrations/000009_payment_links.up.sql`, `migrations/000009_payment_links.down.sql`
- Create: `internal/repository/queries/payment_link.sql`
- Generate: `internal/repository/payment_link.sql.go`, `models.go`, `querier.go` (via `make sqlc`)
- Test: `migrations/migrations_paymentlink_test.go`

**Interfaces:**
- Produces (sqlc-generated): `repository.PaymentLink` model; querier methods `CreatePaymentLink`, `GetPaymentLink`, `GetPaymentLinkByPublicID`, `ListPaymentLinksByMerchant`, `UpdatePaymentLinkStatus` + their `*Params` structs.

- [ ] **Step 1: Write the failing migration test**

Create `migrations/migrations_paymentlink_test.go`:
```go
package migrations_test

import (
	"strings"
	"testing"
)

func TestPaymentLinksMigrationReversible(t *testing.T) {
	up := mustRead(t, "000009_payment_links.up.sql")
	down := mustRead(t, "000009_payment_links.down.sql")

	upU := strings.ToUpper(up)
	if !strings.Contains(upU, "CREATE TABLE PAYMENT_LINKS") {
		t.Fatalf("up does not create payment_links: %s", up)
	}
	if !strings.Contains(strings.ToLower(up), "public_id") || !strings.Contains(upU, "UNIQUE") {
		t.Fatalf("up must have a unique public_id index: %s", up)
	}
	if !strings.Contains(upU, "AMOUNT_MINOR") || !strings.Contains(upU, "CHECK") {
		t.Fatalf("up must CHECK amount_minor > 0: %s", up)
	}
	if !strings.Contains(strings.ToUpper(down), "DROP TABLE") || !strings.Contains(strings.ToLower(down), "payment_links") {
		t.Fatalf("down must drop payment_links: %s", down)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./migrations/ -run TestPaymentLinksMigrationReversible -v`
Expected: FAIL — file not found

- [ ] **Step 3: Write the migrations**

Create `migrations/000009_payment_links.up.sql`:
```sql
-- Shareable payment links. A merchant creates a link (fixed amount + allowed
-- methods); the public_id is the slug in the hosted-checkout URL (/pay/<public_id>).
-- created_by is nullable: links created via API key have no dashboard user.
CREATE TABLE payment_links (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id     UUID NOT NULL REFERENCES merchants(id) ON DELETE CASCADE,
    public_id       TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    amount_minor    BIGINT NOT NULL CHECK (amount_minor > 0),
    currency        TEXT NOT NULL DEFAULT 'THB',
    allowed_methods TEXT[] NOT NULL DEFAULT '{}',
    link_type       TEXT NOT NULL DEFAULT 'single_use',  -- single_use | reusable
    status          TEXT NOT NULL DEFAULT 'active',       -- active | paid | expired | disabled
    reference       TEXT NOT NULL DEFAULT '',
    image_url       TEXT NOT NULL DEFAULT '',
    expires_at      TIMESTAMPTZ,
    created_by      UUID REFERENCES merchant_users(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- The slug is globally unique (it addresses a checkout).
CREATE UNIQUE INDEX payment_links_public_id_idx ON payment_links (public_id);
-- Merchant's link list, newest first.
CREATE INDEX payment_links_merchant_idx ON payment_links (merchant_id, created_at DESC);
```

Create `migrations/000009_payment_links.down.sql`:
```sql
DROP INDEX IF EXISTS payment_links_merchant_idx;
DROP INDEX IF EXISTS payment_links_public_id_idx;
DROP TABLE IF EXISTS payment_links;
```

- [ ] **Step 4: Run the migration test to verify it passes**

Run: `go test ./migrations/ -v`
Expected: PASS (new test + existing pairing test)

- [ ] **Step 5: Write the sqlc queries**

Create `internal/repository/queries/payment_link.sql`:
```sql
-- internal/repository/queries/payment_link.sql
-- Payment links CRUD. Reads/updates are ALWAYS scoped by merchant_id to prevent
-- one merchant reading/altering another's link (IDOR); the public_id lookup is
-- the only unscoped read and is used by the (Phase 3) public checkout page.

-- name: CreatePaymentLink :one
INSERT INTO payment_links (
    id, merchant_id, public_id, title, description, amount_minor, currency,
    allowed_methods, link_type, status, reference, image_url, expires_at, created_by
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
RETURNING *;

-- name: GetPaymentLink :one
SELECT * FROM payment_links WHERE id = $1 AND merchant_id = $2;

-- name: GetPaymentLinkByPublicID :one
SELECT * FROM payment_links WHERE public_id = $1;

-- name: ListPaymentLinksByMerchant :many
SELECT * FROM payment_links
WHERE merchant_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePaymentLinkStatus :one
UPDATE payment_links
SET status = $3, updated_at = NOW()
WHERE id = $1 AND merchant_id = $2
RETURNING *;
```

- [ ] **Step 6: Generate + build**

Run: `make sqlc && go build ./...`
Expected: no errors. Open `internal/repository/models.go`, confirm `PaymentLink` exists and that `AllowedMethods` is generated as `[]string` (pgx maps `TEXT[]` to `[]string`) and `ExpiresAt`/`CreatedBy` are nullable (`pgtype.Timestamptz` / `pgtype.UUID`). Note the exact field names for Task 3.

- [ ] **Step 7: Commit**

```bash
git add migrations/000009_payment_links.up.sql migrations/000009_payment_links.down.sql \
        migrations/migrations_paymentlink_test.go internal/repository/queries/payment_link.sql \
        internal/repository/
git commit -m "feat(db): payment_links table + sqlc queries"
```

---

### Task 2: Domain types + slug/request validation

**Files:**
- Create: `internal/domain/payment_link.go`
- Test: `internal/domain/payment_link_test.go`

**Interfaces:**
- Produces:
  - `domain.CreatePaymentLinkRequest{ Title, Description string; AmountMinor int64; Currency string; AllowedMethods []string; LinkType, Reference, ImageURL string; ExpiresAt *time.Time }` with validate tags
  - `domain.PaymentLink{ ID, MerchantID uuid.UUID; PublicID, Title, Description string; AmountMinor int64; Currency string; AllowedMethods []string; LinkType, Status, Reference, ImageURL, URL string; ExpiresAt *time.Time; CreatedAt, UpdatedAt time.Time }`
  - `domain.ValidPaymentMethods` (the allowed method slugs) and `domain.IsValidMethod(string) bool`

- [ ] **Step 1: Write the failing test**

Create `internal/domain/payment_link_test.go`:
```go
package domain

import "testing"

func TestIsValidMethod(t *testing.T) {
	for _, m := range []string{"card", "promptpay", "truemoney"} {
		if !IsValidMethod(m) {
			t.Fatalf("%q should be valid", m)
		}
	}
	for _, m := range []string{"", "bitcoin", "CARD"} {
		if IsValidMethod(m) {
			t.Fatalf("%q should be invalid", m)
		}
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/domain/ -run TestIsValidMethod -v`
Expected: FAIL — `IsValidMethod` undefined

- [ ] **Step 3: Write the implementation**

Create `internal/domain/payment_link.go`:
```go
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ValidPaymentMethods is the full registry of method slugs a link may allow.
// The backend only truly processes card + promptpay today (Phase 3); the rest
// are Beam-parity slugs backed by mock adapters (Phase 4).
var ValidPaymentMethods = []string{
	"card", "promptpay", "mobile_banking", "truemoney",
	"shopeepay", "alipay", "wechat", "card_installment",
}

// IsValidMethod reports whether m is a known payment-method slug.
func IsValidMethod(m string) bool {
	for _, v := range ValidPaymentMethods {
		if v == m {
			return true
		}
	}
	return false
}

// CreatePaymentLinkRequest is the dashboard/API payload to create a link. The
// merchant id and created_by come from auth context, never the body.
type CreatePaymentLinkRequest struct {
	Title          string     `json:"title" validate:"required,max=200"`
	Description    string     `json:"description" validate:"omitempty,max=2000"`
	AmountMinor    int64      `json:"amount_minor" validate:"required,gt=0"`
	Currency       string     `json:"currency" validate:"omitempty,len=3"`
	AllowedMethods []string   `json:"allowed_methods" validate:"omitempty,dive,oneof=card promptpay mobile_banking truemoney shopeepay alipay wechat card_installment"`
	LinkType       string     `json:"link_type" validate:"omitempty,oneof=single_use reusable"`
	Reference      string     `json:"reference" validate:"omitempty,max=200"`
	ImageURL       string     `json:"image_url" validate:"omitempty,url,max=500"`
	ExpiresAt      *time.Time `json:"expires_at" validate:"omitempty"`
}

// PaymentLink is the API representation of a payment link. URL is the computed
// public checkout URL the merchant shares.
type PaymentLink struct {
	ID             uuid.UUID  `json:"id"`
	MerchantID     uuid.UUID  `json:"merchant_id"`
	PublicID       string     `json:"public_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	AmountMinor    int64      `json:"amount_minor"`
	Currency       string     `json:"currency"`
	AllowedMethods []string   `json:"allowed_methods"`
	LinkType       string     `json:"link_type"`
	Status         string     `json:"status"`
	Reference      string     `json:"reference,omitempty"`
	ImageURL       string     `json:"image_url,omitempty"`
	URL            string     `json:"url"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/domain/ -run TestIsValidMethod -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/domain/payment_link.go internal/domain/payment_link_test.go
git commit -m "feat(domain): payment link request/response types + method registry"
```

---

### Task 3: `PaymentLinkService` — create (with slug) / list / get / disable

**Files:**
- Create: `internal/service/payment_link_service.go`
- Test: `internal/service/payment_link_service_test.go`
- Reuse existing package helpers: `toPgUUID`, `pgUUIDToUUID`, `ptrStr`, `clampLimit`

**Interfaces:**
- Consumes: `repository.Querier` (`CreatePaymentLink`, `GetPaymentLink`, `ListPaymentLinksByMerchant`, `UpdatePaymentLinkStatus`)
- Produces:
  - `service.PaymentLinkService interface { Create(ctx, merchantID uuid.UUID, createdBy *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error); List(ctx, merchantID uuid.UUID, limit, offset int32) ([]*domain.PaymentLink, error); Get(ctx, merchantID, id uuid.UUID) (*domain.PaymentLink, error); Disable(ctx, merchantID, id uuid.UUID) (*domain.PaymentLink, error) }`
  - `service.NewPaymentLinkService(repo repository.Querier, publicBaseURL string, log zerolog.Logger) PaymentLinkService`
  - error `domain.ErrPaymentLinkNotFound` (add to `internal/domain/errors.go`)

- [ ] **Step 1: Add the not-found error + its 404 mapping**

(1) In `internal/domain/errors.go`, inside the existing `var ( ... )` block (which ends after `ErrForbidden` around line 20), add a line aligned with the others:
```go
	ErrPaymentLinkNotFound = errors.New("payment link not found")
```

(2) In `internal/middleware/middleware.go`, the error handler has a `switch { case errors.Is(err, ...) }` (around lines 190-210) that maps domain errors to HTTP responses. Add a case next to the `ErrMerchantNotFound` case:
```go
		case errors.Is(err, domain.ErrPaymentLinkNotFound):
			return domain.Error(c, fiber.StatusNotFound, "PAYMENT_LINK_NOT_FOUND", err.Error())
```
Without this case, a not-found link would fall through to a generic 500 instead of 404.

- [ ] **Step 2: Write the failing test**

Create `internal/service/payment_link_service_test.go`:
```go
package service

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

type fakeLinkRepo struct {
	repository.Querier
	mu      sync.Mutex
	byID    map[uuid.UUID]repository.PaymentLink
	slugs   map[string]bool
}

func newFakeLinkRepo() *fakeLinkRepo {
	return &fakeLinkRepo{byID: map[uuid.UUID]repository.PaymentLink{}, slugs: map[string]bool{}}
}

func (f *fakeLinkRepo) CreatePaymentLink(_ context.Context, a repository.CreatePaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.slugs[a.PublicID] = true
	pl := repository.PaymentLink{
		ID: a.ID, MerchantID: a.MerchantID, PublicID: a.PublicID, Title: a.Title,
		Description: a.Description, AmountMinor: a.AmountMinor, Currency: a.Currency,
		AllowedMethods: a.AllowedMethods, LinkType: a.LinkType, Status: a.Status,
		Reference: a.Reference, ImageUrl: a.ImageUrl,
	}
	f.byID[uuid.UUID(a.ID.Bytes)] = pl
	return pl, nil
}

func (f *fakeLinkRepo) GetPaymentLink(_ context.Context, a repository.GetPaymentLinkParams) (repository.PaymentLink, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	pl, ok := f.byID[uuid.UUID(a.ID.Bytes)]
	if !ok || uuid.UUID(pl.MerchantID.Bytes) != uuid.UUID(a.MerchantID.Bytes) {
		return repository.PaymentLink{}, pgx.ErrNoRows
	}
	return pl, nil
}

func newLinkSvc(repo repository.Querier) PaymentLinkService {
	return NewPaymentLinkService(repo, "https://pay.example", zerolog.Nop())
}

func TestCreatePaymentLinkGeneratesSlugAndURL(t *testing.T) {
	repo := newFakeLinkRepo()
	svc := newLinkSvc(repo)
	mid := uuid.New()
	uid := uuid.New()
	got, err := svc.Create(context.Background(), mid, &uid, domain.CreatePaymentLinkRequest{
		Title: "Coffee", AmountMinor: 5000,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got.PublicID == "" {
		t.Fatal("expected a generated public_id slug")
	}
	if got.Currency != "THB" {
		t.Fatalf("currency default = %q want THB", got.Currency)
	}
	if got.LinkType != "single_use" {
		t.Fatalf("link_type default = %q want single_use", got.LinkType)
	}
	if got.Status != "active" {
		t.Fatalf("status = %q want active", got.Status)
	}
	if !strings.HasSuffix(got.URL, "/pay/"+got.PublicID) || !strings.HasPrefix(got.URL, "https://pay.example") {
		t.Fatalf("url = %q want https://pay.example/pay/<slug>", got.URL)
	}
}

func TestGetPaymentLinkOtherMerchantIsNotFound(t *testing.T) {
	repo := newFakeLinkRepo()
	svc := newLinkSvc(repo)
	owner := uuid.New()
	created, _ := svc.Create(context.Background(), owner, nil, domain.CreatePaymentLinkRequest{Title: "X", AmountMinor: 100})

	// A different merchant must NOT be able to read it.
	_, err := svc.Get(context.Background(), uuid.New(), created.ID)
	if err != domain.ErrPaymentLinkNotFound {
		t.Fatalf("cross-merchant Get err=%v want ErrPaymentLinkNotFound", err)
	}
	// The owner can.
	got, err := svc.Get(context.Background(), owner, created.ID)
	if err != nil || got.ID != created.ID {
		t.Fatalf("owner Get failed: %v", err)
	}
}

var _ = pgtype.UUID{} // keep pgtype import if unused above
```

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/service/ -run 'TestCreatePaymentLink|TestGetPaymentLink' -v`
Expected: FAIL — `NewPaymentLinkService` undefined

- [ ] **Step 4: Write the implementation**

Create `internal/service/payment_link_service.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/repository"
)

// PaymentLinkService creates and manages shareable payment links. Every read and
// update is scoped by merchant id (from auth context) to prevent cross-merchant
// access (IDOR).
type PaymentLinkService interface {
	Create(ctx context.Context, merchantID uuid.UUID, createdBy *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error)
	List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.PaymentLink, error)
	Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error)
	Disable(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error)
}

type paymentLinkService struct {
	repo    repository.Querier
	baseURL string
	log     zerolog.Logger
}

// NewPaymentLinkService wires the service. publicBaseURL is the origin the shared
// checkout URL is built on (e.g. the Next app origin); the URL is
// <baseURL>/pay/<public_id>.
func NewPaymentLinkService(repo repository.Querier, publicBaseURL string, log zerolog.Logger) PaymentLinkService {
	return &paymentLinkService{repo: repo, baseURL: strings_TrimRightSlash(publicBaseURL), log: log.With().Str("service", "payment_link").Logger()}
}

func (s *paymentLinkService) Create(ctx context.Context, merchantID uuid.UUID, createdBy *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, errors.New("payment link repository not configured")
	}
	slug, err := generatePublicID()
	if err != nil {
		return nil, err
	}
	currency := req.Currency
	if currency == "" {
		currency = "THB"
	}
	linkType := req.LinkType
	if linkType == "" {
		linkType = "single_use"
	}
	methods := req.AllowedMethods
	if methods == nil {
		methods = []string{}
	}
	var createdByPg pgtype.UUID
	if createdBy != nil {
		createdByPg = toPgUUID(*createdBy)
	}
	var expiresPg pgtype.Timestamptz
	if req.ExpiresAt != nil {
		expiresPg = pgtype.Timestamptz{Time: *req.ExpiresAt, Valid: true}
	}
	row, err := s.repo.CreatePaymentLink(ctx, repository.CreatePaymentLinkParams{
		ID:             toPgUUID(uuid.New()),
		MerchantID:     toPgUUID(merchantID),
		PublicID:       slug,
		Title:          req.Title,
		Description:    req.Description,
		AmountMinor:    req.AmountMinor,
		Currency:       currency,
		AllowedMethods: methods,
		LinkType:       linkType,
		Status:         "active",
		Reference:      req.Reference,
		ImageUrl:       req.ImageURL,
		ExpiresAt:      expiresPg,
		CreatedBy:      createdByPg,
	})
	if err != nil {
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) List(ctx context.Context, merchantID uuid.UUID, limit, offset int32) ([]*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, nil
	}
	rows, err := s.repo.ListPaymentLinksByMerchant(ctx, repository.ListPaymentLinksByMerchantParams{
		MerchantID: toPgUUID(merchantID),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}
	out := make([]*domain.PaymentLink, 0, len(rows))
	for _, r := range rows {
		out = append(out, s.toDomain(r))
	}
	return out, nil
}

func (s *paymentLinkService) Get(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, domain.ErrPaymentLinkNotFound
	}
	row, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) Disable(ctx context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	if s.repo == nil {
		return nil, domain.ErrPaymentLinkNotFound
	}
	row, err := s.repo.UpdatePaymentLinkStatus(ctx, repository.UpdatePaymentLinkStatusParams{
		ID:         toPgUUID(id),
		MerchantID: toPgUUID(merchantID),
		Status:     "disabled",
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPaymentLinkNotFound
		}
		return nil, err
	}
	return s.toDomain(row), nil
}

func (s *paymentLinkService) toDomain(r repository.PaymentLink) *domain.PaymentLink {
	pl := &domain.PaymentLink{
		ID:             pgUUIDToUUID(r.ID),
		MerchantID:     pgUUIDToUUID(r.MerchantID),
		PublicID:       r.PublicID,
		Title:          r.Title,
		Description:    r.Description,
		AmountMinor:    r.AmountMinor,
		Currency:       r.Currency,
		AllowedMethods: r.AllowedMethods,
		LinkType:       r.LinkType,
		Status:         r.Status,
		Reference:      r.Reference,
		ImageURL:       r.ImageUrl,
		URL:            s.baseURL + "/pay/" + r.PublicID,
		CreatedAt:      r.CreatedAt.Time,
		UpdatedAt:      r.UpdatedAt.Time,
	}
	if r.ExpiresAt.Valid {
		t := r.ExpiresAt.Time
		pl.ExpiresAt = &t
	}
	if pl.AllowedMethods == nil {
		pl.AllowedMethods = []string{}
	}
	return pl
}

// generatePublicID returns a URL-safe base62-ish slug (hex is fine and avoids an
// alphabet dependency) with enough entropy to be unguessable.
func generatePublicID() (string, error) {
	buf := make([]byte, 12)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	const hexdigits = "0123456789abcdefghijklmnopqrstuvwxyz"
	out := make([]byte, len(buf))
	for i, b := range buf {
		out[i] = hexdigits[int(b)%len(hexdigits)]
	}
	return "pl_" + string(out), nil
}

// strings_TrimRightSlash trims a single trailing slash so URL joins don't double.
func strings_TrimRightSlash(s string) string {
	if len(s) > 0 && s[len(s)-1] == '/' {
		return s[:len(s)-1]
	}
	return s
}

var _ = time.Now // retain time import if unused after edits
```

> After writing, remove the two "retain import" guard lines (`var _ = ...`) if the imports are genuinely used (they are: `pgtype`, `time` via `pgtype.Timestamptz{Time:...}`). Run `goimports`/`gofmt`; delete any guard that causes an "unused" contradiction. Do NOT ship `var _ = time.Now` if `time` is already referenced — it will still compile but is dead code; remove it.

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/service/ -run 'TestCreatePaymentLink|TestGetPaymentLink' -v`
Expected: PASS. Then `go test ./internal/service/...` for regressions.

- [ ] **Step 6: Commit**

```bash
git add internal/service/payment_link_service.go internal/service/payment_link_service_test.go \
        internal/domain/errors.go internal/middleware/middleware.go
git commit -m "feat(service): payment link create/list/get/disable (merchant-scoped)"
```

---

### Task 4: Combined `MerchantAuth` middleware (session cookie OR API key)

**Files:**
- Create: `internal/middleware/merchant_auth.go`
- Test: `internal/middleware/merchant_auth_test.go`

**Interfaces:**
- Consumes: `session.Manager`, `middleware.MerchantResolver` (existing), `middleware.SessionCookieName`, `middleware.HashAPIKey`
- Produces: `middleware.MerchantAuth(mgr *session.Manager, resolver MerchantResolver) fiber.Handler` — resolves a merchant from the `pc_session` cookie if present+valid (also sets `LocalUserID`), otherwise from an API key; sets `LocalMerchantID`; 401 on neither.

- [ ] **Step 1: Write the failing test**

Create `internal/middleware/merchant_auth_test.go`:
```go
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

type stubResolver struct {
	merchant *domain.Merchant
}

func (s stubResolver) ResolveByAPIKeyHash(_ context.Context, hash string) (*domain.Merchant, error) {
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
	app := mkApp(mgr, stubResolver{})

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
	app := mkApp(mgr, stubResolver{merchant: m})

	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Authorization", "Bearer good-key")
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fatalf("api key -> %d want 200", resp.StatusCode)
	}
}

func TestMerchantAuthRejectsNeither(t *testing.T) {
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	app := mkApp(mgr, stubResolver{})
	resp, _ := app.Test(httptest.NewRequest("GET", "/x", nil))
	if resp.StatusCode != 401 {
		t.Fatalf("no creds -> %d want 401", resp.StatusCode)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/middleware/ -run TestMerchantAuth -v`
Expected: FAIL — `MerchantAuth` undefined

- [ ] **Step 3: Write the implementation**

Create `internal/middleware/merchant_auth.go`:
```go
package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// MerchantAuth authenticates a merchant from EITHER a dashboard session cookie
// (pc_session) OR an API key, resolving the same merchant context so a route can
// serve both the cookie-based dashboard and API-key clients. The session cookie
// is tried first (dashboard is the common caller); a valid cookie also sets
// LocalUserID. On neither credential it returns 401 without leaking which was
// missing.
func MerchantAuth(mgr *session.Manager, resolver MerchantResolver) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 1. Session cookie.
		if raw := c.Cookies(SessionCookieName); raw != "" {
			if claims, err := mgr.Verify(raw); err == nil {
				c.Locals(LocalMerchantID, claims.MerchantID)
				c.Locals(LocalUserID, claims.UserID)
				return c.Next()
			}
		}
		// 2. API key.
		if rawKey := extractAPIKey(c); rawKey != "" {
			merchant, err := resolver.ResolveByAPIKeyHash(c.UserContext(), HashAPIKey(rawKey))
			if err == nil && merchant != nil {
				c.Locals(LocalMerchant, merchant)
				c.Locals(LocalMerchantID, merchant.ID)
				return c.Next()
			}
			if err != nil && !errors.Is(err, domain.ErrUnauthorized) {
				return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
			}
		}
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "not authenticated")
	}
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/middleware/ -run TestMerchantAuth -v`
Expected: PASS (all three). Then `go test ./internal/middleware/...`.

- [ ] **Step 5: Commit**

```bash
git add internal/middleware/merchant_auth.go internal/middleware/merchant_auth_test.go
git commit -m "feat(middleware): combined session-or-apikey merchant auth"
```

---

### Task 5: `PaymentLinkHandler` + config + routes + main wiring

**Files:**
- Create: `internal/handler/payment_link_handler.go`
- Test: `internal/handler/payment_link_handler_test.go`
- Modify: `internal/config/config.go` (+ `PublicBaseURL`), `.env.example`
- Modify: `internal/handler/handler.go` (add `PaymentLink *PaymentLinkHandler` + `WithPaymentLinks` builder)
- Modify: `internal/router/router.go` (add `merchantAuth fiber.Handler` arg; mount `/payment-links`)
- Modify: `cmd/server/main.go` (build service + handler + merchantAuth; update `router.Setup` call)

**Interfaces:**
- Consumes: `service.PaymentLinkService`, `middleware.MerchantIDFromCtx`, `middleware.UserIDFromCtx`, `paginate`, `validationErrorResponse`
- Produces: `handler.NewPaymentLinkHandler(svc service.PaymentLinkService, log zerolog.Logger) *PaymentLinkHandler` with methods `Create`, `List`, `Get`, `Disable`; `Config.PublicBaseURL string` (default = `OAuthRedirectBase` value, else `http://localhost:3000`)

- [ ] **Step 1: Add config `PublicBaseURL`**

In `internal/config/config.go`: add field near `OAuthRedirectBase`:
```go
	// PublicBaseURL is the origin used to build shareable payment-link / checkout
	// URLs (<base>/pay/<public_id>). Defaults to OAuthRedirectBase when empty.
	PublicBaseURL string `mapstructure:"PUBLIC_BASE_URL"`
```
Add `"PUBLIC_BASE_URL"` to `envKeys`. In `Load()`, after unmarshal + the PORT block, add a fallback:
```go
	if strings.TrimSpace(c.PublicBaseURL) == "" {
		if strings.TrimSpace(c.OAuthRedirectBase) != "" {
			c.PublicBaseURL = c.OAuthRedirectBase
		} else {
			c.PublicBaseURL = "http://localhost:3000"
		}
	}
```
Append to `.env.example`:
```
# Origin for shareable payment-link / checkout URLs (<base>/pay/<id>). Defaults to OAUTH_REDIRECT_BASE.
PUBLIC_BASE_URL=http://localhost:3000
```

- [ ] **Step 2: Write the failing handler test**

Create `internal/handler/payment_link_handler_test.go`:
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

type fakeLinkSvc struct{ lastMerchant uuid.UUID }

func (f *fakeLinkSvc) Create(_ context.Context, merchantID uuid.UUID, _ *uuid.UUID, req domain.CreatePaymentLinkRequest) (*domain.PaymentLink, error) {
	f.lastMerchant = merchantID
	return &domain.PaymentLink{ID: uuid.New(), MerchantID: merchantID, PublicID: "pl_abc", Title: req.Title, AmountMinor: req.AmountMinor, Currency: "THB", Status: "active", URL: "https://pay.example/pay/pl_abc", AllowedMethods: []string{}}, nil
}
func (f *fakeLinkSvc) List(_ context.Context, merchantID uuid.UUID, _, _ int32) ([]*domain.PaymentLink, error) {
	return []*domain.PaymentLink{{ID: uuid.New(), MerchantID: merchantID, Title: "A", AllowedMethods: []string{}}}, nil
}
func (f *fakeLinkSvc) Get(_ context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{ID: id, MerchantID: merchantID, Title: "A", AllowedMethods: []string{}}, nil
}
func (f *fakeLinkSvc) Disable(_ context.Context, merchantID, id uuid.UUID) (*domain.PaymentLink, error) {
	return &domain.PaymentLink{ID: id, MerchantID: merchantID, Status: "disabled", AllowedMethods: []string{}}, nil
}

func TestCreateLinkUsesAuthedMerchantNotBody(t *testing.T) {
	svc := &fakeLinkSvc{}
	h := NewPaymentLinkHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	authed := uuid.New()
	app.Post("/v1/payment-links", withMerchant(authed), h.Create)

	// body tries to inject a different merchant_id — must be ignored.
	body := `{"title":"Coffee","amount_minor":5000,"merchant_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest("POST", "/v1/payment-links", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status=%d want 201 (%s)", resp.StatusCode, b)
	}
	if svc.lastMerchant != authed {
		t.Fatalf("service got merchant %v, want authed %v (body must not override)", svc.lastMerchant, authed)
	}
}

func TestCreateLinkValidationRejectsZeroAmount(t *testing.T) {
	h := NewPaymentLinkHandler(&fakeLinkSvc{}, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/payment-links", withMerchant(uuid.New()), h.Create)

	req := httptest.NewRequest("POST", "/v1/payment-links", strings.NewReader(`{"title":"X","amount_minor":0}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("zero amount -> %d want 400", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var env domain.APIResponse
	_ = json.Unmarshal(body, &env)
	if env.Code != "VALIDATION_ERROR" {
		t.Fatalf("code=%q want VALIDATION_ERROR", env.Code)
	}
}
```

> Note: `withMerchant` already exists in `merchant_handler_test.go` (same package) — reuse it, do NOT redefine.

- [ ] **Step 3: Run test to verify it fails**

Run: `go test ./internal/handler/ -run TestCreateLink -v`
Expected: FAIL — `NewPaymentLinkHandler` undefined

- [ ] **Step 4: Write the handler**

Create `internal/handler/payment_link_handler.go`:
```go
package handler

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/service"
)

type PaymentLinkHandler struct {
	svc      service.PaymentLinkService
	validate *validator.Validate
	log      zerolog.Logger
}

func NewPaymentLinkHandler(svc service.PaymentLinkService, log zerolog.Logger) *PaymentLinkHandler {
	return &PaymentLinkHandler{svc: svc, validate: validator.New(), log: log}
}

// Create godoc
// @Router /v1/payment-links [post]
func (h *PaymentLinkHandler) Create(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	var req domain.CreatePaymentLinkRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	if err := h.validate.Struct(req); err != nil {
		return validationErrorResponse(c, err)
	}
	var createdBy *uuid.UUID
	if uid, ok := middleware.UserIDFromCtx(c); ok {
		createdBy = &uid
	}
	pl, err := h.svc.Create(c.Context(), mid, createdBy, req)
	if err != nil {
		return err
	}
	return domain.Created(c, pl)
}

// List godoc
// @Router /v1/payment-links [get]
func (h *PaymentLinkHandler) List(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	limit, offset := paginate(c)
	items, err := h.svc.List(c.Context(), mid, limit, offset)
	if err != nil {
		return err
	}
	return domain.Success(c, items)
}

// Get godoc
// @Router /v1/payment-links/{id} [get]
func (h *PaymentLinkHandler) Get(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid link id")
	}
	pl, err := h.svc.Get(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, pl)
}

// Disable godoc
// @Router /v1/payment-links/{id} [patch]
func (h *PaymentLinkHandler) Disable(c *fiber.Ctx) error {
	mid, ok := middleware.MerchantIDFromCtx(c)
	if !ok {
		return domain.Error(c, fiber.StatusUnauthorized, "UNAUTHORIZED", "merchant not authenticated")
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_ID", "invalid link id")
	}
	pl, err := h.svc.Disable(c.Context(), mid, id)
	if err != nil {
		return err
	}
	return domain.Success(c, pl)
}
```

- [ ] **Step 5: Run handler test to verify it passes**

Run: `go test ./internal/handler/ -run TestCreateLink -v`
Expected: PASS

- [ ] **Step 6: Wire aggregate + router + main**

`internal/handler/handler.go`: add `PaymentLink *PaymentLinkHandler` to `Handlers`, and a builder:
```go
// WithPaymentLinks attaches the payment-link handler.
func (h *Handlers) WithPaymentLinks(p *PaymentLinkHandler) *Handlers {
	h.PaymentLink = p
	return h
}
```

`internal/router/router.go`: add a `merchantAuth fiber.Handler` parameter to `Setup` (place it right after `sessionAuth`), update the doc comment, and mount the group after the `/auth` block:
```go
	// Payment links (dashboard + API). merchantAuth accepts a session cookie OR
	// an API key. Reads/updates are merchant-scoped in the service (IDOR-safe).
	if h.PaymentLink != nil {
		links := v1.Group("/payment-links", merchantAuth)
		links.Post("/", h.PaymentLink.Create)
		links.Get("/", h.PaymentLink.List)
		links.Get("/:id", h.PaymentLink.Get)
		links.Patch("/:id", h.PaymentLink.Disable)
	}
```

`cmd/server/main.go`: near where `sessions`/`sessionAuth` are built (Phase 1), add:
```go
	merchantAuth := middleware.MerchantAuth(sessions, merchantSvc)
	linkSvc := service.NewPaymentLinkService(repo, cfg.PublicBaseURL, logger)
```
Attach the handler where `h` is assembled:
```go
	h = h.WithPaymentLinks(handler.NewPaymentLinkHandler(linkSvc, logger))
```
Update the `router.Setup(...)` call to pass `merchantAuth` in its new position (immediately after `sessionAuth`):
```go
	router.Setup(app, h, auth, sessionAuth, merchantAuth, adminAuth, rateLimit, signupLimit, publicMetrics, cfg.WebDir, cfg.SandboxMode)
```

- [ ] **Step 7: Build + full suite + manual smoke**

Run: `go build ./... && go test ./...` → expect PASS everywhere (fix the router call-site arg count until green).

Manual smoke (sandbox; reuses Phase 1 dev-login):
```bash
SANDBOX_MODE=true MIGRATE_ON_BOOT=true JWT_SECRET=dev-secret-dev-secret-dev-secret-xx \
  PUBLIC_BASE_URL=http://localhost:3000 make run    # (ensure DATABASE_URL is set/.env present)
# other shell:
curl -s -X POST localhost:8080/v1/auth/dev-login -c /tmp/pc.cookies >/dev/null
curl -s -X POST localhost:8080/v1/payment-links -b /tmp/pc.cookies \
  -H 'Content-Type: application/json' \
  -d '{"title":"Coffee","amount_minor":5000,"allowed_methods":["card","promptpay"]}'
curl -s localhost:8080/v1/payment-links -b /tmp/pc.cookies
```
Expected: create returns 201 with a `url` like `http://localhost:3000/pay/pl_...`; list returns the link. (If starting the server is impractical, skip and say so — automated tests are the gate.)

- [ ] **Step 8: Commit**

```bash
git add internal/handler/payment_link_handler.go internal/handler/payment_link_handler_test.go \
        internal/handler/handler.go internal/router/router.go cmd/server/main.go \
        internal/config/config.go .env.example
git commit -m "feat(api): payment-link routes behind combined merchant auth"
```

---

### Task 6: Dashboard `/links` — list + create

**Files:**
- Create: `web-app/lib/format.ts` (money/format helpers)
- Create: `web-app/app/links/page.tsx` (server component: list)
- Create: `web-app/components/CreateLinkForm.tsx` (client component: create modal/inline form)
- Modify: `web-app/app/page.tsx` (add a link to `/links` in the dashboard shell)

**Interfaces:**
- Consumes: `GET /api/payment-links`, `POST /api/payment-links` (via the proxy), `serverGet` (Phase 1).

- [ ] **Step 1: Money formatting helper**

Create `web-app/lib/format.ts`:
```ts
// formatMoney renders integer minor units (สตางค์) as a THB amount string.
export function formatMoney(amountMinor: number, currency = "THB"): string {
  const major = amountMinor / 100;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(major);
}
```

- [ ] **Step 2: List page (server component)**

Create `web-app/app/links/page.tsx`:
```tsx
import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import CreateLinkForm from "@/components/CreateLinkForm";

type PaymentLink = {
  id: string;
  public_id: string;
  title: string;
  amount_minor: number;
  currency: string;
  status: string;
  url: string;
};

export default async function LinksPage() {
  const res = await serverGet("/payment-links");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`payment-links failed: ${res.status}`);
  const env = await res.json();
  const links: PaymentLink[] = env.data ?? [];

  return (
    <main className="min-h-screen p-8 max-w-3xl mx-auto">
      <header className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold">Payment Links</h1>
        <a href="/" className="text-sm text-paycore-muted hover:text-paycore-text">← Dashboard</a>
      </header>

      <CreateLinkForm />

      <section className="mt-8 space-y-3">
        {links.length === 0 && (
          <p className="text-paycore-muted text-sm">ยังไม่มีลิงก์ — สร้างอันแรกด้านบน</p>
        )}
        {links.map((l) => (
          <div key={l.id} className="rounded-xl2 bg-paycore-surface p-4 flex items-center justify-between">
            <div>
              <a href={`/links/${l.id}`} className="font-medium hover:underline">{l.title}</a>
              <p className="text-paycore-muted text-sm">{formatMoney(l.amount_minor, l.currency)} · {l.status}</p>
            </div>
            <a href={l.url} target="_blank" className="text-paycore-primary text-sm hover:underline">เปิดลิงก์ ↗</a>
          </div>
        ))}
      </section>
    </main>
  );
}
```

- [ ] **Step 3: Create form (client component)**

Create `web-app/components/CreateLinkForm.tsx`:
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

const METHODS = ["card", "promptpay", "mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"];

export default function CreateLinkForm() {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [amount, setAmount] = useState("");
  const [methods, setMethods] = useState<string[]>([]);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");

  function toggle(m: string) {
    setMethods((cur) => (cur.includes(m) ? cur.filter((x) => x !== m) : [...cur, m]));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setErr("");
    const baht = Number(amount);
    if (!title.trim() || !(baht > 0)) {
      setErr("กรอกชื่อและจำนวนเงินให้ถูกต้อง");
      return;
    }
    setBusy(true);
    const res = await fetch("/api/payment-links", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title: title.trim(),
        amount_minor: Math.round(baht * 100),
        allowed_methods: methods,
      }),
    });
    setBusy(false);
    if (res.ok) {
      setTitle(""); setAmount(""); setMethods([]); setOpen(false);
      router.refresh();
    } else {
      const body = await res.json().catch(() => null);
      setErr(body?.message ?? "สร้างลิงก์ไม่สำเร็จ");
    }
  }

  if (!open) {
    return (
      <button onClick={() => setOpen(true)} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2">
        + สร้าง Payment Link
      </button>
    );
  }

  return (
    <form onSubmit={submit} className="rounded-xl2 bg-paycore-surface p-5 space-y-4">
      <div>
        <label className="block text-sm text-paycore-muted mb-1">ชื่อรายการ</label>
        <input value={title} onChange={(e) => setTitle(e.target.value)} className="w-full rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" placeholder="เช่น กาแฟลาเต้" />
      </div>
      <div>
        <label className="block text-sm text-paycore-muted mb-1">จำนวนเงิน (บาท)</label>
        <input value={amount} onChange={(e) => setAmount(e.target.value)} inputMode="decimal" className="w-full rounded-lg bg-paycore-bg border border-white/10 px-3 py-2" placeholder="50.00" />
      </div>
      <div>
        <label className="block text-sm text-paycore-muted mb-2">ช่องทางจ่าย (ว่าง = ทุกช่องทาง)</label>
        <div className="flex flex-wrap gap-2">
          {METHODS.map((m) => (
            <button type="button" key={m} onClick={() => toggle(m)}
              className={`rounded-full px-3 py-1 text-sm border ${methods.includes(m) ? "bg-paycore-primary border-paycore-primary text-white" : "border-white/15 text-paycore-muted"}`}>
              {m}
            </button>
          ))}
        </div>
      </div>
      {err && <p className="text-red-400 text-sm">{err}</p>}
      <div className="flex gap-2">
        <button disabled={busy} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
          {busy ? "กำลังสร้าง…" : "สร้างลิงก์"}
        </button>
        <button type="button" onClick={() => setOpen(false)} className="rounded-lg px-4 py-2 text-paycore-muted hover:text-paycore-text">ยกเลิก</button>
      </div>
    </form>
  );
}
```

- [ ] **Step 4: Link the dashboard shell to /links**

In `web-app/app/page.tsx`, add a nav link into the dashboard (inside the `<section>` or header). Add this line after the merchant id `<p>`:
```tsx
        <a href="/links" className="inline-block mt-4 text-paycore-primary hover:underline">จัดการ Payment Links →</a>
```

- [ ] **Step 5: Build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/links` compiles (dynamic, uses cookies).

- [ ] **Step 6: Commit**

```bash
git add web-app/lib/format.ts web-app/app/links/page.tsx web-app/components/CreateLinkForm.tsx web-app/app/page.tsx
git commit -m "feat(web): payment links list + create form"
```

---

### Task 7: Dashboard `/links/[id]` — detail + copy URL + disable

**Files:**
- Create: `web-app/app/links/[id]/page.tsx` (server component: detail)
- Create: `web-app/components/LinkActions.tsx` (client: copy URL + disable)

**Interfaces:**
- Consumes: `GET /api/payment-links/:id`, `PATCH /api/payment-links/:id` (via proxy), `serverGet`.

> QR-of-URL is intentionally deferred to Phase 3 (checkout polish) to avoid adding a QR npm dependency now — this page ships copy-URL + open + disable, which fully delivers "create and share a link."

- [ ] **Step 1: Detail page (server component)**

Create `web-app/app/links/[id]/page.tsx`:
```tsx
import { redirect, notFound } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import LinkActions from "@/components/LinkActions";

type PaymentLink = {
  id: string;
  public_id: string;
  title: string;
  description?: string;
  amount_minor: number;
  currency: string;
  allowed_methods: string[];
  status: string;
  url: string;
};

export default async function LinkDetail({ params }: { params: { id: string } }) {
  const res = await serverGet(`/payment-links/${params.id}`);
  if (res.status === 401) redirect("/login");
  if (res.status === 404) notFound();
  if (!res.ok) throw new Error(`link failed: ${res.status}`);
  const env = await res.json();
  const link: PaymentLink = env.data;

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <a href="/links" className="text-sm text-paycore-muted hover:text-paycore-text">← Payment Links</a>
      <div className="rounded-xl2 bg-paycore-surface p-6 mt-4">
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-semibold">{link.title}</h1>
            <p className="text-paycore-muted mt-1">{formatMoney(link.amount_minor, link.currency)}</p>
          </div>
          <span className="rounded-full px-3 py-1 text-xs bg-paycore-bg border border-white/10">{link.status}</span>
        </div>

        {link.description && <p className="mt-4 text-sm text-paycore-muted">{link.description}</p>}

        <div className="mt-4 flex flex-wrap gap-2">
          {(link.allowed_methods.length ? link.allowed_methods : ["ทุกช่องทาง"]).map((m) => (
            <span key={m} className="rounded-full px-3 py-1 text-xs border border-white/15 text-paycore-muted">{m}</span>
          ))}
        </div>

        <LinkActions id={link.id} url={link.url} status={link.status} />
      </div>
    </main>
  );
}
```

- [ ] **Step 2: Actions (client component)**

Create `web-app/components/LinkActions.tsx`:
```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function LinkActions({ id, url, status }: { id: string; url: string; status: string }) {
  const router = useRouter();
  const [copied, setCopied] = useState(false);
  const [busy, setBusy] = useState(false);

  async function copy() {
    await navigator.clipboard.writeText(url);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  }

  async function disable() {
    setBusy(true);
    const res = await fetch(`/api/payment-links/${id}`, { method: "PATCH" });
    setBusy(false);
    if (res.ok) router.refresh();
  }

  return (
    <div className="mt-6 space-y-3">
      <div className="flex items-center gap-2">
        <input readOnly value={url} className="flex-1 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2 text-sm" />
        <button onClick={copy} className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm">
          {copied ? "คัดลอกแล้ว ✓" : "คัดลอก"}
        </button>
      </div>
      <div className="flex gap-2">
        <a href={url} target="_blank" className="rounded-lg border border-white/15 px-4 py-2 text-sm hover:bg-white/5">เปิดหน้าจ่าย ↗</a>
        {status === "active" && (
          <button onClick={disable} disabled={busy} className="rounded-lg border border-red-500/40 text-red-400 px-4 py-2 text-sm hover:bg-red-500/10 disabled:opacity-60">
            {busy ? "…" : "ปิดลิงก์"}
          </button>
        )}
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/links/[id]` compiles.

- [ ] **Step 4: Manual end-to-end smoke (sandbox)**

```bash
# backend: SANDBOX_MODE=true MIGRATE_ON_BOOT=true JWT_SECRET=... PUBLIC_BASE_URL=http://localhost:3000 make run
# frontend: cd web-app && BACKEND_URL=http://localhost:8080 npm run dev
```
Browser: login (dev) → Dashboard → "จัดการ Payment Links" → create a link → it appears in the list → open detail → copy URL → disable → status flips to `disabled`.

- [ ] **Step 5: Commit**

```bash
git add web-app/app/links/ web-app/components/LinkActions.tsx
git commit -m "feat(web): payment link detail + copy/disable actions"
```

---

## Self-Review

**Spec coverage (design spec §3.2 payment_links, §5.2 payment-link endpoints, §7 dashboard /links + /links/[id]):**
- payment_links table → Task 1 ✓
- CRUD endpoints (POST/GET list/GET id/PATCH disable) → Task 5 ✓
- merchant-scoped reads/updates (IDOR-safe) → Task 1 (queries WHERE merchant_id) + Task 3 (service) + tests ✓
- combined dashboard(cookie)+API(key) auth so the frontend can call it → Task 4 ✓
- shareable public URL (`/pay/<public_id>`) → Task 3 (URL build) + Task 5 (PublicBaseURL config) ✓
- dashboard list + create → Task 6 ✓
- dashboard detail + copy/disable → Task 7 ✓
- QR-of-URL → **deferred to Phase 3** (documented in Task 7) to avoid a QR npm dep now.

**Placeholder scan:** none — every code step is complete. The two `var _ = ...` import guards in Task 3 are explicitly flagged for removal in the note under Step 4.

**Type consistency:** `PaymentLinkService` (Create/List/Get/Disable) defined in Task 3, consumed in Task 5 handler + tests. `domain.PaymentLink`/`CreatePaymentLinkRequest` defined Task 2, used Tasks 3, 5, 6, 7. `middleware.MerchantAuth` defined Task 4, wired Task 5. sqlc `*Params` names (`CreatePaymentLinkParams` with `ImageUrl`, `AllowedMethods []string`, `ExpiresAt`/`CreatedBy` pgtype) assumed from Task 1 query names — **verify against generated code after `make sqlc`** (Task 1 Step 6 records them). Frontend `serverGet` (Phase 1) reused unchanged; `formatMoney` defined Task 6, used Tasks 6 & 7.

**Cross-phase notes for execution:**
1. `router.Setup` signature grows again (add `merchantAuth`). Single call site in `cmd/server/main.go`.
2. This phase mounts payment-links under the new combined auth but does NOT retrofit the existing `/me`,`/stats` routes onto it — those remain API-key-only until a later phase needs them in the dashboard. Not a regression.
3. `ErrPaymentLinkNotFound` must be mapped to HTTP 404 in the same place `ErrMerchantNotFound` is (Task 3 Step 1) — verify that mapping exists, or the service error will surface as 500.
