# ---- build ----
FROM golang:1.25-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
# Static binary. The migrations/*.sql files are embedded via //go:embed at this
# step, so no SQL ships in the runtime image.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---- run ----
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
# Static UIs (landing, /signup, dashboard, admin, checkout) are served from disk
# via WEB_DIR (default ./web -> /app/web). They must ship in the runtime image,
# otherwise the pages 404 on a PaaS where only the binary would otherwise exist.
COPY --from=build --chown=nonroot:nonroot /src/web /app/web
USER nonroot:nonroot
# 8080 is the local default. On a PaaS (Railway/Fly/Heroku) the platform injects
# $PORT and the server binds 0.0.0.0:$PORT instead (see internal/config). EXPOSE
# is documentation only — Railway routes to whatever $PORT the process listens on.
EXPOSE 8080
ENTRYPOINT ["/app/server"]
