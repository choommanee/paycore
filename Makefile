.PHONY: run dev web build test lint sec sqlc migrate-up migrate-down docker tidy

run:        ## run locally
	go run ./cmd/server

dev:        ## run the API for local dev (sandbox + auto-migrate; no docker/migrate-CLI needed)
	SANDBOX_MODE=true MIGRATE_ON_BOOT=true \
	JWT_SECRET=$${JWT_SECRET:-dev-secret-dev-secret-dev-secret-xx} \
	PUBLIC_BASE_URL=$${PUBLIC_BASE_URL:-http://localhost:3000} \
	go run ./cmd/server

web:        ## run the Next.js dashboard + checkout (installs deps on first run)
	cd web-app && { [ -d node_modules ] || npm install; } && \
	BACKEND_URL=$${BACKEND_URL:-http://localhost:8080} npm run dev

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race -cover

lint:
	golangci-lint run ./...

sec:        ## static security scan (must be 0 HIGH / 0 CRITICAL)
	gosec -quiet -conf .gosec.json ./...

sqlc:       ## generate type-safe DB code
	sqlc generate

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

docker:
	docker compose -f deploy/docker-compose.yml up --build

tidy:
	go mod tidy
