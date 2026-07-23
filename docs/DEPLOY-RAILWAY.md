# Deploy to Railway (TEST / Sandbox) — คู่มือการดีพลอย

Deploy PayCore to [Railway](https://railway.com) as a **public TEST / sandbox**
environment where people can self-sign-up and try it. TH + EN.

> ⚠️ **TEST-grade only.** Public self-signup here has **no email verification, no
> KYC, and no captcha** — it exists so people can try the sandbox instantly. It
> processes **no real money**. For production you MUST add email verification,
> KYC/AML, and a captcha on `POST /v1/merchants` before opening signup.
>
> ⚠️ **สำหรับทดสอบเท่านั้น** การสมัครแบบ self-service ไม่มีการยืนยันอีเมล ไม่มี KYC
> และไม่มี captcha — และไม่มีการเคลื่อนไหวของเงินจริง หากจะใช้งานจริง ต้องเพิ่ม
> email verification, KYC/AML และ captcha ก่อนเปิดให้สมัคร

---

## 1. Prerequisites / สิ่งที่ต้องมี

- A Railway account — https://railway.com
- Railway CLI: `npm i -g @railway/cli` (or `brew install railway`)
- This repo checked out locally.

```bash
railway login          # opens the browser to authenticate
```

---

## 2. Create the project / สร้างโปรเจกต์

From the repo root:

```bash
railway init           # creates a new Railway project (pick a name, e.g. paycore-sandbox)
```

Railway detects the `Dockerfile` and `railway.toml` at the repo root. The build
uses the existing distroless Dockerfile; deploy settings (healthcheck `/readyz`,
`ON_FAILURE` restart) come from `railway.toml`.

---

## 3. Add the Postgres plugin / เพิ่มฐานข้อมูล Postgres

```bash
railway add --database postgres
```

Or in the dashboard: **New → Database → Add PostgreSQL**.

Railway provisions Postgres and exposes a `DATABASE_URL` variable **on the
Postgres service**. Reference it from the app service with Railway's variable
reference syntax so the app always points at the managed DB:

```
DATABASE_URL=${{Postgres.DATABASE_URL}}
```

> Railway variable references use `${{ServiceName.VAR}}`. If your Postgres
> service is named `Postgres`, the reference above works as-is. Set it on the
> **app** service (see next step).
>
> การอ้างอิงตัวแปรของ Railway ใช้รูปแบบ `${{ServiceName.VAR}}` ตั้งค่าตัวแปรนี้ไว้ที่
> **เซอร์วิสของแอป** (ไม่ใช่ที่เซอร์วิส Postgres)

---

## 4. Set variables / ตั้งค่าตัวแปรสภาพแวดล้อม

Set these on the **app** service. Secrets must be **≥ 32 bytes** and **must not**
be the `change-me-in-prod` placeholder — the process fails fast at boot otherwise
(see `internal/config/validate`).

Generate strong secrets:

```bash
openssl rand -hex 32   # 64 hex chars — run once per secret
```

Set them via CLI (or paste into the dashboard **Variables** tab):

```bash
railway variables \
  --set "ENV=production" \
  --set "DATABASE_URL=${{Postgres.DATABASE_URL}}" \
  --set "MIGRATE_ON_BOOT=true" \
  --set "TLS_TERMINATED_UPSTREAM=true" \
  --set "WEB_DIR=./web" \
  --set "ADMIN_API_KEY=$(openssl rand -hex 32)" \
  --set "JWT_SECRET=$(openssl rand -hex 32)" \
  --set "QR_WEBHOOK_SECRET=$(openssl rand -hex 32)" \
  --set "WEBHOOK_SIGNING_SECRET=$(openssl rand -hex 32)" \
  --set "SIGNUP_RATE_LIMIT_PER_HOUR=5"
```

| Variable | ค่า / Value | Notes |
|---|---|---|
| `ENV` | `production` | Turns on the fail-fast secret/TLS checks. |
| `DATABASE_URL` | `${{Postgres.DATABASE_URL}}` | Reference the Railway Postgres plugin. |
| `MIGRATE_ON_BOOT` | `true` | Runs embedded migrations in-process at boot. |
| `TLS_TERMINATED_UPSTREAM` | `true` | Acknowledge Railway terminates TLS at the edge (see below). |
| `WEB_DIR` | `./web` | Serves the landing / signup / dashboard / checkout UIs. |
| `ADMIN_API_KEY` | 32+ bytes | Gates `/v1/admin`. Keep it secret. |
| `JWT_SECRET` | 32+ bytes | Required strong in prod. |
| `QR_WEBHOOK_SECRET` | 32+ bytes | Verifies inbound QR webhooks. |
| `WEBHOOK_SIGNING_SECRET` | 32+ bytes | Signs outbound webhook deliveries. |
| `SIGNUP_RATE_LIMIT_PER_HOUR` | `5` | Per-IP cap on public signups. |
| `CORS_ALLOW_ORIGINS` | *(optional)* | Only if the checkout is hosted on a **different** origin. |
| `PORT` | *(injected)* | **Do not set** — Railway injects it. The app binds `0.0.0.0:$PORT`. |

### TLS / การเข้ารหัส
Railway terminates TLS at its edge and forwards plain HTTP to your container, so
running plaintext HTTP **inside** the container is expected and fine. Because
`ENV=production` refuses to start over plaintext unless TLS is acknowledged, set
`TLS_TERMINATED_UPSTREAM=true`. Do **not** set `TLS_CERT_FILE`/`TLS_KEY_FILE` —
Railway handles the certificate.

Railway จะจัดการ TLS ที่ขอบเครือข่าย (edge) แล้วส่งต่อเป็น HTTP ธรรมดาเข้าคอนเทนเนอร์
จึงต้องตั้ง `TLS_TERMINATED_UPSTREAM=true` เพื่อยืนยันว่ามี upstream ทำ TLS ให้แล้ว

### CORS
The UIs are served same-origin from `WEB_DIR`, so no CORS config is needed unless
you host the checkout on a **separate** domain. In that case set
`CORS_ALLOW_ORIGINS=https://your-checkout-host` (comma-separated; never `*`).

---

## 5. Deploy / ดีพลอย

```bash
railway up             # builds the Dockerfile and deploys
```

On boot the server:
1. Runs all pending migrations in-process (`MIGRATE_ON_BOOT=true`) and logs the
   applied `schema_version`.
2. Opens the Postgres pool and starts serving on `0.0.0.0:$PORT`.
3. Answers `/readyz` with `200` once the DB pings — Railway's healthcheck waits
   for this before routing traffic.

---

## 6. Get the public domain / รับโดเมนสาธารณะ

```bash
railway domain         # generates (or prints) a *.up.railway.app domain
```

Then visit:
- `https://<your-app>.up.railway.app/`         — landing page
- `https://<your-app>.up.railway.app/signup`   — self-service sandbox signup
- `https://<your-app>.up.railway.app/dashboard`— merchant dashboard (login with API key)
- `https://<your-app>.up.railway.app/readyz`   — readiness probe (should be `200`)

---

## 7. Verify / ตรวจสอบ

```bash
BASE=https://<your-app>.up.railway.app

curl -s "$BASE/readyz"                       # -> 200
curl -s "$BASE/signup/" -o /dev/null -w '%{http_code}\n'   # -> 200

# Create a sandbox merchant (returns an api_key ONCE):
curl -s -X POST "$BASE/v1/merchants" \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test Store","mcc":"5999","settlement_currency":"THB"}'
```

The 6th rapid `POST /v1/merchants` from the same IP returns HTTP `429` with the
standard error envelope (`"code":"RATE_LIMITED"`) — that's the signup limiter
working (`SIGNUP_RATE_LIMIT_PER_HOUR`).

---

## Troubleshooting / แก้ปัญหา

- **Boot fails immediately** — a secret is missing / too short / still the
  placeholder. Check the deploy logs; `internal/config` names the offending var.
- **`/readyz` stays 503** — the app can't reach Postgres. Confirm
  `DATABASE_URL=${{Postgres.DATABASE_URL}}` on the **app** service and that the
  Postgres plugin is provisioned.
- **Migrations didn't run** — confirm `MIGRATE_ON_BOOT=true`; the boot log prints
  `migrations applied` with the `schema_version`.
