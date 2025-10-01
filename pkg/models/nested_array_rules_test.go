package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNestedArrayRules_ShouldProcessArray(t *testing.T) {
	tests := []struct {
		name       string
		arrayPath  string
		maxDepth   int
		shouldPass bool
	}{
		{
			name:       "level 1 array (no dots)",
			arrayPath:  "items",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: true,
		},
		{
			name:       "level 2 array (1 dot)",
			arrayPath:  "items.subitems",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: true,
		},
		{
			name:       "level 3 array (2 dots)",
			arrayPath:  "items.subitems.details",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: true,
		},
		{
			name:       "level 4 array (3 dots) - too deep",
			arrayPath:  "items.subitems.details.meta",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: false,
		},
		{
			name:       "level 5 array (4 dots) - way too deep",
			arrayPath:  "items.subitems.details.meta.extra",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: false,
		},
		{
			name:       "custom max depth 2 - level 2 passes",
			arrayPath:  "items.subitems",
			maxDepth:   2,
			shouldPass: true,
		},
		{
			name:       "custom max depth 2 - level 3 fails",
			arrayPath:  "items.subitems.details",
			maxDepth:   2,
			shouldPass: false,
		},
		{
			name:       "empty path",
			arrayPath:  "",
			maxDepth:   MaxNestedArrayDepth,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewNestedArrayRulesWithDepth(tt.maxDepth)
			result := rules.ShouldProcessArray(tt.arrayPath)
			assert.Equal(t, tt.shouldPass, result, "ShouldProcessArray returned unexpected result")
		})
	}
}

func TestNestedArrayRules_GetArrayDepth(t *testing.T) {
	tests := []struct {
		name          string
		arrayPath     string
		expectedDepth int
	}{
		{
			name:          "empty path",
			arrayPath:     "",
			expectedDepth: 0,
		},
		{
			name:          "level 1 - no dots",
			arrayPath:     "items",
			expectedDepth: 1,
		},
		{
			name:          "level 2 - one dot",
			arrayPath:     "items.subitems",
			expectedDepth: 2,
		},
		{
			name:          "level 3 - two dots",
			arrayPath:     "items.subitems.details",
			expectedDepth: 3,
		},
		{
			name:          "level 4 - three dots",
			arrayPath:     "items.subitems.details.meta",
			expectedDepth: 4,
		},
		{
			name:          "level 5 - four dots",
			arrayPath:     "a.b.c.d.e",
			expectedDepth: 5,
		},
		{
			name:          "complex nested path",
			arrayPath:     "order_items.line_items.discounts.rules.conditions",
			expectedDepth: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewNestedArrayRules()
			depth := rules.GetArrayDepth(tt.arrayPath)
			assert.Equal(t, tt.expectedDepth, depth, "GetArrayDepth returned unexpected depth")
		})
	}
}

func TestNestedArrayRules_IsValidArrayPath(t *testing.T) {
	tests := []struct {
		name      string
		arrayPath string
		valid     bool
	}{
		{
			name:      "valid level 1 path",
			arrayPath: "items",
			valid:     true,
		},
		{
			name:      "valid level 2 path",
			arrayPath: "items.subitems",
			valid:     true,
		},
		{
			name:      "valid level 3 path",
			arrayPath: "items.subitems.details",
			valid:     true,
		},
		{
			name:      "invalid - too deep (level 4)",
			arrayPath: "items.subitems.details.meta",
			valid:     false,
		},
		{
			name:      "invalid - empty path",
			arrayPath: "",
			valid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules := NewNestedArrayRules()
			result := rules.IsValidArrayPath(tt.arrayPath)
			assert.Equal(t, tt.valid, result, "IsValidArrayPath returned unexpected result")
		})
	}
}

func TestNestedArrayRules_GetMaxDepth(t *testing.T) {
	t.Run("default max depth", func(t *testing.T) {
		rules := NewNestedArrayRules()
		assert.Equal(t, MaxNestedArrayDepth, rules.GetMaxDepth())
	})

	t.Run("custom max depth", func(t *testing.T) {
		customDepth := 5
		rules := NewNestedArrayRulesWithDepth(customDepth)
		assert.Equal(t, customDepth, rules.GetMaxDepth())
	})
}

func TestNestedArrayRules_DefaultConstant(t *testing.T) {
	t.Run("MaxNestedArrayDepth constant value", func(t *testing.T) {
		// Verify the constant is set to 3 as per business rules
		assert.Equal(t, 3, MaxNestedArrayDepth, "MaxNestedArrayDepth should be 3")
	})
}

// TestNestedArrayRules_RealWorldScenarios tests realistic BigQuery column paths
func TestNestedArrayRules_RealWorldScenarios(t *testing.T) {
	rules := NewNestedArrayRules()

	tests := []struct {
		name       string
		arrayPath  string
		shouldPass bool
		depth      int
	}{
		{
			name:       "BigQuery ARRAY<STRING>",
			arrayPath:  "tags",
			shouldPass: true,
			depth:      1,
		},
		{
			name:       "BigQuery ARRAY<STRUCT>",
			arrayPath:  "line_items",
			shouldPass: true,
			depth:      1,
		},
		{
			name:       "Nested ARRAY in STRUCT",
			arrayPath:  "order_details.items",
			shouldPass: true,
			depth:      2,
		},
		{
			name:       "Deep nested ARRAY",
			arrayPath:  "customer.addresses.history",
			shouldPass: true,
			depth:      3,
		},
		{
			name:       "Too deep - should skip",
			arrayPath:  "customer.orders.items.components",
			shouldPass: false,
			depth:      4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.shouldPass, rules.ShouldProcessArray(tt.arrayPath))
			assert.Equal(t, tt.depth, rules.GetArrayDepth(tt.arrayPath))
		})
	}
}
