# Phase 5 — Merchant Dashboard polish Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the cookie-authed Next.js dashboard into a real merchant operations console — home-page stats, a transactions list + detail with a working refund action, and a settings page (rotate API key / set webhook / view business profile) — by retrofitting the existing merchant-scoped API routes to accept the dashboard session cookie and building the four dashboard pages on top of the handlers/services that already exist.

**Architecture:** No new backend feature logic. The one enabling change is an **auth retrofit** in `internal/router/router.go`: the dashboard-readable routes (`/me`, `/me/rotate-key`, `/me/webhook`, `/stats`, `/settlements`, `/disputes`, and `GET /payments`, `GET /payments/:id`, `POST /payments/:id/refund`) switch from API-key-only `auth` (`middleware.APIKeyAuth`) to the combined session-cookie-OR-API-key `merchantAuth` (`middleware.MerchantAuth`, already built and wired in `cmd/server/main.go:198`). `merchantAuth` sets `LocalMerchantID` identically to `auth` (verified: `internal/middleware/merchant_auth.go:23,32`), so every handler and the per-merchant rate limiter (keys on `MerchantIDFromCtx`, `internal/middleware/middleware.go:66`) are unchanged. Money-moving / lifecycle routes (`POST /payments`, `capture`, `void`, `3ds/return`, `POST /payments/:id/disputes`) stay API-key-only. The frontend adds four pages under `web-app/app/` reusing `serverGet` (cookie-forwarded reads) and client `fetch('/api/...')` (mutations through the Next rewrite proxy).

**Transactions scope — DECISION: Option A (card payments only).** `GET /payments` returns only rows from the `payments` table, which holds **card** payments. PromptPay lives in `qr_payments` (separate table/service) and Phase-4 wallet mocks have **no** `payments` row (they live only on `checkout_sessions`). A unified transactions endpoint aggregating three tables with cross-table pagination is a new query + service + handler — too large for a "polish" phase and out of its intent. Option A ships honest, working value: the Transactions page lists card payments (list + detail + refund) and the home stats come from `GET /stats`. The UI copy must say "card payments" and NOT imply it shows PromptPay or wallet transactions. A unified `GET /transactions` endpoint is recorded as an explicit follow-up (see README step in Task 7).

**Refund auth / CSRF — DECISION.** `POST /payments/:id/refund` moves to `merchantAuth`, so a `pc_session` cookie can trigger a refund. This is acceptable because the session cookie is `SameSite=Lax` (verified `internal/middleware/session_auth.go:54` — `fiber.CookieSameSiteLaxMode`), so a cross-site `POST` does not carry it (CSRF mitigation), and the route keeps its `middleware.RequireIdempotencyKey()` gate and the per-merchant rate limiter. The frontend sends a generated `Idempotency-Key` header on every refund. Refund validation errors (`ErrRefundExceeds` → 422, `ErrInvalidState` → 409, `ErrCardDeclined` → 402/422) already flow through the central `ErrorHandler` into the response envelope; the UI surfaces `envelope.message`.

**Tech Stack:** Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 · shopspring/decimal · Next.js 14 App Router + TypeScript + Tailwind. **No new Go or npm dependency.**

## Global Constraints

- Module `github.com/yourco/payment-gateway`. Branch `feat/dashboard-polish` (already created). **NO new Go dependency; NO new npm dependency.**
- **No migration, no new sqlc query, no `make sqlc`.** Phase 5 reuses the existing `payments`, `merchants`, `settlements` tables through handlers/services that already exist (`MerchantHandler.{Me,Stats,Settlements,RotateKey,SetWebhook}`, `PaymentHandler.{List,Get,Refund}`). If you think a schema change is needed, STOP — re-read the Architecture note; Option A is deliberate.
- **Dashboard routes are retrofitted to `merchantAuth` (session OR key).** Money-moving routes (`POST /payments`, `capture`, `void`, `3ds/return`, `POST /payments/:id/disputes`) stay on API-key-only `auth`.
- **Refund stays `Idempotency-Key`-gated and relies on `SameSite=Lax` for CSRF.** Do not add a new CSRF token; do not remove `RequireIdempotencyKey`.
- **Money units differ by table.** `domain.Payment.{Amount,CapturedAmount,RefundedAmount}` are `decimal.Decimal` in **major units** (baht) and serialize as a **quoted JSON string** (shopspring default `MarshalJSONWithoutQuotes = false`, verified) — format with the new `formatDecimalMoney(major: string)`. `domain.MerchantStats.VolumeMinor` and `payment_links.amount_minor` are **minor units** (satang, integer) — format with the existing `formatMoney(minor: number)`. Never mix the two.
- **Merchant scoping comes from the auth context** (`middleware.MerchantIDFromCtx`), never the request body — unchanged; every reused handler already does this.
- Response envelope: `domain.APIResponse` via `domain.Success` / `domain.Created` / `domain.Error` — `{ success, code, message, data, request_id, timestamp }`.
- Frontend reads use `serverGet(path)` (server component, forwards `pc_session` cookie); mutations use client `fetch('/api/<path>', ...)` which the `next.config.js` rewrite proxies to `${BACKEND_URL}/v1/<path>` (browser attaches the cookie same-origin). Reuse Tailwind `paycore-*` tokens.
- TDD for Go: write the failing test, watch it fail, implement minimally, watch it pass, commit. Complete code in every step. Frequent commits. Frontend has no test runner (no `test` script in `web-app/package.json`) — the frontend gate is `cd web-app && npm run build` plus the documented manual smoke.
- Run set (Go): `go build ./... && go test ./...`. Run set (web): `cd web-app && npm run build`.

---

## File map

