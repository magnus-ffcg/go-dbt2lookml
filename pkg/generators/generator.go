// Package generators provides LookML generation functionality from dbt models.
//
// This package contains the core logic for transforming dbt catalog and manifest
// data into LookML view files. It handles dimension generation, measure creation,
// explore definitions, and nested view structures.
//
// Key components:
//   - LookMLGenerator: Main coordinator for generating all LookML files
//   - ViewGenerator: Generates LookML views from dbt models
//   - DimensionGenerator: Creates dimensions and dimension groups
//   - MeasureGenerator: Generates measures from metadata
//   - ExploreGenerator: Creates explores with join relationships
//
// The package supports:
//   - BigQuery nested/repeated columns (STRUCT/ARRAY)
//   - Custom LookML metadata via dbt meta tags
//   - Dimension groups for date/timestamp fields
//   - Automatic and custom measure generation
//   - Context-aware cancellation for long operations
//
// Example usage:
//
//	cfg := &config.Config{
//	    OutputDir:    "./output",
//	    UseTableName: false,
//	}
//	generator := NewLookMLGenerator(cfg)
package generators

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
)

const (
	// dirPermissions defines the file permissions for created output directories
	dirPermissions = 0755

	// filePermissions defines the file permissions for generated LookML files
	filePermissions = 0644
)

// Compile-time check to ensure LookMLGenerator implements LookMLGeneratorInterface
var _ LookMLGeneratorInterface = (*LookMLGenerator)(nil)

// LookMLGenerator is the main generator that coordinates all LookML generation
type LookMLGenerator struct {
	config             *config.Config
	dimensionGenerator *DimensionGenerator
	viewGenerator      *ViewGenerator
	exploreGenerator   *ExploreGenerator
	measureGenerator   *MeasureGenerator

	// Plugins - hook-based extensibility
	plugins []Plugin
}

// NewLookMLGenerator creates a new LookMLGenerator instance
func NewLookMLGenerator(cfg *config.Config) *LookMLGenerator {
	gen := &LookMLGenerator{
		config:             cfg,
		dimensionGenerator: NewDimensionGenerator(cfg),
		viewGenerator:      NewViewGenerator(cfg),
		exploreGenerator:   NewExploreGenerator(cfg),
		measureGenerator:   NewMeasureGenerator(cfg),
		plugins:            make([]Plugin, 0),
	}

	return gen
}

// RegisterPlugin registers a plugin with the generator
// Plugins receive hooks during generation lifecycle
func (g *LookMLGenerator) RegisterPlugin(plugin Plugin) {
	g.plugins = append(g.plugins, plugin)
}

// LoadManifest passes the raw manifest to plugins for parsing
// This is the new preferred way to handle semantic models and metrics
// Plugins parse what they need internally from the manifest
func (g *LookMLGenerator) LoadManifest(manifest *models.DbtManifest) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnManifestLoaded(manifest)
		}
	}
}

// SetSemanticMeasures sets the semantic measures mapping for generation
// Fires DataIngestionHook to all interested plugins
// Deprecated: Use LoadManifest instead
func (g *LookMLGenerator) SetSemanticMeasures(semanticMeasures map[string][]models.DbtSemanticMeasure) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnSemanticMeasures(semanticMeasures)
		}
	}
}

// SetRatioMetrics sets the ratio metrics for generation
func (g *LookMLGenerator) SetRatioMetrics(ratioMetrics []models.DbtMetric) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnMetrics(ratioMetrics, "ratio")
		}
	}
}

// SetDerivedMetrics sets the derived metrics for generation
func (g *LookMLGenerator) SetDerivedMetrics(derivedMetrics []models.DbtMetric) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnMetrics(derivedMetrics, "derived")
		}
	}
}

// SetSimpleMetrics sets the simple metrics for generation
func (g *LookMLGenerator) SetSimpleMetrics(simpleMetrics []models.DbtMetric) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnMetrics(simpleMetrics, "simple")
		}
	}
}

// SetCumulativeMetrics sets the cumulative metrics for generation
func (g *LookMLGenerator) SetCumulativeMetrics(cumulativeMetrics []models.DbtMetric) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnMetrics(cumulativeMetrics, "cumulative")
		}
	}
}

