package parsers

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelParser_FilterModels tests the FilterModels function directly
func TestModelParser_FilterModels(t *testing.T) {
	manifest := &models.DbtManifest{
		Nodes: map[string]interface{}{},
	}
	parser := NewModelParser(manifest, &config.Config{})

	// Create test models
	models := []*models.DbtModel{
		{
			DbtNode: models.DbtNode{
				Name: "model1",
			},
			Tags: []string{"analytics", "core"},
		},
		{
			DbtNode: models.DbtNode{
				Name: "model2",
			},
			Tags: []string{"reporting"},
		},
		{
			DbtNode: models.DbtNode{
				Name: "model3",
			},
			Tags: []string{"analytics", "staging"},
		},
		{
			DbtNode: models.DbtNode{
				Name: "staging_model",
			},
			Tags: []string{"staging"},
		},
	}

	tests := []struct {
		name           string
		options        ModelFilterOptions
		expectedModels []string
	}{
		{
			name:           "no filters - return all",
			options:        ModelFilterOptions{},
			expectedModels: []string{"model1", "model2", "model3", "staging_model"},
		},
		{
			name: "select single model",
			options: ModelFilterOptions{
				SelectModel: "model2",
			},
			expectedModels: []string{"model2"},
		},
		{
			name: "filter by tag",
			options: ModelFilterOptions{
				Tag: "analytics",
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "include specific models",
			options: ModelFilterOptions{
				IncludeModels: []string{"model1", "model3"},
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "exclude specific models",
			options: ModelFilterOptions{
				ExcludeModels: []string{"staging_model", "model3"},
			},
			expectedModels: []string{"model1", "model2"},
		},
		{
			name: "combined: tag filter + exclude",
			options: ModelFilterOptions{
				Tag:           "analytics",
				ExcludeModels: []string{"model3"},
			},
			expectedModels: []string{"model1"},
		},
		{
			name: "combined: include + exclude (exclude wins)",
			options: ModelFilterOptions{
				IncludeModels: []string{"model1", "model2", "model3"},
				ExcludeModels: []string{"model2"},
			},
			expectedModels: []string{"model1", "model3"},
		},
		{
			name: "select model overrides all other filters",
			options: ModelFilterOptions{
				SelectModel:   "model2",
				Tag:           "analytics",
				IncludeModels: []string{"model1"},
				ExcludeModels: []string{"model2"},
			},
			expectedModels: []string{"model2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := parser.FilterModels(models, tt.options)

			actualNames := make([]string, len(filtered))
			for i, model := range filtered {
				actualNames[i] = model.Name
			}

			assert.ElementsMatch(t, tt.expectedModels, actualNames)
		})
	}
}

// TestModelParser_TagMatching tests tag matching logic
func TestModelParser_TagMatching(t *testing.T) {
	tests := []struct {
		name        string
		searchTag   string
		modelTags   []string
		shouldMatch bool
	}{
		{
			name:        "exact match",
			searchTag:   "analytics",
			modelTags:   []string{"analytics", "core"},
			shouldMatch: true,
		},
		{
			name:        "case insensitive match",
			searchTag:   "Analytics",
			modelTags:   []string{"analytics", "core"},
			shouldMatch: true,
		},
		{
			name:        "no match",
			searchTag:   "reporting",
			modelTags:   []string{"analytics", "core"},
			shouldMatch: false,
		},
		{
			name:        "empty tags",
			searchTag:   "analytics",
			modelTags:   []string{},
			shouldMatch: false,
		},
		{
			name:        "single tag match",
			searchTag:   "staging",
			modelTags:   []string{"staging"},
			shouldMatch: true,
		},
	}

	manifest := &models.DbtManifest{
		Nodes: map[string]interface{}{},
	}
	parser := NewModelParser(manifest, &config.Config{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelsList := []*models.DbtModel{
				{
					DbtNode: models.DbtNode{
						Name: "test_model",
					},
					Tags: tt.modelTags,
				},
			}

			options := ModelFilterOptions{
				Tag: tt.searchTag,
			}

			filtered := parser.FilterModels(modelsList, options)

			if tt.shouldMatch {
				assert.Len(t, filtered, 1, "Should match model with tag")
			} else {
				assert.Empty(t, filtered, "Should not match model")
			}
		})
	}
}

// TestModelParser_GetModelByName tests getting a model by name
func TestModelParser_GetModelByName(t *testing.T) {
	manifest := &models.DbtManifest{
		Nodes: map[string]interface{}{
			"model.test.model1": map[string]interface{}{
				"name":          "model1",
				"resource_type": "model",
				"unique_id":     "model.test.model1",
				"relation_name": "`project.dataset.model1`",
				"schema":        "test_schema",
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
			},
			"model.test.model2": map[string]interface{}{
				"name":          "model2",
				"resource_type": "model",
				"unique_id":     "model.test.model2",
				"relation_name": "`project.dataset.model2`",
				"schema":        "test_schema",
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
			},
		},
	}

	parser := NewModelParser(manifest, &config.Config{})

	tests := []struct {
		name        string
		modelName   string
		expectFound bool
	}{
		{
			name:        "existing model",
			modelName:   "model1",
			expectFound: true,
		},
		{
			name:        "another existing model",
			modelName:   "model2",
			expectFound: true,
		},
		{
			name:        "non-existent model",
			modelName:   "model3",
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := parser.GetModelByName(tt.modelName)

			if tt.expectFound {
				assert.NoError(t, err)
				require.NotNil(t, model)
				assert.Equal(t, tt.modelName, model.Name)
			} else {
				assert.Error(t, err)
				assert.Nil(t, model)
			}
		})
	}
}

