package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/parsers"
	"github.com/magnus-ffcg/go-dbt2lookml/tests/integration/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: LookML parsing structures moved to utils package

// TestIntegration contains integration tests for LKML generation
type TestIntegration struct {
	outputDir string
}

// setupTestEnvironment sets up the test environment
func (t *TestIntegration) setupTestEnvironment() error {
	t.outputDir = "output/tests"
	return os.MkdirAll(t.outputDir, 0755)
}

// cleanupTestEnvironment cleans up test artifacts
func (t *TestIntegration) cleanupTestEnvironment() {
	if t.outputDir != "" {
		os.RemoveAll(t.outputDir)
		// Also clean up parent directory if it's empty
		parentDir := filepath.Dir(t.outputDir)
		if parentDir != "." && parentDir != ".." {
			if files, err := os.ReadDir(parentDir); err == nil && len(files) == 0 {
				os.Remove(parentDir)
			}
		}
	}
}

// loadTestData loads manifest and catalog test data
func (t *TestIntegration) loadTestData() (map[string]interface{}, map[string]interface{}, error) {
	// Load manifest
	manifestData, err := os.ReadFile("../fixtures/data/manifest.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return nil, nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Load catalog
	catalogData, err := os.ReadFile("../fixtures/data/catalog.json")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load catalog: %w", err)
	}

	var catalog map[string]interface{}
	if err := json.Unmarshal(catalogData, &catalog); err != nil {
		return nil, nil, fmt.Errorf("failed to parse catalog: %w", err)
	}

	return manifest, catalog, nil
}

// createTestConfig creates a test configuration
func (t *TestIntegration) createTestConfig(selectModel string) *config.Config {
	return &config.Config{
		ManifestPath:    "../fixtures/data/manifest.json",
		CatalogPath:     "../fixtures/data/catalog.json",
		OutputDir:       t.outputDir,
		TargetDir:       "fixtures/data",
		Select:          selectModel,
		UseTableName:    true,
		LogLevel:        "INFO",
		ContinueOnError: false,
	}
}

// generateLookML generates LookML for a specific model
func (ti *TestIntegration) generateLookML(cfg *config.Config) ([]*models.DbtModel, error) {
	// Load test data
	rawManifest, rawCatalog, err := ti.loadTestData()
	if err != nil {
		return nil, err
	}

	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to create parser: %w", err)
	}

	dbtModels, err := parser.GetModels()
	if err != nil {
		return nil, fmt.Errorf("failed to parse models: %w", err)
	}

	// Debug: check if parsed models have ARRAY columns
	for _, model := range dbtModels {
		if strings.Contains(model.Name, "dq_ICASOI_Current") {
			arrayCount := 0
			for colName, col := range model.Columns {
				if col.DataType != nil && strings.HasPrefix(strings.ToUpper(*col.DataType), "ARRAY") {
					arrayCount++
					log.Printf("DEBUG PARSED: Model %s has ARRAY column %s", model.Name, colName)
				}
			}
			log.Printf("DEBUG PARSED: Model %s returned from parser has %d ARRAY columns", model.Name, arrayCount)
		}
	}

	// Filter models if select is specified
	if cfg.Select != "" {
		var filteredModels []*models.DbtModel
		for _, model := range dbtModels {
			if model.Name == cfg.Select {
				filteredModels = append(filteredModels, model)
			}
		}
		dbtModels = filteredModels
	}

	if len(dbtModels) == 0 {
		return nil, fmt.Errorf("no models found matching criteria")
	}

	// Generate LookML
	generator := generators.NewLookMLGenerator(cfg)
	_, err = generator.GenerateAll(dbtModels)
	if err != nil {
		return nil, fmt.Errorf("failed to generate LookML: %w", err)
	}

	return dbtModels, nil
}

// compareWithExpectedOutput compares generated LookML with expected output using utils
func (t *TestIntegration) compareWithExpectedOutput(testT *testing.T, generatedFile, expectedFile string, model *models.DbtModel) {
	comparator := utils.NewLookMLComparator()
	comparator.CompareWithExpected(testT, generatedFile, expectedFile, model)
}

