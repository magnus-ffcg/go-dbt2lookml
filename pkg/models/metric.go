package models

// DbtMetric represents a metric definition from dbt's Semantic Layer
// Metrics are composed of measures and can be simple, ratio, derived, cumulative, or conversion
type DbtMetric struct {
	Name         string              `json:"name" yaml:"name"`
	ResourceType string              `json:"resource_type" yaml:"resource_type"`
	PackageName  string              `json:"package_name" yaml:"package_name"`
	UniqueID     string              `json:"unique_id" yaml:"unique_id"`
	Description  *string             `json:"description,omitempty" yaml:"description,omitempty"`
	Label        *string             `json:"label,omitempty" yaml:"label,omitempty"`
	Type         string              `json:"type" yaml:"type"` // simple, ratio, derived, cumulative, conversion
	TypeParams   DbtMetricTypeParams `json:"type_params" yaml:"type_params"`
	Filter       *DbtMetricFilter    `json:"filter,omitempty" yaml:"filter,omitempty"`
	Meta         map[string]any      `json:"meta,omitempty" yaml:"meta,omitempty"`
}

// DbtMetricFilter represents filter conditions for a metric
type DbtMetricFilter struct {
	WhereFilters []DbtMetricWhereFilter `json:"where_filters" yaml:"where_filters"`
}

// DbtMetricWhereFilter represents a single WHERE filter condition
type DbtMetricWhereFilter struct {
	WhereSQLTemplate string `json:"where_sql_template" yaml:"where_sql_template"`
}

// DbtMetricTypeParams contains type-specific parameters for metrics
type DbtMetricTypeParams struct {
	// Common
	Measure       *DbtMetricInputMeasure  `json:"measure,omitempty" yaml:"measure,omitempty"`               // For simple metrics
	InputMeasures []DbtMetricInputMeasure `json:"input_measures,omitempty" yaml:"input_measures,omitempty"` // For ratio metrics

	// Ratio-specific
	Numerator   *DbtMetricInput `json:"numerator,omitempty" yaml:"numerator,omitempty"`
	Denominator *DbtMetricInput `json:"denominator,omitempty" yaml:"denominator,omitempty"`

	// Derived-specific
	Expr    *string        `json:"expr,omitempty" yaml:"expr,omitempty"`       // Expression for derived metrics
	Metrics []DbtMetricRef `json:"metrics,omitempty" yaml:"metrics,omitempty"` // Referenced metrics

	// Cumulative-specific
	Window               *string                  `json:"window,omitempty" yaml:"window,omitempty"`
	GrainToDate          *string                  `json:"grain_to_date,omitempty" yaml:"grain_to_date,omitempty"`
	CumulativeTypeParams *DbtCumulativeTypeParams `json:"cumulative_type_params,omitempty" yaml:"cumulative_type_params,omitempty"`

	// Conversion-specific
	ConversionTypeParams *DbtConversionTypeParams `json:"conversion_type_params,omitempty" yaml:"conversion_type_params,omitempty"`
}

// DbtMetricInput represents a metric reference (for numerator/denominator in ratio metrics)
type DbtMetricInput struct {
	Name          string  `json:"name" yaml:"name"`
	Filter        *string `json:"filter,omitempty" yaml:"filter,omitempty"`
	Alias         *string `json:"alias,omitempty" yaml:"alias,omitempty"`
	OffsetWindow  *string `json:"offset_window,omitempty" yaml:"offset_window,omitempty"`
	OffsetToGrain *string `json:"offset_to_grain,omitempty" yaml:"offset_to_grain,omitempty"`
}

// DbtMetricInputMeasure represents a measure input for metrics
type DbtMetricInputMeasure struct {
	Name            string  `json:"name" yaml:"name"`
	Filter          *string `json:"filter,omitempty" yaml:"filter,omitempty"`
	Alias           *string `json:"alias,omitempty" yaml:"alias,omitempty"`
	JoinToTimespine bool    `json:"join_to_timespine" yaml:"join_to_timespine"`
	FillNullsWith   *string `json:"fill_nulls_with,omitempty" yaml:"fill_nulls_with,omitempty"`
}

// DbtMetricRef represents a reference to another metric in derived metrics
type DbtMetricRef struct {
	Name   string  `json:"name" yaml:"name"`
	Alias  *string `json:"alias,omitempty" yaml:"alias,omitempty"`
	Filter *string `json:"filter,omitempty" yaml:"filter,omitempty"`
}

// DbtCumulativeTypeParams contains parameters for cumulative metrics
type DbtCumulativeTypeParams struct {
	Window      *DbtMetricWindow `json:"window,omitempty" yaml:"window,omitempty"`
	GrainToDate *string          `json:"grain_to_date,omitempty" yaml:"grain_to_date,omitempty"`
	PeriodAgg   *string          `json:"period_agg,omitempty" yaml:"period_agg,omitempty"`
}

// DbtMetricWindow represents a time window for cumulative metrics
type DbtMetricWindow struct {
	Count       int    `json:"count" yaml:"count"`
	Granularity string `json:"granularity" yaml:"granularity"`
}

// DbtConversionTypeParams contains parameters for conversion metrics
type DbtConversionTypeParams struct {
	Entity             string                `json:"entity" yaml:"entity"`
	Calculation        *string               `json:"calculation,omitempty" yaml:"calculation,omitempty"`
	BaseMeasure        DbtMetricInputMeasure `json:"base_measure" yaml:"base_measure"`
	ConversionMeasure  DbtMetricInputMeasure `json:"conversion_measure" yaml:"conversion_measure"`
	Window             *DbtMetricWindow      `json:"window,omitempty" yaml:"window,omitempty"`
	ConstantProperties []DbtConstantProperty `json:"constant_properties,omitempty" yaml:"constant_properties,omitempty"`
}

// DbtConstantProperty represents properties that must match for conversion
type DbtConstantProperty struct {
	BaseProperty       string `json:"base_property" yaml:"base_property"`
	ConversionProperty string `json:"conversion_property" yaml:"conversion_property"`
}

// IsSimple returns true if this is a simple metric
func (m *DbtMetric) IsSimple() bool {
	return m.Type == "simple"
}

// IsRatio returns true if this is a ratio metric
func (m *DbtMetric) IsRatio() bool {
	return m.Type == "ratio"
}

// IsDerived returns true if this is a derived metric
func (m *DbtMetric) IsDerived() bool {
	return m.Type == "derived"
}

// IsCumulative returns true if this is a cumulative metric
func (m *DbtMetric) IsCumulative() bool {
	return m.Type == "cumulative"
}

// IsConversion returns true if this is a conversion metric
func (m *DbtMetric) IsConversion() bool {
	return m.Type == "conversion"
}

// GetDisplayName returns the label if set, otherwise the name
func (m *DbtMetric) GetDisplayName() string {
	if m.Label != nil && *m.Label != "" {
		return *m.Label
	}
	return m.Name
}

// HasFilter returns true if this metric has filter conditions
func (m *DbtMetric) HasFilter() bool {
	return m.Filter != nil && len(m.Filter.WhereFilters) > 0
}

// GetFilterSQL returns the combined filter SQL from all where filters
func (m *DbtMetric) GetFilterSQL() string {
	if !m.HasFilter() {
		return ""
	}

	// For now, just return the first filter's SQL
	// Multiple filters would need to be AND'd together
	if len(m.Filter.WhereFilters) > 0 {
		return m.Filter.WhereFilters[0].WhereSQLTemplate
	}

	return ""
}
