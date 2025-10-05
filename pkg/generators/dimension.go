package generators

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
)

// Date/time data type constants
const (
	dataTypeDate      = "DATE"
	dataTypeDateTime  = "DATETIME"
	dataTypeTimestamp = "TIMESTAMP"
	dimGroupTypeTime  = "time"
)

// Compile-time check to ensure DimensionGenerator implements DimensionGeneratorInterface
var _ DimensionGeneratorInterface = (*DimensionGenerator)(nil)

// DimensionGenerator handles generation of LookML dimensions and dimension groups
type DimensionGenerator struct {
	config *config.Config
}

// NewDimensionGenerator creates a new DimensionGenerator instance
func NewDimensionGenerator(cfg *config.Config) *DimensionGenerator {
	return &DimensionGenerator{
		config: cfg,
	}
}

// GenerateDimension generates a LookML dimension from a model column
func (g *DimensionGenerator) GenerateDimension(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimension, error) {
	// Skip date/time columns - they will be dimension_groups
	if g.shouldBeDimensionGroup(column) {
		return nil, nil
	}

	// Generate dimension name - special handling for ARRAY columns in main view
	dimensionName := g.getDimensionNameForMainView(model, column)

	// Check if this is an ARRAY column
	isArrayColumn := false
	if column.DataType != nil {
		dataTypeUpper := strings.ToUpper(*column.DataType)
		isArrayColumn = strings.HasPrefix(dataTypeUpper, "ARRAY")
	}

	dimension := &models.LookMLDimension{
		Name:           dimensionName,
		Type:           g.getDimensionType(column),
		SQL:            g.getDimensionSQL(model, column),
		Description:    g.getDimensionDescription(column),
		GroupLabel:     g.GetDimensionGroupLabel(column),
		GroupItemLabel: g.getDimensionGroupItemLabel(column),
	}

	// Override hidden property for ARRAY columns in main view
	if isArrayColumn {
		hidden := true
		dimension.Hidden = &hidden
	}

	return dimension, nil
}

// GenerateDimensionGroup generates a LookML dimension group from a model column
func (g *DimensionGenerator) GenerateDimensionGroup(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimensionGroup, error) {
	if !g.shouldBeDimensionGroup(column) {
		return nil, nil
	}

	dimensionGroup := &models.LookMLDimensionGroup{
		Name:        g.getDimensionGroupName(column),
		Type:        g.getDimensionGroupType(column),
		SQL:         g.getDimensionSQL(model, column),
		Description: g.getDimensionDescription(column),
		GroupLabel:  g.GetDimensionGroupLabel(column),
		Timeframes:  g.getDimensionGroupTimeframes(column),
	}

	return dimensionGroup, nil
}

// GetDimensionName gets the dimension name from the column
func (g *DimensionGenerator) GetDimensionName(column *models.DbtModelColumn) string {
	// For hierarchical columns (containing dots), always generate the full LookML name
	// Don't use the existing LookMLName as it might just be the leaf part
	if strings.Contains(column.Name, ".") {
		// Use OriginalName if available (preserves PascalCase for proper conversion)
		nameToConvert := column.Name
		if column.OriginalName != nil && *column.OriginalName != "" {
			nameToConvert = *column.OriginalName
		}

		// Generate hierarchical name like "classification__itemsubgroup__code"
		lookmlName := utils.ToLookMLName(nameToConvert)
		return lookmlName
	}

	// For non-hierarchical columns, use existing LookMLName if available
	if column.LookMLName != nil {
		return *column.LookMLName
	}

	// Use the full column name (hierarchy path) for LookML dimension name
	// Use OriginalName if available (preserves PascalCase for proper conversion)
	nameToConvert := column.Name
	if column.OriginalName != nil && *column.OriginalName != "" {
		nameToConvert = *column.OriginalName
	}

	// This converts "SupplierInformation" -> "supplier_information"
	lookmlName := utils.ToLookMLName(nameToConvert)

	return lookmlName
}

// getDimensionGroupName gets the dimension group name from the column
func (g *DimensionGenerator) getDimensionGroupName(column *models.DbtModelColumn) string {
	name := g.GetDimensionName(column)

	// Remove common date suffixes for dimension groups
	// Note: _date_time (from PascalCase DateTime) should NOT be stripped
	// Only strip: _datetime (snake_case), _timestamp, _date
	suffixes := []string{"_datetime", "_timestamp", "_date"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
			break
		}
	}

	return name
}

// getDimensionNameForMainView gets the dimension name for main view columns, with special handling for ARRAY columns
func (g *DimensionGenerator) getDimensionNameForMainView(model *models.DbtModel, column *models.DbtModelColumn) string {
	// Check if this is an ARRAY column
	isArrayColumn := false
	if column.DataType != nil {
		dataTypeUpper := strings.ToUpper(*column.DataType)
		isArrayColumn = strings.HasPrefix(dataTypeUpper, "ARRAY")
	}

	// For ARRAY columns in main view, use the nested view naming pattern
	if isArrayColumn {
		// Get the view name (using the same logic as view generation)
		var viewName string
		if g.config.UseTableName {
			// Extract table name from RelationName
			parts := strings.Split(model.RelationName, ".")
			tableName := parts[len(parts)-1]
			tableName = strings.Trim(tableName, "`")
			viewName = strings.ToLower(tableName)
		} else {
			viewName = model.Name
		}

		// Generate dimension name: {view_name}__{array_name}
		arrayName := strings.ToLower(column.Name)
		return fmt.Sprintf("%s__%s", viewName, arrayName)
	}

	// For non-ARRAY columns, use the standard naming
	return g.GetDimensionName(column)
}

