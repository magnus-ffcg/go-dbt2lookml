package generators

import (
	"context"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// Plugin is the base interface that all generator plugins must implement
type Plugin interface {
	// Name returns the unique name of the plugin
	Name() string

	// Enabled returns whether the plugin is enabled
	Enabled() bool
}

// DataIngestionHook allows plugins to receive data during parsing phase
// Plugins implement this to store semantic models, metrics, etc.
type DataIngestionHook interface {
	Plugin

	// OnManifestLoaded is called when the raw manifest is loaded
	// Plugins parse what they need from the manifest internally
	// This is the new preferred method for data ingestion
	OnManifestLoaded(manifest *models.DbtManifest)

	// OnSemanticMeasures is called when semantic measures are loaded (legacy)
	// Deprecated: Use OnManifestLoaded instead
	OnSemanticMeasures(measures map[string][]models.DbtSemanticMeasure)

	// OnMetrics is called when metrics are loaded by type (legacy)
	// Deprecated: Use OnManifestLoaded instead
	OnMetrics(metrics []models.DbtMetric, metricType string)
}

// ModelGenerationHook allows plugins to participate in model generation lifecycle
type ModelGenerationHook interface {
	Plugin

	// AfterModelGeneration is called after core model generation completes
	// Plugins can generate additional files here (e.g., __metrics.view.lkml)
	AfterModelGeneration(ctx context.Context, model *models.DbtModel) error
}

// ExploreEnrichmentHook allows plugins to enrich explores with additional joins
type ExploreEnrichmentHook interface {
	Plugin

	// EnrichExplore is called to allow plugins to add joins to explores
	// baseName is the explore name that plugins should use as reference
	EnrichExplore(ctx context.Context, model *models.DbtModel, explore *models.LookMLExplore, baseName string) error
}
