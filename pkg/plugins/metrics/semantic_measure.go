package metrics

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// SemanticMeasureGenerator handles generation of LookML measures from dbt semantic models
type SemanticMeasureGenerator struct {
	config *config.Config
}

// NewSemanticMeasureGenerator creates a new SemanticMeasureGenerator instance
func NewSemanticMeasureGenerator(cfg *config.Config) *SemanticMeasureGenerator {
	return &SemanticMeasureGenerator{
		config: cfg,
	}
}

// GenerateMeasureFromSemantic generates a LookML measure from a semantic model measure
func (g *SemanticMeasureGenerator) GenerateMeasureFromSemantic(
	semanticMeasure *models.DbtSemanticMeasure,
	model *models.DbtModel,
) (*models.LookMLMeasure, error) {
	if semanticMeasure == nil {
		return nil, fmt.Errorf("semantic measure is nil")
	}

	// Map aggregation type to LookML measure type
	measureType, err := g.mapAggregationType(semanticMeasure.Agg)
	if err != nil {
		return nil, fmt.Errorf("failed to map aggregation type %s: %w", semanticMeasure.Agg, err)
	}

	measure := &models.LookMLMeasure{
		Name:        semanticMeasure.Name,
		Type:        measureType,
		Label:       g.getMeasureLabel(semanticMeasure),
		Description: g.getMeasureDescription(semanticMeasure),
		SQL:         g.getMeasureSQL(semanticMeasure, model),
	}

	// Handle percentile-specific configuration
	if semanticMeasure.IsPercentile() {
		if percentile, ok := semanticMeasure.GetPercentileValue(); ok {
			// Convert float (0.95) to int (95)
			percentileInt := int(percentile * 100)
			measure.Percentile = &percentileInt
		}
	}

	// TODO: Handle non_additive_dimension for semi-additive measures
	// This would require more complex SQL generation with QUALIFY or window functions
	if semanticMeasure.IsSemiAdditive() {
		g.config.Logger().Warn().
			Str("measure", semanticMeasure.Name).
			Msg("Semi-additive measures (non_additive_dimension) are not yet fully supported")
	}

	return measure, nil
}

// GenerateMeasuresFromSemantic generates LookML measures from all semantic measures for a model
func (g *SemanticMeasureGenerator) GenerateMeasuresFromSemantic(
	semanticMeasures []models.DbtSemanticMeasure,
	model *models.DbtModel,
) ([]*models.LookMLMeasure, error) {
	measures := make([]*models.LookMLMeasure, 0, len(semanticMeasures))

	for i := range semanticMeasures {
		// Note: create_metric flag is for dbt metrics, not LookML generation
		// We generate LookML measures for all semantic measures

		measure, err := g.GenerateMeasureFromSemantic(&semanticMeasures[i], model)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("measure", semanticMeasures[i].Name).
				Msg("Failed to generate measure from semantic model")
			continue
		}

		measures = append(measures, measure)
	}

	return measures, nil
}

// mapAggregationType maps dbt semantic model aggregation type to LookML measure type
func (g *SemanticMeasureGenerator) mapAggregationType(agg string) (enums.LookerMeasureType, error) {
	switch strings.ToLower(agg) {
	case "sum":
		return enums.MeasureSum, nil
	case "average", "avg":
		return enums.MeasureAverage, nil
	case "min":
		return enums.MeasureMin, nil
	case "max":
		return enums.MeasureMax, nil
	case "count":
		return enums.MeasureCount, nil
	case "count_distinct":
		return enums.MeasureCountDistinct, nil
	case "median":
		return enums.MeasureMedian, nil
	case "percentile":
		// Percentile measures in LookML use the type "percentile_XX" where XX is the percentile value
		// For now, we'll use "number" and handle the percentile value separately
		return enums.MeasureNumber, nil
	case "sum_boolean":
		// sum_boolean in dbt is essentially a SUM of boolean values (cast to int)
		// In LookML, this is just a sum measure with appropriate SQL
		return enums.MeasureSum, nil
	default:
		return "", fmt.Errorf("unsupported aggregation type: %s", agg)
	}
}

// getMeasureLabel returns the label for the measure
func (g *SemanticMeasureGenerator) getMeasureLabel(measure *models.DbtSemanticMeasure) *string {
	if measure.Label != nil && *measure.Label != "" {
		return measure.Label
	}
	return nil
}

// getMeasureDescription returns the description for the measure
func (g *SemanticMeasureGenerator) getMeasureDescription(measure *models.DbtSemanticMeasure) *string {
	if measure.Description != nil && *measure.Description != "" {
		return measure.Description
	}
	return nil
}

// getMeasureSQL generates the SQL expression for the measure
func (g *SemanticMeasureGenerator) getMeasureSQL(
	measure *models.DbtSemanticMeasure,
	model *models.DbtModel,
) *string {
	// Count measures typically don't need SQL
	if measure.Agg == "count" {
		return nil
	}

	// If expr is provided, use it
	if measure.Expr != nil && *measure.Expr != "" {
		sql := g.buildSQLFromExpr(*measure.Expr, measure.Agg)
		return &sql
	}

	// For other aggregations without expr, return nil
	// Looker will require SQL to be provided in the YAML or generate an error
	return nil
}

// buildSQLFromExpr builds the SQL expression from the semantic model expr
func (g *SemanticMeasureGenerator) buildSQLFromExpr(expr string, agg string) string {
	// Check if expr is already a reference (starts with ${})
	if strings.HasPrefix(expr, "${") {
		return expr
	}

	// Build LookML-style reference
	// Convert column name to dimension reference
	columnRef := fmt.Sprintf("${TABLE}.%s", strings.ToLower(expr))

	// Special handling for sum_boolean
	if agg == "sum_boolean" {
		// Cast boolean to integer for summing
		return fmt.Sprintf("CAST(%s AS INT64)", columnRef)
	}

	return columnRef
}

// HasSemanticMeasureWithName checks if a semantic measure with the given name exists
func (g *SemanticMeasureGenerator) HasSemanticMeasureWithName(
	semanticMeasures []models.DbtSemanticMeasure,
	name string,
) bool {
	for _, measure := range semanticMeasures {
		if measure.Name == name {
			return true
		}
	}
	return false
}

// MergeWithMetaMeasures merges semantic model measures with meta-defined measures
// Semantic model measures always take precedence over meta measures with the same name
func (g *SemanticMeasureGenerator) MergeWithMetaMeasures(
	semanticMeasures []*models.LookMLMeasure,
	metaMeasures []*models.LookMLMeasure,
) []*models.LookMLMeasure {
	result := make([]*models.LookMLMeasure, 0, len(semanticMeasures)+len(metaMeasures))

	// Add all semantic measures first
	result = append(result, semanticMeasures...)

	// Track semantic measure names
	semanticNames := make(map[string]bool)
	for _, m := range semanticMeasures {
		semanticNames[m.Name] = true
	}

	// Add meta measures that don't conflict with semantic measures
	for _, metaMeasure := range metaMeasures {
		if !semanticNames[metaMeasure.Name] {
			result = append(result, metaMeasure)
		} else {
			g.config.Logger().Debug().
				Str("measure", metaMeasure.Name).
				Msg("Skipping meta measure (overridden by semantic model measure)")
		}
	}

	return result
}