// SetConversionMetrics sets the conversion metrics for generation
func (g *LookMLGenerator) SetConversionMetrics(conversionMetrics []models.DbtMetric) {
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(DataIngestionHook); ok && hook.Enabled() {
			hook.OnMetrics(conversionMetrics, "conversion")
		}
	}
}

// GenerateAll generates all LookML files for the given models.
// Uses the legacy error handling behavior (respects config.ContinueOnError).
func (g *LookMLGenerator) GenerateAll(models []*models.DbtModel) (int, error) {
	// Call the context-aware version with background context
	return g.GenerateAllWithContext(context.Background(), models)
}

// GenerateAllWithOptions generates all LookML files with configurable error handling.
// This is the recommended method for new code as it provides better error control.
func (g *LookMLGenerator) GenerateAllWithOptions(ctx context.Context, models []*models.DbtModel, opts GenerationOptions) (*GenerationResult, error) {
	result := &GenerationResult{
		FilesGenerated:  0,
		Errors:          []ModelError{},
		ModelsProcessed: 0,
	}

	if len(models) == 0 {
		return result, fmt.Errorf("no models provided for generation")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(g.config.OutputDir, dirPermissions); err != nil {
		return result, fmt.Errorf("failed to create output directory: %w", err)
	}

	for _, model := range models {
		result.ModelsProcessed++

		// Check for cancellation before processing each model
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("generation cancelled after %d files: %w", result.FilesGenerated, ctx.Err())
		default:
			// Continue processing
		}

		g.config.Logger().Debug().Str("model", model.Name).Msg("Generating LookML for model")

		// Generate main view file
		if err := g.generateViewFile(ctx, model); err != nil {
			modelErr := ModelError{
				ModelName: model.Name,
				Error:     err,
			}
			result.Errors = append(result.Errors, modelErr)

			// Apply error strategy
			switch opts.ErrorStrategy {
			case FailFast:
				return result, fmt.Errorf("generation failed on model %s: %w", model.Name, err)

			case FailAtEnd:
				g.config.Logger().Warn().Str("model", model.Name).Err(err).Msg("Failed to generate model")
				// Check if we've hit the error limit
				if opts.MaxErrors > 0 && len(result.Errors) >= opts.MaxErrors {
					return result, fmt.Errorf("too many errors (%d), stopping generation", len(result.Errors))
				}
				continue

			case ContinueOnError:
				g.config.Logger().Warn().Str("model", model.Name).Err(err).Msg("Failed to generate model (continuing)")
				continue
			}
		}

		result.FilesGenerated++
	}

	// Return error only if strategy requires it
	if opts.ErrorStrategy == FailAtEnd && result.HasErrors() {
		return result, fmt.Errorf("%d of %d models failed to generate", len(result.Errors), result.ModelsProcessed)
	}

	return result, nil
}

