package generators

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	pluginMetrics "github.com/magnus-ffcg/go-dbt2lookml/pkg/plugins/metrics"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
)

// Compile-time check to ensure ViewGenerator implements ViewGeneratorInterface
var _ ViewGeneratorInterface = (*ViewGenerator)(nil)

// ViewGenerator handles generation of LookML views
type ViewGenerator struct {
	config                   *config.Config
	dimensionGenerator       *DimensionGenerator
	measureGenerator         *MeasureGenerator
	semanticMeasureGenerator *pluginMetrics.SemanticMeasureGenerator
}

// NewViewGenerator creates a new ViewGenerator instance
func NewViewGenerator(cfg *config.Config) *ViewGenerator {
	return &ViewGenerator{
		config:                   cfg,
		dimensionGenerator:       NewDimensionGenerator(cfg),
		measureGenerator:         NewMeasureGenerator(cfg),
		semanticMeasureGenerator: pluginMetrics.NewSemanticMeasureGenerator(cfg),
	}
}

// GenerateView generates a LookML view from a dbt model
func (g *ViewGenerator) GenerateView(model *models.DbtModel) (*models.LookMLView, error) {
	return g.GenerateViewWithSemanticMeasures(model, nil)
}

// GenerateViewWithSemanticMeasures generates a LookML view from a dbt model with optional semantic measures
func (g *ViewGenerator) GenerateViewWithSemanticMeasures(
	model *models.DbtModel,
	semanticMeasures []models.DbtSemanticMeasure,
) (*models.LookMLView, error) {
	// Create column collections once and reuse them
	columnCollections := models.NewColumnCollections(model, nil)

	view := &models.LookMLView{
		Name:         g.getViewName(model),
		SQLTableName: g.getSQLTableName(model),
		Label:        g.getViewLabel(model),
		Description:  g.getViewDescription(model),
		Hidden:       g.getViewHidden(model),
	}

	// Generate dimensions using the shared column collections
	dimensions, err := g.generateDimensionsWithCollections(model, columnCollections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dimensions: %w", err)
	}

	// Add nested view reference dimensions to main view
	nestedViewRefDimensions := g.generateNestedViewReferenceDimensions(model, columnCollections)
	dimensions = append(dimensions, nestedViewRefDimensions...)

	view.Dimensions = dimensions

	// Generate dimension groups (only for main view columns)
	dimensionGroups, err := g.generateDimensionGroups(model, columnCollections)
	if err != nil {
		return nil, fmt.Errorf("failed to generate dimension groups: %w", err)
	}
	view.DimensionGroups = dimensionGroups

	// Apply conflict resolution: rename dimensions that conflict with dimension groups
	if len(dimensions) > 0 && len(dimensionGroups) > 0 {
		dimensions = g.resolveConflicts(dimensions, dimensionGroups, model.Name)
		view.Dimensions = dimensions
	}

	// Generate measures with semantic model support
	measures, err := g.generateMeasuresWithSemanticModels(model, semanticMeasures)
	if err != nil {
		return nil, fmt.Errorf("failed to generate measures: %w", err)
	}
	view.Measures = measures

	return view, nil
}

// getViewName gets the view name from the model
func (g *ViewGenerator) getViewName(model *models.DbtModel) string {
	if g.config.UseTableName {
		// Extract just the table name from relation_name (remove project.dataset prefix and backticks)
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		return strings.ToLower(tableName)
	}
	return model.Name
}

// getSQLTableName gets the SQL table name for the view
func (g *ViewGenerator) getSQLTableName(model *models.DbtModel) string {
	if g.config.UseTableName {
		// When using table names, use the RelationName but remove individual backticks and add single backticks around the whole name
		relationName := model.RelationName
		// Remove individual backticks: `project`.`dataset`.`table` -> project.dataset.table
		relationName = strings.ReplaceAll(relationName, "`", "")
		// Add single backticks around the whole name: project.dataset.table -> `project.dataset.table`
		return fmt.Sprintf("`%s`", relationName)
	}

	// For model names, construct the schema.tableName format
	schema := model.Schema
	if g.config.RemoveSchemaString != "" {
		schema = strings.ReplaceAll(schema, g.config.RemoveSchemaString, "")
	}

	return utils.QuoteColumnNameIfNeeded(fmt.Sprintf("%s.%s", schema, model.Name))
}

// getViewLabel gets the view label from model metadata or generates one
func (g *ViewGenerator) getViewLabel(model *models.DbtModel) *string {
	// Only return label if explicitly defined in metadata (matches fixture behavior)
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		model.Meta.Looker.View != nil &&
		model.Meta.Looker.View.Label != nil {
		return model.Meta.Looker.View.Label
	}

	// Return nil to omit label when not explicitly defined
	return nil
}

// getViewDescription gets the view description from model
func (g *ViewGenerator) getViewDescription(model *models.DbtModel) *string {
	if model.Description != "" {
		return &model.Description
	}
	return nil
}

