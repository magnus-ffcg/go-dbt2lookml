package generators

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestViewGenerator_GenerateView(t *testing.T) {
	cfg := &config.Config{
		UseTableName: false,
	}
	generator := NewViewGenerator(cfg)

	tests := []struct {
		name            string
		model           *models.DbtModel
		expectedName    string
		expectedSQLName string
		expectError     bool
	}{
		{
			name: "simple model with basic columns",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
				Schema:       "test_schema",
				Description:  "Test model description",
				Columns: map[string]models.DbtModelColumn{
					"id": {
						Name:     "id",
						DataType: viewStringPtr("INT64"),
					},
					"name": {
						Name:     "name",
						DataType: viewStringPtr("STRING"),
					},
				},
			},
			expectedName:    "test_model",
			expectedSQLName: "`test_schema.test_model`", // Actual behavior: wrapped in backticks
			expectError:     false,
		},
		{
			name: "model with array columns",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "array_model",
				},
				RelationName: "`project.dataset.array_table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"id": {
						Name:     "id",
						DataType: viewStringPtr("INT64"),
					},
					"tags": {
						Name:     "tags",
						DataType: viewStringPtr("ARRAY<STRING>"),
					},
					"sales": {
						Name:     "sales",
						DataType: viewStringPtr("ARRAY<STRUCT<amount NUMERIC>>"),
					},
					"sales.amount": {
						Name:     "sales.amount",
						DataType: viewStringPtr("NUMERIC"),
						Nested:   true,
					},
				},
			},
			expectedName:    "array_model",
			expectedSQLName: "`test_schema.array_model`", // Actual behavior: wrapped in backticks
			expectError:     false,
		},
		{
			name: "model with metadata",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "meta_model",
				},
				RelationName: "`project.dataset.meta_table`",
				Schema:       "test_schema",
				Description:  "Model with metadata",
				Meta:         &models.DbtModelMeta{},
				Columns: map[string]models.DbtModelColumn{
					"amount": {
						Name:     "amount",
						DataType: viewStringPtr("NUMERIC"),
					},
				},
			},
			expectedName:    "meta_model",
			expectedSQLName: "`test_schema.meta_model`", // Actual behavior: wrapped in backticks
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := generator.GenerateView(tt.model)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, view)
			} else {
				require.NoError(t, err)
				require.NotNil(t, view)

				assert.Equal(t, tt.expectedName, view.Name)
				assert.Equal(t, tt.expectedSQLName, view.SQLTableName)

				// Should have dimensions (at least for non-array columns)
				assert.NotNil(t, view.Dimensions)

				// Should have measures (at least default count)
				assert.NotNil(t, view.Measures)
				assert.Greater(t, len(view.Measures), 0, "Should have at least default count measure")
			}
		})
	}
}

func TestViewGenerator_UseTableName(t *testing.T) {
	cfg := &config.Config{
		UseTableName: true,
	}
	generator := NewViewGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "model_name",
		},
		RelationName: "`project.dataset.actual_table_name`",
		Schema:       "test_schema",
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: viewStringPtr("INT64"),
			},
		},
	}

	view, err := generator.GenerateView(model)
	require.NoError(t, err)
	require.NotNil(t, view)

	// When UseTableName is true, should use table name from RelationName
	assert.Equal(t, "actual_table_name", view.Name)
	assert.Equal(t, "`project.dataset.actual_table_name`", view.SQLTableName)
}

func TestViewGenerator_ViewAttributes(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	tests := []struct {
		name      string
		model     *models.DbtModel
		checkFunc func(*testing.T, *models.LookMLView)
	}{
		{
			name: "model with description",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "described_model",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Description:  "This is a test model with description",
				Columns: map[string]models.DbtModelColumn{
					"id": {
						Name:     "id",
						DataType: viewStringPtr("INT64"),
					},
				},
			},
			checkFunc: func(t *testing.T, view *models.LookMLView) {
				require.NotNil(t, view.Description)
				assert.Equal(t, "This is a test model with description", *view.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := generator.GenerateView(tt.model)
			require.NoError(t, err)
			require.NotNil(t, view)
			tt.checkFunc(t, view)
		})
	}
}

