package models

import (
	"fmt"

	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
)

// LookViewFile represents a file in a looker view directory
type LookViewFile struct {
	Filename string `json:"filename" yaml:"filename"`
	Contents string `json:"contents" yaml:"contents"`
	Schema   string `json:"schema" yaml:"schema"`
}

// DbtMetaLookerBase represents the base class for Looker metadata
type DbtMetaLookerBase struct {
	Label       *string `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden      *bool   `json:"hidden,omitempty" yaml:"hidden,omitempty"`
}

// DbtMetaLookerDimension represents Looker-specific metadata for a dimension
type DbtMetaLookerDimension struct {
	DbtMetaLookerBase
	ConvertTZ        *bool                          `json:"convert_tz,omitempty" yaml:"convert_tz,omitempty"`
	GroupLabel       *string                        `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	ValueFormatName  *enums.LookerValueFormatName   `json:"value_format_name,omitempty" yaml:"value_format_name,omitempty"`
	Timeframes       []enums.LookerTimeFrame        `json:"timeframes,omitempty" yaml:"timeframes,omitempty"`
	CanFilter        interface{}                    `json:"can_filter,omitempty" yaml:"can_filter,omitempty"` // Can be bool or string
}

// DbtMetaLookerMeasureFilter represents a filter for Looker measures
type DbtMetaLookerMeasureFilter struct {
	FilterDimension  string `json:"filter_dimension" yaml:"filter_dimension"`
	FilterExpression string `json:"filter_expression" yaml:"filter_expression"`
}

// DbtMetaLookerMeasure represents Looker metadata for a measure
type DbtMetaLookerMeasure struct {
	DbtMetaLookerBase
	// Required fields
	Type enums.LookerMeasureType `json:"type" yaml:"type"`
	
	// Common optional fields
	Name            *string                        `json:"name,omitempty" yaml:"name,omitempty"`
	GroupLabel      *string                        `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	ValueFormatName *enums.LookerValueFormatName   `json:"value_format_name,omitempty" yaml:"value_format_name,omitempty"`
	Filters         []DbtMetaLookerMeasureFilter   `json:"filters,omitempty" yaml:"filters,omitempty"`
	
	// Fields specific to certain measure types
	Approximate          *bool   `json:"approximate,omitempty" yaml:"approximate,omitempty"`                     // For count_distinct
	ApproximateThreshold *int    `json:"approximate_threshold,omitempty" yaml:"approximate_threshold,omitempty"` // For count_distinct
	Precision            *int    `json:"precision,omitempty" yaml:"precision,omitempty"`                         // For average, sum
	SQLDistinctKey       *string `json:"sql_distinct_key,omitempty" yaml:"sql_distinct_key,omitempty"`           // For count_distinct
	Percentile           *int    `json:"percentile,omitempty" yaml:"percentile,omitempty"`                       // For percentile measures
}

// ValidateMeasureAttributes validates that measure attributes are compatible with the measure type
func (m *DbtMetaLookerMeasure) ValidateMeasureAttributes() error {
	measureType := m.Type
	
	// Validate type-specific attributes
	if (m.Approximate != nil || m.ApproximateThreshold != nil || m.SQLDistinctKey != nil) &&
		measureType != enums.MeasureCountDistinct {
		return fmt.Errorf("approximate, approximate_threshold, and sql_distinct_key can only be used with count_distinct measures")
	}
	
	if m.Percentile != nil && !isPercentileMeasure(string(measureType)) {
		return fmt.Errorf("percentile can only be used with percentile measures")
	}
	
	if m.Precision != nil && measureType != enums.MeasureAverage && measureType != enums.MeasureSum {
		return fmt.Errorf("precision can only be used with average or sum measures")
	}
	
	return nil
}

// isPercentileMeasure checks if the measure type is a percentile measure
func isPercentileMeasure(measureType string) bool {
	return len(measureType) > 10 && measureType[:10] == "percentile"
}

// DbtMetaLookerJoin represents Looker-specific metadata for joins
type DbtMetaLookerJoin struct {
	JoinModel    *string                         `json:"join_model,omitempty" yaml:"join_model,omitempty"`
	SQLON        *string                         `json:"sql_on,omitempty" yaml:"sql_on,omitempty"`
	Type         *enums.LookerJoinType           `json:"type,omitempty" yaml:"type,omitempty"`
	Relationship *enums.LookerRelationshipType  `json:"relationship,omitempty" yaml:"relationship,omitempty"`
}

// DbtMetaLooker represents Looker metadata for a model
type DbtMetaLooker struct {
	View      *DbtMetaLookerBase       `json:"view,omitempty" yaml:"view,omitempty"`
	Dimension *DbtMetaLookerDimension  `json:"dimension,omitempty" yaml:"dimension,omitempty"`
	Measures  []DbtMetaLookerMeasure   `json:"measures,omitempty" yaml:"measures,omitempty"`
	Joins     []DbtMetaLookerJoin      `json:"joins,omitempty" yaml:"joins,omitempty"`
}

// LookMLDimension represents a dimension in LookML
type LookMLDimension struct {
	Name            string                       `json:"name" yaml:"name"`
	Type            string                       `json:"type" yaml:"type"`
	SQL             string                       `json:"sql" yaml:"sql"`
	Label           *string                      `json:"label,omitempty" yaml:"label,omitempty"`
	Description     *string                      `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden          *bool                        `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	GroupLabel      *string                      `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	GroupItemLabel  *string                      `json:"group_item_label,omitempty" yaml:"group_item_label,omitempty"`
	ValueFormatName *enums.LookerValueFormatName `json:"value_format_name,omitempty" yaml:"value_format_name,omitempty"`
	CanFilter       *bool                        `json:"can_filter,omitempty" yaml:"can_filter,omitempty"`
	ConvertTZ       *bool                        `json:"convert_tz,omitempty" yaml:"convert_tz,omitempty"`
}

