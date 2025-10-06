// Package utils provides utility functions for string manipulation and pointer helpers.
//
// This package contains reusable utility functions used throughout the application
// for common operations like string transformation, sanitization, and pointer creation.
//
// String Utilities:
//   - CamelToSnake: Converts CamelCase to snake_case (handles acronyms correctly)
//   - SnakeToCamel: Converts snake_case to CamelCase
//   - ToLookMLName: Converts column names to LookML-compatible identifiers
//   - SanitizeIdentifier: Removes invalid characters from identifiers
//   - QuoteColumnNameIfNeeded: Adds backticks for names requiring quoting
//
// Pointer Helpers:
//   - StringPtr, BoolPtr, IntPtr, Int64Ptr, Float64Ptr: Create pointers to literal values
//
// String Manipulation:
//   - TruncateString: Safely truncate strings to maximum length
//   - Pluralize: Simple English pluralization
//   - ToTitleCase: Convert to title case
//
// The package uses pre-compiled regular expressions for optimal performance
// in frequently-called string transformation functions.
package utils

import (
	"regexp"
	"strings"
	"unicode"
)

// Pre-compiled regular expressions for better performance
// These are compiled once at package initialization
var (
	// acronymPatterns handles acronym-word combinations like GTINId, GTINType
	// Inserts underscore between consecutive uppercase letters followed by lowercase
	acronymPatterns = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)

	// insertUnderscoreBeforeUppercase inserts underscore before uppercase letters that follow lowercase letters or digits
	insertUnderscoreBeforeUppercase = regexp.MustCompile(`([a-z0-9])([A-Z])`)

	// camelCasePatterns handles remaining CamelCase patterns
	camelCasePatterns = regexp.MustCompile(`(.)([A-Z][a-z]+)`)

	// cleanupUnderscores cleans up multiple consecutive underscores
	cleanupUnderscores = regexp.MustCompile(`_+`)

	// sanitizeInvalidChars removes or replaces invalid characters from identifiers
	sanitizeInvalidChars = regexp.MustCompile(`[^a-zA-Z0-9_]`)

	// consecutiveUnderscore Reg removes consecutive underscores
	consecutiveUnderscore = regexp.MustCompile(`_+`)

	// threeOrMoreUnderscores replaces 3+ underscores with double underscores
	threeOrMoreUnderscores = regexp.MustCompile(`_{3,}`)
)

// CamelToSnake converts CamelCase to snake_case
// Handles complex cases like SupplierInformation -> supplier_information, GTINId -> gtin_id
func CamelToSnake(s string) string {
	if s == "" {
		return s
	}

	s1 := acronymPatterns.ReplaceAllString(s, `${1}_${2}`)
	s2 := insertUnderscoreBeforeUppercase.ReplaceAllString(s1, `${1}_${2}`)
	s3 := camelCasePatterns.ReplaceAllString(s2, `${1}_${2}`)
	s4 := cleanupUnderscores.ReplaceAllString(s3, `_`)

	return strings.ToLower(s4)
}

// SnakeToCamel converts snake_case to CamelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	var result strings.Builder

	for _, part := range parts {
		if len(part) > 0 {
			// Capitalize first letter, rest stays as-is
			result.WriteString(strings.ToUpper(string(part[0])) + part[1:])
		}
	}

	return result.String()
}

// QuoteColumnNameIfNeeded adds backticks around column names that need quoting
func QuoteColumnNameIfNeeded(columnName string) string {
	// Check if the column name contains spaces, special characters, or non-ASCII characters
	// Check for spaces
	needsQuoting := strings.Contains(columnName, " ")

	// Check for special characters (anything that's not alphanumeric or underscore)
	for _, r := range columnName {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			needsQuoting = true
			break
		}
		// Check for non-ASCII characters
		if r > 127 {
			needsQuoting = true
			break
		}
	}

	if needsQuoting {
		return "`" + columnName + "`"
	}

	return columnName
}

