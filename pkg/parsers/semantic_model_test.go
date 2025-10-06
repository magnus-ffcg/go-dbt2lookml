package parsers

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSemanticModelParser(t *testing.T) {
	manifest := &models.DbtManifest{
		SemanticModels: map[string]models.DbtSemanticModel{
			"semantic_model.test.orders": {
				Name:  "orders",
				Model: "ref('fact_orders')",
			},
		},
	}

	parser := NewSemanticModelParser(manifest)
	assert.NotNil(t, parser)
	assert.Equal(t, manifest, parser.manifest)
}

func TestGetSemanticModels(t *testing.T) {
	tests := []struct {
		name          string
		manifest      *models.DbtManifest
		expectedCount int
		expectError   bool
	}{
		{
			name: "with semantic models",
			manifest: &models.DbtManifest{
				SemanticModels: map[string]models.DbtSemanticModel{
					"semantic_model.test.orders": {
						Name:  "orders",
						Model: "ref('fact_orders')",
						Measures: []models.DbtSemanticMeasure{
							{
								Name: "total_revenue",
								Agg:  "sum",
								Expr: utils.StringPtr("amount"),
							},
						},
					},
					"semantic_model.test.customers": {
						Name:  "customers",
						Model: "ref('dim_customers')",
					},
				},
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "empty semantic models",
			manifest: &models.DbtManifest{
				SemanticModels: map[string]models.DbtSemanticModel{},
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "nil semantic models",
			manifest: &models.DbtManifest{
				SemanticModels: nil,
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name:          "nil manifest",
			manifest:      nil,
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSemanticModelParser(tt.manifest)
			semanticModels, err := parser.GetSemanticModels()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, semanticModels, tt.expectedCount)
			}
		})
	}
}

func TestGetSemanticModelByName(t *testing.T) {
	manifest := &models.DbtManifest{
		SemanticModels: map[string]models.DbtSemanticModel{
			"semantic_model.test.orders": {
				Name:  "orders",
				Model: "ref('fact_orders')",
			},
			"semantic_model.test.customers": {
				Name:  "customers",
				Model: "ref('dim_customers')",
			},
		},
	}

	tests := []struct {
		name        string
		modelName   string
		expectFound bool
	}{
		{
			name:        "existing model",
			modelName:   "orders",
			expectFound: true,
		},
		{
			name:        "another existing model",
			modelName:   "customers",
			expectFound: true,
		},
		{
			name:        "non-existing model",
			modelName:   "products",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSemanticModelParser(manifest)
			sm, err := parser.GetSemanticModelByName(tt.modelName)

			if tt.expectFound {
				assert.NoError(t, err)
				require.NotNil(t, sm)
				assert.Equal(t, tt.modelName, sm.Name)
			} else {
				assert.Error(t, err)
				assert.Nil(t, sm)
			}
		})
	}
}

func TestGetSemanticModelsForDbtModel(t *testing.T) {
	manifest := &models.DbtManifest{
		SemanticModels: map[string]models.DbtSemanticModel{
			"semantic_model.test.orders": {
				Name:  "orders",
				Model: "ref('fact_orders')",
			},
			"semantic_model.test.order_items": {
				Name:  "order_items",
				Model: "ref('fact_orders')", // Same dbt model
			},
			"semantic_model.test.customers": {
				Name:  "customers",
				Model: "ref('dim_customers')",
			},
		},
	}

	tests := []struct {
		name          string
		dbtModelName  string
		expectedCount int
	}{
		{
			name:          "model with multiple semantic models",
			dbtModelName:  "fact_orders",
			expectedCount: 2,
		},
		{
			name:          "model with one semantic model",
			dbtModelName:  "dim_customers",
			expectedCount: 1,
		},
		{
			name:          "model with no semantic models",
			dbtModelName:  "fact_products",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSemanticModelParser(manifest)
			semanticModels, err := parser.GetSemanticModelsForDbtModel(tt.dbtModelName)

			assert.NoError(t, err)
			assert.Len(t, semanticModels, tt.expectedCount)
		})
	}
}

