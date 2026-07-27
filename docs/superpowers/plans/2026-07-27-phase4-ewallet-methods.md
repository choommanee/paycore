# Phase 4 — E-wallet / redirect payment methods Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** add Beam-parity e-wallet / redirect methods (`mobile_banking`, `truemoney`, `shopeepay`, `alipay`, `wechat`, `card_installment`) to the existing Phase 3 hosted checkout as **mock adapters** — a customer picks one on `/pay/[publicId]`, is shown a mock "approve / decline" step, and the session completes on approval — all sandbox-gated, with no real PSP.

**Architecture:** These six methods share ONE code path, `payWallet`, that lives entirely on the existing `checkout_sessions` row (**Option A — no new table, no new column, no migration**). `payWallet` reserves a `single_use` link exactly like the card path (`ConsumePaymentLinkIfActive`), sets the session to `requires_action` with `selected_method = <slug>`, and returns the view — **it does NOT charge through `PaymentService` or `QRService` and moves no money.** A new **public, sandbox-only** endpoint `POST /v1/checkout/sessions/:token/confirm-mock {approve}` flips the session to `paid` (closing the `single_use` link via `markLinkPaid`) or `failed` (releasing the reservation via `ReleasePaymentLinkReservation`) — mirroring the existing `SandboxService` payer-simulator pattern but for checkout sessions. The frontend `CheckoutClient.tsx` gains wallet method chips, a "continue" button per wallet method, and an inline sandbox "approve / decline" panel that reuses the existing polling + success/failure views.

**Tech Stack:** Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 · shopspring/decimal · Next.js 14 App Router + TypeScript + Tailwind. **No new Go or npm dependency.**

## Global Constraints

- Module `github.com/yourco/payment-gateway`. Branch `feat/ewallet-methods` (already created). **NO new Go dependency; NO new npm dependency.**
- **Option A — no persistence changes.** The wallet flow uses only existing `checkout_sessions` columns (`status`, `selected_method`). It creates NO `payments` or `qr_payments` row (`payment_id`/`qr_payment_id` stay NULL). **No new migration, no new sqlc query, no `make sqlc` run.** (If you believe a column is needed, STOP — re-read Task 2; Option A is deliberate.)
- **Wallet mock is SANDBOX-gated.** When `SANDBOX_MODE=false`: `payWallet` returns `domain.ErrCheckoutMethodUnavailable` ("payment method not available", HTTP 422) and the `confirm-mock` route is NOT mounted (absent → 404). Real PSP integration is explicitly **out of scope** — document this, do not stub it.
- **`single_use` links are reserved + released identically to the card path.** `payWallet` reserves via `ConsumePaymentLinkIfActive` BEFORE going to `requires_action`; a decline releases via `ReleasePaymentLinkReservation`; an approve closes via `markLinkPaid` (idempotent — the link is already `paid` from the reservation). Reusable links are never reserved/closed. (Accepted limitation, identical to the existing card-3DS path: an abandoned `requires_action` wallet session holds the reservation until the session TTL — the card 3DS path already behaves this way and this plan does not change that.)
- **`confirm-mock` is public + sandbox-only.** No auth middleware; the opaque session token in the URL path is the credential; the service scopes everything to the resolved session (merchant context from the row, never the request). Not rate-limited (mirrors the sandbox QR pay endpoint).
- **No money conversion needed for wallets.** Option A charges nothing, so `payWallet`/`ConfirmMock` never call `money.FromMinor` or any payment service. (`money.FromMinor` remains the rule for anything that DOES call `PaymentService`/`QRService` — unchanged card/promptpay paths.)
- **Wallets carry no card data** — `payWallet` reads only `req.Method`; `CheckoutPayRequest.Card` is nil for wallet requests and must not be required.
- Response envelope: `domain.Success` / `domain.Created` / `domain.Error` (unchanged).
- Service tests: the fake repo embeds `repository.Querier` and already implements `ConsumePaymentLinkIfActive`/`ReleasePaymentLinkReservation`/`UpdatePaymentLinkStatus`, recording flips into `repo.linkStatusSets` (`"paid"` on consume, `"active"` on release). Reuse the Task-3-era fakes (`fakeCheckoutRepo`, `fakeCharger`, `fakeQR`, `fakeVault`, `newCheckoutSvc`, `mkLink`, `openSession`) already in `internal/service/checkout_service_test.go` — do NOT redefine them.
- Handler tests fake the service (`fakeCheckoutSvc`) + `app.Test`; public routes need no auth injection.
- TDD: write the failing test first, watch it fail, implement minimally, watch it pass, commit. Complete code in every step. Frequent commits.
- Run set: `go build ./... && go test ./...`; frontend `cd web-app && npm run build`.

---

## File map

| File | Change | Responsibility |
|------|--------|----------------|
| `internal/domain/checkout.go` | modify | Expand method registry; add `CheckoutWalletMethods`, `IsWalletMethod`; add `CheckoutConfirmMockRequest`; widen `CheckoutPayRequest.Method` validator |
| `internal/domain/checkout_test.go` | modify | Update the two Phase-3 `DisplayMethods` tests for the expanded registry; add wallet-registry tests |
| `internal/service/checkout_service.go` | modify | Add `payWallet`; dispatch wallet slugs in `Pay`; add `ConfirmMock` (+ interface method) |
| `internal/service/checkout_service_test.go` | modify | Append wallet + confirm-mock service tests (reuse existing fakes) |
| `internal/handler/checkout_handler.go` | modify | Add `ConfirmMock` HTTP handler |
| `internal/handler/checkout_handler_test.go` | modify | Add `confirmMockFn` to `fakeCheckoutSvc`; mount + test the route |
| `internal/router/router.go` | modify | Mount `POST /checkout/sessions/:token/confirm-mock` gated on existing `sandbox` param |
| `web-app/components/CheckoutClient.tsx` | modify | Wallet chips + per-wallet continue button + inline sandbox approve/decline; reuse polling/success |
| `README.md` | modify | Phase 4 note (methods, confirm-mock, sandbox gate) |

No changes to migrations, sqlc queries, `internal/config/config.go`, `internal/handler/handler.go`, or `cmd/server/main.go` — the checkout service already receives `cfg.SandboxMode` and the router already receives `sandbox bool`.

---

### Task 1: Domain — expand method registry, wallet set, confirm-mock request

**Files:**
- Modify: `internal/domain/checkout.go`
- Modify: `internal/domain/checkout_test.go`

