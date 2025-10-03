# ADR 0010: Testing Strategy with Unit and Integration Tests

**Status:** Accepted

**Date:** 2024

## Context

We need a comprehensive testing strategy to ensure:
- Correct parsing of dbt artifacts
- Valid LookML generation
- Edge cases handled properly
- Regression prevention
- Confidence in refactoring

Testing approaches:
- Unit tests only
- Integration tests only
- Mixed approach
- Fixture-based testing
- Golden file testing

## Decision

We will use a mixed testing strategy with:
- **Unit tests** for individual components (parsers, generators)
- **Integration tests** with real dbt fixtures
- **Table-driven tests** for comprehensive coverage
- **Golden file comparison** for LookML output validation

## Rationale

**Pros:**
- **Comprehensive coverage** - Tests at multiple levels
- **Fast feedback** - Unit tests run quickly
- **Real-world validation** - Integration tests use actual dbt files
- **Regression prevention** - Golden files catch unexpected changes
- **Maintainable** - Table-driven tests are easy to extend
- **Confidence** - Multiple test types provide thorough validation

**Cons:**
- **More test code** - Requires maintaining multiple test types
- **Fixture management** - Need to keep fixtures up to date
- **Test complexity** - Understanding which test type to use

**Alternatives considered:**
- **Unit tests only** - Misses integration issues
- **Integration tests only** - Slow, hard to isolate failures
- **Snapshot testing only** - Brittle, hard to debug

## Consequences

**Positive:**
- High confidence in changes
- Fast unit test feedback loop
- Realistic integration validation
- Easy to add new test cases
- Clear test organization

**Negative:**
- More test code to maintain
- Need to update fixtures when dbt format changes
- Golden files need careful review

## Test Organization

```
pkg/
  parsers/
    parser.go
    parser_test.go          # Unit tests
  generators/
    generator.go
    generator_test.go       # Unit tests

tests/
  integration/
    fixtures/
      manifest.json         # Real dbt artifacts
      catalog.json
      expected/
        view1.lkml           # Golden files
    integration_test.go     # Integration tests
```

## Test Types

### Unit Tests
- Test individual functions
- Mock dependencies
- Fast execution
- Located next to source code

```go
func TestDimensionGenerator(t *testing.T) {
    tests := []struct {
        name     string
        input    Column
        expected Dimension
    }{
        // Test cases
    }
    // Table-driven test
}
```

### Integration Tests
- Test complete workflow
- Use real dbt fixtures
- Validate full output
- Located in `tests/integration/`

```go
func TestFullGeneration(t *testing.T) {
    // Load real manifest and catalog
    // Run generation
    // Compare with golden files
}
```

### Coverage Goals
- Overall: >80%
- Critical paths: >90%
- Parsers: >85%
- Generators: >85%

## Testing Commands

```bash
# All tests
make test

# Unit tests only
go test ./pkg/...

# Integration tests
go test ./tests/integration/...

# With coverage
make test-coverage

# Race detection
make test-race
```

## Fixture Management

- Keep fixtures minimal but realistic
- Update when dbt format changes
- Document fixture structure
- Version control all fixtures
- Validate fixtures against dbt schema

## Notes

This multi-layered approach provides thorough testing while keeping individual tests focused and maintainable. The combination of fast unit tests and comprehensive integration tests enables confident refactoring and feature development.