| File | Change | Responsibility |
|------|--------|----------------|
| `internal/router/router.go` | modify | Retrofit dashboard-read + refund routes to `merchantAuth`; keep money-moving routes on `auth` (split the `/payments` group into a read group and a write group) |
| `internal/router/router_test.go` | create | Prove the retrofit: a `pc_session` cookie now reaches `/me`, `/stats`, `GET /payments`, `POST /payments/:id/refund`; cookie-only is still rejected on `POST /payments` (Create stays API-key-only); no-credential is 401 |
| `web-app/lib/format.ts` | modify | Add `formatDecimalMoney(major, currency)` for major-unit decimal strings (keep `formatMoney` for minor units) |
| `web-app/components/DashboardNav.tsx` | create | Shared top nav (Dashboard / Transactions / Links / Settings / logout) reused by dashboard pages |
| `web-app/app/page.tsx` | modify | Home: KPI stat tiles from `/stats` + greeting from `/auth/me` + nav |
| `web-app/app/transactions/page.tsx` | create | Card-payments list (paginated) from `/payments`, rows link to detail |
| `web-app/app/transactions/[id]/page.tsx` | create | Payment detail from `/payments/:id` + refund panel |
| `web-app/components/RefundForm.tsx` | create | Client refund form: `fetch('/api/payments/:id/refund')` with generated `Idempotency-Key`, surfaces envelope errors |
| `web-app/app/settings/page.tsx` | create | Business profile from `/me` + rotate-key + webhook panels |
| `web-app/components/RotateKeyButton.tsx` | create | Client: `POST /api/me/rotate-key`, reveals the new key once |
| `web-app/components/WebhookForm.tsx` | create | Client: `PUT /api/me/webhook`, reveals the signing secret once |
| `web-app/middleware.ts` | create | Cheap route guard: redirect dashboard paths to `/login` when the `pc_session` cookie is absent |
| `README.md` | modify | Phase 5 note + the Option-A transactions-scope follow-up |

No changes to `cmd/server/main.go` (`merchantAuth` is already constructed at line 198 and passed to `router.Setup` at line 252), migrations, sqlc, services, domain types, or handlers.

---

### Task 1: Auth retrofit — dashboard routes accept the session cookie

**Files:**
- Modify: `internal/router/router.go:57-62` (the `/me` group) and `:107-122` (the `/payments` group)
- Create: `internal/router/router_test.go`

**Interfaces:**
- Consumes: `router.Setup(app, h, auth, sessionAuth, merchantAuth, adminAuth, rateLimit, signupLimit, checkoutLimit, metrics, webDir, sandbox)` (unchanged signature); `middleware.MerchantAuth(mgr *session.Manager, resolver MerchantResolver) fiber.Handler`; `middleware.APIKeyAuth(resolver) fiber.Handler`; `middleware.SessionAuth(mgr)`; `middleware.AdminAuth(key)`; `middleware.RequireIdempotencyKey()`; `session.NewManager(secret string, ttl time.Duration)`; `mgr.Issue(session.Claims{UserID, MerchantID, Email}) (string, error)`; `middleware.SessionCookieName`.
- Produces: no signature change. After this task, `GET /me`, `POST /me/rotate-key`, `PUT /me/webhook`, `GET /stats`, `GET /settlements`, `GET /disputes`, `GET /payments`, `GET /payments/:id`, `GET /payments/:id/disputes`, and `POST /payments/:id/refund` are gated by `merchantAuth`; `POST /payments`, `POST /payments/:id/capture`, `POST /payments/:id/void`, `POST /payments/:id/3ds/return`, `POST /payments/:id/disputes` stay gated by `auth`.

- [ ] **Step 1: Write the failing test**

Create `internal/router/router_test.go`. The test builds the real router via `router.Setup` with the real `merchantAuth`/`auth` middleware and handlers backed by tiny fakes, issues a session cookie, and asserts the cookie now reaches the retrofitted routes while `POST /payments` (Create) still rejects a cookie-only caller. (`router` imports `handler`, so this test lives in `package router` to avoid the import cycle that keeps the equivalent test out of `package handler`.)

```go
package router

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"

	"github.com/yourco/payment-gateway/internal/domain"
	"github.com/yourco/payment-gateway/internal/handler"
	"github.com/yourco/payment-gateway/internal/middleware"
	"github.com/yourco/payment-gateway/internal/pkg/session"
)

// --- fakes: minimal but complete implementations of the two services the
// retrofitted routes touch. Every other handler is left nil in Handlers; its
// routes register (method values bind a nil receiver fine) but are never called.

type fakeMerchantSvc struct{ mid uuid.UUID }

func (f fakeMerchantSvc) Onboard(context.Context, domain.CreateMerchantRequest) (*domain.MerchantCredential, error) {
	return nil, nil
}
func (f fakeMerchantSvc) Get(context.Context, uuid.UUID) (*domain.Merchant, error) { return nil, nil }
func (f fakeMerchantSvc) ResolveByAPIKeyHash(context.Context, string) (*domain.Merchant, error) {
	return nil, domain.ErrUnauthorized // no API key path exercised here
}
func (f fakeMerchantSvc) Profile(context.Context, uuid.UUID) (*domain.MerchantProfile, error) {
	return &domain.MerchantProfile{ID: f.mid, Name: "Acme", SettlementCurrency: "THB", Status: "active"}, nil
}
func (f fakeMerchantSvc) Stats(context.Context, uuid.UUID, time.Time, time.Time) (*domain.MerchantStats, error) {
	return &domain.MerchantStats{Count: 1, VolumeMinor: 10000}, nil
}
func (f fakeMerchantSvc) ListSettlements(context.Context, uuid.UUID, int) ([]*domain.Settlement, error) {
	return []*domain.Settlement{}, nil
}
func (f fakeMerchantSvc) RotateAPIKey(context.Context, uuid.UUID) (*domain.RotatedKey, error) {
	return &domain.RotatedKey{APIKey: "sk_new"}, nil
}
func (f fakeMerchantSvc) SetWebhook(context.Context, uuid.UUID, string) (*domain.WebhookConfig, error) {
	return &domain.WebhookConfig{WebhookURL: "https://x", SigningSecret: "whsec"}, nil
}

type fakePaymentSvc struct{ mid uuid.UUID }

func (f fakePaymentSvc) okPayment() *domain.Payment {
	return &domain.Payment{ID: uuid.New(), MerchantID: f.mid, Amount: decimal.RequireFromString("100.00"), Currency: "THB", Status: domain.StatusCaptured}
}
func (f fakePaymentSvc) Create(context.Context, string, domain.CreatePaymentRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Capture(context.Context, uuid.UUID, uuid.UUID, domain.CaptureRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Void(context.Context, uuid.UUID, uuid.UUID) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Refund(context.Context, uuid.UUID, uuid.UUID, domain.RefundRequest) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) Get(context.Context, uuid.UUID, uuid.UUID) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) List(context.Context, uuid.UUID, int32, int32) ([]*domain.Payment, error) {
	return []*domain.Payment{f.okPayment()}, nil
}
func (f fakePaymentSvc) HandleThreeDSResult(context.Context, uuid.UUID, uuid.UUID, bool) (*domain.Payment, error) {
	return f.okPayment(), nil
}
func (f fakePaymentSvc) VerifyThreeDSResult(uuid.UUID, string, string) bool { return true }

// buildApp wires the real router with real middleware and the fakes above.
func buildApp(t *testing.T, mid uuid.UUID) (*fiber.App, *session.Manager) {
	t.Helper()
	log := zerolog.Nop()
	mgr := session.NewManager("secret-secret-secret-secret-secret!", time.Hour)
	msvc := fakeMerchantSvc{mid: mid}
	psvc := fakePaymentSvc{mid: mid}

	h := &handler.Handlers{
		Merchant: handler.NewMerchantHandler(msvc, log),
		Payment:  handler.NewPaymentHandler(psvc, log),
		Health:   handler.NewHealthHandler(nil),
	}
	auth := middleware.APIKeyAuth(msvc)
	sessionAuth := middleware.SessionAuth(mgr)
	merchantAuth := middleware.MerchantAuth(mgr, msvc)
	adminAuth := middleware.AdminAuth("")

	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(log)})
	Setup(app, h, auth, sessionAuth, merchantAuth, adminAuth, nil, nil, nil, nil, "", false)
	return app, mgr
}

func TestRetrofit_CookieReachesDashboardRoutes(t *testing.T) {
	mid := uuid.New()
	app, mgr := buildApp(t, mid)
	tok, err := mgr.Issue(session.Claims{UserID: uuid.New(), MerchantID: mid, Email: "a@b.co"})
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}

	get := func(path string) int {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
		resp, _ := app.Test(req)
		return resp.StatusCode
	}
	for _, p := range []string{"/v1/me", "/v1/stats", "/v1/settlements", "/v1/payments", "/v1/payments/" + uuid.NewString()} {
		if code := get(p); code != 200 {
			t.Fatalf("GET %s via cookie -> %d want 200", p, code)
		}
	}

	// Refund via cookie: SameSite=Lax + Idempotency-Key gate; expect 200.
	rreq := httptest.NewRequest("POST", "/v1/payments/"+uuid.NewString()+"/refund", nil)
	rreq.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
	rreq.Header.Set("Idempotency-Key", "idem-1")
	rresp, _ := app.Test(rreq)
	if rresp.StatusCode != 200 {
		t.Fatalf("POST refund via cookie -> %d want 200", rresp.StatusCode)
	}
}

func TestRetrofit_CreateStaysApiKeyOnly(t *testing.T) {
	mid := uuid.New()
	app, mgr := buildApp(t, mid)
	tok, _ := mgr.Issue(session.Claims{UserID: uuid.New(), MerchantID: mid, Email: "a@b.co"})

	// A cookie-only caller must NOT be able to create a payment (Create stays on
	// API-key-only auth, whose resolver returns ErrUnauthorized for no key).
	req := httptest.NewRequest("POST", "/v1/payments", nil)
	req.Header.Set("Cookie", middleware.SessionCookieName+"="+tok)
	req.Header.Set("Idempotency-Key", "idem-2")
	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Fatalf("POST /payments via cookie -> %d want 401", resp.StatusCode)
	}
}

func TestRetrofit_NoCredentialsRejected(t *testing.T) {
	app, _ := buildApp(t, uuid.New())
	resp, _ := app.Test(httptest.NewRequest("GET", "/v1/me", nil))
	if resp.StatusCode != 401 {
		t.Fatalf("GET /me no creds -> %d want 401", resp.StatusCode)
	}
}
```

