package generators

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// Compile-time check to ensure MeasureGenerator implements MeasureGeneratorInterface
var _ MeasureGeneratorInterface = (*MeasureGenerator)(nil)

// MeasureGenerator handles generation of LookML measures
type MeasureGenerator struct {
	config *config.Config
}

// NewMeasureGenerator creates a new MeasureGenerator instance
func NewMeasureGenerator(cfg *config.Config) *MeasureGenerator {
	return &MeasureGenerator{
		config: cfg,
	}
}

// GenerateMeasure generates a LookML measure from measure metadata
func (g *MeasureGenerator) GenerateMeasure(model *models.DbtModel, measureMeta *models.DbtMetaLookerMeasure) (*models.LookMLMeasure, error) {
	// Validate measure attributes
	if err := measureMeta.ValidateMeasureAttributes(); err != nil {
		return nil, fmt.Errorf("invalid measure attributes: %w", err)
	}

	measure := &models.LookMLMeasure{
		Name:                 g.getMeasureName(measureMeta),
		Type:                 measureMeta.Type,
		SQL:                  g.getMeasureSQL(model, measureMeta),
		Label:                g.getMeasureLabel(measureMeta),
		Description:          g.getMeasureDescription(measureMeta),
		Hidden:               g.getMeasureHidden(measureMeta),
		GroupLabel:           measureMeta.GroupLabel,
		ValueFormatName:      measureMeta.ValueFormatName,
		Approximate:          measureMeta.Approximate,
		ApproximateThreshold: measureMeta.ApproximateThreshold,
		Precision:            measureMeta.Precision,
		SQLDistinctKey:       measureMeta.SQLDistinctKey,
		Percentile:           measureMeta.Percentile,
		Filters:              measureMeta.Filters,
	}

	return measure, nil
}

// GenerateDefaultCountMeasure generates a default count measure for a model
func (g *MeasureGenerator) GenerateDefaultCountMeasure(model *models.DbtModel) *models.LookMLMeasure {
	measureName := "count"

	// Check if there's already a count measure defined in meta
	if model.Meta != nil && model.Meta.Looker != nil {
		for _, measureMeta := range model.Meta.Looker.Measures {
			if measureMeta.Name != nil && *measureMeta.Name == measureName {
				// Count measure already defined in meta, don't generate default
				return nil
			}
		}
	}

	// Default count measure should be minimal - only name and type (matches Python implementation)
	return &models.LookMLMeasure{
		Name: measureName,
		Type: enums.MeasureCount,
	}
}

// getMeasureName gets the measure name
func (g *MeasureGenerator) getMeasureName(measureMeta *models.DbtMetaLookerMeasure) string {
	if measureMeta.Name != nil {
		return *measureMeta.Name
	}

	// Generate name from type
	return string(measureMeta.Type)
}

// getMeasureSQL gets the SQL expression for the measure
func (g *MeasureGenerator) getMeasureSQL(model *models.DbtModel, measureMeta *models.DbtMetaLookerMeasure) *string {
	// For most measure types, SQL is optional and Looker will generate it
	// Only return SQL if it's explicitly needed or provided

	switch measureMeta.Type {
	case enums.MeasureCount:
		// Count measures don't need SQL
		return nil
	case enums.MeasureCountDistinct:
		if measureMeta.SQLDistinctKey != nil {
			sql := fmt.Sprintf("${TABLE}.%s", strings.ToLower(*measureMeta.SQLDistinctKey))
			return &sql
		}
		return nil
	case enums.MeasureSum, enums.MeasureAverage, enums.MeasureMin, enums.MeasureMax:
		// These measures need a column to aggregate
		// This would need to be specified in the measure metadata
		return nil
	default:
		return nil
	}
}

