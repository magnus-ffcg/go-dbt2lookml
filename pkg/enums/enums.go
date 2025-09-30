package enums

// SupportedDbtAdapters represents the supported dbt adapters
type SupportedDbtAdapters string

const (
	// BigQuery is the only supported adapter
	BigQuery SupportedDbtAdapters = "bigquery"
)

// LookerMeasureType represents Looker measure types
type LookerMeasureType string

const (
	MeasureNumber          LookerMeasureType = "number"
	MeasureString          LookerMeasureType = "string"
	MeasureAverage         LookerMeasureType = "average"
	MeasureAverageDistinct LookerMeasureType = "average_distinct"
	MeasureCount           LookerMeasureType = "count"
	MeasureCountDistinct   LookerMeasureType = "count_distinct"
	MeasureList            LookerMeasureType = "list"
	MeasureMax             LookerMeasureType = "max"
	MeasureMedian          LookerMeasureType = "median"
	MeasureMedianDistinct  LookerMeasureType = "median_distinct"
	MeasureMin             LookerMeasureType = "min"
	MeasureSum             LookerMeasureType = "sum"
	MeasureSumDistinct     LookerMeasureType = "sum_distinct"
)

// LookerValueFormatName represents Looker value format names
type LookerValueFormatName string

const (
	FormatDecimal0 LookerValueFormatName = "decimal_0"
	FormatDecimal1 LookerValueFormatName = "decimal_1"
	FormatDecimal2 LookerValueFormatName = "decimal_2"
	FormatDecimal3 LookerValueFormatName = "decimal_3"
	FormatDecimal4 LookerValueFormatName = "decimal_4"
	FormatUSD0     LookerValueFormatName = "usd_0"
	FormatUSD      LookerValueFormatName = "usd"
	FormatGBP0     LookerValueFormatName = "gbp_0"
	FormatGBP      LookerValueFormatName = "gbp"
	FormatEUR0     LookerValueFormatName = "eur_0"
	FormatEUR      LookerValueFormatName = "eur"
	FormatID       LookerValueFormatName = "id"
	FormatPercent0 LookerValueFormatName = "percent_0"
	FormatPercent1 LookerValueFormatName = "percent_1"
	FormatPercent2 LookerValueFormatName = "percent_2"
	FormatPercent3 LookerValueFormatName = "percent_3"
	FormatPercent4 LookerValueFormatName = "percent_4"
)

// LookerBigQueryDataType maps BigQuery data types to Looker types
type LookerBigQueryDataType string

const (
	DataTypeNumber    LookerBigQueryDataType = "number"
	DataTypeYesNo     LookerBigQueryDataType = "yesno"
	DataTypeString    LookerBigQueryDataType = "string"
	DataTypeTimestamp LookerBigQueryDataType = "timestamp"
	DataTypeDateTime  LookerBigQueryDataType = "datetime"
	DataTypeDate      LookerBigQueryDataType = "date"
)

// GetLookerType returns the appropriate Looker type for a BigQuery data type
func GetLookerType(bqType string) LookerBigQueryDataType {
	switch bqType {
	case "INT64", "INTEGER", "FLOAT", "FLOAT64", "NUMERIC", "DECIMAL", "BIGNUMERIC":
		return DataTypeNumber
	case "BOOLEAN", "BOOL":
		return DataTypeYesNo
	case "TIMESTAMP":
		return DataTypeTimestamp
	case "DATETIME":
		return DataTypeDateTime
	case "DATE":
		return DataTypeDate
	default:
		return DataTypeString
	}
}

// LookerTimeFrame represents Looker time frame options
type LookerTimeFrame string

const (
	TimeFrameRaw     LookerTimeFrame = "raw"
	TimeFrameDate    LookerTimeFrame = "date"
	TimeFrameWeek    LookerTimeFrame = "week"
	TimeFrameMonth   LookerTimeFrame = "month"
	TimeFrameQuarter LookerTimeFrame = "quarter"
	TimeFrameYear    LookerTimeFrame = "year"
	TimeFrameTime    LookerTimeFrame = "time"
)

// LookerRelationshipType represents relationship types in Looker
type LookerRelationshipType string

const (
	RelationshipManyToOne  LookerRelationshipType = "many_to_one"
	RelationshipManyToMany LookerRelationshipType = "many_to_many"
	RelationshipOneToOne   LookerRelationshipType = "one_to_one"
	RelationshipOneToMany  LookerRelationshipType = "one_to_many"
)

// LookerJoinType represents join types in Looker
type LookerJoinType string

const (
	JoinLeftOuter LookerJoinType = "left_outer"
	JoinFullOuter LookerJoinType = "full_outer"
	JoinInner     LookerJoinType = "inner"
	JoinCross     LookerJoinType = "cross"
)

// DbtResourceType represents the type of dbt resource
type DbtResourceType string

const (
	ResourceModel         DbtResourceType = "model"
	ResourceSeed          DbtResourceType = "seed"
	ResourceSnapshot      DbtResourceType = "snapshot"
	ResourceTest          DbtResourceType = "test"
	ResourceAnalysis      DbtResourceType = "analysis"
	ResourceOperation     DbtResourceType = "operation"
	ResourceExposure      DbtResourceType = "exposure"
	ResourceMacro         DbtResourceType = "macro"
	ResourceRPC           DbtResourceType = "rpc"
	ResourceSQLOperation  DbtResourceType = "sql_operation"
	ResourceSource        DbtResourceType = "source"
	ResourceDoc           DbtResourceType = "doc"
	ResourceGroup         DbtResourceType = "group"
	ResourceMetric        DbtResourceType = "metric"
	ResourceSavedQuery    DbtResourceType = "saved_query"
	ResourceSemanticModel DbtResourceType = "semantic_model"
	ResourceUnitTest      DbtResourceType = "unit_test"
	ResourceFixture       DbtResourceType = "fixture"
)
