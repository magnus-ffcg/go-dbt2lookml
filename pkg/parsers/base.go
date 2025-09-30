package parsers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/magnus-ffcg/dbt2lookml/internal/config"
	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
)

// DbtParser is the main DBT parser that coordinates parsing of manifest and catalog files
type DbtParser struct {
	config      *config.Config         // Configuration with CLI arguments
	rawManifest map[string]interface{} // Raw manifest data
	catalog     *models.DbtCatalog
	modelParser    *ModelParser
	catalogParser  *CatalogParser
	exposureParser *ExposureParser
}

// NewDbtParser creates a new DbtParser instance
func NewDbtParser(cliArgs interface{}, rawManifest, rawCatalog map[string]interface{}) (*DbtParser, error) {
	// Parse catalog
	catalogBytes, err := json.Marshal(rawCatalog)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal catalog: %w", err)
	}
	
	var catalog models.DbtCatalog
	if err := json.Unmarshal(catalogBytes, &catalog); err != nil {
		return nil, fmt.Errorf("failed to unmarshal catalog: %w", err)
	}

	// Parse manifest
	manifestBytes, err := json.Marshal(rawManifest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal manifest: %w", err)
	}
	
	var manifest models.DbtManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal manifest: %w", err)
	}

	// Validate adapter
	if err := manifest.Metadata.ValidateAdapter(); err != nil {
		return nil, err
	}

	parser := &DbtParser{
		config:      cliArgs.(*config.Config),
		rawManifest: rawManifest,
		catalog:     &catalog,
	}

	// Initialize sub-parsers
	parser.modelParser = NewModelParser(&manifest)
	parser.catalogParser = NewCatalogParser(&catalog, rawCatalog)
	parser.exposureParser = NewExposureParser(&manifest)

	return parser, nil
}

// GetModels parses dbt models from manifest and filters by criteria
func (p *DbtParser) GetModels() ([]*models.DbtModel, error) {
	// Get all models
	allModels, err := p.modelParser.GetAllModels()
	if err != nil {
		return nil, fmt.Errorf("failed to get all models: %w", err)
	}

	// Get exposed models if needed (simplified for now)
	var exposedNames []string
	// This would need proper CLI args handling
	// if p.shouldFilterByExposures() {
	//     exposedNames = p.exposureParser.GetExposures(exposuresTag)
	// }

	// Filter models based on criteria
	filteredModels := p.modelParser.FilterModels(allModels, ModelFilterOptions{
		SelectModel:    p.getSelectModel(),
		Tag:           p.getTag(),
		ExposedNames:  exposedNames,
		IncludeModels: p.getIncludeModels(),
		ExcludeModels: p.getExcludeModels(),
	})

	// Process models (update with catalog info)
	var processedModels []*models.DbtModel
	var failedModels []string

	for _, model := range filteredModels {
		if processedModel, err := p.catalogParser.ProcessModelColumns(model); err == nil && processedModel != nil {
			// Debug: check if processed model has ARRAY columns
			if strings.Contains(model.Name, "dq_ICASOI_Current") {
				arrayCount := 0
				for colName, col := range processedModel.Columns {
					if col.DataType != nil && strings.HasPrefix(strings.ToUpper(*col.DataType), "ARRAY") {
						arrayCount++
						log.Printf("DEBUG PARSER: Processed model %s has ARRAY column %s", model.Name, colName)
					}
				}
				log.Printf("DEBUG PARSER: Processed model %s has %d ARRAY columns", model.Name, arrayCount)
			}
			
			// Store catalog data reference for generators (would need proper implementation)
			// processedModel.CatalogData = p.catalogParser.rawCatalogData
			
			// Store original raw manifest data for metadata extraction
			if rawNodes, ok := p.rawManifest["nodes"].(map[string]interface{}); ok {
				if manifestData, exists := rawNodes[model.UniqueID]; exists {
					// processedModel.ManifestData = manifestData
					_ = manifestData // Placeholder
				}
			}
			
			processedModels = append(processedModels, processedModel)
		} else {
			failedModels = append(failedModels, model.Name)
			log.Printf("DEBUG PARSER: Failed to process model %s: %v", model.Name, err)
		}
	}

	// Log any models that failed processing
	if len(failedModels) > 0 {
		failedCount := len(failedModels)
		displayNames := failedModels
		if len(displayNames) > 5 {
			displayNames = displayNames[:5]
		}
		
		log.Printf("Failed to process %d models during catalog parsing: %v", failedCount, displayNames)
		if len(failedModels) > 5 {
			log.Printf("... and %d more", len(failedModels)-5)
		}
	}

	return processedModels, nil
}

// Helper methods to extract CLI arguments from config
func (p *DbtParser) getSelectModel() string {
	return p.config.Select
}

func (p *DbtParser) getTag() string {
	return p.config.Tag
}

func (p *DbtParser) getIncludeModels() []string {
	return p.config.IncludeModels
}

func (p *DbtParser) getExcludeModels() []string {
	return p.config.ExcludeModels
}