// getDimensionType gets the dimension type from the column data type
func (g *DimensionGenerator) getDimensionType(column *models.DbtModelColumn) string {
	if column.DataType == nil {
		return "string"
	}

	lookerType := enums.GetLookerType(*column.DataType)
	return string(lookerType)
}

// getDimensionGroupType gets the dimension group type based on the column data type
func (g *DimensionGenerator) getDimensionGroupType(column *models.DbtModelColumn) string {
	if column.DataType == nil {
		return dimGroupTypeTime
	}

	dataType := strings.ToUpper(*column.DataType)
	switch dataType {
	case dataTypeDate:
		return "date" // DATE fields use type: date
	case dataTypeDateTime, dataTypeTimestamp:
		return dimGroupTypeTime // DATETIME and TIMESTAMP use type: time
	default:
		return dimGroupTypeTime
	}
}

// getDimensionSQL gets the SQL expression for the dimension
func (g *DimensionGenerator) getDimensionSQL(model *models.DbtModel, column *models.DbtModelColumn) string {
	// Use OriginalName to preserve PascalCase for SQL references (matches fixture behavior)
	// This is critical for nested columns like Classification.ItemGroup.Code
	columnName := column.Name
	if column.OriginalName != nil && *column.OriginalName != "" {
		columnName = *column.OriginalName
	}

	// For ARRAY columns in main view, use the base column name (e.g., "sales" not "sales.field")
	if column.DataType != nil {
		dataTypeUpper := strings.ToUpper(*column.DataType)
		if strings.HasPrefix(dataTypeUpper, "ARRAY") {
			// Extract the base array name (before any dots)
			baseColumnName := strings.Split(columnName, ".")[0]
			return fmt.Sprintf("${TABLE}.%s", baseColumnName)
		}
	}

	// For nested columns with dots, use dot notation (no backticks needed for PascalCase)
	// For regular columns, use as-is
	// QuoteColumnNameIfNeeded will add backticks only if needed (spaces, special chars)
	return fmt.Sprintf("${TABLE}.%s", columnName)
}

// getDimensionDescription gets the dimension description from column metadata
func (g *DimensionGenerator) getDimensionDescription(column *models.DbtModelColumn) *string {
	// Return column description only if it exists and is not empty
	if column.Description != nil && *column.Description != "" {
		return column.Description
	}
	// Omit description when not present (matches fixture behavior)
	return nil
}

// GetDimensionGroupLabel gets the group label for the dimension
func (g *DimensionGenerator) GetDimensionGroupLabel(column *models.DbtModelColumn) *string {
	// For nested columns like "classification.assortment.code",
	// extract parent path as group label ("classification.assortment")
	if column.OriginalName != nil && strings.Contains(*column.OriginalName, ".") {
		parts := strings.Split(*column.OriginalName, ".")
		if len(parts) > 1 {
			// Join all parts except the last (the field name itself)
			parentPath := strings.Join(parts[:len(parts)-1], ".")
			// Convert to title case
			groupLabel := utils.ToTitleCase(parentPath)
			return &groupLabel
		}
	}
	return nil
}

// getDimensionGroupTimeframes gets the timeframes for a dimension group
func (g *DimensionGenerator) getDimensionGroupTimeframes(column *models.DbtModelColumn) []enums.LookerTimeFrame {
	// Use config timeframes if specified
	if len(g.config.Timeframes) > 0 {
		var timeframes []enums.LookerTimeFrame
		for _, tf := range g.config.Timeframes {
			timeframes = append(timeframes, enums.LookerTimeFrame(tf))
		}
		return timeframes
	}

	// Default timeframes based on data type
	if column.DataType != nil {
		dataType := strings.ToUpper(*column.DataType)
		switch dataType {
		case dataTypeDate:
			return []enums.LookerTimeFrame{
				enums.TimeFrameRaw,
				enums.TimeFrameDate,
				enums.TimeFrameWeek,
				enums.TimeFrameMonth,
				enums.TimeFrameQuarter,
				enums.TimeFrameYear,
			}
		case dataTypeDateTime, dataTypeTimestamp:
			return []enums.LookerTimeFrame{
				enums.TimeFrameRaw,
				enums.TimeFrameTime,
				enums.TimeFrameDate,
				enums.TimeFrameWeek,
				enums.TimeFrameMonth,
				enums.TimeFrameQuarter,
				enums.TimeFrameYear,
			}
		}
	}

	return nil
}

// getDimensionGroupItemLabel gets the group item label for nested columns
func (g *DimensionGenerator) getDimensionGroupItemLabel(column *models.DbtModelColumn) *string {
	// For nested columns like "classification.assortment.code", extract the last part
	if strings.Contains(column.Name, ".") {
		parts := strings.Split(column.Name, ".")
		lastPart := parts[len(parts)-1]
		// Convert to title case
		label := utils.ToTitleCase(lastPart)
		return &label
	}
	return nil
}

// shouldBeDimensionGroup determines if a column should be a dimension group
func (g *DimensionGenerator) shouldBeDimensionGroup(column *models.DbtModelColumn) bool {
	if column.DataType == nil {
		return false
	}

	dataType := strings.ToUpper(*column.DataType)
	return dataType == dataTypeDate || dataType == dataTypeDateTime || dataType == dataTypeTimestamp
}
