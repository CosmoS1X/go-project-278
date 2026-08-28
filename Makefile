.PHONY: tidy fmt build run clean test test-race test-bench test-coverage show-coverage lint lint-fix dev vuln

tidy:
	go mod tidy

fmt:
	golangci-lint fmt

build:
	go build -o ./bin/server ./cmd/server

run: build
	./bin/server

clean:
	rm -rf ./bin coverage.out

test:
	go test -v ./...

test-race:
	go test -race -v ./...

test-bench:
	go test -bench=. -benchmem

test-coverage:
	go test -v -cover -coverprofile=coverage.out ./...

show-coverage:
	go tool cover -html=coverage.out

lint: fmt
	golangci-lint run

lint-fix:
	golangci-lint run --fix

dev:
	air

vuln:
	govulncheck ./...
