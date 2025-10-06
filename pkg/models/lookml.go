package models

import "github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"

// LookMLDimension represents a dimension in LookML
type LookMLDimension struct {
	Name            string                       `json:"name" yaml:"name"`
	Type            string                       `json:"type" yaml:"type"`
	SQL             string                       `json:"sql" yaml:"sql"`
	Label           *string                      `json:"label,omitempty" yaml:"label,omitempty"`
	Description     *string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden          *bool                        `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	PrimaryKey      *bool                        `json:"primary_key,omitempty" yaml:"primary_key,omitempty"`
	GroupLabel      *string                      `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	GroupItemLabel  *string                      `json:"group_item_label,omitempty" yaml:"group_item_label,omitempty"`
	ValueFormatName *enums.LookerValueFormatName `json:"value_format_name,omitempty" yaml:"value_format_name,omitempty"`
}

// LookMLDimensionGroup represents a dimension group in LookML
type LookMLDimensionGroup struct {
	Name        string                  `json:"name" yaml:"name"`
	Type        string                  `json:"type" yaml:"type"`
	SQL         string                  `json:"sql" yaml:"sql"`
	Timeframes  []enums.LookerTimeFrame `json:"timeframes,omitempty" yaml:"timeframes,omitempty"`
	Datatype    *string                 `json:"datatype,omitempty" yaml:"datatype,omitempty"`
	ConvertTZ   *bool                   `json:"convert_tz,omitempty" yaml:"convert_tz,omitempty"`
	Label       *string                 `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string                 `json:"description,omitempty" yaml:"description,omitempty"`
	GroupLabel  *string                 `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	Hidden      *bool                   `json:"hidden,omitempty" yaml:"hidden,omitempty"`
}

// LookMLMeasure represents a measure in LookML
type LookMLMeasure struct {
	Name                 string                       `json:"name" yaml:"name"`
	Type                 enums.LookerMeasureType      `json:"type" yaml:"type"`
	SQL                  *string                      `json:"sql,omitempty" yaml:"sql,omitempty"`
	Label                *string                      `json:"label,omitempty" yaml:"label,omitempty"`
	Description          *string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden               *bool                        `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	GroupLabel           *string                      `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	ValueFormatName      *enums.LookerValueFormatName `json:"value_format_name,omitempty" yaml:"value_format_name,omitempty"`
	Filters              []LookMLFilter               `json:"filters,omitempty" yaml:"filters,omitempty"`
	Percentile           *int                         `json:"percentile,omitempty" yaml:"percentile,omitempty"`
	Approximate          *bool                        `json:"approximate,omitempty" yaml:"approximate,omitempty"`
	ApproximateThreshold *int                         `json:"approximate_threshold,omitempty" yaml:"approximate_threshold,omitempty"`
	Precision            *int                         `json:"precision,omitempty" yaml:"precision,omitempty"`
	SQLDistinctKey       *string                      `json:"sql_distinct_key,omitempty" yaml:"sql_distinct_key,omitempty"`
}

// LookMLFilter represents a filter in LookML
type LookMLFilter struct {
	Field string `json:"field" yaml:"field"`
	Value string `json:"value" yaml:"value"`
}

// LookMLView represents a view in LookML
type LookMLView struct {
	Name            string                 `json:"name" yaml:"name"`
	SQLTableName    string                 `json:"sql_table_name" yaml:"sql_table_name"`
	Label           *string                `json:"label,omitempty" yaml:"label,omitempty"`
	Description     *string                `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden          *bool                  `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	Dimensions      []LookMLDimension      `json:"dimensions,omitempty" yaml:"dimensions,omitempty"`
	DimensionGroups []LookMLDimensionGroup `json:"dimension_groups,omitempty" yaml:"dimension_groups,omitempty"`
	Measures        []LookMLMeasure        `json:"measures,omitempty" yaml:"measures,omitempty"`
}

// LookMLJoin represents a join in LookML
type LookMLJoin struct {
	Name         string                        `json:"name" yaml:"name"`
	ViewLabel    *string                       `json:"view_label,omitempty" yaml:"view_label,omitempty"`
	SQL          *string                       `json:"sql,omitempty" yaml:"sql,omitempty"`
	Type         *enums.LookerJoinType         `json:"type,omitempty" yaml:"type,omitempty"`
	Relationship *enums.LookerRelationshipType `json:"relationship,omitempty" yaml:"relationship,omitempty"`
}

// LookMLExplore represents an explore in LookML
type LookMLExplore struct {
	Name        string       `json:"name" yaml:"name"`
	ViewName    string       `json:"view_name" yaml:"view_name"`
	Label       *string      `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string      `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden      *bool        `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	Joins       []LookMLJoin `json:"joins,omitempty" yaml:"joins,omitempty"`
}
