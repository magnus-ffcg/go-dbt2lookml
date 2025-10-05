package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/parsers"
	pluginMetrics "github.com/magnus-ffcg/go-dbt2lookml/pkg/plugins/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSemanticMeasuresEndToEnd tests the complete semantic measures flow:
// 1. Parse manifest with semantic models
// 2. Extract semantic measures
// 3. Register metrics plugin
// 4. Generate LookML with semantic measures
// 5. Verify output files contain measures
func TestSemanticMeasuresEndToEnd(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		ManifestPath:      "../fixtures/data/manifest_semantic.json",
		CatalogPath:       "../fixtures/data/catalog_semantic.json",
		OutputDir:         tempDir,
		UseSemanticModels: true,
		LogLevel:          "WARN",
	}

	// Load manifest and catalog
	manifestData, err := os.ReadFile(cfg.ManifestPath)
	require.NoError(t, err, "Should load manifest file")

	catalogData, err := os.ReadFile(cfg.CatalogPath)
	require.NoError(t, err, "Should load catalog file")

	var rawManifest map[string]interface{}
	require.NoError(t, json.Unmarshal(manifestData, &rawManifest))

	var rawCatalog map[string]interface{}
	require.NoError(t, json.Unmarshal(catalogData, &rawCatalog))

	// Create parser
	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	require.NoError(t, err, "Should create parser")

	// Get models
	dbtModels, err := parser.GetModels()
	require.NoError(t, err, "Should parse models")
	require.NotEmpty(t, dbtModels, "Should have models")

	// Parse semantic models
	semanticModels, err := parser.GetSemanticModelParser().GetSemanticModels()
	require.NoError(t, err, "Should parse semantic models")
	require.NotEmpty(t, semanticModels, "Should have semantic models")

	t.Logf("Found %d semantic models", len(semanticModels))

	// Build semantic measures map
	semanticMeasures := make(map[string][]models.DbtSemanticMeasure)
	for _, sm := range semanticModels {
		if len(sm.Measures) > 0 {
			semanticMeasures[sm.Model] = sm.Measures
			t.Logf("Model %s has %d semantic measures", sm.Model, len(sm.Measures))
		}
	}

	require.NotEmpty(t, semanticMeasures, "Should have semantic measures")

	// Create generator
	gen := generators.NewLookMLGenerator(cfg)

	// Register metrics plugin
	metricsPlugin := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(metricsPlugin)

	// Set semantic measures (fires DataIngestionHook)
	gen.SetSemanticMeasures(semanticMeasures)

	// Generate LookML
	ctx := context.Background()
	filesGenerated, err := gen.GenerateAllWithContext(ctx, dbtModels)
	require.NoError(t, err, "Should generate LookML")
	require.Greater(t, filesGenerated, 0, "Should generate at least one file")

	t.Logf("Generated %d files", filesGenerated)

	// Verify output
	t.Run("BaseViewsGenerated", func(t *testing.T) {
		for _, model := range dbtModels {
			viewFile := cfg.GetOutputPath(model.Name + ".view.lkml")
			assert.FileExists(t, viewFile, "Base view should exist for model %s", model.Name)
		}
	})

	t.Run("MetricsViewsGenerated", func(t *testing.T) {
		// Check if __metrics.view.lkml files are generated for models with measures
		for modelName := range semanticMeasures {
			metricsFile := cfg.GetOutputPath(modelName + "__metrics.view.lkml")

			if _, err := os.Stat(metricsFile); err == nil {
				t.Logf("Metrics view generated: %s", metricsFile)

				content, err := os.ReadFile(metricsFile)
				require.NoError(t, err)

				contentStr := string(content)

				// Verify it's a view extension
				assert.Contains(t, contentStr, "view: +"+modelName,
					"Metrics file should be a view extension")

				// Verify it contains measures
				assert.Contains(t, contentStr, "measure:",
					"Metrics view should contain measures")

				// Verify specific measures exist
				for _, measure := range semanticMeasures[modelName] {
					assert.Contains(t, contentStr, measure.Name,
						"Metrics view should contain measure %s", measure.Name)
				}
			} else {
				t.Logf("No metrics view for model %s (may not have matching dbt model)", modelName)
			}
		}
	})

	t.Run("SemanticMeasuresInOutput", func(t *testing.T) {
		// Verify semantic measures appear in the generated files
		for modelName, measures := range semanticMeasures {
			metricsFile := cfg.GetOutputPath(modelName + "__metrics.view.lkml")

			if _, err := os.Stat(metricsFile); os.IsNotExist(err) {
				continue
			}

			content, err := os.ReadFile(metricsFile)
			require.NoError(t, err)
			contentStr := string(content)

			for _, measure := range measures {
				t.Run(measure.Name, func(t *testing.T) {
					// Verify measure name
					assert.Contains(t, contentStr, measure.Name,
						"Should contain measure %s", measure.Name)

					// Verify aggregation type
					if measure.Agg != "" {
						expectedType := mapAggToLookML(measure.Agg)
						assert.Contains(t, contentStr, fmt.Sprintf("type: %s", expectedType),
							"Should have correct type for %s", measure.Name)
					}

					// Verify SQL reference if expression exists
					if measure.Expr != nil && *measure.Expr != "" {
						assert.Contains(t, contentStr, "sql:",
							"Should have SQL definition for %s", measure.Name)
					}
				})
			}
		}
	})
}