Note: delete the placeholder `cookieReq` stub above before running — it is only here to flag that no helper is needed; the tests build requests inline. (If you prefer, omit it entirely.)

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/router/ -run TestRetrofit -v`
Expected: FAIL — the tests compile but `TestRetrofit_CookieReachesDashboardRoutes` fails because the routes are still gated by API-key `auth`, so the cookie yields **401** on `/v1/me` (want 200).

- [ ] **Step 3: Retrofit the `/me` group to `merchantAuth`**

In `internal/router/router.go`, replace lines 55-62 (the "Merchant Dashboard (self-service)" block) with:

```go
	// Merchant Dashboard (self-service). Retrofitted to merchantAuth so BOTH the
	// cookie dashboard and API-key clients reach these merchant-scoped routes.
	// merchantAuth sets LocalMerchantID identically to auth, so the handlers are
	// unchanged. Everything resolves the merchant from the auth context; the
	// request body is never trusted for identity.
	v1.Get("/me", merchantAuth, h.Merchant.Me)
	v1.Post("/me/rotate-key", merchantAuth, h.Merchant.RotateKey)
	v1.Put("/me/webhook", merchantAuth, h.Merchant.SetWebhook)
	v1.Get("/stats", merchantAuth, h.Merchant.Stats)
	v1.Get("/settlements", merchantAuth, h.Merchant.Settlements)
	v1.Get("/disputes", merchantAuth, h.Dispute.ListByMerchant)
```

(`GET /merchants/:id` on line 53 stays on `auth` — it is not a dashboard route.)

- [ ] **Step 4: Split the `/payments` group — reads + refund on `merchantAuth`, money-moving on `auth`**

In `internal/router/router.go`, replace the whole payments block (lines 105-122) with:

```go
	// Payments — read + refund routes accept a session cookie OR an API key so
	// the cookie dashboard can list/inspect/refund. The rate limiter is mounted
	// AFTER auth so it keys per merchant (merchantAuth sets LocalMerchantID just
	// like auth). Refund is money-moving: it keeps its idempotency-key gate and
	// relies on the pc_session cookie being SameSite=Lax for CSRF protection.
	readPayments := v1.Group("/payments", merchantAuth)
	if rateLimit != nil {
		readPayments.Use(rateLimit)
	}
	readPayments.Get("/", h.Payment.List)
	readPayments.Get("/:id", h.Payment.Get)
	readPayments.Post("/:id/refund", middleware.RequireIdempotencyKey(), h.Payment.Refund)
	readPayments.Get("/:id/disputes", h.Dispute.ListByPayment)

	// Money-moving / lifecycle routes stay API-key-only (auth). A cookie-only
	// caller is rejected here even though it can reach the read routes above.
	writePayments := v1.Group("/payments", auth)
	if rateLimit != nil {
		writePayments.Use(rateLimit)
	}
	writePayments.Post("/", middleware.RequireIdempotencyKey(), h.Payment.Create)
	writePayments.Post("/:id/capture", middleware.RequireIdempotencyKey(), h.Payment.Capture)
	writePayments.Post("/:id/void", h.Payment.Void)
	writePayments.Post("/:id/3ds/return", h.Payment.ThreeDSReturn)

	// Chargebacks / disputes: opening one is a write (API key); listing is a read.
	writePayments.Post("/:id/disputes", h.Dispute.Open)
