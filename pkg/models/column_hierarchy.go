package models

import "strings"

// ColumnHierarchy represents the parent-child relationships between columns.
// It's used to understand nested STRUCT and ARRAY structures in BigQuery.
type ColumnHierarchy struct {
	hierarchy map[string]*HierarchyInfo
}

// HierarchyInfo contains information about a column's position in the hierarchy.
type HierarchyInfo struct {
	Children []string        // Child column paths
	IsArray  bool            // Whether this column is an ARRAY type
	Column   *DbtModelColumn // Reference to the actual column
}

// NewColumnHierarchy builds a hierarchy from a flat map of columns.
// It analyzes dot-separated paths to build parent-child relationships.
func NewColumnHierarchy(columns map[string]DbtModelColumn) *ColumnHierarchy {
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
	for _, col := range columns {
		if info, exists := hierarchy[col.Name]; exists {
			info.Column = &col
			// Only mark as array if the data type starts with ARRAY
			if col.DataType != nil {
				dataTypeUpper := strings.ToUpper(*col.DataType)
				info.IsArray = strings.HasPrefix(dataTypeUpper, "ARRAY")
			} else {
				info.IsArray = false
			}
		}
	}

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

	return &ColumnHierarchy{hierarchy: hierarchy}
}

// Get returns the hierarchy info for a given column path.
func (h *ColumnHierarchy) Get(path string) *HierarchyInfo {
	return h.hierarchy[path]
}

// HasChildren checks if a column has any child columns.
func (h *ColumnHierarchy) HasChildren(path string) bool {
	info := h.hierarchy[path]
	return info != nil && len(info.Children) > 0
}

// IsArray checks if a column is an ARRAY type.
func (h *ColumnHierarchy) IsArray(path string) bool {
	info := h.hierarchy[path]
	return info != nil && info.IsArray
}

// GetChildren returns all child paths for a given column.
func (h *ColumnHierarchy) GetChildren(path string) []string {
	info := h.hierarchy[path]
	if info == nil {
		return []string{}
	}
	return info.Children
}

// All returns the entire hierarchy map.
func (h *ColumnHierarchy) All() map[string]*HierarchyInfo {
	return h.hierarchy
}
