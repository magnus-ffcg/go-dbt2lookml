package models

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDimensionConflictResolver_NoConflicts(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "id", Type: "number"},
		{Name: "name", Type: "string"},
		{Name: "amount", Type: "number"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate, enums.TimeFrameTime},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	// No conflicts, so dimensions should be unchanged
	assert.Equal(t, len(dimensions), len(result))
	assert.Equal(t, "id", result[0].Name)
	assert.Equal(t, "name", result[1].Name)
	assert.Equal(t, "amount", result[2].Name)
}

func TestDimensionConflictResolver_WithConflicts(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "id", Type: "number"},
		{Name: "created_date", Type: "string"}, // Conflicts with dimension group timeframe
		{Name: "name", Type: "string"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate, enums.TimeFrameTime},
			// This generates: created_date, created_time
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 3, len(result))

	// Non-conflicting dimensions unchanged
	assert.Equal(t, "id", result[0].Name)
	assert.Nil(t, result[0].Hidden)

	// Conflicting dimension renamed and hidden
	assert.Equal(t, "created_date_conflict", result[1].Name)
	require.NotNil(t, result[1].Hidden)
	assert.True(t, *result[1].Hidden)

	// Non-conflicting dimension unchanged
	assert.Equal(t, "name", result[2].Name)
	assert.Nil(t, result[2].Hidden)
}

func TestDimensionConflictResolver_MultipleConflicts(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "created_date", Type: "string"}, // Conflicts
		{Name: "created_time", Type: "string"}, // Conflicts
		{Name: "updated_week", Type: "string"}, // Conflicts
		{Name: "id", Type: "number"},           // No conflict
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate, enums.TimeFrameTime},
		},
		{
			Name:       "updated",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameWeek, enums.TimeFrameMonth},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 4, len(result))

	// All three conflicting dimensions should be renamed
	assert.Equal(t, "created_date_conflict", result[0].Name)
	assert.Equal(t, "created_time_conflict", result[1].Name)
	assert.Equal(t, "updated_week_conflict", result[2].Name)

	// Non-conflicting dimension unchanged
	assert.Equal(t, "id", result[3].Name)
}

func TestDimensionConflictResolver_BaseDimensionGroupName(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "created", Type: "string"}, // Conflicts with base dimension group name
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 1, len(result))
	assert.Equal(t, "created_conflict", result[0].Name)
	require.NotNil(t, result[0].Hidden)
	assert.True(t, *result[0].Hidden)
}

func TestDimensionConflictResolver_CustomSuffix(t *testing.T) {
	resolver := NewDimensionConflictResolverWithOptions("_dimension", true, false)

	dimensions := []LookMLDimension{
		{Name: "created_date", Type: "string"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 1, len(result))
	assert.Equal(t, "created_date_dimension", result[0].Name)
}

func TestDimensionConflictResolver_NoHiding(t *testing.T) {
	resolver := NewDimensionConflictResolverWithOptions("_conflict", false, false)

	dimensions := []LookMLDimension{
		{Name: "created_date", Type: "string"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 1, len(result))
	assert.Equal(t, "created_date_conflict", result[0].Name)
	assert.Nil(t, result[0].Hidden, "Dimension should not be hidden")
}

func TestDimensionConflictResolver_EmptyInputs(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	t.Run("empty dimensions", func(t *testing.T) {
		result := resolver.Resolve([]LookMLDimension{}, []LookMLDimensionGroup{{Name: "created"}}, "test")
		assert.Empty(t, result)
	})

	t.Run("empty dimension groups", func(t *testing.T) {
		dimensions := []LookMLDimension{{Name: "id", Type: "number"}}
		result := resolver.Resolve(dimensions, []LookMLDimensionGroup{}, "test")
		assert.Equal(t, dimensions, result)
	})

	t.Run("both empty", func(t *testing.T) {
		result := resolver.Resolve([]LookMLDimension{}, []LookMLDimensionGroup{}, "test")
		assert.Empty(t, result)
	})
}

func TestDimensionConflictResolver_GetConflictingDimensions(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "id", Type: "number"},
		{Name: "created_date", Type: "string"},
		{Name: "created_time", Type: "string"},
		{Name: "name", Type: "string"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate, enums.TimeFrameTime},
		},
	}

	conflicts := resolver.GetConflictingDimensions(dimensions, dimensionGroups)

	require.Equal(t, 2, len(conflicts))
	assert.Contains(t, conflicts, "created_date")
	assert.Contains(t, conflicts, "created_time")
}

