package parsers

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// SemanticModelParser handles parsing of dbt semantic models from manifest
type SemanticModelParser struct {
	manifest *models.DbtManifest
}

// NewSemanticModelParser creates a new SemanticModelParser instance
func NewSemanticModelParser(manifest *models.DbtManifest) *SemanticModelParser {
	return &SemanticModelParser{
		manifest: manifest,
	}
}

// GetSemanticModels returns all semantic models from the manifest
func (p *SemanticModelParser) GetSemanticModels() ([]models.DbtSemanticModel, error) {
	if p.manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	if p.manifest.SemanticModels == nil {
		// No semantic models in manifest - this is not an error
		return []models.DbtSemanticModel{}, nil
	}

	semanticModels := make([]models.DbtSemanticModel, 0, len(p.manifest.SemanticModels))
	for _, sm := range p.manifest.SemanticModels {
		semanticModels = append(semanticModels, sm)
	}

	return semanticModels, nil
}

// GetSemanticModelByName returns a specific semantic model by name
func (p *SemanticModelParser) GetSemanticModelByName(name string) (*models.DbtSemanticModel, error) {
	if p.manifest == nil || p.manifest.SemanticModels == nil {
		return nil, fmt.Errorf("no semantic models available")
	}

	// Search by name in the map values
	for _, sm := range p.manifest.SemanticModels {
		if sm.Name == name {
			return &sm, nil
		}
	}

	return nil, fmt.Errorf("semantic model %s not found", name)
}

// GetSemanticModelsForDbtModel returns semantic models that reference a specific dbt model
// The modelName should match what's inside the ref() call
func (p *SemanticModelParser) GetSemanticModelsForDbtModel(modelName string) ([]models.DbtSemanticModel, error) {
	allModels, err := p.GetSemanticModels()
	if err != nil {
		return nil, err
	}

	var matchingModels []models.DbtSemanticModel
	for _, sm := range allModels {
		// Extract model name from ref() expression
		refModelName := sm.GetModelRef()
		if refModelName == modelName {
			matchingModels = append(matchingModels, sm)
		}
	}

	return matchingModels, nil
}

// GetMeasuresForDbtModel returns all measures from semantic models that reference a dbt model
func (p *SemanticModelParser) GetMeasuresForDbtModel(modelName string) ([]models.DbtSemanticMeasure, error) {
	semanticModels, err := p.GetSemanticModelsForDbtModel(modelName)
	if err != nil {
		return nil, err
	}

	var allMeasures []models.DbtSemanticMeasure
	for _, sm := range semanticModels {
		allMeasures = append(allMeasures, sm.Measures...)
	}

	return allMeasures, nil
}

// LinkSemanticModelToModel creates a mapping between dbt model name and its semantic models
func (p *SemanticModelParser) LinkSemanticModelToModel(dbtModels []*models.DbtModel) (map[string][]models.DbtSemanticModel, error) {
	modelMap := make(map[string][]models.DbtSemanticModel)

	allSemanticModels, err := p.GetSemanticModels()
	if err != nil {
		return nil, err
	}

	// Create a map of dbt model names for quick lookup
	dbtModelNames := make(map[string]bool)
	for _, model := range dbtModels {
		dbtModelNames[model.Name] = true
	}

	// Link semantic models to dbt models
	for _, sm := range allSemanticModels {
		refModelName := sm.GetModelRef()

		// Only include if the dbt model exists in our list
		if dbtModelNames[refModelName] {
			modelMap[refModelName] = append(modelMap[refModelName], sm)
		}
	}

	return modelMap, nil
}

// ParseRefExpression extracts model name from ref() expressions
// Supports formats:
//   - ref('model_name')
//   - ref("model_name")
//   - ref('package', 'model_name')
//   - ref("package", "model_name")
func ParseRefExpression(refExpr string) string {
	refExpr = strings.TrimSpace(refExpr)

	// Remove "ref(" prefix and ")" suffix
	if !strings.HasPrefix(refExpr, "ref(") || !strings.HasSuffix(refExpr, ")") {
		return refExpr // Not a valid ref expression, return as-is
	}

	// Extract content between ref( and )
	content := refExpr[4 : len(refExpr)-1]
	content = strings.TrimSpace(content)

	// Split by comma for multi-argument refs
	parts := strings.Split(content, ",")

	// Get the last part (model name) and remove quotes
	modelPart := strings.TrimSpace(parts[len(parts)-1])
	modelPart = strings.Trim(modelPart, "'\"")

	return modelPart
}

// HasSemanticModels returns true if the manifest contains any semantic models
func (p *SemanticModelParser) HasSemanticModels() bool {
	if p.manifest == nil || p.manifest.SemanticModels == nil {
		return false
	}
	return len(p.manifest.SemanticModels) > 0
}
