package tests

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	pluginMetrics "github.com/magnus-ffcg/go-dbt2lookml/pkg/plugins/metrics"
)

// TestPluginHookLifecycle tests the complete lifecycle of plugin hooks
// from registration through data ingestion to file generation
func TestPluginHookLifecycle(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	cfg := &config.Config{
		OutputDir:         tempDir,
		UseSemanticModels: true,
	}

	// Create generator
	gen := generators.NewLookMLGenerator(cfg)

	// Register metrics plugin
	metricsPlugin := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(metricsPlugin)

	// Simulate data ingestion phase (as done by CLI)
	semanticMeasures := map[string][]models.DbtSemanticMeasure{
		"orders": {
			{Name: "total_revenue", Agg: "sum"},
			{Name: "order_count", Agg: "count"},
		},
	}
	gen.SetSemanticMeasures(semanticMeasures)

	ratioMetrics := []models.DbtMetric{
		{
			Name: "revenue_per_order",
			Type: "ratio",
			TypeParams: models.DbtMetricTypeParams{
				Numerator: &models.DbtMetricInput{
					Name: "total_revenue",
				},
				Denominator: &models.DbtMetricInput{
					Name: "order_count",
				},
			},
		},
	}
	gen.SetRatioMetrics(ratioMetrics)

	cumulativeMetrics := []models.DbtMetric{
		{
			Name: "cumulative_revenue",
			Type: "cumulative",
			TypeParams: models.DbtMetricTypeParams{
				Measure: &models.DbtMetricInputMeasure{
					Name: "total_revenue",
				},
			},
		},
	}
	gen.SetCumulativeMetrics(cumulativeMetrics)

	// Create test model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
		RelationName: "`project.dataset.orders`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)
	model.Columns["order_id"] = models.DbtModelColumn{
		Name: "order_id",
	}
	model.Columns["revenue"] = models.DbtModelColumn{
		Name: "revenue",
	}

	// Generate files (triggers all hooks)
	ctx := context.Background()
	_, err := gen.GenerateAllWithContext(ctx, []*models.DbtModel{model})
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Verify files were generated
	t.Run("BaseViewGenerated", func(t *testing.T) {
		baseViewPath := filepath.Join(tempDir, "orders.view.lkml")
		if _, err := os.Stat(baseViewPath); os.IsNotExist(err) {
			t.Error("Base view file was not generated")
		}
	})

	t.Run("MetricsViewGenerated", func(t *testing.T) {
		metricsViewPath := filepath.Join(tempDir, "orders__metrics.view.lkml")
		content, err := os.ReadFile(metricsViewPath)
		if err != nil {
			t.Fatalf("Metrics view file was not generated: %v", err)
		}

		contentStr := string(content)

		// Verify semantic measures
		if !strings.Contains(contentStr, "total_revenue") {
			t.Error("Metrics view should contain total_revenue measure")
		}
		if !strings.Contains(contentStr, "order_count") {
			t.Error("Metrics view should contain order_count measure")
		}

		// Verify ratio metric
		if !strings.Contains(contentStr, "revenue_per_order") {
			t.Error("Metrics view should contain revenue_per_order ratio metric")
		}
	})

	t.Run("CumulativeViewGenerated", func(t *testing.T) {
		cumulativeViewPath := filepath.Join(tempDir, "orders__cumulative.view.lkml")
		content, err := os.ReadFile(cumulativeViewPath)
		if err != nil {
			t.Fatalf("Cumulative view file was not generated: %v", err)
		}

		contentStr := string(content)
		if !strings.Contains(contentStr, "cumulative_revenue") {
			t.Error("Cumulative view should contain cumulative_revenue metric")
		}
	})

	t.Run("ExploreEnrichedWithJoins", func(t *testing.T) {
		baseViewPath := filepath.Join(tempDir, "orders.view.lkml")
		content, err := os.ReadFile(baseViewPath)
		if err != nil {
			t.Fatalf("Failed to read base view: %v", err)
		}

		contentStr := string(content)

		// Verify explore has cumulative join
		if !strings.Contains(contentStr, "orders__cumulative") {
			t.Error("Explore should contain join to cumulative metrics view")
		}
	})
}

// TestMultiplePlugins tests that multiple plugins can coexist
func TestMultiplePlugins(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		OutputDir:         tempDir,
		UseSemanticModels: true,
	}

	gen := generators.NewLookMLGenerator(cfg)

	// Register multiple plugins (currently just metrics, but demonstrates pattern)
	plugin1 := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(plugin1)

	// Could register additional plugins here
	// plugin2 := otherPlugin.NewPlugin(cfg)
	// gen.RegisterPlugin(plugin2)

	// Verify both receive hooks
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {{Name: "revenue", Agg: "sum"}},
	}
	gen.SetSemanticMeasures(measures)

	// Verify plugin1 received the data
	if len(plugin1.GetSemanticMeasures("orders")) != 1 {
		t.Error("Plugin should have received semantic measures")
	}
}

// TestPluginDisabled tests that disabled plugins don't participate
func TestPluginDisabled(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		OutputDir:         tempDir,
		UseSemanticModels: false, // Disabled
	}

	gen := generators.NewLookMLGenerator(cfg)
	plugin := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(plugin)

	if plugin.Enabled() {
		t.Error("Plugin should be disabled when UseSemanticModels is false")
	}

	// Create and generate model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
		RelationName: "`project.dataset.orders`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)

	ctx := context.Background()
	_, err := gen.GenerateAllWithContext(ctx, []*models.DbtModel{model})
	if err != nil {
		t.Fatalf("Generation failed: %v", err)
	}

	// Verify no plugin files were generated
	metricsFile := filepath.Join(tempDir, "orders__metrics.view.lkml")
	if _, err := os.Stat(metricsFile); !os.IsNotExist(err) {
		t.Error("Metrics file should not be generated when plugin is disabled")
	}
}

// TestHookErrorHandling tests that hook errors don't crash generation
func TestHookErrorHandling(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		OutputDir:         tempDir,
		UseSemanticModels: true,
	}

	gen := generators.NewLookMLGenerator(cfg)
	plugin := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(plugin)

	// Invalid model (missing required fields) might cause plugin errors
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "", // Invalid - empty name
		},
	}

	ctx := context.Background()

	// Should not panic, even if hooks encounter errors
	_, err := gen.GenerateAllWithContext(ctx, []*models.DbtModel{model})

	// We expect an error, but not a panic
	if err == nil {
		t.Log("Generation completed despite invalid model (hooks handled gracefully)")
	}
}
