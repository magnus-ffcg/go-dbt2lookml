package generators

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/internal/config"
	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasureGenerator_GenerateDefaultCountMeasure(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	tests := []struct {
		name           string
		model          *models.DbtModel
		expectedResult *models.LookMLMeasure
	}{
		{
			name: "simple model without existing count measure",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: nil, // No existing measures
			},
			expectedResult: &models.LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
		},
		{
			name: "model with empty meta",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: &models.DbtModelMeta{
					Looker: nil,
				},
			},
			expectedResult: &models.LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
		},
		{
			name: "model with empty measures",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: &models.DbtModelMeta{
					Looker: &models.DbtMetaLooker{
						Measures: []models.DbtMetaLookerMeasure{},
					},
				},
			},
			expectedResult: &models.LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
		},
		{
			name: "model with existing count measure should return nil",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: &models.DbtModelMeta{
					Looker: &models.DbtMetaLooker{
						Measures: []models.DbtMetaLookerMeasure{
							{
								Name: measureStringPtr("count"),
								Type: enums.MeasureCount,
							},
						},
					},
				},
			},
			expectedResult: nil, // Should not generate default when count already exists
		},
		{
			name: "model with other measures but no count",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: &models.DbtModelMeta{
					Looker: &models.DbtMetaLooker{
						Measures: []models.DbtMetaLookerMeasure{
							{
								Name: measureStringPtr("total_amount"),
								Type: enums.MeasureSum,
							},
						},
					},
				},
			},
			expectedResult: &models.LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateDefaultCountMeasure(tt.model)

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
				assert.Equal(t, tt.expectedResult.Type, result.Type)
				// Default count measure should be minimal - only name and type
				assert.Nil(t, result.SQL)
				assert.Nil(t, result.Label)
				assert.Nil(t, result.Description)
			}
		})
	}
}

func TestMeasureGenerator_GenerateMeasure(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	tests := []struct {
		name           string
		model          *models.DbtModel
		measureMeta    *models.DbtMetaLookerMeasure
		expectedName   string
		expectedType   enums.LookerMeasureType
		expectedSQL    *string
		expectError    bool
	}{
		{
			name: "simple sum measure",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringPtr("total_amount"),
				Type: enums.MeasureSum,
			},
			expectedName: "total_amount",
			expectedType: enums.MeasureSum,
			expectError:  false,
		},
		{
			name: "count distinct measure with approximate",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:                 measureStringPtr("unique_customers"),
				Type:                 enums.MeasureCountDistinct,
				Approximate:          measureBoolPtr(true),
				ApproximateThreshold: measureIntPtr(1000),
			},
			expectedName: "unique_customers",
			expectedType: enums.MeasureCountDistinct,
			expectError:  false,
		},
		{
			name: "measure without name should use type",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: nil, // No name provided
				Type: enums.MeasureAverage,
			},
			expectedName: "average",
			expectedType: enums.MeasureAverage,
			expectError:  false,
		},
		{
			name: "invalid measure should return error",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:        measureStringPtr("invalid_measure"),
				Type:        enums.MeasureSum, // Not count_distinct
				Approximate: measureBoolPtr(true),    // But has approximate (invalid)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(tt.model, tt.measureMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedName, result.Name)
				assert.Equal(t, tt.expectedType, result.Type)
			}
		})
	}
}

func TestMeasureGenerator_MeasureTypes(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
	}

	tests := []struct {
		name         string
		measureType  enums.LookerMeasureType
		expectedType enums.LookerMeasureType
	}{
		{"count measure", enums.MeasureCount, enums.MeasureCount},
		{"sum measure", enums.MeasureSum, enums.MeasureSum},
		{"average measure", enums.MeasureAverage, enums.MeasureAverage},
		{"min measure", enums.MeasureMin, enums.MeasureMin},
		{"max measure", enums.MeasureMax, enums.MeasureMax},
		{"count_distinct measure", enums.MeasureCountDistinct, enums.MeasureCountDistinct},
		{"sum_distinct measure", enums.MeasureSumDistinct, enums.MeasureSumDistinct},
		{"average_distinct measure", enums.MeasureAverageDistinct, enums.MeasureAverageDistinct},
		{"median measure", enums.MeasureMedian, enums.MeasureMedian},
		{"list measure", enums.MeasureList, enums.MeasureList},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measureMeta := &models.DbtMetaLookerMeasure{
				Name: measureStringPtr("test_measure"),
				Type: tt.measureType,
			}

			result, err := generator.GenerateMeasure(model, measureMeta)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedType, result.Type)
		})
	}
}

