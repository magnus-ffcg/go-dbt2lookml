package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDbtModelColumn_IsArrayColumn(t *testing.T) {
	tests := []struct {
		name     string
		dataType *string
		expected bool
	}{
		{"nil dataType", nil, false},
		{"ARRAY<STRING>", stringPtr("ARRAY<STRING>"), true},
		{"ARRAY<INT64>", stringPtr("ARRAY<INT64>"), true},
		{"ARRAY<STRUCT<...>>", stringPtr("ARRAY<STRUCT<field STRING>>"), true},
		{"array lowercase", stringPtr("array<string>"), true},
		{"STRING", stringPtr("STRING"), false},
		{"INT64", stringPtr("INT64"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &DbtModelColumn{DataType: tt.dataType}
			assert.Equal(t, tt.expected, column.IsArrayColumn())
		})
	}
}

func TestDbtModelColumn_IsStructColumn(t *testing.T) {
	tests := []struct {
		name     string
		dataType *string
		expected bool
	}{
		{"nil dataType", nil, false},
		{"STRUCT<field STRING>", stringPtr("STRUCT<field STRING>"), true},
		{"struct lowercase", stringPtr("struct<x INT64>"), true},
		{"ARRAY<STRUCT>", stringPtr("ARRAY<STRUCT<x INT64>>"), false}, // starts with ARRAY, not STRUCT
		{"STRING", stringPtr("STRING"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &DbtModelColumn{DataType: tt.dataType}
			assert.Equal(t, tt.expected, column.IsStructColumn())
		})
	}
}

func TestDbtModelColumn_IsDateTimeColumn(t *testing.T) {
	tests := []struct {
		name     string
		dataType *string
		expected bool
	}{
		{"nil dataType", nil, false},
		{"DATE", stringPtr("DATE"), true},
		{"DATETIME", stringPtr("DATETIME"), true},
		{"TIMESTAMP", stringPtr("TIMESTAMP"), true},
		{"date lowercase", stringPtr("date"), true},
		{"STRING", stringPtr("STRING"), false},
		{"INT64", stringPtr("INT64"), false},
		{"TIME", stringPtr("TIME"), false}, // Not a date/time dimension group type
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &DbtModelColumn{DataType: tt.dataType}
			assert.Equal(t, tt.expected, column.IsDateTimeColumn())
		})
	}
}

func TestDbtModelColumn_IsSimpleArrayColumn(t *testing.T) {
	tests := []struct {
		name     string
		dataType *string
		expected bool
	}{
		{"nil dataType", nil, false},
		{"ARRAY<STRING>", stringPtr("ARRAY<STRING>"), true},
		{"ARRAY<INT64>", stringPtr("ARRAY<INT64>"), true},
		{"ARRAY<STRUCT<...>>", stringPtr("ARRAY<STRUCT<field STRING>>"), false}, // contains STRUCT
		{"array lowercase", stringPtr("array<string>"), true},
		{"STRING", stringPtr("STRING"), false},
		{"STRUCT<...>", stringPtr("STRUCT<field STRING>"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &DbtModelColumn{DataType: tt.dataType}
			assert.Equal(t, tt.expected, column.IsSimpleArrayColumn())
		})
	}
}

func TestDbtModelColumn_GetDataTypeUpper(t *testing.T) {
	tests := []struct {
		name     string
		dataType *string
		expected string
	}{
		{"nil dataType", nil, ""},
		{"lowercase", stringPtr("string"), "STRING"},
		{"uppercase", stringPtr("INT64"), "INT64"},
		{"mixed case", stringPtr("Array<String>"), "ARRAY<STRING>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := &DbtModelColumn{DataType: tt.dataType}
			assert.Equal(t, tt.expected, column.GetDataTypeUpper())
		})
	}
}
