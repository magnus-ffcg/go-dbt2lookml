# Contributing to dbt2lookml

First off, thank you for considering contributing to dbt2lookml! It's people like you that make this tool better for everyone.

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior via GitHub issues.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible using our [bug report template](.github/ISSUE_TEMPLATE/bug_report.md).

**Guidelines:**
- Use a clear and descriptive title
- Describe the exact steps to reproduce the problem
- Provide specific examples (commands, config files, dbt models)
- Describe the behavior you observed and what you expected
- Include your environment details (OS, dbt version, Go version)
- Include relevant logs and error messages

### Suggesting Features

Feature suggestions are tracked as GitHub issues. When creating a feature request, use our [feature request template](.github/ISSUE_TEMPLATE/feature_request.md).

**Guidelines:**
- Use a clear and descriptive title
- Provide a detailed description of the proposed feature
- Explain why this feature would be useful
- Include examples of how it would be used
- Consider any potential drawbacks or alternatives

### Pull Requests

We actively welcome your pull requests! Here's the process:

1. **Fork the repo** and create your branch from `main`
2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Ensure all tests pass** (`make test`)
5. **Run linters** (`make lint`)
6. **Update documentation** as needed
7. **Submit a pull request** using our [PR template](.github/pull_request_template.md)

## Development Setup

### Prerequisites

