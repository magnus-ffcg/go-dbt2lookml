package models

import "strings"

// ColumnCategory represents where a column should be placed.
type ColumnCategory int

const (
	// CategoryExcluded means the column should be excluded from all views
	CategoryExcluded ColumnCategory = iota
	// CategoryMainView means the column belongs in the main view
	CategoryMainView
	// CategoryNestedView means the column belongs in a nested view
	CategoryNestedView
)

// ColumnClassifier determines where each column should be placed based on
// business rules for BigQuery nested structures (ARRAY, STRUCT).
type ColumnClassifier struct {
	hierarchy    *ColumnHierarchy
	arrayColumns map[string]bool
}

// NewColumnClassifier creates a classifier with the given hierarchy and array columns.
func NewColumnClassifier(hierarchy *ColumnHierarchy, arrayColumns map[string]bool) *ColumnClassifier {
	return &ColumnClassifier{
		hierarchy:    hierarchy,
		arrayColumns: arrayColumns,
	}
}

// Classify determines the category for a given column.
func (c *ColumnClassifier) Classify(columnName string, column DbtModelColumn) ColumnCategory {
	// Rule 1: Check if column should be excluded from all views
	if c.shouldExclude(column) {
		return CategoryExcluded
	}

	// Rule 2: Check if this is an array column
	if c.arrayColumns[columnName] {
		return CategoryNestedView
	}

	// Rule 3: Check if column belongs to an array (has array parent)
	arrayParent := c.findArrayParent(columnName)
	if arrayParent != "" {
		return CategoryNestedView
	}

	// Rule 4: Default - belongs to main view
	return CategoryMainView
}

// GetArrayParent returns the array parent for a column, if any.
func (c *ColumnClassifier) GetArrayParent(columnName string) string {
	return c.findArrayParent(columnName)
}

// shouldExclude checks if a column should be excluded from all views.
// Excludes STRUCT parents that have children (but not ARRAY<STRUCT>).
func (c *ColumnClassifier) shouldExclude(column DbtModelColumn) bool {
	if column.DataType == nil {
		return false
	}

	dataTypeUpper := strings.ToUpper(*column.DataType)

	// Only exclude STRUCTs that are not part of an ARRAY
	if !strings.Contains(dataTypeUpper, "STRUCT") || strings.HasPrefix(dataTypeUpper, "ARRAY") {
		return false
	}

	// Check if this STRUCT has nested children
	for otherPath := range c.hierarchy.All() {
		if strings.HasPrefix(otherPath, column.Name+".") {
			return true
		}
	}

	return false
}

// findArrayParent finds which array this column belongs to, if any.
// Returns the most specific (longest) array parent.
func (c *ColumnClassifier) findArrayParent(colName string) string {
	var matchingArrays []string

	for arrayName := range c.arrayColumns {
		if strings.HasPrefix(colName, arrayName+".") {
			matchingArrays = append(matchingArrays, arrayName)
		}
	}

	// Return the longest match (most specific)
	if len(matchingArrays) > 0 {
		longest := matchingArrays[0]
		for _, match := range matchingArrays[1:] {
			if len(match) > len(longest) {
				longest = match
			}
		}
		return longest
	}

	return ""
}

// HasChildren checks if a column has child columns.
func (c *ColumnClassifier) HasChildren(columnName string) bool {
	return c.hierarchy.HasChildren(columnName)
}

// IsNestedArray checks if an array column is nested under another array.
func (c *ColumnClassifier) IsNestedArray(columnName string) bool {
	arrayParent := c.findArrayParent(columnName)
	return arrayParent != "" && c.arrayColumns[arrayParent]
}
