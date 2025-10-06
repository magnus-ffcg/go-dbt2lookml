package metrics

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// CumulativeMetricGenerator handles generation of LookML view extensions
// with derived tables for dbt cumulative metrics
// Note: This is used as a builder by MetricMeasureGenerator
type CumulativeMetricGenerator struct {
	config *config.Config
}

// NewCumulativeMetricGenerator creates a new CumulativeMetricGenerator instance
func NewCumulativeMetricGenerator(cfg *config.Config) *CumulativeMetricGenerator {
	return &CumulativeMetricGenerator{
		config: cfg,
	}
}

// ViewExtension represents a LookML view extension with cumulative metrics
type ViewExtension struct {
	BaseViewName       string
	DerivedTableSQL    string
	CumulativeMeasures []*models.LookMLMeasure
	TimeDimension      string
}

// GenerateViewExtension generates a view extension for cumulative metrics
// grouped by their base semantic model
func (g *CumulativeMetricGenerator) GenerateViewExtension(
	metrics []models.DbtMetric,
	semanticModel *models.DbtSemanticModel,
	baseMeasures map[string]*models.DbtSemanticMeasure,
) (*ViewExtension, error) {

	if len(metrics) == 0 {
		return nil, fmt.Errorf("no cumulative metrics provided")
	}

	// Get the time dimension from semantic model defaults
	timeDimension := g.getTimeDimension(semanticModel)
	if timeDimension == "" {
		return nil, fmt.Errorf("no time dimension found in semantic model")
	}

	// Build the derived table SQL with all cumulative calculations
	derivedSQL, err := g.buildDerivedTableSQL(metrics, semanticModel, baseMeasures, timeDimension)
	if err != nil {
		return nil, fmt.Errorf("failed to build derived table SQL: %w", err)
	}

	// Generate LookML measures for each cumulative metric
	measures := make([]*models.LookMLMeasure, 0, len(metrics))
	for _, metric := range metrics {
		measure, err := g.generateCumulativeMeasure(&metric)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("metric", metric.Name).
				Msg("Failed to generate cumulative measure")
			continue
		}
		measures = append(measures, measure)
	}

	return &ViewExtension{
		BaseViewName:       semanticModel.Model,
		DerivedTableSQL:    derivedSQL,
		CumulativeMeasures: measures,
		TimeDimension:      timeDimension,
	}, nil
}

// getTimeDimension extracts the primary time dimension from semantic model
func (g *CumulativeMetricGenerator) getTimeDimension(model *models.DbtSemanticModel) string {
	// Check defaults.agg_time_dimension first
	if model.Defaults != nil && model.Defaults.AggTimeDimension != "" {
		return model.Defaults.AggTimeDimension
	}

	// Fallback: find first dimension with type: time
	for _, dim := range model.Dimensions {
		if dim.Type == "time" {
			return dim.Name
		}
	}

	return ""
}

// buildDerivedTableSQL generates SQL for the derived table with window functions
func (g *CumulativeMetricGenerator) buildDerivedTableSQL(
	metrics []models.DbtMetric,
	semanticModel *models.DbtSemanticModel,
	baseMeasures map[string]*models.DbtSemanticMeasure,
	timeDimension string,
) (string, error) {

	var sql strings.Builder

	// Start SELECT
	sql.WriteString("SELECT\n")

	// Add all dimensions (to preserve grain)
	sql.WriteString(fmt.Sprintf("  %s,\n", timeDimension))

	// Add entity columns
	for _, entity := range semanticModel.Entities {
		if entity.Expr != nil && *entity.Expr != "" {
			sql.WriteString(fmt.Sprintf("  %s,\n", *entity.Expr))
		}
	}

	// Add other key dimensions
	for _, dim := range semanticModel.Dimensions {
		if dim.Name != timeDimension && dim.Expr != nil {
			sql.WriteString(fmt.Sprintf("  %s as %s,\n", *dim.Expr, dim.Name))
		}
	}

	// Add cumulative calculations as window functions
	for i, metric := range metrics {
		measureName := metric.TypeParams.Measure.Name
		baseMeasure, exists := baseMeasures[measureName]
		if !exists {
			continue
		}

		windowSQL, err := g.buildWindowFunction(&metric, baseMeasure, timeDimension)
		if err != nil {
			return "", err
		}

		// Add comma for all but last
		comma := ","
		if i == len(metrics)-1 {
			comma = ""
		}
		sql.WriteString(fmt.Sprintf("  %s as %s%s\n", windowSQL, metric.Name, comma))
	}

	// FROM clause - reference the base view's SQL table
	sql.WriteString(fmt.Sprintf("FROM ${%s.SQL_TABLE_NAME}", semanticModel.Model))

	return sql.String(), nil
}