// TestGenerateNestedLookMLWithExplore tests LKML generation with explore functionality
func TestGenerateNestedLookMLWithExplore(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Create test configuration
	cfg := testIntegration.createTestConfig("conlaybi_item_dataquality__dq_ICASOI_Current")

	// Generate LookML
	models, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, models, "No models generated")

	// Check that output files were created
	expectedOutputPath := filepath.Join(testIntegration.outputDir, "conlaybi", "item_dataquality", "dq_icasoi_current.view.lkml")
	assert.FileExists(t, expectedOutputPath, "Generated LookML file should exist")

	// Load expected fixture (if it exists)
	expectedFixturePath := "fixtures/expected/dq_icasoi_current.view.lkml"
	if _, err := os.Stat(expectedFixturePath); err == nil {
		// Compare using new utilities
		testIntegration.compareWithExpectedOutput(t, expectedOutputPath, expectedFixturePath, models[0])
	} else {
		t.Logf("Expected fixture not found at %s, skipping comparison", expectedFixturePath)
	}
}

// TestGenerateSalesWasteLookMLWithExplore tests generating LKML for sales waste model
func TestGenerateSalesWasteLookMLWithExplore(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Create test configuration
	cfg := testIntegration.createTestConfig("conlaybi_consumer_sales_secure_versioned__f_store_sales_waste_day")

	// Generate LookML
	models, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, models, "No models generated")

	// Check that output files were created
	expectedOutputPath := filepath.Join(testIntegration.outputDir, "conlaybi", "consumer_sales_secure_versioned", "f_store_sales_waste_day_v1.view.lkml")
	assert.FileExists(t, expectedOutputPath, "Generated LookML file should exist")

	// Load expected fixture (if it exists)
	expectedFixturePath := "fixtures/expected/f_store_sales_waste_day_v1.view.lkml"
	if _, err := os.Stat(expectedFixturePath); err == nil {
		// Compare using new utilities
		testIntegration.compareWithExpectedOutput(t, expectedOutputPath, expectedFixturePath, models[0])
	} else {
		t.Logf("Expected fixture not found at %s, skipping comparison", expectedFixturePath)
	}
}

// TestGenerateDItemV3ComplexStructsWithExplore tests generating LKML for d_item_v3 model
func TestGenerateDItemV3ComplexStructsWithExplore(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Create test configuration
	cfg := testIntegration.createTestConfig("conlaybi_item_versioned__d_item")
	cfg.OutputDir = "output/test_d_item_v3"

	// Generate LookML
	models, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, models, "No models generated")

	// Check that output files were created
	expectedOutputPath := filepath.Join("output/test_d_item_v3", "conlaybi", "item_versioned", "d_item_v3.view.lkml")
	assert.FileExists(t, expectedOutputPath, "Generated LookML file should exist")

	// Load expected fixture (if it exists)
	expectedFixturePath := "fixtures/expected/d_item_v3.view.lkml"
	if _, err := os.Stat(expectedFixturePath); err == nil {
		// Compare using new utilities
		testIntegration.compareWithExpectedOutput(t, expectedOutputPath, expectedFixturePath, models[0])
	} else {
		t.Logf("Expected fixture not found at %s, skipping comparison", expectedFixturePath)
	}
}

// TestSingleFileConsolidation tests that each dbt model generates exactly one view file with all content inline
func TestSingleFileConsolidation(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Test configuration - generate all models
	cfg := testIntegration.createTestConfig("")

	// Generate LookML
	models, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, models, "Should have generated models")

	// Verify single file consolidation for each model
	for _, model := range models {
		t.Run(fmt.Sprintf("Model_%s", model.Name), func(t *testing.T) {
			testIntegration.verifySingleFileConsolidation(t, model, cfg.OutputDir)
		})
	}
}

