package metrics

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

func TestMetricsPlugin_ImplementsHookInterfaces(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Verify plugin implements all required interfaces
	if plugin.Name() != "SemanticMetrics" {
		t.Errorf("Expected plugin name 'SemanticMetrics', got '%s'", plugin.Name())
	}

	if !plugin.Enabled() {
		t.Error("Plugin should be enabled when UseSemanticModels is true")
	}
}

func TestMetricsPlugin_OnSemanticMeasures(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Create test measures
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {
			{Name: "total_revenue", Agg: "sum"},
			{Name: "order_count", Agg: "count"},
		},
		"customers": {
			{Name: "customer_count", Agg: "count_distinct"},
		},
	}

	// Fire hook
	plugin.OnSemanticMeasures(measures)

	// Verify data was stored
	if len(plugin.semanticMeasures) != 2 {
		t.Errorf("Expected 2 models in semanticMeasures, got %d", len(plugin.semanticMeasures))
	}

	ordersMeasures := plugin.GetSemanticMeasures("orders")
	if len(ordersMeasures) != 2 {
		t.Errorf("Expected 2 measures for orders, got %d", len(ordersMeasures))
	}
}

func TestMetricsPlugin_OnMetrics(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Test different metric types
	testCases := []struct {
		metricType string
		verify     func() int
	}{
		{"ratio", func() int { return len(plugin.GetRatioMetrics()) }},
		{"derived", func() int { return len(plugin.GetDerivedMetrics()) }},
		{"simple", func() int { return len(plugin.GetSimpleMetrics()) }},
		{"cumulative", func() int { return len(plugin.GetCumulativeMetrics()) }},
		{"conversion", func() int { return len(plugin.GetConversionMetrics()) }},
	}

	for _, tc := range testCases {
		t.Run(tc.metricType, func(t *testing.T) {
			metrics := []models.DbtMetric{
				{Name: "test_metric_1"},
				{Name: "test_metric_2"},
			}

			plugin.OnMetrics(metrics, tc.metricType)

			if count := tc.verify(); count != 2 {
				t.Errorf("Expected 2 %s metrics, got %d", tc.metricType, count)
			}
		})
	}
}

func TestMetricsPlugin_AfterModelGeneration(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Set up test data
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {
			{Name: "total_revenue", Agg: "sum"},
			{Name: "order_count", Agg: "count"},
		},
	}
	plugin.OnSemanticMeasures(measures)

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
	plugin.OnMetrics(ratioMetrics, "ratio")

	// Create test model
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
		RelationName: "`project.dataset.orders`",
	}
	model.Columns = make(map[string]models.DbtModelColumn)

	ctx := context.Background()

	// Fire hook
	err := plugin.AfterModelGeneration(ctx, model)
	if err != nil {
		t.Fatalf("AfterModelGeneration failed: %v", err)
	}

	// Verify files were generated
	metricsFile := cfg.GetOutputPath("orders__metrics.view.lkml")
	if _, err := os.Stat(metricsFile); os.IsNotExist(err) {
		t.Error("Expected metrics view file to be generated")
	}

	// Verify content
	content, err := os.ReadFile(metricsFile)
	if err != nil {
		t.Fatalf("Failed to read metrics file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "total_revenue") {
		t.Error("Metrics file should contain total_revenue measure")
	}
	if !strings.Contains(contentStr, "revenue_per_order") {
		t.Error("Metrics file should contain revenue_per_order ratio metric")
	}
}

func TestMetricsPlugin_EnrichExplore(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Set up semantic measures for the model (required for metric detection)
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {
			{Name: "total_revenue", Agg: "sum"},
		},
	}
	plugin.OnSemanticMeasures(measures)

	// Set up cumulative metrics that reference the measure
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
	plugin.OnMetrics(cumulativeMetrics, "cumulative")

	// Create test model and explore
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
	}

	explore := &models.LookMLExplore{
		Name:     "orders",
		ViewName: "orders",
		Joins:    []models.LookMLJoin{},
	}

	ctx := context.Background()

	// Fire hook
	err := plugin.EnrichExplore(ctx, model, explore, "orders")
	if err != nil {
		t.Fatalf("EnrichExplore failed: %v", err)
	}

	// Verify joins were added
	if len(explore.Joins) == 0 {
		t.Error("Expected explore to have joins added by plugin")
	}

	// Verify cumulative join was added
	foundCumulativeJoin := false
	for _, join := range explore.Joins {
		if strings.Contains(join.Name, "cumulative") {
			foundCumulativeJoin = true
			break
		}
	}

	if !foundCumulativeJoin {
		t.Error("Expected explore to have cumulative metrics join")
	}
}

func TestMetricsPlugin_Disabled(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: false, // Disabled
	}

	plugin := NewMetricsPlugin(cfg)

	if plugin.Enabled() {
		t.Error("Plugin should be disabled when UseSemanticModels is false")
	}

	// Set up test data
	measures := map[string][]models.DbtSemanticMeasure{
		"orders": {{Name: "total_revenue", Agg: "sum"}},
	}
	plugin.OnSemanticMeasures(measures)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
	}

	ctx := context.Background()

	// Fire hooks - should be no-ops
	err := plugin.AfterModelGeneration(ctx, model)
	if err != nil {
		t.Fatalf("AfterModelGeneration should not error when disabled: %v", err)
	}

	explore := &models.LookMLExplore{
		Name:  "orders",
		Joins: []models.LookMLJoin{},
	}

	err = plugin.EnrichExplore(ctx, model, explore, "orders")
	if err != nil {
		t.Fatalf("EnrichExplore should not error when disabled: %v", err)
	}

	// Verify no files were generated
	metricsFile := cfg.GetOutputPath("orders__metrics.view.lkml")
	if _, err := os.Stat(metricsFile); !os.IsNotExist(err) {
		t.Error("Expected no metrics file when plugin is disabled")
	}

	// Verify no joins were added
	if len(explore.Joins) != 0 {
		t.Error("Expected no joins when plugin is disabled")
	}
}

func TestMetricsPlugin_MultipleCalls(t *testing.T) {
	cfg := &config.Config{
		OutputDir:         t.TempDir(),
		UseSemanticModels: true,
	}

	plugin := NewMetricsPlugin(cfg)

	// Call OnMetrics multiple times with different data
	metrics1 := []models.DbtMetric{{Name: "metric1"}}
	metrics2 := []models.DbtMetric{{Name: "metric2"}}

	plugin.OnMetrics(metrics1, "ratio")
	plugin.OnMetrics(metrics2, "ratio") // Should replace, not append

	ratioMetrics := plugin.GetRatioMetrics()
	if len(ratioMetrics) != 1 {
		t.Errorf("Expected 1 ratio metric (replaced), got %d", len(ratioMetrics))
	}

	if ratioMetrics[0].Name != "metric2" {
		t.Errorf("Expected last metric to be 'metric2', got '%s'", ratioMetrics[0].Name)
	}
}