**Interfaces:**
- Consumes: existing `CheckoutSupportedMethods`, `DisplayMethods`, `CheckoutPayRequest`.
- Produces:
  - `domain.CheckoutSupportedMethods` widened to `{"card","promptpay","mobile_banking","truemoney","shopeepay","alipay","wechat","card_installment"}`
  - `domain.CheckoutWalletMethods []string` (= the six wallet/redirect slugs)
  - `domain.IsWalletMethod(m string) bool`
  - `domain.CheckoutConfirmMockRequest{ Approve bool }`
  - `CheckoutPayRequest.Method` validator widened to accept all eight slugs

- [ ] **Step 1: Update the failing tests for the expanded registry**

The two existing Phase-3 tests assert the registry is exactly `{card, promptpay}`; the expansion makes them fail. REPLACE the entire body of `internal/domain/checkout_test.go` with:
```go
package domain

import (
	"reflect"
	"testing"
)

func TestDisplayMethodsEmptyMeansAllSupported(t *testing.T) {
	got := DisplayMethods(nil)
	want := []string{"card", "promptpay", "mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("empty allowed = %v want %v", got, want)
	}
}

func TestDisplayMethodsIntersectsAndKeepsSupportedOrder(t *testing.T) {
	// Link allows a wallet + promptpay + card, in a different order.
	got := DisplayMethods([]string{"truemoney", "promptpay", "card"})
	// All three are now supported; survivors keep CheckoutSupportedMethods order
	// (card, promptpay, then truemoney).
	want := []string{"card", "promptpay", "truemoney"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestDisplayMethodsCardOnly(t *testing.T) {
	got := DisplayMethods([]string{"card"})
	if !reflect.DeepEqual(got, []string{"card"}) {
		t.Fatalf("got %v want [card]", got)
	}
}

func TestDisplayMethodsWalletOnly(t *testing.T) {
	got := DisplayMethods([]string{"shopeepay", "alipay"})
	if !reflect.DeepEqual(got, []string{"shopeepay", "alipay"}) {
		t.Fatalf("got %v want [shopeepay alipay]", got)
	}
}

func TestIsWalletMethod(t *testing.T) {
	for _, m := range []string{"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"} {
		if !IsWalletMethod(m) {
			t.Fatalf("IsWalletMethod(%q) = false, want true", m)
		}
	}
	for _, m := range []string{"card", "promptpay", "", "paypal"} {
		if IsWalletMethod(m) {
			t.Fatalf("IsWalletMethod(%q) = true, want false", m)
		}
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/domain/ -run 'TestDisplayMethods|TestIsWalletMethod' -v`
Expected: FAIL — `TestDisplayMethodsEmptyMeansAllSupported`/`...IntersectsAndKeepsSupportedOrder` fail on the old 2-element registry; `TestDisplayMethodsWalletOnly` fails (wallets not yet supported); `IsWalletMethod` undefined (compile error).

- [ ] **Step 3: Expand the registry + add the wallet helpers**

In `internal/domain/checkout.go`, REPLACE the `CheckoutSupportedMethods` var (and its doc comment) with:
```go
// CheckoutSupportedMethods is the set of methods the hosted checkout can process.
// Order is the display order on the payment page. card (sandbox raw PAN) and
// promptpay (all modes) come from Phase 3; the six wallet / redirect methods are
// Phase 4 MOCK adapters (sandbox-gated) that complete via /confirm-mock.
var CheckoutSupportedMethods = []string{
	"card", "promptpay",
	"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment",
}

// CheckoutWalletMethods are the Phase 4 e-wallet / redirect methods. They share a
// single MOCK adapter (service.payWallet): no card data, no PaymentService /
// QRService call, no money moved — the session goes to requires_action and is
// completed by the sandbox-only /confirm-mock endpoint. In production (sandbox
// off) they are refused (ErrCheckoutMethodUnavailable); real PSP wiring is out of
// scope for this phase.
var CheckoutWalletMethods = []string{
	"mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment",
}

// IsWalletMethod reports whether m is a Phase 4 wallet / redirect method handled
// by the mock wallet adapter.
func IsWalletMethod(m string) bool {
	for _, w := range CheckoutWalletMethods {
		if w == m {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Widen the pay-request validator + add the confirm-mock request**

In `internal/domain/checkout.go`, REPLACE the `CheckoutPayRequest` struct with the widened `oneof` (add the six slugs), and ADD the confirm-mock request type after it:
```go
// CheckoutPayRequest selects a method and carries its data. Card is required only
// when Method == "card" (enforced in the service). Wallet methods carry no data.
type CheckoutPayRequest struct {
	Method        string     `json:"method" validate:"required,oneof=card promptpay mobile_banking truemoney shopeepay alipay wechat card_installment"`
	Card          *CardInput `json:"card" validate:"omitempty"`
	CustomerEmail string     `json:"customer_email" validate:"omitempty,email"`
}