```

- [ ] **Step 5: Run the retrofit tests to verify they pass**

Run: `go test ./internal/router/ -run TestRetrofit -v`
Expected: PASS (all three tests).

- [ ] **Step 6: Run the full suite to confirm nothing regressed**

Run: `go build ./... && go test ./...`
Expected: build succeeds; all packages PASS (existing `internal/middleware` and `internal/handler` auth tests still green — handlers and the rate limiter are untouched).

- [ ] **Step 7: Commit**

```bash
git add internal/router/router.go internal/router/router_test.go
git commit -m "feat(router): retrofit dashboard-read + refund routes to merchantAuth (session OR key)"
```

---

### Task 2: Frontend — major-unit money formatter

**Files:**
- Modify: `web-app/lib/format.ts`

**Interfaces:**
- Consumes: nothing.
- Produces: `formatDecimalMoney(major: string | number, currency?: string): string` (major-unit decimal, e.g. Payment.Amount `"100.00"`). Existing `formatMoney(amountMinor: number, currency?)` is unchanged and stays the tool for minor-unit values (`VolumeMinor`, `amount_minor`).

- [ ] **Step 1: Add the formatter**

Replace the contents of `web-app/lib/format.ts` with:

```ts
// formatMoney renders integer minor units (สตางค์) as a THB amount string.
export function formatMoney(amountMinor: number, currency = "THB"): string {
  const major = amountMinor / 100;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(major);
}

// formatDecimalMoney renders a MAJOR-unit decimal (baht) as a THB amount string.
// domain.Payment.{Amount,CapturedAmount,RefundedAmount} are shopspring decimals
// serialized as quoted JSON strings (e.g. "100.00") — NOT minor units. Use this
// for payment amounts; use formatMoney for minor-unit values (stats volume,
// payment_links.amount_minor).
export function formatDecimalMoney(major: string | number, currency = "THB"): string {
  const n = typeof major === "number" ? major : parseFloat(major);
  const safe = Number.isFinite(n) ? n : 0;
  return new Intl.NumberFormat("th-TH", { style: "currency", currency }).format(safe);
}
```

- [ ] **Step 2: Verify it type-checks / builds**

Run: `cd web-app && npm run build`
Expected: build succeeds (no type errors; the new export is unused until Task 4 imports it — that is fine).

- [ ] **Step 3: Commit**

```bash
git add web-app/lib/format.ts
git commit -m "feat(web): add formatDecimalMoney for major-unit payment amounts"
```

---

### Task 3: Frontend — dashboard nav + home stats

**Files:**
- Create: `web-app/components/DashboardNav.tsx`
- Modify: `web-app/app/page.tsx`

**Interfaces:**
- Consumes: `serverGet('/auth/me')`, `serverGet('/stats')`; `formatMoney`; envelope `{ data }`. `MerchantStats` JSON: `{ count, volumeMinor, byStatus:{authorized,captured,refunded,failed}, successRate, refundRatio, chargebackRatio, from, to }` (ratios are fractions in [0,1]).
- Produces: `DashboardNav` (default export) — a server-safe presentational nav with links `/`, `/transactions`, `/settings` and a logout form posting to `/api/auth/logout`.

- [ ] **Step 1: Create the nav component**

Create `web-app/components/DashboardNav.tsx`:

```tsx
export default function DashboardNav({ active }: { active?: string }) {
  const items = [
    { href: "/", label: "หน้าหลัก" },
    { href: "/transactions", label: "ธุรกรรม" },
    { href: "/links", label: "ลิงก์ชำระเงิน" },
    { href: "/settings", label: "ตั้งค่า" },
  ];
  return (
    <header className="flex items-center justify-between mb-8">
      <nav className="flex items-center gap-4">
        <span className="text-xl font-semibold">PayCore</span>
        {items.map((it) => (
          <a
            key={it.href}
            href={it.href}
            className={
              "text-sm hover:text-paycore-text " +
              (active === it.href ? "text-paycore-text font-medium" : "text-paycore-muted")
            }
          >
            {it.label}
          </a>
        ))}
      </nav>
      <form action="/api/auth/logout" method="post">
        <button className="text-sm text-paycore-muted hover:text-paycore-text">ออกจากระบบ</button>
      </form>
    </header>
  );
}
```

- [ ] **Step 2: Rewrite the home page to show stats**

Replace `web-app/app/page.tsx` with:

```tsx
import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

type AuthMe = { email: string; name: string; merchant_name: string; merchant_id: string };
type Stats = {
  count: number;
  volumeMinor: number;
  byStatus: { authorized: number; captured: number; refunded: number; failed: number };
  successRate: number;
  refundRatio: number;
};

function pct(fraction: number): string {
  return `${(fraction * 100).toFixed(1)}%`;
}

export default async function DashboardHome() {
  const [meRes, statsRes] = await Promise.all([serverGet("/auth/me"), serverGet("/stats")]);
  if (meRes.status === 401 || statsRes.status === 401) redirect("/login");
  if (!meRes.ok) throw new Error(`auth/me failed: ${meRes.status}`);
  if (!statsRes.ok) throw new Error(`stats failed: ${statsRes.status}`);
  const me: AuthMe = (await meRes.json()).data;
  const s: Stats = (await statsRes.json()).data;

  const tiles = [
    { label: "ยอดรับชำระ (30 วัน)", value: formatMoney(s.volumeMinor) },
    { label: "จำนวนธุรกรรม", value: String(s.count) },
    { label: "อัตราสำเร็จ", value: pct(s.successRate) },
    { label: "อัตราคืนเงิน", value: pct(s.refundRatio) },
  ];

  return (
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <DashboardNav active="/" />
      <p className="text-paycore-muted text-sm mb-6">สวัสดี {me.name || me.email} · {me.merchant_name}</p>

      <section className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {tiles.map((t) => (
          <div key={t.label} className="rounded-xl2 bg-paycore-surface p-5">
            <p className="text-paycore-muted text-xs">{t.label}</p>
            <p className="text-2xl font-semibold mt-2">{t.value}</p>
          </div>
        ))}
      </section>

      <p className="text-paycore-muted text-xs mt-6">
        ตัวเลขคำนวณจากการชำระด้วยบัตร (card) ในรอบ 30 วันล่าสุด
      </p>

      <div className="mt-6 flex gap-4">
        <a href="/transactions" className="text-paycore-primary hover:underline">ดูธุรกรรมทั้งหมด →</a>
        <a href="/links" className="text-paycore-primary hover:underline">จัดการลิงก์ชำระเงิน →</a>
      </div>
    </main>
  );
}
```

- [ ] **Step 3: Verify the build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/` compiles as a dynamic route (uses `cookies()` via `serverGet`).

