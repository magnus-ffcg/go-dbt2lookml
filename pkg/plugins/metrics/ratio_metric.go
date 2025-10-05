package metrics

import (
	"fmt"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// RatioMetricGenerator handles generation of LookML measures from dbt ratio metrics
// Note: This is now used as a builder by MetricMeasureGenerator
// For direct usage, prefer MetricMeasureGenerator.GenerateMetricMeasures()
type RatioMetricGenerator struct {
	config *config.Config
}

// NewRatioMetricGenerator creates a new RatioMetricGenerator instance
func NewRatioMetricGenerator(cfg *config.Config) *RatioMetricGenerator {
	return &RatioMetricGenerator{
		config: cfg,
	}
}

// GenerateMeasureFromRatioMetric generates a LookML measure from a dbt ratio metric
// A ratio metric divides a numerator measure by a denominator measure
func (g *RatioMetricGenerator) GenerateMeasureFromRatioMetric(
	metric *models.DbtMetric,
	measureMap map[string]*models.DbtSemanticMeasure,
) (*models.LookMLMeasure, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	if !metric.IsRatio() {
		return nil, fmt.Errorf("metric %s is not a ratio type", metric.Name)
	}

	// Validate we have numerator and denominator
	if metric.TypeParams.Numerator == nil {
		return nil, fmt.Errorf("ratio metric %s missing numerator", metric.Name)
	}
	if metric.TypeParams.Denominator == nil {
		return nil, fmt.Errorf("ratio metric %s missing denominator", metric.Name)
	}

	// Look up the measures for numerator and denominator
	numeratorName := metric.TypeParams.Numerator.Name
	denominatorName := metric.TypeParams.Denominator.Name

	numeratorMeasure, numeratorExists := measureMap[numeratorName]
	denominatorMeasure, denominatorExists := measureMap[denominatorName]

	if !numeratorExists {
		return nil, fmt.Errorf("numerator measure %s not found for ratio metric %s", numeratorName, metric.Name)
	}
	if !denominatorExists {
		return nil, fmt.Errorf("denominator measure %s not found for ratio metric %s", denominatorName, metric.Name)
	}

	// Build the SQL expression
	sql := g.buildRatioSQL(numeratorMeasure, denominatorMeasure)

	measure := &models.LookMLMeasure{
		Name:        metric.Name,
		Type:        enums.MeasureNumber, // Ratio metrics are type: number
		Label:       g.getMetricLabel(metric),
		Description: g.getMetricDescription(metric),
		SQL:         &sql,
		// Note: value_format not yet supported in models.LookMLMeasure
		// Users can add this via meta tags if needed
	}

	return measure, nil
}

// buildRatioSQL builds the SQL expression for a ratio metric
// Returns: numerator_agg / NULLIF(denominator_agg, 0)
func (g *RatioMetricGenerator) buildRatioSQL(
	numerator *models.DbtSemanticMeasure,
	denominator *models.DbtSemanticMeasure,
) string {
	// Build numerator aggregation
	numeratorSQL := g.buildAggregationSQL(numerator)

	// Build denominator aggregation with NULLIF to prevent division by zero
	denominatorSQL := g.buildAggregationSQL(denominator)

	// Return: numerator / NULLIF(denominator, 0)
	return fmt.Sprintf("%s / NULLIF(%s, 0)", numeratorSQL, denominatorSQL)
}

// buildAggregationSQL builds the aggregation SQL for a measure
func (g *RatioMetricGenerator) buildAggregationSQL(measure *models.DbtSemanticMeasure) string {
	agg := measure.Agg

	// For count, no column needed
	if agg == "count" {
		return "COUNT(*)"
	}

	// Get the column expression
	var expr string
	if measure.Expr != nil && *measure.Expr != "" {
		expr = *measure.Expr
	} else {
		expr = measure.Name
	}

	// Build aggregation based on type
	switch agg {
	case "sum":
		return fmt.Sprintf("SUM(${TABLE}.%s)", expr)
	case "average", "avg":
		return fmt.Sprintf("AVG(${TABLE}.%s)", expr)
	case "min":
		return fmt.Sprintf("MIN(${TABLE}.%s)", expr)
	case "max":
		return fmt.Sprintf("MAX(${TABLE}.%s)", expr)
	case "count_distinct":
		return fmt.Sprintf("COUNT(DISTINCT ${TABLE}.%s)", expr)
	case "median":
		return fmt.Sprintf("PERCENTILE_CONT(${TABLE}.%s, 0.5)", expr)
	case "sum_boolean":
		return fmt.Sprintf("SUM(CAST(${TABLE}.%s AS INT64))", expr)
	default:
		// Fallback: just use the aggregation function
		return fmt.Sprintf("%s(${TABLE}.%s)", agg, expr)
	}
}

// getMetricLabel returns the label for the metric
func (g *RatioMetricGenerator) getMetricLabel(metric *models.DbtMetric) *string {
	if metric.Label != nil && *metric.Label != "" {
		return metric.Label
	}
	return nil
}

// getMetricDescription returns the description for the metric
func (g *RatioMetricGenerator) getMetricDescription(metric *models.DbtMetric) *string {
	if metric.Description != nil && *metric.Description != "" {
		return metric.Description
	}
	return nil
}

// GenerateMeasuresFromRatioMetrics generates LookML measures from all ratio metrics
func (g *RatioMetricGenerator) GenerateMeasuresFromRatioMetrics(
	ratioMetrics []models.DbtMetric,
	measureMap map[string]*models.DbtSemanticMeasure,
) ([]*models.LookMLMeasure, error) {
	measures := make([]*models.LookMLMeasure, 0, len(ratioMetrics))

	for i := range ratioMetrics {
		measure, err := g.GenerateMeasureFromRatioMetric(&ratioMetrics[i], measureMap)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("metric", ratioMetrics[i].Name).
				Msg("Failed to generate measure from ratio metric")
			continue
		}

		measures = append(measures, measure)
	}

	return measures, nil
}