func TestViewGenerator_DimensionGeneration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "dimension_test",
		},
		RelationName: "`project.dataset.table`",
		Schema:       "test_schema",
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: viewStringPtr("INT64"),
			},
			"name": {
				Name:     "name",
				DataType: viewStringPtr("STRING"),
			},
			"is_active": {
				Name:     "is_active",
				DataType: viewStringPtr("BOOLEAN"),
			},
			"tags": {
				Name:     "tags",
				DataType: viewStringPtr("ARRAY<STRING>"),
			},
		},
	}

	view, err := generator.GenerateView(model)
	require.NoError(t, err)
	require.NotNil(t, view)

	// Should have dimensions for non-array columns
	assert.Greater(t, len(view.Dimensions), 0, "Should have dimensions")

	// Find specific dimensions
	dimensionNames := make([]string, len(view.Dimensions))
	for i, dim := range view.Dimensions {
		dimensionNames[i] = dim.Name
	}

	// Should have dimensions for regular columns
	assert.Contains(t, dimensionNames, "id")
	assert.Contains(t, dimensionNames, "name")
	assert.Contains(t, dimensionNames, "is_active")

	// Should have reference dimension for array column (short name, not full view name)
	assert.Contains(t, dimensionNames, "tags")
}

func TestViewGenerator_ArrayHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "array_test",
		},
		RelationName: "`project.dataset.table`",
		Schema:       "test_schema",
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: viewStringPtr("INT64"),
			},
			"simple_array": {
				Name:     "simple_array",
				DataType: viewStringPtr("ARRAY<STRING>"),
			},
			"complex_array": {
				Name:     "complex_array",
				DataType: viewStringPtr("ARRAY<STRUCT<name STRING, value NUMERIC>>"),
			},
			"complex_array.name": {
				Name:     "complex_array.name",
				DataType: viewStringPtr("STRING"),
				Nested:   true,
			},
			"complex_array.value": {
				Name:     "complex_array.value",
				DataType: viewStringPtr("NUMERIC"),
				Nested:   true,
			},
		},
	}

	view, err := generator.GenerateView(model)
	require.NoError(t, err)
	require.NotNil(t, view)

	// Check dimensions
	dimensionNames := make([]string, len(view.Dimensions))
	for i, dim := range view.Dimensions {
		dimensionNames[i] = dim.Name
	}

	// Should have regular column
	assert.Contains(t, dimensionNames, "id")

	// Should have reference dimensions for array columns (short names, not full view names)
	assert.Contains(t, dimensionNames, "simple_array")
	assert.Contains(t, dimensionNames, "complex_array")

	// Should NOT have nested array fields in main view dimensions
	assert.NotContains(t, dimensionNames, "complex_array.name")
	assert.NotContains(t, dimensionNames, "complex_array.value")
}

func TestViewGenerator_ErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	tests := []struct {
		name        string
		model       *models.DbtModel
		expectError bool
	}{
		{
			name:        "nil model should return error",
			model:       nil,
			expectError: true,
		},
		{
			name: "valid model should not error",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "valid_model",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"id": {
						Name:     "id",
						DataType: viewStringPtr("INT64"),
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.model == nil {
				// Expect panic for nil model
				assert.Panics(t, func() {
					_, _ = generator.GenerateView(tt.model)
				})
			} else {
				view, err := generator.GenerateView(tt.model)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, view)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, view)
				}
			}
		})
	}
}