- **Go 1.21 or later** ([download](https://go.dev/dl/))
- **golangci-lint v2.5+** for linting
- **make** for running build tasks
- **Git** for version control

### Getting Started

1. **Clone the repository:**
   ```bash
   git clone https://github.com/magnus-ffcg/go-dbt2lookml.git
   cd go-dbt2lookml
   ```

2. **Install dependencies:**
   ```bash
   make deps
   ```

3. **Install development tools:**
   ```bash
   # golangci-lint
   go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
   ```

4. **Verify setup:**
   ```bash
   make test
   make lint
   ```

### Development Workflow

#### Running Tests

```bash
# Run all tests
make test

# Run tests with race detector (like CI)
make test-race

# Run tests with coverage
make test-coverage

# Run only unit tests
go test ./pkg/...

# Run only integration tests
go test ./tests/integration/...

# Run specific test
go test -run TestName ./pkg/generators/
```

#### Building

```bash
# Build the binary
make build

# The binary will be at ./dbt2lookml
./dbt2lookml --help

# Build for all platforms (like release)
make build-all
```

#### Code Quality

```bash
# Format code
make fmt

# Run linters (uses golangci-lint v2.5+)
make lint

# Run all CI checks locally
make ci-check

# Quick pre-commit checks
make pre-commit
```

#### Using Pre-commit Hooks

Set up Git hooks to run checks before each commit:

```bash
# Option 1: Use .githooks directory
git config core.hooksPath .githooks

# Option 2: Use pre-commit (requires Python)
pip install pre-commit
pre-commit install
```

### Project Structure

```
dbt2lookml/
├── cmd/dbt2lookml/       # CLI entry point
├── internal/             # Private application code
│   ├── cli/              # Command-line interface
│   └── config/           # Configuration
├── pkg/                  # Public packages
│   ├── enums/            # Enumerations
│   ├── models/           # Domain models & business logic
│   ├── generators/       # LookML generation
│   ├── parsers/          # dbt file parsing
│   └── utils/            # Utility functions
├── tests/                # Tests and fixtures
│   ├── integration/      # Integration tests
│   ├── unit/             # Unit tests (1:1 with pkg/)
│   └── fixtures/         # Test data
└── docs/                 # Documentation
    ├── public/           # User documentation
    └── development/      # Developer documentation
```

**Key Principles:**
- **Domain logic** goes in `pkg/models/` (not in generators)
- **Tests** mirror the structure of what they test
- **One test file per source file** in `tests/unit/`
- **All exported functions must be tested**

See [Architecture Documentation](docs/development/architecture.md) for details.

## Coding Standards

### Go Style

- Follow standard Go style ([Effective Go](https://go.dev/doc/effective_go))
- Use `gofmt` for formatting (automatic via `make fmt`)
- Use `goimports` for import management
- Run `golangci-lint` before committing

### Naming Conventions

- **Packages:** Short, lowercase, single-word names (`models`, `parsers`)
- **Types:** PascalCase (`DbtModel`, `LookMLView`)
- **Functions/Methods:** PascalCase for exported, camelCase for private
- **Variables:** camelCase (`modelName`, `columnType`)
- **Constants:** PascalCase or SCREAMING_SNAKE_CASE for enums

### Code Organization

```go
// 1. Package declaration
package models

// 2. Imports (grouped: stdlib, external, internal)
import (
    "fmt"
    "strings"
    
    "github.com/external/package"
    
    "github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
)

// 3. Constants
const MaxDepth = 3

// 4. Types
type MyService struct {
    config *Config
}

// 5. Constructor
func NewMyService(config *Config) *MyService {
    return &MyService{config: config}
}

// 6. Methods
func (s *MyService) DoSomething() error {
    // Implementation
}
```

### Documentation

- **All exported types, functions, and methods must have doc comments**
- Doc comments start with the name of the item
- Use complete sentences
- Include examples for complex functionality

```go
// NewColumnClassifier creates a new column classifier for categorizing columns
// based on their structure and relationships.
//
// The classifier uses the provided hierarchy to understand parent-child
// relationships and determines whether columns should be included in the
// main view, nested views, or excluded entirely.
func NewColumnClassifier(hierarchy *ColumnHierarchy, arrayColumns map[string]bool) *ColumnClassifier {
    // Implementation
}
```

### Error Handling

- Return errors, don't panic (except for truly unrecoverable situations)
- Wrap errors with context using `fmt.Errorf("context: %w", err)`
- Use custom error types for domain-specific errors
- Validate inputs early

```go
// ✅ GOOD
func ProcessModel(model *DbtModel) error {
    if model == nil {
        return fmt.Errorf("model cannot be nil")
    }
    
    if err := validate(model); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    return nil
}

// ❌ BAD
func ProcessModel(model *DbtModel) {
    if model == nil {
        panic("model is nil")  // Don't panic!
    }
}
```

## Testing Guidelines

### Test Requirements

- **All exported functions must have tests**
- **Test coverage:** Aim for 80%+ on new code
- **Test file location:** Mirror the file being tested
  - Source: `pkg/models/classifier.go`
  - Test: `tests/unit/models/classifier_test.go`

### Writing Good Tests

```go
func TestFunctionName_Scenario(t *testing.T) {
    // Arrange
    input := setupTestData()
    expected := expectedResult()
    
    // Act
    actual := FunctionName(input)
    
    // Assert
    assert.Equal(t, expected, actual)
}
```

**Guidelines:**
- **Use table-driven tests** for multiple scenarios
- **Test edge cases** (empty, nil, max values)
- **Test error conditions**
- **Use descriptive test names:** `TestName_Scenario_ExpectedBehavior`
- **Keep tests independent** (no shared state)

### Test Structure Example

```go
func TestColumnClassifier_Classify(t *testing.T) {
    tests := []struct {
        name     string
        column   string
        expected Category
    }{
        {
            name:     "simple column should be in main view",
            column:   "customer_id",
            expected: CategoryMainView,
        },
        {
            name:     "nested array should be excluded",
            column:   "items.details.meta",
            expected: CategoryExcluded,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            classifier := setupClassifier()
            result := classifier.Classify(tt.column, nil)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Tests

- **Do NOT modify fixtures** in `tests/fixtures/`
- Output must match fixtures exactly (except minor whitespace)
- Add new fixtures for new test scenarios
- Document what each fixture tests

## Documentation Standards

### User Documentation (`docs/public/`)

- **Audience:** End users
- **Focus:** How to use the tool
- **Style:** Clear, example-rich, actionable
- **Update when:** Adding features, changing behavior

### Developer Documentation (`docs/development/`)

- **Audience:** Contributors
- **Focus:** How the code works
- **Style:** Technical, detailed, with diagrams
- **Update when:** Changing architecture, adding components

### README Updates

Update README.md when:
- Adding new features
- Changing installation process
- Updating requirements
- Adding new dependencies

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <short description>

<optional longer description>

<optional footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring (no behavior change)
- `perf`: Performance improvement
- `test`: Adding or updating tests
- `docs`: Documentation only
- `style`: Code style/formatting
- `chore`: Build, dependencies, tooling

**Examples:**

```
feat: add support for GEOGRAPHY column type

Adds detection and LookML generation for BigQuery GEOGRAPHY columns.
Includes tests and documentation.

Closes #123
```

```
fix: handle empty array columns correctly

Previously crashed on models with empty ARRAY fields.
Now properly handles and logs a warning.

Fixes #456
```

```
refactor: extract nested array rules to domain service

Moves business logic from generator to dedicated service.
Adds comprehensive tests. No behavioral changes.
```

## Pull Request Process

1. **Create a branch** from `main`:
   ```bash
   git checkout -b feature/my-feature
   # or
   git checkout -b fix/issue-123
   ```

2. **Make your changes:**
   - Write code following our standards
   - Add/update tests
   - Update documentation
   - Run `make ci-check` locally

3. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add my feature"
   ```

4. **Push to your fork:**
   ```bash
   git push origin feature/my-feature
   ```

5. **Create a Pull Request:**
   - Use our PR template
   - Fill out all sections
   - Link related issues
   - Request review

6. **Address review feedback:**
   - Make requested changes
   - Push updates to the same branch
   - Respond to comments

7. **Merge:**
   - Maintainer will merge when approved
   - Branch will be deleted automatically

## CI/CD Pipeline

All PRs must pass CI checks:

- ✅ **Tests:** All unit and integration tests
- ✅ **Race Detector:** `go test -race`
- ✅ **Linting:** golangci-lint v2.5+
- ✅ **Security:** govulncheck
- ✅ **Build:** Must compile successfully
- ✅ **Coverage:** Uploaded to Codecov

**Local CI Simulation:**
```bash
make ci-check
```

## Release Process

Maintainers handle releases:

1. Update `CHANGELOG.md`
2. Create version tag: `git tag v1.0.0`
3. Push tag: `git push origin v1.0.0`
4. GitHub Actions automatically:
   - Builds binaries for all platforms
   - Creates GitHub release
   - Generates release notes
   - Uploads artifacts

## Getting Help

- **Questions:** [GitHub Discussions](https://github.com/magnus-ffcg/go-dbt2lookml/discussions)
- **Bugs:** [GitHub Issues](https://github.com/magnus-ffcg/go-dbt2lookml/issues)
- **Docs:** [Documentation](docs/)
- **Architecture:** [Architecture Guide](docs/development/architecture.md)

## Recognition

Contributors are recognized in:
- GitHub contributors page
- Release notes
- CHANGELOG.md (for significant contributions)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to dbt2lookml! 🎉
