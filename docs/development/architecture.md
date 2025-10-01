# Architecture Overview

Understanding the dbt2lookml codebase architecture.

## High-Level Architecture

```
┌─────────────┐
│     CLI     │  Command-line interface
└──────┬──────┘
       │
┌──────▼──────────────────────────────────────┐
│              Coordinator                    │
│  (internal/cli - application logic)         │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│            Generators                       │
│  (pkg/generators - LookML generation)       │
│  • LookMLGenerator (coordinator)            │
│  • ViewGenerator                            │
│  • DimensionGenerator                       │
│  • MeasureGenerator                         │
│  • ExploreGenerator                         │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│           Domain Models                     │
│  (pkg/models - business logic)              │
│  • DbtModel, DbtColumn                      │
│  • LookMLView, LookMLDimension              │
│  • NestedArrayRules ← business rules        │
│  • DimensionConflictResolver                │
│  • ColumnHierarchy                          │
│  • ColumnClassifier                         │
└──────┬──────────────────────────────────────┘
       │
┌──────▼──────────────────────────────────────┐
│             Parsers                         │
│  (pkg/parsers - data extraction)            │
│  • ManifestParser                           │
│  • CatalogParser                            │
│  • DbtParser (coordinator)                  │
└─────────────────────────────────────────────┘
```

## Design Principles

### 1. **Domain-Driven Design (DDD)**

Business logic lives in the domain layer (`pkg/models`), not in generators.

**Example:**
```go
// ❌ BAD: Business logic in generator
if strings.Count(arrayName, ".") > 2 {
    continue  // Magic number, unclear intent
}

// ✅ GOOD: Business logic in domain service
rules := models.NewNestedArrayRules()
if !rules.ShouldProcessArray(arrayName) {
    continue  // Explicit rule, testable
}
```

### 2. **Separation of Concerns**

Each layer has a single responsibility:

| Layer | Responsibility | Example |
|-------|---------------|---------|
| **CLI** | User interaction, configuration | Parse flags, load config |
| **Generators** | Coordinate generation flow | Call domain services, write files |
| **Domain** | Business rules and logic | Classify columns, resolve conflicts |
| **Parsers** | Extract data from dbt files | Read JSON, parse structures |
| **Utils** | Pure utility functions | String conversion, helpers |

### 3. **Dependency Injection**

Services receive dependencies, don't create them:

```go
// ✅ GOOD: Dependencies injected
type ViewGenerator struct {
    config             *config.Config
    dimensionGenerator *DimensionGenerator
    // ...
}

// ✅ GOOD: Service created externally
resolver := models.NewDimensionConflictResolver()
result := resolver.Resolve(dimensions, dimensionGroups, modelName)
```

### 4. **Testability**

Each component can be tested independently:

```go
// Domain service test (no dependencies on generators)
func TestNestedArrayRules_ShouldProcessArray(t *testing.T) {
    rules := NewNestedArrayRules()
    assert.True(t, rules.ShouldProcessArray("items"))
    assert.True(t, rules.ShouldProcessArray("items.subitems"))
    assert.False(t, rules.ShouldProcessArray("items.subitems.details.meta"))
}
```

## Directory Structure

```
dbt2lookml/
├── cmd/
│   └── dbt2lookml/          # CLI entry point
│       └── main.go
│
├── internal/                # Private application code
│   ├── cli/                 # Command-line interface
│   │   ├── root.go          # Main command
│   │   └── version.go       # Version command
│   └── config/              # Configuration
│       └── config.go
│
├── pkg/                     # Public packages (importable)
│   ├── enums/               # Enumerations
│   │   └── looker.go        # LookML enums
│   │
│   ├── models/              # Domain models & business logic
│   │   ├── dbt.go           # dbt model types
│   │   ├── looker.go        # LookML types
│   │   ├── nested_array_rules.go           # ← Domain service
│   │   ├── dimension_conflict_resolver.go  # ← Domain service
│   │   ├── column_hierarchy.go             # ← Domain service
│   │   ├── column_classifier.go            # ← Domain service
│   │   └── column_collections.go
│   │
│   ├── generators/          # LookML generation
│   │   ├── generator.go     # Main coordinator
│   │   ├── view.go          # View generation
│   │   ├── dimension.go     # Dimension generation
│   │   ├── measure.go       # Measure generation
│   │   ├── explore.go       # Explore generation
│   │   ├── error_strategy.go # ← Error handling
│   │   └── interfaces.go    # Generator interfaces
│   │
│   ├── parsers/             # dbt file parsing
│   │   ├── base.go          # Base parser
│   │   ├── manifest.go      # Manifest parser
│   │   ├── catalog.go       # Catalog parser
│   │   └── model.go         # Model parser
│   │
│   └── utils/               # Utility functions
│       ├── string_utils.go  # String conversions
│       └── pointers.go      # Pointer helpers
│
├── tests/                   # Tests
│   ├── integration/         # Integration tests
│   └── fixtures/            # Test data
│
└── docs/                    # Documentation
    ├── public/              # User documentation
    └── development/         # Developer documentation
```

## Key Components

### 1. LookMLGenerator (Coordinator)

**Role:** Orchestrates the entire generation process

**Responsibilities:**
- Create output directories
- Process each model
- Coordinate other generators
- Handle errors according to strategy
- Report results

**Does NOT:**
- Contain business logic
- Make classification decisions
- Resolve naming conflicts

### 2. Domain Services (Business Logic)

#### NestedArrayRules

**Purpose:** Enforce array nesting depth limits