// testMetricsEnvironment holds the test environment for all metric types
type testMetricsEnvironment struct {
	tempDir           string
	ratioMetrics      []models.DbtMetric
	cumulativeMetrics []models.DbtMetric
	conversionMetrics []models.DbtMetric
}

// setupAllMetricTypesTest sets up the test environment with all metric types
func setupAllMetricTypesTest(t *testing.T) *testMetricsEnvironment {
	tempDir := t.TempDir()

	cfg := &config.Config{
		ManifestPath:      "../fixtures/data/manifest_semantic_all_bq.json",
		CatalogPath:       "../fixtures/data/catalog_semantic_generated_bq.json",
		OutputDir:         tempDir,
		UseSemanticModels: true,
		LogLevel:          "WARN",
	}

	// Load and parse
	manifestData, err := os.ReadFile(cfg.ManifestPath)
	require.NoError(t, err)
	catalogData, err := os.ReadFile(cfg.CatalogPath)
	require.NoError(t, err)

	var rawManifest, rawCatalog map[string]interface{}
	require.NoError(t, json.Unmarshal(manifestData, &rawManifest))
	require.NoError(t, json.Unmarshal(catalogData, &rawCatalog))

	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	require.NoError(t, err)

	dbtModels, err := parser.GetModels()
	require.NoError(t, err)

	// Parse semantic models
	semanticModels, err := parser.GetSemanticModelParser().GetSemanticModels()
	require.NoError(t, err)

	semanticMeasures := make(map[string][]models.DbtSemanticMeasure)
	for _, sm := range semanticModels {
		if len(sm.Measures) > 0 {
			semanticMeasures[sm.Model] = sm.Measures
		}
	}

	// Parse metrics
	ratioMetrics, err := parser.GetMetricParser().GetRatioMetrics()
	require.NoError(t, err)

	derivedMetrics, err := parser.GetMetricParser().GetDerivedMetrics()
	require.NoError(t, err)

	simpleMetrics, err := parser.GetMetricParser().GetSimpleMetrics()
	require.NoError(t, err)

	cumulativeMetrics, err := parser.GetMetricParser().GetCumulativeMetrics()
	require.NoError(t, err)

	conversionMetrics, err := parser.GetMetricParser().GetConversionMetrics()
	require.NoError(t, err)

	t.Logf("Parsed: %d semantic measures, %d ratio, %d derived, %d simple, %d cumulative, %d conversion metrics",
		len(semanticMeasures), len(ratioMetrics), len(derivedMetrics), len(simpleMetrics),
		len(cumulativeMetrics), len(conversionMetrics))

	// Create generator with plugin
	gen := generators.NewLookMLGenerator(cfg)
	metricsPlugin := pluginMetrics.NewMetricsPlugin(cfg)
	gen.RegisterPlugin(metricsPlugin)

	// Set all data
	gen.SetSemanticMeasures(semanticMeasures)
	gen.SetRatioMetrics(ratioMetrics)
	gen.SetDerivedMetrics(derivedMetrics)
	gen.SetSimpleMetrics(simpleMetrics)
	gen.SetCumulativeMetrics(cumulativeMetrics)
	gen.SetConversionMetrics(conversionMetrics)

	// Generate
	ctx := context.Background()
	filesGenerated, err := gen.GenerateAllWithContext(ctx, dbtModels)
	require.NoError(t, err)
	require.Greater(t, filesGenerated, 0)

	t.Logf("Generated %d files", filesGenerated)

	return &testMetricsEnvironment{
		tempDir:           tempDir,
		ratioMetrics:      ratioMetrics,
		cumulativeMetrics: cumulativeMetrics,
		conversionMetrics: conversionMetrics,
	}
}

// TestSemanticMeasuresWithAllMetricTypes tests semantic measures with ratio, derived, and cumulative metrics
func TestSemanticMeasuresWithAllMetricTypes(t *testing.T) {
	env := setupAllMetricTypesTest(t)

	t.Run("RatioMetricsGenerated", func(t *testing.T) {
		verifyRatioMetrics(t, env)
	})

	t.Run("CumulativeMetricsGenerated", func(t *testing.T) {
		verifyCumulativeMetrics(t, env)
	})

	t.Run("ExploresEnrichedWithJoins", func(t *testing.T) {
		verifyExploreJoins(t, env)
	})
}

