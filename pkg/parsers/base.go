// Package parsers provides functionality for parsing dbt manifest and catalog files.
//
// This package handles reading and parsing dbt's generated JSON files (manifest.json
// and catalog.json) into Go structs for further processing. It manages the complex
// task of merging catalog metadata with manifest model definitions.
//
// Key components:
//   - DbtParser: Main coordinator for parsing dbt artifacts
//   - ModelParser: Parses dbt model definitions
//   - CatalogParser: Parses dbt catalog metadata
//   - ExposureParser: Parses dbt exposure definitions
//
// The parser supports:
//   - BigQuery data types and nested structures
//   - Column metadata and descriptions
//   - Model relationships and dependencies
//   - Custom meta tags for LookML generation
//   - Filtering by tags and model selection
//
// Example usage:
//
//	cfg := &config.Config{
//	    ManifestPath: "./target/manifest.json",
//	    CatalogPath:  "./target/catalog.json",
//	}
//	parser := NewDbtParser(cfg)
//	models, err := parser.Parse()
package parsers

import (
	"encoding/json"
	"fmt"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// DbtParser is the main DBT parser that coordinates parsing of manifest and catalog files
type DbtParser struct {
	config              *config.Config         // Configuration with CLI arguments
	rawManifest         map[string]interface{} // Raw manifest data
	manifest            *models.DbtManifest    // Parsed manifest
	catalog             *models.DbtCatalog
	modelParser         *ModelParser
	catalogParser       *CatalogParser
	exposureParser      *ExposureParser
	semanticModelParser *SemanticModelParser
	metricParser        *MetricParser
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
		manifest:    &manifest,
		catalog:     &catalog,
	}

	// Initialize sub-parsers
	parser.modelParser = NewModelParser(&manifest, parser.config)
	parser.catalogParser = NewCatalogParser(&catalog, rawCatalog, parser.config)
	parser.exposureParser = NewExposureParser(&manifest)
	parser.semanticModelParser = NewSemanticModelParser(&manifest)
	parser.metricParser = NewMetricParser(&manifest)

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
		SelectModel:   p.getSelectModel(),
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
			if err != nil {
				p.config.Logger().Warn().Str("model", model.Name).Err(err).Msg("Failed to process model")
			}
		}
	}

	// Log any models that failed processing
	if len(failedModels) > 0 {
		failedCount := len(failedModels)
		displayNames := failedModels
		if len(displayNames) > 5 {
			displayNames = displayNames[:5]
		}

		p.config.Logger().Warn().Int("failed_count", failedCount).Strs("models", displayNames).Msg("Failed to process models during catalog parsing")
		if len(failedModels) > 5 {
			p.config.Logger().Warn().Int("additional", len(failedModels)-5).Msg("Additional models failed")
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

// GetSemanticModelParser returns the semantic model parser
// Deprecated: Use GetManifest and parse internally in plugins
func (p *DbtParser) GetSemanticModelParser() *SemanticModelParser {
	return p.semanticModelParser
}

// GetMetricParser returns the metric parser
// Deprecated: Use GetManifest and parse internally in plugins
func (p *DbtParser) GetMetricParser() *MetricParser {
	return p.metricParser
}

// GetManifest returns the parsed manifest
func (p *DbtParser) GetManifest() *models.DbtManifest {
	return p.manifest
}