// getViewHidden gets the view hidden setting from model metadata
func (g *ViewGenerator) getViewHidden(model *models.DbtModel) *bool {
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		model.Meta.Looker.View != nil &&
		model.Meta.Looker.View.Hidden != nil {
		return model.Meta.Looker.View.Hidden
	}
	return nil
}

// generateDimensions generates dimensions for the view (legacy method)
func (g *ViewGenerator) generateDimensions(model *models.DbtModel) ([]models.LookMLDimension, error) {
	// Create column collections and delegate to the new method
	columnCollections := models.NewColumnCollections(model, nil)
	return g.generateDimensionsWithCollections(model, columnCollections)
}

// generateDimensionsWithCollections generates dimensions for the view using provided column collections
func (g *ViewGenerator) generateDimensionsWithCollections(model *models.DbtModel, columnCollections *models.ColumnCollections) ([]models.LookMLDimension, error) {
	var dimensions []models.LookMLDimension

	// Generate dimensions for ALL main view columns (including those that will become dimension groups)
	// This is needed to generate conflict dimensions for date/time fields before classification

	for colName, column := range columnCollections.MainViewColumns {
		// Create a proper deep copy of the column to avoid shared pointer issues
		columnCopy := models.DbtModelColumn{
			Name:         colName, // Use the full path from the map key
			Nested:       column.Nested,
			IsPrimaryKey: column.IsPrimaryKey,
			InnerTypes:   column.InnerTypes, // Slice is copied by value
			Meta:         column.Meta,       // Pointer to metadata (shared is OK)
		}

		// Deep copy all pointer fields to avoid shared references
		if column.OriginalName != nil {
			originalNameCopy := *column.OriginalName
			columnCopy.OriginalName = &originalNameCopy
		}
		if column.DataType != nil {
			dataTypeCopy := *column.DataType
			columnCopy.DataType = &dataTypeCopy
		}
		if column.Description != nil {
			descriptionCopy := *column.Description
			columnCopy.Description = &descriptionCopy
		}
		if column.LookMLName != nil {
			lookmlNameCopy := *column.LookMLName
			columnCopy.LookMLName = &lookmlNameCopy
		}
		if column.LookMLLongName != nil {
			lookmlLongNameCopy := *column.LookMLLongName
			columnCopy.LookMLLongName = &lookmlLongNameCopy
		}

		dimension, err := g.dimensionGenerator.GenerateDimension(model, &columnCopy)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dimension for column %s: %w", column.Name, err)
		}

		if dimension != nil {
			dimensions = append(dimensions, *dimension)
		}
	}

	return dimensions, nil
}

// resolveConflicts uses the domain service to resolve dimension/dimension-group naming conflicts
func (g *ViewGenerator) resolveConflicts(dimensions []models.LookMLDimension, dimensionGroups []models.LookMLDimensionGroup, modelName string) []models.LookMLDimension {
	resolver := models.NewDimensionConflictResolver(g.config)
	return resolver.Resolve(dimensions, dimensionGroups, modelName)
}

// generateNestedViewReferenceDimensions generates hidden dimensions in main view that reference nested views
func (g *ViewGenerator) generateNestedViewReferenceDimensions(model *models.DbtModel, columnCollections *models.ColumnCollections) []models.LookMLDimension {
	var dimensions []models.LookMLDimension

	// Apply business rules for nested array processing
	arrayRules := models.NewNestedArrayRules()

	// For each nested view, create a corresponding reference dimension in main view
	for arrayName, nestedCols := range columnCollections.NestedViewColumns {
		// Skip arrays that exceed the maximum nesting depth
		if !arrayRules.ShouldProcessArray(arrayName) {
			continue
		}

		// Skip arrays that are nested inside other arrays
		// e.g., "sales.f_sale_receipt_pseudo_keys" where "sales" is also an array
		// Only include arrays that are direct children of the table or children of STRUCTs
		if strings.Contains(arrayName, ".") {
			// Check if the parent is also an array
			parentName := arrayName[:strings.LastIndex(arrayName, ".")]
			if _, isParentArray := columnCollections.NestedViewColumns[parentName]; isParentArray {
				// Parent is an array, so skip this nested array
				continue
			}
		}

		// Find the array column to get its OriginalName
		var arrayOriginalName string
		for _, col := range nestedCols {
			if col.Name == arrayName {
				if col.OriginalName != nil && *col.OriginalName != "" {
					arrayOriginalName = *col.OriginalName
				}
				break
			}
		}

		// Generate the nested view name (same logic as nested view generation)
		var baseName string
		if g.config.UseTableName {
			// Extract table name from RelationName
			parts := strings.Split(model.RelationName, ".")
			tableName := parts[len(parts)-1]
			tableName = strings.Trim(tableName, "`")
			baseName = strings.ToLower(tableName)
		} else {
			baseName = model.Name
		}

		// Use OriginalName if available for proper PascalCase conversion
		var nestedSuffix string
		if arrayOriginalName != "" {
			nestedSuffix = utils.ToLookMLName(arrayOriginalName)
		} else {
			nestedSuffix = utils.ToLookMLName(arrayName)
		}

		// The dimension name in the main view should be just the array name (short form)
		// e.g., "sales", "supplier_information"
		dimensionName := nestedSuffix

		// SQL reference should be the full nested view name
		sqlRef := fmt.Sprintf("%s__%s", baseName, nestedSuffix)

		// Create hidden dimension that references the nested view
		hidden := true
		dimension := models.LookMLDimension{
			Name:   dimensionName,
			Type:   "string",
			SQL:    sqlRef, // Reference nested view with full name
			Hidden: &hidden,
		}

		dimensions = append(dimensions, dimension)
	}

	return dimensions
}

