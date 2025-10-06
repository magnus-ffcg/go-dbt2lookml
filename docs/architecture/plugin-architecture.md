# Plugin Architecture

## Overview

The semantic models/metrics functionality is implemented as a **pluggable extension** to the core LookML generator. This architecture provides clear separation of concerns and makes the codebase more maintainable.

## Design Principles

### 1. Separation of Concerns
- **Core generator** handles basic LookML generation (views, dimensions, explores)
- **Metrics plugin** handles all semantic model and metric logic
- No coupling between core and plugin code

### 2. Optional Activation
- Plugin only initialized when `--use-semantic-models` flag is enabled
- Zero overhead when plugin is disabled
- Core functionality remains fast and focused

### 3. Clean Interfaces
- Plugin exposes clear methods for generation and data access
- Core generator doesn't need to know plugin internals
- Easy to add new plugins in the future

## Architecture Diagram

```
┌─────────────────────────────────────────────────────┐
│                  LookMLGenerator                     │
│  ┌───────────────────────────────────────────────┐  │
│  │           Core Generators (Always)            │  │
│  │  • ViewGenerator                              │  │
│  │  • DimensionGenerator                         │  │
│  │  • ExploreGenerator                           │  │
│  │  • MeasureGenerator                           │  │
│  └───────────────────────────────────────────────┘  │
│                                                      │
│  ┌───────────────────────────────────────────────┐  │
│  │         MetricsPlugin (Optional)              │  │
│  │  • Initialized only if UseSemanticModels      │  │
│  │  • Generates metric view extensions           │  │
│  │  • Generates cumulative views                 │  │
│  │  • Generates conversion views                 │  │
│  │  • Provides explore joins                     │  │
│  └───────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## File Structure

```
pkg/generators/
├── generator.go              # Core generator with optional plugin
├── view.go                   # View generation
├── dimension.go              # Dimension generation
├── explore.go                # Explore generation
├── measure.go                # Measure generation
└── plugins/
    ├── metrics_plugin.go     # Plugin interface and data management
    └── metrics_generator.go  # Metric view generation logic
```

## How It Works

### 1. Plugin Initialization

```go
func NewLookMLGenerator(cfg *config.Config) *LookMLGenerator {
    gen := &LookMLGenerator{
        config:           cfg,
        viewGenerator:    NewViewGenerator(cfg),
        exploreGenerator: NewExploreGenerator(cfg),
        // ... other core generators
    }
    
    // Initialize plugin only if enabled
    if cfg.UseSemanticModels {
        gen.metricsPlugin = plugins.NewMetricsPlugin(cfg)
    }
    
    return gen
}
```

### 2. Data Injection

When semantic model data is loaded, it's injected into both the generator and the plugin:

```go
func (g *LookMLGenerator) SetSemanticMeasures(semanticMeasures map[string][]models.DbtSemanticMeasure) {
    g.semanticMeasures = semanticMeasures  // Backward compatibility
    
    // Forward to plugin if enabled
    if g.metricsPlugin != nil {
        g.metricsPlugin.SetSemanticMeasures(semanticMeasures)
    }
}
```

### 3. Generation Flow

```go
func (g *LookMLGenerator) generateViewFile(model *models.DbtModel) error {
    // 1. Core view generation
    g.viewGenerator.Generate(model)
    
    // 2. Core explore generation
    explore := g.exploreGenerator.GenerateExplore(model)
    
    // 3. Plugin extension (if enabled)
    if g.metricsPlugin != nil {
        // Add metric joins to explore
        metricJoins := g.metricsPlugin.GetExploreJoins(model, explore.Name)
        explore.Joins = append(explore.Joins, metricJoins...)
    }
    
    return nil
}

func (g *LookMLGenerator) generateViewExtensionFile(model *models.DbtModel) error {
    // Generate view extension with semantic measures...
    
    // Plugin generates separate views (if enabled)
    if g.metricsPlugin != nil {
        g.metricsPlugin.GenerateForModel(model)
    }
    
    return nil
}
```

## Plugin Interface

### Core Methods

```go
type MetricsPlugin struct {
    config            *config.Config
    semanticMeasures  map[string][]models.DbtSemanticMeasure
    cumulativeMetrics []models.DbtMetric
    conversionMetrics []models.DbtMetric
    // ...
}

// Check if plugin is enabled
func (p *MetricsPlugin) Enabled() bool

// Generate all metric views for a model
func (p *MetricsPlugin) GenerateForModel(model *models.DbtModel) error

// Get explore joins for metric views
func (p *MetricsPlugin) GetExploreJoins(model *models.DbtModel, baseName string) []models.LookMLJoin