- [ ] **Step 4: Commit**

```bash
git add web-app/components/DashboardNav.tsx web-app/app/page.tsx
git commit -m "feat(web): home dashboard stat tiles + shared nav"
```

---

### Task 4: Frontend — transactions list

**Files:**
- Create: `web-app/app/transactions/page.tsx`

**Interfaces:**
- Consumes: `serverGet('/payments?limit=&offset=')`; `formatDecimalMoney`; `DashboardNav`. `domain.Payment` JSON: `{ id, merchant_id, amount:"<major>", captured_amount:"<major>", refunded_amount:"<major>", currency, status, card_brand?, card_last4?, reference?, created_at }` (amounts are quoted decimal strings).
- Produces: route `/transactions` with `?page=` pagination; each row links to `/transactions/<id>`.

- [ ] **Step 1: Create the list page**

Create `web-app/app/transactions/page.tsx`:

```tsx
import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";

type Payment = {
  id: string;
  amount: string;
  currency: string;
  status: string;
  card_brand?: string;
  card_last4?: string;
  reference?: string;
  created_at: string;
};

const PAGE_SIZE = 25;

const STATUS_LABEL: Record<string, string> = {
  authorized: "อนุมัติแล้ว",
  captured: "เรียกเก็บแล้ว",
  partial_refunded: "คืนบางส่วน",
  refunded: "คืนเงินแล้ว",
  voided: "ยกเลิก",
  failed: "ล้มเหลว",
  requires_action: "รอยืนยัน",
};

export default async function TransactionsPage({
  searchParams,
}: {
  searchParams: { page?: string };
}) {
  const page = Math.max(1, parseInt(searchParams.page ?? "1", 10) || 1);
  const offset = (page - 1) * PAGE_SIZE;
  const res = await serverGet(`/payments?limit=${PAGE_SIZE}&offset=${offset}`);
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`payments failed: ${res.status}`);
  const items: Payment[] = (await res.json()).data ?? [];

  return (
    <main className="min-h-screen p-8 max-w-4xl mx-auto">
      <DashboardNav active="/transactions" />
      <h1 className="text-xl font-semibold mb-1">ธุรกรรม (บัตร)</h1>
      <p className="text-paycore-muted text-xs mb-6">
        แสดงเฉพาะการชำระด้วยบัตร PromptPay และ e-wallet ยังไม่รวมในรายการนี้
      </p>

      <section className="space-y-2">
        {items.length === 0 && (
          <p className="text-paycore-muted text-sm">ยังไม่มีธุรกรรม</p>
        )}
        {items.map((p) => (
          <a
            key={p.id}
            href={`/transactions/${p.id}`}
            className="rounded-xl2 bg-paycore-surface p-4 flex items-center justify-between hover:bg-white/5"
          >
            <div>
              <p className="font-medium">{formatDecimalMoney(p.amount, p.currency)}</p>
              <p className="text-paycore-muted text-xs mt-1">
                {(p.card_brand || "card")}{p.card_last4 ? ` ····${p.card_last4}` : ""}
                {p.reference ? ` · ${p.reference}` : ""}
              </p>
            </div>
            <div className="text-right">
              <span className="rounded-full px-3 py-1 text-xs bg-paycore-bg border border-white/10">
                {STATUS_LABEL[p.status] ?? p.status}
              </span>
              <p className="text-paycore-muted text-xs mt-1">
                {new Date(p.created_at).toLocaleString("th-TH")}
              </p>
            </div>
          </a>
        ))}
      </section>

      <nav className="mt-6 flex items-center justify-between text-sm">
        {page > 1 ? (
          <a href={`/transactions?page=${page - 1}`} className="text-paycore-primary hover:underline">← ก่อนหน้า</a>
        ) : (
          <span />
        )}
        <span className="text-paycore-muted">หน้า {page}</span>
        {items.length === PAGE_SIZE ? (
          <a href={`/transactions?page=${page + 1}`} className="text-paycore-primary hover:underline">ถัดไป →</a>
        ) : (
          <span />
        )}
      </nav>
    </main>
  );
}
```

