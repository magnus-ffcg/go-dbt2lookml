package generators

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMeasureGenerator_Filters tests measure filters functionality
func TestMeasureGenerator_Filters(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "sales",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		checkFunc   func(*testing.T, *models.LookMLMeasure)
	}{
		{
			name: "measure with single filter",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringTestPtr("active_sales_count"),
				Type: enums.MeasureCount,
				Filters: []models.DbtMetaLookerMeasureFilter{
					{
						FilterDimension:  "status",
						FilterExpression: "active",
					},
				},
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				assert.Equal(t, "active_sales_count", measure.Name)
				assert.Len(t, measure.Filters, 1)
				assert.Equal(t, "status", measure.Filters[0].FilterDimension)
				assert.Equal(t, "active", measure.Filters[0].FilterExpression)
			},
		},
		{
			name: "measure with multiple filters",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringTestPtr("premium_sales_total"),
				Type: enums.MeasureSum,
				Filters: []models.DbtMetaLookerMeasureFilter{
					{
						FilterDimension:  "status",
						FilterExpression: "completed",
					},
					{
						FilterDimension:  "tier",
						FilterExpression: "premium",
					},
				},
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				assert.Len(t, measure.Filters, 2)
				assert.Equal(t, "status", measure.Filters[0].FilterDimension)
				assert.Equal(t, "completed", measure.Filters[0].FilterExpression)
				assert.Equal(t, "tier", measure.Filters[1].FilterDimension)
				assert.Equal(t, "premium", measure.Filters[1].FilterExpression)
			},
		},
		{
			name: "measure without filters",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:    measureStringTestPtr("all_sales_count"),
				Type:    enums.MeasureCount,
				Filters: nil,
			},
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				assert.Nil(t, measure.Filters)
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

// TestMeasureGenerator_Precision tests precision attribute for sum/average measures
func TestMeasureGenerator_Precision(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "sales",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		expectError bool
		checkFunc   func(*testing.T, *models.LookMLMeasure)
	}{
		{
			name: "sum measure with precision",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:      measureStringTestPtr("total_amount"),
				Type:      enums.MeasureSum,
				Precision: measureIntTestPtr(2),
			},
			expectError: false,
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.Precision)
				assert.Equal(t, 2, *measure.Precision)
			},
		},
		{
			name: "average measure with precision",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:      measureStringTestPtr("avg_price"),
				Type:      enums.MeasureAverage,
				Precision: measureIntTestPtr(3),
			},
			expectError: false,
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.Precision)
				assert.Equal(t, 3, *measure.Precision)
			},
		},
		{
			name: "count measure with precision should error",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:      measureStringTestPtr("count_with_precision"),
				Type:      enums.MeasureCount,
				Precision: measureIntTestPtr(2),
			},
			expectError: true,
		},
		{
			name: "count_distinct with precision should error",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:      measureStringTestPtr("unique_users"),
				Type:      enums.MeasureCountDistinct,
				Precision: measureIntTestPtr(2),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestMeasureGenerator_SQLDistinctKey tests sql_distinct_key for count_distinct measures
func TestMeasureGenerator_SQLDistinctKey(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		expectError bool
		checkFunc   func(*testing.T, *models.LookMLMeasure)
	}{
		{
			name: "count_distinct with sql_distinct_key",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:           measureStringTestPtr("unique_customers"),
				Type:           enums.MeasureCountDistinct,
				SQLDistinctKey: measureStringTestPtr("${TABLE}.customer_id"),
			},
			expectError: false,
			checkFunc: func(t *testing.T, measure *models.LookMLMeasure) {
				require.NotNil(t, measure.SQLDistinctKey)
				assert.Equal(t, "${TABLE}.customer_id", *measure.SQLDistinctKey)
			},
		},
		{
			name: "sum with sql_distinct_key should error",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:           measureStringTestPtr("total_amount"),
				Type:           enums.MeasureSum,
				SQLDistinctKey: measureStringTestPtr("${TABLE}.id"),
			},
			expectError: true,
		},
		{
			name: "average with sql_distinct_key should error",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:           measureStringTestPtr("avg_price"),
				Type:           enums.MeasureAverage,
				SQLDistinctKey: measureStringTestPtr("${TABLE}.id"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestMeasureGenerator_MultipleMeasuresIntegration tests generating multiple measures
func TestMeasureGenerator_MultipleMeasuresIntegration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "sales",
		},
		Meta: &models.DbtModelMeta{
			Looker: &models.DbtMetaLooker{
				Measures: []models.DbtMetaLookerMeasure{
					{
						Name: measureStringTestPtr("total_revenue"),
						Type: enums.MeasureSum,
						DbtMetaLookerBase: models.DbtMetaLookerBase{
							Label:       measureStringTestPtr("Total Revenue"),
							Description: measureStringTestPtr("Sum of all sales revenue"),
						},
						ValueFormatName: valueFormatNamePtr(enums.FormatUSD),
						Precision:       measureIntTestPtr(2),
					},
					{
						Name: measureStringTestPtr("avg_order_value"),
						Type: enums.MeasureAverage,
						DbtMetaLookerBase: models.DbtMetaLookerBase{
							Label: measureStringTestPtr("Average Order Value"),
						},
						ValueFormatName: valueFormatNamePtr(enums.FormatUSD),
						Precision:       measureIntTestPtr(2),
					},
					{
						Name: measureStringTestPtr("unique_customers"),
						Type: enums.MeasureCountDistinct,
						DbtMetaLookerBase: models.DbtMetaLookerBase{
							Label: measureStringTestPtr("Unique Customers"),
						},
						Approximate:          measureBoolTestPtr(true),
						ApproximateThreshold: measureIntTestPtr(10000),
					},
					{
						Name: measureStringTestPtr("active_orders"),
						Type: enums.MeasureCount,
						Filters: []models.DbtMetaLookerMeasureFilter{
							{
								FilterDimension:  "status",
								FilterExpression: "active",
							},
						},
					},
				},
			},
		},
	}

	t.Run("generate all measures from model meta", func(t *testing.T) {
		var measures []*models.LookMLMeasure

		for _, measureMeta := range model.Meta.Looker.Measures {
			measure, err := generator.GenerateMeasure(model, &measureMeta)
			require.NoError(t, err)
			require.NotNil(t, measure)
			measures = append(measures, measure)
		}

		assert.Len(t, measures, 4)

		// Check total_revenue
		assert.Equal(t, "total_revenue", measures[0].Name)
		assert.Equal(t, enums.MeasureSum, measures[0].Type)
		require.NotNil(t, measures[0].Label)
		assert.Equal(t, "Total Revenue", *measures[0].Label)
		require.NotNil(t, measures[0].Precision)
		assert.Equal(t, 2, *measures[0].Precision)

		// Check avg_order_value
		assert.Equal(t, "avg_order_value", measures[1].Name)
		assert.Equal(t, enums.MeasureAverage, measures[1].Type)

		// Check unique_customers
		assert.Equal(t, "unique_customers", measures[2].Name)
		assert.Equal(t, enums.MeasureCountDistinct, measures[2].Type)
		require.NotNil(t, measures[2].Approximate)
		assert.True(t, *measures[2].Approximate)

		// Check active_orders
		assert.Equal(t, "active_orders", measures[3].Name)
		assert.Len(t, measures[3].Filters, 1)
	})

	t.Run("default count measure with existing measures", func(t *testing.T) {
		// Should still generate default count if no explicit count exists
		countMeasure := generator.GenerateDefaultCountMeasure(model)
		require.NotNil(t, countMeasure)
		assert.Equal(t, "count", countMeasure.Name)
		assert.Equal(t, enums.MeasureCount, countMeasure.Type)
	})
}

