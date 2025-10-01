# dbt2lookml (Go)

Generate LookML views from BigQuery via dbt models.

## Features

- Parse dbt manifest and catalog files
- Generate LookML views, dimensions, measures, and explores
- Support for complex nested BigQuery structures (ARRAY, STRUCT)
- CLI interface with rich configuration options
- Comprehensive validation and error handling

## Installation

```bash
go install github.com/magnus-ffcg/dbt2lookml/cmd/dbt2lookml@latest
```

## Usage

```bash
dbt2lookml --config config.yaml
```

## Development

### Quick Start

```bash
# Install dependencies
make deps

# Run all CI checks locally (same as CI)
make ci-check

# Quick pre-commit checks
make pre-commit
```

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detector (like CI)
make test-race

# Run with coverage
make test-coverage
```

### Building

```bash
# Build binary
make build

# Build for all platforms
make build-all
```

### Code Quality

```bash
# Format code
make fmt

# Lint code (requires golangci-lint v2.5+)
make lint

# Run go vet
go vet ./...
```

**Install golangci-lint v2.5+:**

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

### Available Make Targets

Run `make help` to see all available targets:

```bash
make help
```
```

## Architecture

- `cmd/` - CLI application entry points
- `pkg/` - Public packages (models, parsers, generators)
- `internal/` - Private application code
- `tests/` - Test files and fixtures
- `docs/` - Project documentation, split on development and public docs.