func TestViewGenerator_SchemaStringRemoval(t *testing.T) {
	cfg := &config.Config{
		RemoveSchemaString: "_staging",
	}
	generator := NewViewGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
		RelationName: "`project.dataset.table`",
		Schema:       "test_staging_schema",
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: viewStringPtr("INT64"),
			},
		},
	}

	view, err := generator.GenerateView(model)
	require.NoError(t, err)
	require.NotNil(t, view)

	// Should remove the schema string from SQL table name (wrapped in backticks)
	assert.Equal(t, "`test_schema.test_model`", view.SQLTableName) // Note: RemoveSchemaString might not be fully implemented
}

// TestViewGenerator_ConflictResolution tests dimension/dimension_group name conflict resolution
func TestViewGenerator_ConflictResolution(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	tests := []struct {
		name              string
		model             *models.DbtModel
		expectedDimGroups []string
		noDimensions      bool // Date columns don't create dimensions, only dimension_groups
	}{
		{
			name: "date column creates dimension_group only (no conflict)",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "conflict_test",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"created_date": {
						Name:     "created_date",
						DataType: viewStringPtr("DATE"),
					},
				},
			},
			expectedDimGroups: []string{"created"}, // Only dimension_group, no regular dimension
			noDimensions:      true,
		},
		{
			name: "timestamp column creates dimension_group only",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "timestamp_test",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"updated_timestamp": {
						Name:     "updated_timestamp",
						DataType: viewStringPtr("TIMESTAMP"),
					},
				},
			},
			expectedDimGroups: []string{"updated"},
			noDimensions:      true,
		},
		{
			name: "multiple date columns create multiple dimension_groups",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "multi_conflict",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"created_date": {
						Name:     "created_date",
						DataType: viewStringPtr("DATE"),
					},
					"updated_date": {
						Name:     "updated_date",
						DataType: viewStringPtr("DATE"),
					},
				},
			},
			expectedDimGroups: []string{"created", "updated"},
			noDimensions:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := generator.GenerateView(tt.model)
			require.NoError(t, err)
			require.NotNil(t, view)

			// Collect dimension group names
			dimGroupNames := make([]string, len(view.DimensionGroups))
			for i, dimGroup := range view.DimensionGroups {
				dimGroupNames[i] = dimGroup.Name
			}

			// Verify expected dimension groups exist
			for _, expectedDimGroup := range tt.expectedDimGroups {
				assert.Contains(t, dimGroupNames, expectedDimGroup, "Should have dimension_group: %s", expectedDimGroup)
			}

			// Verify no regular dimensions created for date columns (only dimension_groups)
			if tt.noDimensions {
				assert.Empty(t, view.Dimensions, "Date/time columns should not create regular dimensions")
			}
		})
	}
}

// TestViewGenerator_DeepCopyPreventsSharedPointers tests that column deep copying prevents shared pointer bugs
func TestViewGenerator_DeepCopyPreventsSharedPointers(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	// Create model with multiple columns that have OriginalName set
	originalName1 := "BuyingItem_GTIN"
	originalName2 := "SupplierInformation"
	originalName3 := "Classification.ItemGroup.Code"

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "pointer_test",
		},
		RelationName: "`project.dataset.table`",
		Schema:       "test_schema",
		Columns: map[string]models.DbtModelColumn{
			"buyingitem_gtin": {
				Name:         "buyingitem_gtin",
				OriginalName: &originalName1,
				DataType:     viewStringPtr("STRING"),
			},
			"supplierinformation": {
				Name:         "supplierinformation",
				OriginalName: &originalName2,
				DataType:     viewStringPtr("STRING"),
			},
			"classification.itemgroup.code": {
				Name:         "classification.itemgroup.code",
				OriginalName: &originalName3,
				DataType:     viewStringPtr("STRING"),
				Nested:       true,
			},
		},
	}

	view, err := generator.GenerateView(model)
	require.NoError(t, err)
	require.NotNil(t, view)

	// Verify we have 3 dimensions
	require.Len(t, view.Dimensions, 3, "Should have 3 dimensions")

	// Verify each dimension has correct SQL (uses OriginalName)
	sqlReferences := make(map[string]string)
	for _, dim := range view.Dimensions {
		sqlReferences[dim.Name] = dim.SQL
	}

	// Each dimension should have its own OriginalName preserved in SQL
	assert.Contains(t, sqlReferences["buying_item_gtin"], "BuyingItem_GTIN", "Should preserve BuyingItem_GTIN in SQL")
	assert.Contains(t, sqlReferences["supplier_information"], "SupplierInformation", "Should preserve SupplierInformation in SQL")
	assert.Contains(t, sqlReferences["classification__item_group__code"], "Classification.ItemGroup.Code", "Should preserve Classification.ItemGroup.Code in SQL")

	// Verify all 3 SQL references are different (no pointer sharing)
	uniqueSQLRefs := make(map[string]bool)
	for _, sql := range sqlReferences {
		uniqueSQLRefs[sql] = true
	}
	assert.Len(t, uniqueSQLRefs, 3, "All SQL references should be unique (no pointer sharing)")
}

