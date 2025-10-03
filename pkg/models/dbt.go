// Package models defines data structures for dbt and LookML representations.
//
// This package contains the domain models used throughout the application,
// including dbt artifacts (models, columns, metadata) and LookML structures
// (views, dimensions, measures, explores).
//
// dbt Models:
//   - DbtModel: Represents a dbt model with columns and metadata
//   - DbtModelColumn: Column definition with data type and metadata
//   - DbtCatalog: Parsed catalog.json structure
//   - DbtManifest: Parsed manifest.json structure
//   - DbtExposure: dbt exposure definitions
//
// LookML Models:
//   - LookMLView: Complete LookML view structure
//   - LookMLDimension: Dimension field definition
//   - LookMLDimensionGroup: Dimension group for dates/timestamps
//   - LookMLMeasure: Measure field definition
//   - LookMLExplore: Explore with joins
//
// Metadata:
//   - DbtMetaLooker: Custom LookML metadata from dbt meta tags
//   - DbtMetaLookerDimension: Dimension-specific metadata
//   - DbtMetaLookerMeasure: Measure-specific metadata
//
// All LookML models include Validate() methods to ensure data integrity
// before generation.
package models

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
)

// DbtBaseModel represents the base model for dbt objects
type DbtBaseModel struct{}

// DbtNode represents a dbt node, extensible to models, seeds, etc.
type DbtNode struct {
	Name         string                `json:"name" yaml:"name"`
	UniqueID     string                `json:"unique_id" yaml:"unique_id"`
	ResourceType enums.DbtResourceType `json:"resource_type" yaml:"resource_type"`
}

// DbtExposureRef represents a reference in a dbt exposure
type DbtExposureRef struct {
	Name    string      `json:"name" yaml:"name"`
	Package *string     `json:"package,omitempty" yaml:"package,omitempty"`
	Version interface{} `json:"version,omitempty" yaml:"version,omitempty"` // Can be string or int
}

// DbtDependsOn represents dependencies between dbt objects
type DbtDependsOn struct {
	Macros []string `json:"macros" yaml:"macros"`
	Nodes  []string `json:"nodes" yaml:"nodes"`
}

// DbtExposure represents a dbt exposure
type DbtExposure struct {
	DbtNode
	Description *string          `json:"description,omitempty" yaml:"description,omitempty"`
	URL         *string          `json:"url,omitempty" yaml:"url,omitempty"`
	Refs        []DbtExposureRef `json:"refs" yaml:"refs"`
	Tags        []string         `json:"tags" yaml:"tags"`
	DependsOn   DbtDependsOn     `json:"depends_on" yaml:"depends_on"`
}

// DbtCatalogNodeMetadata represents metadata about a dbt catalog node
type DbtCatalogNodeMetadata struct {
	Type    string  `json:"type" yaml:"type"`
	Schema  string  `json:"schema" yaml:"schema"`
	Name    string  `json:"name" yaml:"name"`
	Comment *string `json:"comment,omitempty" yaml:"comment,omitempty"`
	Owner   *string `json:"owner,omitempty" yaml:"owner,omitempty"`
}

// DbtCatalogNodeColumn represents a column in a dbt catalog node
type DbtCatalogNodeColumn struct {
	Type         string                `json:"type" yaml:"type"`
	DataType     string                `json:"data_type" yaml:"data_type"`
	InnerTypes   []string              `json:"inner_types" yaml:"inner_types"`
	Comment      *string               `json:"comment,omitempty" yaml:"comment,omitempty"`
	Index        int                   `json:"index" yaml:"index"`
	Name         string                `json:"name" yaml:"name"`
	OriginalName string                `json:"original_name" yaml:"original_name"`
	Parent       *DbtCatalogNodeColumn `json:"parent,omitempty" yaml:"parent,omitempty"`
}

// ProcessColumnType processes the column type and extracts data type and inner types
func (c *DbtCatalogNodeColumn) ProcessColumnType() {
	// Extract data type (everything before '<' or '(') - like Python does
	dataType := c.Type
	if idx := strings.Index(dataType, "<"); idx != -1 {
		dataType = dataType[:idx]
	}
	if idx := strings.Index(dataType, "("); idx != -1 {
		dataType = dataType[:idx]
	}
	c.DataType = dataType

	// Parse inner types using schema parser (simplified version)
	// This would need to be implemented based on the Python schema parser
	c.InnerTypes = parseInnerTypes(c.Type)
}

// parseInnerTypes is a simplified version of the schema parser
func parseInnerTypes(columnType string) []string {
	// This is a simplified implementation
	// The full implementation would need to parse complex BigQuery types
	var innerTypes []string

	// Basic ARRAY<TYPE> parsing
	if strings.HasPrefix(columnType, "ARRAY<") && strings.HasSuffix(columnType, ">") {
		inner := columnType[6 : len(columnType)-1]
		innerTypes = append(innerTypes, inner)
	}

	return innerTypes
}

// DbtCatalogNode represents a dbt catalog node
type DbtCatalogNode struct {
	Metadata DbtCatalogNodeMetadata          `json:"metadata" yaml:"metadata"`
	Columns  map[string]DbtCatalogNodeColumn `json:"columns" yaml:"columns"`
}

// NormalizeColumnNames converts all column names to lowercase for case-insensitive matching
// but preserves the original name for LookML generation
func (n *DbtCatalogNode) NormalizeColumnNames() {
	normalizedColumns := make(map[string]DbtCatalogNodeColumn, len(n.Columns))
	for name, column := range n.Columns {
		lowerName := strings.ToLower(name)
		// Preserve the original name for proper LookML naming
		column.OriginalName = name
		column.Name = lowerName

		normalizedColumns[lowerName] = column
	}
	n.Columns = normalizedColumns
}

