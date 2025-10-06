package metrics

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// SimpleMetricGenerator handles generation of LookML measures from dbt simple metrics with filters
// Note: This is now used as a builder by MetricMeasureGenerator
// For direct usage, prefer MetricMeasureGenerator.GenerateMetricMeasures()
type SimpleMetricGenerator struct {
	config *config.Config
}

// NewSimpleMetricGenerator creates a new SimpleMetricGenerator instance
func NewSimpleMetricGenerator(cfg *config.Config) *SimpleMetricGenerator {
	return &SimpleMetricGenerator{
		config: cfg,
	}
}

// GenerateMeasureFromSimpleMetric generates a LookML measure from a dbt simple metric with filters
// Simple metrics with filters add WHERE conditions to the base measure
func (g *SimpleMetricGenerator) GenerateMeasureFromSimpleMetric(
	metric *models.DbtMetric,
	baseMeasure *models.LookMLMeasure,
) (*models.LookMLMeasure, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	if !metric.IsSimple() {
		return nil, fmt.Errorf("metric %s is not a simple type", metric.Name)
	}

	if !metric.HasFilter() {
		// Simple metric without filter - just use the base measure
		return nil, fmt.Errorf("simple metric %s has no filter, should not be generated separately", metric.Name)
	}

	// Get the filter SQL and convert to LookML format
	filterSQL := metric.GetFilterSQL()
	lookMLFilter, err := g.convertDbtFilterToLookML(filterSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to convert filter for metric %s: %w", metric.Name, err)
	}

	// Create a new measure based on the base measure but with filter
	measure := &models.LookMLMeasure{
		Name:        metric.Name,
		Type:        baseMeasure.Type, // Use same type as base measure
		Label:       g.getMetricLabel(metric),
		Description: g.getMetricDescription(metric),
	}

	// Build the filtered SQL
	filteredSQL := g.buildFilteredSQL(baseMeasure, lookMLFilter)
	if filteredSQL != "" {
		measure.SQL = &filteredSQL
	} else {
		g.config.Logger().Warn().
			Str("metric", metric.Name).
			Msg("No SQL generated for filtered metric")
	}

	return measure, nil
}

// convertDbtFilterToLookML converts dbt filter syntax to LookML SQL
// dbt uses: {{ Dimension('entity__column') }} operator value
// LookML uses: ${TABLE}.column operator value
func (g *SimpleMetricGenerator) convertDbtFilterToLookML(dbtFilter string) (string, error) {
	// Clean up the filter string
	filter := strings.TrimSpace(dbtFilter)

	// Pattern to match {{ Dimension('entity__column') }}
	// The entity prefix (before __) is typically the semantic model name
	dimensionPattern := regexp.MustCompile(`\{\{\s*Dimension\(['"]([^'"]+)['"]\)\s*\}\}`)

	// Replace {{ Dimension('order__status') }} with ${TABLE}.status
	result := dimensionPattern.ReplaceAllStringFunc(filter, func(match string) string {
		// Extract the dimension reference
		matches := dimensionPattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match
		}

		dimensionRef := matches[1]

		// Split on __ to get entity and column
		parts := strings.Split(dimensionRef, "__")
		var columnName string
		if len(parts) >= 2 {
			// Take the last part as the column name
			columnName = parts[len(parts)-1]
		} else {
			// No entity prefix, use as-is
			columnName = dimensionRef
		}

		// Return LookML table reference
		return fmt.Sprintf("${TABLE}.%s", columnName)
	})

	return result, nil
}

