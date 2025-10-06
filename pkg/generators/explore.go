package generators

import (
	"fmt"
	"strings"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
)

// Compile-time check to ensure ExploreGenerator implements ExploreGeneratorInterface
var _ ExploreGeneratorInterface = (*ExploreGenerator)(nil)

// ExploreGenerator handles generation of LookML explores
type ExploreGenerator struct {
	config *config.Config
}

// NewExploreGenerator creates a new ExploreGenerator instance
func NewExploreGenerator(cfg *config.Config) *ExploreGenerator {
	return &ExploreGenerator{
		config: cfg,
	}
}

// GenerateExplore generates a LookML explore from a dbt model
func (g *ExploreGenerator) GenerateExplore(model *models.DbtModel) (*models.LookMLExplore, error) {
	return g.GenerateExploreWithMetrics(model, false, false)
}

// GenerateExploreWithMetrics generates a LookML explore with optional metric view joins
func (g *ExploreGenerator) GenerateExploreWithMetrics(model *models.DbtModel, hasCumulativeMetrics bool, hasConversionMetrics bool) (*models.LookMLExplore, error) {
	joins := g.getExploreJoins(model)

	// Add joins for cumulative and conversion metric views
	metricJoins := g.generateMetricViewJoins(model, hasCumulativeMetrics, hasConversionMetrics)
	joins = append(joins, metricJoins...)

	explore := &models.LookMLExplore{
		Name:        g.getExploreName(model),
		ViewName:    g.getExploreViewName(model),
		Label:       g.getExploreLabel(model),
		Description: g.getExploreDescription(model),
		Hidden:      g.getExploreHidden(model),
		Joins:       joins,
	}

	return explore, nil
}

// getExploreName gets the explore name from the model
func (g *ExploreGenerator) getExploreName(model *models.DbtModel) string {
	if g.config.UseTableName {
		// Extract just the table name from relation_name (remove project.dataset prefix and backticks)
		parts := strings.Split(model.RelationName, ".")
		tableName := parts[len(parts)-1]
		tableName = strings.Trim(tableName, "`")
		return strings.ToLower(tableName)
	}
	return model.Name
}

// getExploreViewName gets the view name that the explore should reference
func (g *ExploreGenerator) getExploreViewName(model *models.DbtModel) string {
	// Usually the same as the explore name
	return g.getExploreName(model)
}

// getExploreLabel gets the explore label
func (g *ExploreGenerator) getExploreLabel(model *models.DbtModel) *string {
	// Generate from model name
	label := utils.ToTitleCase(model.Name)
	return &label
}

// getExploreDescription gets the explore description
func (g *ExploreGenerator) getExploreDescription(model *models.DbtModel) *string {
	// Use model description if available
	if model.Description != "" {
		return &model.Description
	}
	return nil
}

// getExploreHidden gets the explore hidden setting
func (g *ExploreGenerator) getExploreHidden(model *models.DbtModel) *bool {
	// Explores are always visible by default
	return nil
}

// getExploreJoins gets the joins for the explore from model metadata and generates nested view joins
func (g *ExploreGenerator) getExploreJoins(model *models.DbtModel) []models.LookMLJoin {
	// Generate automatic joins for nested views (ARRAY columns)
	return g.generateNestedViewJoins(model)
}

// generateNestedViewJoins generates joins for nested views based on ARRAY columns
func (g *ExploreGenerator) generateNestedViewJoins(model *models.DbtModel) []models.LookMLJoin {
	var joins []models.LookMLJoin

	// Use column collections to identify ARRAY columns that need nested view joins
	columnCollections := models.NewColumnCollections(model, nil)

	// Generate a join for each nested view
	for arrayColumnName := range columnCollections.NestedViewColumns {
		join := g.createNestedViewJoin(model, arrayColumnName)
		joins = append(joins, join)
	}

	return joins
}

// createNestedViewJoin creates a join for a specific nested view
func (g *ExploreGenerator) createNestedViewJoin(model *models.DbtModel, arrayColumnName string) models.LookMLJoin {
	// Generate view name for the nested view
	nestedViewName := g.getNestedViewName(model, arrayColumnName)

	// Generate view label
	viewLabel := g.getNestedViewLabel(model, arrayColumnName)

	// Generate SQL for the join
	sql := g.getNestedViewJoinSQL(model, arrayColumnName, nestedViewName)

	// Create relationship enum value
	relationship := enums.LookerRelationshipType("one_to_many")

	return models.LookMLJoin{
		Name:         nestedViewName,
		ViewLabel:    &viewLabel,
		SQL:          &sql,
		Relationship: &relationship,
	}
}