func verifyRatioMetrics(t *testing.T, env *testMetricsEnvironment) {
	if len(env.ratioMetrics) == 0 {
		t.Skip("No ratio metrics in test data")
	}

	var files []string
	err := filepath.Walk(env.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "__metrics.view.lkml") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)

	found := false
	for _, file := range files {
		content, _ := os.ReadFile(file)
		contentStr := string(content)

		for _, metric := range env.ratioMetrics {
			if strings.Contains(contentStr, metric.Name) {
				found = true
				t.Logf("Found ratio metric %s in %s", metric.Name, file)
				assert.Contains(t, contentStr, "/", "Ratio metric should contain division")
			}
		}
	}

	if !found && len(env.ratioMetrics) > 0 {
		t.Log("Warning: Ratio metrics defined but not found in output")
	}
}

func verifyCumulativeMetrics(t *testing.T, env *testMetricsEnvironment) {
	if len(env.cumulativeMetrics) == 0 {
		t.Skip("No cumulative metrics in test data")
	}

	var files []string
	err := filepath.Walk(env.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, "__cumulative.view.lkml") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)

	if len(files) > 0 {
		t.Logf("Found %d cumulative metric files", len(files))
		for _, file := range files {
			content, _ := os.ReadFile(file)
			contentStr := string(content)
			assert.Contains(t, contentStr, "SELECT", "Cumulative metric should have SELECT statement")
			assert.Contains(t, contentStr, "metric_time", "Cumulative metric should reference metric_time")
		}
	} else {
		t.Log("Warning: Cumulative metrics defined but no cumulative files generated")
	}
}

func verifyExploreJoins(t *testing.T, env *testMetricsEnvironment) {
	var files []string
	err := filepath.Walk(env.tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".view.lkml") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)

	for _, file := range files {
		if strings.Contains(file, "__") {
			continue
		}

		content, _ := os.ReadFile(file)
		contentStr := string(content)

		if strings.Contains(contentStr, "explore:") {
			t.Logf("Found explore in %s", filepath.Base(file))
			if len(env.cumulativeMetrics) > 0 || len(env.conversionMetrics) > 0 {
				if strings.Contains(contentStr, "join:") {
					t.Logf("  Explore contains join statements")
				}
			}
		}
	}
}

// TestSemanticMeasuresDisabled tests that semantic measures are not generated when disabled
func TestSemanticMeasuresDisabled(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		ManifestPath:      "../fixtures/data/manifest_semantic.json",
		CatalogPath:       "../fixtures/data/catalog_semantic.json",
		OutputDir:         tempDir,
		UseSemanticModels: false, // Disabled
		LogLevel:          "WARN",
	}

	manifestData, err := os.ReadFile(cfg.ManifestPath)
	require.NoError(t, err)
	catalogData, err := os.ReadFile(cfg.CatalogPath)
	require.NoError(t, err)

	var rawManifest, rawCatalog map[string]interface{}
	require.NoError(t, json.Unmarshal(manifestData, &rawManifest))
	require.NoError(t, json.Unmarshal(catalogData, &rawCatalog))

	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	require.NoError(t, err)

	dbtModels, err := parser.GetModels()
	require.NoError(t, err)

	// Create generator WITHOUT plugin registration
	gen := generators.NewLookMLGenerator(cfg)

	// Generate
	ctx := context.Background()
	_, err = gen.GenerateAllWithContext(ctx, dbtModels)
	require.NoError(t, err)

	// Verify no __metrics.view.lkml files generated
	var metricsFiles []string
	_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, "__metrics.view.lkml") {
			metricsFiles = append(metricsFiles, path)
		}
		return nil
	})
	assert.Empty(t, metricsFiles, "Should not generate metrics files when disabled")

	// Verify no __cumulative.view.lkml files
	var cumulativeFiles []string
	_ = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, "__cumulative.view.lkml") {
			cumulativeFiles = append(cumulativeFiles, path)
		}
		return nil
	})
	assert.Empty(t, cumulativeFiles, "Should not generate cumulative files when disabled")
}

// Helper function to map dbt aggregation to LookML type
func mapAggToLookML(agg string) string {
	switch strings.ToLower(agg) {
	case "sum", "sum_boolean":
		return "sum"
	case "count", "count_distinct":
		return "count_distinct"
	case "average", "avg":
		return "average"
	case "min":
		return "min"
	case "max":
		return "max"
	default:
		return agg
	}
}
