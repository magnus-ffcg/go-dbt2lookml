package enums

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetLookerType tests BigQuery to Looker type mapping
func TestGetLookerType(t *testing.T) {
	tests := []struct {
		name         string
		bigQueryType string
		expectedType LookerBigQueryDataType
	}{
		// Numeric types
		{"INT64", "INT64", DataTypeNumber},
		{"INTEGER", "INTEGER", DataTypeNumber},
		{"NUMERIC", "NUMERIC", DataTypeNumber},
		{"DECIMAL", "DECIMAL", DataTypeNumber},
		{"BIGNUMERIC", "BIGNUMERIC", DataTypeNumber},
		{"FLOAT64", "FLOAT64", DataTypeNumber},
		{"FLOAT", "FLOAT", DataTypeNumber},

		// Boolean types
		{"BOOLEAN", "BOOLEAN", DataTypeYesNo},
		{"BOOL", "BOOL", DataTypeYesNo},

		// String types
		{"STRING", "STRING", DataTypeString},
		{"BYTES", "BYTES", DataTypeString},

		// Date/Time types (specific types for dimension groups)
		{"DATE", "DATE", DataTypeDate},
		{"DATETIME", "DATETIME", DataTypeDateTime},
		{"TIMESTAMP", "TIMESTAMP", DataTypeTimestamp},
		{"TIME", "TIME", DataTypeString},

		// Complex types
		{"ARRAY<STRING>", "ARRAY<STRING>", DataTypeString},
		{"STRUCT<field STRING>", "STRUCT<field STRING>", DataTypeString},

		// Unknown/fallback
		{"UNKNOWN_TYPE", "UNKNOWN_TYPE", DataTypeString},
		{"", "", DataTypeString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLookerType(tt.bigQueryType)
			assert.Equal(t, tt.expectedType, result,
				"BigQuery type %s should map to Looker type %s", tt.bigQueryType, tt.expectedType)
		})
	}
}

// TestGetLookerType_CaseSensitive tests that type mapping is case-sensitive
func TestGetLookerType_CaseSensitive(t *testing.T) {
	// GetLookerType is case-sensitive - only uppercase types match
	// Lowercase types should fall back to string
	assert.Equal(t, DataTypeString, GetLookerType("int64"))
	assert.Equal(t, DataTypeString, GetLookerType("boolean"))
	assert.Equal(t, DataTypeString, GetLookerType("string"))

	// Test mixed case also falls back to string
	assert.Equal(t, DataTypeString, GetLookerType("Int64"))
	assert.Equal(t, DataTypeString, GetLookerType("Boolean"))
	assert.Equal(t, DataTypeString, GetLookerType("String"))
}

// TestAdapterType tests adapter type enum
func TestAdapterType(t *testing.T) {
	assert.Equal(t, "bigquery", string(BigQuery))
}

// TestMeasureType tests measure type enums
func TestMeasureType(t *testing.T) {
	assert.Equal(t, "count", string(MeasureCount))
	assert.Equal(t, "count_distinct", string(MeasureCountDistinct))
	assert.Equal(t, "sum", string(MeasureSum))
	assert.Equal(t, "average", string(MeasureAverage))
	assert.Equal(t, "min", string(MeasureMin))
	assert.Equal(t, "max", string(MeasureMax))
}

// TestTimeFrame tests timeframe enums
func TestTimeFrame(t *testing.T) {
	timeframes := []LookerTimeFrame{
		TimeFrameRaw,
		TimeFrameTime,
		TimeFrameDate,
		TimeFrameWeek,
		TimeFrameMonth,
		TimeFrameQuarter,
		TimeFrameYear,
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
	assert.Equal(t, "decimal_0", string(FormatDecimal0))
	assert.Equal(t, "decimal_1", string(FormatDecimal1))
	assert.Equal(t, "decimal_2", string(FormatDecimal2))
	assert.Equal(t, "usd", string(FormatUSD))
	assert.Equal(t, "eur", string(FormatEUR))
	assert.Equal(t, "gbp", string(FormatGBP))
	assert.Equal(t, "percent_0", string(FormatPercent0))
	assert.Equal(t, "percent_1", string(FormatPercent1))
	assert.Equal(t, "percent_2", string(FormatPercent2))
}
