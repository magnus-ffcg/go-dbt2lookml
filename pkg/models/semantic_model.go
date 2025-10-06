package models

// DbtSemanticModel represents a dbt semantic model definition
// Semantic models are the foundation for dbt's Semantic Layer
type DbtSemanticModel struct {
	Name          string                    `json:"name" yaml:"name"`
	Description   *string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Model         string                    `json:"model" yaml:"model"` // ref() reference like "ref('model_name')"
	Label         *string                   `json:"label,omitempty" yaml:"label,omitempty"`
	Entities      []DbtSemanticEntity       `json:"entities" yaml:"entities"`
	Dimensions    []DbtSemanticDimension    `json:"dimensions" yaml:"dimensions"`
	Measures      []DbtSemanticMeasure      `json:"measures" yaml:"measures"`
	Defaults      *DbtSemanticModelDefaults `json:"defaults,omitempty" yaml:"defaults,omitempty"`
	PrimaryEntity *string                   `json:"primary_entity,omitempty" yaml:"primary_entity,omitempty"`
	NodeRelation  *DbtSemanticNodeRelation  `json:"node_relation,omitempty" yaml:"node_relation,omitempty"`
	UniqueID      *string                   `json:"unique_id,omitempty" yaml:"unique_id,omitempty"`
}

// DbtSemanticModelDefaults represents default settings for a semantic model
type DbtSemanticModelDefaults struct {
	AggTimeDimension string `json:"agg_time_dimension" yaml:"agg_time_dimension"`
}

// DbtSemanticEntity represents an entity in a semantic model
// Entities are join keys between models
type DbtSemanticEntity struct {
	Name string  `json:"name" yaml:"name"`
	Type string  `json:"type" yaml:"type"` // "primary", "foreign", "unique", "natural"
	Expr *string `json:"expr,omitempty" yaml:"expr,omitempty"`
	Role *string `json:"role,omitempty" yaml:"role,omitempty"`
}

// DbtSemanticDimension represents a dimension in a semantic model
// Dimensions are categorical or time-based grouping fields
type DbtSemanticDimension struct {
	Name        string                          `json:"name" yaml:"name"`
	Type        string                          `json:"type" yaml:"type"` // "categorical", "time"
	Expr        *string                         `json:"expr,omitempty" yaml:"expr,omitempty"`
	Label       *string                         `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string                         `json:"description,omitempty" yaml:"description,omitempty"`
	TypeParams  *DbtSemanticDimensionTypeParams `json:"type_params,omitempty" yaml:"type_params,omitempty"`
}

// DbtSemanticDimensionTypeParams represents type-specific parameters for dimensions
type DbtSemanticDimensionTypeParams struct {
	TimeGranularity string `json:"time_granularity,omitempty" yaml:"time_granularity,omitempty"` // "day", "week", "month", etc.
}

// DbtSemanticMeasure represents a measure in a semantic model
// Measures are aggregations that can be used in metrics
type DbtSemanticMeasure struct {
	Name                 string                           `json:"name" yaml:"name"`
	Description          *string                          `json:"description,omitempty" yaml:"description,omitempty"`
	Agg                  string                           `json:"agg" yaml:"agg"` // sum, average, min, max, median, count_distinct, percentile, sum_boolean
	Expr                 *string                          `json:"expr,omitempty" yaml:"expr,omitempty"`
	Label                *string                          `json:"label,omitempty" yaml:"label,omitempty"`
	CreateMetric         *bool                            `json:"create_metric,omitempty" yaml:"create_metric,omitempty"`
	NonAdditiveDimension *DbtSemanticNonAdditiveDimension `json:"non_additive_dimension,omitempty" yaml:"non_additive_dimension,omitempty"`
	AggParams            *DbtSemanticMeasureAggParams     `json:"agg_params,omitempty" yaml:"agg_params,omitempty"`
	AggTimeDimension     *string                          `json:"agg_time_dimension,omitempty" yaml:"agg_time_dimension,omitempty"`
}

// DbtSemanticNonAdditiveDimension represents a non-additive dimension configuration
// Used for semi-additive measures (like account balances)
type DbtSemanticNonAdditiveDimension struct {
	Name            string   `json:"name" yaml:"name"`
	WindowChoice    *string  `json:"window_choice,omitempty" yaml:"window_choice,omitempty"` // "min", "max"
	WindowGroupings []string `json:"window_groupings,omitempty" yaml:"window_groupings,omitempty"`
}

// DbtSemanticMeasureAggParams represents aggregation parameters for measures
// Used for measures that need additional configuration (like percentile)
type DbtSemanticMeasureAggParams struct {
	Percentile            *float64 `json:"percentile,omitempty" yaml:"percentile,omitempty"`
	UseDiscretePercentile *bool    `json:"use_discrete_percentile,omitempty" yaml:"use_discrete_percentile,omitempty"`
}

// DbtSemanticNodeRelation represents the relation information for a semantic model
type DbtSemanticNodeRelation struct {
	Alias        string  `json:"alias" yaml:"alias"`
	SchemaName   string  `json:"schema_name" yaml:"schema_name"`
	Database     *string `json:"database,omitempty" yaml:"database,omitempty"`
	RelationName string  `json:"relation_name" yaml:"relation_name"`
}

// GetModelRef extracts the model name from a ref() expression
// Example: "ref('customers')" -> "customers"
func (s *DbtSemanticModel) GetModelRef() string {
	// Remove "ref('" prefix and "')" suffix
	model := s.Model
	if len(model) > 6 && model[:4] == "ref(" {
		// Extract text between quotes
		start := 5            // After "ref('"
		end := len(model) - 2 // Before "')"
		if end > start {
			return model[start:end]
		}
	}
	return model
}

// GetDisplayName returns the label if set, otherwise the name
func (s *DbtSemanticModel) GetDisplayName() string {
	if s.Label != nil && *s.Label != "" {
		return *s.Label
	}
	return s.Name
}

// GetDisplayName returns the label if set, otherwise the name
func (m *DbtSemanticMeasure) GetDisplayName() string {
	if m.Label != nil && *m.Label != "" {
		return *m.Label
	}
	return m.Name
}

// IsSemiAdditive returns true if the measure has a non-additive dimension
func (m *DbtSemanticMeasure) IsSemiAdditive() bool {
	return m.NonAdditiveDimension != nil
}

// ShouldCreateMetric returns true if create_metric is not explicitly false
func (m *DbtSemanticMeasure) ShouldCreateMetric() bool {
	// Default is true if not specified
	if m.CreateMetric == nil {
		return true
	}
	return *m.CreateMetric
}

// IsPercentile returns true if the measure uses percentile aggregation
func (m *DbtSemanticMeasure) IsPercentile() bool {
	return m.Agg == "percentile"
}

// GetPercentileValue returns the percentile value if this is a percentile measure
func (m *DbtSemanticMeasure) GetPercentileValue() (float64, bool) {
	if !m.IsPercentile() || m.AggParams == nil || m.AggParams.Percentile == nil {
		return 0, false
	}
	return *m.AggParams.Percentile, true
}