// DbtCatalog represents a dbt catalog
type DbtCatalog struct {
	Nodes map[string]DbtCatalogNode `json:"nodes" yaml:"nodes"`
}

// DbtModelColumnMeta represents metadata about a column in a dbt model
type DbtModelColumnMeta struct {
	Looker *DbtMetaLooker `json:"looker,omitempty" yaml:"looker,omitempty"`
}

// DbtModelColumn represents a column in a dbt model
type DbtModelColumn struct {
	Name           string              `json:"name" yaml:"name"`
	Description    *string             `json:"description,omitempty" yaml:"description,omitempty"`
	LookMLLongName *string             `json:"lookml_long_name,omitempty" yaml:"lookml_long_name,omitempty"`
	LookMLName     *string             `json:"lookml_name,omitempty" yaml:"lookml_name,omitempty"`
	OriginalName   *string             `json:"original_name,omitempty" yaml:"original_name,omitempty"`
	DataType       *string             `json:"data_type,omitempty" yaml:"data_type,omitempty"`
	InnerTypes     []string            `json:"inner_types" yaml:"inner_types"`
	Meta           *DbtModelColumnMeta `json:"meta,omitempty" yaml:"meta,omitempty"`
	Nested         bool                `json:"nested" yaml:"nested"`
	IsPrimaryKey   bool                `json:"is_primary_key" yaml:"is_primary_key"`
}

// ProcessColumn processes the column and sets derived fields
func (c *DbtModelColumn) ProcessColumn() {
	// Set original name if not already set
	if c.OriginalName == nil {
		originalName := c.Name
		c.OriginalName = &originalName
	}

	// Check if nested (contains dot)
	if strings.Contains(c.Name, ".") {
		c.Nested = true
	}

	// Convert to lowercase for processing
	c.Name = strings.ToLower(c.Name)

	// Generate LookML names
	c.generateLookMLNames()

	// Do NOT set a default description - let it be nil if not provided
	// This matches fixture behavior where descriptions are omitted when not present
}

// generateLookMLNames generates LookML long name and name from the original name
func (c *DbtModelColumn) generateLookMLNames() {
	originalName := c.Name
	if c.OriginalName != nil {
		originalName = *c.OriginalName
	}

	// Split by dots and convert each part
	parts := strings.Split(originalName, ".")
	var snakeParts []string

	for _, part := range parts {
		if isLowerCaseWithoutUnderscore(part) {
			snakeParts = append(snakeParts, part)
		} else {
			snakeParts = append(snakeParts, utils.CamelToSnake(part))
		}
	}

	// LookML long name uses double underscores
	longName := strings.Join(snakeParts, "__")
	c.LookMLLongName = &longName

	// LookML name is just the last part
	if len(snakeParts) > 0 {
		name := snakeParts[len(snakeParts)-1]
		c.LookMLName = &name
	}
}

// isLowerCaseWithoutUnderscore checks if a string is pure lowercase without underscores
func isLowerCaseWithoutUnderscore(s string) bool {
	return strings.ToLower(s) == s && !strings.Contains(s, "_")
}

// DbtModelMeta represents metadata about a dbt model
type DbtModelMeta struct {
	Looker *DbtMetaLooker `json:"looker,omitempty" yaml:"looker,omitempty"`
}

// DbtModel represents a dbt model
type DbtModel struct {
	DbtNode
	ResourceType string                    `json:"resource_type" yaml:"resource_type"`
	RelationName string                    `json:"relation_name" yaml:"relation_name"`
	Schema       string                    `json:"schema" yaml:"schema"`
	Description  string                    `json:"description" yaml:"description"`
	Columns      map[string]DbtModelColumn `json:"columns" yaml:"columns"`
	Tags         []string                  `json:"tags" yaml:"tags"`
	Meta         *DbtModelMeta             `json:"meta,omitempty" yaml:"meta,omitempty"`
	Path         string                    `json:"path" yaml:"path"`
}

// NormalizeColumnNames converts all column names to lowercase for case-insensitive matching
func (m *DbtModel) NormalizeColumnNames() {
	normalizedColumns := make(map[string]DbtModelColumn, len(m.Columns))
	for name, column := range m.Columns {
		lowerName := strings.ToLower(name)
		column.Name = lowerName
		normalizedColumns[lowerName] = column
	}
	m.Columns = normalizedColumns
}

// DbtManifestMetadata represents metadata about a dbt manifest
type DbtManifestMetadata struct {
	AdapterType string `json:"adapter_type" yaml:"adapter_type"`
}

// ValidateAdapter validates that the adapter type is supported
func (m *DbtManifestMetadata) ValidateAdapter() error {
	supportedAdapters := []string{string(enums.BigQuery)}
	for _, adapter := range supportedAdapters {
		if m.AdapterType == adapter {
			return nil
		}
	}
	return fmt.Errorf("adapter type %s is not supported. Supported adapters are: %v",
		m.AdapterType, supportedAdapters)
}

// DbtManifest represents a dbt manifest
type DbtManifest struct {
	Nodes          map[string]interface{}      `json:"nodes" yaml:"nodes"` // Can be DbtModel or DbtNode
	Metadata       DbtManifestMetadata         `json:"metadata" yaml:"metadata"`
	Exposures      map[string]DbtExposure      `json:"exposures" yaml:"exposures"`
	SemanticModels map[string]DbtSemanticModel `json:"semantic_models,omitempty" yaml:"semantic_models,omitempty"`
}
