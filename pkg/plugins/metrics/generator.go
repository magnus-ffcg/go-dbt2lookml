package metrics

import (
	"fmt"
	"os"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// Measure type constants
const (
	measureTypeAverage       = "average"
	measureTypeSum           = "sum"
	measureTypeCount         = "count"
	measureTypeMin           = "min"
	measureTypeMax           = "max"
	measureTypeCountDistinct = "count_distinct"
	measureTypeSumBoolean    = "sum_boolean"
)

// GenerateForModel generates all metric-related files for a model
// This includes view extensions, cumulative views, and conversion views
func (p *MetricsPlugin) GenerateForModel(model *models.DbtModel) error {
	if !p.Enabled() {
		return nil // Plugin disabled, skip
	}

	// Get semantic measures for this model
	semanticMeasures := p.GetSemanticMeasures(model.Name)
	if len(semanticMeasures) == 0 {
		return nil // No semantic measures, nothing to generate
	}

	// Generate view extension (semantic measures + simple/ratio/derived metrics)
	if err := p.GenerateViewExtension(model, semanticMeasures); err != nil {
		return fmt.Errorf("failed to generate view extension: %w", err)
	}

	// Generate cumulative metrics view if applicable
	if err := p.generateCumulativeViewFile(model, semanticMeasures); err != nil {
		return fmt.Errorf("failed to generate cumulative view: %w", err)
	}

	// Generate conversion metrics view if applicable
	if err := p.generateConversionViewFile(model, semanticMeasures); err != nil {
		return fmt.Errorf("failed to generate conversion view: %w", err)
	}

	return nil
}

// GenerateViewExtension generates a view extension file containing semantic layer content
// This includes: semantic measures, simple metrics, ratio metrics, and derived metrics
func (p *MetricsPlugin) GenerateViewExtension(model *models.DbtModel, semanticMeasures []models.DbtSemanticMeasure) error {
	// Check if we have metrics
	hasMetrics := len(p.simpleMetrics) > 0 || len(p.ratioMetrics) > 0 || len(p.derivedMetrics) > 0

	// If no semantic measures AND no metrics, skip extension
	if len(semanticMeasures) == 0 && !hasMetrics {
		return nil
	}

	var extension strings.Builder

	// 1. Add include statement
	extension.WriteString(fmt.Sprintf("include: \"%s.view.lkml\"\n\n", model.Name))

	// 2. Start view extension
	extension.WriteString(fmt.Sprintf("view: +%s {\n", model.Name))

	// 3. Add all semantic measures
	for _, semanticMeasure := range semanticMeasures {
		measureLookML := p.generateSemanticMeasureLookML(&semanticMeasure)
		extension.WriteString(measureLookML)
		extension.WriteString("\n")
	}

	// 4. Generate metric measures (simple with filters, ratio, derived)
	if hasMetrics && p.metricMeasureGenerator != nil {
		// Build existing measures list from semantic measures for metric references
		existingMeasures := make([]models.LookMLMeasure, 0, len(semanticMeasures))
		for _, sm := range semanticMeasures {
			measure := models.LookMLMeasure{
				Name:        sm.Name,
				Type:        enums.LookerMeasureType(sm.Agg),
				SQL:         sm.Expr,
				Label:       sm.Label,
				Description: sm.Description,
			}
			existingMeasures = append(existingMeasures, measure)
		}

		ctx := &GenerationContext{
			SemanticMeasures: semanticMeasures,
			ExistingMeasures: existingMeasures,
		}

		metricMeasures, err := p.metricMeasureGenerator.GenerateMetricMeasures(
			p.simpleMetrics,
			p.ratioMetrics,
			p.derivedMetrics,
			ctx,
		)

		if err != nil {
			p.config.Logger().Warn().
				Err(err).
				Str("model", model.Name).
				Msg("Failed to generate metric measures")
		} else {
			for _, measure := range metricMeasures {
				measureLookML := p.generateMeasureLookML(measure)
				extension.WriteString(measureLookML)
				extension.WriteString("\n")
			}
		}
	}

	// 5. Close view extension
	extension.WriteString("}\n")

	// 6. Write to file
	filename := fmt.Sprintf("%s__metrics.view.lkml", model.Name)
	filePath := p.config.GetOutputPath(filename)

	if err := os.WriteFile(filePath, []byte(extension.String()), filePermissions); err != nil {
		return fmt.Errorf("failed to write view extension file: %w", err)
	}

	p.config.Logger().Info().
		Str("file", filePath).
		Msg("Generated view extension file")

	return nil
}

// generateSemanticMeasureLookML generates LookML for a semantic measure
func (p *MetricsPlugin) generateSemanticMeasureLookML(measure *models.DbtSemanticMeasure) string {
	var lookML strings.Builder

	lookML.WriteString(fmt.Sprintf("  measure: %s {\n", measure.Name))
	lookML.WriteString(fmt.Sprintf("    type: %s\n", measure.Agg))

	if measure.Expr != nil && *measure.Expr != "" {
		lookML.WriteString(fmt.Sprintf("    sql: %s ;;\n", *measure.Expr))
	}

	if measure.Label != nil {
		lookML.WriteString(fmt.Sprintf("    label: \"%s\"\n", *measure.Label))
	}

	if measure.Description != nil {
		lookML.WriteString(fmt.Sprintf("    description: \"%s\"\n", *measure.Description))
	}

	lookML.WriteString("  }\n")
	return lookML.String()
}

// generateCumulativeViewFile generates a separate view file for cumulative metrics
func (p *MetricsPlugin) generateCumulativeViewFile(
	model *models.DbtModel,
	semanticMeasures []models.DbtSemanticMeasure,
) error {
	// Filter cumulative metrics for this model
	var modelCumulativeMetrics []models.DbtMetric
	baseMeasureMap := make(map[string]*models.DbtSemanticMeasure)

	for _, sm := range semanticMeasures {
		baseMeasureMap[sm.Name] = &sm
	}

	for _, metric := range p.cumulativeMetrics {
		if metric.TypeParams.Measure != nil {
			measureName := metric.TypeParams.Measure.Name
			if _, exists := baseMeasureMap[measureName]; exists {
				modelCumulativeMetrics = append(modelCumulativeMetrics, metric)
			}
		}
	}

	if len(modelCumulativeMetrics) == 0 {
		return nil
	}

	var viewContent strings.Builder

	// 1. Add include statement
	viewContent.WriteString(fmt.Sprintf("include: \"%s.view.lkml\"\n\n", model.Name))

	// 2. Start view
	viewContent.WriteString(fmt.Sprintf("view: %s__cumulative {\n", model.Name))

	// 3. Add derived table
	viewContent.WriteString("  derived_table: {\n")
	viewContent.WriteString("    sql:\n")
	viewContent.WriteString("      SELECT\n")

	// Get primary key and time dimension from semantic model
	// For now, assume order_id and order_date - TODO: get from semantic model
	viewContent.WriteString("        order_id,\n")
	viewContent.WriteString("        order_date,\n")
	viewContent.WriteString("        customer_id,\n")

	// Add window function calculations
	for i, metric := range modelCumulativeMetrics {
		measureName := metric.TypeParams.Measure.Name
		baseMeasure := baseMeasureMap[measureName]

		windowSQL := p.buildWindowFunctionSQL(&metric, baseMeasure)
		comma := ","
		if i == len(modelCumulativeMetrics)-1 {
			comma = ""
		}
		viewContent.WriteString(fmt.Sprintf("        %s as %s%s\n", windowSQL, metric.Name, comma))
	}

	viewContent.WriteString(fmt.Sprintf("      FROM ${%s.SQL_TABLE_NAME}\n", model.Name))
	viewContent.WriteString("      ;;\n")
	viewContent.WriteString("  }\n\n")

	// 4. Add primary key dimension (hidden, for joining)
	viewContent.WriteString("  dimension: order_id {\n")
	viewContent.WriteString("    primary_key: yes\n")
	viewContent.WriteString("    hidden: yes\n")
	viewContent.WriteString("    sql: ${TABLE}.order_id ;;\n")
	viewContent.WriteString("  }\n\n")

	// 5. Add measures
	for _, metric := range modelCumulativeMetrics {
		measure := &models.LookMLMeasure{
			Name: metric.Name,
			Type: enums.MeasureSum,
		}

		sql := fmt.Sprintf("${TABLE}.%s", metric.Name)
		measure.SQL = &sql

		if metric.Label != nil {
			measure.Label = metric.Label
		}

		if metric.Description != nil {
			measure.Description = metric.Description
		}

		measureLookML := p.generateMeasureLookML(measure)
		viewContent.WriteString(measureLookML)
		viewContent.WriteString("\n")
	}

	// 6. Close view
	viewContent.WriteString("}\n")

	// 7. Write to file
	filename := fmt.Sprintf("%s__cumulative.view.lkml", model.Name)
	filePath := p.config.GetOutputPath(filename)

	if err := os.WriteFile(filePath, []byte(viewContent.String()), filePermissions); err != nil {
		return fmt.Errorf("failed to write cumulative view file: %w", err)
	}

	p.config.Logger().Info().
		Str("file", filePath).
		Int("metrics", len(modelCumulativeMetrics)).
		Msg("Generated cumulative metrics view file")

	return nil
}

// buildWindowFunctionSQL creates SQL window function for a cumulative metric
func (p *MetricsPlugin) buildWindowFunctionSQL(
	metric *models.DbtMetric,
	baseMeasure *models.DbtSemanticMeasure,
) string {
	// Get aggregation type
	aggType := strings.ToUpper(string(baseMeasure.Agg))

	// Get column expression
	column := "*"
	if baseMeasure.Expr != nil && *baseMeasure.Expr != "" {
		column = *baseMeasure.Expr
	}

	// Build OVER clause
	overClause := "ORDER BY order_date" // TODO: Get time dimension from semantic model

	// Check for window parameters
	if metric.TypeParams.CumulativeTypeParams != nil {
		params := metric.TypeParams.CumulativeTypeParams

		// grain_to_date (e.g., month-to-date)
		if params.GrainToDate != nil && *params.GrainToDate != "" {
			grain := *params.GrainToDate
			partitionBy := fmt.Sprintf("DATE_TRUNC(order_date, %s)", strings.ToUpper(grain))
			overClause = fmt.Sprintf("PARTITION BY %s %s", partitionBy, overClause)
		} else if params.Window != nil && params.Window.Count > 0 {
			// Rolling window
			rowsBetween := fmt.Sprintf("ROWS BETWEEN %d PRECEDING AND CURRENT ROW", params.Window.Count-1)
			overClause = fmt.Sprintf("%s %s", overClause, rowsBetween)
		}
	}

	return fmt.Sprintf("%s(%s) OVER (%s)", aggType, column, overClause)
}

// generateMeasureLookML generates LookML for a measure
func (p *MetricsPlugin) generateMeasureLookML(measure *models.LookMLMeasure) string {
	var lookML strings.Builder

	lookML.WriteString(fmt.Sprintf("  measure: %s {\n", measure.Name))
	lookML.WriteString(fmt.Sprintf("    type: %s\n", measure.Type))

	if measure.SQL != nil {
		lookML.WriteString(fmt.Sprintf("    sql: %s ;;\n", *measure.SQL))
	}

	if measure.Label != nil {
		lookML.WriteString(fmt.Sprintf("    label: \"%s\"\n", *measure.Label))
	}

	if measure.Description != nil {
		lookML.WriteString(fmt.Sprintf("    description: \"%s\"\n", *measure.Description))
	}

	lookML.WriteString("  }\n")
	return lookML.String()
}

// generateConversionViewFile generates a separate view file for conversion metrics
func (p *MetricsPlugin) generateConversionViewFile(
	model *models.DbtModel,
	semanticMeasures []models.DbtSemanticMeasure,
) error {
	// Filter conversion metrics for this model
	var modelConversionMetrics []models.DbtMetric
	baseMeasureMap := make(map[string]*models.DbtSemanticMeasure)

	for _, sm := range semanticMeasures {
		baseMeasureMap[sm.Name] = &sm
	}

	for _, metric := range p.conversionMetrics {
		if metric.TypeParams.ConversionTypeParams != nil {
			params := metric.TypeParams.ConversionTypeParams
			baseMeasureName := params.BaseMeasure.Name
			conversionMeasureName := params.ConversionMeasure.Name

			// Check if both measures exist in this model
			if _, hasBase := baseMeasureMap[baseMeasureName]; hasBase {
				if _, hasConv := baseMeasureMap[conversionMeasureName]; hasConv {
					modelConversionMetrics = append(modelConversionMetrics, metric)
				}
			}
		}
	}

	if len(modelConversionMetrics) == 0 {
		return nil
	}

	var viewContent strings.Builder

	// 1. Add include statement
	viewContent.WriteString(fmt.Sprintf("include: \"%s.view.lkml\"\n\n", model.Name))

	// 2. Start view
	viewContent.WriteString(fmt.Sprintf("view: %s__conversion {\n", model.Name))

	// 3. Add derived table with MetricFlow-style SQL
	viewContent.WriteString("  derived_table: {\n")
	viewContent.WriteString("    sql:\n")

	// Generate SQL for each conversion metric
	for i, metric := range modelConversionMetrics {
		params := metric.TypeParams.ConversionTypeParams

		// Get entity column (e.g., "customer" -> "customer_id")
		entityColumn := fmt.Sprintf("%s_id", params.Entity)

		// TODO: Get time column from semantic model - for now assume order_date
		timeColumn := "order_date"

		// Add WITH clause for first metric, comma for subsequent
		if i == 0 {
			viewContent.WriteString("      WITH ")
		} else {
			viewContent.WriteString("      ,")
		}

		// Generate CTEs for this metric
		viewContent.WriteString(fmt.Sprintf("%s_base_events AS (\n", metric.Name))
		viewContent.WriteString(fmt.Sprintf("        -- Base: First time each %s appears\n", params.Entity))
		viewContent.WriteString("        SELECT\n")
		viewContent.WriteString(fmt.Sprintf("          %s as entity_id,\n", entityColumn))
		viewContent.WriteString(fmt.Sprintf("          MIN(%s) as base_time\n", timeColumn))
		viewContent.WriteString(fmt.Sprintf("        FROM ${%s.SQL_TABLE_NAME}\n", model.Name))
		viewContent.WriteString(fmt.Sprintf("        GROUP BY %s\n", entityColumn))
		viewContent.WriteString("      ),\n")

		viewContent.WriteString(fmt.Sprintf("      %s_conversion_events AS (\n", metric.Name))
		viewContent.WriteString(fmt.Sprintf("        -- Conversion: Subsequent times %s appears\n", params.Entity))
		viewContent.WriteString("        SELECT\n")
		viewContent.WriteString(fmt.Sprintf("          t.%s as entity_id,\n", entityColumn))
		viewContent.WriteString(fmt.Sprintf("          MIN(t.%s) as conversion_time\n", timeColumn))
		viewContent.WriteString(fmt.Sprintf("        FROM ${%s.SQL_TABLE_NAME} t\n", model.Name))
		viewContent.WriteString(fmt.Sprintf("        INNER JOIN %s_base_events b\n", metric.Name))
		viewContent.WriteString(fmt.Sprintf("          ON t.%s = b.entity_id\n", entityColumn))
		viewContent.WriteString(fmt.Sprintf("          AND t.%s > b.base_time\n", timeColumn))
		viewContent.WriteString(fmt.Sprintf("        GROUP BY t.%s\n", entityColumn))
		viewContent.WriteString("      )\n")
	}

	// Final SELECT combining all metrics
	viewContent.WriteString("      SELECT\n")

	// Get entity column from first metric for the SELECT
	firstMetric := modelConversionMetrics[0]
	firstParams := firstMetric.TypeParams.ConversionTypeParams
	entityColumn := fmt.Sprintf("%s_id", firstParams.Entity)

	viewContent.WriteString(fmt.Sprintf("        b.entity_id as %s,\n", entityColumn))

	// Add conversion flag for each metric
	for i, metric := range modelConversionMetrics {
		params := metric.TypeParams.ConversionTypeParams
		windowInterval := "INTERVAL 30 DAY"
		if params.Window != nil {
			windowInterval = fmt.Sprintf("INTERVAL %d %s",
				params.Window.Count,
				strings.ToUpper(params.Window.Granularity))
		}

		comma := ","
		if i == len(modelConversionMetrics)-1 {
			comma = ""
		}

		viewContent.WriteString("        CASE\n")
		viewContent.WriteString(fmt.Sprintf("          WHEN c%d.conversion_time <= DATE_ADD(b.base_time, %s)\n", i, windowInterval))
		viewContent.WriteString("          THEN 1 ELSE 0\n")
		viewContent.WriteString(fmt.Sprintf("        END as %s%s\n", metric.Name, comma))
	}

	viewContent.WriteString(fmt.Sprintf("      FROM %s_base_events b\n", firstMetric.Name))

	// LEFT JOIN all conversion event CTEs
	for i, metric := range modelConversionMetrics {
		viewContent.WriteString(fmt.Sprintf("      LEFT JOIN %s_conversion_events c%d\n", metric.Name, i))
		viewContent.WriteString(fmt.Sprintf("        ON b.entity_id = c%d.entity_id\n", i))
	}

	viewContent.WriteString("      ;;\n")
	viewContent.WriteString("  }\n\n")

	// 5. Add primary key dimension (hidden, for joining)
	viewContent.WriteString("  dimension: customer_id {\n")
	viewContent.WriteString("    primary_key: yes\n")
	viewContent.WriteString("    hidden: yes\n")
	viewContent.WriteString("    sql: ${TABLE}.customer_id ;;\n")
	viewContent.WriteString("  }\n\n")

	// 6. Add measures for conversion metrics
	for _, metric := range modelConversionMetrics {
		params := metric.TypeParams.ConversionTypeParams

		viewContent.WriteString(fmt.Sprintf("  measure: %s {\n", metric.Name))

		// Determine measure type based on calculation
		measureType := measureTypeAverage // Default for conversion_rate
		if params.Calculation != nil {
			calc := *params.Calculation
			if calc == "conversions" || calc == "converted_entity_count" {
				measureType = measureTypeSum
			}
		}

		viewContent.WriteString(fmt.Sprintf("    type: %s\n", measureType))
		viewContent.WriteString(fmt.Sprintf("    sql: ${TABLE}.%s ;;\n", metric.Name))

		if metric.Label != nil {
			viewContent.WriteString(fmt.Sprintf("    label: \"%s\"\n", *metric.Label))
		}

		if metric.Description != nil {
			viewContent.WriteString(fmt.Sprintf("    description: \"%s\"\n", *metric.Description))
		}

		// Add value format for rates
		if measureType == measureTypeAverage {
			viewContent.WriteString("    value_format_name: percent_1\n")
		}

		viewContent.WriteString("  }\n\n")
	}

	// 7. Close view
	viewContent.WriteString("}\n")

	// 8. Write to file
	filename := fmt.Sprintf("%s__conversion.view.lkml", model.Name)
	filePath := p.config.GetOutputPath(filename)

	if err := os.WriteFile(filePath, []byte(viewContent.String()), filePermissions); err != nil {
		return fmt.Errorf("failed to write conversion view file: %w", err)
	}

	p.config.Logger().Info().
		Str("file", filePath).
		Int("metrics", len(modelConversionMetrics)).
		Msg("Generated conversion metrics view file with MetricFlow-style funnel SQL")

	return nil
}
