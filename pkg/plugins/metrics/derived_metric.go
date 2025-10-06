package metrics

import (
	"fmt"
	"regexp"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// DerivedMetricGenerator handles generation of LookML measures from dbt derived metrics
// Note: This is now used as a builder by MetricMeasureGenerator
// For direct usage, prefer MetricMeasureGenerator.GenerateMetricMeasures()
type DerivedMetricGenerator struct {
	config *config.Config
}

// NewDerivedMetricGenerator creates a new DerivedMetricGenerator instance
func NewDerivedMetricGenerator(cfg *config.Config) *DerivedMetricGenerator {
	return &DerivedMetricGenerator{
		config: cfg,
	}
}

// GenerateMeasureFromDerivedMetric generates a LookML measure from a dbt derived metric
// A derived metric is an expression composed of other metrics
func (g *DerivedMetricGenerator) GenerateMeasureFromDerivedMetric(
	metric *models.DbtMetric,
	metricToMeasureMap map[string]string, // Maps metric name to its LookML measure name
) (*models.LookMLMeasure, error) {
	if metric == nil {
		return nil, fmt.Errorf("metric is nil")
	}

	if !metric.IsDerived() {
		return nil, fmt.Errorf("metric %s is not a derived type", metric.Name)
	}

	// Validate we have expression
	if metric.TypeParams.Expr == nil || *metric.TypeParams.Expr == "" {
		return nil, fmt.Errorf("derived metric %s missing expr", metric.Name)
	}

	// Build the SQL expression by replacing metric names with measure references
	sql, err := g.buildDerivedSQL(metric, metricToMeasureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL for derived metric %s: %w", metric.Name, err)
	}

	measure := &models.LookMLMeasure{
		Name:        metric.Name,
		Type:        enums.MeasureNumber, // Derived metrics are type: number
		Label:       g.getMetricLabel(metric),
		Description: g.getMetricDescription(metric),
		SQL:         &sql,
	}

	return measure, nil
}

// buildDerivedSQL builds the SQL expression for a derived metric
// Replaces metric names in the expression with LookML measure references
func (g *DerivedMetricGenerator) buildDerivedSQL(
	metric *models.DbtMetric,
	metricToMeasureMap map[string]string,
) (string, error) {
	expr := *metric.TypeParams.Expr

	g.config.Logger().Debug().
		Str("metric", metric.Name).
		Str("expr", expr).
		Msg("Building derived SQL")

	// Get all metric references from the expression
	metricNames := g.extractMetricNames(metric)

	g.config.Logger().Debug().
		Str("metric", metric.Name).
		Strs("referenced_metrics", metricNames).
		Msg("Extracted metric names")

	// Replace each metric name with its LookML measure reference
	result := expr
	for _, metricName := range metricNames {
		measureName, exists := metricToMeasureMap[metricName]
		if !exists {
			g.config.Logger().Warn().
				Str("metric", metric.Name).
				Str("referenced_metric", metricName).
				Msg("Metric not found in map")
			return "", fmt.Errorf("metric %s references unknown metric %s", metric.Name, metricName)
		}

		g.config.Logger().Debug().
			Str("metric_name", metricName).
			Str("measure_name", measureName).
			Msg("Replacing metric with measure")

		// Replace metric name with LookML measure reference ${measure_name}
		// Use word boundaries to avoid partial replacements
		// Note: $$ escapes the $ in replacement strings
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(metricName))
		replacement := fmt.Sprintf("$${%s}", measureName)
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, replacement)
	}

	g.config.Logger().Debug().
		Str("metric", metric.Name).
		Str("result", result).
		Msg("Final derived SQL")

	return result, nil
}

// extractMetricNames extracts metric names from a derived metric's type_params
func (g *DerivedMetricGenerator) extractMetricNames(metric *models.DbtMetric) []string {
	if metric.TypeParams.Metrics == nil {
		return []string{}
	}

	names := make([]string, 0, len(metric.TypeParams.Metrics))
	for _, metricRef := range metric.TypeParams.Metrics {
		names = append(names, metricRef.Name)
	}
	return names
}

