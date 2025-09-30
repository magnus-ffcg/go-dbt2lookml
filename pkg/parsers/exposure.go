package parsers

import (
	"strings"

	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
)

// ExposureParser handles parsing of DBT exposures from manifest
type ExposureParser struct {
	manifest *models.DbtManifest
}

// NewExposureParser creates a new ExposureParser instance
func NewExposureParser(manifest *models.DbtManifest) *ExposureParser {
	return &ExposureParser{
		manifest: manifest,
	}
}

// GetExposures gets exposure names, optionally filtered by tag
func (p *ExposureParser) GetExposures(tag string) []string {
	var exposureNames []string
	
	for _, exposure := range p.manifest.Exposures {
		// If no tag filter specified, include all exposures
		if tag == "" {
			exposureNames = append(exposureNames, p.getExposedModelNames(exposure)...)
			continue
		}
		
		// Check if exposure has the specified tag
		if p.hasTag(exposure, tag) {
			exposureNames = append(exposureNames, p.getExposedModelNames(exposure)...)
		}
	}
	
	// Remove duplicates
	return p.removeDuplicates(exposureNames)
}

// GetExposureByName gets a specific exposure by name
func (p *ExposureParser) GetExposureByName(name string) (*models.DbtExposure, bool) {
	for _, exposure := range p.manifest.Exposures {
		if exposure.Name == name {
			return &exposure, true
		}
	}
	return nil, false
}

// GetExposuresByTag gets all exposures with a specific tag
func (p *ExposureParser) GetExposuresByTag(tag string) []models.DbtExposure {
	var result []models.DbtExposure
	
	for _, exposure := range p.manifest.Exposures {
		if p.hasTag(exposure, tag) {
			result = append(result, exposure)
		}
	}
	
	return result
}

// GetAllExposures gets all exposures from the manifest
func (p *ExposureParser) GetAllExposures() []models.DbtExposure {
	var result []models.DbtExposure
	
	for _, exposure := range p.manifest.Exposures {
		result = append(result, exposure)
	}
	
	return result
}

// getExposedModelNames extracts model names from an exposure's refs
func (p *ExposureParser) getExposedModelNames(exposure models.DbtExposure) []string {
	var modelNames []string
	
	for _, ref := range exposure.Refs {
		modelNames = append(modelNames, ref.Name)
	}
	
	return modelNames
}

// hasTag checks if an exposure has a specific tag
func (p *ExposureParser) hasTag(exposure models.DbtExposure, tag string) bool {
	for _, exposureTag := range exposure.Tags {
		if strings.EqualFold(exposureTag, tag) {
			return true
		}
	}
	return false
}

// removeDuplicates removes duplicate strings from a slice
func (p *ExposureParser) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// GetExposureModelDependencies gets all model dependencies for an exposure
func (p *ExposureParser) GetExposureModelDependencies(exposureName string) ([]string, bool) {
	exposure, found := p.GetExposureByName(exposureName)
	if !found {
		return nil, false
	}
	
	var dependencies []string
	
	// Add direct refs
	dependencies = append(dependencies, p.getExposedModelNames(*exposure)...)
	
	// Add dependencies from depends_on.nodes (filter for models only)
	for _, nodeID := range exposure.DependsOn.Nodes {
		// Extract model name from node ID (format: model.project.model_name)
		parts := strings.Split(nodeID, ".")
		if len(parts) >= 3 && parts[0] == "model" {
			modelName := parts[len(parts)-1]
			dependencies = append(dependencies, modelName)
		}
	}
	
	return p.removeDuplicates(dependencies), true
}

// ValidateExposureRefs validates that all exposure refs point to existing models
func (p *ExposureParser) ValidateExposureRefs(modelNames []string) map[string][]string {
	modelSet := make(map[string]bool)
	for _, name := range modelNames {
		modelSet[name] = true
	}
	
	invalidRefs := make(map[string][]string)
	
	for exposureName, exposure := range p.manifest.Exposures {
		var invalid []string
		
		for _, ref := range exposure.Refs {
			if !modelSet[ref.Name] {
				invalid = append(invalid, ref.Name)
			}
		}
		
		if len(invalid) > 0 {
			invalidRefs[exposureName] = invalid
		}
	}
	
	return invalidRefs
}
