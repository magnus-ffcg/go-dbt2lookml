# Go dbt2lookml Makefile

.PHONY: build test clean install fmt lint deps help

# Build variables
BINARY_NAME=dbt2lookml
BUILD_DIR=bin
MAIN_PATH=./cmd/dbt2lookml

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Default target
all: deps fmt test build

# Help target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Install dependencies
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt: ## Format Go code
	$(GOFMT) -s -w .
	$(GOCMD) fmt ./...

# Lint code
lint: ## Lint Go code (requires golangci-lint v2.5+)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0"; \
	fi

# Run tests
test: ## Run tests
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Run tests with race detector (like CI)
test-race: ## Run tests with race detector
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...

# Build binary
build: ## Build binary
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all: ## Build for multiple platforms
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Install binary to GOPATH/bin
install: ## Install binary
	$(GOCMD) install $(MAIN_PATH)

# Clean build artifacts
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# Run the application
run: ## Run the application
	$(GOCMD) run $(MAIN_PATH)

# Run with example config
run-example: ## Run with example configuration
	$(GOCMD) run $(MAIN_PATH) --config example.config.yaml

# Development targets
dev-deps: ## Install development dependencies
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Check for security vulnerabilities
security: ## Check for security vulnerabilities
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs: ## Generate documentation
	@if command -v godoc >/dev/null 2>&1; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Benchmark tests
benchmark: ## Run benchmark tests
	$(GOTEST) -bench=. -benchmem ./...

# Profile CPU usage
profile-cpu: ## Profile CPU usage
	$(GOTEST) -cpuprofile=cpu.prof -bench=. ./...
	$(GOCMD) tool pprof cpu.prof

# Profile memory usage
profile-mem: ## Profile memory usage
	$(GOTEST) -memprofile=mem.prof -bench=. ./...
	$(GOCMD) tool pprof mem.prof

# Check Go modules
mod-check: ## Check Go modules
	$(GOMOD) verify
	$(GOMOD) tidy
	git diff --exit-code go.mod go.sum

# Update dependencies
update-deps: ## Update dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

# CI checks - run all checks that CI runs
ci-check: ## Run all CI checks locally
	@echo "Running CI checks..."
	@echo "\n==> Checking formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "The following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	else \
		echo "✓ All files are properly formatted"; \
	fi
	@echo "\n==> Running go vet..."
	@$(GOCMD) vet ./... && echo "✓ go vet passed"
	@echo "\n==> Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m --config .golangci.yml && echo "✓ golangci-lint passed"; \
	else \
		echo "⚠ golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0"; \
		exit 1; \
	fi
	@echo "\n==> Running tests with race detector..."
	@$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./... && echo "✓ All tests passed"
	@echo "\n==> Building binary..."
	@$(GOBUILD) -o /tmp/$(BINARY_NAME) $(MAIN_PATH) && echo "✓ Build successful"
	@echo "\n✅ All CI checks passed!"

# Quick pre-commit check (faster than full CI)
pre-commit: fmt lint test ## Run quick pre-commit checks
	@echo "✅ Pre-commit checks passed!"

# Versioning targets
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION_PARTS := $(subst ., ,$(subst v,,$(CURRENT_VERSION)))
MAJOR := $(word 1,$(VERSION_PARTS))
MINOR := $(word 2,$(VERSION_PARTS))
PATCH := $(word 3,$(VERSION_PARTS))

version-current: ## Show current version
	@echo "Current version: $(CURRENT_VERSION)"

version-patch: ## Bump patch version (e.g., v1.2.3 -> v1.2.4)
	@$(eval NEW_VERSION := v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))))
	@echo "Bumping version from $(CURRENT_VERSION) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "✅ Tagged and pushed $(NEW_VERSION)"

version-minor: ## Bump minor version (e.g., v1.2.3 -> v1.3.0)
	@$(eval NEW_VERSION := v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0)
	@echo "Bumping version from $(CURRENT_VERSION) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "✅ Tagged and pushed $(NEW_VERSION)"

version-major: ## Bump major version (e.g., v1.2.3 -> v2.0.0)
	@$(eval NEW_VERSION := v$(shell echo $$(($(MAJOR)+1))).0.0)
	@echo "Bumping version from $(CURRENT_VERSION) to $(NEW_VERSION)"
	@git tag -a $(NEW_VERSION) -m "Release $(NEW_VERSION)"
	@git push origin $(NEW_VERSION)
	@echo "✅ Tagged and pushed $(NEW_VERSION)"

# Documentation targets
docs-serve: ## Serve documentation locally (requires hugo)
	@if ! command -v hugo > /dev/null; then \
		echo "❌ Hugo not found. Install from: https://gohugo.io/installation/"; \
		exit 1; \
	fi
	cd docs && hugo server -D

docs-build: ## Build documentation (requires hugo)
	@if ! command -v hugo > /dev/null; then \
		echo "❌ Hugo not found. Install from: https://gohugo.io/installation/"; \
		exit 1; \
	fi
	cd docs && hugo --minify

docs-clean: ## Clean Hugo build artifacts
	rm -rf docs/public docs/resources

docs-api: ## Generate API documentation from Go packages (requires gomarkdoc)
	@if ! command -v gomarkdoc > /dev/null; then \
		echo "❌ gomarkdoc not found. Install with: go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@latest"; \
		exit 1; \
	fi
	./scripts/generate-api-docs.sh

docs-full: docs-api docs-build ## Generate API docs and build Hugo site
