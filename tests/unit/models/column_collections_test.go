package models

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestColumnCollections_NewColumnCollections(t *testing.T) {
	// Create a test model with various column types
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: stringPtr("INT64"),
			},
			"name": {
				Name:     "name",
				DataType: stringPtr("STRING"),
			},
			"sales": {
				Name:     "sales",
				DataType: stringPtr("ARRAY<STRUCT<amount NUMERIC, date DATE>>"),
			},
			"sales.amount": {
				Name:     "sales.amount",
				DataType: stringPtr("NUMERIC"),
				Nested:   true,
			},
			"sales.date": {
				Name:     "sales.date",
				DataType: stringPtr("DATE"),
				Nested:   true,
			},
			"tags": {
				Name:     "tags",
				DataType: stringPtr("ARRAY<STRING>"),
			},
		},
	}

	collections := models.NewColumnCollections(model, nil)
	require.NotNil(t, collections)

	// Test main view columns - should exclude ARRAY columns
	assert.Contains(t, collections.MainViewColumns, "id")
	assert.Contains(t, collections.MainViewColumns, "name")
	assert.NotContains(t, collections.MainViewColumns, "sales")     // ARRAY should be excluded
	assert.NotContains(t, collections.MainViewColumns, "tags")     // ARRAY should be excluded
	assert.NotContains(t, collections.MainViewColumns, "sales.amount") // Nested should be excluded

	// Test nested view columns - should contain ARRAY columns
	assert.Contains(t, collections.NestedViewColumns, "sales")
	assert.Contains(t, collections.NestedViewColumns, "tags")
	
	// Check that sales nested view contains its children
	if salesNested, exists := collections.NestedViewColumns["sales"]; exists {
		assert.Contains(t, salesNested, "sales.amount")
		assert.Contains(t, salesNested, "sales.date")
		assert.Contains(t, salesNested, "sales") // Should contain the array field itself as hidden reference
	}
}

func TestColumnCollections_ArrayClassification(t *testing.T) {
	tests := []struct {
		name                string
		columns             map[string]models.DbtModelColumn
		expectedMainView    []string
		expectedNestedViews []string
	}{
		{
			name: "simple array without children",
			columns: map[string]models.DbtModelColumn{
				"tags": {
					Name:     "tags",
					DataType: stringPtr("ARRAY<STRING>"),
				},
				"id": {
					Name:     "id",
					DataType: stringPtr("INT64"),
				},
			},
			expectedMainView:    []string{"id"},
			expectedNestedViews: []string{"tags"},
		},
		{
			name: "array with struct children",
			columns: map[string]models.DbtModelColumn{
				"sales": {
					Name:     "sales",
					DataType: stringPtr("ARRAY<STRUCT<amount NUMERIC>>"),
				},
				"sales.amount": {
					Name:     "sales.amount",
					DataType: stringPtr("NUMERIC"),
					Nested:   true,
				},
				"id": {
					Name:     "id",
					DataType: stringPtr("INT64"),
				},
			},
			expectedMainView:    []string{"id"},
			expectedNestedViews: []string{"sales"},
		},
		{
			name: "nested arrays",
			columns: map[string]models.DbtModelColumn{
				"sales": {
					Name:     "sales",
					DataType: stringPtr("ARRAY<STRUCT<items ARRAY<STRING>>>"),
				},
				"sales.items": {
					Name:     "sales.items",
					DataType: stringPtr("ARRAY<STRING>"),
					Nested:   true,
				},
				"id": {
					Name:     "id",
					DataType: stringPtr("INT64"),
				},
			},
			expectedMainView:    []string{"id"},
			expectedNestedViews: []string{"sales", "sales.items"},
		},
		{
			name: "mixed column types",
			columns: map[string]models.DbtModelColumn{
				"id": {
					Name:     "id",
					DataType: stringPtr("INT64"),
				},
				"name": {
					Name:     "name",
					DataType: stringPtr("STRING"),
				},
				"created_at": {
					Name:     "created_at",
					DataType: stringPtr("TIMESTAMP"),
				},
				"tags": {
					Name:     "tags",
					DataType: stringPtr("ARRAY<STRING>"),
				},
				"metadata": {
					Name:     "metadata",
					DataType: stringPtr("STRUCT<version INT64>"),
				},
				"metadata.version": {
					Name:     "metadata.version",
					DataType: stringPtr("INT64"),
					Nested:   true,
				},
			},
			expectedMainView:    []string{"id", "name", "created_at"},
			expectedNestedViews: []string{"tags"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Columns: tt.columns,
			}

			collections := models.NewColumnCollections(model, nil)
			require.NotNil(t, collections)

			// Check main view columns
			for _, expectedCol := range tt.expectedMainView {
				assert.Contains(t, collections.MainViewColumns, expectedCol, 
					"Expected %s to be in MainViewColumns", expectedCol)
			}

			// Check nested view columns
			for _, expectedNestedView := range tt.expectedNestedViews {
				assert.Contains(t, collections.NestedViewColumns, expectedNestedView,
					"Expected %s to be in NestedViewColumns", expectedNestedView)
			}

			// Verify ARRAY columns are not in main view
			for colName, col := range tt.columns {
				if col.DataType != nil && isArrayType(*col.DataType) {
					assert.NotContains(t, collections.MainViewColumns, colName,
						"ARRAY column %s should not be in MainViewColumns", colName)
				}
			}
		})
	}
}

