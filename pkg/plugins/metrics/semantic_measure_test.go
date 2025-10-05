package metrics

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSemanticMeasureGenerator(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	assert.NotNil(t, gen)
	assert.Equal(t, cfg, gen.config)
}

func TestMapAggregationType(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	tests := []struct {
		name        string
		agg         string
		expected    enums.LookerMeasureType
		expectError bool
	}{
		{"sum", "sum", enums.MeasureSum, false},
		{"average", "average", enums.MeasureAverage, false},
		{"avg alias", "avg", enums.MeasureAverage, false},
		{"min", "min", enums.MeasureMin, false},
		{"max", "max", enums.MeasureMax, false},
		{"count", "count", enums.MeasureCount, false},
		{"count_distinct", "count_distinct", enums.MeasureCountDistinct, false},
		{"median", "median", enums.MeasureMedian, false},
		{"percentile", "percentile", enums.MeasureNumber, false},
		{"sum_boolean", "sum_boolean", enums.MeasureSum, false},
		{"case insensitive SUM", "SUM", enums.MeasureSum, false},
		{"case insensitive Average", "Average", enums.MeasureAverage, false},
		{"unsupported type", "variance", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.mapAggregationType(tt.agg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGenerateMeasureFromSemantic(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{Name: "test_model"},
	}

	tests := []struct {
		name            string
		semanticMeasure *models.DbtSemanticMeasure
		expectedName    string
		expectedType    enums.LookerMeasureType
		expectError     bool
	}{
		{
			name: "simple sum measure",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name: "total_revenue",
				Agg:  "sum",
				Expr: utils.StringPtr("amount"),
			},
			expectedName: "total_revenue",
			expectedType: enums.MeasureSum,
			expectError:  false,
		},
		{
			name: "average measure with label",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name:  "avg_order_value",
				Agg:   "average",
				Expr:  utils.StringPtr("order_total"),
				Label: utils.StringPtr("Average Order Value"),
			},
			expectedName: "avg_order_value",
			expectedType: enums.MeasureAverage,
			expectError:  false,
		},
		{
			name: "count_distinct measure",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name:        "unique_customers",
				Agg:         "count_distinct",
				Expr:        utils.StringPtr("customer_id"),
				Description: utils.StringPtr("Number of unique customers"),
			},
			expectedName: "unique_customers",
			expectedType: enums.MeasureCountDistinct,
			expectError:  false,
		},
		{
			name: "count measure without expr",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name: "order_count",
				Agg:  "count",
			},
			expectedName: "order_count",
			expectedType: enums.MeasureCount,
			expectError:  false,
		},
		{
			name: "median measure",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name: "median_price",
				Agg:  "median",
				Expr: utils.StringPtr("price"),
			},
			expectedName: "median_price",
			expectedType: enums.MeasureMedian,
			expectError:  false,
		},
		{
			name: "percentile measure with params",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name: "p95_latency",
				Agg:  "percentile",
				Expr: utils.StringPtr("latency_ms"),
				AggParams: &models.DbtSemanticMeasureAggParams{
					Percentile: utils.Float64Ptr(0.95),
				},
			},
			expectedName: "p95_latency",
			expectedType: enums.MeasureNumber,
			expectError:  false,
		},
		{
			name: "sum_boolean measure",
			semanticMeasure: &models.DbtSemanticMeasure{
				Name: "active_count",
				Agg:  "sum_boolean",
				Expr: utils.StringPtr("is_active"),
			},
			expectedName: "active_count",
			expectedType: enums.MeasureSum,
			expectError:  false,
		},
		{
			name:            "nil measure",
			semanticMeasure: nil,
			expectError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.GenerateMeasureFromSemantic(tt.semanticMeasure, model)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedName, result.Name)
				assert.Equal(t, tt.expectedType, result.Type)

				// Check label if provided
				if tt.semanticMeasure.Label != nil {
					assert.Equal(t, tt.semanticMeasure.Label, result.Label)
				}

				// Check description if provided
				if tt.semanticMeasure.Description != nil {
					assert.Equal(t, tt.semanticMeasure.Description, result.Description)
				}

				// Check percentile if this is a percentile measure
				if tt.semanticMeasure.IsPercentile() && tt.semanticMeasure.AggParams != nil && tt.semanticMeasure.AggParams.Percentile != nil {
					// Percentile should be converted from float (0.95) to int (95)
					expectedPercentile := int(*tt.semanticMeasure.AggParams.Percentile * 100)
					require.NotNil(t, result.Percentile)
					assert.Equal(t, expectedPercentile, *result.Percentile)
				}
			}
		})
	}
}