// GenerateAllWithContext generates all LookML files for the given models with cancellation support
func (g *LookMLGenerator) GenerateAllWithContext(ctx context.Context, models []*models.DbtModel) (int, error) {
	if len(models) == 0 {
		return 0, fmt.Errorf("no models provided for generation")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(g.config.OutputDir, dirPermissions); err != nil {
		return 0, fmt.Errorf("failed to create output directory: %w", err)
	}

	var filesGenerated int
	var errors []string

	for _, model := range models {
		// Check for cancellation before processing each model
		select {
		case <-ctx.Done():
			return filesGenerated, fmt.Errorf("generation cancelled after %d files: %w", filesGenerated, ctx.Err())
		default:
			// Continue processing
		}

		g.config.Logger().Debug().Str("model", model.Name).Msg("Generating LookML for model")

		// Generate main view file (includes explore and nested views inline)
		if err := g.generateViewFile(ctx, model); err != nil {
			if g.config.ContinueOnError {
				errorMsg := fmt.Sprintf("failed to generate view for model %s: %v", model.Name, err)
				g.config.Logger().Warn().Str("model", model.Name).Err(err).Msg("Failed to generate view")
				errors = append(errors, errorMsg)
				continue
			} else {
				return filesGenerated, fmt.Errorf("failed to generate view for model %s: %w", model.Name, err)
			}
		}
		filesGenerated++

		// Note: Nested views and explores are now generated inline in the main view file
		// No separate files needed
	}

	if len(errors) > 0 {
		return filesGenerated, fmt.Errorf("generation completed with %d errors: %s", len(errors), strings.Join(errors, "; "))
	}

	return filesGenerated, nil
}

// generateViewFile generates a LookML view file for a model (includes explore and nested views)
// Base view contains ONLY catalog-based dimensions and a default count measure
func (g *LookMLGenerator) generateViewFile(ctx context.Context, model *models.DbtModel) error {
	var fullContent strings.Builder

	// 1. Generate BASE view from catalog only (no dbt semantic measures)
	// All dbt semantic layer content goes into view extension
	view, err := g.viewGenerator.GenerateView(model)

	if err != nil {
		return fmt.Errorf("failed to generate view: %w", err)
	}

	viewContent, err := g.viewToLookML(view)
	if err != nil {
		return fmt.Errorf("failed to convert view to LookML: %w", err)
	}
	fullContent.WriteString(viewContent)

	// 2. Generate nested views and append them to the same file
	nestedViewsCount, err := g.generateNestedViewsInline(model, &fullContent)
	if err != nil {
		return fmt.Errorf("failed to generate nested views: %w", err)
	}

	// 3. Generate explore section at the bottom
	explore, err := g.exploreGenerator.GenerateExplore(model)
	if err != nil {
		return fmt.Errorf("failed to generate explore: %w", err)
	}

	// 3b. Fire ExploreEnrichmentHook to allow plugins to add joins
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(ExploreEnrichmentHook); ok && hook.Enabled() {
			if err := hook.EnrichExplore(ctx, model, explore, explore.Name); err != nil {
				g.config.Logger().Warn().
					Err(err).
					Str("plugin", plugin.Name()).
					Msg("Plugin explore enrichment failed")
			}
		}
	}

	exploreContent, err := g.exploreToLookML(explore)
	if err != nil {
		return fmt.Errorf("failed to convert explore to LookML: %w", err)
	}
	fullContent.WriteString(exploreContent)

	g.config.Logger().Debug().Int("count", nestedViewsCount).Str("model", model.Name).Msg("Generated nested views inline")

	// Write to file
	filename := g.getViewFilename(model)
	filePath := g.config.GetOutputPath(filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), dirPermissions); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(fullContent.String()), filePermissions); err != nil {
		return fmt.Errorf("failed to write view file: %w", err)
	}

	g.config.Logger().Debug().Str("file", filePath).Msg("Generated view file")

	// 4. Fire AfterModelGeneration hook for plugins to generate additional files
	for _, plugin := range g.plugins {
		if hook, ok := plugin.(ModelGenerationHook); ok && hook.Enabled() {
			if err := hook.AfterModelGeneration(ctx, model); err != nil {
				g.config.Logger().Warn().
					Err(err).
					Str("plugin", plugin.Name()).
					Str("model", model.Name).
					Msg("Plugin after-generation hook failed")
			}
		}
	}

	return nil
}

// generateNestedViews generates nested views for ARRAY/STRUCT columns
func (g *LookMLGenerator) generateNestedViews(model *models.DbtModel) (int, error) {
	// Create column collections to identify array columns
	columnCollections := models.NewColumnCollections(model, nil)

	logger := g.config.Logger()
	logger.Debug().Str("model", model.Name).Int("total_columns", len(model.Columns)).Int("nested_views", len(columnCollections.NestedViewColumns)).Msg("Processing model columns")

	for arrayName := range columnCollections.NestedViewColumns {
		logger.Debug().Str("array", arrayName).Msg("Found array column for nested view")
	}

	var filesGenerated int

	// Generate a nested view for each array column
	for arrayName, nestedColumns := range columnCollections.NestedViewColumns {
		if err := g.generateNestedViewFile(model, arrayName, nestedColumns); err != nil {
			return filesGenerated, fmt.Errorf("failed to generate nested view for %s: %w", arrayName, err)
		}
		filesGenerated++
	}

	return filesGenerated, nil
}