// generateDimensionGroups generates dimension groups for the view
func (g *ViewGenerator) generateDimensionGroups(model *models.DbtModel, columnCollections *models.ColumnCollections) ([]models.LookMLDimensionGroup, error) {
	var dimensionGroups []models.LookMLDimensionGroup

	// Only process main view columns, not nested columns
	for _, column := range columnCollections.MainViewColumns {
		// Only process columns that should be dimension groups
		if !g.shouldBeDimensionGroup(column) {
			continue
		}

		dimensionGroup, err := g.dimensionGenerator.GenerateDimensionGroup(model, &column)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dimension group for column %s: %w", column.Name, err)
		}

		if dimensionGroup != nil {
			dimensionGroups = append(dimensionGroups, *dimensionGroup)
		}
	}

	return dimensionGroups, nil
}

// generateMeasures generates measures for the view
func (g *ViewGenerator) generateMeasures(model *models.DbtModel) ([]models.LookMLMeasure, error) {
	return g.generateMeasuresWithSemanticModels(model, nil)
}

// generateMeasuresWithSemanticModels generates measures, optionally including semantic model measures
func (g *ViewGenerator) generateMeasuresWithSemanticModels(
	model *models.DbtModel,
	semanticMeasures []models.DbtSemanticMeasure,
) ([]models.LookMLMeasure, error) {
	var metaMeasures []*models.LookMLMeasure
	var semanticLookMLMeasures []*models.LookMLMeasure

	// Generate measures from semantic models if enabled and provided
	if g.config.UseSemanticModels && len(semanticMeasures) > 0 {
		generated, err := g.semanticMeasureGenerator.GenerateMeasuresFromSemantic(semanticMeasures, model)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("model", model.Name).
				Msg("Failed to generate measures from semantic models")
		} else {
			semanticLookMLMeasures = generated
		}
	}

	// Generate measures from model meta
	if model.Meta != nil && model.Meta.Looker != nil {
		for _, measureMeta := range model.Meta.Looker.Measures {
			measure, err := g.measureGenerator.GenerateMeasure(model, &measureMeta)
			if err != nil {
				return nil, fmt.Errorf("failed to generate measure: %w", err)
			}

			if measure != nil {
				metaMeasures = append(metaMeasures, measure)
			}
		}
	}

	// Merge semantic and meta measures (semantic measures take precedence)
	var mergedMeasures []*models.LookMLMeasure
	if len(semanticLookMLMeasures) > 0 && len(metaMeasures) > 0 {
		mergedMeasures = g.semanticMeasureGenerator.MergeWithMetaMeasures(
			semanticLookMLMeasures,
			metaMeasures,
		)
	} else if len(semanticLookMLMeasures) > 0 {
		mergedMeasures = semanticLookMLMeasures
	} else {
		mergedMeasures = metaMeasures
	}

	// Convert pointer slice to value slice
	measures := make([]models.LookMLMeasure, 0, len(mergedMeasures))
	for _, m := range mergedMeasures {
		measures = append(measures, *m)
	}

	// Generate default count measure only if no count measure exists
	hasCountMeasure := false
	for _, m := range measures {
		if m.Name == DefaultCountMeasureName {
			hasCountMeasure = true
			break
		}
	}

	if !hasCountMeasure {
		countMeasure := g.measureGenerator.GenerateDefaultCountMeasure(model)
		if countMeasure != nil {
			measures = append(measures, *countMeasure)
		}
	}

	return measures, nil
}

// shouldBeDimensionGroup determines if a column should be a dimension group
func (g *ViewGenerator) shouldBeDimensionGroup(column models.DbtModelColumn) bool {
	if column.DataType == nil {
		return false
	}

	dataType := strings.ToUpper(*column.DataType)
	return dataType == "DATE" || dataType == "DATETIME" || dataType == "TIMESTAMP"
}

// GenerateNestedView generates a nested view for array/struct columns
func (g *ViewGenerator) GenerateNestedView(model *models.DbtModel, arrayColumn *models.DbtModelColumn) (*models.LookMLView, error) {
	// This would implement nested view generation for ARRAY<STRUCT> columns
	// Simplified implementation for now

	nestedViewName := fmt.Sprintf("%s__%s", model.Name, *arrayColumn.LookMLName)

	view := &models.LookMLView{
		Name:         nestedViewName,
		SQLTableName: fmt.Sprintf("${%s.SQL_TABLE_NAME}", model.Name),
	}

	// Generate dimensions for nested fields
	// This would need to parse the nested structure from catalog data

	return view, nil
}
