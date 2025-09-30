package models

import "strings"

// IsArrayColumn returns true if the column is an ARRAY type
func (c *DbtModelColumn) IsArrayColumn() bool {
	if c.DataType == nil {
		return false
	}
	dataType := strings.ToUpper(*c.DataType)
	return strings.HasPrefix(dataType, "ARRAY")
}

// IsStructColumn returns true if the column is a STRUCT type
func (c *DbtModelColumn) IsStructColumn() bool {
	if c.DataType == nil {
		return false
	}
	dataType := strings.ToUpper(*c.DataType)
	return strings.HasPrefix(dataType, "STRUCT")
}

// IsDateTimeColumn returns true if the column is a date/time type (DATE, DATETIME, TIMESTAMP)
func (c *DbtModelColumn) IsDateTimeColumn() bool {
	if c.DataType == nil {
		return false
	}
	dataType := strings.ToUpper(*c.DataType)
	return dataType == "DATE" || dataType == "DATETIME" || dataType == "TIMESTAMP"
}

// IsSimpleArrayColumn returns true if the column is a simple ARRAY without STRUCT
// (e.g., ARRAY<STRING>, ARRAY<INT64>)
func (c *DbtModelColumn) IsSimpleArrayColumn() bool {
	if c.DataType == nil {
		return false
	}
	dataType := strings.ToUpper(*c.DataType)
	return strings.HasPrefix(dataType, "ARRAY") && !strings.Contains(dataType, "STRUCT")
}

// GetDataTypeUpper returns the uppercase data type string, or empty string if nil
func (c *DbtModelColumn) GetDataTypeUpper() string {
	if c.DataType == nil {
		return ""
	}
	return strings.ToUpper(*c.DataType)
}