// CheckoutConfirmMockRequest is the body of the sandbox-only confirm-mock endpoint
// that simulates a wallet approve (true) or decline (false).
type CheckoutConfirmMockRequest struct {
	Approve bool `json:"approve"`
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `go test ./internal/domain/ -v`
Expected: PASS (all `DisplayMethods` tests + `TestIsWalletMethod`). Then `go build ./...`.

- [ ] **Step 6: Commit**

```bash
git add internal/domain/checkout.go internal/domain/checkout_test.go
git commit -m "feat(domain): phase-4 wallet method registry + confirm-mock request"
```

---

### Task 2: Service — `payWallet` + wallet dispatch in `Pay`

**Files:**
- Modify: `internal/service/checkout_service.go`
- Modify: `internal/service/checkout_service_test.go`

**Interfaces:**
- Consumes: `repository.GetPaymentLink`, `repository.ConsumePaymentLinkIfActive`, `repository.UpdateCheckoutSession`, `repository.ReleasePaymentLinkReservation` (via existing `releaseLinkReservation`), `domain.IsWalletMethod`, existing helpers `strFallback`, `buildView`, `transition`.
- Produces: `payWallet` transitions an `open` session → `requires_action` with `selected_method = <slug>` (reserving a `single_use` link first); `Pay` routes the six wallet slugs to `payWallet`. **No `payments`/`qr_payments` row; no money moved.**

- [ ] **Step 1: Write the failing tests**

Append to `internal/service/checkout_service_test.go` (reuses `newFakeCheckoutRepo`, `mkLink`, `newCheckoutSvc`, `openSession`, `middleware.HashAPIKey`):
```go
// ---- Task: payWallet -------------------------------------------------------

func TestPayWalletRequiresActionAndReservesSingleUse(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"card", "truemoney"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "truemoney"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if view.SelectedMethod != "truemoney" {
		t.Fatalf("selected_method = %q want truemoney", view.SelectedMethod)
	}
	// Wallet mock moves no money: no QR payload, no 3DS redirect URL.
	if view.QRPayload != "" || view.NextActionURL != "" {
		t.Fatalf("wallet must not set qr_payload/next_action_url: %q / %q", view.QRPayload, view.NextActionURL)
	}
	// single_use link reserved (flipped to paid) exactly like the card path.
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid] (reserved)", repo.linkStatusSets)
	}
}

func TestPayWalletBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"alipay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false) // sandbox OFF
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "alipay"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (wallet is sandbox-only)", err)
	}
	// Refused before any reservation.
	if len(repo.linkStatusSets) != 0 {
		t.Fatalf("no link reservation expected, got %v", repo.linkStatusSets)
	}
}

func TestPayWalletMethodNotAllowedByLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"}) // wallet NOT allowed
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "shopeepay"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable", err)
	}
}

func TestPayWalletReusableLinkNotReserved(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_abc", 5000, []string{"wechat"})
	l.LinkType = "reusable"
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := openSession(t, repo, svc)

	view, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "wechat"})
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action", view.Status)
	}
	if len(repo.linkStatusSets) != 0 {
		t.Fatalf("reusable link must not reserve, got %v", repo.linkStatusSets)
	}
}

func TestPayWalletSingleUseAlreadyConsumedRefuses(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	l := mkLink(repo, merchant, "pl_abc", 5000, []string{"truemoney"})
	l.Status = "paid" // already consumed by a prior session
	repo.putLink(l)
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	// Open a session directly against the (now paid) link's stored row via a fresh
	// active link, then simulate the race: re-mark it paid before paying.
	// Simpler: open while active, then flip to paid, then pay.
	l.Status = "active"
	repo.putLink(l)
	tok := openSession(t, repo, svc)
	l.Status = "paid"
	repo.putLink(l)

	_, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "truemoney"})
	if err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (link already consumed)", err)
	}
	row, _ := repo.GetCheckoutSessionByTokenHash(context.Background(), middleware.HashAPIKey(tok))
	if row.Status != string(domain.CheckoutFailed) {
		t.Fatalf("session status = %q want failed", row.Status)
	}
}
```

> Note: `openSession` (defined in Task-3-era code, ~line 330) creates a session from `"pl_abc"`. The fake repo's `ConsumePaymentLinkIfActive` returns `pgx.ErrNoRows` when the stored link's status is not `"active"` (that is how it mirrors the real query), which drives `TestPayWalletSingleUseAlreadyConsumedRefuses`.

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/service/ -run TestPayWallet -v`
Expected: FAIL — the wallet slugs fall into the `default` case of `Pay`'s switch, returning `ErrCheckoutMethodUnavailable` for the success cases (assertions fail).

- [ ] **Step 3: Route wallet slugs + implement `payWallet`**

In `internal/service/checkout_service.go`, in `Pay`, REPLACE the `switch req.Method { ... }` block with one that dispatches wallet slugs:
```go
	switch req.Method {
	case "promptpay":
		return s.payPromptPay(ctx, row, req)
	case "card":
		return s.payCard(ctx, row, req)
	default:
		// The six Phase 4 wallet / redirect methods share one mock adapter.
		if domain.IsWalletMethod(req.Method) {
			return s.payWallet(ctx, row, req)
		}
		return nil, domain.ErrCheckoutMethodUnavailable
	}
```

Then ADD `payWallet` (place it right after `payCard`):
```go
// payWallet is the MOCK adapter for the Phase 4 e-wallet / redirect methods
// (mobile_banking, truemoney, shopeepay, alipay, wechat, card_installment). It is
// SANDBOX-ONLY: in production these are stubs for a real PSP redirect, which is
// out of scope, so it refuses. Unlike card/promptpay it charges NOTHING — it
// reserves a single_use link (exactly like the card path, so two sessions cannot
// both complete the same link) and moves the session to requires_action. The
// sandbox-only ConfirmMock endpoint then flips it to paid (closing the link) or
// failed (releasing the reservation). No payments/qr_payments row is created and
// no money is converted or moved.
func (s *checkoutService) payWallet(ctx context.Context, row repository.CheckoutSession, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error) {
	if !s.sandbox {
		// Real PSP redirect integration is out of scope: refuse rather than
		// silently mark anything paid.
		return nil, domain.ErrCheckoutMethodUnavailable
	}

	// Reserve a single_use link BEFORE going to requires_action — identical to the
	// card path. ConsumePaymentLinkIfActive atomically flips active -> paid; no row
	// means another session already consumed it, so refuse without proceeding.
	var reservedLink *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		link, err := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		})
		// Fail closed on any load error: a session referencing a link must load it
		// to know whether it is single_use and needs reservation.
		if err != nil {
			return nil, err
		}
		if link.LinkType == "single_use" {
			consumed, cerr := s.repo.ConsumePaymentLinkIfActive(ctx, repository.ConsumePaymentLinkIfActiveParams{
				ID: row.PaymentLinkID, MerchantID: row.MerchantID,
			})
			if cerr != nil {
				if errors.Is(cerr, pgx.ErrNoRows) {
					_, _ = s.transition(ctx, row, domain.CheckoutFailed)
					return nil, domain.ErrCheckoutMethodUnavailable
				}
				return nil, cerr
			}
			reservedLink = &consumed
		}
	}

	updated, err := s.repo.UpdateCheckoutSession(ctx, repository.UpdateCheckoutSessionParams{
		ID:             row.ID,
		Status:         string(domain.CheckoutRequiresAction),
		SelectedMethod: req.Method,
		PaymentID:      row.PaymentID,   // stays NULL
		QrPaymentID:    row.QrPaymentID, // stays NULL
		CustomerEmail:  strFallback(req.CustomerEmail, row.CustomerEmail),
	})
	if err != nil {
		// Persisting requires_action failed: release the reservation so the
		// single_use link can be retried.
		s.releaseLinkReservation(ctx, reservedLink, row)
		return nil, err
	}
	// No QRPayload / NextActionURL: the mock approve/decline is driven inline by
	// the frontend against ConfirmMock; there is no external redirect.
	return s.buildView(ctx, updated, nil), nil
}
```

> This reuses the existing `releaseLinkReservation(ctx, reserved *repository.PaymentLink, row)`, `transition`, `strFallback`, and `buildView`. It adds no new imports (`errors`, `pgx`, `domain`, `repository` are already imported).

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/ -run 'TestPayWallet|TestPay|TestGet|TestCreateFromLink' -v`
Expected: PASS. Then `go build ./... && go test ./internal/service/...`.

