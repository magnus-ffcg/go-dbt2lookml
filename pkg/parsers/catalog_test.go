package parsers

import (
	"fmt"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCatalogParser_NewCatalogParser tests catalog parser creation
func TestCatalogParser_NewCatalogParser(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{},
	}
	rawCatalog := map[string]interface{}{}

	parser := NewCatalogParser(catalog, rawCatalog, &config.Config{})
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
			parser := NewCatalogParser(tt.catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

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

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

	dataType, found := parser.GetColumnDataType("model.test.test_model", "tags")
	assert.True(t, found)
	assert.Equal(t, "ARRAY", dataType)
}

// TestCatalogParser_ColumnNormalization tests that column names are normalized to lowercase
func TestCatalogParser_ColumnNormalization(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"BuyingItem_GTIN": {
						Name:         "BuyingItem_GTIN",
						OriginalName: "",
						Type:         "STRING",
						DataType:     "STRING",
					},
					"SupplierInformation": {
						Name:         "SupplierInformation",
						OriginalName: "",
						Type:         "STRING",
						DataType:     "STRING",
					},
				},
			},
		},
	}

	// Normalize column names
	catalogNode := catalog.Nodes["model.test.test_model"]
	catalogNode.NormalizeColumnNames()

	// Check that names are now lowercase
	_, foundLower1 := catalogNode.Columns["buyingitem_gtin"]
	assert.True(t, foundLower1, "Column name should be normalized to lowercase")

	_, foundLower2 := catalogNode.Columns["supplierinformation"]
	assert.True(t, foundLower2, "Column name should be normalized to lowercase")

	// Check that OriginalName preserves PascalCase
	col1 := catalogNode.Columns["buyingitem_gtin"]
	assert.Equal(t, "BuyingItem_GTIN", col1.OriginalName, "OriginalName should preserve PascalCase")

	col2 := catalogNode.Columns["supplierinformation"]
	assert.Equal(t, "SupplierInformation", col2.OriginalName, "OriginalName should preserve PascalCase")

	// Verify lowercase name is set
	assert.Equal(t, "buyingitem_gtin", col1.Name)
	assert.Equal(t, "supplierinformation", col2.Name)
}

// TestCatalogParser_ProcessModelColumnsWithNested tests processing nested columns
func TestCatalogParser_ProcessModelColumnsWithNested(t *testing.T) {
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name:     "nested_model",
			UniqueID: "model.test.nested_model",
		},
		Columns: map[string]models.DbtModelColumn{},
	}

	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.nested_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"Classification.ItemGroup.Code": {
						Name:         "classification.itemgroup.code",
						OriginalName: "Classification.ItemGroup.Code",
						Type:         "STRING",
						DataType:     "STRING",
					},
					"Product.ItemSubGroup.Name": {
						Name:         "product.itemsubgroup.name",
						OriginalName: "Product.ItemSubGroup.Name",
						Type:         "STRING",
						DataType:     "STRING",
					},
				},
			},
		},
	}

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

	processedModel, err := parser.ProcessModelColumns(model)
	require.NoError(t, err)
	require.NotNil(t, processedModel)

	// Should have 2 columns
	assert.Len(t, processedModel.Columns, 2)

	// Check nested column with PascalCase preservation
	col1, exists := processedModel.Columns["classification.itemgroup.code"]
	require.True(t, exists, "Should have normalized lowercase column name")
	require.NotNil(t, col1.OriginalName, "Should have OriginalName set")
	assert.Equal(t, "Classification.ItemGroup.Code", *col1.OriginalName, "OriginalName should preserve PascalCase")
	assert.NotNil(t, col1.DataType)
	assert.Equal(t, "STRING", *col1.DataType)
}

