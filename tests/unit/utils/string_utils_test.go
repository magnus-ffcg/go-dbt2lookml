package utils

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
)

// TestStringUtilsExtended demonstrates comprehensive string utility testing
func TestStringUtilsExtended(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		testFunc func(string) string
	}{
		{"camel to snake - simple", "CamelCase", "camel_case", utils.CamelToSnake},
		{"camel to snake - single word", "Simple", "simple", utils.CamelToSnake},
		{"camel to snake - already snake", "already_snake", "already_snake", utils.CamelToSnake},
		{"camel to snake - empty", "", "", utils.CamelToSnake},
		{"camel to snake - complex", "VeryLongCamelCaseString", "very_long_camel_case_string", utils.CamelToSnake},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestColumnQuoting demonstrates column name quoting logic
func TestColumnQuoting(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple column", "simple_column", "simple_column"},
		{"column with spaces", "column with spaces", "`column with spaces`"},
		{"column with dashes", "column-with-dashes", "`column-with-dashes`"},
		{"column with special chars", "column@special", "`column@special`"},
		{"already quoted", "`already_quoted`", "``already_quoted``"}, // Current behavior - double quotes
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.QuoteColumnNameIfNeeded(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSanitizeIdentifier demonstrates identifier sanitization
func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"valid identifier", "valid_identifier", "valid_identifier"},
		{"spaces to underscores", "column with spaces", "column_with_spaces"},
		{"special chars removed", "column@with#special$chars", "column_with_special_chars"},
		{"multiple spaces", "column  with   spaces", "column_with_spaces"},
		{"leading/trailing spaces", " column ", "column"},
		{"numbers preserved", "column123", "column123"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := utils.SanitizeIdentifier(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestEdgeCases demonstrates edge case handling
func TestEdgeCases(t *testing.T) {
	t.Run("nil safety", func(t *testing.T) {
		// Test that functions handle edge cases gracefully
		assert.Equal(t, "", utils.CamelToSnake(""))
		assert.Equal(t, "", utils.QuoteColumnNameIfNeeded(""))
		assert.Equal(t, "", utils.SanitizeIdentifier(""))
	})

	t.Run("unicode handling", func(t *testing.T) {
		// Test unicode character handling
		result := utils.SanitizeIdentifier("column_with_ñ")
		assert.NotEmpty(t, result)
	})

	t.Run("very long strings", func(t *testing.T) {
		longString := "VeryLongCamelCaseStringThatGoesOnAndOnAndOn"
		result := utils.CamelToSnake(longString)
		assert.Contains(t, result, "_")
		assert.True(t, len(result) > len(longString))
	})
}