// buildWindowFunction creates the SQL window function for a cumulative metric
func (g *CumulativeMetricGenerator) buildWindowFunction(
	metric *models.DbtMetric,
	baseMeasure *models.DbtSemanticMeasure,
	timeDimension string,
) (string, error) {

	// Get the aggregation type
	aggType := strings.ToUpper(string(baseMeasure.Agg))

	// Get the column expression
	column := g.getColumnExpr(baseMeasure)

	// Build OVER clause
	overClause := g.buildOverClause(metric, timeDimension)

	return fmt.Sprintf("%s(%s) OVER (%s)", aggType, column, overClause), nil
}

// getColumnExpr extracts the column expression from a semantic measure
func (g *CumulativeMetricGenerator) getColumnExpr(measure *models.DbtSemanticMeasure) string {
	if measure.Expr != nil && *measure.Expr != "" {
		return *measure.Expr
	}
	return "*" // For COUNT without expression
}

// buildOverClause builds the OVER clause for window function
func (g *CumulativeMetricGenerator) buildOverClause(
	metric *models.DbtMetric,
	timeDimension string,
) string {

	orderBy := fmt.Sprintf("ORDER BY %s", timeDimension)

	// Check for window parameters
	if metric.TypeParams.CumulativeTypeParams == nil {
		// No window params = unbounded (lifetime cumulative)
		return orderBy
	}

	params := metric.TypeParams.CumulativeTypeParams

	// grain_to_date (e.g., month-to-date, year-to-date)
	if params.GrainToDate != nil && *params.GrainToDate != "" {
		// For grain-to-date, we need to partition by the grain
		grain := *params.GrainToDate
		partitionBy := g.getPartitionByForGrain(grain, timeDimension)
		return fmt.Sprintf("PARTITION BY %s %s", partitionBy, orderBy)
	}

	// window (e.g., "7 days", "30 days")
	if params.Window != nil {
		window := params.Window
		if window.Count > 0 {
			// Rolling window: ROWS BETWEEN n PRECEDING AND CURRENT ROW
			rowsBetween := fmt.Sprintf("ROWS BETWEEN %d PRECEDING AND CURRENT ROW", window.Count-1)
			return fmt.Sprintf("%s %s", orderBy, rowsBetween)
		}
	}

	// Default: unbounded window (all-time cumulative)
	return orderBy
}

// getPartitionByForGrain returns partition expression for grain-to-date
func (g *CumulativeMetricGenerator) getPartitionByForGrain(grain string, timeDimension string) string {
	switch strings.ToLower(grain) {
	case "day":
		return fmt.Sprintf("DATE_TRUNC('%s', DAY)", timeDimension)
	case "week":
		return fmt.Sprintf("DATE_TRUNC('%s', WEEK)", timeDimension)
	case "month":
		return fmt.Sprintf("DATE_TRUNC('%s', MONTH)", timeDimension)
	case "quarter":
		return fmt.Sprintf("DATE_TRUNC('%s', QUARTER)", timeDimension)
	case "year":
		return fmt.Sprintf("DATE_TRUNC('%s', YEAR)", timeDimension)
	default:
		return fmt.Sprintf("DATE_TRUNC('%s', %s)", timeDimension, strings.ToUpper(grain))
	}
}

// generateCumulativeMeasure creates a LookML measure for a cumulative metric
func (g *CumulativeMetricGenerator) generateCumulativeMeasure(
	metric *models.DbtMetric,
) (*models.LookMLMeasure, error) {

	measure := &models.LookMLMeasure{
		Name: metric.Name,
		Type: "sum", // Cumulative values are pre-calculated in derived table
	}

	// SQL references the pre-calculated column
	sql := fmt.Sprintf("${TABLE}.%s", metric.Name)
	measure.SQL = &sql

	// Add label
	utils := &MeasureUtils{}
	measure.Label = utils.GetMetricLabel(metric)
	measure.Description = utils.GetMetricDescription(metric)

	return measure, nil
}
