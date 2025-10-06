// Package metrics provides pluggable extensions for LookML generation
package metrics

import (
	"context"
	"fmt"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

const (
	filePermissions = 0644
)

// MetricsPlugin handles all semantic model metric generation
// This plugin is responsible for generating view extensions and derived tables
// for semantic measures, ratio metrics, derived metrics, cumulative metrics, and conversion metrics
type MetricsPlugin struct {
	config *config.Config

	// Internal parser for semantic models and metrics
	parser *parser

	// Parsed metric data (populated by OnManifestLoaded)
	semanticMeasures  map[string][]models.DbtSemanticMeasure // model name -> semantic measures
	ratioMetrics      []models.DbtMetric                     // global ratio metrics
	derivedMetrics    []models.DbtMetric                     // global derived metrics
	simpleMetrics     []models.DbtMetric                     // global simple metrics (with filters)
	cumulativeMetrics []models.DbtMetric                     // global cumulative metrics
	conversionMetrics []models.DbtMetric                     // global conversion metrics

	// Metric measure generator (for simple/ratio/derived metrics)
	metricMeasureGenerator *MetricMeasureGenerator
}

// NewMetricsPlugin creates a new MetricsPlugin instance
func NewMetricsPlugin(cfg *config.Config) *MetricsPlugin {
	return &MetricsPlugin{
		config:                 cfg,
		semanticMeasures:       make(map[string][]models.DbtSemanticMeasure),
		ratioMetrics:           make([]models.DbtMetric, 0),
		derivedMetrics:         make([]models.DbtMetric, 0),
		simpleMetrics:          make([]models.DbtMetric, 0),
		cumulativeMetrics:      make([]models.DbtMetric, 0),
		conversionMetrics:      make([]models.DbtMetric, 0),
		metricMeasureGenerator: NewMetricMeasureGenerator(cfg),
	}
}

// Enabled returns whether this plugin is enabled
func (p *MetricsPlugin) Enabled() bool {
	return p.config.UseSemanticModels
}

// Name returns the plugin name
func (p *MetricsPlugin) Name() string {
	return "SemanticMetrics"
}

// ============================================================================
// DataIngestionHook Implementation
// ============================================================================

// OnManifestLoaded is called when the manifest is loaded
// The plugin parses semantic models and metrics internally from the raw manifest
func (p *MetricsPlugin) OnManifestLoaded(manifest *models.DbtManifest) {
	if !p.Enabled() {
		return
	}

	// Create parser from manifest
	p.parser = newParser(manifest)

	// Parse semantic measures
	p.semanticMeasures = p.parser.parseSemanticMeasures()

	// Parse all metric types
	p.ratioMetrics = p.parser.parseRatioMetrics()
	p.derivedMetrics = p.parser.parseDerivedMetrics()
	p.simpleMetrics = p.parser.parseSimpleMetrics()
	p.cumulativeMetrics = p.parser.parseCumulativeMetrics()
	p.conversionMetrics = p.parser.parseConversionMetrics()
}

// ============================================================================
// Legacy Methods (for backward compatibility - will be removed)
// ============================================================================

// SetSemanticMeasures sets the semantic measures mapping
// Deprecated: Use OnManifestLoaded instead
func (p *MetricsPlugin) SetSemanticMeasures(semanticMeasures map[string][]models.DbtSemanticMeasure) {
	p.semanticMeasures = semanticMeasures
}

// SetRatioMetrics sets the ratio metrics
// Deprecated: Use OnManifestLoaded instead
func (p *MetricsPlugin) SetRatioMetrics(ratioMetrics []models.DbtMetric) {
	p.ratioMetrics = ratioMetrics
}

// SetDerivedMetrics sets the derived metrics
// Deprecated: Use OnManifestLoaded instead
func (p *MetricsPlugin) SetDerivedMetrics(derivedMetrics []models.DbtMetric) {
	p.derivedMetrics = derivedMetrics
}

// SetSimpleMetrics sets the simple metrics
// Deprecated: Use OnManifestLoaded instead
func (p *MetricsPlugin) SetSimpleMetrics(simpleMetrics []models.DbtMetric) {
	p.simpleMetrics = simpleMetrics
}

// SetCumulativeMetrics sets the cumulative metrics
// Deprecated: Use OnManifestLoaded instead
func (p *MetricsPlugin) SetCumulativeMetrics(cumulativeMetrics []models.DbtMetric) {
	p.cumulativeMetrics = cumulativeMetrics
}

// SetConversionMetrics sets the conversion metrics
func (p *MetricsPlugin) SetConversionMetrics(conversionMetrics []models.DbtMetric) {
	p.conversionMetrics = conversionMetrics
}

// HasMetricsForModel checks if there are any metrics for the given model
func (p *MetricsPlugin) HasMetricsForModel(modelName string) bool {
	if !p.Enabled() {
		return false
	}

	// Check semantic measures
	if measures, ok := p.semanticMeasures[modelName]; ok && len(measures) > 0 {
		return true
	}

	// Check if any metrics reference this model
	hasMetrics := len(p.ratioMetrics) > 0 ||
		len(p.derivedMetrics) > 0 ||
		len(p.simpleMetrics) > 0 ||
		len(p.cumulativeMetrics) > 0 ||
		len(p.conversionMetrics) > 0

	return hasMetrics
}

// GetExploreJoins returns the joins needed for metric views
func (p *MetricsPlugin) GetExploreJoins(model *models.DbtModel, baseName string) []models.LookMLJoin {
	if !p.Enabled() {
		return nil
	}

	var joins []models.LookMLJoin

	// Check if cumulative metrics exist for this model
	hasCumulative := p.hasCumulativeMetricsForModel(model)
	if hasCumulative {
		joins = append(joins, p.createCumulativeJoin(baseName))
	}

	// Check if conversion metrics exist for this model
	hasConversion := p.hasConversionMetricsForModel(model)
	if hasConversion {
		joins = append(joins, p.createConversionJoin(baseName))
	}

	return joins
}

// createCumulativeJoin creates a join for cumulative metrics view
func (p *MetricsPlugin) createCumulativeJoin(baseName string) models.LookMLJoin {
	cumulativeViewName := fmt.Sprintf("%s__cumulative", baseName)
	relationship := enums.LookerRelationshipType("one_to_one")
	joinType := enums.JoinLeftOuter

	// TODO: Detect primary key from semantic model instead of hardcoding
	sqlOn := fmt.Sprintf("${%s.order_id} = ${%s.order_id}", baseName, cumulativeViewName)

	label := "Cumulative Metrics"

	return models.LookMLJoin{
		Name:         cumulativeViewName,
		ViewLabel:    &label,
		SQL:          &sqlOn,
		Type:         &joinType,
		Relationship: &relationship,
	}
}

// createConversionJoin creates a join for conversion metrics view
func (p *MetricsPlugin) createConversionJoin(baseName string) models.LookMLJoin {
	conversionViewName := fmt.Sprintf("%s__conversion", baseName)
	relationship := enums.LookerRelationshipType("one_to_one")
	joinType := enums.JoinLeftOuter

	// TODO: Detect entity key from semantic model instead of hardcoding
	sqlOn := fmt.Sprintf("${%s.customer_id} = ${%s.customer_id}", baseName, conversionViewName)

	label := "Conversion Metrics"

	return models.LookMLJoin{
		Name:         conversionViewName,
		ViewLabel:    &label,
		SQL:          &sqlOn,
		Type:         &joinType,
		Relationship: &relationship,
	}
}

// hasCumulativeMetricsForModel checks if model has cumulative metrics
func (p *MetricsPlugin) hasCumulativeMetricsForModel(model *models.DbtModel) bool {
	if len(p.cumulativeMetrics) == 0 {
		return false
	}

	// Get semantic measures for this model
	semanticMeasures, ok := p.semanticMeasures[model.Name]
	if !ok {
		return false
	}

	baseMeasureMap := make(map[string]bool)
	for _, sm := range semanticMeasures {
		baseMeasureMap[sm.Name] = true
	}

	// Check if any cumulative metric references measures from this model
	for _, metric := range p.cumulativeMetrics {
		if metric.TypeParams.Measure != nil {
			if baseMeasureMap[metric.TypeParams.Measure.Name] {
				return true
			}
		}
	}

	return false
}

// hasConversionMetricsForModel checks if model has conversion metrics
func (p *MetricsPlugin) hasConversionMetricsForModel(model *models.DbtModel) bool {
	if len(p.conversionMetrics) == 0 {
		return false
	}

	// Get semantic measures for this model
	semanticMeasures, ok := p.semanticMeasures[model.Name]
	if !ok {
		return false
	}

	baseMeasureMap := make(map[string]bool)
	for _, sm := range semanticMeasures {
		baseMeasureMap[sm.Name] = true
	}

	// Check if any conversion metric references measures from this model
	for _, metric := range p.conversionMetrics {
		if metric.TypeParams.ConversionTypeParams != nil {
			params := metric.TypeParams.ConversionTypeParams
			baseMeasureName := params.BaseMeasure.Name
			conversionMeasureName := params.ConversionMeasure.Name

			if baseMeasureMap[baseMeasureName] && baseMeasureMap[conversionMeasureName] {
				return true
			}
		}
	}

	return false
}

// GetSemanticMeasures returns semantic measures for a model
func (p *MetricsPlugin) GetSemanticMeasures(modelName string) []models.DbtSemanticMeasure {
	if !p.Enabled() {
		return nil
	}
	return p.semanticMeasures[modelName]
}

// GetRatioMetrics returns all ratio metrics
func (p *MetricsPlugin) GetRatioMetrics() []models.DbtMetric {
	if !p.Enabled() {
		return nil
	}
	return p.ratioMetrics
}

// GetDerivedMetrics returns all derived metrics
func (p *MetricsPlugin) GetDerivedMetrics() []models.DbtMetric {
	if !p.Enabled() {
		return nil
	}
	return p.derivedMetrics
}

// GetSimpleMetrics returns all simple metrics
func (p *MetricsPlugin) GetSimpleMetrics() []models.DbtMetric {
	if !p.Enabled() {
		return nil
	}
	return p.simpleMetrics
}

// GetCumulativeMetrics returns all cumulative metrics
func (p *MetricsPlugin) GetCumulativeMetrics() []models.DbtMetric {
	if !p.Enabled() {
		return nil
	}
	return p.cumulativeMetrics
}

// GetConversionMetrics returns all conversion metrics
func (p *MetricsPlugin) GetConversionMetrics() []models.DbtMetric {
	if !p.Enabled() {
		return nil
	}
	return p.conversionMetrics
}

// ========== Hook Interface Implementations ==========

// OnSemanticMeasures implements DataIngestionHook (legacy)
// Deprecated: This is kept for backward compatibility. Use OnManifestLoaded instead
func (p *MetricsPlugin) OnSemanticMeasures(measures map[string][]models.DbtSemanticMeasure) {
	p.semanticMeasures = measures
}

// OnMetrics implements DataIngestionHook (legacy)
// Deprecated: This is kept for backward compatibility. Use OnManifestLoaded instead
func (p *MetricsPlugin) OnMetrics(metrics []models.DbtMetric, metricType string) {
	switch metricType {
	case "ratio":
		p.ratioMetrics = metrics
	case "derived":
		p.derivedMetrics = metrics
	case "simple":
		p.simpleMetrics = metrics
	case "cumulative":
		p.cumulativeMetrics = metrics
	case "conversion":
		p.conversionMetrics = metrics
	}
}

// AfterModelGeneration implements ModelGenerationHook
// Generates semantic layer files after core model generation
func (p *MetricsPlugin) AfterModelGeneration(ctx context.Context, model *models.DbtModel) error {
	return p.GenerateForModel(model)
}

// EnrichExplore implements ExploreEnrichmentHook
// Adds joins for metric views to the explore
func (p *MetricsPlugin) EnrichExplore(ctx context.Context, model *models.DbtModel, explore *models.LookMLExplore, baseName string) error {
	if !p.Enabled() {
		return nil
	}

	joins := p.GetExploreJoins(model, baseName)
	explore.Joins = append(explore.Joins, joins...)

	return nil
}