- [ ] **Step 5: Commit**

```bash
git add internal/service/checkout_service.go internal/service/checkout_service_test.go
git commit -m "feat(service): payWallet mock adapter + wallet dispatch (sandbox-gated)"
```

---

### Task 3: Service — `ConfirmMock` (approve → paid / decline → failed)

**Files:**
- Modify: `internal/service/checkout_service.go` (add `ConfirmMock` to the `CheckoutService` interface + implement it)
- Modify: `internal/service/checkout_service_test.go` (append confirm-mock tests)

**Interfaces:**
- Consumes: `loadByToken`, `transition`, `markLinkPaid`, `releaseLinkReservation`, `buildView`, `Get`, `repository.GetPaymentLink`, `domain.IsWalletMethod`.
- Produces: `CheckoutService.ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error)` — approve → `paid` + close single_use link; decline → `failed` + release reservation; idempotent for non-`requires_action`/non-wallet sessions (returns current view); sandbox-only.

> **Cross-task note:** adding `ConfirmMock` to the `CheckoutService` interface makes the handler-package test fake (`fakeCheckoutSvc`) stop satisfying the interface until Task 4 adds its method. Therefore Task 3's gate is `go build ./...` (build excludes `_test.go`) + `go test ./internal/service/...` ONLY — the full `go test ./...` goes green again in Task 4. This mirrors the Phase 3 plan's staged interface growth.

- [ ] **Step 1: Write the failing tests**

Append to `internal/service/checkout_service_test.go`:
```go
// ---- Task: ConfirmMock -----------------------------------------------------

// walletRequiresAction opens a session and drives it to requires_action via a
// wallet method, returning the token.
func walletRequiresAction(t *testing.T, repo *fakeCheckoutRepo, svc CheckoutService, method string) string {
	t.Helper()
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: method}); err != nil {
		t.Fatalf("Pay(%s): %v", method, err)
	}
	return tok
}

func TestConfirmMockApproveMarksPaidAndClosesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"truemoney"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, svc, "truemoney")

	view, err := svc.ConfirmMock(context.Background(), tok, true)
	if err != nil {
		t.Fatalf("ConfirmMock approve: %v", err)
	}
	if view.Status != string(domain.CheckoutPaid) {
		t.Fatalf("status = %q want paid", view.Status)
	}
	// Reserved at pay time -> [paid]; approve's markLinkPaid is an idempotent
	// no-op (link already paid), so the ledger stays [paid].
	if len(repo.linkStatusSets) != 1 || repo.linkStatusSets[0] != "paid" {
		t.Fatalf("link status sets = %v want [paid]", repo.linkStatusSets)
	}
}

func TestConfirmMockDeclineMarksFailedAndReleasesLink(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"shopeepay"})
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, svc, "shopeepay")

	view, err := svc.ConfirmMock(context.Background(), tok, false)
	if err != nil {
		t.Fatalf("ConfirmMock decline: %v", err)
	}
	if view.Status != string(domain.CheckoutFailed) {
		t.Fatalf("status = %q want failed", view.Status)
	}
	// Reserve then release -> [paid active].
	if len(repo.linkStatusSets) != 2 || repo.linkStatusSets[0] != "paid" || repo.linkStatusSets[1] != "active" {
		t.Fatalf("link status sets = %v want [paid active]", repo.linkStatusSets)
	}
}

func TestConfirmMockBlockedOutsideSandbox(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"alipay"})
	// Reach requires_action while sandbox is on, then confirm with a sandbox-off
	// service pointed at the same repo.
	on := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	tok := walletRequiresAction(t, repo, on, "alipay")
	off := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, false)

	if _, err := off.ConfirmMock(context.Background(), tok, true); err != domain.ErrCheckoutMethodUnavailable {
		t.Fatalf("err = %v want ErrCheckoutMethodUnavailable (confirm-mock is sandbox-only)", err)
	}
}

func TestConfirmMockNonWalletSessionIsNoop(t *testing.T) {
	repo := newFakeCheckoutRepo()
	merchant := uuid.New()
	mkLink(repo, merchant, "pl_abc", 5000, []string{"promptpay"})
	qr := &fakeQR{} // default: awaiting_payment
	svc := newCheckoutSvc(repo, &fakeCharger{}, qr, &fakeVault{}, true)
	tok := openSession(t, repo, svc)
	if _, err := svc.Pay(context.Background(), tok, domain.CheckoutPayRequest{Method: "promptpay"}); err != nil {
		t.Fatalf("Pay promptpay: %v", err)
	}
	// promptpay is requires_action but NOT a wallet method: confirm-mock must not
	// flip it; it returns the current (requires_action) view unchanged.
	view, err := svc.ConfirmMock(context.Background(), tok, true)
	if err != nil {
		t.Fatalf("ConfirmMock: %v", err)
	}
	if view.Status != string(domain.CheckoutRequiresAction) {
		t.Fatalf("status = %q want requires_action (unchanged)", view.Status)
	}
}

func TestConfirmMockUnknownTokenIs404(t *testing.T) {
	repo := newFakeCheckoutRepo()
	svc := newCheckoutSvc(repo, &fakeCharger{}, &fakeQR{}, &fakeVault{}, true)
	if _, err := svc.ConfirmMock(context.Background(), "cs_nope", true); err != domain.ErrCheckoutSessionNotFound {
		t.Fatalf("err = %v want ErrCheckoutSessionNotFound", err)
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/service/ -run TestConfirmMock -v`
Expected: FAIL — `ConfirmMock` undefined (compile error).

- [ ] **Step 3: Add `ConfirmMock` to the interface + implement**

In `internal/service/checkout_service.go`, ADD the method to the `CheckoutService` interface (after `Pay`):
```go
	// ConfirmMock simulates a wallet approve (true) / decline (false) for a session
	// awaiting action. SANDBOX ONLY — the HTTP route is absent in production.
	ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error)
```

