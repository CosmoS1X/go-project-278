.PHONY: tidy fmt build run clean test test-race test-bench test-coverage coverage-func coverage-html coverage-show lint lint-fix dev dev-backend dev-frontend vuln sqlc-generate migrate-up migrate-down

MIGRATIONS_DIR := db/migrations
-include .env

sqlc-generate:
	cd internal/storage/sqlc && go tool sqlc generate

migrate-up:
	go tool goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" up

migrate-down:
	go tool goose -dir $(MIGRATIONS_DIR) postgres "$(DATABASE_URL)" down

tidy:
	go mod tidy

fmt:
	golangci-lint fmt

build:
	go build -o ./bin/server .

run: build
	./bin/server

clean:
	rm -rf ./bin/server coverage.out coverage.html

test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-bench:
	go test -bench=. -benchmem

test-coverage:
	go test -v -coverpkg=./... -coverprofile=coverage.out ./...

coverage-func: test-coverage
	go tool cover -func=coverage.out

coverage-html: test-coverage
	go tool cover -html=coverage.out -o coverage.html

coverage-show: test-coverage
	go tool cover -html=coverage.out

lint: fmt
	golangci-lint run

lint-fix:
	golangci-lint run --fix

dev:
	concurrently "make dev-backend" "make dev-frontend"

dev-backend:
	air

dev-frontend:
	cd frontend && npm exec start-hexlet-url-shortener-frontend

vuln:
	govulncheck ./...
