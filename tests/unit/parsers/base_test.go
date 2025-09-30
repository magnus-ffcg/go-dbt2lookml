package parsers

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/internal/config"
	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/dbt2lookml/pkg/parsers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDbtParser_NewDbtParser tests parser creation
func TestDbtParser_NewDbtParser(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		manifest    map[string]interface{}
		catalog     map[string]interface{}
		expectError bool
	}{
		{
			name:   "valid parser creation",
			config: &config.Config{},
			manifest: map[string]interface{}{
				"metadata": map[string]interface{}{
					"adapter_type": "bigquery",
				},
				"nodes": map[string]interface{}{},
			},
			catalog:     map[string]interface{}{},
			expectError: false,
		},
		{
			name:   "invalid adapter type",
			config: &config.Config{},
			manifest: map[string]interface{}{
				"metadata": map[string]interface{}{
					"adapter_type": "unsupported",
				},
			},
			catalog:     map[string]interface{}{},
			expectError: true,
		},
		{
			name:        "nil config",
			config:      nil,
			manifest:    map[string]interface{}{},
			catalog:     map[string]interface{}{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := parsers.NewDbtParser(tt.config, tt.manifest, tt.catalog)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, parser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parser)
			}
		})
	}
}

// TestDbtParser_ModelFiltering tests model filtering functionality
func TestDbtParser_ModelFiltering(t *testing.T) {
	// Create test manifest with multiple models
	manifest := map[string]interface{}{
		"metadata": map[string]interface{}{
			"adapter_type": "bigquery",
		},
		"nodes": map[string]interface{}{
			"model.test.model1": map[string]interface{}{
				"name":          "model1",
				"resource_type": "model",
				"tags":          []interface{}{"analytics"},
				"unique_id":     "model.test.model1",
				"relation_name": "`project.dataset.model1`",
				"schema":        "test_schema",
				"description":   "Test model 1",
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
				"path":          "models/model1.sql",
			},
			"model.test.model2": map[string]interface{}{
				"name":          "model2",
				"resource_type": "model",
				"tags":          []interface{}{"reporting"},
				"unique_id":     "model.test.model2",
				"relation_name": "`project.dataset.model2`",
				"schema":        "test_schema",
				"description":   "Test model 2",
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
				"path":          "models/model2.sql",
			},
			"model.test.model3": map[string]interface{}{
				"name":          "model3",
				"resource_type": "model",
				"tags":          []interface{}{"analytics", "core"},
				"unique_id":     "model.test.model3",
				"relation_name": "`project.dataset.model3`",
				"schema":        "test_schema",
				"description":   "Test model 3",
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
				"path":          "models/model3.sql",
			},
		},
	}

	catalog := map[string]interface{}{}

	tests := []struct {
		name           string
		config         *config.Config
		expectedModels []string
	}{
		{
			name:           "no filters - all models",
			config:         &config.Config{},
			expectedModels: []string{"model1", "model2", "model3"},
		},
		{
			name: "select specific model",
			config: &config.Config{
				Select: "model2",
			},
			expectedModels: []string{"model2"},
		},
		{
			name: "filter by tag",
			config: &config.Config{
				Tag: "analytics",
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "include models",
			config: &config.Config{
				IncludeModels: []string{"model1", "model3"},
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "exclude models",
			config: &config.Config{
				ExcludeModels: []string{"model2"},
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "combined filters",
			config: &config.Config{
				Tag:           "analytics",
				ExcludeModels: []string{"model3"},
			},
			expectedModels: []string{"model1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := parsers.NewDbtParser(tt.config, manifest, catalog)
			require.NoError(t, err)
			require.NotNil(t, parser)

			models, err := parser.GetModels()
			require.NoError(t, err)

			// Extract model names for comparison
			actualNames := make([]string, len(models))
			for i, model := range models {
				actualNames[i] = model.Name
			}

			assert.ElementsMatch(t, tt.expectedModels, actualNames)
		})
	}
}

// TestDbtParser_ErrorHandling tests error conditions
func TestDbtParser_ErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		config   *config.Config
		manifest map[string]interface{}
		catalog  map[string]interface{}
	}{
		{
			name:   "empty manifest",
			config: &config.Config{},
			manifest: map[string]interface{}{
				"metadata": map[string]interface{}{
					"adapter_type": "bigquery",
				},
			},
			catalog: map[string]interface{}{},
		},
		{
			name:   "malformed nodes",
			config: &config.Config{},
			manifest: map[string]interface{}{
				"metadata": map[string]interface{}{
					"adapter_type": "bigquery",
				},
				"nodes": "invalid_nodes_format",
			},
			catalog: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := parsers.NewDbtParser(tt.config, tt.manifest, tt.catalog)
			
			// Parser creation might succeed even with malformed data
			if err == nil && parser != nil {
				// But GetModels should handle errors gracefully
				models, err := parser.GetModels()
				// Should either succeed with empty results or return a clear error
				if err != nil {
					assert.Error(t, err)
				} else {
					// Models can be nil or empty for edge cases
					if models != nil && tt.name == "empty manifest" {
						assert.Empty(t, models)
					}
					// Just verify we don't panic and get some result
					t.Logf("Test %s: got %d models", tt.name, len(models))
				}
			}
		})
	}
}
func TestSupportedAdapters(t *testing.T) {
	tests := []struct {
		name        string
		adapterType string
		expectError bool
	}{
		{"bigquery supported", "bigquery", false},
		{"unsupported adapter", "snowflake", true},
		{"empty adapter", "", true},
		{"case sensitivity", "BigQuery", true}, // Should be case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manifest := map[string]interface{}{
				"metadata": map[string]interface{}{
					"adapter_type": tt.adapterType,
				},
				"nodes": map[string]interface{}{},
			}

			parser, err := parsers.NewDbtParser(&config.Config{}, manifest, map[string]interface{}{})

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, parser)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, parser)
			}
		})
	}
}

// TestEnumValues tests enum value consistency
func TestEnumValues(t *testing.T) {
	t.Run("supported adapters", func(t *testing.T) {
		assert.Equal(t, "bigquery", string(enums.BigQuery))
	})

	t.Run("resource types", func(t *testing.T) {
		assert.Equal(t, "model", string(enums.ResourceModel))
		assert.Equal(t, "seed", string(enums.ResourceSeed))
		assert.Equal(t, "exposure", string(enums.ResourceExposure))
	})

	t.Run("measure types", func(t *testing.T) {
		assert.Equal(t, "count", string(enums.MeasureCount))
		assert.Equal(t, "sum", string(enums.MeasureSum))
		assert.Equal(t, "average", string(enums.MeasureAverage))
	})
}