Then ADD the implementation (place it after `payWallet`):
```go
// ConfirmMock completes a wallet session in requires_action: approve -> paid
// (closing a single_use link), decline -> failed (releasing the reservation). It
// is SANDBOX ONLY (the route is not mounted in production) and applies ONLY to a
// wallet session awaiting action; anything else (already terminal, promptpay
// QR-awaiting, expired) is a no-op that returns the current view. It scopes
// everything to the session the token resolves; merchant context comes from the
// row.
func (s *checkoutService) ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
	if !s.sandbox {
		return nil, domain.ErrCheckoutMethodUnavailable
	}
	row, err := s.loadByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if row.ExpiresAt.Valid && time.Now().After(row.ExpiresAt.Time) {
		_, _ = s.transition(ctx, row, domain.CheckoutExpired)
		return nil, domain.ErrCheckoutSessionExpired
	}
	// Only a wallet session actually awaiting action can be confirmed. Anything
	// else is an idempotent no-op returning the current view (double-submit, an
	// already-terminal session, or a non-wallet requires_action like promptpay
	// which confirms via the QR webhook instead).
	if domain.CheckoutStatus(row.Status) != domain.CheckoutRequiresAction || !domain.IsWalletMethod(row.SelectedMethod) {
		return s.Get(ctx, token)
	}

	if approve {
		updated, err := s.transition(ctx, row, domain.CheckoutPaid)
		if err != nil {
			return nil, err
		}
		// single_use link was already reserved (flipped to paid) in payWallet;
		// this is an idempotent no-op for it and handles reusable (no-op) too.
		s.markLinkPaid(ctx, updated)
		return s.buildView(ctx, updated, nil), nil
	}

	updated, err := s.transition(ctx, row, domain.CheckoutFailed)
	if err != nil {
		return nil, err
	}
	// Release the single_use reservation so the link can be retried. Load the link
	// to pass releaseLinkReservation a non-nil pointer ONLY when it is single_use
	// (reusable links were never reserved, so pass nil -> no-op).
	var reserved *repository.PaymentLink
	if row.PaymentLinkID.Valid {
		if l, lerr := s.repo.GetPaymentLink(ctx, repository.GetPaymentLinkParams{
			ID: row.PaymentLinkID, MerchantID: row.MerchantID,
		}); lerr == nil && l.LinkType == "single_use" {
			reserved = &l
		}
	}
	s.releaseLinkReservation(ctx, reserved, row)
	return s.buildView(ctx, updated, nil), nil
}
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `go test ./internal/service/ -run TestConfirmMock -v`
Expected: PASS. Then `go build ./... && go test ./internal/service/...`.
> Do NOT run the full `go test ./...` yet — the handler test fake gets its `ConfirmMock` method in Task 4 (see the cross-task note above).

- [ ] **Step 5: Commit**

```bash
git add internal/service/checkout_service.go internal/service/checkout_service_test.go
git commit -m "feat(service): ConfirmMock wallet approve/decline (sandbox-only)"
```

---

### Task 4: Handler + router — public sandbox-gated `confirm-mock` route

**Files:**
- Modify: `internal/handler/checkout_handler.go`
- Modify: `internal/handler/checkout_handler_test.go`
- Modify: `internal/router/router.go`

**Interfaces:**
- Consumes: `service.CheckoutService.ConfirmMock`, `domain.CheckoutConfirmMockRequest`.
- Produces: `(*CheckoutHandler).ConfirmMock(c *fiber.Ctx) error`; route `POST /v1/checkout/sessions/:token/confirm-mock` mounted ONLY when `sandbox` is true.

- [ ] **Step 1: Write the failing handler tests + extend the fake**

In `internal/handler/checkout_handler_test.go`, ADD a `confirmMockFn` field + method to `fakeCheckoutSvc` and mount the route in `newCheckoutApp`. REPLACE the `fakeCheckoutSvc` struct + its methods + `newCheckoutApp` with:
```go
type fakeCheckoutSvc struct {
	createFn      func(ctx context.Context, publicID string) (*domain.CheckoutSessionView, error)
	getFn         func(ctx context.Context, token string) (*domain.CheckoutSessionView, error)
	payFn         func(ctx context.Context, token string, req domain.CheckoutPayRequest) (*domain.CheckoutSessionView, error)
	confirmMockFn func(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error)
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
func (f *fakeCheckoutSvc) ConfirmMock(ctx context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
	return f.confirmMockFn(ctx, token, approve)
}

func newCheckoutApp(svc *fakeCheckoutSvc) *fiber.App {
	h := NewCheckoutHandler(svc, zerolog.Nop())
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler(zerolog.Nop())})
	app.Post("/v1/checkout/sessions", h.Create)
	app.Get("/v1/checkout/sessions/:token", h.Get)
	app.Post("/v1/checkout/sessions/:token/pay", h.Pay)
	app.Post("/v1/checkout/sessions/:token/confirm-mock", h.ConfirmMock)
	return app
}
```

Then ADD these test functions:
```go
func TestCheckoutConfirmMockApprove(t *testing.T) {
	var gotToken string
	var gotApprove bool
	svc := &fakeCheckoutSvc{confirmMockFn: func(_ context.Context, token string, approve bool) (*domain.CheckoutSessionView, error) {
		gotToken, gotApprove = token, approve
		return &domain.CheckoutSessionView{ID: uuid.New(), Status: "paid", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/confirm-mock", strings.NewReader(`{"approve":true}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d want 200 (%s)", resp.StatusCode, b)
	}
	if gotToken != "cs_tok" || !gotApprove {
		t.Fatalf("token/approve = %q/%v want cs_tok/true", gotToken, gotApprove)
	}
}

func TestCheckoutConfirmMockDecline(t *testing.T) {
	var gotApprove = true
	svc := &fakeCheckoutSvc{confirmMockFn: func(_ context.Context, _ string, approve bool) (*domain.CheckoutSessionView, error) {
		gotApprove = approve
		return &domain.CheckoutSessionView{Status: "failed", AllowedMethods: []string{}}, nil
	}}
	app := newCheckoutApp(svc)

	req := httptest.NewRequest("POST", "/v1/checkout/sessions/cs_tok/confirm-mock", strings.NewReader(`{"approve":false}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d want 200", resp.StatusCode)
	}
	if gotApprove {
		t.Fatalf("approve = true want false")
	}
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `go test ./internal/handler/ -run TestCheckoutConfirmMock -v`
Expected: FAIL — `h.ConfirmMock` undefined (compile error).

- [ ] **Step 3: Implement the handler**

In `internal/handler/checkout_handler.go`, ADD after `Pay`:
```go
// ConfirmMock simulates a wallet approve/decline for a session awaiting action.
// PUBLIC + SANDBOX ONLY — the router mounts this route only when sandbox is on.
// @Router /v1/checkout/sessions/{token}/confirm-mock [post]
func (h *CheckoutHandler) ConfirmMock(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_TOKEN", "missing session token")
	}
	var req domain.CheckoutConfirmMockRequest
	if err := c.BodyParser(&req); err != nil {
		return domain.Error(c, fiber.StatusBadRequest, "INVALID_BODY", "invalid request body")
	}
	view, err := h.svc.ConfirmMock(c.Context(), token, req.Approve)
	if err != nil {
		return err
	}
	return domain.Success(c, view)
}
```
> `CheckoutConfirmMockRequest` has only a bool, so there is nothing to validate — no `h.validate.Struct` call (a missing `approve` correctly defaults to `false` = decline).

- [ ] **Step 4: Run handler tests to verify they pass**

Run: `go test ./internal/handler/ -run TestCheckout -v`
Expected: PASS (existing + the two new confirm-mock tests).

- [ ] **Step 5: Mount the route (sandbox-gated) in the router**

In `internal/router/router.go`, inside the `if h.Checkout != nil { ... }` block, ADD the sandbox-gated route after the `pay` route (and extend the block comment):
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
		// Wallet approve/decline simulator. Mounted ONLY when SANDBOX_MODE=true so
		// production has no way to mark a session paid without a real PSP callback
		// (route absent -> 404). Public: the session token is the credential.
		if sandbox {
			checkout.Post("/sessions/:token/confirm-mock", h.Checkout.ConfirmMock)
		}
	}
```
> `sandbox` is already a parameter of `Setup` (used by the `/v1/sandbox` block below); no signature change. `cmd/server/main.go` already passes `cfg.SandboxMode`.

- [ ] **Step 6: Build + full suite**

Run: `go build ./... && go test ./...`
Expected: PASS everywhere (the handler-package build is now green again — the fake has `ConfirmMock`).

- [ ] **Step 7: Commit**

```bash
git add internal/handler/checkout_handler.go internal/handler/checkout_handler_test.go internal/router/router.go
git commit -m "feat(api): public sandbox-only /checkout confirm-mock route + handler"
```

---

### Task 5: Frontend — wallet chips, continue button, inline sandbox approve/decline

**Files:**
- Modify: `web-app/components/CheckoutClient.tsx`

**Interfaces:**
- Consumes: `POST /api/checkout/sessions/:token/pay {method}` (wallet), `POST /api/checkout/sessions/:token/confirm-mock {approve}` (sandbox), `GET /api/checkout/sessions/:token` (poll) — all via the existing `/api/*` proxy (`web-app/next.config.js` → `${BACKEND_URL}/v1/*`).

> The page routes any non-`open` session to `CheckoutStatusView`. That view already handles `paid`/`expired`/`failed` and PromptPay QR; this task adds the wallet `requires_action` branch (inline approve/decline) and wallet method chips + a per-wallet "continue" button in the selector. Reuses the existing polling loop and success/failure UI. No new npm dependency.

- [ ] **Step 1: Replace `CheckoutClient.tsx` with the wallet-aware version**

REPLACE the entire contents of `web-app/components/CheckoutClient.tsx` with:
```tsx
"use client";

import { useEffect, useRef, useState } from "react";
import { formatMoney } from "@/lib/format";

declare global {
  interface Window {
    QRCode?: {
      new (el: HTMLElement, opts: { text: string; width: number; height: number; correctLevel: number }): unknown;
      CorrectLevel: { L: number; M: number; Q: number; H: number };
    };
  }
}

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

// Labels for every method the registry can surface (Phase 3 card/promptpay +
// Phase 4 wallet / redirect methods).
const METHOD_LABEL: Record<string, string> = {
  card: "บัตรเครดิต/เดบิต",
  promptpay: "PromptPay",
  mobile_banking: "Mobile Banking",
  truemoney: "TrueMoney Wallet",
  shopeepay: "ShopeePay",
  alipay: "Alipay",
  wechat: "WeChat Pay",
  card_installment: "ผ่อนชำระบัตร",
};

// The six Phase 4 wallet / redirect methods share one mock flow.
const WALLET_METHODS = ["mobile_banking", "truemoney", "shopeepay", "alipay", "wechat", "card_installment"];
const isWallet = (m: string) => WALLET_METHODS.includes(m);

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

  // Any non-open session (requires_action for promptpay QR or a wallet mock, or a
  // terminal state) is driven by the status view.
  if (view.status !== "open") {
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
        <PayMethodButton token={token} method="promptpay" label="สร้าง QR PromptPay" busyLabel="กำลังสร้าง QR…" onDone={setView} setErr={setErr} />
      )}

      {isWallet(method) && (
        <>
          {!view.sandbox && (
            <p className="text-xs rounded-lg bg-yellow-500/10 text-yellow-300 px-3 py-2">
              ช่องทางนี้ยังไม่พร้อมใช้งานบนระบบนี้
            </p>
          )}
          {view.sandbox && (
            <PayMethodButton
              token={token}
              method={method}
              label={`ดำเนินการต่อด้วย ${METHOD_LABEL[method] ?? method}`}
              busyLabel="กำลังดำเนินการ…"
              onDone={setView}
              setErr={setErr}
            />
          )}
        </>
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

// PayMethodButton POSTs /pay for a data-less method (promptpay or a wallet slug),
// then hands the returned requires_action view to the status view via onDone.
function PayMethodButton({
  token, method, label, busyLabel, onDone, setErr,
}: {
  token: string; method: string; label: string; busyLabel: string;
  onDone: (v: CheckoutView) => void; setErr: (s: string) => void;
}) {
  const [busy, setBusy] = useState(false);
  async function start() {
    setErr("");
    setBusy(true);
    const res = await fetch(`/api/checkout/sessions/${token}/pay`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ method }),
    });
    const env = await res.json().catch(() => null);
    setBusy(false);
    if (!res.ok) {
      setErr(env?.message ?? "ดำเนินการไม่สำเร็จ");
      return;
    }
    onDone(env.data as CheckoutView);
  }
  return (
    <button onClick={start} disabled={busy} className="w-full rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2 disabled:opacity-60">
      {busy ? busyLabel : label}
    </button>
  );
}

// CheckoutStatusView renders the PromptPay QR or the wallet mock approve/decline
// panel while awaiting action, polls session status until a terminal state, then
// shows success + optional return_url.
function CheckoutStatusView({ token, initial }: { token: string; initial: CheckoutView }) {
  const [view, setView] = useState<CheckoutView>(initial);
  const qrBox = useRef<HTMLDivElement>(null);

  const walletAwaiting =
    view.status === "requires_action" && !!view.selected_method && WALLET_METHODS.includes(view.selected_method);

  // Render the QR whenever a PromptPay payload is present and not yet paid.
  useEffect(() => {
    if (!view.qr_payload || view.status === "paid") return;
    let cancelled = false;
    (async () => {
      try {
        const QR = await waitForQRCode();
        if (cancelled || !qrBox.current) return;
        qrBox.current.innerHTML = "";
        new QR(qrBox.current, { text: view.qr_payload!, width: 220, height: 220, correctLevel: QR.CorrectLevel.M });
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

  // Mock wallet approve/decline (sandbox only) → flips the session server-side.
  async function confirmMock(approve: boolean) {
    const res = await fetch(`/api/checkout/sessions/${token}/confirm-mock`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ approve }),
    });
    if (!res.ok) return;
    const env = await res.json();
    setView(env.data as CheckoutView);
  }

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

  // Wallet mock: simulate the PSP approve/decline screen (sandbox only).
  if (walletAwaiting) {
    return (
      <div className="max-w-md w-full rounded-xl2 bg-paycore-surface p-6 mt-10 text-center space-y-4">
        <p className="text-paycore-muted text-sm">{view.merchant_name}</p>
        <h1 className="text-lg font-semibold">{METHOD_LABEL[view.selected_method ?? ""] ?? view.selected_method}</h1>
        <p className="text-2xl font-bold">{formatMoney(view.amount_minor, view.currency)}</p>
        <p className="text-xs rounded-lg bg-yellow-500/10 text-yellow-300 px-3 py-2">
          โหมดทดสอบ (Sandbox) — จำลองหน้าอนุมัติของผู้ให้บริการ
        </p>
        <div className="flex gap-2">
          <button onClick={() => confirmMock(true)} className="flex-1 rounded-lg bg-paycore-primary hover:bg-paycore-primaryHover text-white font-medium px-4 py-2">
            อนุมัติการชำระเงิน
          </button>
          <button onClick={() => confirmMock(false)} className="flex-1 rounded-lg border border-white/15 text-paycore-muted px-4 py-2">
            ปฏิเสธ
          </button>
        </div>
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

> This replaces the Phase 3 `PayPromptPayButton` with the generic `PayMethodButton` (used for both promptpay and wallets) and adds the `walletAwaiting` branch to `CheckoutStatusView`. The QR `correctLevel` fix (`QR.CorrectLevel.M`) and the polling loop are preserved verbatim from Phase 3.

- [ ] **Step 2: Build**

Run: `cd web-app && npm run build`
Expected: build succeeds; `/pay/[publicId]` compiles. No new dependency added (check `git diff web-app/package.json` is empty).

- [ ] **Step 3: Commit**

```bash
git add web-app/components/CheckoutClient.tsx
git commit -m "feat(web): wallet method chips + inline sandbox approve/decline"
```

---

### Task 6: End-to-end smoke (sandbox) + docs note

**Files:**
- Modify: `README.md`

**Interfaces:** none (integration verification of Tasks 1–5).

- [ ] **Step 1: Run the full automated suite**

Run: `go build ./... && go test ./...` and `cd web-app && npm run build`
Expected: all PASS / build succeeds. This is the real gate; the manual smoke below is best-effort.

- [ ] **Step 2: Manual backend smoke — wallet approve (sandbox)**

Start the backend (sandbox):
```bash
SANDBOX_MODE=true MIGRATE_ON_BOOT=true JWT_SECRET=dev-secret-dev-secret-dev-secret-xx \
  PUBLIC_BASE_URL=http://localhost:3000 PROMPTPAY_MOBILE_NO=0812345678 \
  QR_WEBHOOK_SECRET=dev-secret-dev-secret-dev-secret-xx make run
```
In another shell:
```bash
# 1. Dashboard login (dev) + create a link allowing a wallet method.
curl -s -X POST localhost:8080/v1/auth/dev-login -c /tmp/pc.cookies >/dev/null
LINK=$(curl -s -X POST localhost:8080/v1/payment-links -b /tmp/pc.cookies \
  -H 'Content-Type: application/json' \
  -d '{"title":"Coffee","amount_minor":5000,"allowed_methods":["truemoney","promptpay","card"]}')
PUBLIC_ID=$(echo "$LINK" | sed -E 's/.*"public_id":"([^"]+)".*/\1/')

# 2. Public: create a checkout session (no auth).
SESS=$(curl -s -X POST localhost:8080/v1/checkout/sessions \
  -H 'Content-Type: application/json' -d "{\"link\":\"$PUBLIC_ID\"}")
TOKEN=$(echo "$SESS" | sed -E 's/.*"session_token":"([^"]+)".*/\1/')

# 3. Pay with a wallet -> requires_action, selected_method=truemoney, no qr_payload.
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/pay" \
  -H 'Content-Type: application/json' -d '{"method":"truemoney"}'

# 4. Approve via confirm-mock -> paid.
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/confirm-mock" \
  -H 'Content-Type: application/json' -d '{"approve":true}'
curl -s "localhost:8080/v1/checkout/sessions/$TOKEN"   # expect "status":"paid"
```
Expected: step 3 returns `"status":"requires_action"` with `"selected_method":"truemoney"` and no `qr_payload`; step 4 returns `"status":"paid"`.

- [ ] **Step 3: Manual backend smoke — wallet decline + prod gate**

```bash
# Fresh session, decline -> failed (single_use link released for retry).
SESS=$(curl -s -X POST localhost:8080/v1/checkout/sessions -H 'Content-Type: application/json' -d "{\"link\":\"$PUBLIC_ID\"}")
TOKEN=$(echo "$SESS" | sed -E 's/.*"session_token":"([^"]+)".*/\1/')
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/pay" -H 'Content-Type: application/json' -d '{"method":"shopeepay"}' >/dev/null
curl -s -X POST "localhost:8080/v1/checkout/sessions/$TOKEN/confirm-mock" -H 'Content-Type: application/json' -d '{"approve":false}'  # expect "status":"failed"
```
Then restart with `SANDBOX_MODE=false` and confirm the wallet is refused and the route is absent:
```bash
# pay wallet -> HTTP 422 CHECKOUT_METHOD_UNAVAILABLE
# POST .../confirm-mock -> HTTP 404 (route not mounted)
```
Expected: decline returns `"status":"failed"`; in prod mode wallet pay is `422 CHECKOUT_METHOD_UNAVAILABLE` and confirm-mock is `404`. (If starting the server is impractical, skip and say so — automated tests are the gate.)

- [ ] **Step 4: Manual full-stack browser smoke (optional)**

```bash
# backend as in Step 2; frontend:
cd web-app && BACKEND_URL=http://localhost:8080 npm run dev
```
Browser: open `http://localhost:3000/pay/<PUBLIC_ID>` → pick "TrueMoney Wallet" → "ดำเนินการต่อ" → sandbox approve/decline panel appears → "อนุมัติการชำระเงิน" → page flips to "ชำระเงินสำเร็จ". Repeat with "ปฏิเสธ" → "ชำระเงินไม่สำเร็จ".

- [ ] **Step 5: README note**

Add a short subsection to `README.md` under the existing hosted-checkout docs:
```markdown
### E-wallet / redirect methods (Phase 4)

Beam-parity wallet / redirect methods — `mobile_banking`, `truemoney`,
`shopeepay`, `alipay`, `wechat`, `card_installment` — are **mock adapters**,
enabled ONLY when `SANDBOX_MODE=true`. They share one code path that moves no
money: `POST /v1/checkout/sessions/:token/pay {"method":"truemoney"}` reserves a
`single_use` link and sets the session to `requires_action`; the sandbox-only
`POST /v1/checkout/sessions/:token/confirm-mock {"approve":true|false}` then flips
it to `paid` (closing the link) or `failed` (releasing the reservation). No
`payments`/`qr_payments` row is created (they will not appear under `/payments`).

In production (`SANDBOX_MODE=false`) wallet pay returns `422
CHECKOUT_METHOD_UNAVAILABLE` and `confirm-mock` is absent (`404`); real PSP
integration is out of scope for this phase.
```

- [ ] **Step 6: Commit**

```bash
git add README.md
git commit -m "docs: phase-4 e-wallet mock methods + confirm-mock endpoint"
```

---

## Self-Review

**1. Spec coverage (design spec §4 method registry, §5.3 confirm-mock, §11 Phase 4):**
- §4 all six methods (`mobile_banking`, `truemoney`, `shopeepay`, `alipay`, `wechat`, `card_installment`) as mock adapters, sandbox-gated, "method not configured" in prod → `CheckoutSupportedMethods`/`CheckoutWalletMethods` (Task 1) + `payWallet` sandbox gate returning `ErrCheckoutMethodUnavailable` (Task 2) ✓
- §4 registry exposed to the frontend so it knows which methods exist → `DisplayMethods` (already surfaced via `CheckoutSessionView.AllowedMethods`; expanded in Task 1) ✓
- §5.3 `POST /v1/checkout/sessions/:token/confirm-mock` (SANDBOX) wallet approve/decline → service `ConfirmMock` (Task 3) + handler + sandbox-gated route (Task 4) ✓
- §11 Phase 4 "mock wallet/redirect adapter + methods + UI selector ครบ + mock-approve" → service (Tasks 2–3), API (Task 4), frontend chips + inline mock-approve (Task 5) ✓
- §6 flow (`wallet/mobile_banking: mock redirect → approve → confirm-mock → paid`; `single_use` link → paid) → `payWallet` reserve + `ConfirmMock` approve/decline (Tasks 2–3), frontend (Task 5) ✓
- §8 security (mock adapters gated by `SANDBOX_MODE`; prod → "method not configured", never silently paid) → sandbox gate in both `payWallet` and `ConfirmMock` + route absent in prod (Tasks 2–4) ✓

**2. Placeholder scan:** No `TODO`/`fill in`/"add error handling" placeholders. Every code step is complete. The frontend replacement gives the full file (no stub components). The one deliberately-empty test-helper edge (`TestPayWalletSingleUseAlreadyConsumedRefuses` reuses the existing fake's `pgx.ErrNoRows` behavior) is explained inline.

**3. Type consistency:**
- `CheckoutService.ConfirmMock(ctx, token string, approve bool) (*domain.CheckoutSessionView, error)` — defined in the interface (Task 3), implemented on `*checkoutService` (Task 3), faked in `fakeCheckoutSvc` (Task 4), called by the handler (Task 4). Signatures match.
- `payWallet(ctx, row repository.CheckoutSession, req domain.CheckoutPayRequest)` matches the `payPromptPay`/`payCard` shape it sits beside and is dispatched from `Pay` (Task 2).
- Reuses verified real signatures: `ConsumePaymentLinkIfActive(ctx, ConsumePaymentLinkIfActiveParams{ID, MerchantID})`, `ReleasePaymentLinkReservation(...same params)`, `GetPaymentLink(ctx, GetPaymentLinkParams{ID, MerchantID})`, `UpdateCheckoutSession(ctx, UpdateCheckoutSessionParams{ID, Status, SelectedMethod, PaymentID, QrPaymentID, CustomerEmail})` — all confirmed in `internal/repository/*.sql.go`. `releaseLinkReservation(ctx, *repository.PaymentLink, repository.CheckoutSession)`, `markLinkPaid(ctx, repository.CheckoutSession)`, `transition`, `strFallback`, `buildView`, `loadByToken`, `Get` are existing methods in `checkout_service.go`.
- `domain.IsWalletMethod` / `CheckoutWalletMethods` (Task 1) used in `Pay` dispatch + `ConfirmMock` (Tasks 2–3) and mirrored by the frontend `WALLET_METHODS` (Task 5).
- `domain.CheckoutConfirmMockRequest{Approve bool}` (Task 1) parsed by the handler (Task 4); frontend sends `{approve}` (Task 5).
- Error mapping unchanged: `ErrCheckoutMethodUnavailable`→422, `ErrCheckoutSessionNotFound`→404, `ErrCheckoutSessionExpired`→410 already in `middleware.go` (verified) — reused, no new sentinel.
- Router `sandbox bool` param already exists (used by `/v1/sandbox`); no `Setup` signature change; `main.go` already passes `cfg.SandboxMode` and constructs `checkoutSvc` with it.

**4. Wallet-representation decision — Option A (no new table/column).** Justification: the wallet mock moves no money and needs no PSP artifact, so the session's existing `status` + `selected_method` fully represent it; adding a `payments` row (Option B) would require a mock wallet PaymentService adapter, a new payment method type, and reconciliation with `syncStatus` for no functional gain in a mock. Verified composition: `syncStatus` guards QR polling on `selected_method == "promptpay"`, so a wallet `requires_action` session is never QR-polled (it just ages out on TTL); `ConfirmMock` is the only thing that advances it. Accepted trade-off (documented in the README note): wallet payments create no `payments` row and so do not surface under `/payments`/`/stats` — acceptable for a sandbox mock; real PSP integration (which would create ledger rows) is out of scope.
