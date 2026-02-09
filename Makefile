# Deep Clone Library for Go
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

# Tool versions
GOLANGCI_LINT_BINARY := $(GOBIN)/golangci-lint
GOLANGCI_LINT_VERSION := $(shell $(GOLANGCI_LINT_BINARY) version --format short 2>/dev/null || $(GOLANGCI_LINT_BINARY) version --short 2>/dev/null || echo "not-installed")
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version 2>/dev/null || echo "2.4.0")

.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo "Deep Clone Library for Go"
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts and caches
	@echo "[clean] Cleaning build artifacts..."
	@rm -rf $(GOBIN)
	@go clean -cache -testcache
	@rm -f coverage.out coverage.html

.PHONY: deps
deps: ## Download Go module dependencies
	@echo "[deps] Downloading dependencies..."
	@go mod download
	@go mod tidy

.PHONY: test
test: ## Run all tests
	@echo "[test] Running all tests..."
	@go test -race ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "[test] Running tests with coverage..."
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "[test] Coverage report generated: coverage.html"

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "[test] Running tests with verbose output..."
	@go test -race -v ./...

.PHONY: bench
bench: ## Run benchmarks
	@echo "[bench] Running benchmarks..."
	@go test -bench=. -benchmem ./...

.PHONY: bench-comparison
bench-comparison: ## Run comparison benchmarks
	@echo "[bench] Running comparison benchmarks..."
	@cd benchmarks && go test -bench=. -benchmem -benchtime=1s

.PHONY: install-golangci-lint
install-golangci-lint:
	@mkdir -p $(GOBIN)
	@if [ "$(GOLANGCI_LINT_VERSION)" != "$(REQUIRED_GOLANGCI_LINT_VERSION)" ]; then \
		echo "[lint] Installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) (current: $(GOLANGCI_LINT_VERSION))"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION); \
		echo "[lint] golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) installed successfully"; \
	else \
		echo "[lint] golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION) already installed"; \
	fi

.PHONY: golangci-lint
golangci-lint: install-golangci-lint ## Run golangci-lint
	@echo "[lint] Running $(shell $(GOLANGCI_LINT_BINARY) version)"
	@$(GOLANGCI_LINT_BINARY) run --timeout=10m

.PHONY: tidy-lint
tidy-lint: ## Check if go.mod and go.sum are tidy
	@echo "[lint] Checking go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || (echo "go.mod or go.sum is not tidy" && exit 1)

.PHONY: fmt
fmt: ## Format Go code
	@echo "[fmt] Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "[vet] Running go vet..."
	@go vet ./...

.PHONY: lint
lint: golangci-lint tidy-lint ## Run all linters

.PHONY: verify
verify: deps fmt vet lint test ## Run all verification steps
	@echo "[verify] All verification steps completed successfully ✅"

.PHONY: all
all: verify ## Run all verification steps (alias for verify)