// verifySingleFileConsolidation verifies that a model generates exactly one consolidated file
func (ti *TestIntegration) verifySingleFileConsolidation(t *testing.T, model *models.DbtModel, outputDir string) {
	// 1. Verify exactly one .view.lkml file exists for this model
	modelFiles := ti.findModelFiles(model, outputDir)

	// Should have exactly one .view.lkml file
	viewFiles := []string{}
	exploreFiles := []string{}
	for _, file := range modelFiles {
		if strings.HasSuffix(file, ".view.lkml") {
			viewFiles = append(viewFiles, file)
		} else if strings.HasSuffix(file, ".explore.lkml") {
			exploreFiles = append(exploreFiles, file)
		}
	}

	assert.Len(t, viewFiles, 1, "Should have exactly one .view.lkml file for model %s, found: %v", model.Name, viewFiles)
	assert.Len(t, exploreFiles, 0, "Should have no separate .explore.lkml files for model %s, found: %v", model.Name, exploreFiles)

	if len(viewFiles) != 1 {
		return // Skip further checks if basic requirement not met
	}

	// 2. Verify the single file contains all required sections
	mainViewFile := viewFiles[0]
	content, err := os.ReadFile(mainViewFile)
	require.NoError(t, err, "Should be able to read main view file")

	contentStr := string(content)

	// 3. Verify explore section exists (if model has ARRAY columns)
	hasArrayColumns := ti.modelHasArrayColumns(model)
	if hasArrayColumns {
		assert.Contains(t, contentStr, "explore:", "File should contain explore section for model with ARRAY columns")
		assert.Contains(t, contentStr, "join:", "File should contain join statements for nested views")
	}

	// 4. Verify main view exists
	// The view name should match the filename (without .view.lkml extension)
	fileName := filepath.Base(mainViewFile)
	// Remove .view.lkml extension properly
	expectedMainViewName := strings.TrimSuffix(fileName, ".view.lkml")
	searchPattern := fmt.Sprintf("view: %s {", expectedMainViewName)

	assert.Contains(t, contentStr, searchPattern, "File should contain main view definition")

	// 5. Verify nested views are inline (if model has ARRAY columns)
	if hasArrayColumns {
		nestedViewCount := strings.Count(contentStr, "view:")
		assert.GreaterOrEqual(t, nestedViewCount, 2, "File should contain main view plus nested views inline")

		// Verify no separate nested view files exist
		nestedFiles := ti.findNestedViewFiles(model, outputDir)
		assert.Len(t, nestedFiles, 0, "Should have no separate nested view files for model %s, found: %v", model.Name, nestedFiles)
	}

	// 6. Verify file structure integrity
	ti.verifyFileStructureIntegrity(t, contentStr, model)

	// 7. Compare with expected output (semantic comparison)
	expectedFileName := ti.getExpectedFileName(model)
	expectedFilePath := filepath.Join("fixtures/expected", expectedFileName)
	if _, err := os.Stat(expectedFilePath); err == nil {
		ti.compareWithExpectedOutput(t, mainViewFile, expectedFilePath, model)
	} else {
		t.Logf("Expected fixture not found at %s, skipping comparison", expectedFilePath)
	}
}

// findModelFiles finds all files related to a specific model
func (ti *TestIntegration) findModelFiles(model *models.DbtModel, outputDir string) []string {
	var files []string

	// Walk through output directory to find files related to this model
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(path, ".view.lkml") || strings.HasSuffix(path, ".explore.lkml")) {
			// Check if file is related to this model
			fileName := filepath.Base(path)
			fileNameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			fileNameWithoutExt = strings.TrimSuffix(fileNameWithoutExt, ".view")
			fileNameWithoutExt = strings.TrimSuffix(fileNameWithoutExt, ".explore")

			// Extract simplified name (e.g., "dq_icasoi_current" from "conlaybi_item_dataquality__dq_ICASOI_Current")
			simplifiedName := ti.getSimplifiedModelName(model.Name)

			// Match if the filename exactly matches or starts with simplified name + underscore (for versions)
			if fileNameWithoutExt == simplifiedName || strings.HasPrefix(fileNameWithoutExt, simplifiedName+"_") {
				files = append(files, path)
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking directory: %v", err)
	}

	return files
}

// getSimplifiedModelName extracts the simplified name from a full model name
func (ti *TestIntegration) getSimplifiedModelName(fullName string) string {
	// For names like "conlaybi_item_dataquality__dq_ICASOI_Current", extract "dq_ICASOI_Current"
	parts := strings.Split(fullName, "__")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return strings.ToLower(fullName)
}

// isModelRelatedPath checks if a file path is related to a model based on directory structure
func (ti *TestIntegration) isModelRelatedPath(filePath string, model *models.DbtModel) bool {
	// Check if the path contains schema-related directories
	// e.g., "output/tests/conlaybi/item_dataquality/dq_icasoi_current.view.lkml"

	modelNameLower := strings.ToLower(model.Name)
	pathLower := strings.ToLower(filePath)

	// Extract schema parts from model name
	if strings.Contains(modelNameLower, "item_dataquality") && strings.Contains(pathLower, "item_dataquality") {
		return true
	}
	if strings.Contains(modelNameLower, "consumer_sales") && strings.Contains(pathLower, "consumer_sales") {
		return true
	}
	if strings.Contains(modelNameLower, "item_versioned") && strings.Contains(pathLower, "item_versioned") {
		return true
	}

	return false
}

