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
		Label:          g.getDimensionLabel(column),
		Description:    g.getDimensionDescription(column),
		Hidden:         g.getDimensionHidden(column),
		GroupLabel:     g.GetDimensionGroupLabel(column),
		GroupItemLabel: g.getDimensionGroupItemLabel(column),
	}

	// Override hidden property for ARRAY columns in main view
	if isArrayColumn {
		hidden := true
		dimension.Hidden = &hidden
	}

	// Set additional properties based on type
	switch dimension.Type {
	case "yesno":
		// Boolean dimensions don't need additional properties
	case "number":
		// Number dimensions might have value format
		dimension.ValueFormatName = g.getDimensionValueFormat(column)
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
		Label:       g.getDimensionLabel(column),
		Description: g.getDimensionDescription(column),
		Hidden:      g.getDimensionHidden(column),
		GroupLabel:  g.GetDimensionGroupLabel(column),
		Timeframes:  g.getDimensionGroupTimeframes(column),
		ConvertTZ:   g.getDimensionGroupConvertTZ(column),
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
		return dimGroupTypeTime
	case dataTypeDateTime:
		return dimGroupTypeTime
	case dataTypeTimestamp:
		return dimGroupTypeTime
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

// getDimensionLabel gets the dimension label
func (g *DimensionGenerator) getDimensionLabel(column *models.DbtModelColumn) *string {
	// Only return label if explicitly defined in metadata
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.Label != nil {
		return column.Meta.Looker.Dimension.Label
	}

	// Return nil to omit label when not explicitly defined (matches fixture behavior)
	return nil
}

// getDimensionDescription gets the dimension description
func (g *DimensionGenerator) getDimensionDescription(column *models.DbtModelColumn) *string {
	// Check meta looker dimension description first
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.Description != nil {
		return column.Meta.Looker.Dimension.Description
	}

	// Return column description only if it exists and is not empty
	if column.Description != nil && *column.Description != "" {
		return column.Description
	}

	// Return nil to omit description when not present (matches fixture behavior)
	return nil
}

// getDimensionHidden gets the dimension hidden setting
func (g *DimensionGenerator) getDimensionHidden(column *models.DbtModelColumn) *bool {
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.Hidden != nil {
		return column.Meta.Looker.Dimension.Hidden
	}
	return nil
}

// GetDimensionGroupLabel gets the group label for the dimension
func (g *DimensionGenerator) GetDimensionGroupLabel(column *models.DbtModelColumn) *string {
	// Check metadata first
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.GroupLabel != nil {
		return column.Meta.Looker.Dimension.GroupLabel
	}

	// For nested columns like "classification.assortment.code",
	// create group label from parent path: "Classification Assortment"
	if strings.Contains(column.Name, ".") {
		parts := strings.Split(column.Name, ".")
		if len(parts) > 1 {
			// Take all parts except the last one
			parentParts := parts[:len(parts)-1]
			// Convert each part to title case and join with space
			var titleParts []string
			for _, part := range parentParts {
				titleParts = append(titleParts, utils.ToTitleCase(part))
			}
			label := strings.Join(titleParts, " ")
			return &label
		}
	}

	return nil
}

// getDimensionValueFormat gets the value format for number dimensions
func (g *DimensionGenerator) getDimensionValueFormat(column *models.DbtModelColumn) *enums.LookerValueFormatName {
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.ValueFormatName != nil {
		return column.Meta.Looker.Dimension.ValueFormatName
	}
	return nil
}

// getDimensionGroupTimeframes gets the timeframes for dimension groups
func (g *DimensionGenerator) getDimensionGroupTimeframes(column *models.DbtModelColumn) []enums.LookerTimeFrame {
	// Check meta looker dimension timeframes
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		len(column.Meta.Looker.Dimension.Timeframes) > 0 {
		return column.Meta.Looker.Dimension.Timeframes
	}

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

// getDimensionGroupConvertTZ gets the convert_tz setting for dimension groups
func (g *DimensionGenerator) getDimensionGroupConvertTZ(column *models.DbtModelColumn) *bool {
	if column.Meta != nil &&
		column.Meta.Looker != nil &&
		column.Meta.Looker.Dimension != nil &&
		column.Meta.Looker.Dimension.ConvertTZ != nil {
		return column.Meta.Looker.Dimension.ConvertTZ
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
