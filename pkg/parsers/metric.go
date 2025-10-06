package parsers

import (
	"fmt"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// MetricParser handles parsing of dbt metrics from manifest
type MetricParser struct {
	manifest *models.DbtManifest
}

// NewMetricParser creates a new MetricParser instance
func NewMetricParser(manifest *models.DbtManifest) *MetricParser {
	return &MetricParser{
		manifest: manifest,
	}
}

// GetMetrics returns all metrics from the manifest
func (p *MetricParser) GetMetrics() ([]models.DbtMetric, error) {
	if p.manifest == nil {
		return nil, fmt.Errorf("manifest is nil")
	}

	if p.manifest.Metrics == nil {
		// No metrics in manifest - this is not an error
		return []models.DbtMetric{}, nil
	}

	metrics := make([]models.DbtMetric, 0, len(p.manifest.Metrics))
	for _, metric := range p.manifest.Metrics {
		metrics = append(metrics, metric)
	}

	return metrics, nil
}

// GetMetricByName returns a specific metric by name
func (p *MetricParser) GetMetricByName(name string) (*models.DbtMetric, error) {
	if p.manifest == nil || p.manifest.Metrics == nil {
		return nil, fmt.Errorf("no metrics available")
	}

	// Search by name in the map values
	for _, metric := range p.manifest.Metrics {
		if metric.Name == name {
			return &metric, nil
		}
	}

	return nil, fmt.Errorf("metric %s not found", name)
}

// GetRatioMetrics returns all ratio-type metrics
func (p *MetricParser) GetRatioMetrics() ([]models.DbtMetric, error) {
	allMetrics, err := p.GetMetrics()
	if err != nil {
		return nil, err
	}

	var ratioMetrics []models.DbtMetric
	for _, metric := range allMetrics {
		if metric.IsRatio() {
			ratioMetrics = append(ratioMetrics, metric)
		}
	}

	return ratioMetrics, nil
}

// GetSimpleMetrics returns all simple-type metrics
func (p *MetricParser) GetSimpleMetrics() ([]models.DbtMetric, error) {
	allMetrics, err := p.GetMetrics()
	if err != nil {
		return nil, err
	}

	var simpleMetrics []models.DbtMetric
	for _, metric := range allMetrics {
		if metric.IsSimple() {
			simpleMetrics = append(simpleMetrics, metric)
		}
	}

	return simpleMetrics, nil
}

// GetDerivedMetrics returns all derived-type metrics
func (p *MetricParser) GetDerivedMetrics() ([]models.DbtMetric, error) {
	allMetrics, err := p.GetMetrics()
	if err != nil {
		return nil, err
	}

	var derivedMetrics []models.DbtMetric
	for _, metric := range allMetrics {
		if metric.IsDerived() {
			derivedMetrics = append(derivedMetrics, metric)
		}
	}

	return derivedMetrics, nil
}

// HasMetrics returns true if the manifest contains any metrics
func (p *MetricParser) HasMetrics() bool {
	if p.manifest == nil || p.manifest.Metrics == nil {
		return false
	}
	return len(p.manifest.Metrics) > 0
}

// HasRatioMetrics returns true if there are any ratio metrics
func (p *MetricParser) HasRatioMetrics() bool {
	ratioMetrics, err := p.GetRatioMetrics()
	if err != nil {
		return false
	}
	return len(ratioMetrics) > 0
}

// HasDerivedMetrics returns true if there are any derived metrics
func (p *MetricParser) HasDerivedMetrics() bool {
	derivedMetrics, err := p.GetDerivedMetrics()
	if err != nil {
		return false
	}
	return len(derivedMetrics) > 0
}

// GetSimpleMetricsWithFilters returns simple metrics that have filter conditions
func (p *MetricParser) GetSimpleMetricsWithFilters() ([]models.DbtMetric, error) {
	simpleMetrics, err := p.GetSimpleMetrics()
	if err != nil {
		return nil, err
	}

	// Filter to only those with filters
	var filteredMetrics []models.DbtMetric
	for _, metric := range simpleMetrics {
		if metric.HasFilter() {
			filteredMetrics = append(filteredMetrics, metric)
		}
	}

	return filteredMetrics, nil
}

// GetCumulativeMetrics returns all cumulative metrics from the manifest
func (p *MetricParser) GetCumulativeMetrics() ([]models.DbtMetric, error) {
	if !p.HasMetrics() {
		return []models.DbtMetric{}, nil
	}

	var cumulativeMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsCumulative() {
			cumulativeMetrics = append(cumulativeMetrics, metric)
		}
	}

	return cumulativeMetrics, nil
}

// GetConversionMetrics returns all conversion metrics from the manifest
func (p *MetricParser) GetConversionMetrics() ([]models.DbtMetric, error) {
	if !p.HasMetrics() {
		return []models.DbtMetric{}, nil
	}

	var conversionMetrics []models.DbtMetric
	for _, metric := range p.manifest.Metrics {
		if metric.IsConversion() {
			conversionMetrics = append(conversionMetrics, metric)
		}
	}

	return conversionMetrics, nil
}
