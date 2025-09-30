package parsers

import (
	"log"
	"strings"

	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
)

// CatalogParser handles parsing of DBT catalog data
type CatalogParser struct {
	catalog         *models.DbtCatalog
	rawCatalogData  map[string]interface{}
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
	
	// Debug: log processing for our test model
	if strings.Contains(model.Name, "dq_ICASOI_Current") {
		log.Printf("DEBUG: Processing catalog for model %s with %d manifest columns, %d catalog columns", model.Name, len(model.Columns), len(catalogNode.Columns))
		
		// Check if manifest has the ARRAY columns we found in catalog
		arrayColumnsInCatalog := []string{"item_information_claim_detail", "central_department", "ica_ethical_accreditation"}
		for _, arrayCol := range arrayColumnsInCatalog {
			if _, exists := model.Columns[arrayCol]; exists {
				log.Printf("DEBUG: Manifest HAS column %s", arrayCol)
			} else {
				log.Printf("DEBUG: Manifest MISSING column %s", arrayCol)
			}
		}
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
	log.Printf("DEBUG: Model %s - creating all columns from catalog (%d catalog columns)", model.Name, len(catalogNode.Columns))
	
	for catalogColumnName, catalogColumn := range catalogNode.Columns {
		// Create a new model column from catalog data
		dataTypeCopy := catalogColumn.DataType // Create a copy of the string
		newColumn := models.DbtModelColumn{
			Name:        catalogColumnName,  // Use normalized (lowercase) name for matching
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
			
			// Debug: log OriginalName setting for specific columns
			if strings.Contains(strings.ToLower(catalogColumnName), "buying") {
				log.Printf("DEBUG ORIGINAL: Column '%s' -> OriginalName '%s'", catalogColumnName, catalogColumn.OriginalName)
			}
		}
		
		// Debug: log what we're setting for ARRAY columns
		if catalogColumn.DataType != "" && strings.HasPrefix(strings.ToUpper(catalogColumn.DataType), "ARRAY") {
			log.Printf("DEBUG CREATE: Creating column %s with DataType: '%s'", catalogColumnName, catalogColumn.DataType)
		}
		
		newColumn.ProcessColumn()
		processedColumns[catalogColumnName] = newColumn
	}

	processedModel.Columns = processedColumns
	
	// Debug: check final processed model - specifically look for our expected columns
	if strings.Contains(model.Name, "dq_ICASOI_Current") {
		expectedArrayCols := []string{"format", "supplierinformation", "markings.marking"}
		arrayCount := 0
		
		log.Printf("DEBUG FINAL: Checking final model %s with %d total columns", model.Name, len(processedModel.Columns))
		
		for _, expectedCol := range expectedArrayCols {
			if col, exists := processedModel.Columns[expectedCol]; exists {
				if col.DataType != nil && strings.HasPrefix(strings.ToUpper(*col.DataType), "ARRAY") {
					log.Printf("DEBUG FINAL: ✅ Found expected ARRAY column %s with type %s", expectedCol, *col.DataType)
					arrayCount++
				} else {
					log.Printf("DEBUG FINAL: ❌ Found column %s but DataType is %v", expectedCol, col.DataType)
				}
			} else {
				log.Printf("DEBUG FINAL: ❌ Missing expected column %s", expectedCol)
			}
		}
		
		log.Printf("DEBUG FINAL: Final processed model %s has %d expected ARRAY columns", model.Name, arrayCount)
	}
	
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
