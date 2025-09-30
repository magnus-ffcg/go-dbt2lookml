package generators

import (
	"context"
	"testing"
	"time"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// TestGenerateAllWithContext tests the context-aware generation
func TestGenerateAllWithContext(t *testing.T) {
	t.Run("successful generation with valid context", func(t *testing.T) {
		cfg := &config.Config{
			OutputDir:       t.TempDir(),
			UseTableName:    false,
			ContinueOnError: false,
		}

		gen := NewLookMLGenerator(cfg)

		// Create test models using proper structure
		model1 := &models.DbtModel{}
		model1.Name = "test_model_1"
		model1.RelationName = "`project.dataset.test_model_1`"
		model1.Columns = make(map[string]models.DbtModelColumn)

		model2 := &models.DbtModel{}
		model2.Name = "test_model_2"
		model2.RelationName = "`project.dataset.test_model_2`"
		model2.Columns = make(map[string]models.DbtModelColumn)

		testModels := []*models.DbtModel{model1, model2}

		ctx := context.Background()
		filesGenerated, err := gen.GenerateAllWithContext(ctx, testModels)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if filesGenerated != 2 {
			t.Errorf("expected 2 files generated, got %d", filesGenerated)
		}
	})

	t.Run("generation respects context cancellation", func(t *testing.T) {
		cfg := &config.Config{
			OutputDir:       t.TempDir(),
			UseTableName:    false,
			ContinueOnError: false,
		}

		gen := NewLookMLGenerator(cfg)

		// Create test model
		model := &models.DbtModel{}
		model.Name = "test_model"
		model.RelationName = "`project.dataset.test_model`"
		model.Columns = make(map[string]models.DbtModelColumn)

		testModels := []*models.DbtModel{model}

		// Create a context that's already cancelled
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		filesGenerated, err := gen.GenerateAllWithContext(ctx, testModels)

		// Should get an error due to cancellation
		if err == nil {
			t.Error("expected cancellation error, got nil")
		}

		t.Logf("Files generated before cancellation: %d", filesGenerated)
		t.Logf("Error: %v", err)
	})

	t.Run("generation with timeout context", func(t *testing.T) {
		cfg := &config.Config{
			OutputDir:       t.TempDir(),
			UseTableName:    false,
			ContinueOnError: false,
		}

		gen := NewLookMLGenerator(cfg)

		model := &models.DbtModel{}
		model.Name = "test_model_1"
		model.RelationName = "`project.dataset.test_model_1`"
		model.Columns = make(map[string]models.DbtModelColumn)

		testModels := []*models.DbtModel{model}

		// Create context with reasonable timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		filesGenerated, err := gen.GenerateAllWithContext(ctx, testModels)

		// Should succeed as timeout is generous
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if filesGenerated != 1 {
			t.Errorf("expected 1 file generated, got %d", filesGenerated)
		}
	})

	t.Run("GenerateAll uses background context", func(t *testing.T) {
		cfg := &config.Config{
			OutputDir:       t.TempDir(),
			UseTableName:    false,
			ContinueOnError: false,
		}

		gen := NewLookMLGenerator(cfg)

		model := &models.DbtModel{}
		model.Name = "test_model"
		model.RelationName = "`project.dataset.test_model`"
		model.Columns = make(map[string]models.DbtModelColumn)

		testModels := []*models.DbtModel{model}

		// GenerateAll should work without explicit context
		filesGenerated, err := gen.GenerateAll(testModels)

		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		if filesGenerated != 1 {
			t.Errorf("expected 1 file generated, got %d", filesGenerated)
		}
	})
}