```go
rules := models.NewNestedArrayRules()
if rules.ShouldProcessArray("items.subitems.details") {
    // Process (level 3 - allowed)
}
// "items.subitems.details.meta" - level 4 - skipped
```

#### DimensionConflictResolver

**Purpose:** Resolve dimension/dimension_group name conflicts

```go
resolver := models.NewDimensionConflictResolver()
resolvedDimensions := resolver.Resolve(dimensions, dimensionGroups, modelName)
// Conflicting dimensions renamed with "_conflict" suffix and hidden
```

#### ColumnHierarchy

**Purpose:** Understand column structure and relationships

```go
hierarchy := models.NewColumnHierarchy(columns)
if hierarchy.IsArray("items") {
    children := hierarchy.GetChildren("items")
    // Process children
}
```

#### ColumnClassifier

**Purpose:** Classify columns based on business rules

```go
classifier := models.NewColumnClassifier(hierarchy, arrayColumns)
category := classifier.Classify("address", column)
// Returns: CategoryExcluded, CategoryMainView, or CategoryNestedView
```

### 3. Error Handling System

**Three Strategies:**

```go
type ErrorStrategy int

const (
    FailFast         // Stop on first error
    FailAtEnd        // Collect errors, fail at end
    ContinueOnError  // Log errors, don't fail
)
```

**Usage:**

```go
opts := GenerationOptions{
    ErrorStrategy: FailAtEnd,
    MaxErrors:     10,  // Stop after 10 errors
    Verbose:       true,
}

result, err := generator.GenerateAllWithOptions(ctx, models, opts)
// result.FilesGenerated, result.Errors, result.ModelsProcessed
```

## Data Flow

### 1. Parsing Phase

```
dbt files → Parser → Domain Models
```

```go
manifest.json + catalog.json
    ↓
DbtParser.Parse()
    ↓
[]*models.DbtModel
```

### 2. Classification Phase

```
Domain Models → Classifiers → Organized Columns
```

```go
DbtModel.Columns
    ↓
ColumnHierarchy (structure analysis)
    ↓
ColumnClassifier (business rules)
    ↓
ColumnCollections (MainView / NestedView / Excluded)
```

### 3. Generation Phase

```
Organized Columns → Generators → LookML
```

```go
ColumnCollections
    ↓
ViewGenerator (coordinates)
    ├→ DimensionGenerator (creates dimensions)
    ├→ MeasureGenerator (creates measures)
    ├→ ExploreGenerator (creates explores)
    └→ DimensionConflictResolver (resolves conflicts)
    ↓
LookMLView
    ↓
WriteToFile("model.view.lkml")
```

## Design Patterns

### 1. **Strategy Pattern** (Error Handling)

Different strategies for the same operation:

```go
type ErrorStrategy int  // Strategy enum
type GenerationOptions struct { ErrorStrategy ErrorStrategy }
func GenerateAllWithOptions(..., opts GenerationOptions)
```

### 2. **Builder Pattern** (Configuration)

Flexible configuration building:

```go
opts := GenerationOptions{
    ErrorStrategy: FailAtEnd,
    MaxErrors:     10,
    Verbose:       true,
}
```

### 3. **Service Layer** (Domain Services)

Business logic encapsulated in services:

```go
// Services are stateless and reusable
rules := NewNestedArrayRules()
resolver := NewDimensionConflictResolver()
classifier := NewColumnClassifier(hierarchy, arrayColumns)
```

### 4. **Coordinator Pattern** (Generators)

High-level coordination, low-level delegation:

```go
// ViewGenerator coordinates but delegates
func (g *ViewGenerator) Generate(model *DbtModel) {
    dimensions := g.dimensionGenerator.Generate(...)
    measures := g.measureGenerator.Generate(...)
    resolved := g.resolver.Resolve(...)
    // Assemble and return
}
```

## Testing Strategy

### Unit Tests

Test each component independently:

```go
// pkg/models/*_test.go
// pkg/generators/*_test.go
// pkg/parsers/*_test.go
```

### Integration Tests

Test end-to-end scenarios:

```go
// tests/integration/integration_test.go
// Verifies actual LookML output matches fixtures
```

### Test Metrics

- **Unit Test Lines:** 1,090+ (55 functions, 175+ cases)
- **Coverage:** 72.6% overall, 85%+ for critical paths
- **Integration Tests:** 5 fixture files, full validation

## Performance Considerations

### Current Optimizations

1. **Pre-compiled Regexes:** String utilities use package-level regexes
2. **Minimal Allocations:** Reuse buffers where possible
3. **Context Support:** Cancellable operations

### Future Optimizations

See `docs/development/performance.md` for detailed analysis.

## Extension Points

### Adding a New Domain Service

1. Create file: `pkg/models/my_service.go`
2. Define interface and implementation
3. Add comprehensive tests: `pkg/models/my_service_test.go`
4. Inject into generator if needed
5. Document in this file

### Adding a New Generator

1. Implement `GeneratorInterface` from `pkg/generators/interfaces.go`
2. Create file: `pkg/generators/my_generator.go`
3. Add tests: `pkg/generators/my_generator_test.go`
4. Inject into `LookMLGenerator`

### Adding a New Error Strategy

1. Add to `ErrorStrategy` enum in `error_strategy.go`
2. Implement strategy in `GenerateAllWithOptions()`
3. Add tests
4. Document in user docs

## Related Documentation

- [Domain Services](domain-services.md) - Detailed service documentation
- [Testing Guide](testing.md) - Testing strategies and guidelines
- [Performance](performance.md) - Performance analysis and optimization

---

**Last Updated:** 2025-10-01 (Phase 1 & 2 refactoring complete)