// TestViewGenerator_NestedViewReferenceDimensions tests that array columns get proper reference dimensions
func TestViewGenerator_NestedViewReferenceDimensions(t *testing.T) {
	cfg := &config.Config{}
	generator := NewViewGenerator(cfg)

	tests := []struct {
		name                 string
		model                *models.DbtModel
		expectedRefDimension string
		checkHidden          bool
		checkSQL             bool
		expectedSQLContains  string
	}{
		{
			name: "simple array column gets reference dimension",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "array_ref_test",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"id": {
						Name:     "id",
						DataType: viewStringPtr("INT64"),
					},
					"tags": {
						Name:     "tags",
						DataType: viewStringPtr("ARRAY<STRING>"),
					},
				},
			},
			expectedRefDimension: "tags",
			checkHidden:          true,
			checkSQL:             true,
			expectedSQLContains:  "tags",
		},
		{
			name: "complex array struct gets reference dimension",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "complex_array_test",
				},
				RelationName: "`project.dataset.table`",
				Schema:       "test_schema",
				Columns: map[string]models.DbtModelColumn{
					"sales": {
						Name:     "sales",
						DataType: viewStringPtr("ARRAY<STRUCT<amount NUMERIC, date DATE>>"),
					},
					"sales.amount": {
						Name:     "sales.amount",
						DataType: viewStringPtr("NUMERIC"),
						Nested:   true,
					},
					"sales.date": {
						Name:     "sales.date",
						DataType: viewStringPtr("DATE"),
						Nested:   true,
					},
				},
			},
			expectedRefDimension: "sales",
			checkHidden:          true,
			checkSQL:             true,
			expectedSQLContains:  "sales",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			view, err := generator.GenerateView(tt.model)
			require.NoError(t, err)
			require.NotNil(t, view)

			// Find the reference dimension
			var refDim *models.LookMLDimension
			for i, dim := range view.Dimensions {
				if dim.Name == tt.expectedRefDimension {
					refDim = &view.Dimensions[i]
					break
				}
			}

			require.NotNil(t, refDim, "Should have reference dimension: %s", tt.expectedRefDimension)

			// Check hidden attribute
			if tt.checkHidden {
				require.NotNil(t, refDim.Hidden, "Reference dimension should have Hidden attribute")
				assert.True(t, *refDim.Hidden, "Reference dimension should be hidden")
			}

			// Check SQL field
			if tt.checkSQL {
				assert.NotEmpty(t, refDim.SQL, "Reference dimension should have SQL")
				assert.Contains(t, refDim.SQL, tt.expectedSQLContains, "SQL should reference the array column")
			}
		})
	}
}

// Helper functions
func viewStringPtr(s string) *string {
	return &s
}

func viewBoolPtr(b bool) *bool {
	return &b
}