// findNestedViewFiles finds separate nested view files (should be empty)
func (ti *TestIntegration) findNestedViewFiles(model *models.DbtModel, outputDir string) []string {
	var nestedFiles []string

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".view.lkml") {
			fileName := filepath.Base(path)
			// Check if this is a nested view file (contains __ pattern but is not the main view)
			if strings.Contains(fileName, "__") && strings.Contains(fileName, model.Name) {
				// Check if it's the main view file or a separate nested view file
				expectedMainFile := ti.getExpectedViewFileName(model)
				if fileName != expectedMainFile {
					nestedFiles = append(nestedFiles, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Error walking directory for nested files: %v", err)
	}

	return nestedFiles
}

// modelHasArrayColumns checks if a model has ARRAY columns
func (ti *TestIntegration) modelHasArrayColumns(model *models.DbtModel) bool {
	for _, column := range model.Columns {
		if column.DataType != nil && strings.HasPrefix(strings.ToUpper(*column.DataType), "ARRAY") {
			return true
		}
	}
	return false
}

// getExpectedViewName gets the expected view name for a model
func (ti *TestIntegration) getExpectedViewName(model *models.DbtModel) string {
	// The generator uses simplified names, not full model names
	// e.g., "dq_icasoi_current" instead of "conlaybi_item_dataquality__dq_ICASOI_Current"
	return ti.getSimplifiedModelName(model.Name)
}

// getExpectedViewFileName gets the expected view file name for a model
func (ti *TestIntegration) getExpectedViewFileName(model *models.DbtModel) string {
	return fmt.Sprintf("%s.view.lkml", model.Name)
}

// verifyFileStructureIntegrity verifies the internal structure of the consolidated file
func (ti *TestIntegration) verifyFileStructureIntegrity(t *testing.T, content string, model *models.DbtModel) {
	// Verify proper LookML syntax
	// This is a basic check - could be expanded with more sophisticated parsing

	// Check for balanced braces
	openBraces := strings.Count(content, "{")
	closeBraces := strings.Count(content, "}")
	assert.Equal(t, openBraces, closeBraces, "File should have balanced braces")

	// Check for required LookML keywords
	assert.Contains(t, content, "sql_table_name:", "File should contain sql_table_name")

	// Verify proper dimension syntax
	dimensionCount := strings.Count(content, "dimension:")
	if dimensionCount > 0 {
		// Should have proper SQL references
		assert.Contains(t, content, "sql:", "Should contain SQL references for dimensions")
	}
}

// getExpectedFileName determines the expected file name for a model
func (ti *TestIntegration) getExpectedFileName(model *models.DbtModel) string {
	// Extract table name from RelationName for expected file matching
	parts := strings.Split(model.RelationName, ".")
	if len(parts) > 0 {
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		return fmt.Sprintf("%s.view.lkml", strings.ToLower(tableName))
	}
	return fmt.Sprintf("%s.view.lkml", strings.ToLower(model.Name))
}

// TestBasicLookMLGeneration tests basic LKML generation functionality
func TestBasicLookMLGeneration(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Create test configuration for any available model
	cfg := testIntegration.createTestConfig("") // No specific model selection

	// Generate LookML
	models, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, models, "No models generated")

	// Check that at least one output file was created
	var outputFiles []string
	err = filepath.Walk(testIntegration.outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".lkml") {
			outputFiles = append(outputFiles, path)
		}
		return nil
	})
	require.NoError(t, err)
	assert.NotEmpty(t, outputFiles, "At least one LookML file should be generated")

	// Verify that generated files are valid (not empty and contain basic LookML structure)
	for _, file := range outputFiles {
		content, err := os.ReadFile(file)
		require.NoError(t, err, "Should be able to read generated file %s", file)

		contentStr := string(content)
		assert.NotEmpty(t, contentStr, "Generated file %s should not be empty", file)

		// Basic LookML structure validation
		if strings.HasSuffix(file, ".view.lkml") {
			assert.Contains(t, contentStr, "view:", "View file should contain 'view:' declaration")
			assert.Contains(t, contentStr, "sql_table_name:", "View file should contain 'sql_table_name:' declaration")
		} else if strings.HasSuffix(file, ".explore.lkml") {
			assert.Contains(t, contentStr, "explore:", "Explore file should contain 'explore:' declaration")
		}
	}
}

