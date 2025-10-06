package metrics

import (
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// parser handles parsing of semantic models and metrics from dbt manifest
// This is internal to the metrics plugin - semantic data only needs to be parsed here
type parser struct {
	manifest *models.DbtManifest
}

// newParser creates a new parser instance
func newParser(manifest *models.DbtManifest) *parser {
	return &parser{
		manifest: manifest,
	}
}

// ============================================================================
// Semantic Model Parsing
// ============================================================================

// parseSemanticMeasures extracts semantic measures grouped by model ref
func (p *parser) parseSemanticMeasures() map[string][]models.DbtSemanticMeasure {
	measures := make(map[string][]models.DbtSemanticMeasure)

	if p.manifest == nil || p.manifest.SemanticModels == nil {
		return measures
	}

	for _, sm := range p.manifest.SemanticModels {
		if len(sm.Measures) > 0 {
			modelRef := sm.GetModelRef()
			measures[modelRef] = append(measures[modelRef], sm.Measures...)
		}
	}

	return measures
}

// getSemanticModels returns all semantic models from the manifest
func (p *parser) getSemanticModels() []models.DbtSemanticModel {
	if p.manifest == nil || p.manifest.SemanticModels == nil {
		return []models.DbtSemanticModel{}
	}

	semanticModels := make([]models.DbtSemanticModel, 0, len(p.manifest.SemanticModels))
	for _, sm := range p.manifest.SemanticModels {
		semanticModels = append(semanticModels, sm)
	}

	return semanticModels
}

// getMeasuresForModel returns all measures from semantic models that reference a dbt model
func (p *parser) getMeasuresForModel(modelName string) []models.DbtSemanticMeasure {
	var allMeasures []models.DbtSemanticMeasure

	for _, sm := range p.getSemanticModels() {
		refModelName := sm.GetModelRef()
		if refModelName == modelName {
			allMeasures = append(allMeasures, sm.Measures...)
		}
	}

	return allMeasures
}

// hasSemanticModels returns true if the manifest contains any semantic models
func (p *parser) hasSemanticModels() bool {
	return p.manifest != nil && p.manifest.SemanticModels != nil && len(p.manifest.SemanticModels) > 0
}

// ============================================================================
// Metric Parsing
// ============================================================================

// parseRatioMetrics returns all ratio-type metrics
func (p *parser) parseRatioMetrics() []models.DbtMetric {
	if !p.hasMetrics() {
		return []models.DbtMetric{}
	}

	var ratioMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsRatio() {
			ratioMetrics = append(ratioMetrics, metric)
		}
	}

	return ratioMetrics
}

// parseSimpleMetrics returns all simple-type metrics
func (p *parser) parseSimpleMetrics() []models.DbtMetric {
	if !p.hasMetrics() {
		return []models.DbtMetric{}
	}

	var simpleMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsSimple() {
			simpleMetrics = append(simpleMetrics, metric)
		}
	}

	return simpleMetrics
}

// parseDerivedMetrics returns all derived-type metrics
func (p *parser) parseDerivedMetrics() []models.DbtMetric {
	if !p.hasMetrics() {
		return []models.DbtMetric{}
	}

	var derivedMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsDerived() {
			derivedMetrics = append(derivedMetrics, metric)
		}
	}

	return derivedMetrics
}

// parseCumulativeMetrics returns all cumulative metrics from the manifest
func (p *parser) parseCumulativeMetrics() []models.DbtMetric {
	if !p.hasMetrics() {
		return []models.DbtMetric{}
	}

	var cumulativeMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsCumulative() {
			cumulativeMetrics = append(cumulativeMetrics, metric)
		}
	}

	return cumulativeMetrics
}

// parseConversionMetrics returns all conversion metrics from the manifest
func (p *parser) parseConversionMetrics() []models.DbtMetric {
	if !p.hasMetrics() {
		return []models.DbtMetric{}
	}

	var conversionMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsConversion() {
			conversionMetrics = append(conversionMetrics, metric)
		}
	}

	return conversionMetrics
}

// hasMetrics returns true if the manifest contains any metrics
func (p *parser) hasMetrics() bool {
	return p.manifest != nil && p.manifest.Metrics != nil && len(p.manifest.Metrics) > 0
}

// ============================================================================
// Utility Functions
// ============================================================================

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

// findSemanticMeasureByName finds a measure by name across all semantic models
func (p *parser) findSemanticMeasureByName(measureName string) *models.DbtSemanticMeasure {
	for _, sm := range p.getSemanticModels() {
		for _, measure := range sm.Measures {
			if measure.Name == measureName {
				return &measure
			}
		}
	}
	return nil
}

// getModelRefForMeasure returns the model ref for a given measure name
func (p *parser) getModelRefForMeasure(measureName string) string {
	for _, sm := range p.getSemanticModels() {
		for _, measure := range sm.Measures {
			if measure.Name == measureName {
				return sm.GetModelRef()
			}
		}
	}
	return ""
}
