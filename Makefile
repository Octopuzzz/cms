# CMS Backend Makefile
.PHONY: all build run dev test test-unit test-integration test-uat \
        swagger lint fmt deps docker-up docker-down clean help

APP_NAME   := cms-backend
VERSION    := 1.0.0
BUILD_DIR  := bin

all: build

# ---- Build ----
build:
	@echo "Building $(APP_NAME)..."
	go build -ldflags "-X main.Version=$(VERSION)" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

build-prod:
	CGO_ENABLED=1 GOOS=linux go build -a \
		-ldflags "-w -s -X main.Version=$(VERSION)" \
		-o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

# ---- Run ----
run:
	go run ./cmd/server

dev:
	air -c .air.toml

# ---- Dependencies ----
deps:
	go mod download
	go mod tidy

# ---- Tests ----
test:
	go test -v -race ./tests/...

test-unit:
	go test -v -race ./tests/unit/...

test-integration:
	go test -v -race ./tests/integration/...

test-uat:
	go test -v ./tests/uat/...

test-coverage:
	go test -v -race -coverprofile=coverage.out -coverpkg=./... ./tests/...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ---- Swagger ----
swagger:
	swag init -g cmd/server/main.go \
		--parseDependency --parseInternal \
		--output docs

# ---- Code Quality ----
lint:
	golangci-lint run --timeout=5m

fmt:
	go fmt ./...
	gofumpt -w .

vet:
	go vet ./...

# ---- Docker ----
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-build:
	docker-compose build

# ---- Database ----
migrate:
	go run ./cmd/server  # Auto-migrates on start

# ---- Clean ----
clean:
	rm -rf $(BUILD_DIR)/
	rm -f coverage.out coverage.html
	rm -rf docs/

# ---- Help ----
help:
	@echo "CMS Backend - Make Targets"
	@echo ""
	@echo "  build         Build the binary"
	@echo "  run           Run the server"
	@echo "  dev           Run with hot reload (requires air)"
	@echo "  deps          Download and tidy dependencies"
	@echo "  test          Run all tests"
	@echo "  test-unit     Run unit tests"
	@echo "  test-integration  Run integration tests"
	@echo "  test-uat      Run UAT / E2E tests"
	@echo "  test-coverage Generate coverage report"
	@echo "  swagger       Generate Swagger docs"
	@echo "  lint          Run linter"
	@echo "  fmt           Format code"
	@echo "  docker-up     Start Docker services"
	@echo "  docker-down   Stop Docker services"
	@echo "  clean         Clean build artifacts"
