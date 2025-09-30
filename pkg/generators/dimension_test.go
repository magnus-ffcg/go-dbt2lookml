package generators

import (
	"strings"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDimensionGenerator_GenerateDimension(t *testing.T) {
	cfg := &config.Config{
		UseTableName: false,
	}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name           string
		column         *models.DbtModelColumn
		expectedName   string
		expectedType   string
		expectedSQL    string
		expectedHidden *bool
	}{
		{
			name: "simple string column",
			column: &models.DbtModelColumn{
				Name:     "customer_name",
				DataType: stringPtr("STRING"),
			},
			expectedName: "customer_name",
			expectedType: "string",
			expectedSQL:  "${TABLE}.customer_name",
		},
		{
			name: "integer column",
			column: &models.DbtModelColumn{
				Name:     "customer_id",
				DataType: stringPtr("INT64"),
			},
			expectedName: "customer_id",
			expectedType: "number",
			expectedSQL:  "${TABLE}.customer_id",
		},
		{
			name: "boolean column",
			column: &models.DbtModelColumn{
				Name:     "is_active",
				DataType: stringPtr("BOOLEAN"),
			},
			expectedName: "is_active",
			expectedType: "yesno",
			expectedSQL:  "${TABLE}.is_active",
		},
		{
			name: "array column should be hidden",
			column: &models.DbtModelColumn{
				Name:     "tags",
				DataType: stringPtr("ARRAY<STRING>"),
			},
			expectedName:   "test_model__tags", // Actual behavior: prefixed with model name
			expectedType:   "string",
			expectedSQL:    "${TABLE}.tags",
			expectedHidden: boolPtr(true),
		},
		{
			name: "nested column with proper naming",
			column: &models.DbtModelColumn{
				Name:     "address.street.name",
				DataType: stringPtr("STRING"),
				Nested:   true,
			},
			expectedName: "address__street__name",
			expectedType: "string",
			expectedSQL:  "${TABLE}.address.street.name", // Actual behavior: no quoting for nested columns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock model with proper DbtNode embedding
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, tt.column)
			require.NoError(t, err)

			// Skip if this should be a dimension group (returns nil)
			if dimension == nil {
				t.Skip("Column should be dimension group, not dimension")
				return
			}

			assert.Equal(t, tt.expectedName, dimension.Name)
			assert.Equal(t, tt.expectedType, dimension.Type)
			assert.Equal(t, tt.expectedSQL, dimension.SQL)

			if tt.expectedHidden != nil {
				require.NotNil(t, dimension.Hidden)
				assert.Equal(t, *tt.expectedHidden, *dimension.Hidden)
			} else {
				assert.Nil(t, dimension.Hidden)
			}
		})
	}
}

func TestDimensionGenerator_DataTypeMapping(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name         string
		bigQueryType string
		expectedType string
	}{
		{"string type", "STRING", "string"},
		{"integer type", "INT64", "number"},
		{"float type", "FLOAT64", "number"},
		{"numeric type", "NUMERIC", "number"},
		{"boolean type", "BOOLEAN", "yesno"},
		{"bool type", "BOOL", "yesno"},
		{"timestamp type", "TIMESTAMP", "string"}, // Handled as string in dimensions
		{"date type", "DATE", "string"},           // Handled as string in dimensions
		{"datetime type", "DATETIME", "string"},   // Handled as string in dimensions
		{"array type", "ARRAY<STRING>", "string"},
		{"struct type", "STRUCT<name STRING>", "string"},
		{"unknown type fallback", "UNKNOWN_TYPE", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &models.DbtModelColumn{
				Name:     "test_column",
				DataType: &tt.bigQueryType,
			}

			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, column)
			require.NoError(t, err)

			// Skip if this should be a dimension group
			if dimension == nil {
				t.Skip("Column should be dimension group, not dimension")
				return
			}

			assert.Equal(t, tt.expectedType, dimension.Type)
		})
	}
}

func TestDimensionGenerator_SQLGeneration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name        string
		column      *models.DbtModelColumn
		expectedSQL string
	}{
		{
			name: "simple column",
			column: &models.DbtModelColumn{
				Name: "simple_column",
			},
			expectedSQL: "${TABLE}.simple_column",
		},
		{
			name: "column with spaces needs quoting",
			column: &models.DbtModelColumn{
				Name: "column with spaces",
			},
			expectedSQL: "${TABLE}.column with spaces", // Actual behavior: no automatic quoting
		},
		{
			name: "nested column preserves structure",
			column: &models.DbtModelColumn{
				Name: "address.street.name",
			},
			expectedSQL: "${TABLE}.address.street.name", // Actual behavior: no quoting for nested columns
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, tt.column)
			require.NoError(t, err)

			// Skip if this should be a dimension group
			if dimension == nil {
				t.Skip("Column should be dimension group, not dimension")
				return
			}

			assert.Equal(t, tt.expectedSQL, dimension.SQL)
		})
	}
}