// TestModelParser_GetModelsByTag tests getting models by tag
func TestModelParser_GetModelsByTag(t *testing.T) {
	manifest := &models.DbtManifest{
		Nodes: map[string]interface{}{
			"model.test.model1": map[string]interface{}{
				"name":          "model1",
				"resource_type": "model",
				"unique_id":     "model.test.model1",
				"relation_name": "`project.dataset.model1`",
				"schema":        "test_schema",
				"tags":          []interface{}{"analytics", "core"},
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
			},
			"model.test.model2": map[string]interface{}{
				"name":          "model2",
				"resource_type": "model",
				"unique_id":     "model.test.model2",
				"relation_name": "`project.dataset.model2`",
				"schema":        "test_schema",
				"tags":          []interface{}{"reporting"},
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
			},
			"model.test.model3": map[string]interface{}{
				"name":          "model3",
				"resource_type": "model",
				"unique_id":     "model.test.model3",
				"relation_name": "`project.dataset.model3`",
				"schema":        "test_schema",
				"tags":          []interface{}{"analytics"},
				"columns":       map[string]interface{}{},
				"meta":          map[string]interface{}{},
			},
		},
	}

	parser := NewModelParser(manifest, &config.Config{})

	tests := []struct {
		name           string
		tag            string
		expectedModels []string
	}{
		{
			name:           "analytics tag",
			tag:            "analytics",
			expectedModels: []string{"model1", "model3"},
		},
		{
			name:           "reporting tag",
			tag:            "reporting",
			expectedModels: []string{"model2"},
		},
		{
			name:           "core tag",
			tag:            "core",
			expectedModels: []string{"model1"},
		},
		{
			name:           "non-existent tag",
			tag:            "staging",
			expectedModels: []string{},
		},
		{
			name:           "case insensitive",
			tag:            "Analytics",
			expectedModels: []string{"model1", "model3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models, err := parser.GetModelsByTag(tt.tag)
			assert.NoError(t, err)

			actualNames := make([]string, len(models))
			for i, model := range models {
				actualNames[i] = model.Name
			}

			assert.ElementsMatch(t, tt.expectedModels, actualNames)
		})
	}
}

// TestModelParser_EdgeCases tests edge cases in model parsing
func TestModelParser_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		manifest *models.DbtManifest
	}{
		{
			name: "empty manifest",
			manifest: &models.DbtManifest{
				Nodes: map[string]interface{}{},
			},
		},
		{
			name: "manifest with no models (only seeds)",
			manifest: &models.DbtManifest{
				Nodes: map[string]interface{}{
					"seed.test.seed1": map[string]interface{}{
						"name":          "seed1",
						"resource_type": "seed",
						"unique_id":     "seed.test.seed1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewModelParser(tt.manifest, &config.Config{})

			models, err := parser.GetAllModels()
			assert.NoError(t, err)
			assert.Empty(t, models, "Should return empty list for no models")
		})
	}
}