// generateNestedViewFile generates a single nested view file
func (g *LookMLGenerator) generateNestedViewFile(model *models.DbtModel, arrayName string, nestedColumns map[string]models.DbtModelColumn) error {
	// Create a nested view name
	viewName := g.getNestedViewName(model, arrayName)

	// Generate the nested view
	view := &models.LookMLView{
		Name:         viewName,
		SQLTableName: fmt.Sprintf("${%s.SQL_TABLE_NAME}", g.viewGenerator.getViewName(model)),
		Label:        &viewName,
	}

	// Generate dimensions for nested columns
	var dimensions []models.LookMLDimension
	for _, column := range nestedColumns {
		dimension, err := g.dimensionGenerator.GenerateDimension(model, &column)
		if err != nil {
			return fmt.Errorf("failed to generate dimension for nested column %s: %w", column.Name, err)
		}
		if dimension != nil {
			dimensions = append(dimensions, *dimension)
		}
	}
	view.Dimensions = dimensions

	// Convert to LookML string
	lookmlContent, err := g.viewToLookML(view)
	if err != nil {
		return fmt.Errorf("failed to convert nested view to LookML: %w", err)
	}

	// Write to file
	filename := g.getNestedViewFilename(model, arrayName)
	filePath := g.config.GetOutputPath(filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), dirPermissions); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(lookmlContent), filePermissions); err != nil {
		return fmt.Errorf("failed to write nested view file: %w", err)
	}

	g.config.Logger().Debug().Str("file", filePath).Msg("Generated nested view file")
	return nil
}

// generateExploreFile generates a LookML explore file for a model
func (g *LookMLGenerator) generateExploreFile(model *models.DbtModel) error {
	// Generate the explore
	explore, err := g.exploreGenerator.GenerateExplore(model)
	if err != nil {
		return fmt.Errorf("failed to generate explore: %w", err)
	}

	// Convert to LookML string
	lookmlContent, err := g.exploreToLookML(explore)
	if err != nil {
		return fmt.Errorf("failed to convert explore to LookML: %w", err)
	}

	// Write to file
	filename := g.getExploreFilename(model)
	filePath := g.config.GetOutputPath(filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), dirPermissions); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, []byte(lookmlContent), filePermissions); err != nil {
		return fmt.Errorf("failed to write explore file: %w", err)
	}

	g.config.Logger().Debug().Str("file", filePath).Msg("Generated explore file")
	return nil
}

// shouldGenerateExplore determines if an explore should be generated for a model
func (g *LookMLGenerator) shouldGenerateExplore(model *models.DbtModel) bool {
	// Check if model has joins defined in meta
	if model.Meta != nil && model.Meta.Looker != nil && len(model.Meta.Looker.Joins) > 0 {
		return true
	}

	// Default to generating explores for all models (can be configured)
	return true
}

// getViewFilename generates the filename for a view file
func (g *LookMLGenerator) getViewFilename(model *models.DbtModel) string {
	var name string
	var directory string

	if g.config.UseTableName && model.RelationName != "" {
		// Extract just the table name from relation_name (remove project.dataset prefix and backticks)
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		name = strings.ToLower(tableName)

		// Fallback to model name if table name is empty (e.g., ephemeral models)
		if name == "" {
			name = model.Name
		}

		// Use directory structure from model path (unless flatten is enabled)
		if model.Path != "" && !g.config.Flatten {
			directory = filepath.Dir(model.Path)
			directory = strings.Trim(directory, "/")
		}
	} else {
		name = model.Name
		// Use directory structure from model path (unless flatten is enabled)
		if model.Path != "" && !g.config.Flatten {
			directory = filepath.Dir(model.Path)
			directory = strings.Trim(directory, "/")
		}
	}

	// Remove schema string if configured
	if g.config.RemoveSchemaString != "" {
		name = strings.ReplaceAll(name, g.config.RemoveSchemaString, "")
		directory = strings.ReplaceAll(directory, g.config.RemoveSchemaString, "")
	}

	if directory != "" && !g.config.Flatten {
		return fmt.Sprintf("%s/%s.view.lkml", directory, name)
	}
	return fmt.Sprintf("%s.view.lkml", name)
}

