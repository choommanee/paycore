# PayCore Merchant Checkout — Design Spec

**Date:** 2026-07-24
**Status:** Approved (design), pending implementation plan
**Goal:** เพิ่ม *merchant product layer* แบบ Beam Checkout วางบน PayCore acquiring backend เดิม — ให้ร้านค้าล็อกอิน สร้าง payment link เลือกช่องทางจ่าย และรับเงินผ่านหน้า hosted checkout ที่สวยงามในแบรนด์ PayCore

---

## 1. Decisions (locked)

| หัวข้อ | เลือก |
|--------|-------|
| ขอบเขต | Merchant product layer บน PayCore เดิม (ไม่เริ่มใหม่) |
| ฟีเจอร์ MVP | Payment Links + Hosted Checkout + Merchant Dashboard + หลายช่องทางจ่าย |
| Login | **OAuth (Google OIDC)** + dev-login fallback (เฉพาะ `SANDBOX_MODE=true`) |
| ช่องทางจ่าย | ครบแบบ Beam: card, PromptPay QR, mobile banking, TrueMoney, ShopeePay, Alipay, WeChat, card installment (e-wallet/redirect ทั้งหมดเป็น **mock adapter**) |
| ดีไซน์ | แบรนด์ PayCore เอง (ไม่ก็อป Beam) |
| สถาปัตยกรรม | **แนวทางที่ 3** — แยก Next.js frontend คุย Go API ผ่าน proxy (ไม่มี CORS) |

---

## 2. สถาปัตยกรรม & Topology

Browser คุยกับ **origin เดียว** (Next.js) เท่านั้น → ไม่ต้องแตะ CORS, session cookie เป็น first-party

```
Browser ──▶ Next.js app (web-app/)  ──rewrites /api/* (server-side proxy)──▶ Go API (PayCore) ──▶ PostgreSQL
             - Dashboard (ร้านค้า, session-auth)
             - Hosted checkout /pay/[publicId] (public, session-token)
```

- **Monorepo:** Next app อยู่ที่ `web-app/` ในโปรเจกต์เดิม; static `web/` เดิมคงไว้ (ค่อย deprecate)
- **No CORS:** Next `rewrites` proxy ทุก `/api/*` → Go (`BACKEND_URL`); browser เห็นแค่ origin ของ Next
- **Go เป็นเจ้าของ identity/OAuth ทั้งหมด** (source of truth เดียว)
- **Deploy:** Railway 2 service — Go (เดิม) + Next (Node ใหม่); Next รู้จัก Go ผ่าน internal URL
- **Frontend stack:** Next.js App Router + TypeScript + Tailwind (map design token จาก `paycore.css` เป็น Tailwind theme)

---

## 3. Data Model (migration ใหม่)

### 3.1 `merchant_users` — human identity (ล็อกอินได้)
```
id             UUID PK
merchant_id    UUID FK → merchants(id)
email          TEXT UNIQUE NOT NULL
name           TEXT
avatar_url     TEXT
oauth_provider TEXT NOT NULL          -- 'google' | 'dev'
oauth_subject  TEXT NOT NULL          -- provider 'sub'
role           TEXT NOT NULL DEFAULT 'owner'
last_login_at  TIMESTAMPTZ
created_at, updated_at TIMESTAMPTZ
UNIQUE(oauth_provider, oauth_subject)
```
Google login ครั้งแรกที่ไม่มี user → สร้าง `merchants` + `merchant_users` ให้อัตโนมัติ (login = signup). Merchant เดิมยังไม่มี email → เพิ่ม user record ผูกเข้าไป

### 3.2 `payment_links` — ลิงก์รับเงินที่แชร์ได้
```
id, merchant_id FK
public_id       TEXT UNIQUE            -- slug ใน URL (base62 สุ่ม)
title, description TEXT
amount_minor    BIGINT NOT NULL        -- integer สตางค์ (คงคอนเวนชันเดิม)
currency        TEXT NOT NULL DEFAULT 'THB'
allowed_methods TEXT[]                 -- ว่าง = อนุญาตทุก method ที่เปิด
link_type       TEXT DEFAULT 'single_use'  -- single_use | reusable
status          TEXT DEFAULT 'active'      -- active | paid | expired | disabled
reference, image_url TEXT
expires_at      TIMESTAMPTZ
created_by      UUID FK → merchant_users
created_at, updated_at
```

