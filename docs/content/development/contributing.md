---
title: Contributing
weight: 10
---

# Contributing to go-dbt2lookml

Thank you for your interest in contributing! This guide will help you get started.

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Make (optional, but recommended)
- golangci-lint v2.5+ (for linting)

### Setup

```bash
# Clone the repository
git clone https://github.com/magnus-ffcg/go-dbt2lookml.git
cd go-dbt2lookml

# Install dependencies
make deps

# Run tests to verify setup
make test
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Write your code following our coding standards (see below).

### 3. Run Tests

```bash
# Run all tests
make test

# Run with race detector
make test-race

# Run with coverage
make test-coverage
```

### 4. Run Quality Checks

```bash
# Run all CI checks locally
make ci-check

# Or run individually:
make fmt      # Format code
make lint     # Lint code
make vet      # Run go vet
```

### 5. Commit Changes

We use [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug in parser"
git commit -m "docs: update README"
```

**Commit Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting)
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

### 6. Push and Create PR

```bash
git push origin feature/your-feature-name
```

Then create a Pull Request on GitHub.

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Write clear, self-documenting code
- Add comments for exported functions and types

### Testing

- Write unit tests for new features
- Maintain or improve code coverage
- Use table-driven tests where appropriate
- Mock external dependencies

### Documentation

- Update relevant documentation
- Add examples for new features
- Keep README.md up to date

## Testing

### Running Tests

```bash
# All tests
make test

# Unit tests only
go test ./pkg/...

# Integration tests
go test ./tests/integration/...

# With coverage
make test-coverage
open coverage.html
```

### Writing Tests

Example test structure:

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "basic case",
            input:    "input",
            expected: "output",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := YourFunction(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

## Code Review Process

1. All PRs require at least one approval
2. CI checks must pass
3. Code coverage should not decrease
4. Documentation should be updated

## Getting Help

- **Questions:** Open a [Discussion](https://github.com/magnus-ffcg/go-dbt2lookml/discussions)
- **Bugs:** Open an [Issue](https://github.com/magnus-ffcg/go-dbt2lookml/issues)
- **Security:** See [SECURITY.md](https://github.com/magnus-ffcg/go-dbt2lookml/blob/main/SECURITY.md)

## Makefile Targets

Run `make help` to see all available targets:

```bash
make help
```

Common targets:
- `make build` - Build binary
- `make test` - Run tests
- `make lint` - Lint code
- `make ci-check` - Run all CI checks
- `make clean` - Clean build artifacts

## Thank You!

Your contributions make this project better for everyone!