func TestGetMeasuresForDbtModel(t *testing.T) {
	manifest := &models.DbtManifest{
		SemanticModels: map[string]models.DbtSemanticModel{
			"semantic_model.test.orders": {
				Name:  "orders",
				Model: "ref('fact_orders')",
				Measures: []models.DbtSemanticMeasure{
					{Name: "total_revenue", Agg: "sum", Expr: utils.StringPtr("amount")},
					{Name: "order_count", Agg: "count"},
				},
			},
			"semantic_model.test.order_metrics": {
				Name:  "order_metrics",
				Model: "ref('fact_orders')",
				Measures: []models.DbtSemanticMeasure{
					{Name: "avg_order_value", Agg: "average", Expr: utils.StringPtr("amount")},
				},
			},
		},
	}

	parser := NewSemanticModelParser(manifest)
	measures, err := parser.GetMeasuresForDbtModel("fact_orders")

	assert.NoError(t, err)
	assert.Len(t, measures, 3)

	// Verify measure names
	measureNames := make([]string, len(measures))
	for i, m := range measures {
		measureNames[i] = m.Name
	}
	assert.Contains(t, measureNames, "total_revenue")
	assert.Contains(t, measureNames, "order_count")
	assert.Contains(t, measureNames, "avg_order_value")
}

func TestLinkSemanticModelToModel(t *testing.T) {
	manifest := &models.DbtManifest{
		SemanticModels: map[string]models.DbtSemanticModel{
			"semantic_model.test.orders": {
				Name:  "orders",
				Model: "ref('fact_orders')",
			},
			"semantic_model.test.customers": {
				Name:  "customers",
				Model: "ref('dim_customers')",
			},
			"semantic_model.test.orphan": {
				Name:  "orphan",
				Model: "ref('fact_unknown')", // Not in dbt models list
			},
		},
	}

	dbtModels := []*models.DbtModel{
		{DbtNode: models.DbtNode{Name: "fact_orders"}},
		{DbtNode: models.DbtNode{Name: "dim_customers"}},
	}

	parser := NewSemanticModelParser(manifest)
	modelMap, err := parser.LinkSemanticModelToModel(dbtModels)

	assert.NoError(t, err)
	assert.Len(t, modelMap, 2)
	assert.Contains(t, modelMap, "fact_orders")
	assert.Contains(t, modelMap, "dim_customers")
	assert.NotContains(t, modelMap, "fact_unknown") // Orphan should not be included

	assert.Len(t, modelMap["fact_orders"], 1)
	assert.Len(t, modelMap["dim_customers"], 1)
}

func TestParseRefExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple ref with single quotes",
			input:    "ref('customers')",
			expected: "customers",
		},
		{
			name:     "simple ref with double quotes",
			input:    `ref("customers")`,
			expected: "customers",
		},
		{
			name:     "ref with package - single quotes",
			input:    "ref('my_package', 'customers')",
			expected: "customers",
		},
		{
			name:     "ref with package - double quotes",
			input:    `ref("my_package", "customers")`,
			expected: "customers",
		},
		{
			name:     "ref with whitespace",
			input:    "ref( 'customers' )",
			expected: "customers",
		},
		{
			name:     "not a ref expression",
			input:    "customers",
			expected: "customers",
		},
		{
			name:     "empty ref",
			input:    "ref()",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRefExpression(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasSemanticModels(t *testing.T) {
	tests := []struct {
		name     string
		manifest *models.DbtManifest
		expected bool
	}{
		{
			name: "with semantic models",
			manifest: &models.DbtManifest{
				SemanticModels: map[string]models.DbtSemanticModel{
					"semantic_model.test.orders": {
						Name:  "orders",
						Model: "ref('fact_orders')",
					},
				},
			},
			expected: true,
		},
		{
			name: "empty semantic models map",
			manifest: &models.DbtManifest{
				SemanticModels: map[string]models.DbtSemanticModel{},
			},
			expected: false,
		},
		{
			name: "nil semantic models",
			manifest: &models.DbtManifest{
				SemanticModels: nil,
			},
			expected: false,
		},
		{
			name:     "nil manifest",
			manifest: nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewSemanticModelParser(tt.manifest)
			result := parser.HasSemanticModels()
			assert.Equal(t, tt.expected, result)
		})
	}
}