// getExploreFilename generates the filename for an explore file
func (g *LookMLGenerator) getExploreFilename(model *models.DbtModel) string {
	var name string
	var directory string

	if g.config.UseTableName && model.RelationName != "" {
		// Extract just the table name from relation_name (remove project.dataset prefix and backticks)
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		name = strings.ToLower(tableName)

		// Fallback to model name if table name is empty (e.g., ephemeral models)
		if name == "" {
			name = model.Name
		}

		// Use directory structure from model path (unless flatten is enabled)
		if model.Path != "" && !g.config.Flatten {
			directory = filepath.Dir(model.Path)
			directory = strings.Trim(directory, "/")
		}
	} else {
		name = model.Name
		// Use directory structure from model path (unless flatten is enabled)
		if model.Path != "" && !g.config.Flatten {
			directory = filepath.Dir(model.Path)
			directory = strings.Trim(directory, "/")
		}
	}

	// Remove schema string if configured
	if g.config.RemoveSchemaString != "" {
		name = strings.ReplaceAll(name, g.config.RemoveSchemaString, "")
		directory = strings.ReplaceAll(directory, g.config.RemoveSchemaString, "")
	}

	// Return the explore filename
	if directory != "" && !g.config.Flatten {
		return fmt.Sprintf("%s/%s.explore.lkml", directory, name)
	}
	return fmt.Sprintf("%s.explore.lkml", name)
}

// getNestedViewName generates the view name for a nested view
func (g *LookMLGenerator) getNestedViewName(model *models.DbtModel, arrayName string) string {
	var baseName string

	if g.config.UseTableName {
		// Extract just the table name from relation_name
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		baseName = strings.ToLower(tableName)
	} else {
		baseName = model.Name
	}

	// Create nested view name by converting to proper LookML naming
	// This handles both PascalCase conversion and dot replacement
	nestedSuffix := utils.ToLookMLName(arrayName)
	return fmt.Sprintf("%s__%s", baseName, nestedSuffix)
}

// getNestedViewNameWithOriginal generates the view name using the original PascalCase name
func (g *LookMLGenerator) getNestedViewNameWithOriginal(model *models.DbtModel, originalArrayName string) string {
	var baseName string

	if g.config.UseTableName {
		// Extract just the table name from relation_name
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		baseName = strings.ToLower(tableName)
	} else {
		baseName = model.Name
	}

	// Create nested view name by converting PascalCase to snake_case with double underscores
	// e.g., "SupplierInformation" -> "supplier_information"
	// e.g., "Markings.Marking" -> "markings__marking"
	nestedSuffix := utils.ToLookMLName(originalArrayName)
	return fmt.Sprintf("%s__%s", baseName, nestedSuffix)
}

// getNestedViewFilename generates the filename for a nested view file
func (g *LookMLGenerator) getNestedViewFilename(model *models.DbtModel, arrayName string) string {
	var directory string

	// Use directory structure from model path
	if model.Path != "" {
		directory = filepath.Dir(model.Path)
		directory = strings.Trim(directory, "/")
	}

	// Remove schema string if configured
	if g.config.RemoveSchemaString != "" {
		directory = strings.ReplaceAll(directory, g.config.RemoveSchemaString, "")
	}

	// Generate nested view name
	viewName := g.getNestedViewName(model, arrayName)

	if directory != "" {
		return fmt.Sprintf("%s/%s.view.lkml", directory, viewName)
	}
	return fmt.Sprintf("%s.view.lkml", viewName)
}

// viewToLookML converts a LookMLView to LookML string format
func (g *LookMLGenerator) viewToLookML(view *models.LookMLView) (string, error) {
	// This is a simplified implementation
	// A full implementation would use a proper LookML serializer
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("view: %s {\n", view.Name))
	builder.WriteString(fmt.Sprintf("  sql_table_name: %s ;;\n", view.SQLTableName))

	if view.Label != nil {
		builder.WriteString(fmt.Sprintf("  label: \"%s\"\n", *view.Label))
	}

	if view.Description != nil {
		builder.WriteString(fmt.Sprintf("  description: \"%s\"\n", *view.Description))
	}

	// Add dimensions
	for _, dimension := range view.Dimensions {
		builder.WriteString(g.dimensionToLookML(&dimension))
	}

	// Add dimension groups
	for _, dimensionGroup := range view.DimensionGroups {
		builder.WriteString(g.dimensionGroupToLookML(&dimensionGroup))
	}

	// Add measures
	for _, measure := range view.Measures {
		builder.WriteString(g.measureToLookML(&measure))
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// lookmlJoinToLookML converts a LookML join to LookML string
func (g *LookMLGenerator) lookmlJoinToLookML(join *models.LookMLJoin) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("  join: %s {\n", join.Name))

	if join.ViewLabel != nil {
		builder.WriteString(fmt.Sprintf("    view_label: \"%s\"\n", *join.ViewLabel))
	}

	if join.SQL != nil {
		builder.WriteString(fmt.Sprintf("    sql: %s ;;\n", *join.SQL))
	}

	if join.Relationship != nil {
		builder.WriteString(fmt.Sprintf("    relationship: %s\n", string(*join.Relationship)))
	}

	builder.WriteString("  }\n")

	return builder.String()
}

