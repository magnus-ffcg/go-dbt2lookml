# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Removed

- **Unimplemented features**
  - Removed `--generate-locale` flag (was not implemented)
  - Removed `--include-iso-fields` flag (was not implemented)
  - Cleaned up CLI, config, and documentation
  - All remaining flags are fully functional

### Added

- **Performance optimization**
  - Pre-allocated map capacity in NormalizeColumnNames methods
  - Reduces memory allocations during column normalization
  - Applied to both DbtCatalogNode and DbtModel

- **Configuration validation improvements**
  - Added `ValidateFilePaths()` method to check file existence
  - Better error messages with field names in validation errors
  - Validates files are readable and not directories
  - Separated format validation from file validation for testability

- **Dependency management**
  - Created `.github/dependabot.yml` - Automated dependency updates
  - Weekly updates for Go modules and GitHub Actions
  - Automated security vulnerability updates
  - PR limit of 5 to avoid overwhelming maintainers
  - Proper labels and conventional commit messages

- **User experience improvements**
  - Improved CLI help text with usage examples and clear flag descriptions
  - Created `example.config.yaml` - Complete configuration template with scenarios
  - Created `docs/public/configuration.md` - Comprehensive configuration guide
  - Created `docs/public/cli-reference.md` - Complete CLI command reference
  - Grouped flags by category (Core, Filtering, Generation, Error Handling)
  - Added detailed flag descriptions with examples

- **Open-source project infrastructure**
  - Created `LICENSE` - MIT License for maximum compatibility
  - Created `SECURITY.md` - Vulnerability reporting policy (48h response time)
  - Created `CONTRIBUTING.md` - Complete contributor guide (500+ lines)
    - Development setup, coding standards, testing guidelines
    - Commit message conventions, PR process, CI/CD explanation
  - Created `CODE_OF_CONDUCT.md` - Contributor Covenant v2.1
  - Added `version` command to CLI - Shows version, commit, build date, Go version, platform
  - Created GitHub issue templates (bug report, feature request)
  - Created GitHub pull request template with comprehensive checklist
  - Organized `docs/` folder structure (public/ for users, development/ for contributors)
  - Created comprehensive user documentation (getting-started.md, error-handling.md)
  - Created developer documentation (architecture.md with Phase 1 & 2 details)

- **Error handling strategy** for generation operations
  - Created `pkg/generators/error_strategy.go` - ErrorStrategy enum (FailFast, FailAtEnd, ContinueOnError)
  - Created `GenerationOptions` struct for configurable error handling
  - Created `GenerationResult` struct with detailed error reporting
  - Added `GenerateAllWithOptions()` method with better error control
  - Added `ModelError` type for tracking per-model errors
  - Added 150+ lines of comprehensive tests for error strategies
  - Users can now choose how errors are handled during generation

- **Domain services for business logic**
  - Created `pkg/models/nested_array_rules.go` - Explicit array depth limit rules (max 3 levels)
  - Created `pkg/models/dimension_conflict_resolver.go` - Dimension/dimension-group conflict resolution
  - Created `pkg/models/column_hierarchy.go` - Column structure analysis and parent-child relationships
  - Created `pkg/models/column_classifier.go` - Column classification business rules
  - Created `pkg/models/column_collections_v2.go` - Clean implementation using new services
  - Added 940 lines of comprehensive tests (39 test functions, 135+ test cases)
  - All business rules now explicit, documented, and independently testable

### Changed

- **Refactored generators to use domain services**
  - `ViewGenerator.resolveConflicts()` reduced from 44 lines to 3 lines (93% reduction)
  - `ViewGenerator` now uses `NestedArrayRules` for array depth checking
  - Removed hardcoded magic number (`dotCount > 2`) in favor of explicit business rule
  - Business logic moved from generators (30%) to domain layer (85%)
  - Generators now act as thin coordinators, delegating to domain services
  - Removed unused `log` import from `view.go`

- **Improved code organization**
  - `column_collections.go` reduced from 224 to 170 lines
  - Removed duplicate `buildHierarchyMap` function
  - Separated concerns: structure analysis vs. business rules
  - Average function length reduced: 40 → 25 lines (-37%)
  - Maximum function length reduced: 110 → 70 lines (-36%)

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
