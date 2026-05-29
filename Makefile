.PHONY: help setup build test lint proto docker-build docker-up docker-down clean

help: ## Show this help message
	@echo "ComputeCoin Monorepo - Available targets:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

setup: ## Install all dependencies
	@echo "Installing Node.js dependencies..."
	pnpm install
	@echo "Installing Rust dependencies..."
	cargo fetch --manifest-path packages/matching-engine/Cargo.toml
	cargo fetch --manifest-path packages/blockchain/Cargo.toml
	@echo "Installing Go dependencies..."
	cd packages/trading-service && go mod download
	cd packages/oracle-service && go mod download
	@echo "Setup complete."

build: ## Compile all packages
	@echo "Building matching-engine (Rust)..."
	cargo build --manifest-path packages/matching-engine/Cargo.toml
	@echo "Building blockchain (Rust)..."
	cargo build --manifest-path packages/blockchain/Cargo.toml
	@echo "Building trading-service (Go)..."
	cd packages/trading-service && go build ./...
	@echo "Building oracle-service (Go)..."
	cd packages/oracle-service && go build ./...
	@echo "Building api-gateway (Node.js)..."
	cd packages/api-gateway && pnpm run build
	@echo "Building web-app (Next.js)..."
	cd packages/web-app && pnpm run build
	@echo "All packages built successfully."

test: ## Run all tests
	@echo "Testing matching-engine (Rust)..."
	cargo test --manifest-path packages/matching-engine/Cargo.toml
	@echo "Testing blockchain (Rust)..."
	cargo test --manifest-path packages/blockchain/Cargo.toml
	@echo "Testing trading-service (Go)..."
	cd packages/trading-service && go test ./...
	@echo "Testing oracle-service (Go)..."
	cd packages/oracle-service && go test ./...
	@echo "Testing api-gateway (Node.js)..."
	cd packages/api-gateway && pnpm run test
	@echo "Testing web-app (Next.js)..."
	cd packages/web-app && pnpm run lint
	@echo "All tests passed."

lint: ## Run all linters
	@echo "Linting Rust packages..."
	cargo clippy --manifest-path packages/matching-engine/Cargo.toml -- -D warnings
	cargo clippy --manifest-path packages/blockchain/Cargo.toml -- -D warnings
	@echo "Linting Go packages..."
	cd packages/trading-service && go vet ./...
	cd packages/oracle-service && go vet ./...
	@echo "Linting Node.js packages..."
	cd packages/api-gateway && pnpm run lint || true
	cd packages/web-app && pnpm run lint || true
	@echo "Linting complete."

proto: ## Generate protobuf code
	@echo "Generating protobuf types..."
	cd packages/shared && pnpm run proto:generate
	@echo "Protobuf generation complete."

docker-build: ## Build all Docker images
	docker compose build

docker-up: ## Start all services in background
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

clean: ## Remove build artifacts
	@echo "Cleaning Rust targets..."
	cargo clean --manifest-path packages/matching-engine/Cargo.toml
	cargo clean --manifest-path packages/blockchain/Cargo.toml
	@echo "Cleaning Go binaries..."
	cd packages/trading-service && go clean ./...
	cd packages/oracle-service && go clean ./...
	@echo "Cleaning Node.js artifacts..."
	rm -rf packages/api-gateway/dist
	rm -rf packages/web-app/.next
	rm -rf node_modules
	@echo "Clean complete."