// TestMeasureGenerator_PercentileMeasures tests percentile measure type validation
func TestMeasureGenerator_PercentileMeasures(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "test_model",
		},
	}

	tests := []struct {
		name        string
		measureMeta *models.DbtMetaLookerMeasure
		expectError bool
	}{
		{
			name: "percentile measure with percentile attribute",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:       measureStringTestPtr("p95_response_time"),
				Type:       enums.LookerMeasureType("percentile_95"), // Must be at least 10 chars starting with "percentile"
				Percentile: measureIntTestPtr(95),
			},
			expectError: false,
		},
		{
			name: "non-percentile measure with percentile attribute should error",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name:       measureStringTestPtr("total_amount"),
				Type:       enums.MeasureSum,
				Percentile: measureIntTestPtr(50),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
			}
		})
	}
}

// TestMeasureGenerator_MeasureNameGeneration tests measure name fallback logic
func TestMeasureGenerator_MeasureNameGeneration(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: "orders",
		},
	}

	tests := []struct {
		name         string
		measureMeta  *models.DbtMetaLookerMeasure
		expectedName string
	}{
		{
			name: "explicit name is used",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: measureStringTestPtr("custom_measure_name"),
				Type: enums.MeasureSum,
			},
			expectedName: "custom_measure_name",
		},
		{
			name: "fallback to measure type when name is nil",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: nil,
				Type: enums.MeasureAverage,
			},
			expectedName: "average",
		},
		{
			name: "count type fallback",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: nil,
				Type: enums.MeasureCount,
			},
			expectedName: "count",
		},
		{
			name: "count_distinct type fallback",
			measureMeta: &models.DbtMetaLookerMeasure{
				Name: nil,
				Type: enums.MeasureCountDistinct,
			},
			expectedName: "count_distinct",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := generator.GenerateMeasure(model, tt.measureMeta)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.expectedName, result.Name)
		})
	}
}

// Helper functions specific to measure integration tests
func measureIntTestPtr(i int) *int {
	return &i
}

func measureStringTestPtr(s string) *string {
	return &s
}

func measureBoolTestPtr(b bool) *bool {
	return &b
}

func valueFormatNamePtr(vf enums.LookerValueFormatName) *enums.LookerValueFormatName {
	return &vf
}