// buildFilteredSQL builds SQL for a measure with a filter condition
func (g *SimpleMetricGenerator) buildFilteredSQL(baseMeasure *models.LookMLMeasure, filter string) string {
	baseSQL := ""
	if baseMeasure.SQL != nil {
		baseSQL = *baseMeasure.SQL
	}

	// For most aggregations, we can use CASE WHEN to filter
	// Example: SUM(CASE WHEN condition THEN column ELSE NULL END)
	switch baseMeasure.Type {
	case "count":
		// COUNT(*) with filter becomes COUNT(CASE WHEN condition THEN 1 END)
		return fmt.Sprintf("COUNT(CASE WHEN %s THEN 1 END)", filter)
	case "count_distinct":
		// COUNT(DISTINCT column) with filter
		// Extract the column from baseSQL
		if baseSQL == "" {
			// If no SQL specified, we can't build the filter
			g.config.Logger().Warn().
				Str("measure", baseMeasure.Name).
				Msg("Cannot add filter to count_distinct without SQL column")
			return ""
		}
		column := g.extractColumnFromSQL(baseSQL)
		return fmt.Sprintf("COUNT(DISTINCT CASE WHEN %s THEN %s END)", filter, column)
	case "sum", "average", "min", "max":
		// AGG(column) with filter becomes AGG(CASE WHEN condition THEN column END)
		if baseSQL == "" {
			// If no SQL specified, we can't build the filter
			g.config.Logger().Warn().
				Str("measure", baseMeasure.Name).
				Str("type", string(baseMeasure.Type)).
				Msg("Cannot add filter without SQL column")
			return ""
		}
		column := g.extractColumnFromSQL(baseSQL)
		aggType := strings.ToUpper(string(baseMeasure.Type))
		return fmt.Sprintf("%s(CASE WHEN %s THEN %s END)", aggType, filter, column)
	default:
		// For other types, wrap in CASE WHEN
		if baseSQL == "" {
			return fmt.Sprintf("CASE WHEN %s THEN 1 ELSE 0 END", filter)
		}
		return fmt.Sprintf("CASE WHEN %s THEN %s END", filter, baseSQL)
	}
}

// extractColumnFromSQL extracts the column reference from an aggregation SQL
// Example: "${TABLE}.amount" or "SUM(${TABLE}.amount)" -> "${TABLE}.amount"
func (g *SimpleMetricGenerator) extractColumnFromSQL(sql string) string {
	// Remove aggregation function if present
	sql = strings.TrimSpace(sql)

	// Pattern to match ${TABLE}.column_name
	pattern := regexp.MustCompile(`\$\{TABLE\}\.(\w+)`)
	if pattern.MatchString(sql) {
		// Extract just the table reference
		matches := pattern.FindStringSubmatch(sql)
		if len(matches) >= 1 {
			return matches[0] // Return full ${TABLE}.column
		}
	}

	// If no pattern match, return as-is
	return sql
}

// getMetricLabel returns the label for the metric
func (g *SimpleMetricGenerator) getMetricLabel(metric *models.DbtMetric) *string {
	if metric.Label != nil && *metric.Label != "" {
		return metric.Label
	}
	return nil
}

// getMetricDescription returns the description for the metric
func (g *SimpleMetricGenerator) getMetricDescription(metric *models.DbtMetric) *string {
	if metric.Description != nil && *metric.Description != "" {
		return metric.Description
	}
	return nil
}

// GenerateMeasuresFromSimpleMetrics generates LookML measures from simple metrics with filters
func (g *SimpleMetricGenerator) GenerateMeasuresFromSimpleMetrics(
	simpleMetrics []models.DbtMetric,
	existingMeasures map[string]*models.LookMLMeasure,
) ([]*models.LookMLMeasure, error) {
	measures := make([]*models.LookMLMeasure, 0)

	for i := range simpleMetrics {
		metric := &simpleMetrics[i]

		// Skip simple metrics without filters (they're just aliases)
		if !metric.HasFilter() {
			continue
		}

		// Get the base measure name
		if metric.TypeParams.Measure == nil {
			g.config.Logger().Warn().
				Str("metric", metric.Name).
				Msg("Simple metric has no base measure")
			continue
		}

		baseMeasureName := metric.TypeParams.Measure.Name
		baseMeasure, exists := existingMeasures[baseMeasureName]
		if !exists {
			g.config.Logger().Warn().
				Str("metric", metric.Name).
				Str("base_measure", baseMeasureName).
				Msg("Base measure not found for simple metric")
			continue
		}

		measure, err := g.GenerateMeasureFromSimpleMetric(metric, baseMeasure)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("metric", metric.Name).
				Msg("Failed to generate measure from simple metric")
			continue
		}

		measures = append(measures, measure)
	}

	return measures, nil
}
