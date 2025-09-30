package models

import (
	"fmt"
	"strings"
)

// ColumnCollections organizes model columns by their intended use
type ColumnCollections struct {
	MainViewColumns   map[string]DbtModelColumn            // Columns for the main view
	NestedViewColumns map[string]map[string]DbtModelColumn // array_name -> columns for nested views
	ExcludedColumns   map[string]DbtModelColumn            // Columns excluded from all views
}

// HierarchyInfo contains information about column hierarchy
type HierarchyInfo struct {
	Children []string
	IsArray  bool
	Column   *DbtModelColumn
}

// FromModel creates column collections from a dbt model with optimized processing
func NewColumnCollections(model *DbtModel, arrayModels []string) *ColumnCollections {
	if arrayModels == nil {
		arrayModels = []string{}
	}

	// Get all columns from the model
	allColumns := model.Columns
	
	

	// Build hierarchy map for proper nested array detection
	hierarchy := buildHierarchyMap(allColumns)

	// Convert array_models to string names and find all array columns
	arrayModelNames := make(map[string]bool)
	for _, name := range arrayModels {
		arrayModelNames[name] = true
	}

	// Find all array columns (including nested ones) from hierarchy
	for colName, colInfo := range hierarchy {
		if colInfo.IsArray && colInfo.Column != nil {
			arrayModelNames[colName] = true
			fmt.Printf("DEBUG HIERARCHY: Found ARRAY column %s in hierarchy\n", colName)
		}
	}

	// Single-pass column classification with proper nested array handling
	mainViewColumns := make(map[string]DbtModelColumn)
	nestedViewColumns := make(map[string]map[string]DbtModelColumn)
	excludedColumns := make(map[string]DbtModelColumn)

	for colName, column := range allColumns {
		// Check if column should be excluded from all views
		if shouldExcludeFromAllViews(column, hierarchy) {
			excludedColumns[colName] = column
			continue
		}

		// Find the most specific array parent
		arrayParent := findArrayParent(colName, arrayModelNames)
		

		// Array parent columns need special handling
		if arrayModelNames[colName] {
			// Check if this array has child columns
			hasChildren := false
			for otherName := range allColumns {
				if strings.HasPrefix(otherName, colName+".") {
					hasChildren = true
					break
				}
			}

			// Check if this array is itself a child of another array
			isNestedArray := arrayParent != "" && arrayModelNames[arrayParent]
			

			if hasChildren {
				// Array with children (ARRAY<STRUCT>): only create nested view, don't add to main view
				// The nested view generation will handle creating the hidden reference dimension in main view
				if nestedViewColumns[colName] == nil {
					nestedViewColumns[colName] = make(map[string]DbtModelColumn)
				}
				// For ARRAY<STRUCT> fields, add the array field itself to its nested view as a hidden dimension
				nestedViewColumns[colName][colName] = column
			} else {
				// Array without children (pure ARRAY): only create nested view, don't add to main view
				// The nested view generation will handle creating the hidden reference dimension in main view
				if nestedViewColumns[colName] == nil {
					nestedViewColumns[colName] = make(map[string]DbtModelColumn)
				}
				// For pure ARRAY fields, add the array field itself to its nested view
				nestedViewColumns[colName][colName] = column
			}

			// If this array is nested under another array, also add it to the parent's nested view
			if isNestedArray && arrayParent != "" {
				if nestedViewColumns[arrayParent] == nil {
					nestedViewColumns[arrayParent] = make(map[string]DbtModelColumn)
				}
				nestedViewColumns[arrayParent][colName] = column
			}
		} else if arrayParent != "" {
			// This column belongs to a nested view
			if nestedViewColumns[arrayParent] == nil {
				nestedViewColumns[arrayParent] = make(map[string]DbtModelColumn)
			}
			nestedViewColumns[arrayParent][colName] = column
		} else {
			// This column belongs to the main view
			mainViewColumns[colName] = column
		}
	}


	return &ColumnCollections{
		MainViewColumns:   mainViewColumns,
		NestedViewColumns: nestedViewColumns,
		ExcludedColumns:   excludedColumns,
	}
}