// LookMLDimensionGroup represents a dimension group in LookML
type LookMLDimensionGroup struct {
	Name        string                    `json:"name" yaml:"name"`
	Type        string                    `json:"type" yaml:"type"`
	SQL         string                    `json:"sql" yaml:"sql"`
	Label       *string                   `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string                   `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden      *bool                     `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	GroupLabel  *string                   `json:"group_label,omitempty" yaml:"group_label,omitempty"`
	Timeframes  []enums.LookerTimeFrame   `json:"timeframes,omitempty" yaml:"timeframes,omitempty"`
	ConvertTZ   *bool                     `json:"convert_tz,omitempty" yaml:"convert_tz,omitempty"`
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
	Approximate          *bool                        `json:"approximate,omitempty" yaml:"approximate,omitempty"`
	ApproximateThreshold *int                         `json:"approximate_threshold,omitempty" yaml:"approximate_threshold,omitempty"`
	Precision            *int                         `json:"precision,omitempty" yaml:"precision,omitempty"`
	SQLDistinctKey       *string                      `json:"sql_distinct_key,omitempty" yaml:"sql_distinct_key,omitempty"`
	Percentile           *int                         `json:"percentile,omitempty" yaml:"percentile,omitempty"`
	Filters              []DbtMetaLookerMeasureFilter `json:"filters,omitempty" yaml:"filters,omitempty"`
}

// LookMLView represents a view in LookML
type LookMLView struct {
	Name            string                  `json:"name" yaml:"name"`
	SQLTableName    string                  `json:"sql_table_name" yaml:"sql_table_name"`
	Label           *string                 `json:"label,omitempty" yaml:"label,omitempty"`
	Description     *string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden          *bool                   `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	Dimensions      []LookMLDimension       `json:"dimensions,omitempty" yaml:"dimensions,omitempty"`
	DimensionGroups []LookMLDimensionGroup  `json:"dimension_groups,omitempty" yaml:"dimension_groups,omitempty"`
	Measures        []LookMLMeasure         `json:"measures,omitempty" yaml:"measures,omitempty"`
}

// LookMLJoin represents a join in LookML explores
type LookMLJoin struct {
	Name         string                       `json:"name" yaml:"name"`
	ViewLabel    *string                      `json:"view_label,omitempty" yaml:"view_label,omitempty"`
	SQL          *string                      `json:"sql,omitempty" yaml:"sql,omitempty"`
	Type         *enums.LookerJoinType        `json:"type,omitempty" yaml:"type,omitempty"`
	Relationship *enums.LookerRelationshipType `json:"relationship,omitempty" yaml:"relationship,omitempty"`
}

// LookMLExplore represents an explore in LookML
type LookMLExplore struct {
	Name        string        `json:"name" yaml:"name"`
	ViewName    string        `json:"view_name" yaml:"view_name"`
	Label       *string       `json:"label,omitempty" yaml:"label,omitempty"`
	Description *string       `json:"description,omitempty" yaml:"description,omitempty"`
	Hidden      *bool         `json:"hidden,omitempty" yaml:"hidden,omitempty"`
	Joins       []LookMLJoin  `json:"joins,omitempty" yaml:"joins,omitempty"`
}