// exploreToLookML converts an explore to LookML string
func (g *LookMLGenerator) exploreToLookML(explore *models.LookMLExplore) (string, error) {
	var builder strings.Builder

	builder.WriteString("\n# Un-hide and use this explore, or copy the joins into another explore, to get all the fully nested relationships from this view\n")
	builder.WriteString(fmt.Sprintf("explore: %s {\n", explore.Name))
	builder.WriteString("  hidden: yes\n")

	// Add joins
	for _, join := range explore.Joins {
		builder.WriteString(g.lookmlJoinToLookML(&join))
	}

	builder.WriteString("}\n")

	return builder.String(), nil
}

// dimensionToLookML converts a dimension to LookML string
func (g *LookMLGenerator) dimensionToLookML(dimension *models.LookMLDimension) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("  dimension: %s {\n", dimension.Name))
	builder.WriteString(fmt.Sprintf("    type: %s\n", dimension.Type))
	builder.WriteString(fmt.Sprintf("    sql: %s ;;\n", dimension.SQL))

	// Add group_label if present
	if dimension.GroupLabel != nil {
		builder.WriteString(fmt.Sprintf("    group_label: \"%s\"\n", *dimension.GroupLabel))
	}

	// Add group_item_label if present
	if dimension.GroupItemLabel != nil {
		builder.WriteString(fmt.Sprintf("    group_item_label: \"%s\"\n", *dimension.GroupItemLabel))
	}

	if dimension.Label != nil {
		builder.WriteString(fmt.Sprintf("    label: \"%s\"\n", *dimension.Label))
	}

	if dimension.Description != nil {
		builder.WriteString(fmt.Sprintf("    description: \"%s\"\n", *dimension.Description))
	}

	if dimension.Hidden != nil && *dimension.Hidden {
		builder.WriteString("    hidden: yes\n")
	}

	builder.WriteString("  }\n\n")

	return builder.String()
}

// generateNestedViewsInline generates nested views and appends them to the content builder
func (g *LookMLGenerator) generateNestedViewsInline(model *models.DbtModel, contentBuilder *strings.Builder) (int, error) {
	// Create column collections to identify array columns
	columnCollections := models.NewColumnCollections(model, nil)

	var viewsGenerated int

	// Generate a nested view for each array column
	for arrayName := range columnCollections.NestedViewColumns {
		nestedView, err := g.generateSingleNestedView(model, arrayName, columnCollections.NestedViewColumns[arrayName])
		if err != nil {
			return viewsGenerated, fmt.Errorf("failed to generate nested view for %s: %w", arrayName, err)
		}

		// Convert nested view to LookML and append to content
		nestedViewContent, err := g.viewToLookML(nestedView)
		if err != nil {
			return viewsGenerated, fmt.Errorf("failed to convert nested view to LookML for %s: %w", arrayName, err)
		}

		contentBuilder.WriteString("\n")
		contentBuilder.WriteString(nestedViewContent)
		viewsGenerated++

		g.config.Logger().Debug().Str("view", nestedView.Name).Msg("Generated inline nested view")
	}

	return viewsGenerated, nil
}