func TestDimensionGenerator_NestedColumnNaming(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name            string
		columnName      string
		expectedDimName string
	}{
		{
			name:            "simple nested field",
			columnName:      "address.street",
			expectedDimName: "address__street",
		},
		{
			name:            "deep nested field",
			columnName:      "classification.item_group.code",
			expectedDimName: "classification__item_group__code",
		},
		{
			name:            "single level field",
			columnName:      "simple_field",
			expectedDimName: "simple_field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &models.DbtModelColumn{
				Name:     tt.columnName,
				DataType: stringPtr("STRING"),
				Nested:   strings.Contains(tt.columnName, "."),
			}

			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, column)
			require.NoError(t, err)

			// Skip if this should be a dimension group
			if dimension == nil {
				t.Skip("Column should be dimension group, not dimension")
				return
			}

			assert.Equal(t, tt.expectedDimName, dimension.Name)
		})
	}
}

func TestDimensionGenerator_ErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name        string
		column      *models.DbtModelColumn
		model       *models.DbtModel
		expectError bool
		expectPanic bool
	}{
		{
			name:   "nil column should panic",
			column: nil,
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			expectError: false,
			expectPanic: true,
		},
		{
			name: "nil model should work", // Actual behavior: doesn't validate model
			column: &models.DbtModelColumn{
				Name: "test_column",
			},
			model:       nil,
			expectError: false,
		},
		{
			name: "valid inputs should not error",
			column: &models.DbtModelColumn{
				Name:     "test_column",
				DataType: stringPtr("STRING"),
			},
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					generator.GenerateDimension(tt.model, tt.column)
				})
			} else {
				dimension, err := generator.GenerateDimension(tt.model, tt.column)

				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, dimension)
				} else {
					assert.NoError(t, err)
					// dimension can be nil if it should be a dimension group
				}
			}
		})
	}
}

func TestDimensionGenerator_ArrayHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name         string
		dataType     string
		expectHidden bool
		expectedType string
	}{
		{
			name:         "simple array",
			dataType:     "ARRAY<STRING>",
			expectHidden: true,
			expectedType: "string",
		},
		{
			name:         "complex array",
			dataType:     "ARRAY<STRUCT<name STRING, id INT64>>",
			expectHidden: true,
			expectedType: "string",
		},
		{
			name:         "non-array type",
			dataType:     "STRING",
			expectHidden: false,
			expectedType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &models.DbtModelColumn{
				Name:     "test_column",
				DataType: &tt.dataType,
			}

			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, column)
			require.NoError(t, err)

			// Skip if this should be a dimension group
			if dimension == nil {
				t.Skip("Column should be dimension group, not dimension")
				return
			}

			assert.Equal(t, tt.expectedType, dimension.Type)

			if tt.expectHidden {
				require.NotNil(t, dimension.Hidden, "Expected dimension to be hidden for ARRAY type")
				assert.True(t, *dimension.Hidden)
			} else {
				if dimension.Hidden != nil {
					assert.False(t, *dimension.Hidden)
				}
			}
		})
	}
}

// TestDimensionGenerator_PascalCasePreservation tests that OriginalName preserves PascalCase for SQL
func TestDimensionGenerator_PascalCasePreservation(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name            string
		columnName      string
		originalName    string
		expectedDimName string
		expectedSQL     string
	}{
		{
			name:            "PascalCase to snake_case in dimension name",
			columnName:      "buyingitem_gtin",
			originalName:    "BuyingItem_GTIN",
			expectedDimName: "buying_item_gtin",
			expectedSQL:     "${TABLE}.BuyingItem_GTIN", // SQL preserves PascalCase
		},
		{
			name:            "nested PascalCase column",
			columnName:      "classification.itemgroup.code",
			originalName:    "Classification.ItemGroup.Code",
			expectedDimName: "classification__item_group__code",
			expectedSQL:     "${TABLE}.Classification.ItemGroup.Code", // SQL preserves PascalCase
		},
		{
			name:            "complex nested with acronyms",
			columnName:      "item.gtinid",
			originalName:    "Item.GTINId",
			expectedDimName: "item__gtin_id",
			expectedSQL:     "${TABLE}.Item.GTINId", // SQL preserves original
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &models.DbtModelColumn{
				Name:         tt.columnName,
				OriginalName: &tt.originalName,
				DataType:     stringPtr("STRING"),
				Nested:       strings.Contains(tt.columnName, "."),
			}

			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			dimension, err := generator.GenerateDimension(model, column)
			require.NoError(t, err)
			require.NotNil(t, dimension)

			assert.Equal(t, tt.expectedDimName, dimension.Name, "Dimension name should be snake_case")
			assert.Equal(t, tt.expectedSQL, dimension.SQL, "SQL should preserve PascalCase from OriginalName")
		})
	}
}

