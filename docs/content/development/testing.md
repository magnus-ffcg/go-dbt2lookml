---
title: Testing
weight: 20
---

# Testing Guide

Comprehensive testing is essential for maintaining code quality. This guide covers our testing practices and how to write effective tests.

## Test Structure

```
tests/
├── integration/          # Integration tests
│   ├── fixtures/        # Test fixtures (manifest, catalog, expected output)
│   ├── cli_test.go      # CLI integration tests
│   └── integration_test.go  # Full workflow tests
└── ...

pkg/
├── parsers/
│   ├── parser.go
│   └── parser_test.go   # Unit tests alongside code
├── generators/
│   ├── generator.go
│   └── generator_test.go
└── ...
```

## 🎯 Types of Tests

### Unit Tests

Test individual functions and methods in isolation.

**Location:** `pkg/*/`

**Example:**

```go
func TestDimensionGenerator_GenerateDimension(t *testing.T) {
    tests := []struct {
        name     string
        column   models.DbtModelColumn
        expected models.LookMLDimension
    }{
        {
            name: "string dimension",
            column: models.DbtModelColumn{
                Name: "customer_name",
                Type: "STRING",
            },
            expected: models.LookMLDimension{
                Name: "customer_name",
                Type: "string",
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            gen := NewDimensionGenerator(&config.Config{})
            result := gen.GenerateDimension(tt.column)
            assert.Equal(t, tt.expected.Name, result.Name)
            assert.Equal(t, tt.expected.Type, result.Type)
        })
    }
}
```

### Integration Tests

Test complete workflows with real fixtures.

**Location:** `tests/integration/`

**Running:**

```bash
# All integration tests
go test ./tests/integration/...

# Specific test
go test ./tests/integration/ -run TestFixtureComparison

# With verbose output
go test -v ./tests/integration/...
```

### Table-Driven Tests

Our preferred pattern for comprehensive test coverage.

```go
func TestParser(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        expected    string
        expectError bool
    }{
        {"valid input", "test", "TEST", false},
        {"empty input", "", "", true},
        {"special chars", "a@b", "A@B", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Parse(tt.input)
            
            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

## 🎬 Running Tests

### Basic Commands

```bash
# All tests
make test

# With race detector
make test-race

# Short mode (skip long tests)
go test -short ./...

# Verbose output
go test -v ./...

# Specific package
go test ./pkg/generators/...

# Specific test
go test ./pkg/parsers -run TestModelParser
```

### Coverage

```bash
# Generate coverage report
make test-coverage

# View in browser
open coverage.html

# Coverage for specific package
go test -cover ./pkg/generators/
```

## Coverage Goals

- **Overall:** > 80%
- **pkg/parsers:** > 85%
- **pkg/generators:** > 85%
- **pkg/models:** > 75%

Check current coverage:

```bash
make test-coverage
```

## Writing Good Tests

### DO ✅

```go
// Clear test names
func TestParseManifest_WithValidJSON_ReturnsModels(t *testing.T) { }

// Test error cases
func TestGenerator_WithInvalidModel_ReturnsError(t *testing.T) { }

// Use table-driven tests
tests := []struct {
    name string
    // ...
}{}

// Test edge cases
{"empty input", "", "", true},
{"nil input", nil, nil, true},
{"max length", strings.Repeat("a", 1000), ...},
```

### DON'T ❌

```go
// Vague test names
func TestParse(t *testing.T) { }

// No error checking
result := Parse(input)
// Missing: if err != nil

// Hard-coded values without explanation
assert.Equal(t, 42, result) // Why 42?

// No cleanup
file := createTempFile()
// Missing: defer os.Remove(file)
```

## Testing Utilities

### Assertions

We use [testify](https://github.com/stretchr/testify):

```go
import "github.com/stretchr/testify/assert"

assert.Equal(t, expected, actual)
assert.NoError(t, err)
assert.Error(t, err)
assert.Contains(t, haystack, needle)
assert.Len(t, slice, 3)
```

### Test Fixtures

Located in `tests/integration/fixtures/`:

```go
// Load fixture
manifest := loadFixture(t, "fixtures/manifest.json")

// Helper function
func loadFixture(t *testing.T, path string) map[string]interface{} {
    data, err := os.ReadFile(path)
    require.NoError(t, err)
    
    var result map[string]interface{}
    require.NoError(t, json.Unmarshal(data, &result))
    
    return result
}
```

## 🐛 Debugging Tests

### Verbose Output

```bash
go test -v ./pkg/parsers/...
```

### Specific Test

```bash
go test -v ./pkg/parsers -run TestModelParser_FilterModels
```

### Print Debug Info

```go
func TestDebug(t *testing.T) {
    result := Parse(input)
    t.Logf("Result: %+v", result)  // Only shows on failure or -v
    fmt.Printf("Debug: %+v\n", result)  // Always shows
}
```

### Race Detector

```bash
go test -race ./...
```

## Performance Testing

### Benchmarks

```go
func BenchmarkParser(b *testing.B) {
    parser := NewParser()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.Parse(testData)
    }
}
```

**Run benchmarks:**

```bash
go test -bench=. ./pkg/parsers/
go test -bench=BenchmarkParser -benchmem
```

## CI/CD Testing

Our CI runs:

```bash
# Format check
gofmt -l .

# Linting
golangci-lint run

# Vet
go vet ./...

# Tests with race detector
go test -race -v ./...

# Coverage
go test -coverprofile=coverage.out ./...
```

Run the same locally:

```bash
make ci-check
```

## 📝 Best Practices

1. **Test behavior, not implementation**
2. **Keep tests simple and readable**
3. **One assertion per test case** (when possible)
4. **Clean up resources** (use defer)
5. **Avoid test interdependencies**
6. **Mock external dependencies**
7. **Test error paths**
8. **Document complex test scenarios**

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Table Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Test Fixtures](https://dave.cheney.net/2016/05/10/test-fixtures-in-go)
