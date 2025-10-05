package metrics

import (
	"fmt"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// MetricMeasureGenerator consolidates generation of all dbt metric types
// into a single coordinated generator
type MetricMeasureGenerator struct {
	config *config.Config
	utils  *MeasureUtils

	// Specialized builders for each metric type
	ratioBuilder   *RatioMetricGenerator
	derivedBuilder *DerivedMetricGenerator
	simpleBuilder  *SimpleMetricGenerator
}

// NewMetricMeasureGenerator creates a new unified metric measure generator
func NewMetricMeasureGenerator(cfg *config.Config) *MetricMeasureGenerator {
	return &MetricMeasureGenerator{
		config:         cfg,
		utils:          &MeasureUtils{},
		ratioBuilder:   NewRatioMetricGenerator(cfg),
		derivedBuilder: NewDerivedMetricGenerator(cfg),
		simpleBuilder:  NewSimpleMetricGenerator(cfg),
	}
}

// GenerationContext holds the context needed for generating metrics
type GenerationContext struct {
	SemanticMeasures   []models.DbtSemanticMeasure
	ExistingMeasures   []models.LookMLMeasure
	MetricToMeasureMap map[string]string
}

// GenerateMetricMeasures generates all metric measures based on the provided metrics
// Returns the generated measures and updates the context with new measures
func (g *MetricMeasureGenerator) GenerateMetricMeasures(
	simpleMetrics []models.DbtMetric,
	ratioMetrics []models.DbtMetric,
	derivedMetrics []models.DbtMetric,
	ctx *GenerationContext,
) ([]*models.LookMLMeasure, error) {

	allMeasures := make([]*models.LookMLMeasure, 0)

	// Initialize context if needed
	if ctx.MetricToMeasureMap == nil {
		ctx.MetricToMeasureMap = g.utils.BuildMetricToMeasureMap(
			ctx.SemanticMeasures,
			ctx.ExistingMeasures,
		)
	}

	// Step 1: Generate simple metrics with filters first
	// (they can be referenced by other metrics)
	if len(simpleMetrics) > 0 {
		simpleMeasures, err := g.generateSimpleMetrics(simpleMetrics, ctx)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Msg("Failed to generate simple metrics")
		} else {
			allMeasures = append(allMeasures, simpleMeasures...)
			// Add to context for reference by other metrics
			for _, m := range simpleMeasures {
				ctx.MetricToMeasureMap[m.Name] = m.Name
				ctx.ExistingMeasures = append(ctx.ExistingMeasures, *m)
			}
		}
	}

	// Step 2: Generate ratio metrics
	// (they can be referenced by derived metrics)
	if len(ratioMetrics) > 0 {
		ratioMeasures, err := g.generateRatioMetrics(ratioMetrics, ctx)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Msg("Failed to generate ratio metrics")
		} else {
			allMeasures = append(allMeasures, ratioMeasures...)
			// Add to context for reference by derived metrics
			for _, m := range ratioMeasures {
				ctx.MetricToMeasureMap[m.Name] = m.Name
				ctx.ExistingMeasures = append(ctx.ExistingMeasures, *m)
			}
		}
	}

	// Step 3: Generate derived metrics last
	// (they may reference simple or ratio metrics)
	if len(derivedMetrics) > 0 {
		derivedMeasures, err := g.generateDerivedMetrics(derivedMetrics, ctx)
		if err != nil {
			g.config.Logger().Warn().
				Err(err).
				Msg("Failed to generate derived metrics")
		} else {
			allMeasures = append(allMeasures, derivedMeasures...)
		}
	}

	return allMeasures, nil
}

// generateSimpleMetrics generates measures from simple metrics with filters
func (g *MetricMeasureGenerator) generateSimpleMetrics(
	metrics []models.DbtMetric,
	ctx *GenerationContext,
) ([]*models.LookMLMeasure, error) {

	// Build measure map for lookup
	measureMap := g.utils.BuildLookMLMeasureMap(ctx.ExistingMeasures)

	// Use the simple metric builder
	measures, err := g.simpleBuilder.GenerateMeasuresFromSimpleMetrics(metrics, measureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to generate simple metrics: %w", err)
	}

	g.config.Logger().Debug().
		Int("count", len(measures)).
		Msg("Generated simple metric measures")

	return measures, nil
}

// generateRatioMetrics generates measures from ratio metrics
func (g *MetricMeasureGenerator) generateRatioMetrics(
	metrics []models.DbtMetric,
	ctx *GenerationContext,
) ([]*models.LookMLMeasure, error) {

	// Filter to only ratio metrics that can be resolved
	measureMap := g.utils.BuildMeasureMap(ctx.SemanticMeasures)

	var applicableMetrics []models.DbtMetric
	for _, metric := range metrics {
		if !metric.IsRatio() {
			continue
		}

		// Check if both numerator and denominator are available
		if metric.TypeParams.Numerator != nil && metric.TypeParams.Denominator != nil {
			numeratorName := metric.TypeParams.Numerator.Name
			denominatorName := metric.TypeParams.Denominator.Name

			_, hasNumerator := measureMap[numeratorName]
			_, hasDenominator := measureMap[denominatorName]

			if hasNumerator && hasDenominator {
				applicableMetrics = append(applicableMetrics, metric)
			}
		}
	}

	if len(applicableMetrics) == 0 {
		return []*models.LookMLMeasure{}, nil
	}

	// Use the ratio metric builder
	measures, err := g.ratioBuilder.GenerateMeasuresFromRatioMetrics(applicableMetrics, measureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ratio metrics: %w", err)
	}

	g.config.Logger().Debug().
		Int("count", len(measures)).
		Msg("Generated ratio metric measures")

	return measures, nil
}

// generateDerivedMetrics generates measures from derived metrics
func (g *MetricMeasureGenerator) generateDerivedMetrics(
	metrics []models.DbtMetric,
	ctx *GenerationContext,
) ([]*models.LookMLMeasure, error) {

	// Filter to only derived metrics that can be resolved
	var applicableMetrics []models.DbtMetric
	for _, metric := range metrics {
		if !metric.IsDerived() {
			continue
		}

		// Check if all referenced metrics are available
		canResolve := true
		if metric.TypeParams.Metrics != nil {
			for _, ref := range metric.TypeParams.Metrics {
				if _, exists := ctx.MetricToMeasureMap[ref.Name]; !exists {
					canResolve = false
					break
				}
			}
		}

		if canResolve {
			applicableMetrics = append(applicableMetrics, metric)
		}
	}

	if len(applicableMetrics) == 0 {
		return []*models.LookMLMeasure{}, nil
	}

	// Sort metrics by dependency order
	sortedMetrics, err := g.derivedBuilder.TopologicalSortMetrics(applicableMetrics)
	if err != nil {
		return nil, fmt.Errorf("failed to sort derived metrics: %w", err)
	}

	// Use the derived metric builder
	measures, err := g.derivedBuilder.GenerateMeasuresFromDerivedMetrics(sortedMetrics, ctx.MetricToMeasureMap)
	if err != nil {
		return nil, fmt.Errorf("failed to generate derived metrics: %w", err)
	}

	g.config.Logger().Debug().
		Int("count", len(measures)).
		Msg("Generated derived metric measures")

	return measures, nil
}