// getMeasureLabel gets the measure label
func (g *MeasureGenerator) getMeasureLabel(measureMeta *models.DbtMetaLookerMeasure) *string {
	if measureMeta.Label != nil {
		return measureMeta.Label
	}

	// Generate label from name or type
	var label string
	if measureMeta.Name != nil {
		label = *measureMeta.Name
	} else {
		label = string(measureMeta.Type)
	}

	// Convert to title case
	label = g.toTitleCase(label)
	return &label
}

// getMeasureDescription gets the measure description
func (g *MeasureGenerator) getMeasureDescription(measureMeta *models.DbtMetaLookerMeasure) *string {
	return measureMeta.Description
}

// getMeasureHidden gets the measure hidden setting
func (g *MeasureGenerator) getMeasureHidden(measureMeta *models.DbtMetaLookerMeasure) *bool {
	return measureMeta.Hidden
}

// toTitleCase converts a string to title case
func (g *MeasureGenerator) toTitleCase(s string) string {
	// Simple title case conversion
	if len(s) == 0 {
		return s
	}

	// Convert first character to uppercase
	result := string(s[0])
	if result >= "a" && result <= "z" {
		result = string(s[0] - 32)
	}

	// Add the rest of the string
	if len(s) > 1 {
		result += s[1:]
	}

	return result
}

// GeneratePrimaryKeyMeasure generates a count distinct measure for primary key columns
func (g *MeasureGenerator) GeneratePrimaryKeyMeasure(model *models.DbtModel, pkColumn *models.DbtModelColumn) *models.LookMLMeasure {
	measureName := fmt.Sprintf("count_distinct_%s", *pkColumn.LookMLName)
	label := fmt.Sprintf("Count Distinct %s", *pkColumn.LookMLName)
	description := fmt.Sprintf("Count of distinct %s values", *pkColumn.LookMLName)

	sql := fmt.Sprintf("${TABLE}.%s", strings.ToLower(pkColumn.Name))

	return &models.LookMLMeasure{
		Name:        measureName,
		Type:        enums.MeasureCountDistinct,
		SQL:         &sql,
		Label:       &label,
		Description: &description,
	}
}

// GenerateNumericMeasures generates sum/average measures for numeric columns
func (g *MeasureGenerator) GenerateNumericMeasures(model *models.DbtModel, column *models.DbtModelColumn) []*models.LookMLMeasure {
	if column.DataType == nil {
		return nil
	}

	// Check if column is numeric
	dataType := *column.DataType
	if !g.isNumericType(dataType) {
		return nil
	}

	var measures []*models.LookMLMeasure
	columnName := *column.LookMLName
	sql := fmt.Sprintf("${TABLE}.%s", strings.ToLower(column.Name))

	// Generate sum measure
	sumName := fmt.Sprintf("sum_%s", columnName)
	sumLabel := fmt.Sprintf("Sum %s", columnName)
	sumDescription := fmt.Sprintf("Sum of %s values", columnName)

	sumMeasure := &models.LookMLMeasure{
		Name:        sumName,
		Type:        enums.MeasureSum,
		SQL:         &sql,
		Label:       &sumLabel,
		Description: &sumDescription,
	}
	measures = append(measures, sumMeasure)

	// Generate average measure
	avgName := fmt.Sprintf("avg_%s", columnName)
	avgLabel := fmt.Sprintf("Average %s", columnName)
	avgDescription := fmt.Sprintf("Average of %s values", columnName)

	avgMeasure := &models.LookMLMeasure{
		Name:        avgName,
		Type:        enums.MeasureAverage,
		SQL:         &sql,
		Label:       &avgLabel,
		Description: &avgDescription,
	}
	measures = append(measures, avgMeasure)

	return measures
}

// isNumericType checks if a data type is numeric
func (g *MeasureGenerator) isNumericType(dataType string) bool {
	numericTypes := []string{
		"INT64", "INTEGER", "FLOAT", "FLOAT64",
		"NUMERIC", "DECIMAL", "BIGNUMERIC",
	}

	for _, numType := range numericTypes {
		if dataType == numType {
			return true
		}
	}

	return false
}
