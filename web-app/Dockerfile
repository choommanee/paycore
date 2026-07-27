# PayCore web (Next.js) — production image for Railway.
# Multi-stage: install+build, then run the standalone server (output: "standalone").
# Runtime env: PORT (Railway injects), BACKEND_URL (internal/public URL of the Go API).

FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
# npm install (not `npm ci`): the lock carries a benign nested-picomatch quirk that
# the builder's stricter npm rejects under `npm ci`; `npm install` reconciles it.
RUN npm install --no-audit --no-fund

FROM node:20-alpine AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
# Standalone output ships its own minimal node_modules + server.js.
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
ENV PORT=3000
# Next standalone honors PORT/HOSTNAME; bind all interfaces for the platform edge.
ENV HOSTNAME=0.0.0.0
CMD ["node", "server.js"]
