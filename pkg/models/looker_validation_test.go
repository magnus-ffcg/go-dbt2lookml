package models

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/stretchr/testify/assert"
)

func TestLookMLDimension_Validate(t *testing.T) {
	tests := []struct {
		name        string
		dimension   LookMLDimension
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid dimension",
			dimension: LookMLDimension{
				Name: "test_dim",
				Type: "string",
				SQL:  "${TABLE}.test",
			},
			expectError: false,
		},
		{
			name: "missing name",
			dimension: LookMLDimension{
				Type: "string",
				SQL:  "${TABLE}.test",
			},
			expectError: true,
			errorMsg:    "dimension name is required",
		},
		{
			name: "missing type",
			dimension: LookMLDimension{
				Name: "test_dim",
				SQL:  "${TABLE}.test",
			},
			expectError: true,
			errorMsg:    "dimension type is required",
		},
		{
			name: "missing SQL",
			dimension: LookMLDimension{
				Name: "test_dim",
				Type: "string",
			},
			expectError: true,
			errorMsg:    "dimension SQL is required",
		},
		{
			name: "invalid type",
			dimension: LookMLDimension{
				Name: "test_dim",
				Type: "invalid_type",
				SQL:  "${TABLE}.test",
			},
			expectError: true,
			errorMsg:    "invalid dimension type",
		},
		{
			name: "valid number type",
			dimension: LookMLDimension{
				Name: "count",
				Type: "number",
				SQL:  "${TABLE}.count",
			},
			expectError: false,
		},
		{
			name: "valid yesno type",
			dimension: LookMLDimension{
				Name: "is_active",
				Type: "yesno",
				SQL:  "${TABLE}.active",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.dimension.Validate()
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

func TestLookMLMeasure_Validate(t *testing.T) {
	sql := "${TABLE}.amount"

	tests := []struct {
		name        string
		measure     LookMLMeasure
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid count measure",
			measure: LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
			expectError: false,
		},
		{
			name: "valid sum measure",
			measure: LookMLMeasure{
				Name: "total",
				Type: enums.MeasureSum,
				SQL:  &sql,
			},
			expectError: false,
		},
		{
			name: "missing name",
			measure: LookMLMeasure{
				Type: enums.MeasureCount,
			},
			expectError: true,
			errorMsg:    "measure name is required",
		},
		{
			name: "missing SQL for sum",
			measure: LookMLMeasure{
				Name: "total",
				Type: enums.MeasureSum,
			},
			expectError: true,
			errorMsg:    "measure SQL is required",
		},
		{
			name: "invalid approximate on sum",
			measure: LookMLMeasure{
				Name:        "total",
				Type:        enums.MeasureSum,
				SQL:         &sql,
				Approximate: boolPtr(true),
			},
			expectError: true,
			errorMsg:    "approximate, approximate_threshold, and sql_distinct_key can only be used with count_distinct",
		},
		{
			name: "valid count_distinct with approximate",
			measure: LookMLMeasure{
				Name:        "unique_users",
				Type:        enums.MeasureCountDistinct,
				SQL:         &sql,
				Approximate: boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "invalid precision on count",
			measure: LookMLMeasure{
				Name:      "count",
				Type:      enums.MeasureCount,
				Precision: intPtr(2),
			},
			expectError: true,
			errorMsg:    "precision can only be used with average or sum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.measure.Validate()
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

func TestLookMLView_Validate(t *testing.T) {
	tests := []struct {
		name        string
		view        LookMLView
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid view",
			view: LookMLView{
				Name:         "test_view",
				SQLTableName: "schema.table",
				Dimensions: []LookMLDimension{
					{
						Name: "id",
						Type: "string",
						SQL:  "${TABLE}.id",
					},
				},
			},
			expectError: false,
		},
		{
			name: "missing name",
			view: LookMLView{
				SQLTableName: "schema.table",
			},
			expectError: true,
			errorMsg:    "view name is required",
		},
		{
			name: "missing sql_table_name",
			view: LookMLView{
				Name: "test_view",
			},
			expectError: true,
			errorMsg:    "sql_table_name is required",
		},
		{
			name: "invalid dimension",
			view: LookMLView{
				Name:         "test_view",
				SQLTableName: "schema.table",
				Dimensions: []LookMLDimension{
					{
						Name: "bad_dim",
						// Missing Type and SQL
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid dimension",
		},
		{
			name: "invalid measure",
			view: LookMLView{
				Name:         "test_view",
				SQLTableName: "schema.table",
				Measures: []LookMLMeasure{
					{
						// Missing Name and Type
					},
				},
			},
			expectError: true,
			errorMsg:    "invalid measure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.view.Validate()
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
