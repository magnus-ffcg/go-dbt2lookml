package utils

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LookMLComparator handles semantic comparison of LookML files
type LookMLComparator struct {
	parser *LookMLParser
}

// NewLookMLComparator creates a new comparator instance
func NewLookMLComparator() *LookMLComparator {
	return &LookMLComparator{
		parser: NewLookMLParser(),
	}
}

// CompareWithExpected compares generated LookML with expected output
func (c *LookMLComparator) CompareWithExpected(t *testing.T, generatedFile, expectedFile string, model *models.DbtModel) {
	// Parse both files
	generated, err := c.parser.ParseFile(generatedFile)
	require.NoError(t, err, "Should parse generated file: %s", generatedFile)

	expected, err := c.parser.ParseFile(expectedFile)
	require.NoError(t, err, "Should parse expected file: %s", expectedFile)

	// Compare explores
	c.compareExplores(t, generated.Explores, expected.Explores, model)

	// Compare views
	c.compareViews(t, generated.Views, expected.Views, model)
}

// compareExplores compares explore blocks
func (c *LookMLComparator) compareExplores(t *testing.T, generated, expected []ParsedExplore, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of explores for model %s", model.Name)

	if len(generated) == 0 && len(expected) == 0 {
		return
	}

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedExplore)
	expMap := make(map[string]ParsedExplore)

	for _, e := range generated {
		genMap[e.Name] = e
	}
	for _, e := range expected {
		expMap[e.Name] = e
	}

	// Compare each explore
	for name, expectedExplore := range expMap {
		generatedExplore, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain explore: %s", name)

		if exists {
			c.compareExplore(t, generatedExplore, expectedExplore, model)
		}
	}

	// Check for extra explores in generated
	for name := range genMap {
		_, exists := expMap[name]
		assert.True(t, exists, "Generated file contains unexpected explore: %s", name)
	}
}

// compareExplore compares a single explore
func (c *LookMLComparator) compareExplore(t *testing.T, generated, expected ParsedExplore, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "Explore names should match")

	// Compare hidden property
	if expected.Hidden != nil {
		assert.Equal(t, *expected.Hidden, getBoolValue(generated.Hidden), "Explore hidden property should match for %s", expected.Name)
	}

	// Compare joins
	c.compareJoins(t, generated.Joins, expected.Joins, model)
}

// compareJoins compares join blocks
func (c *LookMLComparator) compareJoins(t *testing.T, generated, expected []ParsedJoin, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of joins for model %s", model.Name)

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedJoin)
	expMap := make(map[string]ParsedJoin)

	for _, j := range generated {
		genMap[j.Name] = j
	}
	for _, j := range expected {
		expMap[j.Name] = j
	}

	// Compare each join
	for name, expectedJoin := range expMap {
		generatedJoin, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain join: %s", name)

		if exists {
			c.compareJoin(t, generatedJoin, expectedJoin, model)
		}
	}
}

// compareJoin compares a single join
func (c *LookMLComparator) compareJoin(t *testing.T, generated, expected ParsedJoin, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "Join names should match")
	assert.Equal(t, expected.Relationship, generated.Relationship, "Join relationship should match for %s", expected.Name)

	// Compare view labels (optional)
	if expected.ViewLabel != nil {
		assert.Equal(t, *expected.ViewLabel, getStringValue(generated.ViewLabel), "Join view_label should match for %s", expected.Name)
	}

	// Compare SQL (normalize whitespace)
	assert.Equal(t, normalizeSQL(expected.SQL), normalizeSQL(generated.SQL), "Join SQL should match for %s", expected.Name)
}

// compareViews compares view blocks
func (c *LookMLComparator) compareViews(t *testing.T, generated, expected []ParsedView, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of views for model %s", model.Name)

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedView)
	expMap := make(map[string]ParsedView)

	for _, v := range generated {
		genMap[v.Name] = v
	}
	for _, v := range expected {
		expMap[v.Name] = v
	}

	// Compare each view
	for name, expectedView := range expMap {
		generatedView, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain view: %s", name)

		if exists {
			c.compareView(t, generatedView, expectedView, model)
		}
	}

	// Check for extra views in generated
	for name := range genMap {
		_, exists := expMap[name]
		assert.True(t, exists, "Generated file contains unexpected view: %s", name)
	}
}

// compareView compares a single view
func (c *LookMLComparator) compareView(t *testing.T, generated, expected ParsedView, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "View names should match")
	assert.Equal(t, expected.SQLTableName, generated.SQLTableName, "SQL table names should match for view %s", expected.Name)

	// Compare optional properties
	if expected.Label != nil {
		assert.Equal(t, *expected.Label, getStringValue(generated.Label), "View label should match for %s", expected.Name)
	}
	if expected.Description != nil {
		assert.Equal(t, *expected.Description, getStringValue(generated.Description), "View description should match for %s", expected.Name)
	}

	// Compare dimensions (order-independent)
	c.compareDimensions(t, generated.Dimensions, expected.Dimensions, model)

	// Compare dimension groups (order-independent)
	c.compareDimensionGroups(t, generated.DimensionGroups, expected.DimensionGroups, model)

	// Compare measures (order-independent)
	c.compareMeasures(t, generated.Measures, expected.Measures, model)
}

