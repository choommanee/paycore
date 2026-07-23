.PHONY: run build test lint sec sqlc migrate-up migrate-down docker tidy

run:        ## run locally
	go run ./cmd/server

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