func TestGetMeasureSQL(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{Name: "test_model"},
	}

	tests := []struct {
		name        string
		measure     *models.DbtSemanticMeasure
		expectedSQL *string
	}{
		{
			name: "count measure - no SQL needed",
			measure: &models.DbtSemanticMeasure{
				Name: "order_count",
				Agg:  "count",
			},
			expectedSQL: nil,
		},
		{
			name: "sum with expr",
			measure: &models.DbtSemanticMeasure{
				Name: "total_revenue",
				Agg:  "sum",
				Expr: utils.StringPtr("amount"),
			},
			expectedSQL: utils.StringPtr("${TABLE}.amount"),
		},
		{
			name: "sum_boolean with expr",
			measure: &models.DbtSemanticMeasure{
				Name: "active_count",
				Agg:  "sum_boolean",
				Expr: utils.StringPtr("is_active"),
			},
			expectedSQL: utils.StringPtr("CAST(${TABLE}.is_active AS INT64)"),
		},
		{
			name: "average without expr",
			measure: &models.DbtSemanticMeasure{
				Name: "avg_price",
				Agg:  "average",
			},
			expectedSQL: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.getMeasureSQL(tt.measure, model)

			if tt.expectedSQL == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, *tt.expectedSQL, *result)
			}
		})
	}
}

func TestBuildSQLFromExpr(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	tests := []struct {
		name     string
		expr     string
		agg      string
		expected string
	}{
		{
			name:     "simple column reference",
			expr:     "amount",
			agg:      "sum",
			expected: "${TABLE}.amount",
		},
		{
			name:     "column with uppercase",
			expr:     "TotalAmount",
			agg:      "sum",
			expected: "${TABLE}.totalamount",
		},
		{
			name:     "sum_boolean column",
			expr:     "is_active",
			agg:      "sum_boolean",
			expected: "CAST(${TABLE}.is_active AS INT64)",
		},
		{
			name:     "already a LookML reference",
			expr:     "${dimension_name}",
			agg:      "sum",
			expected: "${dimension_name}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.buildSQLFromExpr(tt.expr, tt.agg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateMeasuresFromSemantic(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	model := &models.DbtModel{
		DbtNode: models.DbtNode{Name: "orders"},
	}

	tests := []struct {
		name          string
		measures      []models.DbtSemanticMeasure
		expectedCount int
	}{
		{
			name: "multiple valid measures",
			measures: []models.DbtSemanticMeasure{
				{Name: "total_revenue", Agg: "sum", Expr: utils.StringPtr("amount")},
				{Name: "order_count", Agg: "count"},
				{Name: "avg_order_value", Agg: "average", Expr: utils.StringPtr("amount")},
			},
			expectedCount: 3,
		},
		{
			name: "generate all measures regardless of create_metric flag",
			measures: []models.DbtSemanticMeasure{
				{Name: "total_revenue", Agg: "sum", Expr: utils.StringPtr("amount")},
				{Name: "internal_metric", Agg: "count", CreateMetric: utils.BoolPtr(false)},
			},
			// We generate LookML measures for ALL semantic measures (create_metric is for dbt metrics)
			expectedCount: 2,
		},
		{
			name:          "empty measures",
			measures:      []models.DbtSemanticMeasure{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.GenerateMeasuresFromSemantic(tt.measures, model)

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)
		})
	}
}

func TestHasSemanticMeasureWithName(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	measures := []models.DbtSemanticMeasure{
		{Name: "total_revenue", Agg: "sum"},
		{Name: "order_count", Agg: "count"},
	}

	tests := []struct {
		name       string
		searchName string
		expected   bool
	}{
		{"existing measure", "total_revenue", true},
		{"another existing measure", "order_count", true},
		{"non-existing measure", "avg_price", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.HasSemanticMeasureWithName(measures, tt.searchName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMergeWithMetaMeasures_WithConflict(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	semanticMeasures := []*models.LookMLMeasure{
		{Name: "total_revenue", Type: enums.MeasureSum},
		{Name: "order_count", Type: enums.MeasureCount},
	}

	metaMeasures := []*models.LookMLMeasure{
		{Name: "total_revenue", Type: enums.MeasureSum}, // Conflict - semantic wins
		{Name: "custom_measure", Type: enums.MeasureAverage},
	}

	result := gen.MergeWithMetaMeasures(semanticMeasures, metaMeasures)

	// Should have 3 measures: 2 semantic + 1 non-conflicting meta
	assert.Len(t, result, 3)

	// Check that all expected measures are present
	names := make([]string, len(result))
	for i, m := range result {
		names[i] = m.Name
	}
	assert.Contains(t, names, "total_revenue")
	assert.Contains(t, names, "order_count")
	assert.Contains(t, names, "custom_measure")

	// Verify semantic measure comes first (takes precedence)
	assert.Equal(t, "total_revenue", result[0].Name)
	assert.Equal(t, enums.MeasureSum, result[0].Type)
}

func TestMergeWithMetaMeasures_NoConflicts(t *testing.T) {
	cfg := &config.Config{}
	gen := NewSemanticMeasureGenerator(cfg)

	semanticMeasures := []*models.LookMLMeasure{
		{Name: "total_revenue", Type: enums.MeasureSum},
	}

	metaMeasures := []*models.LookMLMeasure{
		{Name: "custom_measure", Type: enums.MeasureAverage},
	}

	result := gen.MergeWithMetaMeasures(semanticMeasures, metaMeasures)

	// Should have both measures
	assert.Len(t, result, 2)

	names := make([]string, len(result))
	for i, m := range result {
		names[i] = m.Name
	}
	assert.Contains(t, names, "total_revenue")
	assert.Contains(t, names, "custom_measure")
}
