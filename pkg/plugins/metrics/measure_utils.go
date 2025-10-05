package metrics

import (
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// MeasureUtils provides common utilities for measure generation
type MeasureUtils struct{}

// GetMetricLabel returns the label for a metric, or nil if not set
func (u *MeasureUtils) GetMetricLabel(metric *models.DbtMetric) *string {
	if metric.Label != nil && *metric.Label != "" {
		return metric.Label
	}
	return nil
}

// GetMetricDescription returns the description for a metric, or nil if not set
func (u *MeasureUtils) GetMetricDescription(metric *models.DbtMetric) *string {
	if metric.Description != nil && *metric.Description != "" {
		return metric.Description
	}
	return nil
}

// GetMeasureLabel returns the label for a semantic measure, or nil if not set
func (u *MeasureUtils) GetMeasureLabel(measure *models.DbtSemanticMeasure) *string {
	if measure.Label != nil && *measure.Label != "" {
		return measure.Label
	}
	return nil
}

// GetMeasureDescription returns the description for a semantic measure, or nil if not set
func (u *MeasureUtils) GetMeasureDescription(measure *models.DbtSemanticMeasure) *string {
	if measure.Description != nil && *measure.Description != "" {
		return measure.Description
	}
	return nil
}

// BuildMeasureMap creates a map of measure name to measure for quick lookup
func (u *MeasureUtils) BuildMeasureMap(measures []models.DbtSemanticMeasure) map[string]*models.DbtSemanticMeasure {
	measureMap := make(map[string]*models.DbtSemanticMeasure)
	for i := range measures {
		measureMap[measures[i].Name] = &measures[i]
	}
	return measureMap
}

// BuildLookMLMeasureMap creates a map of LookML measure name to measure for quick lookup
func (u *MeasureUtils) BuildLookMLMeasureMap(measures []models.LookMLMeasure) map[string]*models.LookMLMeasure {
	measureMap := make(map[string]*models.LookMLMeasure)
	for i := range measures {
		measureMap[measures[i].Name] = &measures[i]
	}
	return measureMap
}

// BuildMetricToMeasureMap creates a mapping from metric names to their corresponding measure names
// Used for derived metrics to resolve metric references
func (u *MeasureUtils) BuildMetricToMeasureMap(
	semanticMeasures []models.DbtSemanticMeasure,
	existingMeasures []models.LookMLMeasure,
) map[string]string {
	metricToMeasureMap := make(map[string]string)

	// Add semantic measures (measure name → measure name)
	for _, measure := range semanticMeasures {
		metricToMeasureMap[measure.Name] = measure.Name
	}

	// Add existing LookML measures
	for _, measure := range existingMeasures {
		metricToMeasureMap[measure.Name] = measure.Name
	}

	return metricToMeasureMap
}
