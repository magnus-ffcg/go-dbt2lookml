package generators

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// MockPlugin is a test plugin that implements all hook interfaces
type MockPlugin struct {
	name    string
	enabled bool

	// Track hook calls
	semanticMeasuresCalled bool
	metricsCalled          map[string]int // metricType -> call count
	afterGenCalled         int
	enrichExploreCalled    int

	// Store data
	storedMeasures map[string][]models.DbtSemanticMeasure
	storedMetrics  map[string][]models.DbtMetric
}

func NewMockPlugin(name string, enabled bool) *MockPlugin {
	return &MockPlugin{
		name:           name,
		enabled:        enabled,
		metricsCalled:  make(map[string]int),
		storedMeasures: make(map[string][]models.DbtSemanticMeasure),
		storedMetrics:  make(map[string][]models.DbtMetric),
	}
}

func (p *MockPlugin) Name() string {
	return p.name
}

func (p *MockPlugin) Enabled() bool {
	return p.enabled
}

func (p *MockPlugin) OnSemanticMeasures(measures map[string][]models.DbtSemanticMeasure) {
	p.semanticMeasuresCalled = true
	p.storedMeasures = measures
}

func (p *MockPlugin) OnMetrics(metrics []models.DbtMetric, metricType string) {
	p.metricsCalled[metricType]++
	p.storedMetrics[metricType] = metrics
}

func (p *MockPlugin) AfterModelGeneration(ctx context.Context, model *models.DbtModel) error {
	p.afterGenCalled++
	return nil
}

func (p *MockPlugin) EnrichExplore(ctx context.Context, model *models.DbtModel, explore *models.LookMLExplore, baseName string) error {
	p.enrichExploreCalled++
	// Add a test join
	explore.Joins = append(explore.Joins, models.LookMLJoin{
		Name: "test_join_from_plugin",
	})
	return nil
}

func TestPluginRegistration(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	// Initially no plugins
	if len(gen.plugins) != 0 {
		t.Errorf("Expected 0 plugins initially, got %d", len(gen.plugins))
	}

	// Register one plugin
	plugin1 := NewMockPlugin("TestPlugin1", true)
	gen.RegisterPlugin(plugin1)

	if len(gen.plugins) != 1 {
		t.Errorf("Expected 1 plugin after registration, got %d", len(gen.plugins))
	}

	// Register another plugin
	plugin2 := NewMockPlugin("TestPlugin2", true)
	gen.RegisterPlugin(plugin2)

	if len(gen.plugins) != 2 {
		t.Errorf("Expected 2 plugins after second registration, got %d", len(gen.plugins))
	}
}

func TestDataIngestionHook_SemanticMeasures(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	// Register enabled and disabled plugins
	enabledPlugin := NewMockPlugin("EnabledPlugin", true)
	disabledPlugin := NewMockPlugin("DisabledPlugin", false)
	gen.RegisterPlugin(enabledPlugin)
	gen.RegisterPlugin(disabledPlugin)

	// Create test data
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {
			{Name: "total_revenue", Agg: "sum"},
		},
	}

	// Fire the hook
	gen.SetSemanticMeasures(measures)

	// Verify enabled plugin received the hook
	if !enabledPlugin.semanticMeasuresCalled {
		t.Error("Expected enabled plugin to receive OnSemanticMeasures hook")
	}

	if len(enabledPlugin.storedMeasures) != 1 {
		t.Errorf("Expected enabled plugin to store 1 model's measures, got %d", len(enabledPlugin.storedMeasures))
	}

	// Verify disabled plugin did NOT receive the hook
	if disabledPlugin.semanticMeasuresCalled {
		t.Error("Expected disabled plugin to NOT receive OnSemanticMeasures hook")
	}
}

