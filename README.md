# Payment Gateway (Go / Fiber)

โครงโปรเจ็ค **Full Acquiring Payment Gateway** ตาม Clean Architecture ออกแบบให้สอดคล้องกับ
พ.ร.บ. ระบบการชำระเงิน พ.ศ. 2560 และ PCI-DSS เพื่อยื่นขอใบอนุญาตจาก ธปท. ได้

> สถานะ: ชั้น service ถูก implement เต็มแล้ว (authorize / capture / refund / void / 3DS /
> QR / disputes / settlement / reconciliation / outbound webhook). Mock adapters ของ
> acquirer / 3DS / QR provider และ tokenization vault พร้อมใช้งานสำหรับ sandbox/dev.

## เอกสาร
- [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) — สถาปัตยกรรม, flow, data model, PCI/security
- [`docs/QR-PAYMENT.md`](docs/QR-PAYMENT.md) — QR payment (PromptPay/Thai QR EMVCo), flow, security
- [`docs/COMPLIANCE-TH.md`](docs/COMPLIANCE-TH.md) — กฎหมายไทย, ประเภทใบอนุญาต, ทุน, ขั้นตอนยื่นขอ
- [`docs/DIAGRAMS.md`](docs/DIAGRAMS.md) — context / component / sequence / state / ERD (mermaid)
- [`docs/ROADMAP.md`](docs/ROADMAP.md) — เฟส, timeline, ทีม, ประมาณการต้นทุน

## Tech Stack
Go 1.24 · Fiber v2 · PostgreSQL 16 · sqlc · pgx/v5 · zerolog · Viper · validator · Docker · GitLab CI

## โครงสร้าง
```
payment-gateway/
├── cmd/server/main.go            # entry point + graceful shutdown
├── internal/
│   ├── config/                   # Viper config loader
│   ├── middleware/               # auth, logger, idempotency, recovery, error handler
│   ├── handler/                  # HTTP handlers (Fiber)
│   ├── service/                  # business logic (payment state machine)
│   ├── repository/               # data access (sqlc) + queries/*.sql
│   ├── domain/                   # models, DTOs, errors, response envelope
│   ├── router/                   # route registration
│   ├── external/                 # acquirer + 3DS interfaces
│   └── pkg/                      # money, idempotency, crypto(vault), threeds
├── migrations/                   # PostgreSQL schema (up/down)
├── deploy/docker-compose.yml
├── docs/                         # เอกสารออกแบบ (ด้านบน)
├── sqlc.yaml · Dockerfile · Makefile · .gitlab-ci.yml · .env.example
```

## API (v1)
| Method | Path | หน้าที่ | Auth |
|--------|------|--------|------|
| POST | `/v1/merchants` | onboard merchant — คืน API key ครั้งเดียว | none |
| GET  | `/v1/merchants/{id}` | ดูโปรไฟล์ merchant (ตัวเองเท่านั้น) | API key |
| POST | `/v1/payments` | สร้าง payment (authorize/charge) — ต้องมี `Idempotency-Key` | API key |
| GET  | `/v1/payments` | list payments (paginated: `limit`≤200, `offset`) | API key |
| GET  | `/v1/payments/{id}` | ดูสถานะ | API key |
| POST | `/v1/payments/{id}/capture` | capture (รองรับ partial) — ต้องมี `Idempotency-Key` | API key |
| POST | `/v1/payments/{id}/refund` | refund (รองรับ partial) — ต้องมี `Idempotency-Key` | API key |
| POST | `/v1/payments/{id}/void` | ยกเลิก authorization | API key |
| POST | `/v1/payments/{id}/3ds/return` | callback หลัง 3DS | API key |
| POST | `/v1/payments/{id}/disputes` | เปิด dispute / chargeback | API key |
| GET  | `/v1/payments/{id}/disputes` | list disputes ของ payment | API key |
| POST | `/v1/qr-payments` | สร้าง QR (PromptPay/card/cross-border) | API key |
| GET  | `/v1/qr-payments/{id}` | poll สถานะ QR | API key |
| POST | `/v1/webhooks/qr` | callback ยืนยันจ่าย QR จากธนาคาร/PSP (verify HMAC) | HMAC sig |
| POST | `/v1/checkout/sessions` | สร้าง hosted-checkout session จาก payment link | none (IP rate-limited) |
| GET  | `/v1/checkout/sessions/{token}` | poll สถานะ checkout session | none (session token) |
| POST | `/v1/checkout/sessions/{token}/pay` | จ่ายเงินผ่าน checkout session (card/promptpay) | none (session token) |
| GET  | `/healthz` · `/readyz` | probes | none |

### Authentication (API key)
- Onboard ด้วย `POST /v1/merchants` → response คืน `api_key` (`sk_live_...`) **ครั้งเดียว**; server เก็บเฉพาะ SHA-256 hash.
- ส่ง key ในทุก money route ผ่าน `Authorization: Bearer sk_live_...` หรือ `X-API-Key: sk_live_...`.
- ทุก request ถูก scope ด้วย merchant ที่ auth แล้ว; `merchant_id` ใน body ถูกละเว้นเสมอ. `GET /v1/merchants/{id}` อ่านได้เฉพาะ merchant ของตัวเอง (มิฉะนั้นคืน `404 MERCHANT_NOT_FOUND` เพื่อไม่ให้เป็น existence oracle).