// generateSingleNestedView generates a single nested view for an array column
func (g *LookMLGenerator) generateSingleNestedView(model *models.DbtModel, arrayName string, nestedColumns map[string]models.DbtModelColumn) (*models.LookMLView, error) {
	// Find the array column to get its OriginalName for proper view naming
	var arrayColumn *models.DbtModelColumn
	for _, col := range nestedColumns {
		if col.Name == arrayName {
			arrayColumn = &col
			break
		}
	}

	// Generate nested view name using OriginalName if available
	var viewName string
	if arrayColumn != nil && arrayColumn.OriginalName != nil && *arrayColumn.OriginalName != "" {
		viewName = g.getNestedViewNameWithOriginal(model, *arrayColumn.OriginalName)
	} else {
		viewName = g.getNestedViewName(model, arrayName)
	}

	// Create the nested view
	nestedView := &models.LookMLView{
		Name:         viewName,
		SQLTableName: "", // Nested views don't have SQL table names
	}

	// Generate dimensions for nested columns using nested view-specific logic
	var dimensions []models.LookMLDimension
	for _, column := range nestedColumns {
		// Check if this is the array field itself (hidden self-reference)
		if column.Name == arrayName {
			// Determine if we should include the hidden self-reference dimension
			// Rule from Python: include if single-value array OR top-level ARRAY<STRUCT>
			isSingleValueArray := g.isSingleValueArray(&column)
			isTopLevelArray := !strings.Contains(arrayName, ".")

			// Skip if it's a nested ARRAY<STRUCT> (has dot and not single-value)
			if !isSingleValueArray && !isTopLevelArray {
				continue
			}
		}

		dimension, err := g.generateNestedViewDimension(model, viewName, arrayName, &column)
		if err != nil {
			return nil, fmt.Errorf("failed to generate dimension for %s: %w", column.Name, err)
		}
		if dimension != nil {
			dimensions = append(dimensions, *dimension)
		}
	}

	// Assign dimensions to the nested view
	nestedView.Dimensions = dimensions

	return nestedView, nil
}

// isSingleValueArray checks if a column is a single-value array (ARRAY<primitive>, not ARRAY<STRUCT>)
func (g *LookMLGenerator) isSingleValueArray(column *models.DbtModelColumn) bool {
	if column.DataType == nil {
		return false
	}

	dataType := strings.ToUpper(*column.DataType)
	// Single value array = ARRAY<primitive> (no STRUCT)
	return strings.HasPrefix(dataType, "ARRAY") && !strings.Contains(dataType, "STRUCT")
}

// generateNestedViewDimension generates a dimension for a nested view with correct SQL references
func (g *LookMLGenerator) generateNestedViewDimension(model *models.DbtModel, viewName string, arrayName string, column *models.DbtModelColumn) (*models.LookMLDimension, error) {
	// For nested views, create simpler dimensions without extra grouping attributes
	dimension := &models.LookMLDimension{
		Name: g.generateNestedViewDimensionName(model, arrayName, column),
		Type: g.dimensionGenerator.getDimensionType(column),
		SQL:  g.generateNestedViewSQL(viewName, arrayName, column),
	}

	// Add description if available
	if description := g.dimensionGenerator.getDimensionDescription(column); description != nil {
		dimension.Description = description
	}

	// Override hidden property for array fields
	if column.Name == arrayName {
		hidden := true
		dimension.Hidden = &hidden
	}

	return dimension, nil
}

// generateNestedViewDimensionName generates the dimension name for nested view dimensions
func (g *LookMLGenerator) generateNestedViewDimensionName(model *models.DbtModel, arrayName string, column *models.DbtModelColumn) string {
	columnName := column.Name

	// For the array field itself (hidden dimension), use the full nested view name
	if columnName == arrayName {
		// Get the array column's OriginalName for proper conversion
		var arrayOriginalName string
		if column.OriginalName != nil && *column.OriginalName != "" {
			arrayOriginalName = *column.OriginalName
		} else {
			arrayOriginalName = arrayName
		}

		// Generate the full nested view name (same as the view name)
		var baseName string
		if g.config.UseTableName {
			parts := strings.Split(model.RelationName, ".")
			tableName := parts[len(parts)-1]
			tableName = strings.Trim(tableName, "`")
			baseName = strings.ToLower(tableName)
		} else {
			baseName = model.Name
		}

		nestedSuffix := utils.ToLookMLName(arrayOriginalName)
		return fmt.Sprintf("%s__%s", baseName, nestedSuffix)
	}

	// Use OriginalName to preserve PascalCase for proper conversion
	originalName := columnName
	if column.OriginalName != nil && *column.OriginalName != "" {
		originalName = *column.OriginalName
	}

	// For nested fields, extract the path relative to the array
	// e.g., "SupplierInformation.PalletType" with arrayName "supplierinformation" -> "PalletType"
	// Need to find the array prefix in OriginalName (case-insensitive)
	if strings.HasPrefix(strings.ToLower(originalName), strings.ToLower(arrayName)+".") {
		// Find the position after the array name and dot
		prefixLen := len(arrayName) + 1 // +1 for the dot
		nestedPath := originalName[prefixLen:]

		// Convert nested path to LookML dimension name
		// e.g., "GTIN.GTINId" -> "gtin__gtin_id"
		// e.g., "PalletType" -> "pallet_type"
		return utils.ToLookMLName(nestedPath)
	}

	// Fallback: use the original name converted to LookML format
	return utils.ToLookMLName(originalName)
}