func TestDataIngestionHook_Metrics(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	plugin := NewMockPlugin("TestPlugin", true)
	gen.RegisterPlugin(plugin)

	// Test different metric types
	testCases := []struct {
		metricType string
		setter     func([]models.DbtMetric)
	}{
		{"ratio", gen.SetRatioMetrics},
		{"derived", gen.SetDerivedMetrics},
		{"simple", gen.SetSimpleMetrics},
		{"cumulative", gen.SetCumulativeMetrics},
		{"conversion", gen.SetConversionMetrics},
	}

	for _, tc := range testCases {
		t.Run(tc.metricType, func(t *testing.T) {
			metrics := []models.DbtMetric{
				{Name: "test_metric"},
			}

			tc.setter(metrics)

			if plugin.metricsCalled[tc.metricType] != 1 {
				t.Errorf("Expected %s metrics hook to be called once, got %d",
					tc.metricType, plugin.metricsCalled[tc.metricType])
			}

			if len(plugin.storedMetrics[tc.metricType]) != 1 {
				t.Errorf("Expected plugin to store 1 %s metric, got %d",
					tc.metricType, len(plugin.storedMetrics[tc.metricType]))
			}
		})
	}
}

func TestModelGenerationHook(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	plugin := NewMockPlugin("TestPlugin", true)
	gen.RegisterPlugin(plugin)

	// Create a test model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		RelationName: "`test_db.test_schema.test_table`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)

	ctx := context.Background()

	// Generate view file (which fires AfterModelGeneration hook)
	err := gen.generateViewFile(ctx, model)
	if err != nil {
		t.Fatalf("generateViewFile failed: %v", err)
	}

	// Verify hook was called
	if plugin.afterGenCalled != 1 {
		t.Errorf("Expected AfterModelGeneration hook to be called once, got %d", plugin.afterGenCalled)
	}
}

func TestExploreEnrichmentHook(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	plugin := NewMockPlugin("TestPlugin", true)
	gen.RegisterPlugin(plugin)

	// Create a test model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		RelationName: "`test_db.test_schema.test_table`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)

	ctx := context.Background()

	// Generate view file (which creates explore and fires EnrichExplore hook)
	err := gen.generateViewFile(ctx, model)
	if err != nil {
		t.Fatalf("generateViewFile failed: %v", err)
	}

	// Verify hook was called
	if plugin.enrichExploreCalled != 1 {
		t.Errorf("Expected EnrichExplore hook to be called once, got %d", plugin.enrichExploreCalled)
	}

	// Verify the generated file contains the join added by the plugin
	filePath := cfg.GetOutputPath(model.Name + ".view.lkml")
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read generated file: %v", err)
	}

	if !strings.Contains(string(content), "test_join_from_plugin") {
		t.Error("Expected explore to contain join added by plugin")
	}
}

func TestMultiplePluginsReceiveHooks(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	// Register multiple plugins
	plugin1 := NewMockPlugin("Plugin1", true)
	plugin2 := NewMockPlugin("Plugin2", true)
	plugin3 := NewMockPlugin("Plugin3", false) // disabled
	gen.RegisterPlugin(plugin1)
	gen.RegisterPlugin(plugin2)
	gen.RegisterPlugin(plugin3)

	// Fire data ingestion hook
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {{Name: "revenue", Agg: "sum"}},
	}
	gen.SetSemanticMeasures(measures)

	// Verify both enabled plugins received the hook
	if !plugin1.semanticMeasuresCalled {
		t.Error("Plugin1 should have received hook")
	}
	if !plugin2.semanticMeasuresCalled {
		t.Error("Plugin2 should have received hook")
	}
	if plugin3.semanticMeasuresCalled {
		t.Error("Plugin3 (disabled) should NOT have received hook")
	}
}

func TestPluginHookOrder(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	// Register plugins in specific order
	plugin1 := NewMockPlugin("First", true)
	plugin2 := NewMockPlugin("Second", true)
	gen.RegisterPlugin(plugin1)
	gen.RegisterPlugin(plugin2)

	// Create test model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test",
		},
		RelationName: "`db.schema.table`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)

	ctx := context.Background()

	// Generate (fires hooks)
	err := gen.generateViewFile(ctx, model)
	if err != nil {
		t.Fatalf("generateViewFile failed: %v", err)
	}

	// Both plugins should be called
	if plugin1.afterGenCalled != 1 {
		t.Error("First plugin should be called")
	}
	if plugin2.afterGenCalled != 1 {
		t.Error("Second plugin should be called")
	}
}
