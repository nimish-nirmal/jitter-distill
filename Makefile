.PHONY: help build test lint bench clean fmt examples install-tools

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the project
	go build -v ./...

test: ## Run tests with coverage
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | tail -1

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run --timeout=5m

bench: ## Run benchmarks
	go test -bench=. -benchmem -benchtime=3s ./...

fmt: ## Format code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

clean: ## Clean build artifacts
	go clean ./...
	rm -f coverage.out
	rm -f jitter-token
	rm -rf dist/

install-tools: ## Install development tools
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

examples: ## Build and run examples
	@echo "Building examples..."
	go build -o /tmp/jitter-token ./cmd/jitter-token
	@echo "\n=== Example 1: Generate 3 tokens ==="
	/tmp/jitter-token --count 3
	@echo "\n=== Example 2: Benchmark ==="
	/tmp/jitter-token --bench --bench-duration 2s

all: fmt vet lint test bench ## Run all checks
