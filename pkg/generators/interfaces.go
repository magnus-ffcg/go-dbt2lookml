package generators

import (
	"context"

	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// ViewGeneratorInterface defines the interface for generating LookML views
type ViewGeneratorInterface interface {
	// GenerateView generates a LookML view from a dbt model
	GenerateView(model *models.DbtModel) (*models.LookMLView, error)

	// GenerateNestedView generates a nested view for array/struct columns
	GenerateNestedView(model *models.DbtModel, arrayColumn *models.DbtModelColumn) (*models.LookMLView, error)
}

// DimensionGeneratorInterface defines the interface for generating LookML dimensions
type DimensionGeneratorInterface interface {
	// GenerateDimension generates a LookML dimension from a model column
	GenerateDimension(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimension, error)

	// GenerateDimensionGroup generates a LookML dimension group from a model column
	GenerateDimensionGroup(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimensionGroup, error)

	// GetDimensionName gets the dimension name from the column
	GetDimensionName(column *models.DbtModelColumn) string

	// GetDimensionGroupLabel gets the group label for the dimension
	GetDimensionGroupLabel(column *models.DbtModelColumn) *string
}

// MeasureGeneratorInterface defines the interface for generating LookML measures
type MeasureGeneratorInterface interface {
	// GenerateMeasure generates a LookML measure from measure metadata
	GenerateMeasure(model *models.DbtModel, measureMeta *models.DbtMetaLookerMeasure) (*models.LookMLMeasure, error)

	// GenerateDefaultCountMeasure generates a default count measure for a model
	GenerateDefaultCountMeasure(model *models.DbtModel) *models.LookMLMeasure

	// GeneratePrimaryKeyMeasure generates a count distinct measure for primary key columns
	GeneratePrimaryKeyMeasure(model *models.DbtModel, pkColumn *models.DbtModelColumn) *models.LookMLMeasure

	// GenerateNumericMeasures generates sum/average measures for numeric columns
	GenerateNumericMeasures(model *models.DbtModel, column *models.DbtModelColumn) []*models.LookMLMeasure
}

// ExploreGeneratorInterface defines the interface for generating LookML explores
type ExploreGeneratorInterface interface {
	// GenerateExplore generates a LookML explore from a dbt model
	GenerateExplore(model *models.DbtModel) (*models.LookMLExplore, error)

	// GenerateExploreWithJoins generates an explore with automatic joins based on foreign keys
	GenerateExploreWithJoins(model *models.DbtModel, relatedModels []*models.DbtModel) (*models.LookMLExplore, error)

	// ValidateExplore validates that an explore is properly configured
	ValidateExplore(explore *models.LookMLExplore) []string
}

// LookMLGeneratorInterface defines the interface for the main LookML generator
type LookMLGeneratorInterface interface {
	// GenerateAll generates all LookML files for the given models
	GenerateAll(models []*models.DbtModel) (int, error)

	// GenerateAllWithContext generates all LookML files with cancellation support
	GenerateAllWithContext(ctx context.Context, models []*models.DbtModel) (int, error)
}