### 3.3 `checkout_sessions` — ความพยายามจ่าย 1 ครั้ง (ขับหน้า hosted)
```
id, merchant_id FK
payment_link_id UUID FK NULL           -- NULL = session ที่สร้างผ่าน API ตรง
session_token_hash TEXT UNIQUE         -- เก็บ SHA-256 hash (เหมือน api_key_hash); token ดิบให้ browser
amount_minor, currency
status          TEXT                   -- open | processing | requires_action | paid | failed | expired
selected_method TEXT
payment_id      UUID FK NULL → payments
qr_payment_id   UUID FK NULL → qr_payments
customer_email  TEXT
return_url      TEXT
expires_at      TIMESTAMPTZ NOT NULL   -- อายุสั้น (เช่น 30 นาที)
created_at, updated_at
```

---

## 4. Payment Method Registry (ใน code, ไม่ใช่ table)

แต่ละ method map เข้า "family" ที่รู้วิธี initiate + confirm:

| method | family | ทำงานจริงผ่าน |
|--------|--------|----------------|
| `card` | card | payments authorize/capture + 3DS (adapter เดิม) |
| `card_installment` | card | payments (mock installment) |
| `promptpay` | qr | qr_payments (adapter เดิม) |
| `mobile_banking` | redirect | mock redirect adapter |
| `truemoney` / `shopeepay` / `alipay` / `wechat` | wallet | **mock wallet adapter ตัวเดียว** (ต่อยอด pattern "sandbox bank") |

- Mock adapter ทั้งหมด gated ด้วย `SANDBOX_MODE`; ใน prod จะเป็นจุดต่อ PSP จริง (นอก scope นี้ → คืน "method not configured")
- Registry เปิดเผยผ่าน endpoint ให้ frontend query ว่ามี method ไหนบ้าง

---

## 5. Backend API (endpoint ใหม่)

### 5.1 Auth
| Method | Path | หน้าที่ |
|--------|------|--------|
| GET | `/v1/auth/google/start` | 302 → Google (state + PKCE) |
| GET | `/v1/auth/google/callback` | verify → upsert user/merchant → set `pc_session` cookie → 302 dashboard |
| POST | `/v1/auth/dev-login` | (SANDBOX only) สร้าง/คืน dev session |
| POST | `/v1/auth/logout` | ล้าง cookie |
| GET | `/v1/auth/me` | user + merchant ปัจจุบัน (session-auth) |

- **`pc_session`**: JWT เซ็นด้วย `JWT_SECRET` (มี config อยู่แล้ว), httpOnly + Secure(prod) + SameSite=Lax
- **`sessionAuth` middleware** ใหม่ (parallel กับ API-key); auth middleware กลางลองอ่าน API key **หรือ** session cookie → resolve merchant context เดียวกัน → route ที่ merchant-scoped อยู่แล้ว (`/me`, `/stats`, `/payments`, `/settlements`, `/disputes`) รับได้ทั้งสองแบบ

### 5.2 Dashboard (session-auth; reuse service เดิมที่ scope merchant อยู่แล้ว)
| Method | Path | หน้าที่ |
|--------|------|--------|
| POST/GET | `/v1/payment-links` | สร้าง / list link |
| GET | `/v1/payment-links/:id` | รายละเอียด link + payments ที่มาจาก link |
| PATCH | `/v1/payment-links/:id` | disable / แก้สถานะ |
| (reuse) | `/v1/me`, `/me/rotate-key`, `/me/webhook`, `/stats`, `/settlements`, `/payments*` | ของเดิม |

### 5.3 Public Checkout (session-token; ไม่มี cookie/API key — ตรงกับ flow ที่ README ออกแบบไว้)
| Method | Path | หน้าที่ |
|--------|------|--------|
| POST | `/v1/checkout/sessions` | สร้าง session จาก `{ link: public_id }` → คืน `session_token` + display (ชื่อร้าน/โลโก้/ยอด/methods). Rate-limit ต่อ IP |
| GET | `/v1/checkout/sessions/:token` | สถานะ session + display (โหลดหน้า/poll) |
| POST | `/v1/checkout/sessions/:token/pay` | `{ method, ...data }` → initiate: card→next_action(3ds/success); promptpay→qr payload+poll; wallet→mock redirect+poll |
| POST | `/v1/checkout/sessions/:token/confirm-mock` | (SANDBOX) จำลอง wallet approve/decline |

- QR ยังยืนยันผ่าน `/v1/webhooks/qr` (bank callback, HMAC) + sandbox pay endpoint เดิม

---

## 6. Checkout Payment Flow (sequence)