func TestDimensionConflictResolver_GetReservedNames(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate, enums.TimeFrameTime, enums.TimeFrameWeek},
		},
		{
			Name:       "updated",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameMonth},
		},
	}

	reserved := resolver.GetReservedNames(dimensionGroups)

	// Should have 6 reserved names:
	// created, created_date, created_time, created_week, updated, updated_month
	assert.Equal(t, 6, len(reserved))
	assert.Contains(t, reserved, "created")
	assert.Contains(t, reserved, "created_date")
	assert.Contains(t, reserved, "created_time")
	assert.Contains(t, reserved, "created_week")
	assert.Contains(t, reserved, "updated")
	assert.Contains(t, reserved, "updated_month")
}

func TestDimensionConflictResolver_DimensionGroupWithoutTimeframes(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{Name: "created", Type: "string"},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: nil, // No timeframes
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	// Should still conflict with base name
	require.Equal(t, 1, len(result))
	assert.Equal(t, "created_conflict", result[0].Name)
}

func TestDimensionConflictResolver_PreservesOtherAttributes(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	dimensions := []LookMLDimension{
		{
			Name:        "created_date",
			Type:        "string",
			SQL:         "${TABLE}.created_date",
			Description: utils.StringPtr("Creation date"),
			Label:       utils.StringPtr("Created Date"),
		},
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name:       "created",
			Type:       "time",
			Timeframes: []enums.LookerTimeFrame{enums.TimeFrameDate},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "test_model")

	require.Equal(t, 1, len(result))

	// Name and Hidden should be modified
	assert.Equal(t, "created_date_conflict", result[0].Name)
	require.NotNil(t, result[0].Hidden)
	assert.True(t, *result[0].Hidden)

	// Other attributes should be preserved
	assert.Equal(t, "string", result[0].Type)
	assert.Equal(t, "${TABLE}.created_date", result[0].SQL)
	require.NotNil(t, result[0].Description)
	assert.Equal(t, "Creation date", *result[0].Description)
	require.NotNil(t, result[0].Label)
	assert.Equal(t, "Created Date", *result[0].Label)
}

// Test real-world scenario from BigQuery with nested STRUCT and TIMESTAMP fields
func TestDimensionConflictResolver_RealWorldScenario(t *testing.T) {
	resolver := NewDimensionConflictResolver()

	// Scenario: Model has nested STRUCT with fields that conflict with dimension groups
	dimensions := []LookMLDimension{
		{Name: "order_id", Type: "number"},
		{Name: "item_creation_date", Type: "string"}, // From nested STRUCT, conflicts
		{Name: "item_creation_time", Type: "string"}, // From nested STRUCT, conflicts
		{Name: "customer_name", Type: "string"},
		{Name: "shipping_updated_week", Type: "string"}, // From nested STRUCT, conflicts
	}

	dimensionGroups := []LookMLDimensionGroup{
		{
			Name: "item_creation",
			Type: "time",
			Timeframes: []enums.LookerTimeFrame{
				enums.TimeFrameDate,
				enums.TimeFrameTime,
				enums.TimeFrameWeek,
				enums.TimeFrameMonth,
			},
		},
		{
			Name: "shipping_updated",
			Type: "time",
			Timeframes: []enums.LookerTimeFrame{
				enums.TimeFrameWeek,
				enums.TimeFrameMonth,
			},
		},
	}

	result := resolver.Resolve(dimensions, dimensionGroups, "f_order_items")

	require.Equal(t, 5, len(result))

	// Non-conflicting dimensions unchanged
	assert.Equal(t, "order_id", result[0].Name)
	assert.Nil(t, result[0].Hidden)

	// Conflicting dimensions renamed and hidden
	assert.Equal(t, "item_creation_date_conflict", result[1].Name)
	assert.NotNil(t, result[1].Hidden)

	assert.Equal(t, "item_creation_time_conflict", result[2].Name)
	assert.NotNil(t, result[2].Hidden)

	// Non-conflicting dimension unchanged
	assert.Equal(t, "customer_name", result[3].Name)
	assert.Nil(t, result[3].Hidden)

	// Conflicting dimension renamed and hidden
	assert.Equal(t, "shipping_updated_week_conflict", result[4].Name)
	assert.NotNil(t, result[4].Hidden)
}
