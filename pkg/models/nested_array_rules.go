package models

import "strings"

// MaxNestedArrayDepth defines the maximum nesting level for array processing.
//
// Nesting levels:
//   - Level 1: items (0 dots in path)
//   - Level 2: items.subitems (1 dot in path)
//   - Level 3: items.subitems.details (2 dots in path)
//   - Level 4+: Considered too deeply nested (3+ dots)
//
// Arrays beyond Level 3 are typically too complex for efficient querying
// and can cause performance issues in most BI tools.
const MaxNestedArrayDepth = 3

// NestedArrayRules encapsulates business rules for nested array processing.
// It determines which nested arrays should be processed based on their depth
// in the column hierarchy.
type NestedArrayRules struct {
	maxDepth int
}

// NewNestedArrayRules creates a new NestedArrayRules instance with default settings.
func NewNestedArrayRules() *NestedArrayRules {
	return &NestedArrayRules{
		maxDepth: MaxNestedArrayDepth,
	}
}

// NewNestedArrayRulesWithDepth creates a new NestedArrayRules instance with custom max depth.
func NewNestedArrayRulesWithDepth(maxDepth int) *NestedArrayRules {
	return &NestedArrayRules{
		maxDepth: maxDepth,
	}
}

// ShouldProcessArray determines if an array at the given path should be processed
// based on its nesting depth.
//
// Examples:
//   - "items" (level 1) → true
//   - "items.subitems" (level 2) → true
//   - "items.subitems.details" (level 3) → true
//   - "items.subitems.details.meta" (level 4) → false
//   - "" (empty) → false
func (r *NestedArrayRules) ShouldProcessArray(arrayPath string) bool {
	if arrayPath == "" {
		return false
	}
	depth := r.GetArrayDepth(arrayPath)
	return depth <= r.maxDepth
}

// GetArrayDepth returns the nesting level of an array based on its path.
// The depth is calculated by counting dots in the path and adding 1.
//
// Examples:
//   - "items" → 1 (0 dots)
//   - "items.subitems" → 2 (1 dot)
//   - "items.subitems.details" → 3 (2 dots)
//   - "items.subitems.details.meta" → 4 (3 dots)
func (r *NestedArrayRules) GetArrayDepth(arrayPath string) int {
	if arrayPath == "" {
		return 0
	}
	dotCount := strings.Count(arrayPath, ".")
	return dotCount + 1
}

// IsValidArrayPath checks if the array path is valid (non-empty and not too deep).
func (r *NestedArrayRules) IsValidArrayPath(arrayPath string) bool {
	if arrayPath == "" {
		return false
	}
	return r.ShouldProcessArray(arrayPath)
}

// GetMaxDepth returns the maximum allowed nesting depth.
func (r *NestedArrayRules) GetMaxDepth() int {
	return r.maxDepth
}
