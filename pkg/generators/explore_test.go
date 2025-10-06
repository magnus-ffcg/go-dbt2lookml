package generators

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExploreGenerator_GenerateExplore(t *testing.T) {
	cfg := &config.Config{
		UseTableName: false,
	}
	generator := NewExploreGenerator(cfg)

	tests := []struct {
		name             string
		model            *models.DbtModel
		expectedName     string
		expectedViewName string
		expectedLabel    *string
		expectError      bool
	}{
		{
			name: "simple model",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.test_table`",
				Description:  "Test model description",
			},
			expectedName:     "test_model",
			expectedViewName: "test_model",
			expectedLabel:    exploreStringPtr("Test Model"),
			expectError:      false,
		},
		{
			name: "model with underscores in name",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "customer_order_summary",
				},
				RelationName: "`project.dataset.customer_order_summary`",
			},
			expectedName:     "customer_order_summary",
			expectedViewName: "customer_order_summary",
			expectedLabel:    exploreStringPtr("Customer Order Summary"),
			expectError:      false,
		},
		{
			name: "model with description",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "meta_model",
				},
				RelationName: "`project.dataset.meta_table`",
				Description:  "Model with custom metadata",
			},
			expectedName:     "meta_model",
			expectedViewName: "meta_model",
			expectedLabel:    exploreStringPtr("Meta Model"), // Auto-generated from name
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explore, err := generator.GenerateExplore(tt.model)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, explore)
			} else {
				require.NoError(t, err)
				require.NotNil(t, explore)

				assert.Equal(t, tt.expectedName, explore.Name)
				assert.Equal(t, tt.expectedViewName, explore.ViewName)

				if tt.expectedLabel != nil {
					require.NotNil(t, explore.Label)
					assert.Equal(t, *tt.expectedLabel, *explore.Label)
				}
			}
		})
	}
}

func TestExploreGenerator_UseTableName(t *testing.T) {
	cfg := &config.Config{
		UseTableName: true,
	}
	generator := NewExploreGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "model_name",
		},
		RelationName: "`project.dataset.actual_table_name`",
	}

	explore, err := generator.GenerateExplore(model)
	require.NoError(t, err)
	require.NotNil(t, explore)

	// When UseTableName is true, should use table name from RelationName
	assert.Equal(t, "actual_table_name", explore.Name)
	assert.Equal(t, "actual_table_name", explore.ViewName)
}

func TestExploreGenerator_ExploreAttributes(t *testing.T) {
	cfg := &config.Config{}
	generator := NewExploreGenerator(cfg)

	tests := []struct {
		name      string
		model     *models.DbtModel
		checkFunc func(*testing.T, *models.LookMLExplore)
	}{
		{
			name: "explore with description from model",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "described_model",
				},
				RelationName: "`project.dataset.table`",
				Description:  "This is a test model with description",
			},
			checkFunc: func(t *testing.T, explore *models.LookMLExplore) {
				require.NotNil(t, explore.Description)
				assert.Equal(t, "This is a test model with description", *explore.Description)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explore, err := generator.GenerateExplore(tt.model)
			require.NoError(t, err)
			require.NotNil(t, explore)
			tt.checkFunc(t, explore)
		})
	}
}

func TestExploreGenerator_LabelGeneration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewExploreGenerator(cfg)

	tests := []struct {
		name          string
		modelName     string
		expectedLabel string
	}{
		{
			name:          "simple name",
			modelName:     "users",
			expectedLabel: "Users",
		},
		{
			name:          "name with underscores",
			modelName:     "customer_orders",
			expectedLabel: "Customer Orders",
		},
		{
			name:          "complex name with multiple underscores",
			modelName:     "dim_customer_order_summary",
			expectedLabel: "Dim Customer Order Summary",
		},
		{
			name:          "single character",
			modelName:     "a",
			expectedLabel: "A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: tt.modelName,
				},
				RelationName: "`project.dataset.table`",
			}

			explore, err := generator.GenerateExplore(model)
			require.NoError(t, err)
			require.NotNil(t, explore)
			require.NotNil(t, explore.Label)
			assert.Equal(t, tt.expectedLabel, *explore.Label)
		})
	}
}

func TestExploreGenerator_JoinGeneration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewExploreGenerator(cfg)

	tests := []struct {
		name          string
		model         *models.DbtModel
		expectedJoins int
		checkJoins    func(*testing.T, []models.LookMLJoin)
	}{
		{
			name: "model without joins",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "simple_model",
				},
				RelationName: "`project.dataset.table`",
			},
			expectedJoins: 0,
			checkJoins: func(t *testing.T, joins []models.LookMLJoin) {
				assert.Empty(t, joins)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			explore, err := generator.GenerateExplore(tt.model)
			require.NoError(t, err)
			require.NotNil(t, explore)

			assert.Len(t, explore.Joins, tt.expectedJoins)
			tt.checkJoins(t, explore.Joins)
		})
	}
}

func TestExploreGenerator_ErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := NewExploreGenerator(cfg)

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
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.model == nil {
				// Expect panic for nil model
				assert.Panics(t, func() {
					_, _ = generator.GenerateExplore(tt.model)
				})
			} else {
				explore, err := generator.GenerateExplore(tt.model)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, explore)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, explore)
				}
			}
		})
	}
}

func TestExploreGenerator_DescriptionPriority(t *testing.T) {
	cfg := &config.Config{}
	generator := NewExploreGenerator(cfg)

	tests := []struct {
		name                string
		modelDescription    string
		expectedDescription *string
	}{
		{
			name:                "model with description",
			modelDescription:    "Model description",
			expectedDescription: exploreStringPtr("Model description"),
		},
		{
			name:                "model without description",
			modelDescription:    "",
			expectedDescription: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				RelationName: "`project.dataset.table`",
				Description:  tt.modelDescription,
			}

			explore, err := generator.GenerateExplore(model)
			require.NoError(t, err)
			require.NotNil(t, explore)

			if tt.expectedDescription == nil {
				assert.Nil(t, explore.Description)
			} else {
				require.NotNil(t, explore.Description)
				assert.Equal(t, *tt.expectedDescription, *explore.Description)
			}
		})
	}
}

// Helper functions
func exploreStringPtr(s string) *string {
	return &s
}

func exploreBoolPtr(b bool) *bool {
	return &b
}

func joinTypePtr(jt enums.LookerJoinType) *enums.LookerJoinType {
	return &jt
}

func relationshipPtr(rt enums.LookerRelationshipType) *enums.LookerRelationshipType {
	return &rt
}
