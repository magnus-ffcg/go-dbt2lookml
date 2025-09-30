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

### Running Tests

```bash
# Run all tests
go test ./...

# Run unit tests only
go test ./tests/unit/...

# Run with coverage
go test ./tests... -cover
```

### Building

```bash
# Build binary
go build -o bin/dbt2lookml ./cmd/dbt2lookml

# Format code
go fmt ./...

# Lint
golangci-lint run
```

## Architecture

- `cmd/` - CLI application entry points
- `pkg/` - Public packages (models, parsers, generators)
- `internal/` - Private application code
- `tests/` - Test files and fixtures
- `docs/` - Project documentation, split on development and public docs.
