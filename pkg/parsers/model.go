package parsers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// ModelParser handles parsing of DBT models from manifest
type ModelParser struct {
	manifest *models.DbtManifest
}

// ModelFilterOptions contains options for filtering models
type ModelFilterOptions struct {
	SelectModel   string
	Tag           string
	ExposedNames  []string
	IncludeModels []string
	ExcludeModels []string
}

// NewModelParser creates a new ModelParser instance
func NewModelParser(manifest *models.DbtManifest) *ModelParser {
	return &ModelParser{
		manifest: manifest,
	}
}

// GetAllModels gets all models from manifest
func (p *ModelParser) GetAllModels() ([]*models.DbtModel, error) {
	allModels := p.filterNodesByType(p.manifest.Nodes, string(enums.ResourceModel))

	// Validate models
	var validModels []*models.DbtModel
	for _, model := range allModels {
		if model.Name == "" {
			log.Printf("Cannot parse model with id: %q - is the model file empty?", model.UniqueID)
			continue
		}
		validModels = append(validModels, model)
	}

	return validModels, nil
}

// FilterModels filters models based on multiple criteria
func (p *ModelParser) FilterModels(modelsList []*models.DbtModel, options ModelFilterOptions) []*models.DbtModel {
	filtered := modelsList

	// Single model selection takes precedence
	if options.SelectModel != "" {
		var result []*models.DbtModel
		for _, model := range filtered {
			if model.Name == options.SelectModel {
				result = append(result, model)
			}
		}
		return result
	}

	// Filter by tag
	if options.Tag != "" {
		var result []*models.DbtModel
		for _, model := range filtered {
			if p.tagsMatch(options.Tag, model) {
				result = append(result, model)
			}
		}
		filtered = result
	}

	// Filter by exposed names
	if len(options.ExposedNames) > 0 {
		var result []*models.DbtModel
		for _, model := range filtered {
			for _, exposedName := range options.ExposedNames {
				if model.Name == exposedName {
					result = append(result, model)
					break
				}
			}
		}
		filtered = result
	}

	// Filter by include models
	if len(options.IncludeModels) > 0 {
		var result []*models.DbtModel
		for _, model := range filtered {
			for _, includeName := range options.IncludeModels {
				if model.Name == includeName {
					result = append(result, model)
					break
				}
			}
		}
		filtered = result
	}

	// Filter by exclude models
	if len(options.ExcludeModels) > 0 {
		var result []*models.DbtModel
		for _, model := range filtered {
			excluded := false
			for _, excludeName := range options.ExcludeModels {
				if model.Name == excludeName {
					excluded = true
					break
				}
			}
			if !excluded {
				result = append(result, model)
			}
		}
		filtered = result
	}

	return filtered
}

// filterNodesByType filters nodes by resource type and ensures they have names
func (p *ModelParser) filterNodesByType(nodes map[string]interface{}, resourceType string) []*models.DbtModel {
	var result []*models.DbtModel

	for _, node := range nodes {
		// Convert node to DbtModel
		if model := p.convertToModel(node); model != nil && model.ResourceType == resourceType {
			result = append(result, model)
		}
	}

	return result
}

// convertToModel converts a generic node interface to a DbtModel
func (p *ModelParser) convertToModel(node interface{}) *models.DbtModel {
	// Convert to JSON and back to properly unmarshal into struct
	nodeBytes, err := json.Marshal(node)
	if err != nil {
		return nil
	}

	var model models.DbtModel
	if err := json.Unmarshal(nodeBytes, &model); err != nil {
		return nil
	}

	// Process columns
	model.NormalizeColumnNames()
	for name, column := range model.Columns {
		column.ProcessColumn()
		model.Columns[name] = column
	}

	return &model
}

// tagsMatch checks if model has the specified tag
func (p *ModelParser) tagsMatch(tag string, model *models.DbtModel) bool {
	for _, modelTag := range model.Tags {
		if strings.EqualFold(modelTag, tag) {
			return true
		}
	}
	return false
}

// GetModelByName gets a specific model by name
func (p *ModelParser) GetModelByName(name string) (*models.DbtModel, error) {
	allModels, err := p.GetAllModels()
	if err != nil {
		return nil, err
	}

	for _, model := range allModels {
		if model.Name == name {
			return model, nil
		}
	}

	return nil, fmt.Errorf("model %q not found", name)
}

// GetModelsByTag gets all models with a specific tag
func (p *ModelParser) GetModelsByTag(tag string) ([]*models.DbtModel, error) {
	allModels, err := p.GetAllModels()
	if err != nil {
		return nil, err
	}

	var result []*models.DbtModel
	for _, model := range allModels {
		if p.tagsMatch(tag, model) {
			result = append(result, model)
		}
	}

	return result, nil
}
