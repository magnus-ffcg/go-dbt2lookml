package models

// NewColumnCollectionsV2 creates column collections using the refactored services.
// This is a cleaner implementation that delegates to specialized services.
func NewColumnCollectionsV2(model *DbtModel, arrayModels []string) *ColumnCollections {
	if arrayModels == nil {
		arrayModels = []string{}
	}

	// Step 1: Build hierarchy to understand parent-child relationships
	hierarchy := NewColumnHierarchy(model.Columns)

	// Step 2: Build set of array columns (from config + detected from hierarchy)
	arrayColumns := buildArrayColumnSet(arrayModels, hierarchy)

	// Step 3: Create classifier with business rules
	classifier := NewColumnClassifier(hierarchy, arrayColumns)

	// Step 4: Initialize result collections
	collections := &ColumnCollections{
		MainViewColumns:   make(map[string]DbtModelColumn),
		NestedViewColumns: make(map[string]map[string]DbtModelColumn),
		ExcludedColumns:   make(map[string]DbtModelColumn),
	}

	// Step 5: Classify each column and place it in the appropriate collection
	for colName, column := range model.Columns {
		category := classifier.Classify(colName, column)

		switch category {
		case CategoryExcluded:
			collections.ExcludedColumns[colName] = column

		case CategoryMainView:
			collections.MainViewColumns[colName] = column

		case CategoryNestedView:
			handleNestedViewColumn(collections, classifier, colName, column, arrayColumns)
		}
	}

	return collections
}

// buildArrayColumnSet creates a set of all array column names.
func buildArrayColumnSet(arrayModels []string, hierarchy *ColumnHierarchy) map[string]bool {
	arrayColumns := make(map[string]bool)

	// Add explicitly configured array models
	for _, name := range arrayModels {
		arrayColumns[name] = true
	}

	// Add arrays detected from hierarchy
	for colName, colInfo := range hierarchy.All() {
		if colInfo.IsArray && colInfo.Column != nil {
			arrayColumns[colName] = true
		}
	}

	return arrayColumns
}

// handleNestedViewColumn handles the complex logic for placing columns in nested views.
func handleNestedViewColumn(
	collections *ColumnCollections,
	classifier *ColumnClassifier,
	colName string,
	column DbtModelColumn,
	arrayColumns map[string]bool,
) {
	// If this is an array column itself
	if arrayColumns[colName] {
		// Initialize nested view for this array
		if collections.NestedViewColumns[colName] == nil {
			collections.NestedViewColumns[colName] = make(map[string]DbtModelColumn)
		}

		// Add the array field itself to its nested view
		// (both ARRAY<STRUCT> and pure ARRAY types)
		collections.NestedViewColumns[colName][colName] = column

		// If this array is nested under another array, also add it to parent's nested view
		if classifier.IsNestedArray(colName) {
			arrayParent := classifier.GetArrayParent(colName)
			if arrayParent != "" {
				if collections.NestedViewColumns[arrayParent] == nil {
					collections.NestedViewColumns[arrayParent] = make(map[string]DbtModelColumn)
				}
				collections.NestedViewColumns[arrayParent][colName] = column
			}
		}
	} else {
		// This column belongs to a nested view (child of an array)
		arrayParent := classifier.GetArrayParent(colName)
		if arrayParent != "" {
			if collections.NestedViewColumns[arrayParent] == nil {
				collections.NestedViewColumns[arrayParent] = make(map[string]DbtModelColumn)
			}
			collections.NestedViewColumns[arrayParent][colName] = column
		}
	}
}