// TestFixtureComparison compares generated LookML with expected fixtures
func TestFixtureComparison(t *testing.T) {
	testIntegration := &TestIntegration{}

	// Setup test environment
	err := testIntegration.setupTestEnvironment()
	require.NoError(t, err)
	defer testIntegration.cleanupTestEnvironment()

	// Generate all models with UseTableName=true to match fixtures
	cfg := testIntegration.createTestConfig("")
	cfg.UseTableName = true

	dbtModels, err := testIntegration.generateLookML(cfg)
	require.NoError(t, err)
	require.NotEmpty(t, dbtModels, "Should have generated models")

	// Define fixture mappings: model name -> expected fixture file
	fixtureMap := map[string]string{
		"conlaybi_item_dataquality__dq_ICASOI_Current":                      "dq_icasoi_current.view.lkml",
		"conlaybi_item_dataquality__dq_ItemEBO_Current":                     "dq_item_ebo_current.view.lkml",
		"conlaybi_item_versioned__d_item":                                   "d_item_v3.view.lkml",
		"conlaybi_consumer_sales_secure_versioned__f_store_sales_waste_day": "f_store_sales_waste_day_v1.view.lkml",
		"conlaybi_consumer_sales_looker__f_store_sales_day_selling_entity":  "f_store_sales_day_selling_entity_v1.view.lkml",
	}

	// Test each fixture
	for modelName, fixtureName := range fixtureMap {
		t.Run(fixtureName, func(t *testing.T) {
			var model *models.DbtModel
			for _, m := range dbtModels {
				if m.Name == modelName {
					model = m
					break
				}
			}
			if model == nil {
				t.Skipf("Model %s not found in generated models", modelName)
				return
			}

			// Find generated file
			generatedFile := testIntegration.findGeneratedFileForModel(t, model, cfg.OutputDir)
			if generatedFile == "" {
				t.Fatalf("Could not find generated file for model %s", modelName)
			}

			// Expected fixture path
			expectedFile := filepath.Join("..", "fixtures", "expected", fixtureName)
			if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
				t.Skipf("Expected fixture file not found: %s", expectedFile)
				return
			}

			// Read both files
			generatedContent, err := os.ReadFile(generatedFile)
			require.NoError(t, err, "Should read generated file")

			expectedContent, err := os.ReadFile(expectedFile)
			require.NoError(t, err, "Should read expected file")

			generatedStr := string(generatedContent)
			expectedStr := string(expectedContent)

			t.Logf("  Generated: %d bytes, Expected: %d bytes", len(generatedContent), len(expectedContent))

			// Count views
			genViews := countOccurrences(generatedStr, "\nview:")
			expViews := countOccurrences(expectedStr, "\nview:")

			t.Logf("  Views: %d (expected %d)", genViews, expViews)
			assert.Equal(t, expViews, genViews, "View count should match")

			// Count dimensions per view for more precise comparison
			genViewDims := countDimensionsPerView(generatedStr)
			expViewDims := countDimensionsPerView(expectedStr)

			// Log main view dimensions (first view)
			if len(genViewDims) > 0 && len(expViewDims) > 0 {
				t.Logf("  Main view dimensions: %d (expected %d)", genViewDims[0].Dimensions, expViewDims[0].Dimensions)
				t.Logf("  Main view dimension groups: %d (expected %d)", genViewDims[0].DimensionGroups, expViewDims[0].DimensionGroups)
				t.Logf("  Main view measures: %d (expected %d)", genViewDims[0].Measures, expViewDims[0].Measures)

				// Assert main view counts match
				assert.Equal(t, expViewDims[0].Dimensions, genViewDims[0].Dimensions, "Main view dimension count should match")
				assert.Equal(t, expViewDims[0].DimensionGroups, genViewDims[0].DimensionGroups, "Main view dimension group count should match")
				assert.Equal(t, expViewDims[0].Measures, genViewDims[0].Measures, "Main view measure count should match")
			}

			// Quality checks
			t.Run("NoDuplicateDimensions", func(t *testing.T) {
				duplicates := findDuplicateDimensions(generatedStr)
				assert.Empty(t, duplicates, "Should have no duplicate dimension names: %v", duplicates)
			})
			t.Run("NoBackticksInSQL", func(t *testing.T) {
				backtickCount := countOccurrences(generatedStr, "sql: ${TABLE}.`")
				assert.Equal(t, 0, backtickCount, "Should not use backticks in SQL references")
			})

			t.Run("NoPlaceholderDescriptions", func(t *testing.T) {
				placeholderCount := countOccurrences(generatedStr, "This field is missing a description")
				assert.Equal(t, 0, placeholderCount, "Should not have placeholder descriptions")
			})

			t.Run("ProperSQLReferences", func(t *testing.T) {
				// Check that we have PascalCase SQL references for nested columns
				// Only icasoi models have Classification fields
				if strings.Contains(fixtureName, "icasoi") {
					assert.Contains(t, generatedStr, "${TABLE}.Classification.", "Should have PascalCase SQL references")
				}
			})
		})
	}
}

