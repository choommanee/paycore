# Deploying the full PayCore product (API + Next.js app)

The product is two services on Railway (project `paycore`, env `production`):

| Service | Dir | Build | Role |
|---------|-----|-------|------|
| `paycore-api` | repo root | `Dockerfile` (Go, distroless) | REST API `/v1/*` (+ legacy static `web/` at `/`) + Postgres |
| `paycore-web` | `web-app/` | `web-app/Dockerfile` (Next standalone) | Landing `/`, dashboard, hosted checkout `/pay/*`; proxies `/api/*` → API `/v1/*` |
| `Postgres` | — | Railway plugin | `DATABASE_URL=${{Postgres.DATABASE_URL}}` |

Customers hit **paycore-web**; it calls the API server-side via `BACKEND_URL`.

## 1. Redeploy the API with the latest code (Phases 1–5)
The API must be the current build (auth / payment-links / checkout / e-wallets / dashboard endpoints).
```bash
# repo root, Railway CLI already logged in
railway up -s paycore-api --ci
# log stream may time out — the build still runs; poll:
curl -s -o /dev/null -w "%{http_code}\n" https://paycore-api-production.up.railway.app/readyz   # want 200
```
Railway won't shift traffic until `/readyz` passes, so a bad build can't take the live service down.

**API env vars** (set in Railway, most already present): `ENV=production`, `MIGRATE_ON_BOOT=true`, `TLS_TERMINATED_UPSTREAM=true`, `SANDBOX_MODE=true` (test deploy), strong `JWT_SECRET`/`QR_WEBHOOK_SECRET`/`WEBHOOK_SIGNING_SECRET`/`ADMIN_API_KEY`, `PROMPTPAY_MOBILE_NO`. **Add for this product:** `PUBLIC_BASE_URL=<paycore-web URL>` and `OAUTH_REDIRECT_BASE=<paycore-web URL>` (so payment-link URLs and the Google OAuth callback point at the web app). `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` are optional (dev-login works in sandbox without them).

## 2. Create + deploy the Next.js web service
```bash
# once, from repo root — create the service in the project
railway add -s paycore-web
# point it at the API (public URL; internal networking also works)
railway variables -s paycore-web --set "BACKEND_URL=https://paycore-api-production.up.railway.app"
# deploy from the web-app dir (its Dockerfile + context)
cd web-app && railway up -s paycore-web --ci
```
Then in the Railway dashboard give `paycore-web` a public domain (Settings → Networking → Generate Domain). Set the API's `PUBLIC_BASE_URL`/`OAUTH_REDIRECT_BASE` to that domain and redeploy the API (step 1) so link/checkout URLs are absolute-correct.

## 3. Verify
```bash
WEB=https://<paycore-web-domain>
curl -s -o /dev/null -w "%{http_code}\n" $WEB/                 # landing 200
curl -s -X POST $WEB/api/auth/dev-login -c /tmp/c -o /dev/null # 200 (sandbox)
curl -s $WEB/api/auth/me -b /tmp/c                             # dev user JSON
```
Open `$WEB/` → Dev login → create a link → `/pay/<id>` → pay (test card `4111 1111 1111 1111` / PromptPay / e-wallet) → see it in the dashboard.

> Notes: `SANDBOX_MODE=true` exposes dev-login + the payer simulator — TEST deploys only, never a real-money deploy. Local dev uses `make dev` + `make web` (no Docker); see README.