// Data accessors
func (p *MetricsPlugin) GetSemanticMeasures(modelName string) []models.DbtSemanticMeasure
func (p *MetricsPlugin) GetCumulativeMetrics() []models.DbtMetric
func (p *MetricsPlugin) GetConversionMetrics() []models.DbtMetric
```

### Generation Methods

```go
// Internal generation methods
func (p *MetricsPlugin) generateCumulativeViewFile(model, measures) error
func (p *MetricsPlugin) generateConversionViewFile(model, measures) error
func (p *MetricsPlugin) buildWindowFunctionSQL(metric, measure) string
```

## Benefits

### 1. Clean Core Generator

**Before (tightly coupled):**
```go
type LookMLGenerator struct {
    config                 *config.Config
    viewGenerator          *ViewGenerator
    semanticMeasures       map[string][]models.DbtSemanticMeasure
    ratioMetrics           []models.DbtMetric
    derivedMetrics         []models.DbtMetric
    simpleMetrics          []models.DbtMetric
    cumulativeMetrics      []models.DbtMetric
    conversionMetrics      []models.DbtMetric
    // 6+ metric-related fields polluting core struct!
}
```

**After (plugin pattern):**
```go
type LookMLGenerator struct {
    config           *config.Config
    viewGenerator    *ViewGenerator
    metricsPlugin    *plugins.MetricsPlugin  // Single optional plugin
}
```

### 2. Easy to Disable

```bash
# Without semantic models - plugin never created
dbt2lookml --manifest manifest.json --catalog catalog.json

# With semantic models - plugin initialized and used
dbt2lookml --manifest manifest.json --catalog catalog.json --use-semantic-models
```

### 3. Testability

```go
// Test core generator without metrics
func TestCoreGenerator(t *testing.T) {
    cfg := &config.Config{UseSemanticModels: false}
    gen := NewLookMLGenerator(cfg)
    // No metric code involved
}

// Test plugin independently
func TestMetricsPlugin(t *testing.T) {
    plugin := plugins.NewMetricsPlugin(cfg)
    plugin.SetCumulativeMetrics(metrics)
    // Test plugin in isolation
}
```

### 4. Extensibility

Easy to add more plugins in the future:

```go
type LookMLGenerator struct {
    config           *config.Config
    viewGenerator    *ViewGenerator
    
    // Plugins
    metricsPlugin    *plugins.MetricsPlugin    // Existing
    tableauPlugin    *plugins.TableauPlugin    // New!
    powerbiPlugin    *plugins.PowerBIPlugin    // New!
}
```

## Backward Compatibility

The refactoring maintains **100% backward compatibility**:

### 1. Same Public API

```go
// All existing methods still work
generator.SetSemanticMeasures(measures)
generator.SetCumulativeMetrics(metrics)
generator.GenerateAll(models)
```

### 2. Same Output

Generated files are identical:
- `orders.view.lkml` - base view with explore
- `orders__metrics.view.lkml` - view extension
- `orders__cumulative.view.lkml` - cumulative metrics view
- `orders__conversion.view.lkml` - conversion metrics view

### 3. Same Behavior

- Plugin only active when `--use-semantic-models` flag is used
- No performance impact when disabled
- All existing tests pass

## Migration Path

The transition was done incrementally:

1. ✅ Create plugin package structure
2. ✅ Move cumulative metrics generation to plugin
3. ✅ Move conversion metrics generation to plugin
4. ✅ Update core generator to use plugin
5. ✅ Keep backward compatibility (deprecated fields still exist)
6. 🔄 Future: Remove deprecated fields after confirming stability

## Future Enhancements

### 1. Plugin Registry

```go
type PluginRegistry struct {
    plugins []Plugin
}

func (r *PluginRegistry) Register(plugin Plugin)
func (r *PluginRegistry) GenerateForModel(model *models.DbtModel)
```

### 2. Plugin Interface

```go
type Plugin interface {
    Name() string
    Enabled() bool
    GenerateForModel(model *models.DbtModel) error
    GetExploreJoins(model *models.DbtModel) []models.LookMLJoin
}
```

### 3. Configuration-Based Plugins

```yaml
plugins:
  - name: semantic_metrics
    enabled: true
  - name: tableau_export
    enabled: true
    config:
      output_format: "tdsx"
```

## Performance Impact

### With Plugin Disabled (Default)

- **Memory**: No overhead (plugin = nil)
- **CPU**: No plugin code executed
- **Output**: Only core files generated

### With Plugin Enabled

- **Memory**: ~5-10MB for metric data structures
- **CPU**: Additional processing for metric views
- **Output**: Core + 2-3 additional views per model

## Testing

### Integration Tests

```bash
# Test with plugin enabled
go run cmd/dbt2lookml/main.go \
  --manifest manifest_semantic.json \
  --use-semantic-models

# Test with plugin disabled (backward compatibility)
go run cmd/dbt2lookml/main.go \
  --manifest manifest.json

# Verify output
diff core_output/ plugin_output/  # Should show additional metric files only
```

### Unit Tests

```go
func TestPluginIsolation(t *testing.T) {
    // Core without plugin
    cfg := &config.Config{UseSemanticModels: false}
    gen := NewLookMLGenerator(cfg)
    assert.Nil(t, gen.metricsPlugin)
    
    // Core with plugin
    cfg.UseSemanticModels = true
    gen = NewLookMLGenerator(cfg)
    assert.NotNil(t, gen.metricsPlugin)
}
```

## Summary

The plugin architecture provides:

✅ **Clean separation** - Core and metrics logic decoupled  
✅ **Optional activation** - Zero overhead when disabled  
✅ **Backward compatible** - Existing API unchanged  
✅ **Extensible** - Easy to add new plugins  
✅ **Testable** - Components can be tested independently  
✅ **Maintainable** - Clear ownership of code  

This architecture sets a solid foundation for future enhancements while keeping the codebase clean and maintainable.