- [ ] **Step 2: Verify the build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/transactions` compiles.

- [ ] **Step 3: Commit**

```bash
git add web-app/app/transactions/page.tsx
git commit -m "feat(web): transactions list (card payments, paginated)"
```

---

### Task 5: Frontend — transaction detail + refund

**Files:**
- Create: `web-app/app/transactions/[id]/page.tsx`
- Create: `web-app/components/RefundForm.tsx`

**Interfaces:**
- Consumes: `serverGet('/payments/:id')`; `formatDecimalMoney`; `RefundRequest` body `{ amount: <number|string>, reason?: string }` (amount is a major-unit decimal); refund route `POST /api/payments/:id/refund` with header `Idempotency-Key`; envelope error `{ success:false, code, message }`.
- Produces: route `/transactions/[id]`; `RefundForm` client component (`{ paymentId, remaining, currency }`) that refunds and calls `router.refresh()` on success, surfacing `envelope.message` on failure.

- [ ] **Step 1: Create the refund form (client component)**

Create `web-app/components/RefundForm.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function RefundForm({
  paymentId,
  remaining,
  currency,
}: {
  paymentId: string;
  remaining: number; // major units still refundable
  currency: string;
}) {
  const router = useRouter();
  const [amount, setAmount] = useState(remaining.toFixed(2));
  const [reason, setReason] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  if (remaining <= 0) {
    return <p className="text-paycore-muted text-sm">คืนเงินครบแล้ว — ไม่มียอดคงเหลือให้คืน</p>;
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    const res = await fetch(`/api/payments/${paymentId}/refund`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Idempotency-Key": crypto.randomUUID(),
      },
      body: JSON.stringify({ amount, reason: reason || undefined }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (res.ok) {
      setDone(true);
      router.refresh();
      return;
    }
    setErr(env?.message ?? `คืนเงินไม่สำเร็จ (${res.status})`);
  }

  return (
    <form onSubmit={submit} className="mt-6 space-y-3 border-t border-white/10 pt-6">
      <h2 className="font-medium">คืนเงิน</h2>
      <div className="flex gap-2">
        <input
          type="number"
          step="0.01"
          min="0.01"
          max={remaining}
          value={amount}
          onChange={(e) => setAmount(e.target.value)}
          className="w-40 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2 text-sm"
        />
        <input
          type="text"
          placeholder="เหตุผล (ไม่บังคับ)"
          maxLength={140}
          value={reason}
          onChange={(e) => setReason(e.target.value)}
          className="flex-1 rounded-lg bg-paycore-bg border border-white/10 px-3 py-2 text-sm"
        />
      </div>
      <p className="text-paycore-muted text-xs">คืนได้สูงสุด {remaining.toFixed(2)} {currency}</p>
      {err && <p className="text-red-400 text-sm">{err}</p>}
      {done && <p className="text-green-400 text-sm">คืนเงินสำเร็จ</p>}
      <button
        type="submit"
        disabled={busy}
        className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm disabled:opacity-60"
      >
        {busy ? "กำลังคืนเงิน…" : "ยืนยันคืนเงิน"}
      </button>
    </form>
  );
}
```

- [ ] **Step 2: Create the detail page**

Create `web-app/app/transactions/[id]/page.tsx`:

```tsx
import { redirect, notFound } from "next/navigation";
import { serverGet } from "@/lib/api";
import { formatDecimalMoney } from "@/lib/format";
import DashboardNav from "@/components/DashboardNav";
import RefundForm from "@/components/RefundForm";

type Payment = {
  id: string;
  amount: string;
  captured_amount: string;
  refunded_amount: string;
  currency: string;
  status: string;
  card_brand?: string;
  card_last4?: string;
  acquirer_ref?: string;
  auth_code?: string;
  reference?: string;
  created_at: string;
  updated_at: string;
};

// Refund is only allowed by the service from captured / partial_refunded.
const REFUNDABLE = new Set(["captured", "partial_refunded"]);

export default async function TransactionDetail({ params }: { params: { id: string } }) {
  const res = await serverGet(`/payments/${params.id}`);
  if (res.status === 401) redirect("/login");
  if (res.status === 404) notFound();
  if (!res.ok) throw new Error(`payment failed: ${res.status}`);
  const p: Payment = (await res.json()).data;

  const captured = parseFloat(p.captured_amount || "0");
  const refunded = parseFloat(p.refunded_amount || "0");
  const remaining = Math.max(0, captured - refunded);

  const rows: [string, string][] = [
    ["ยอด", formatDecimalMoney(p.amount, p.currency)],
    ["เรียกเก็บแล้ว", formatDecimalMoney(p.captured_amount, p.currency)],
    ["คืนแล้ว", formatDecimalMoney(p.refunded_amount, p.currency)],
    ["บัตร", `${p.card_brand || "card"}${p.card_last4 ? ` ····${p.card_last4}` : ""}`],
    ["อ้างอิงผู้รับชำระ", p.acquirer_ref || "—"],
    ["รหัสอนุมัติ", p.auth_code || "—"],
    ["อ้างอิงร้านค้า", p.reference || "—"],
    ["สร้างเมื่อ", new Date(p.created_at).toLocaleString("th-TH")],
  ];

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <DashboardNav active="/transactions" />
      <a href="/transactions" className="text-sm text-paycore-muted hover:text-paycore-text">← ธุรกรรม</a>

      <div className="rounded-xl2 bg-paycore-surface p-6 mt-4">
        <div className="flex items-start justify-between">
          <h1 className="text-2xl font-semibold">{formatDecimalMoney(p.amount, p.currency)}</h1>
          <span className="rounded-full px-3 py-1 text-xs bg-paycore-bg border border-white/10">{p.status}</span>
        </div>

        <dl className="mt-6 space-y-2 text-sm">
          {rows.map(([k, v]) => (
            <div key={k} className="flex justify-between">
              <dt className="text-paycore-muted">{k}</dt>
              <dd>{v}</dd>
            </div>
          ))}
        </dl>

        {REFUNDABLE.has(p.status) ? (
          <RefundForm paymentId={p.id} remaining={remaining} currency={p.currency} />
        ) : (
          <p className="mt-6 border-t border-white/10 pt-6 text-paycore-muted text-sm">
            สถานะนี้คืนเงินไม่ได้
          </p>
        )}
      </div>
    </main>
  );
}
```

- [ ] **Step 3: Verify the build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/transactions/[id]` compiles; `RefundForm` is bundled as a client component.

- [ ] **Step 4: Commit**

```bash
git add web-app/app/transactions/[id]/page.tsx web-app/components/RefundForm.tsx
git commit -m "feat(web): transaction detail + refund UI (idempotency-keyed)"
```

---

### Task 6: Frontend — settings (profile / rotate key / webhook)

**Files:**
- Create: `web-app/app/settings/page.tsx`
- Create: `web-app/components/RotateKeyButton.tsx`
- Create: `web-app/components/WebhookForm.tsx`

**Interfaces:**
- Consumes: `serverGet('/me')` → `MerchantProfile` `{ id, name, mcc?, settlement_currency, status, webhook_url? }`; `POST /api/me/rotate-key` → `RotatedKey` `{ api_key }` (shown once); `PUT /api/me/webhook` body `{ url }` → `WebhookConfig` `{ webhook_url, signing_secret }` (secret shown once).
- Produces: route `/settings`; `RotateKeyButton` (client); `WebhookForm` (client, `{ initialUrl }`).

- [ ] **Step 1: Create the rotate-key button (client component)**

Create `web-app/components/RotateKeyButton.tsx`:

```tsx
"use client";

import { useState } from "react";

export default function RotateKeyButton() {
  const [busy, setBusy] = useState(false);
  const [key, setKey] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);

  async function rotate() {
    setBusy(true);
    setErr(null);
    const res = await fetch("/api/me/rotate-key", { method: "POST" });
    const env = await res.json().catch(() => null);
    setBusy(false);
    setConfirming(false);
    if (res.ok) {
      setKey(env?.data?.api_key ?? null);
      return;
    }
    setErr(env?.message ?? `หมุนคีย์ไม่สำเร็จ (${res.status})`);
  }

  return (
    <div className="space-y-3">
      <p className="text-paycore-muted text-sm">
        API key จะแสดงเพียงครั้งเดียวตอนสร้างหรือหมุนคีย์ ระบบเก็บเฉพาะค่าแฮชจึงแสดงคีย์เดิมซ้ำไม่ได้
      </p>
      {key ? (
        <div className="rounded-lg bg-paycore-bg border border-white/10 p-3">
          <p className="text-paycore-muted text-xs mb-1">คีย์ใหม่ (บันทึกทันที จะไม่แสดงอีก)</p>
          <code className="text-sm break-all">{key}</code>
        </div>
      ) : confirming ? (
        <div className="flex gap-2">
          <button
            onClick={rotate}
            disabled={busy}
            className="rounded-lg bg-red-500/90 hover:bg-red-500 text-white px-4 py-2 text-sm disabled:opacity-60"
          >
            {busy ? "กำลังหมุนคีย์…" : "ยืนยันหมุนคีย์ (คีย์เดิมใช้ไม่ได้ทันที)"}
          </button>
          <button onClick={() => setConfirming(false)} className="rounded-lg border border-white/15 px-4 py-2 text-sm">
            ยกเลิก
          </button>
        </div>
      ) : (
        <button
          onClick={() => setConfirming(true)}
          className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm"
        >
          หมุน API key
        </button>
      )}
      {err && <p className="text-red-400 text-sm">{err}</p>}
    </div>
  );
}
```

