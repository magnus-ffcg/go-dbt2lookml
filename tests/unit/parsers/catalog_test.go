package parsers

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/dbt2lookml/pkg/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCatalogParser_NewCatalogParser tests catalog parser creation
func TestCatalogParser_NewCatalogParser(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{},
	}
	rawCatalog := map[string]interface{}{}

	parser := parsers.NewCatalogParser(catalog, rawCatalog)
	assert.NotNil(t, parser)
}

// TestCatalogParser_ProcessModelColumns tests column processing with catalog data
func TestCatalogParser_ProcessModelColumns(t *testing.T) {
	tests := []struct {
		name           string
		model          *models.DbtModel
		catalog        *models.DbtCatalog
		expectedCols   int
		checkDataTypes bool
	}{
		{
			name: "model with catalog columns",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name:     "test_model",
					UniqueID: "model.test.test_model",
				},
				Columns: map[string]models.DbtModelColumn{},
			},
			catalog: &models.DbtCatalog{
				Nodes: map[string]models.DbtCatalogNode{
					"model.test.test_model": {
						Columns: map[string]models.DbtCatalogNodeColumn{
							"id": {
								Name:         "id",
								OriginalName: "ID",
								Type:         "INT64",
								DataType:     "INT64",
							},
							"name": {
								Name:         "name",
								OriginalName: "Name",
								Type:         "STRING",
								DataType:     "STRING",
							},
						},
					},
				},
			},
			expectedCols:   2,
			checkDataTypes: true,
		},
		{
			name: "model without catalog entry",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name:     "missing_model",
					UniqueID: "model.test.missing_model",
				},
				Columns: map[string]models.DbtModelColumn{
					"existing": {
						Name: "existing",
					},
				},
			},
			catalog: &models.DbtCatalog{
				Nodes: map[string]models.DbtCatalogNode{},
			},
			expectedCols:   1,
			checkDataTypes: false,
		},
		{
			name: "model with ARRAY columns",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name:     "array_model",
					UniqueID: "model.test.array_model",
				},
				Columns: map[string]models.DbtModelColumn{},
			},
			catalog: &models.DbtCatalog{
				Nodes: map[string]models.DbtCatalogNode{
					"model.test.array_model": {
						Columns: map[string]models.DbtCatalogNodeColumn{
							"tags": {
								Name:         "tags",
								OriginalName: "Tags",
								Type:         "ARRAY<STRING>",
								DataType:     "ARRAY<STRING>",
							},
							"metadata": {
								Name:         "metadata",
								OriginalName: "Metadata",
								Type:         "ARRAY<STRUCT<key STRING, value STRING>>",
								DataType:     "ARRAY",
								InnerTypes:   []string{"STRUCT", "STRING", "STRING"},
							},
						},
					},
				},
			},
			expectedCols:   2,
			checkDataTypes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := parsers.NewCatalogParser(tt.catalog, map[string]interface{}{})
			
			processedModel, err := parser.ProcessModelColumns(tt.model)
			require.NoError(t, err)
			require.NotNil(t, processedModel)
			
			assert.Len(t, processedModel.Columns, tt.expectedCols)
			
			if tt.checkDataTypes {
				for _, col := range processedModel.Columns {
					assert.NotNil(t, col.DataType, "Column %s should have DataType", col.Name)
				}
			}
		})
	}
}

// TestCatalogParser_GetCatalogColumn tests getting a specific column
func TestCatalogParser_GetCatalogColumn(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"id": {
						Name:     "id",
						Type:     "INT64",
						DataType: "INT64",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	// Test existing column
	col, found := parser.GetCatalogColumn("model.test.test_model", "id")
	assert.True(t, found)
	assert.NotNil(t, col)
	assert.Equal(t, "INT64", col.Type)

	// Test non-existent column
	col, found = parser.GetCatalogColumn("model.test.test_model", "missing")
	assert.False(t, found)
	assert.Nil(t, col)

	// Test non-existent model
	col, found = parser.GetCatalogColumn("model.test.missing", "id")
	assert.False(t, found)
	assert.Nil(t, col)
}

// TestCatalogParser_IsArrayType tests ARRAY type detection
func TestCatalogParser_IsArrayType(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"tags": {
						Name: "tags",
						Type: "ARRAY<STRING>",
					},
					"id": {
						Name: "id",
						Type: "INT64",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	assert.True(t, parser.IsArrayType("model.test.test_model", "tags"))
	assert.False(t, parser.IsArrayType("model.test.test_model", "id"))
	assert.False(t, parser.IsArrayType("model.test.test_model", "missing"))
}

// TestCatalogParser_IsStructType tests STRUCT type detection
func TestCatalogParser_IsStructType(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"address": {
						Name: "address",
						Type: "STRUCT<street STRING, city STRING>",
					},
					"id": {
						Name: "id",
						Type: "INT64",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	assert.True(t, parser.IsStructType("model.test.test_model", "address"))
	assert.False(t, parser.IsStructType("model.test.test_model", "id"))
	assert.False(t, parser.IsStructType("model.test.test_model", "missing"))
}

// TestCatalogParser_GetNestedColumns tests getting nested columns
func TestCatalogParser_GetNestedColumns(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"address": {
						Name: "address",
						Type: "STRUCT<street STRING, city STRING>",
					},
					"address.street": {
						Name: "address.street",
						Type: "STRING",
					},
					"address.city": {
						Name: "address.city",
						Type: "STRING",
					},
					"id": {
						Name: "id",
						Type: "INT64",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	nestedCols := parser.GetNestedColumns("model.test.test_model", "address")
	assert.Len(t, nestedCols, 2)

	// Check that we got the right nested columns
	foundStreet := false
	foundCity := false
	for _, col := range nestedCols {
		if col.Name == "address.street" {
			foundStreet = true
		}
		if col.Name == "address.city" {
			foundCity = true
		}
	}
	assert.True(t, foundStreet)
	assert.True(t, foundCity)
}

// TestCatalogParser_GetColumnType tests getting column types
func TestCatalogParser_GetColumnType(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"id": {
						Name:     "id",
						Type:     "INT64",
						DataType: "INT64",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	colType, found := parser.GetColumnType("model.test.test_model", "id")
	assert.True(t, found)
	assert.Equal(t, "INT64", colType)

	colType, found = parser.GetColumnType("model.test.test_model", "missing")
	assert.False(t, found)
	assert.Empty(t, colType)
}

// TestCatalogParser_GetColumnDataType tests getting simplified data types
func TestCatalogParser_GetColumnDataType(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"tags": {
						Name:     "tags",
						Type:     "ARRAY<STRING>",
						DataType: "ARRAY",
					},
				},
			},
		},
	}

	parser := parsers.NewCatalogParser(catalog, map[string]interface{}{})

	dataType, found := parser.GetColumnDataType("model.test.test_model", "tags")
	assert.True(t, found)
	assert.Equal(t, "ARRAY", dataType)
}