```
1. ร้านค้า (dashboard) POST /v1/payment-links → ได้ public_id → แชร์ URL https://app/pay/<public_id>
2. ลูกค้าเปิด /pay/<public_id> → Next server เรียก POST /v1/checkout/sessions {link} → ได้ session_token (ถือใน memory หน้าเว็บ)
3. หน้า checkout แสดง ชื่อร้าน/ยอด/รายการ + method selector (จาก registry ∩ allowed_methods)
4. ลูกค้าเลือก method + กรอกข้อมูล → POST /pay
   - card: authorize → ถ้า 3DS ต้อง redirect → return → capture → paid
   - promptpay: สร้าง QR → หน้าเว็บ render + poll GET session จน paid
   - wallet/mobile_banking: mock redirect → หน้า approve → confirm-mock → paid
5. session.status = paid → payment_links.status = paid (ถ้า single_use) → หน้า success + return_url
6. ร้านค้าเห็น transaction ใน dashboard (reuse /payments, /stats)
```

---

## 7. Frontend (Next.js) — โครง route

**Dashboard (session):** `/login` · `/` (stats+recent) · `/transactions` (+detail+refund) · `/links` (+create modal) · `/links/[id]` (copy URL, QR ของ URL, payments) · `/settings` (API keys rotate, webhook, business profile/logo)

**Hosted checkout (public):** `/pay/[publicId]` · `/pay/[publicId]/mock-approve` (sandbox wallet) · success state

---

## 8. Security

- **Session JWT:** httpOnly, Secure(prod), SameSite=Lax, อายุจำกัด
- **CSRF:** dashboard mutation ใช้ double-submit CSRF token (SameSite=Lax ยังต้องกันเพิ่ม)
- **OAuth:** state param + PKCE กัน CSRF/replay
- **Public checkout:** rate-limit ต่อ IP; session token = single-payment + TTL สั้น + เก็บเป็น hash; **ไม่มี secret key ใน browser** เด็ดขาด (คงหลักการเดิม)
- **Card form:** โหมด sandbox เท่านั้น + ป้ายกำกับชัด; prod จริงต้องใช้ hosted fields/tokenization (นอก scope, ระบุเป็น future)
- **Mock adapters:** gated ด้วย `SANDBOX_MODE`; prod ที่ไม่ตั้งค่า → "method not configured" (ไม่ mark paid เงียบ ๆ)

---

## 9. Testing

- **Go:** unit test service ใหม่ (auth/session, payment_links, checkout_sessions, method adapters) ตาม pattern `_test.go` เดิม; integration test flow checkout ครบ (สร้าง link → session → จ่ายแต่ละ family → paid)
- **Next:** เทสเบา ๆ (component/e2e หลัก ๆ ของ checkout); โฟกัส backend test ตามคอนเวนชันเดิม

---

## 10. Deployment / config

- เพิ่ม Railway service ที่ 2 (Next); env: `BACKEND_URL` (internal), `GOOGLE_CLIENT_ID/SECRET`, `OAUTH_REDIRECT_BASE`, cookie settings
- อัปเดต `.env.example`, `railway.toml`, README

---

## 11. Phasing (แต่ละ phase = vertical slice ที่ใช้งานได้ + มี 1 แผน implementation)

- **Phase 1 — Human auth:** `merchant_users` migration + Google OIDC + dev-login + `pc_session` + `sessionAuth` + `/auth/me`. Next: `/login` + shell + guard. *เสร็จ = ล็อกอินเข้า dashboard เปล่าได้*
- **Phase 2 — Payment links:** `payment_links` migration + CRUD endpoints + dashboard `/links`, `/links/[id]`. *เสร็จ = สร้าง/แชร์ลิงก์ได้*
- **Phase 3 — Hosted checkout (card + PromptPay):** `checkout_sessions` + public checkout endpoints + method registry (card, promptpay) + `/pay/[publicId]`. *เสร็จ = จ่ายจริง end-to-end 2 ช่องทาง*
- **Phase 4 — E-wallet methods:** mock wallet/redirect adapter + methods (mobile_banking, truemoney, shopeepay, alipay, wechat, installment) + UI selector ครบ + mock-approve. *เสร็จ = ครบทุกช่องทางแบบ Beam*
- **Phase 5 — Dashboard polish:** transactions list/detail + refund UI + stats + settings (API keys/webhook/profile). *เสร็จ = merchant ops ครบ*

**เริ่ม Phase 1 ก่อน**