func TestColumnCollections_HierarchyBuilding(t *testing.T) {
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		Columns: map[string]models.DbtModelColumn{
			"sales": {
				Name:     "sales",
				DataType: stringPtr("ARRAY<STRUCT<amount NUMERIC, items ARRAY<STRING>>>"),
			},
			"sales.amount": {
				Name:     "sales.amount",
				DataType: stringPtr("NUMERIC"),
				Nested:   true,
			},
			"sales.items": {
				Name:     "sales.items",
				DataType: stringPtr("ARRAY<STRING>"),
				Nested:   true,
			},
		},
	}

	collections := models.NewColumnCollections(model, nil)
	require.NotNil(t, collections)

	// Should have nested views for both sales and sales.items
	assert.Contains(t, collections.NestedViewColumns, "sales")
	assert.Contains(t, collections.NestedViewColumns, "sales.items")

	// Sales nested view should contain its direct children
	if salesNested, exists := collections.NestedViewColumns["sales"]; exists {
		assert.Contains(t, salesNested, "sales.amount")
		assert.Contains(t, salesNested, "sales") // Hidden reference dimension
	}

	// Sales.items nested view should contain itself
	if itemsNested, exists := collections.NestedViewColumns["sales.items"]; exists {
		assert.Contains(t, itemsNested, "sales.items") // Hidden reference dimension
	}
}

func TestColumnCollections_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		model   *models.DbtModel
		wantErr bool
	}{
		{
			name: "empty model",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "empty_model",
				},
				Columns: map[string]models.DbtModelColumn{},
			},
			wantErr: false,
		},
		{
			name: "nil columns map",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "nil_columns_model",
				},
				Columns: nil,
			},
			wantErr: false,
		},
		{
			name: "columns with nil data types",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "nil_datatype_model",
				},
				Columns: map[string]models.DbtModelColumn{
					"col1": {
						Name:     "col1",
						DataType: nil,
					},
					"col2": {
						Name:     "col2",
						DataType: stringPtr("STRING"),
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collections := models.NewColumnCollections(tt.model, nil)
			
			if tt.wantErr {
				assert.Nil(t, collections)
			} else {
				assert.NotNil(t, collections)
				assert.NotNil(t, collections.MainViewColumns)
				assert.NotNil(t, collections.NestedViewColumns)
				assert.NotNil(t, collections.ExcludedColumns)
			}
		})
	}
}

func TestColumnCollections_ArrayModelsParameter(t *testing.T) {
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		Columns: map[string]models.DbtModelColumn{
			"sales": {
				Name:     "sales",
				DataType: stringPtr("ARRAY<STRUCT<amount NUMERIC>>"),
			},
			"sales.amount": {
				Name:     "sales.amount",
				DataType: stringPtr("NUMERIC"),
				Nested:   true,
			},
		},
	}

	// Test with array models parameter
	arrayModels := []string{"sales"} // Explicitly specify sales as array model
	collections := models.NewColumnCollections(model, arrayModels)
	require.NotNil(t, collections)

	// Should still work correctly
	assert.Contains(t, collections.NestedViewColumns, "sales")
	if salesNested, exists := collections.NestedViewColumns["sales"]; exists {
		assert.Contains(t, salesNested, "sales.amount")
	}
}

func TestColumnCollections_NonArrayStructHandling(t *testing.T) {
	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		Columns: map[string]models.DbtModelColumn{
			"metadata": {
				Name:     "metadata",
				DataType: stringPtr("STRUCT<version INT64, name STRING>"),
			},
			"metadata.version": {
				Name:     "metadata.version",
				DataType: stringPtr("INT64"),
				Nested:   true,
			},
			"metadata.name": {
				Name:     "metadata.name",
				DataType: stringPtr("STRING"),
				Nested:   true,
			},
		},
	}

	collections := models.NewColumnCollections(model, nil)
	require.NotNil(t, collections)

	// Actual behavior: Non-ARRAY STRUCT parent is not in main view, only nested fields
	assert.NotContains(t, collections.MainViewColumns, "metadata")
	
	// Nested fields of non-ARRAY STRUCT should be in main view
	assert.Contains(t, collections.MainViewColumns, "metadata.version")
	assert.Contains(t, collections.MainViewColumns, "metadata.name")
	
	// Should not create nested views for non-ARRAY STRUCT
	assert.NotContains(t, collections.NestedViewColumns, "metadata")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func isArrayType(dataType string) bool {
	return len(dataType) > 5 && dataType[:5] == "ARRAY"
}