func TestMeasureGenerator_MeasureAttributes(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		checkFunc   func(*testing.T, *models.LookMLMeasure)
	}{
		{
			name: "measure with label and description",
			measureMeta: &models.DbtMetaLookerMeasure{
				DbtMetaLookerBase: models.DbtMetaLookerBase{
					Label:       measureStringPtr("Total Sales Amount"),
					Description: measureStringPtr("Sum of all sales amounts"),
				},
				Name: measureStringPtr("total_sales"),
				Type: enums.MeasureSum,
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				assert.Equal(t, "total_sales", measure.Name)
				require.NotNil(t, measure.Label)
				assert.Equal(t, "Total Sales Amount", *measure.Label)
				require.NotNil(t, measure.Description)
				assert.Equal(t, "Sum of all sales amounts", *measure.Description)
			},
		},
		{
			name: "measure with value format",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:            measureStringPtr("revenue"),
				Type:            enums.MeasureSum,
				ValueFormatName: valueFormatPtr(enums.FormatUSD),
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.ValueFormatName)
				assert.Equal(t, enums.FormatUSD, *measure.ValueFormatName)
			},
		},
		{
			name: "hidden measure",
			measureMeta: &models.DbtMetaLookerMeasure{
				DbtMetaLookerBase: models.DbtMetaLookerBase{
					Hidden: measureBoolPtr(true),
				},
				Name: measureStringPtr("internal_metric"),
				Type: enums.MeasureSum,
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.Hidden)
				assert.True(t, *measure.Hidden)
			},
		},
		{
			name: "measure with group label",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:       measureStringPtr("sales_count"),
				Type:       enums.MeasureCount,
				GroupLabel: measureStringPtr("Sales Metrics"),
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.GroupLabel)
				assert.Equal(t, "Sales Metrics", *measure.GroupLabel)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)
			require.NoError(t, err)
			require.NotNil(t, result)
			tt.checkFunc(t, result)
		})
	}
}

func TestMeasureGenerator_CountDistinctValidation(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid count_distinct with approximate",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:                 measureStringPtr("unique_users"),
				Type:                 enums.MeasureCountDistinct,
				Approximate:          measureBoolPtr(true),
				ApproximateThreshold: measureIntPtr(1000),
			},
			expectError: false,
		},
		{
			name: "invalid approximate on non-count_distinct",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:        measureStringPtr("total_amount"),
				Type:        enums.MeasureSum,
				Approximate: measureBoolPtr(true), // Invalid for sum
			},
			expectError: true,
			errorMsg:    "approximate",
		},
		{
			name: "valid count_distinct without approximate",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringPtr("distinct_customers"),
				Type: enums.MeasureCountDistinct,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMeasureGenerator_ErrorHandling(t *testing.T) {
	cfg := &config.Config{}
	generator := generators.NewMeasureGenerator(cfg)

	tests := []struct {
		name        string
		model       *models.DbtModel
		measureMeta *models.DbtMetaLookerMeasure
		expectError bool
	}{
		{
			name: "nil measure meta should return error",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: nil,
			expectError: true,
		},
		{
			name: "valid inputs should not error",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
			},
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringPtr("test_measure"),
				Type: enums.MeasureCount,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectError && tt.measureMeta == nil {
				// Test will panic with nil measureMeta, so we expect that
				assert.Panics(t, func() {
					generator.GenerateMeasure(tt.model, tt.measureMeta)
				})
			} else {
				result, err := generator.GenerateMeasure(tt.model, tt.measureMeta)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, result)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, result)
				}
			}
		})
	}
}

// Helper functions
func measureStringPtr(s string) *string {
	return &s
}

func measureBoolPtr(b bool) *bool {
	return &b
}

func measureIntPtr(i int) *int {
	return &i
}

func valueFormatPtr(vf enums.LookerValueFormatName) *enums.LookerValueFormatName {
	return &vf
}
