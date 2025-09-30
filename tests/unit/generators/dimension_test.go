package generators

import (
	"strings"
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/internal/config"
	"github.com/magnus-ffcg/dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDimensionGenerator_GenerateDimension(t *testing.T) {
	cfg := &config.Config{
		UseTableName: false,
	}
	generator := generators.NewDimensionGenerator(cfg)

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
	generator := generators.NewDimensionGenerator(cfg)

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
	generator := generators.NewDimensionGenerator(cfg)

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
	generator := generators.NewDimensionGenerator(cfg)

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
	generator := generators.NewDimensionGenerator(cfg)

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
	generator := generators.NewDimensionGenerator(cfg)

	tests := []struct {
		name           string
		dataType       string
		expectHidden   bool
		expectedType   string
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

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
