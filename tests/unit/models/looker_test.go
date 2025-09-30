package models

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
)

// TestDbtMetaLookerMeasure_ValidateMeasureAttributes tests measure validation
func TestDbtMetaLookerMeasure_ValidateMeasureAttributes(t *testing.T) {
	tests := []struct {
		name        string
		measure     models.DbtMetaLookerMeasure
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid count_distinct with approximate",
			measure: models.DbtMetaLookerMeasure{
				Type:                 enums.MeasureCountDistinct,
				Approximate:          boolPtr(true),
				ApproximateThreshold: intPtr(1000),
			},
			expectError: false,
		},
		{
			name: "invalid approximate on sum measure",
			measure: models.DbtMetaLookerMeasure{
				Type:        enums.MeasureSum,
				Approximate: boolPtr(true),
			},
			expectError: true,
			errorMsg:    "approximate",
		},
		{
			name: "invalid sql_distinct_key on average measure",
			measure: models.DbtMetaLookerMeasure{
				Type:           enums.MeasureAverage,
				SQLDistinctKey: stringPtr("user_id"),
			},
			expectError: true,
			errorMsg:    "sql_distinct_key",
		},
		{
			name: "valid precision on sum measure",
			measure: models.DbtMetaLookerMeasure{
				Type:      enums.MeasureSum,
				Precision: intPtr(2),
			},
			expectError: false,
		},
		{
			name: "valid precision on average measure",
			measure: models.DbtMetaLookerMeasure{
				Type:      enums.MeasureAverage,
				Precision: intPtr(3),
			},
			expectError: false,
		},
		{
			name: "invalid precision on count measure",
			measure: models.DbtMetaLookerMeasure{
				Type:      enums.MeasureCount,
				Precision: intPtr(2),
			},
			expectError: true,
			errorMsg:    "precision",
		},
		{
			name: "basic count measure",
			measure: models.DbtMetaLookerMeasure{
				Type: enums.MeasureCount,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.measure.ValidateMeasureAttributes()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLookMLDimension_Structure tests dimension structure
func TestLookMLDimension_Structure(t *testing.T) {
	label := "Customer Name"
	description := "The name of the customer"
	hidden := true
	groupLabel := "Customer Info"
	
	dimension := models.LookMLDimension{
		Name:        "customer_name",
		Type:        "string",
		SQL:         "${TABLE}.customer_name",
		Label:       &label,
		Description: &description,
		Hidden:      &hidden,
		GroupLabel:  &groupLabel,
	}

	assert.Equal(t, "customer_name", dimension.Name)
	assert.Equal(t, "string", dimension.Type)
	assert.Equal(t, "${TABLE}.customer_name", dimension.SQL)
	assert.NotNil(t, dimension.Label)
	assert.Equal(t, "Customer Name", *dimension.Label)
	assert.NotNil(t, dimension.Hidden)
	assert.True(t, *dimension.Hidden)
}

// TestLookMLDimensionGroup_Structure tests dimension group structure
func TestLookMLDimensionGroup_Structure(t *testing.T) {
	label := "Created"
	
	dimensionGroup := models.LookMLDimensionGroup{
		Name:  "created_at",
		Type:  "time",
		SQL:   "${TABLE}.created_at",
		Label: &label,
		Timeframes: []enums.LookerTimeFrame{
			enums.TimeFrameRaw,
			enums.TimeFrameDate,
			enums.TimeFrameWeek,
			enums.TimeFrameMonth,
		},
	}

	assert.Equal(t, "created_at", dimensionGroup.Name)
	assert.Equal(t, "time", dimensionGroup.Type)
	assert.Len(t, dimensionGroup.Timeframes, 4)
	assert.Contains(t, dimensionGroup.Timeframes, enums.TimeFrameDate)
}

// TestLookMLMeasure_Structure tests measure structure
func TestLookMLMeasure_Structure(t *testing.T) {
	sql := "${TABLE}.amount"
	label := "Total Amount"
	precision := 2
	
	measure := models.LookMLMeasure{
		Name:      "total_amount",
		Type:      enums.MeasureSum,
		SQL:       &sql,
		Label:     &label,
		Precision: &precision,
	}

	assert.Equal(t, "total_amount", measure.Name)
	assert.Equal(t, enums.MeasureSum, measure.Type)
	assert.NotNil(t, measure.SQL)
	assert.Equal(t, "${TABLE}.amount", *measure.SQL)
	assert.NotNil(t, measure.Precision)
	assert.Equal(t, 2, *measure.Precision)
}

// TestLookMLView_Structure tests view structure
func TestLookMLView_Structure(t *testing.T) {
	label := "Customer View"
	
	view := models.LookMLView{
		Name:         "customers",
		SQLTableName: "`project.dataset.customers`",
		Label:        &label,
		Dimensions: []models.LookMLDimension{
			{Name: "id", Type: "number", SQL: "${TABLE}.id"},
			{Name: "name", Type: "string", SQL: "${TABLE}.name"},
		},
		Measures: []models.LookMLMeasure{
			{Name: "count", Type: enums.MeasureCount},
		},
	}

	assert.Equal(t, "customers", view.Name)
	assert.Equal(t, "`project.dataset.customers`", view.SQLTableName)
	assert.Len(t, view.Dimensions, 2)
	assert.Len(t, view.Measures, 1)
	assert.NotNil(t, view.Label)
	assert.Equal(t, "Customer View", *view.Label)
}

// TestLookMLExplore_Structure tests explore structure
func TestLookMLExplore_Structure(t *testing.T) {
	label := "Customer Analysis"
	hidden := true
	
	explore := models.LookMLExplore{
		Name:        "customers",
		ViewName:    "customers",
		Label:       &label,
		Hidden:      &hidden,
		Joins: []models.LookMLJoin{
			{
				Name: "orders",
				Type: &[]enums.LookerJoinType{enums.JoinLeftOuter}[0],
			},
		},
	}

	assert.Equal(t, "customers", explore.Name)
	assert.Equal(t, "customers", explore.ViewName)
	assert.Len(t, explore.Joins, 1)
	assert.NotNil(t, explore.Hidden)
	assert.True(t, *explore.Hidden)
}

// TestLookMLJoin_Structure tests join structure
func TestLookMLJoin_Structure(t *testing.T) {
	sql := "${customers.id} = ${orders.customer_id}"
	viewLabel := "Order Information"
	joinType := enums.JoinLeftOuter
	relationship := enums.RelationshipManyToOne
	
	join := models.LookMLJoin{
		Name:         "orders",
		ViewLabel:    &viewLabel,
		SQL:          &sql,
		Type:         &joinType,
		Relationship: &relationship,
	}

	assert.Equal(t, "orders", join.Name)
	assert.NotNil(t, join.SQL)
	assert.NotNil(t, join.Type)
	assert.Equal(t, enums.JoinLeftOuter, *join.Type)
	assert.NotNil(t, join.Relationship)
	assert.Equal(t, enums.RelationshipManyToOne, *join.Relationship)
}

// TestDbtMetaLooker_Structure tests meta looker structure
func TestDbtMetaLooker_Structure(t *testing.T) {
	viewLabel := "Test View"
	dimLabel := "Test Dimension"
	
	meta := models.DbtMetaLooker{
		View: &models.DbtMetaLookerBase{
			Label: &viewLabel,
		},
		Dimension: &models.DbtMetaLookerDimension{
			DbtMetaLookerBase: models.DbtMetaLookerBase{
				Label: &dimLabel,
			},
		},
		Measures: []models.DbtMetaLookerMeasure{
			{Type: enums.MeasureCount},
			{Type: enums.MeasureSum},
		},
	}

	assert.NotNil(t, meta.View)
	assert.NotNil(t, meta.Dimension)
	assert.Len(t, meta.Measures, 2)
	assert.Equal(t, "Test View", *meta.View.Label)
	assert.Equal(t, "Test Dimension", *meta.Dimension.Label)
}

// TestDbtMetaLookerMeasureFilter_Structure tests measure filter structure
func TestDbtMetaLookerMeasureFilter_Structure(t *testing.T) {
	filter := models.DbtMetaLookerMeasureFilter{
		FilterDimension:  "status",
		FilterExpression: "active",
	}

	assert.Equal(t, "status", filter.FilterDimension)
	assert.Equal(t, "active", filter.FilterExpression)
}

// TestDbtMetaLookerDimension_WithTimeframes tests dimension with timeframes
func TestDbtMetaLookerDimension_WithTimeframes(t *testing.T) {
	dimension := models.DbtMetaLookerDimension{
		Timeframes: []enums.LookerTimeFrame{
			enums.TimeFrameDate,
			enums.TimeFrameWeek,
			enums.TimeFrameMonth,
			enums.TimeFrameYear,
		},
	}

	assert.Len(t, dimension.Timeframes, 4)
	assert.Contains(t, dimension.Timeframes, enums.TimeFrameDate)
	assert.Contains(t, dimension.Timeframes, enums.TimeFrameYear)
}

// TestLookViewFile_Structure tests view file structure
func TestLookViewFile_Structure(t *testing.T) {
	file := models.LookViewFile{
		Filename: "customers.view.lkml",
		Contents: "view: customers { }",
		Schema:   "public",
	}

	assert.Equal(t, "customers.view.lkml", file.Filename)
	assert.Equal(t, "view: customers { }", file.Contents)
	assert.Equal(t, "public", file.Schema)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