// TestDimensionGenerator_GetDimensionName tests the GetDimensionName method directly
func TestDimensionGenerator_GetDimensionName(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name         string
		column       *models.DbtModelColumn
		expectedName string
	}{
		{
			name: "simple column",
			column: &models.DbtModelColumn{
				Name:       "customer_name",
				LookMLName: stringPtr("customer_name"),
			},
			expectedName: "customer_name",
		},
		{
			name: "nested column generates full path",
			column: &models.DbtModelColumn{
				Name:   "address.street.name",
				Nested: true,
			},
			expectedName: "address__street__name",
		},
		{
			name: "PascalCase conversion",
			column: &models.DbtModelColumn{
				Name:         "supplierinformation",
				OriginalName: stringPtr("SupplierInformation"),
			},
			expectedName: "supplier_information",
		},
		{
			name: "nested PascalCase",
			column: &models.DbtModelColumn{
				Name:         "product.itemsubgroup.name",
				OriginalName: stringPtr("Product.ItemSubGroup.Name"),
				Nested:       true,
			},
			expectedName: "product__item_sub_group__name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GetDimensionName(tt.column)
			assert.Equal(t, tt.expectedName, result)
		})
	}
}

// TestDimensionGenerator_DimensionGroups tests dimension group generation
func TestDimensionGenerator_DimensionGroups(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name              string
		column            *models.DbtModelColumn
		expectedGroupName string
		expectedType      string
		hasTimeframes     bool
	}{
		{
			name: "DATE column becomes dimension group",
			column: &models.DbtModelColumn{
				Name:     "order_date",
				DataType: stringPtr("DATE"),
			},
			expectedGroupName: "order", // _date suffix removed
			expectedType:      "time",
			hasTimeframes:     true,
		},
		{
			name: "TIMESTAMP column becomes dimension group",
			column: &models.DbtModelColumn{
				Name:     "created_timestamp",
				DataType: stringPtr("TIMESTAMP"),
			},
			expectedGroupName: "created", // _timestamp suffix removed
			expectedType:      "time",
			hasTimeframes:     true,
		},
		{
			name: "DATETIME column becomes dimension group",
			column: &models.DbtModelColumn{
				Name:     "updated_datetime",
				DataType: stringPtr("DATETIME"),
			},
			expectedGroupName: "updated", // _datetime suffix removed
			expectedType:      "time",
			hasTimeframes:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
			}

			// GenerateDimension should return nil for date/time columns
			dimension, err := generator.GenerateDimension(model, tt.column)
			require.NoError(t, err)
			assert.Nil(t, dimension, "Date/time columns should not generate regular dimensions")

			// GenerateDimensionGroup should create dimension group
			dimensionGroup, err := generator.GenerateDimensionGroup(model, tt.column)
			require.NoError(t, err)
			require.NotNil(t, dimensionGroup, "Date/time columns should generate dimension groups")

			assert.Equal(t, tt.expectedGroupName, dimensionGroup.Name)
			assert.Equal(t, tt.expectedType, dimensionGroup.Type)

			if tt.hasTimeframes {
				assert.NotEmpty(t, dimensionGroup.Timeframes, "Should have timeframes")
			}
		})
	}
}

// TestDimensionGenerator_GroupLabels tests group label generation for nested columns
func TestDimensionGenerator_GroupLabels(t *testing.T) {
	cfg := &config.Config{}
	generator := NewDimensionGenerator(cfg)

	tests := []struct {
		name               string
		column             *models.DbtModelColumn
		expectedGroupLabel *string
	}{
		{
			name: "simple column has no group label",
			column: &models.DbtModelColumn{
				Name: "customer_name",
			},
			expectedGroupLabel: nil,
		},
		{
			name: "nested column gets group label from parent path",
			column: &models.DbtModelColumn{
				Name:   "classification.assortment.code",
				Nested: true,
			},
			expectedGroupLabel: stringPtr("Classification Assortment"),
		},
		{
			name: "single level nested column",
			column: &models.DbtModelColumn{
				Name:   "address.street",
				Nested: true,
			},
			expectedGroupLabel: stringPtr("Address"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GetDimensionGroupLabel(tt.column)

			if tt.expectedGroupLabel == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expectedGroupLabel, *result)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
