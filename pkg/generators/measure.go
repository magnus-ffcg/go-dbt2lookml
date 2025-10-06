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

const (
	// DefaultCountMeasureName is the name of the default count measure
	DefaultCountMeasureName = "count"
)

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

// GenerateMeasure is deprecated - use semantic models instead
// Kept for backward compatibility but returns nil
func (g *MeasureGenerator) GenerateMeasure(model *models.DbtModel, measureMeta interface{}) (*models.LookMLMeasure, error) {
	// Meta measures are no longer supported - use semantic models
	return nil, fmt.Errorf("meta.looker.measures is deprecated - use dbt semantic models instead")
}

// GenerateDefaultCountMeasure generates a default count measure for a model
func (g *MeasureGenerator) GenerateDefaultCountMeasure(model *models.DbtModel) *models.LookMLMeasure {
	// Default count measure should be minimal - only name and type (matches Python implementation)
	return &models.LookMLMeasure{
		Name: DefaultCountMeasureName,
		Type: enums.MeasureCount,
	}
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