// TestCatalogParser_ProcessModelColumnsPreservesOriginalName tests that OriginalName is copied properly
func TestCatalogParser_ProcessModelColumnsPreservesOriginalName(t *testing.T) {
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name:     "original_name_test",
			UniqueID: "model.test.original_name_test",
		},
		Columns: map[string]models.DbtModelColumn{},
	}

	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.original_name_test": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					// Catalog has PascalCase names that will be normalized
					"BuyingItem_GTIN": {
						Name:     "BuyingItem_GTIN",
						Type:     "STRING",
						DataType: "STRING",
					},
					"SupplierInformation": {
						Name:     "SupplierInformation",
						Type:     "STRING",
						DataType: "STRING",
					},
					"Classification.ItemGroup.Code": {
						Name:     "Classification.ItemGroup.Code",
						Type:     "STRING",
						DataType: "STRING",
					},
				},
			},
		},
	}

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

	processedModel, err := parser.ProcessModelColumns(model)
	require.NoError(t, err)
	require.NotNil(t, processedModel)

	// After processing, columns are normalized to lowercase
	col1 := processedModel.Columns["buyingitem_gtin"]
	col2 := processedModel.Columns["supplierinformation"]
	col3 := processedModel.Columns["classification.itemgroup.code"]

	require.NotNil(t, col1.OriginalName)
	require.NotNil(t, col2.OriginalName)
	require.NotNil(t, col3.OriginalName)

	// OriginalName preserves PascalCase from catalog
	assert.Equal(t, "BuyingItem_GTIN", *col1.OriginalName)
	assert.Equal(t, "SupplierInformation", *col2.OriginalName)
	assert.Equal(t, "Classification.ItemGroup.Code", *col3.OriginalName)

	// Verify they are different pointers (no sharing)
	assert.NotEqual(t, fmt.Sprintf("%p", col1.OriginalName), fmt.Sprintf("%p", col2.OriginalName), "Should have different pointers")
	assert.NotEqual(t, fmt.Sprintf("%p", col1.OriginalName), fmt.Sprintf("%p", col3.OriginalName), "Should have different pointers")
}

// TestCatalogParser_GetColumnInnerTypes tests getting inner types for complex columns
func TestCatalogParser_GetColumnInnerTypes(t *testing.T) {
	catalog := &models.DbtCatalog{
		Nodes: map[string]models.DbtCatalogNode{
			"model.test.test_model": {
				Columns: map[string]models.DbtCatalogNodeColumn{
					"metadata": {
						Name:       "metadata",
						Type:       "ARRAY<STRUCT<key STRING, value STRING>>",
						DataType:   "ARRAY",
						InnerTypes: []string{"STRUCT", "STRING", "STRING"},
					},
				},
			},
		},
	}

	parser := NewCatalogParser(catalog, map[string]interface{}{}, &config.Config{})

	innerTypes, found := parser.GetColumnInnerTypes("model.test.test_model", "metadata")
	assert.True(t, found)
	assert.Equal(t, []string{"STRUCT", "STRING", "STRING"}, innerTypes)

	// Test non-existent column
	innerTypes, found = parser.GetColumnInnerTypes("model.test.test_model", "missing")
	assert.False(t, found)
	assert.Nil(t, innerTypes)
}

// TestCatalogParser_ProcessModelColumnsEdgeCases tests edge cases
func TestCatalogParser_ProcessModelColumnsEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		model         *models.DbtModel
		catalog       *models.DbtCatalog
		expectColumns int
	}{
		{
			name: "empty catalog node",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name:     "empty_model",
					UniqueID: "model.test.empty_model",
				},
				Columns: map[string]models.DbtModelColumn{},
			},
			catalog: &models.DbtCatalog{
				Nodes: map[string]models.DbtCatalogNode{
					"model.test.empty_model": {
						Columns: map[string]models.DbtCatalogNodeColumn{},
					},
				},
			},
			expectColumns: 0,
		},
		{
			name: "model with only ARRAY columns",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name:     "array_only",
					UniqueID: "model.test.array_only",
				},
				Columns: map[string]models.DbtModelColumn{},
			},
			catalog: &models.DbtCatalog{
				Nodes: map[string]models.DbtCatalogNode{
					"model.test.array_only": {
						Columns: map[string]models.DbtCatalogNodeColumn{
							"tags": {
								Name:       "tags",
								Type:       "ARRAY<STRING>",
								DataType:   "ARRAY<STRING>",
								InnerTypes: []string{"STRING"},
							},
							"items": {
								Name:       "items",
								Type:       "ARRAY<INT64>",
								DataType:   "ARRAY<INT64>",
								InnerTypes: []string{"INT64"},
							},
						},
					},
				},
			},
			expectColumns: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCatalogParser(tt.catalog, map[string]interface{}{}, &config.Config{})

			processedModel, err := parser.ProcessModelColumns(tt.model)
			require.NoError(t, err)
			require.NotNil(t, processedModel)

			assert.Len(t, processedModel.Columns, tt.expectColumns)
		})
	}
}