// compareDimensions compares dimension blocks
func (c *LookMLComparator) compareDimensions(t *testing.T, generated, expected []ParsedDimension, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of dimensions for model %s", model.Name)

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedDimension)
	expMap := make(map[string]ParsedDimension)

	for _, d := range generated {
		genMap[d.Name] = d
	}
	for _, d := range expected {
		expMap[d.Name] = d
	}

	// Compare each dimension
	for name, expectedDim := range expMap {
		generatedDim, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain dimension: %s", name)

		if exists {
			c.compareDimension(t, generatedDim, expectedDim, model)
		}
	}

	// Check for extra dimensions in generated
	for name := range genMap {
		_, exists := expMap[name]
		if !exists {
			t.Logf("WARNING: Generated file contains extra dimension: %s", name)
		}
	}
}

// compareDimension compares a single dimension
func (c *LookMLComparator) compareDimension(t *testing.T, generated, expected ParsedDimension, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "Dimension names should match")
	assert.Equal(t, expected.Type, generated.Type, "Dimension type should match for %s", expected.Name)
	assert.Equal(t, normalizeSQL(expected.SQL), normalizeSQL(generated.SQL), "Dimension SQL should match for %s", expected.Name)

	// Compare optional properties
	if expected.Description != nil {
		assert.Equal(t, *expected.Description, getStringValue(generated.Description), "Dimension description should match for %s", expected.Name)
	}
	if expected.Hidden != nil {
		assert.Equal(t, *expected.Hidden, getBoolValue(generated.Hidden), "Dimension hidden property should match for %s", expected.Name)
	}
	if expected.Label != nil {
		assert.Equal(t, *expected.Label, getStringValue(generated.Label), "Dimension label should match for %s", expected.Name)
	}
}

// compareDimensionGroups compares dimension_group blocks
func (c *LookMLComparator) compareDimensionGroups(t *testing.T, generated, expected []ParsedDimensionGroup, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of dimension_groups for model %s", model.Name)

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedDimensionGroup)
	expMap := make(map[string]ParsedDimensionGroup)

	for _, dg := range generated {
		genMap[dg.Name] = dg
	}
	for _, dg := range expected {
		expMap[dg.Name] = dg
	}

	// Compare each dimension_group
	for name, expectedDG := range expMap {
		generatedDG, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain dimension_group: %s", name)

		if exists {
			c.compareDimensionGroup(t, generatedDG, expectedDG, model)
		}
	}
}

// compareDimensionGroup compares a single dimension_group
func (c *LookMLComparator) compareDimensionGroup(t *testing.T, generated, expected ParsedDimensionGroup, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "Dimension group names should match")
	assert.Equal(t, expected.Type, generated.Type, "Dimension group type should match for %s", expected.Name)
	assert.Equal(t, normalizeSQL(expected.SQL), normalizeSQL(generated.SQL), "Dimension group SQL should match for %s", expected.Name)

	// Compare timeframes (order-independent)
	if len(expected.Timeframes) > 0 {
		sort.Strings(expected.Timeframes)
		sort.Strings(generated.Timeframes)
		assert.Equal(t, expected.Timeframes, generated.Timeframes, "Dimension group timeframes should match for %s", expected.Name)
	}

	// Compare optional properties
	if expected.ConvertTZ != nil {
		assert.Equal(t, *expected.ConvertTZ, getBoolValue(generated.ConvertTZ), "Dimension group convert_tz should match for %s", expected.Name)
	}
	if expected.Datatype != nil {
		assert.Equal(t, *expected.Datatype, getStringValue(generated.Datatype), "Dimension group datatype should match for %s", expected.Name)
	}
}

// compareMeasures compares measure blocks
func (c *LookMLComparator) compareMeasures(t *testing.T, generated, expected []ParsedMeasure, model *models.DbtModel) {
	assert.Len(t, generated, len(expected), "Should have same number of measures for model %s", model.Name)

	// Create maps for easier comparison (order-independent)
	genMap := make(map[string]ParsedMeasure)
	expMap := make(map[string]ParsedMeasure)

	for _, m := range generated {
		genMap[m.Name] = m
	}
	for _, m := range expected {
		expMap[m.Name] = m
	}

	// Compare each measure
	for name, expectedMeasure := range expMap {
		generatedMeasure, exists := genMap[name]
		assert.True(t, exists, "Generated file should contain measure: %s", name)

		if exists {
			c.compareMeasure(t, generatedMeasure, expectedMeasure, model)
		}
	}
}

// compareMeasure compares a single measure
func (c *LookMLComparator) compareMeasure(t *testing.T, generated, expected ParsedMeasure, model *models.DbtModel) {
	assert.Equal(t, expected.Name, generated.Name, "Measure names should match")
	assert.Equal(t, expected.Type, generated.Type, "Measure type should match for %s", expected.Name)
	assert.Equal(t, expected.Label, generated.Label, "Measure label should match for %s", expected.Name)
}

// Helper functions

// getStringValue safely gets a string value from a pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getBoolValue safely gets a bool value from a pointer
func getBoolValue(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

// normalizeSQL normalizes SQL strings for comparison
func normalizeSQL(sql string) string {
	// Remove extra whitespace and normalize
	sql = strings.TrimSpace(sql)
	// Could add more normalization rules here if needed
	return sql
}

// getExpectedFileName determines the expected file name for a model
func getExpectedFileName(model *models.DbtModel) string {
	// This should match the naming convention used in the expected files
	// For now, assume it matches the simplified table name
	parts := strings.Split(model.RelationName, ".")
	if len(parts) > 0 {
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		return fmt.Sprintf("%s.view.lkml", strings.ToLower(tableName))
	}
	return fmt.Sprintf("%s.view.lkml", model.Name)
}
