# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Interface-based architecture** for all generator types
  - Created `pkg/generators/interfaces.go` with 5 generator interfaces
  - `ViewGeneratorInterface`, `DimensionGeneratorInterface`, `MeasureGeneratorInterface`
  - `ExploreGeneratorInterface`, `LookMLGeneratorInterface`
  - Compile-time interface checks for all generators
  - Mock implementation examples in `pkg/generators/interfaces_test.go`
  - Enables dependency injection and easier testing

- **Context support** for cancellable operations
  - New `GenerateAllWithContext(ctx context.Context, models []*models.DbtModel)` method
  - Graceful cancellation support for long-running generation tasks
  - Timeout support via context deadlines
  - Context checked before processing each model
  - Comprehensive tests in `pkg/generators/context_test.go`
  - Backward compatible - `GenerateAll()` uses `context.Background()`

- **Validation methods** for LookML domain models
  - `LookMLDimension.Validate()` - validates required fields and dimension types
  - `LookMLView.Validate()` - validates view structure and fields
  - `LookMLMeasure.Validate()` - validates measure configuration and type-specific attributes
  - 19 comprehensive validation tests
  - Clear, actionable error messages

- **CLI flag: `--flatten`**
  - Generates all LookML files in a single directory without nested structure
  - Useful for projects that prefer flat file organization
  - Works seamlessly with `--use-table-name` flag
  - Tested with 4,273+ files

### Changed

- **Performance: Pre-compiled regex patterns**
  - Moved regex compilation to package-level variables in `pkg/utils/string_utils.go`
  - Added `sanitizeInvalidChars`, `consecutiveUnderscore`, `threeOrMoreUnderscores` patterns
  - Updated `SanitizeIdentifier()` to use pre-compiled regexes
  - Updated `ToLookMLName()` to use pre-compiled regexes
  - Eliminates runtime regex compilation overhead (significant performance improvement)

- **Improved regex variable naming**
  - `camelToSnakeReg1` → `acronymPatterns` (more descriptive)
  - `camelToSnakeReg2` → `insertUnderscoreBeforeUppercase` (clearer intent)
  - `camelToSnakeReg3` → `camelCasePatterns` (better naming)
  - `camelToSnakeReg4` → `cleanupUnderscores` (self-documenting)
  - `sanitizeInvalidCharsReg` → `sanitizeInvalidChars` (consistent naming)

- **CLI state management refactoring**
  - Refactored 15 package-level variables into single `cliFlags` struct
  - Eliminated hidden global state
  - Improved testability and maintainability
  - More explicit dependency management

- **Code formatting and whitespace cleanup**
  - Consistent indentation throughout codebase
  - Removed trailing whitespace
  - Improved code readability
  - Professional code style

- **Removed hardcoded business logic from utilities**
  - Removed "supplierinformation" special case from `ToLookMLName()`
  - Function now uses purely algorithmic conversion
  - Added clear documentation about input expectations
  - Utilities are now domain-agnostic and more maintainable

### Fixed

- **Empty filename bug** for ephemeral models
  - Ephemeral models with empty `RelationName` now fallback to model name
  - No more `.view.lkml` files with empty prefixes
  - All 70+ ephemeral models now have proper filenames

### Documentation

- **Added package-level documentation**
  - Comprehensive godoc comments for `pkg/generators`
  - Comprehensive godoc comments for `pkg/parsers`
  - Comprehensive godoc comments for `pkg/models`
  - Comprehensive godoc comments for `pkg/utils`
  - All packages now include usage examples
  - Documentation will appear on pkg.go.dev

- **Added CHANGELOG.md**
  - Follows [Keep a Changelog](https://keepachangelog.com/) format
  - Tracks all changes chronologically
  - Semantic versioning ready

### CI/CD & Automation

- **GitHub Actions workflows**
  - Streamlined CI pipeline (`.github/workflows/ci.yml`)
  - Tests on Ubuntu with Go 1.23 (single reliable runner)
  - Single lint job with golangci-lint v2.5.0 (includes fmt, vet, staticcheck, etc.)
  - Security scanning with govulncheck
  - Code coverage reporting with Codecov
  - 4 focused jobs: test, lint, build, security

- **Automated release process**
  - GoReleaser configuration (`.goreleaser.yml`)
  - Multi-platform binary builds (Linux, macOS, Windows)
  - Multi-architecture support (amd64, arm64)
  - Automated CHANGELOG generation
  - Package archives and checksums
  - GitHub releases with download instructions

- **Quality tooling**
  - golangci-lint v2.5.0 configuration (`.golangci.yml`)
  - Same linter version in CI and local development
  - Comprehensive linter rules (9 linters enabled)
  - Format and style checking
  - Consistent configuration across all environments

- **Local development tools**
  - Makefile with `ci-check` target (runs same checks as CI)
  - Pre-commit hooks (`.githooks/pre-commit`)
  - Pre-commit configuration (`.pre-commit-config.yaml`)
  - Quick feedback before pushing to CI
  - Single `.golangci.yml` works everywhere

### Security

- **Updated .gitignore**
  - Added `.cursor`, `.windsurf` to prevent IDE files from being committed
  - Added `.work/*` with exceptions for `.work/docs/` (allows documentation access)
  - Added `output/` directory to ignore generated files
  - Added compiled binary `dbt2lookml` to ignore list

- **Security scanning**
  - gosec security scanner in CI
  - govulncheck for vulnerability detection
  - Automated security checks on every PR

---

## [0.1.0] - Previous Release

### Added
- Initial release
- Basic LookML generation from dbt catalog and manifest
- Support for dimensions, measures, and explores
- BigQuery adapter support
- CLI interface with multiple flags
- Comprehensive test suite

### Features
- Generate LookML views from dbt models
- Generate LookML explores with joins
- Handle nested/repeated columns (BigQuery STRUCT/ARRAY)
- Dimension groups for date/timestamp fields
- Automatic measure generation
- Custom LookML metadata via dbt meta tags
- Schema-based directory organization

---

## Future Enhancements

### Planned
- Package-level documentation comments
- Extract magic numbers to named constants
- Method receiver consistency review
- Additional performance optimizations
- Enhanced error messages with suggestions
- Support for more dbt adapters

### Under Consideration
- Configuration file support
- Custom template system
- Plugin architecture for extensibility
- Interactive mode
- Watch mode for continuous generation

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on how to contribute to this project.

## License

See [LICENSE](LICENSE) for license information.
