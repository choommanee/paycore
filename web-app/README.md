# PayCore Web (Next.js)

Dashboard + hosted checkout. Proxies `/api/*` -> Go API `/v1/*` (see `next.config.js`),
so the browser only ever calls this origin (no CORS, first-party session cookie).

## Dev
    cp .env.example .env    # set BACKEND_URL if not localhost:8080
    npm install
    npm run dev             # http://localhost:3000

The Go API must run with SANDBOX_MODE=true to expose dev-login.
