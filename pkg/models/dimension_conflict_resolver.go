package models

import (
	"fmt"
	"log"
)

// DimensionConflictResolver handles conflicts between dimensions and dimension groups.
//
// When a dimension has the same name as a dimension group or one of its timeframe
// variations, it creates a naming conflict in LookML. This resolver detects such
// conflicts and renames the conflicting dimensions by adding a "_conflict" suffix
// and marking them as hidden.
//
// Example conflict:
//   - Dimension: "created_date" (from STRUCT field or nested column)
//   - Dimension Group: "created" with timeframe "date" → generates "created_date"
//   - Resolution: Rename dimension to "created_date_conflict" and hide it
type DimensionConflictResolver struct {
	// conflictSuffix is the suffix added to conflicting dimension names
	conflictSuffix string

	// hideConflicts determines if conflicting dimensions should be hidden
	hideConflicts bool

	// logConflicts determines if conflicts should be logged
	logConflicts bool
}

// NewDimensionConflictResolver creates a new resolver with default settings.
func NewDimensionConflictResolver() *DimensionConflictResolver {
	return &DimensionConflictResolver{
		conflictSuffix: "_conflict",
		hideConflicts:  true,
		logConflicts:   true,
	}
}

// NewDimensionConflictResolverWithOptions creates a resolver with custom settings.
func NewDimensionConflictResolverWithOptions(suffix string, hideConflicts, logConflicts bool) *DimensionConflictResolver {
	return &DimensionConflictResolver{
		conflictSuffix: suffix,
		hideConflicts:  hideConflicts,
		logConflicts:   logConflicts,
	}
}

// Resolve detects and resolves naming conflicts between dimensions and dimension groups.
// Returns a new slice with conflicting dimensions renamed and optionally hidden.
func (r *DimensionConflictResolver) Resolve(
	dimensions []LookMLDimension,
	dimensionGroups []LookMLDimensionGroup,
	modelName string,
) []LookMLDimension {
	// Build set of reserved names from dimension groups
	reservedNames := r.buildReservedNames(dimensionGroups)

	// Check if there are any actual conflicts
	if !r.hasConflicts(dimensions, reservedNames) {
		return dimensions
	}

	// Resolve conflicts by renaming
	return r.renameConflicts(dimensions, reservedNames, modelName)
}

// buildReservedNames creates a set of all names that dimension groups will generate.
func (r *DimensionConflictResolver) buildReservedNames(dimensionGroups []LookMLDimensionGroup) map[string]bool {
	reserved := make(map[string]bool)

	for _, group := range dimensionGroups {
		// Add the base dimension group name
		reserved[group.Name] = true

		// Add all timeframe variations
		// e.g., "created" with timeframes [date, time, week] generates:
		// created_date, created_time, created_week
		if group.Timeframes != nil {
			for _, timeframe := range group.Timeframes {
				generatedName := fmt.Sprintf("%s_%s", group.Name, timeframe)
				reserved[generatedName] = true
			}
		}
	}

	return reserved
}

// hasConflicts checks if any dimensions conflict with reserved names.
func (r *DimensionConflictResolver) hasConflicts(dimensions []LookMLDimension, reservedNames map[string]bool) bool {
	for _, dimension := range dimensions {
		if reservedNames[dimension.Name] {
			return true
		}
	}
	return false
}

// renameConflicts creates a new slice with conflicting dimensions renamed.
func (r *DimensionConflictResolver) renameConflicts(
	dimensions []LookMLDimension,
	reservedNames map[string]bool,
	modelName string,
) []LookMLDimension {
	result := make([]LookMLDimension, 0, len(dimensions))

	for _, dimension := range dimensions {
		if reservedNames[dimension.Name] {
			// Conflict detected - create modified copy
			originalName := dimension.Name
			dimension.Name = fmt.Sprintf("%s%s", originalName, r.conflictSuffix)

			// Hide the conflicting dimension if configured
			if r.hideConflicts {
				hidden := true
				dimension.Hidden = &hidden
			}

			// Log the conflict if configured
			if r.logConflicts {
				log.Printf("Renamed conflicting dimension '%s' to '%s' in model '%s'",
					originalName, dimension.Name, modelName)
			}
		}

		result = append(result, dimension)
	}

	return result
}

// GetConflictingDimensions returns a list of dimension names that conflict with dimension groups.
// Useful for debugging or reporting.
func (r *DimensionConflictResolver) GetConflictingDimensions(
	dimensions []LookMLDimension,
	dimensionGroups []LookMLDimensionGroup,
) []string {
	reservedNames := r.buildReservedNames(dimensionGroups)
	var conflicts []string

	for _, dimension := range dimensions {
		if reservedNames[dimension.Name] {
			conflicts = append(conflicts, dimension.Name)
		}
	}

	return conflicts
}

// GetReservedNames returns all names that are reserved by dimension groups.
// Useful for validation or documentation.
func (r *DimensionConflictResolver) GetReservedNames(dimensionGroups []LookMLDimensionGroup) []string {
	reservedMap := r.buildReservedNames(dimensionGroups)
	reserved := make([]string, 0, len(reservedMap))

	for name := range reservedMap {
		reserved = append(reserved, name)
	}

	return reserved
}
