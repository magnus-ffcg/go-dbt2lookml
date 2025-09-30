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
	explore := &models.LookMLExplore{
		Name:        g.getExploreName(model),
		ViewName:    g.getExploreViewName(model),
		Label:       g.getExploreLabel(model),
		Description: g.getExploreDescription(model),
		Hidden:      g.getExploreHidden(model),
		Joins:       g.getExploreJoins(model),
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
	// Check if there's a custom label in model meta
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		model.Meta.Looker.View != nil &&
		model.Meta.Looker.View.Label != nil {
		return model.Meta.Looker.View.Label
	}

	// Generate from model name
	label := strings.ReplaceAll(model.Name, "_", " ")
	label = strings.Title(label)
	return &label
}

// getExploreDescription gets the explore description
func (g *ExploreGenerator) getExploreDescription(model *models.DbtModel) *string {
	// Use model description if available
	if model.Description != "" {
		return &model.Description
	}

	// Check meta description
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		model.Meta.Looker.View != nil &&
		model.Meta.Looker.View.Description != nil {
		return model.Meta.Looker.View.Description
	}

	return nil
}

// getExploreHidden gets the explore hidden setting
func (g *ExploreGenerator) getExploreHidden(model *models.DbtModel) *bool {
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		model.Meta.Looker.View != nil &&
		model.Meta.Looker.View.Hidden != nil {
		return model.Meta.Looker.View.Hidden
	}
	return nil
}

// getExploreJoins gets the joins for the explore from model metadata and generates nested view joins
func (g *ExploreGenerator) getExploreJoins(model *models.DbtModel) []models.LookMLJoin {
	var joins []models.LookMLJoin

	// Convert metadata joins to LookML joins if available
	if model.Meta != nil &&
		model.Meta.Looker != nil &&
		len(model.Meta.Looker.Joins) > 0 {
		for _, metaJoin := range model.Meta.Looker.Joins {
			lookmlJoin := g.convertMetaJoinToLookMLJoin(metaJoin)
			joins = append(joins, lookmlJoin)
		}
	}

	// Generate automatic joins for nested views (ARRAY columns)
	nestedViewJoins := g.generateNestedViewJoins(model)
	joins = append(joins, nestedViewJoins...)

	return joins
}

// convertMetaJoinToLookMLJoin converts a metadata join to a LookML join
func (g *ExploreGenerator) convertMetaJoinToLookMLJoin(metaJoin models.DbtMetaLookerJoin) models.LookMLJoin {
	return models.LookMLJoin{
		Name:         "",  // Would need to be derived from JoinModel
		ViewLabel:    nil, // Not available in metadata join
		SQL:          metaJoin.SQLON,
		Type:         metaJoin.Type,
		Relationship: metaJoin.Relationship,
	}
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
	modelLabel := strings.Title(strings.ReplaceAll(baseName, "_", " "))

	// Convert array column name to title case
	arrayLabel := strings.Title(strings.ReplaceAll(strings.ReplaceAll(arrayColumnName, ".", " "), "_", " "))

	return fmt.Sprintf("%s: %s", modelLabel, arrayLabel)
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

// GenerateExploreWithJoins generates an explore with automatic joins based on foreign keys
func (g *ExploreGenerator) GenerateExploreWithJoins(model *models.DbtModel, relatedModels []*models.DbtModel) (*models.LookMLExplore, error) {
	explore, err := g.GenerateExplore(model)
	if err != nil {
		return nil, err
	}

	// Add automatic joins based on foreign key relationships
	// TODO: Fix type conversion for autoJoins
	// autoJoins := g.generateAutoJoins(model, relatedModels)
	// explore.Joins = append(explore.Joins, autoJoins...)

	return explore, nil
}

// generateAutoJoins generates automatic joins based on foreign key patterns
func (g *ExploreGenerator) generateAutoJoins(model *models.DbtModel, relatedModels []*models.DbtModel) []models.DbtMetaLookerJoin {
	var joins []models.DbtMetaLookerJoin

	// Look for foreign key patterns in column names
	for _, column := range model.Columns {
		// Look for columns ending with "_id" that might be foreign keys
		if strings.HasSuffix(column.Name, "_id") {
			// Extract the potential table name
			tableName := strings.TrimSuffix(column.Name, "_id")

			// Find matching model
			for _, relatedModel := range relatedModels {
				if g.isRelatedModel(tableName, relatedModel) {
					join := g.createAutoJoin(column, relatedModel)
					if join != nil {
						joins = append(joins, *join)
					}
					break
				}
			}
		}
	}

	return joins
}

// isRelatedModel checks if a table name matches a related model
func (g *ExploreGenerator) isRelatedModel(tableName string, model *models.DbtModel) bool {
	// Simple matching - could be made more sophisticated
	return strings.Contains(model.Name, tableName) ||
		strings.Contains(model.RelationName, tableName)
}

// createAutoJoin creates an automatic join for a foreign key relationship
func (g *ExploreGenerator) createAutoJoin(fkColumn models.DbtModelColumn, relatedModel *models.DbtModel) *models.DbtMetaLookerJoin {
	// Find the primary key column in the related model
	var pkColumnName string
	for _, column := range relatedModel.Columns {
		if column.IsPrimaryKey {
			pkColumnName = column.Name
			break
		}
	}

	// Default to "id" if no primary key found
	if pkColumnName == "" {
		pkColumnName = "id"
	}

	joinModel := relatedModel.Name
	sqlOn := g.generateJoinSQL(fkColumn.Name, relatedModel.Name, pkColumnName)
	joinType := enums.JoinLeftOuter
	relationship := enums.RelationshipManyToOne

	return &models.DbtMetaLookerJoin{
		JoinModel:    &joinModel,
		SQLON:        &sqlOn,
		Type:         &joinType,
		Relationship: &relationship,
	}
}

// generateJoinSQL generates the SQL ON clause for a join
func (g *ExploreGenerator) generateJoinSQL(fkColumn, joinTable, pkColumn string) string {
	return "${" + joinTable + "." + pkColumn + "} = ${TABLE." + fkColumn + "}"
}

// ValidateExplore validates that an explore is properly configured
func (g *ExploreGenerator) ValidateExplore(explore *models.LookMLExplore) []string {
	var issues []string

	// Check that view name is not empty
	if explore.ViewName == "" {
		issues = append(issues, "explore view_name cannot be empty")
	}

	// Validate joins
	// TODO: Update validation for LookMLJoin type
	// for i, join := range explore.Joins {
	//	if join.Name == "" {
	//		issues = append(issues, fmt.Sprintf("join %d: name cannot be empty", i))
	//	}
	//
	//	if join.SQL == nil || *join.SQL == "" {
	//		issues = append(issues, fmt.Sprintf("join %d: sql cannot be empty", i))
	//	}
	// }

	return issues
}
