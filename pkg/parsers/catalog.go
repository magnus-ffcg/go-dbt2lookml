package parsers

import (
	"log"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// CatalogParser handles parsing of DBT catalog data
type CatalogParser struct {
	catalog        *models.DbtCatalog
	rawCatalogData map[string]interface{}
}

// NewCatalogParser creates a new CatalogParser instance
func NewCatalogParser(catalog *models.DbtCatalog, rawCatalog map[string]interface{}) *CatalogParser {
	return &CatalogParser{
		catalog:        catalog,
		rawCatalogData: rawCatalog,
	}
}

// ProcessModelColumns processes model columns by merging with catalog information
func (p *CatalogParser) ProcessModelColumns(model *models.DbtModel) (*models.DbtModel, error) {
	// Find corresponding catalog node
	catalogNode, exists := p.catalog.Nodes[model.UniqueID]
	if !exists {
		log.Printf("No catalog entry found for model: %s", model.Name)
		return model, nil
	}

	// Normalize catalog column names and process column types
	catalogNode.NormalizeColumnNames()

	// Process each catalog column to extract DataType from Type
	for columnName, catalogColumn := range catalogNode.Columns {
		catalogColumn.ProcessColumnType()
		// Update the column back in the map since Go passes by value
		catalogNode.Columns[columnName] = catalogColumn
	}

	// Create a copy of the model to avoid modifying the original
	processedModel := *model
	processedColumns := make(map[string]models.DbtModelColumn)

	// Always create columns from catalog (manifest columns are typically empty)
	for catalogColumnName, catalogColumn := range catalogNode.Columns {
		// Create a new model column from catalog data
		dataTypeCopy := catalogColumn.DataType // Create a copy of the string
		newColumn := models.DbtModelColumn{
			Name:        catalogColumnName, // Use normalized (lowercase) name for matching
			DataType:    &dataTypeCopy,
			InnerTypes:  catalogColumn.InnerTypes,
			Description: catalogColumn.Comment,
		}

		// Set OriginalName for proper LookML naming (preserves PascalCase)
		// CRITICAL: Must create a new string copy to avoid pointer sharing!
		// Use catalogColumn.OriginalName (set by NormalizeColumnNames) which preserves PascalCase
		if catalogColumn.OriginalName != "" {
			originalNameCopy := catalogColumn.OriginalName
			newColumn.OriginalName = &originalNameCopy
		}

		newColumn.ProcessColumn()
		processedColumns[catalogColumnName] = newColumn
	}

	processedModel.Columns = processedColumns

	return &processedModel, nil
}

// GetCatalogColumn gets a specific column from the catalog
func (p *CatalogParser) GetCatalogColumn(modelUniqueID, columnName string) (*models.DbtCatalogNodeColumn, bool) {
	catalogNode, exists := p.catalog.Nodes[modelUniqueID]
	if !exists {
		return nil, false
	}

	column, found := catalogNode.Columns[strings.ToLower(columnName)]
	if !found {
		return nil, false
	}

	return &column, true
}

// GetModelCatalogData gets the raw catalog data for a specific model
func (p *CatalogParser) GetModelCatalogData(modelUniqueID string) (map[string]interface{}, bool) {
	if nodes, ok := p.rawCatalogData["nodes"].(map[string]interface{}); ok {
		if modelData, exists := nodes[modelUniqueID]; exists {
			if modelMap, ok := modelData.(map[string]interface{}); ok {
				return modelMap, true
			}
		}
	}
	return nil, false
}

// GetColumnType gets the BigQuery type for a specific column
func (p *CatalogParser) GetColumnType(modelUniqueID, columnName string) (string, bool) {
	if column, found := p.GetCatalogColumn(modelUniqueID, columnName); found {
		return column.Type, true
	}
	return "", false
}

// GetColumnDataType gets the simplified data type for a specific column
func (p *CatalogParser) GetColumnDataType(modelUniqueID, columnName string) (string, bool) {
	if column, found := p.GetCatalogColumn(modelUniqueID, columnName); found {
		return column.DataType, true
	}
	return "", false
}

// GetColumnInnerTypes gets the inner types for a specific column (for ARRAY/STRUCT types)
func (p *CatalogParser) GetColumnInnerTypes(modelUniqueID, columnName string) ([]string, bool) {
	if column, found := p.GetCatalogColumn(modelUniqueID, columnName); found {
		return column.InnerTypes, true
	}
	return nil, false
}

// IsArrayType checks if a column is an ARRAY type
func (p *CatalogParser) IsArrayType(modelUniqueID, columnName string) bool {
	if columnType, found := p.GetColumnType(modelUniqueID, columnName); found {
		return strings.HasPrefix(strings.ToUpper(columnType), "ARRAY")
	}
	return false
}

// IsStructType checks if a column is a STRUCT type
func (p *CatalogParser) IsStructType(modelUniqueID, columnName string) bool {
	if columnType, found := p.GetColumnType(modelUniqueID, columnName); found {
		return strings.HasPrefix(strings.ToUpper(columnType), "STRUCT")
	}
	return false
}

// GetNestedColumns gets all nested columns for a given parent column
func (p *CatalogParser) GetNestedColumns(modelUniqueID, parentColumnName string) []models.DbtCatalogNodeColumn {
	catalogNode, exists := p.catalog.Nodes[modelUniqueID]
	if !exists {
		return nil
	}

	var nestedColumns []models.DbtCatalogNodeColumn
	parentPrefix := strings.ToLower(parentColumnName) + "."

	for columnName, column := range catalogNode.Columns {
		if strings.HasPrefix(strings.ToLower(columnName), parentPrefix) {
			nestedColumns = append(nestedColumns, column)
		}
	}

	return nestedColumns
}