### Idempotency
- Money-moving POST (create / capture / refund) บังคับ header `Idempotency-Key`.
- ขอบเขต key = (merchant + key). ยิงซ้ำด้วย key เดิม + payload เดิม → คืนผลลัพธ์เดิม (safe retry).
- key เดิมแต่ payload ต่างกัน (fingerprint mismatch) → `409 IDEMPOTENCY_KEY_REUSED`.

### Error envelope & codes
ทุก response ใช้ envelope เดียว: `{ success, code, message, data?, request_id, timestamp }`.
Validation ล้มเหลวคืน `400 VALIDATION_ERROR` พร้อม `data.errors: [{ field, code, message }]` (field เป็นชื่อ JSON, code เป็น validator tag — machine-parseable).

| Code | HTTP | ความหมาย |
|------|------|----------|
| `VALIDATION_ERROR` | 400 | request ไม่ผ่าน validation (ดู `data.errors`) |
| `INVALID_BODY` / `INVALID_ID` | 400 | body/param parse ไม่ได้ |
| `INVALID_REQUEST` | 400 | คำขอไม่ถูกต้อง |
| `UNAUTHORIZED` | 401 | ไม่มี/ผิด API key |
| `FORBIDDEN` | 403 | ไม่มีสิทธิ์ |
| `INSUFFICIENT_FUNDS` / `CARD_DECLINED` | 402 | ถูกปฏิเสธ |
| `PAYMENT_NOT_FOUND` / `MERCHANT_NOT_FOUND` / `DISPUTE_NOT_FOUND` | 404 | ไม่พบ |
| `DUPLICATE_REQUEST` / `IDEMPOTENCY_KEY_REUSED` / `THREE_DS_REQUIRED` / `INVALID_STATE` | 409 | conflict (รวมถึง state-transition ที่ไม่ถูกต้อง เช่น void หลัง capture) |
| `REFUND_EXCEEDS_CAPTURED` | 422 | refund เกินยอด capture |
| `RATE_LIMITED` | 429 | เกิน per-merchant rate limit |
| `INTERNAL_ERROR` | 500 | ข้อผิดพลาดภายใน |

### Hosted checkout & auth model
`web/checkout/*.html` เป็น **demo**. secret API key (`sk_live_...`) **ห้ามอยู่ใน browser JS** เด็ดขาด.
Money routes อยู่หลัง API-key auth ดังนั้น browser ที่ไม่มี key จะได้ `401`. flow ที่ถูกต้อง:
1. server สร้าง checkout session → ออก **short-lived / single-payment session token** ที่ปลอดภัยพอจะส่งเข้า browser
2. browser เรียก unauthenticated checkout-session route group ด้วย session token นั้น (ไม่ใช่ secret key)
3. server (ถือ secret key) เป็นผู้ settle / capture

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

Config: `PUBLIC_BASE_URL` (base URL used to build the checkout link shown to
merchants) and `CHECKOUT_RATE_LIMIT_PER_MIN` (default 30 — IP-keyed limit on
`POST /v1/checkout/sessions` only; the other two checkout routes are unlimited).

### Performance & rate limiting
- Rate limit เป็น **per-merchant** (keyed `m:<merchant_id>`) จาก `RATE_LIMIT_PER_SEC` (default 600/s); health/metrics/webhook ไม่ถูก limit. ไม่มี global per-IP limiter.
- **Load test ต้องรู้:** harness แบบ single-source ยิงจาก merchant id เดียว → budget ทั้งหมดคือของ merchant นั้น. ตั้ง `RATE_LIMIT_PER_SEC` ≥ concurrency ที่ต้องการ (เช่น ≥300/s สำหรับ 167 TPS target) มิฉะนั้น 429 flood จะทำให้ err rate พุ่งด้วยเหตุผลผิด. สำหรับ staging/load ตั้งค่านี้สูงกว่า prod ceiling ได้.
- pgx pool: `DB_MAX_CONNS` (default 50) / `DB_MIN_CONNS` (default 10).

## เริ่มต้น
```bash
cp .env.example .env
make docker          # รัน postgres + app ด้วย docker compose
# หรือ local:
make tidy && make sqlc && make migrate-up && make run
```

## หมายเหตุความปลอดภัย
- ระบบหลัก **ไม่เก็บ/ไม่ log** full PAN, CVV, PIN — เก็บได้แค่ `card_last4`
- เงินเก็บเป็น **integer minor units (สตางค์)** เสมอ
- ทุก money endpoint บังคับ **Idempotency-Key**; ledger เป็น **append-only**
- **Transport (prod):** เมื่อ `ENV=production` server จะ **ปฏิเสธการ start** ถ้าไม่ได้ตั้ง TLS
  (`TLS_CERT_FILE`+`TLS_KEY_FILE`) และไม่ได้ตั้ง `TLS_TERMINATED_UPSTREAM=true` เพื่อยืนยันว่ามี
  upstream (LB/mesh) terminate TLS ให้ — ไม่ยอมให้ serve cleartext เงียบ ๆ ใน prod
