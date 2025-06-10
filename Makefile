# Deep Clone Library for Go
# Set up GOBIN so that our binaries are installed to ./bin instead of $GOPATH/bin.
PROJECT_ROOT = $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
export GOBIN = $(PROJECT_ROOT)/bin

GOLANGCI_LINT_VERSION := $(shell $(GOBIN)/golangci-lint version --format short 2>/dev/null)
REQUIRED_GOLANGCI_LINT_VERSION := $(shell cat .golangci.version)

.PHONY: all
all: lint test

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
test: ## Run all tests with race detection
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

.PHONY: lint
lint: golangci-lint tidy-lint ## Run all linters

# Install golangci-lint with the required version in GOBIN if it is not already installed.
.PHONY: install-golangci-lint
install-golangci-lint:
    ifneq ($(GOLANGCI_LINT_VERSION),$(REQUIRED_GOLANGCI_LINT_VERSION))
		@echo "[lint] installing golangci-lint v$(REQUIRED_GOLANGCI_LINT_VERSION)"
		@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v$(REQUIRED_GOLANGCI_LINT_VERSION)
    endif

.PHONY: golangci-lint
golangci-lint: install-golangci-lint ## Run golangci-lint
	@echo "[lint] $(shell $(GOBIN)/golangci-lint version)"
	@$(GOBIN)/golangci-lint run --timeout=10m

.PHONY: tidy-lint
tidy-lint: ## Check if go.mod and go.sum are tidy
	@echo "[lint] mod tidy"
	@go mod tidy

.PHONY: fmt
fmt: ## Format Go code
	@echo "[fmt] Formatting Go code..."
	@go fmt ./...

.PHONY: vet
vet: ## Run go vet
	@echo "[vet] Running go vet..."
	@go vet ./...

.PHONY: verify
verify: deps fmt vet lint test ## Run all verification steps
	@echo "[verify] All verification steps completed successfully ✅"