// buildHierarchyMap builds a map of parent -> children relationships based on dot notation
func buildHierarchyMap(columns map[string]DbtModelColumn) map[string]*HierarchyInfo {
	fmt.Printf("DEBUG HIERARCHY: Building hierarchy for %d columns\n", len(columns))
	hierarchy := make(map[string]*HierarchyInfo)

	// First pass: create all hierarchy entries
	for _, col := range columns {
		parts := strings.Split(col.Name, ".")
		for i := range parts {
			parentPath := strings.Join(parts[:i+1], ".")
			if hierarchy[parentPath] == nil {
				hierarchy[parentPath] = &HierarchyInfo{
					Children: []string{},
					IsArray:  false,
					Column:   nil,
				}
			}
		}
	}

	// Second pass: set correct column references and array flags
	arrayCount := 0
	for _, col := range columns {
		if info, exists := hierarchy[col.Name]; exists {
			// Debug: check for duplicate column assignments
			if strings.Contains(col.Name, "sales") && info.Column != nil {
				fmt.Printf("DEBUG HIERARCHY DUPLICATE: Column '%s' already has a Column assigned! Existing: %p, New: %p\n", 
					col.Name, info.Column, &col)
			}
			info.Column = &col
			// Only mark as array if the data type starts with ARRAY
			if col.DataType != nil {
				dataTypeUpper := strings.ToUpper(*col.DataType)
				info.IsArray = strings.HasPrefix(dataTypeUpper, "ARRAY")
				// Debug: log array detection
				if info.IsArray {
					fmt.Printf("DEBUG HIERARCHY BUILD: Found ARRAY column: %s with type: %s\n", col.Name, *col.DataType)
					arrayCount++
				}
			} else {
				// Debug: log missing data type for first few columns
				if arrayCount < 3 {
					fmt.Printf("DEBUG HIERARCHY BUILD: Column %s has no DataType\n", col.Name)
				}
				info.IsArray = false
			}
		}
	}
	fmt.Printf("DEBUG HIERARCHY: Found %d ARRAY columns in hierarchy\n", arrayCount)

	// Third pass: build child relationships
	for _, col := range columns {
		parts := strings.Split(col.Name, ".")
		for i := 0; i < len(parts)-1; i++ {
			parentPath := strings.Join(parts[:i+1], ".")
			childPath := strings.Join(parts[:i+2], ".")
			if parentInfo, exists := hierarchy[parentPath]; exists {
				// Check if child already exists to avoid duplicates
				found := false
				for _, existing := range parentInfo.Children {
					if existing == childPath {
						found = true
						break
					}
				}
				if !found {
					parentInfo.Children = append(parentInfo.Children, childPath)
				}
			}
		}
	}

	return hierarchy
}

// shouldExcludeFromAllViews checks if a column should be excluded from all views
func shouldExcludeFromAllViews(column DbtModelColumn, hierarchy map[string]*HierarchyInfo) bool {
	// Exclude STRUCT parents that have children (but not ARRAY<STRUCT>)
	if column.DataType != nil {
		dataTypeUpper := strings.ToUpper(*column.DataType)
		if strings.Contains(dataTypeUpper, "STRUCT") && !strings.HasPrefix(dataTypeUpper, "ARRAY") {
			// Check if this STRUCT has nested children
			for otherPath := range hierarchy {
				if strings.HasPrefix(otherPath, column.Name+".") {
					return true
				}
			}
		}
	}
	return false
}

// findArrayParent finds which array model this column belongs to, if any
func findArrayParent(colName string, arrayModelNames map[string]bool) string {
	// Find the most specific (longest) array parent that matches
	var matchingArrays []string
	for arrayName := range arrayModelNames {
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

// GetArrayModels extracts all array models from the column collections
func (cc *ColumnCollections) GetArrayModels() []string {
	var arrayModels []string
	for arrayName := range cc.NestedViewColumns {
		arrayModels = append(arrayModels, arrayName)
	}
	return arrayModels
}