// generateNestedViewSQL generates the SQL reference for a nested view dimension
func (g *LookMLGenerator) generateNestedViewSQL(viewName string, arrayName string, column *models.DbtModelColumn) string {
	// For nested view dimensions, we need to reference the nested field structure

	columnName := column.Name

	// Remove the array prefix to get the nested path
	// e.g., "supplierinformation.gtin.gtinid" -> "gtin.gtinid"
	if strings.HasPrefix(columnName, arrayName+".") {
		nestedPath := strings.TrimPrefix(columnName, arrayName+".")
		// Convert to lowercase for consistent SQL references
		nestedPath = strings.ToLower(nestedPath)

		// Use explicit view_name.column if configured, otherwise ${TABLE}.column
		if g.config.NestedViewExplicitReference {
			return fmt.Sprintf("%s.%s", viewName, nestedPath)
		}
		return fmt.Sprintf("${TABLE}.%s", nestedPath)
	}

	// For the array field itself (hidden dimension), use just the view name
	// e.g., "f_store_sales_week_v1__sales" not "${TABLE}.sales"
	if columnName == arrayName {
		return viewName
	}

	// Fallback: regular nested fields
	if g.config.NestedViewExplicitReference {
		return fmt.Sprintf("%s.%s", viewName, strings.ToLower(columnName))
	}
	return fmt.Sprintf("${TABLE}.%s", strings.ToLower(columnName))
}

// dimensionGroupToLookML converts a dimension group to LookML string
func (g *LookMLGenerator) dimensionGroupToLookML(dimensionGroup *models.LookMLDimensionGroup) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("  dimension_group: %s {\n", dimensionGroup.Name))
	builder.WriteString(fmt.Sprintf("    type: %s\n", dimensionGroup.Type))
	builder.WriteString(fmt.Sprintf("    sql: %s ;;\n", dimensionGroup.SQL))

	if len(dimensionGroup.Timeframes) > 0 {
		timeframes := make([]string, len(dimensionGroup.Timeframes))
		for i, tf := range dimensionGroup.Timeframes {
			timeframes[i] = string(tf)
		}
		builder.WriteString(fmt.Sprintf("    timeframes: [%s]\n", strings.Join(timeframes, ", ")))
	}

	builder.WriteString("  }\n\n")

	return builder.String()
}

// measureToLookML converts a measure to LookML string
func (g *LookMLGenerator) measureToLookML(measure *models.LookMLMeasure) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("  measure: %s {\n", measure.Name))
	builder.WriteString(fmt.Sprintf("    type: %s\n", string(measure.Type)))

	if measure.SQL != nil {
		builder.WriteString(fmt.Sprintf("    sql: %s ;;\n", *measure.SQL))
	}

	if measure.Label != nil {
		builder.WriteString(fmt.Sprintf("    label: \"%s\"\n", *measure.Label))
	}

	builder.WriteString("  }\n\n")

	return builder.String()
}

// joinToLookML converts a join to LookML string
func (g *LookMLGenerator) joinToLookML(join *models.DbtMetaLookerJoin) string {
	var builder strings.Builder

	if join.JoinModel != nil {
		builder.WriteString(fmt.Sprintf("  join: %s {\n", *join.JoinModel))

		if join.SQLON != nil {
			builder.WriteString(fmt.Sprintf("    sql_on: %s ;;\n", *join.SQLON))
		}

		if join.Type != nil {
			builder.WriteString(fmt.Sprintf("    type: %s\n", string(*join.Type)))
		}

		if join.Relationship != nil {
			builder.WriteString(fmt.Sprintf("    relationship: %s\n", string(*join.Relationship)))
		}

		builder.WriteString("  }\n\n")
	}

	return builder.String()
}
