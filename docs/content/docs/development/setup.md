---
title: Development Setup
weight: 5
---

# Development Setup

Guide for setting up your local development environment.

## Prerequisites

- Go 1.21 or higher
- Make (optional, but recommended)
- golangci-lint v2.5+ (for linting)
- Git

## Quick Start

```bash
# Clone the repository
git clone https://github.com/magnus-ffcg/go-dbt2lookml.git
cd go-dbt2lookml

# Install dependencies
make deps

# Run all CI checks locally (same as CI)
make ci-check

# Quick pre-commit checks
make pre-commit
```

## Running Tests

```bash
# Run all tests
make test

# Run tests with race detector (like CI)
make test-race

# Run with coverage
make test-coverage

# View coverage report
open coverage.html
```

## Building

```bash
# Build binary for your platform
make build

# Build for all platforms
make build-all

# Binary will be in bin/
./bin/dbt2lookml --help
```

## Code Quality

### Formatting

```bash
# Format code
make fmt

# Check formatting
gofmt -l .
```

### Linting

```bash
# Lint code (requires golangci-lint v2.5+)
make lint

# Run go vet
go vet ./...
```

**Install golangci-lint:**

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
```

### Pre-commit Hooks

Set up Git hooks to run checks before each commit:

```bash
# Option 1: Use the .githooks directory
git config core.hooksPath .githooks

# Option 2: Install pre-commit (requires Python)
pip install pre-commit
pre-commit install
```

## Available Make Targets

Run `make help` to see all available targets:

```bash
make help
```

Common targets:
- `make build` - Build binary
- `make test` - Run tests
- `make test-race` - Run tests with race detector
- `make test-coverage` - Generate coverage report
- `make lint` - Lint code
- `make fmt` - Format code
- `make ci-check` - Run all CI checks
- `make pre-commit` - Run pre-commit checks
- `make clean` - Clean build artifacts
- `make deps` - Install dependencies

## Project Structure

```
go-dbt2lookml/
├── cmd/                    # CLI application entry points
│   └── dbt2lookml/        # Main CLI
├── pkg/                    # Public packages
│   ├── models/            # Core data models
│   ├── parsers/           # Parsing logic
│   ├── generators/        # LookML generation
│   ├── enums/             # Enumerations
│   └── utils/             # Utilities
├── internal/              # Private application code
│   ├── cli/               # CLI implementation
│   └── config/            # Configuration
├── tests/                 # Test files and fixtures
│   └── integration/       # Integration tests
├── docs/                  # Documentation
│   └── content/           # Hugo content
└── scripts/               # Build and utility scripts
```

## Development Workflow

1. **Create a branch** - `git checkout -b feature/your-feature`
2. **Make changes** - Write code and tests
3. **Run tests** - `make test`
4. **Run quality checks** - `make ci-check`
5. **Commit** - Use conventional commits
6. **Push** - `git push origin feature/your-feature`
7. **Create PR** - Open a pull request

See [Contributing](contributing) for detailed guidelines.

## Debugging

### Running with Debug Logging

```bash
# Build and run with debug logging
go run ./cmd/dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir output \
  --log-level DEBUG \
  --log-format console
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd/dbt2lookml -- \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir output
```

## Troubleshooting

### Tests Failing

```bash
# Clean and rebuild
make clean
make deps
make test
```

### Linter Errors

```bash
# Update golangci-lint
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

# Run linter with verbose output
golangci-lint run --verbose
```

### Build Issues

```bash
# Clear Go cache
go clean -cache -modcache -testcache

# Reinstall dependencies
make deps
```

## Next Steps

- Read the [Contributing Guide](contributing)
- Check out the [Testing Guide](testing)
- Browse the [API Reference](../api)