- [ ] **Step 2: Create the webhook form (client component)**

Create `web-app/components/WebhookForm.tsx`:

```tsx
"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

export default function WebhookForm({ initialUrl }: { initialUrl: string }) {
  const router = useRouter();
  const [url, setUrl] = useState(initialUrl);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState<string | null>(null);
  const [secret, setSecret] = useState<string | null>(null);

  async function save(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    const res = await fetch("/api/me/webhook", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ url }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (res.ok) {
      setSecret(env?.data?.signing_secret ?? null);
      router.refresh();
      return;
    }
    setErr(env?.message ?? `บันทึกไม่สำเร็จ (${res.status})`);
  }

  return (
    <form onSubmit={save} className="space-y-3">
      <input
        type="url"
        required
        maxLength={2048}
        placeholder="https://example.com/webhooks/paycore"
        value={url}
        onChange={(e) => setUrl(e.target.value)}
        className="w-full rounded-lg bg-paycore-bg border border-white/10 px-3 py-2 text-sm"
      />
      {err && <p className="text-red-400 text-sm">{err}</p>}
      {secret && (
        <div className="rounded-lg bg-paycore-bg border border-white/10 p-3">
          <p className="text-paycore-muted text-xs mb-1">Signing secret ใหม่ (แสดงครั้งเดียว)</p>
          <code className="text-sm break-all">{secret}</code>
        </div>
      )}
      <button
        type="submit"
        disabled={busy}
        className="rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white px-4 py-2 text-sm disabled:opacity-60"
      >
        {busy ? "กำลังบันทึก…" : "บันทึก Webhook"}
      </button>
    </form>
  );
}
```

- [ ] **Step 3: Create the settings page**

Create `web-app/app/settings/page.tsx`:

```tsx
import { redirect } from "next/navigation";
import { serverGet } from "@/lib/api";
import DashboardNav from "@/components/DashboardNav";
import RotateKeyButton from "@/components/RotateKeyButton";
import WebhookForm from "@/components/WebhookForm";

type Profile = {
  id: string;
  name: string;
  mcc?: string;
  settlement_currency: string;
  status: string;
  webhook_url?: string;
};

export default async function SettingsPage() {
  const res = await serverGet("/me");
  if (res.status === 401) redirect("/login");
  if (!res.ok) throw new Error(`me failed: ${res.status}`);
  const p: Profile = (await res.json()).data;

  const rows: [string, string][] = [
    ["ชื่อร้านค้า", p.name],
    ["Merchant ID", p.id],
    ["สกุลเงินตั้งจ่าย", p.settlement_currency],
    ["MCC", p.mcc || "—"],
    ["สถานะ", p.status],
  ];

  return (
    <main className="min-h-screen p-8 max-w-2xl mx-auto">
      <DashboardNav active="/settings" />
      <h1 className="text-xl font-semibold mb-6">ตั้งค่า</h1>

      <section className="rounded-xl2 bg-paycore-surface p-6 mb-6">
        <h2 className="font-medium mb-4">ข้อมูลร้านค้า</h2>
        <dl className="space-y-2 text-sm">
          {rows.map(([k, v]) => (
            <div key={k} className="flex justify-between">
              <dt className="text-paycore-muted">{k}</dt>
              <dd className="break-all text-right">{v}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section className="rounded-xl2 bg-paycore-surface p-6 mb-6">
        <h2 className="font-medium mb-4">API key</h2>
        <RotateKeyButton />
      </section>

      <section className="rounded-xl2 bg-paycore-surface p-6">
        <h2 className="font-medium mb-4">Webhook</h2>
        {p.webhook_url && (
          <p className="text-paycore-muted text-xs mb-3">ปัจจุบัน: {p.webhook_url}</p>
        )}
        <WebhookForm initialUrl={p.webhook_url ?? ""} />
      </section>
    </main>
  );
}
```