// getMetricLabel returns the label for the metric
func (g *DerivedMetricGenerator) getMetricLabel(metric *models.DbtMetric) *string {
	if metric.Label != nil && *metric.Label != "" {
		return metric.Label
	}
	return nil
}

// getMetricDescription returns the description for the metric
func (g *DerivedMetricGenerator) getMetricDescription(metric *models.DbtMetric) *string {
	if metric.Description != nil && *metric.Description != "" {
		return metric.Description
	}
	return nil
}

// GenerateMeasuresFromDerivedMetrics generates LookML measures from all derived metrics
// The metrics must be provided in dependency order (referenced metrics before metrics that reference them)
func (g *DerivedMetricGenerator) GenerateMeasuresFromDerivedMetrics(
	derivedMetrics []models.DbtMetric,
	metricToMeasureMap map[string]string,
) ([]*models.LookMLMeasure, error) {
	measures := make([]*models.LookMLMeasure, 0, len(derivedMetrics))

	for i := range derivedMetrics {
		measure, err := g.GenerateMeasureFromDerivedMetric(&derivedMetrics[i], metricToMeasureMap)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Str("metric", derivedMetrics[i].Name).
				Msg("Failed to generate measure from derived metric")
			continue
		}

		measures = append(measures, measure)

		// Add this derived metric to the map so it can be referenced by other derived metrics
		metricToMeasureMap[derivedMetrics[i].Name] = measure.Name
	}

	return measures, nil
}

// TopologicalSortMetrics sorts metrics in dependency order (dependencies first)
// This ensures that when we generate measures, all referenced metrics are already generated
func (g *DerivedMetricGenerator) TopologicalSortMetrics(metrics []models.DbtMetric) ([]models.DbtMetric, error) {
	// Build dependency graph
	graph := make(map[string][]string) // metric -> metrics it depends on
	inDegree := make(map[string]int)   // metric -> number of metrics depending on it
	allMetrics := make(map[string]models.DbtMetric)

	for _, metric := range metrics {
		allMetrics[metric.Name] = metric
		if _, exists := inDegree[metric.Name]; !exists {
			inDegree[metric.Name] = 0
		}

		// For derived metrics, add dependencies
		if metric.IsDerived() && metric.TypeParams.Metrics != nil {
			for _, ref := range metric.TypeParams.Metrics {
				graph[metric.Name] = append(graph[metric.Name], ref.Name)
				inDegree[ref.Name]++
			}
		}
	}

	// Kahn's algorithm for topological sort
	var result []models.DbtMetric
	queue := make([]string, 0)

	// Start with metrics that don't depend on anything (inDegree == 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	visited := 0
	for len(queue) > 0 {
		// Dequeue
		current := queue[0]
		queue = queue[1:]

		if metric, exists := allMetrics[current]; exists {
			result = append(result, metric)
			visited++
		}

		// Reduce in-degree for dependents
		for dependent := range graph {
			deps := graph[dependent]
			for _, dep := range deps {
				if dep == current {
					inDegree[dependent]--
					if inDegree[dependent] == 0 {
						queue = append(queue, dependent)
					}
				}
			}
		}
	}

	// Check for cycles
	if visited != len(metrics) {
		return nil, fmt.Errorf("circular dependency detected in derived metrics")
	}

	return result, nil
}

// ValidateDerivedMetric validates that a derived metric can be generated
func (g *DerivedMetricGenerator) ValidateDerivedMetric(
	metric *models.DbtMetric,
	metricToMeasureMap map[string]string,
) error {
	if metric.TypeParams.Expr == nil || *metric.TypeParams.Expr == "" {
		return fmt.Errorf("missing expression")
	}

	if len(metric.TypeParams.Metrics) == 0 {
		return fmt.Errorf("no referenced metrics")
	}

	// Check if all referenced metrics exist
	for _, ref := range metric.TypeParams.Metrics {
		if _, exists := metricToMeasureMap[ref.Name]; !exists {
			return fmt.Errorf("referenced metric %s not found", ref.Name)
		}
	}

	return nil
}