// getNestedViewName generates the view name for a nested view join
func (g *ExploreGenerator) getNestedViewName(model *models.DbtModel, arrayColumnName string) string {
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

	// Create nested view name by converting to LookML naming convention
	// This handles PascalCase -> snake_case conversion (SupplierInformation -> supplier_information)
	nestedSuffix := utils.ToLookMLName(arrayColumnName)
	return fmt.Sprintf("%s__%s", baseName, nestedSuffix)
}

// getNestedViewLabel generates a human-readable label for the nested view
func (g *ExploreGenerator) getNestedViewLabel(model *models.DbtModel, arrayColumnName string) string {
	// Use the same naming logic as explore names for consistency
	baseName := g.getExploreName(model)
	modelLabel := utils.ToTitleCase(baseName)

	// Convert array column name to title case
	arrayLabel := utils.ToTitleCase(arrayColumnName)

	return fmt.Sprintf("%s: %s", modelLabel, arrayLabel)
}

// generateMetricViewJoins generates joins for cumulative and conversion metric views
func (g *ExploreGenerator) generateMetricViewJoins(model *models.DbtModel, hasCumulativeMetrics bool, hasConversionMetrics bool) []models.LookMLJoin {
	var joins []models.LookMLJoin

	baseName := g.getExploreName(model)

	// Add join for cumulative metrics view
	if hasCumulativeMetrics {
		cumulativeViewName := fmt.Sprintf("%s__cumulative", baseName)
		relationship := enums.LookerRelationshipType("one_to_one")
		joinType := enums.JoinLeftOuter

		// Dynamically detect the primary key from the semantic model
		primaryKey := g.findPrimaryKey(model)
		sqlOn := fmt.Sprintf("${%s.%s} = ${%s.%s}", baseName, primaryKey, cumulativeViewName, primaryKey)

		label := "Cumulative Metrics"

		joins = append(joins, models.LookMLJoin{
			Name:         cumulativeViewName,
			ViewLabel:    &label,
			SQL:          &sqlOn,
			Type:         &joinType,
			Relationship: &relationship,
		})
	}
	if hasConversionMetrics {
		conversionViewName := fmt.Sprintf("%s__conversion", baseName)
		relationship := enums.LookerRelationshipType("one_to_one")
		joinType := enums.JoinLeftOuter

		// Dynamically detect the entity key from the semantic model
		entityKey := g.findEntityKey(model)
		sqlOn := fmt.Sprintf("${%s.%s} = ${%s.%s}", baseName, entityKey, conversionViewName, entityKey)

		label := "Conversion Metrics"

		joins = append(joins, models.LookMLJoin{
			Name:         conversionViewName,
			ViewLabel:    &label,
			SQL:          &sqlOn,
			Type:         &joinType,
			Relationship: &relationship,
		})
	}

	return joins
}

// findPrimaryKey finds the primary key for a model.
// TODO: This needs to be improved to dynamically find the primary key from the model's metadata.
func (g *ExploreGenerator) findPrimaryKey(model *models.DbtModel) string {
	// For now, we'll use a hardcoded default. This should be updated to inspect model constraints or tags.
	return "order_id"
}

// findEntityKey finds the entity key for a model.
// TODO: This should be updated to dynamically find the entity key from the semantic model.
func (g *ExploreGenerator) findEntityKey(model *models.DbtModel) string {
	// For now, we'll use a hardcoded default.
	return "customer_id"
}

// getNestedViewJoinSQL generates the SQL for joining a nested view
func (g *ExploreGenerator) getNestedViewJoinSQL(model *models.DbtModel, arrayColumnName string, nestedViewName string) string {
	// Generate the main view reference using the same logic as explore/view names
	mainViewName := g.getExploreName(model)

	// Convert array column name to LookML reference (dots to double underscores)
	arrayFieldRef := strings.ReplaceAll(arrayColumnName, ".", "__")

	return fmt.Sprintf("LEFT JOIN UNNEST(${%s.%s}) as %s", mainViewName, arrayFieldRef, nestedViewName)
}

// getStringPtr returns a pointer to a string
func getStringPtr(s string) *string {
	return &s
}