// SanitizeIdentifier removes or replaces invalid characters from identifiers
func SanitizeIdentifier(s string) string {
	// Replace spaces and special characters with underscores (using pre-compiled regex)
	sanitized := sanitizeInvalidChars.ReplaceAllString(s, "_")

	// Remove consecutive underscores (using pre-compiled regex)
	sanitized = consecutiveUnderscore.ReplaceAllString(sanitized, "_")

	// Remove leading/trailing underscores
	sanitized = strings.Trim(sanitized, "_")

	// Ensure it doesn't start with a number
	if len(sanitized) > 0 && unicode.IsDigit(rune(sanitized[0])) {
		sanitized = "_" + sanitized
	}

	return sanitized
}

// TruncateString truncates a string to a maximum length
func TruncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength]
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// ContainsAny checks if a string contains any of the given substrings
func ContainsAny(s string, substrings []string) bool {
	for _, substring := range substrings {
		if strings.Contains(s, substring) {
			return true
		}
	}
	return false
}

// RemovePrefix removes a prefix from a string if it exists
func RemovePrefix(s, prefix string) string {
	if strings.HasPrefix(s, prefix) {
		return s[len(prefix):]
	}
	return s
}

// RemoveSuffix removes a suffix from a string if it exists
func RemoveSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		return s[:len(s)-len(suffix)]
	}
	return s
}

// SplitAndTrim splits a string by delimiter and trims whitespace from each part
func SplitAndTrim(s, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	var result []string

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// JoinNonEmpty joins non-empty strings with a delimiter
func JoinNonEmpty(parts []string, delimiter string) string {
	var nonEmpty []string

	for _, part := range parts {
		if !IsEmpty(part) {
			nonEmpty = append(nonEmpty, part)
		}
	}

	return strings.Join(nonEmpty, delimiter)
}

// Pluralize adds 's' to a word (simple pluralization)
func Pluralize(word string) string {
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") ||
		strings.HasSuffix(word, "z") || strings.HasSuffix(word, "ch") ||
		strings.HasSuffix(word, "sh") {
		return word + "es"
	}

	if strings.HasSuffix(word, "y") && len(word) > 1 {
		if !isVowel(rune(word[len(word)-2])) {
			return word[:len(word)-1] + "ies"
		}
	}

	return word + "s"
}

// ToTitleCase converts a string from snake_case or kebab-case to Title Case.
func ToTitleCase(s string) string {
	if s == "" {
		return ""
	}

	// Replace underscores and hyphens with spaces
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")

	words := strings.Fields(s)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// ToLookMLName converts a column name to LookML naming convention.
// Converts PascalCase to snake_case and replaces dots with double underscores.
//
// Note: This function expects proper PascalCase or camelCase input. If your column
// names are already lowercase (e.g., "supplierinformation"), they will remain as-is
// since there's no case information to work with. Ensure your source data maintains
// proper case conventions (e.g., "SupplierInformation") for correct conversion.
func ToLookMLName(s string) string {
	if s == "" {
		return s
	}

	// Split by dots to handle nested column names
	parts := strings.Split(s, ".")
	var convertedParts []string

	for _, part := range parts {
		// Convert each part from PascalCase/camelCase to snake_case
		snakePart := CamelToSnake(part)
		convertedParts = append(convertedParts, snakePart)
	}

	// Join with double underscores (LookML convention for nested fields)
	lookmlName := strings.Join(convertedParts, "__")

	// Sanitize any remaining special characters (but preserve underscores) - using pre-compiled regex
	lookmlName = sanitizeInvalidChars.ReplaceAllString(lookmlName, "_")

	// Remove consecutive underscores (but preserve double underscores)
	// Replace 3 or more underscores with double underscores - using pre-compiled regex
	lookmlName = threeOrMoreUnderscores.ReplaceAllString(lookmlName, "__")

	// Remove leading/trailing underscores
	lookmlName = strings.Trim(lookmlName, "_")

	// Ensure it doesn't start with a number
	if len(lookmlName) > 0 && unicode.IsDigit(rune(lookmlName[0])) {
		lookmlName = "_" + lookmlName
	}

	return lookmlName
}

// isVowel checks if a character is a vowel
func isVowel(r rune) bool {
	vowels := "aeiouAEIOU"
	return strings.ContainsRune(vowels, r)
}
