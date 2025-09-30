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
lint: ## Lint Go code (requires golangci-lint)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run tests
test: ## Run tests
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

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
