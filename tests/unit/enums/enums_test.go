package enums

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
	"github.com/stretchr/testify/assert"
)

// TestGetLookerType tests BigQuery to Looker type mapping
func TestGetLookerType(t *testing.T) {
	tests := []struct {
		name         string
		bigQueryType string
		expectedType enums.LookerBigQueryDataType
	}{
		// Numeric types
		{"INT64", "INT64", enums.DataTypeNumber},
		{"INTEGER", "INTEGER", enums.DataTypeNumber},
		{"NUMERIC", "NUMERIC", enums.DataTypeNumber},
		{"DECIMAL", "DECIMAL", enums.DataTypeNumber},
		{"BIGNUMERIC", "BIGNUMERIC", enums.DataTypeNumber},
		{"FLOAT64", "FLOAT64", enums.DataTypeNumber},
		{"FLOAT", "FLOAT", enums.DataTypeNumber},
		
		// Boolean types
		{"BOOLEAN", "BOOLEAN", enums.DataTypeYesNo},
		{"BOOL", "BOOL", enums.DataTypeYesNo},
		
		// String types
		{"STRING", "STRING", enums.DataTypeString},
		{"BYTES", "BYTES", enums.DataTypeString},
		
		// Date/Time types (specific types for dimension groups)
		{"DATE", "DATE", enums.DataTypeDate},
		{"DATETIME", "DATETIME", enums.DataTypeDateTime},
		{"TIMESTAMP", "TIMESTAMP", enums.DataTypeTimestamp},
		{"TIME", "TIME", enums.DataTypeString},
		
		// Complex types
		{"ARRAY<STRING>", "ARRAY<STRING>", enums.DataTypeString},
		{"STRUCT<field STRING>", "STRUCT<field STRING>", enums.DataTypeString},
		
		// Unknown/fallback
		{"UNKNOWN_TYPE", "UNKNOWN_TYPE", enums.DataTypeString},
		{"", "", enums.DataTypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := enums.GetLookerType(tt.bigQueryType)
			assert.Equal(t, tt.expectedType, result, 
				"BigQuery type %s should map to Looker type %s", tt.bigQueryType, tt.expectedType)
		})
	}
}

// TestGetLookerType_CaseSensitive tests that type mapping is case-sensitive
func TestGetLookerType_CaseSensitive(t *testing.T) {
	// GetLookerType is case-sensitive - only uppercase types match
	// Lowercase types should fall back to string
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("int64"))
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("boolean"))
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("string"))
	
	// Test mixed case also falls back to string
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("Int64"))
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("Boolean"))
	assert.Equal(t, enums.DataTypeString, enums.GetLookerType("String"))
}

// TestAdapterType tests adapter type enum
func TestAdapterType(t *testing.T) {
	assert.Equal(t, "bigquery", string(enums.BigQuery))
}

// TestMeasureType tests measure type enums
func TestMeasureType(t *testing.T) {
	assert.Equal(t, "count", string(enums.MeasureCount))
	assert.Equal(t, "count_distinct", string(enums.MeasureCountDistinct))
	assert.Equal(t, "sum", string(enums.MeasureSum))
	assert.Equal(t, "average", string(enums.MeasureAverage))
	assert.Equal(t, "min", string(enums.MeasureMin))
	assert.Equal(t, "max", string(enums.MeasureMax))
}

// TestTimeFrame tests timeframe enums
func TestTimeFrame(t *testing.T) {
	timeframes := []enums.LookerTimeFrame{
		enums.TimeFrameRaw,
		enums.TimeFrameTime,
		enums.TimeFrameDate,
		enums.TimeFrameWeek,
		enums.TimeFrameMonth,
		enums.TimeFrameQuarter,
		enums.TimeFrameYear,
	}

	expectedValues := []string{
		"raw", "time", "date", "week", "month", "quarter", "year",
	}

	for i, tf := range timeframes {
		assert.Equal(t, expectedValues[i], string(tf))
	}
}

// TestLookerValueFormatName tests value format name enums
func TestLookerValueFormatName(t *testing.T) {
	assert.Equal(t, "decimal_0", string(enums.FormatDecimal0))
	assert.Equal(t, "decimal_1", string(enums.FormatDecimal1))
	assert.Equal(t, "decimal_2", string(enums.FormatDecimal2))
	assert.Equal(t, "usd", string(enums.FormatUSD))
	assert.Equal(t, "eur", string(enums.FormatEUR))
	assert.Equal(t, "gbp", string(enums.FormatGBP))
	assert.Equal(t, "percent_0", string(enums.FormatPercent0))
	assert.Equal(t, "percent_1", string(enums.FormatPercent1))
	assert.Equal(t, "percent_2", string(enums.FormatPercent2))
}
