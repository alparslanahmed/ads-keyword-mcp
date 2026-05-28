.PHONY: run run-http test lint build migrate research tidy fmt clean

# Default product idea for `make research` (override: make research IDEA="...").
IDEA ?= portable espresso maker for campers
COUNTRY ?= Turkey
LANGUAGE ?= Turkish

BIN_DIR := bin

# Run the MCP server over stdio (primary transport).
run:
	go run ./cmd/keyword-market-mcp --stdio

# Run the Fiber HTTP debug server.
run-http:
	go run ./cmd/keyword-market-mcp --http

# Run all tests.
test:
	go test ./...

# Static analysis + format check.
lint:
	go vet ./...
	@test -z "$$(gofmt -l .)" || (echo "gofmt needed on:"; gofmt -l .; exit 1)

# Build both binaries into ./bin.
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/keyword-market-mcp ./cmd/keyword-market-mcp
	go build -o $(BIN_DIR)/research ./cmd/research

# Apply database migrations (AutoMigrate) and exit.
migrate:
	go run ./cmd/keyword-market-mcp --migrate

# Validate a product idea via the CLI (uses mock data without credentials).
research:
	GOOGLE_ADS_MOCK=true go run ./cmd/research --idea "$(IDEA)" --country "$(COUNTRY)" --language "$(LANGUAGE)"

tidy:
	go mod tidy

fmt:
	gofmt -w .

clean:
	rm -rf $(BIN_DIR) *.db *.db-shm *.db-wal