- [ ] **Step 4: Verify the build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/settings` compiles; both client components bundle.

- [ ] **Step 5: Commit**

```bash
git add web-app/app/settings/page.tsx web-app/components/RotateKeyButton.tsx web-app/components/WebhookForm.tsx
git commit -m "feat(web): settings page (profile, rotate key, webhook)"
```

---

### Task 7: Route guard middleware + smoke + docs

**Files:**
- Create: `web-app/middleware.ts`
- Modify: `README.md`

**Interfaces:**
- Consumes: Next.js `NextRequest`/`NextResponse` (framework built-ins, no new dep); the `pc_session` cookie name.
- Produces: dashboard paths (`/`, `/transactions`, `/settings`, `/links` and subpaths) redirect to `/login` when the `pc_session` cookie is absent.

**Route-guard DECISION:** add a cheap `web-app/middleware.ts` that redirects to `/login` when the `pc_session` cookie is **absent** (this was flagged as a Phase-1 follow-up as pages grow). It is a first-line UX guard only — cookie *presence* is not proof of a valid session, so every page still keeps its `serverGet(...) → 401 → redirect('/login')` handling (the real authorization boundary is the backend `merchantAuth`). Do NOT verify the JWT in middleware (the signing secret lives only in the Go backend).

- [ ] **Step 1: Create the route guard**

Create `web-app/middleware.ts`:

```ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// First-line UX guard: bounce to /login when the session cookie is absent.
// This is NOT the authorization boundary — the Go backend's merchantAuth is.
// Cookie presence is not proof of a valid session, so pages still handle a 401
// from serverGet by redirecting to /login.
export function middleware(req: NextRequest) {
  const hasSession = req.cookies.has("pc_session");
  if (!hasSession) {
    const url = req.nextUrl.clone();
    url.pathname = "/login";
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

// Guard only the dashboard pages. /login, /pay/*, /api/*, and Next internals are
// excluded so public checkout and the login page stay reachable without a cookie.
export const config = {
  matcher: ["/", "/transactions/:path*", "/settings/:path*", "/links/:path*"],
};
```

- [ ] **Step 2: Verify the build**

Run: `cd web-app && npm run build`
Expected: build succeeds; the build output lists `middleware` (ƒ Middleware).

- [ ] **Step 3: Manual smoke (documented, run against a local stack)**

Run the backend in sandbox mode and the web app, then walk the flow. Expected results in each line:

```bash
# Terminal A — backend (sandbox so dev-login + card checkout work)
SANDBOX_MODE=true MIGRATE_ON_BOOT=true go run ./cmd/server
# Terminal B — web
cd web-app && BACKEND_URL=http://localhost:8080 npm run dev
```

- Visit `http://localhost:3000/` with no cookie → redirected to `/login` (middleware guard).
- Click "Dev login (sandbox)" → lands on `/` showing four stat tiles (volume/count/success/refund).
- Create a card payment so transactions are non-empty: create a payment link (`/links`), open its `/pay/<publicId>`, pay with the sandbox card, complete. Return to the dashboard.
- `/transactions` → the card payment appears; click it → detail page renders amounts via `formatDecimalMoney` (baht, not satang).
- On a `captured` payment, refund a partial amount → success message, `refunded` line updates after refresh; refund more than remaining → the page shows the envelope message ("refund exceeds captured amount", HTTP 422).
- `/settings` → profile rows render; "หมุน API key" reveals a new key once; save a webhook URL → signing secret shown once.
- Log out → `/` redirects to `/login`.

- [ ] **Step 4: Update the README**

Add a "Phase 5 — Dashboard polish" subsection to `README.md` (place it after the Phase 4 note). Content:

```markdown
### Phase 5 — Merchant dashboard polish

The cookie-authed Next.js dashboard is now a working merchant console:

- **Home** (`/`) — KPI tiles from `GET /v1/stats` (30-day card volume, count, success/refund ratios).
- **Transactions** (`/transactions`, `/transactions/[id]`) — card-payment list + detail, with a
  refund action (`POST /v1/payments/:id/refund`, idempotency-keyed).
- **Settings** (`/settings`) — business profile (`GET /v1/me`), API-key rotation
  (`POST /v1/me/rotate-key`, new key shown once) and webhook config (`PUT /v1/me/webhook`,
  signing secret shown once).

**Auth:** the dashboard-read routes and refund were retrofitted from API-key-only `auth` to the
combined session-OR-key `merchantAuth`, so the same endpoints serve both the cookie dashboard and
API-key clients. Money-moving routes (create / capture / void / 3DS) stay API-key-only. Refund
relies on the `pc_session` cookie being `SameSite=Lax` for CSRF protection and keeps its
`Idempotency-Key` gate.

**Transactions scope (known limitation / follow-up):** the Transactions page lists **card**
payments only (the `payments` table). PromptPay (`qr_payments`) and e-wallet mock transactions
(`checkout_sessions`, no `payments` row) are NOT yet in this unified list. A future
`GET /v1/transactions` endpoint aggregating `payments` + `qr_payments` + paid `checkout_sessions`
(with cross-table pagination) is the planned follow-up. The UI states this explicitly and does not
imply it shows payments it does not.
```

- [ ] **Step 5: Final full verification**

Run: `go build ./... && go test ./... && cd web-app && npm run build`
Expected: Go build + all tests PASS; web build succeeds.

- [ ] **Step 6: Commit**

```bash
git add web-app/middleware.ts README.md
git commit -m "feat(web): dashboard route guard + Phase 5 docs/smoke"
```

---

## Self-Review

**1. Spec coverage (§7 + §11 Phase 5):**
- §11 Phase 5 "transactions list/detail + refund UI" → Tasks 4, 5. ✓
- §11 Phase 5 "stats" → Task 3 (home tiles from `/stats`). ✓
- §11 Phase 5 "settings (API keys/webhook/profile)" → Task 6. ✓
- §7 dashboard routes `/` (stats+recent) → Task 3; `/transactions` (+detail+refund) → Tasks 4-5; `/settings` (rotate/webhook/profile) → Task 6; `/login`, `/links`, `/links/[id]`, `/pay/*` already exist (Phases 1-3) and are untouched. ✓
- §7 "business profile/logo" — profile is covered; **logo upload is intentionally out of scope** (no upload endpoint exists; adding storage + an endpoint would be a new subsystem, not "polish"). Noted here as a deliberate gap, not an oversight.
- §8 CSRF: the spec mentions double-submit tokens as an *additional* measure; Phase 5 relies on the shipped `SameSite=Lax` cookie + idempotency key and documents the reasoning (Architecture note, Task 1). A double-submit token is a reasonable future hardening but is not required to ship Phase 5 and is not built here.
- Enabler (spec §5.1/§5.2 "auth middleware ... /me, /stats, /payments, /settlements, /disputes รับได้ทั้งสองแบบ") → Task 1 retrofit. ✓

**2. Placeholder scan:** No "TODO/TBD/implement later" steps; every code step has complete, compilable code. All test commands include exact expected output (fail reason in Task 1 Step 2, PASS in Step 5).

**3. Type consistency:**
- `formatDecimalMoney(major: string | number, currency?)` defined in Task 2, used in Tasks 4-5 with `(p.amount, p.currency)` — signature matches. `formatMoney(minor: number)` used only for `VolumeMinor` in Task 3 — correct unit.
- `DashboardNav` default export with `{ active?: string }` — imported and passed `active="..."` in Tasks 3-6. ✓
- `RefundForm` props `{ paymentId, remaining: number, currency }` — Task 5 detail page passes exactly these (`remaining` computed as `captured - refunded`, major units). ✓
- Refund body `{ amount, reason? }` matches `domain.RefundRequest{ Amount decimal (required), Reason (max 140) }`; `amount` sent as a string is accepted by shopspring decimal's `UnmarshalJSON`. ✓
- Envelope fields read as `.message` / `.data` match `domain.APIResponse` JSON tags. ✓
- `MerchantProfile` fields (`webhook_url`, `settlement_currency`, `mcc`, `status`) match `internal/domain/dashboard.go`. `RotatedKey.api_key`, `WebhookConfig.{webhook_url,signing_secret}` match. ✓
- Router test fakes implement the exact current interface method sets (`PaymentService`: Create/Capture/Void/Refund/Get/List/HandleThreeDSResult/VerifyThreeDSResult; `MerchantService`: Onboard/Get/ResolveByAPIKeyHash/Profile/Stats/ListSettlements/RotateAPIKey/SetWebhook). ✓

**4. Transactions-scope honesty:** Option A is stated in the Architecture note, enforced by UI copy in Tasks 3 (home caption) and 4 (list caption), and documented as a follow-up in the README (Task 7). The dashboard never claims to show PromptPay or wallet transactions. ✓
</content>
</invoke>