// Helper functions for fixture comparison
func countOccurrences(s, substr string) int {
	return strings.Count(s, substr)
}

type ViewCounts struct {
	ViewName        string
	Dimensions      int
	DimensionGroups int
	Measures        int
}

func countDimensionsPerView(content string) []ViewCounts {
	var views []ViewCounts
	var currentView *ViewCounts

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track view boundaries
		if strings.HasPrefix(trimmed, "view:") {
			// Save previous view if exists
			if currentView != nil {
				views = append(views, *currentView)
			}

			// Start new view
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				viewName := strings.TrimSuffix(parts[1], " {")
				currentView = &ViewCounts{
					ViewName: viewName,
				}
			}
		}

		// Count elements in current view
		if currentView != nil {
			if strings.HasPrefix(trimmed, "dimension:") {
				currentView.Dimensions++
			} else if strings.HasPrefix(trimmed, "dimension_group:") {
				currentView.DimensionGroups++
			} else if strings.HasPrefix(trimmed, "measure:") {
				currentView.Measures++
			}
		}
	}

	// Save last view
	if currentView != nil {
		views = append(views, *currentView)
	}

	return views
}

func findDuplicateDimensions(content string) []string {
	// Track dimensions per view to detect duplicates within the same view
	type viewDimensions struct {
		viewName   string
		dimensions map[string]int
	}

	var views []viewDimensions
	var currentView *viewDimensions

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track view boundaries
		if strings.HasPrefix(trimmed, "view:") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				viewName := strings.TrimSuffix(parts[1], " {")
				currentView = &viewDimensions{
					viewName:   viewName,
					dimensions: make(map[string]int),
				}
				views = append(views, *currentView)
			}
		}

		// Track dimensions within current view
		if currentView != nil && (strings.HasPrefix(trimmed, "dimension:") || strings.HasPrefix(trimmed, "dimension_group:")) {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				name := strings.TrimSuffix(parts[1], " {")
				currentView.dimensions[name]++
			}
		}
	}

	// Find duplicates within each view
	var duplicates []string
	for _, view := range views {
		for name, count := range view.dimensions {
			if count > 1 {
				duplicates = append(duplicates, fmt.Sprintf("%s in %s (x%d)", name, view.viewName, count))
			}
		}
	}
	return duplicates
}

func (ti *TestIntegration) findGeneratedFileForModel(t *testing.T, model *models.DbtModel, outputDir string) string {
	var foundFiles []string

	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".view.lkml") {
			foundFiles = append(foundFiles, path)
		}
		return nil
	})

	require.NoError(t, err)

	// Try to match by table name from RelationName
	if model.RelationName != "" {
		parts := strings.Split(model.RelationName, ".")
		if len(parts) > 0 {
			tableName := strings.Trim(parts[len(parts)-1], "`")
			tableName = strings.ToLower(tableName)

			for _, file := range foundFiles {
				if strings.Contains(strings.ToLower(file), tableName) {
					return file
				}
			}
		}
	}

	// Fallback: return first file if only one exists
	if len(foundFiles) == 1 {
		return foundFiles[0]
	}

	return ""
}
