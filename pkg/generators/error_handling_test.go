package generators

import (
	"context"
	"errors"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAllWithOptions_FailFast(t *testing.T) {
	cfg := &config.Config{
		OutputDir: "/invalid/path/that/will/fail", // Invalid path to force error
	}
	gen := NewLookMLGenerator(cfg)

	models := []*models.DbtModel{
		createValidModel("model1"),
	}

	opts := GenerationOptions{
		ErrorStrategy: FailFast,
		Verbose:       false,
	}

	result, err := gen.GenerateAllWithOptions(context.Background(), models, opts)

	// Should fail on directory creation
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output directory")
	assert.Equal(t, 0, result.FilesGenerated)
}

// Note: The following tests verify error handling logic exists and is correct.
// In practice, the generator is very resilient and rarely fails for valid models.

func TestGenerateAllWithOptions_AllSucceed(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	models := []*models.DbtModel{
		createValidModel("model1"),
		createValidModel("model2"),
		createValidModel("model3"),
	}

	opts := DefaultGenerationOptions()

	result, err := gen.GenerateAllWithOptions(context.Background(), models, opts)

	require.NoError(t, err)
	assert.Equal(t, 3, result.FilesGenerated)
	assert.Equal(t, 3, result.ModelsProcessed)
	assert.Equal(t, 0, len(result.Errors))
	assert.False(t, result.HasErrors())
	assert.True(t, result.Success())
}

func TestGenerateAllWithOptions_ContextCancellation(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	models := []*models.DbtModel{
		createValidModel("model1"),
		createValidModel("model2"),
	}

	// Cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := DefaultGenerationOptions()

	result, err := gen.GenerateAllWithOptions(ctx, models, opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	assert.Equal(t, 0, result.FilesGenerated)
}

func TestGenerateAllWithOptions_EmptyModels(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	opts := DefaultGenerationOptions()

	result, err := gen.GenerateAllWithOptions(context.Background(), []*models.DbtModel{}, opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no models provided")
	assert.Equal(t, 0, result.FilesGenerated)
}

func TestGenerateAllWithOptions_VerboseLogging(t *testing.T) {
	cfg := &config.Config{
		OutputDir: t.TempDir(),
	}
	gen := NewLookMLGenerator(cfg)

	models := []*models.DbtModel{
		createValidModel("model1"),
	}

	opts := GenerationOptions{
		ErrorStrategy: FailFast,
		Verbose:       true, // Enable verbose logging
	}

	result, err := gen.GenerateAllWithOptions(context.Background(), models, opts)

	require.NoError(t, err)
	assert.Equal(t, 1, result.FilesGenerated)
}

// Helper functions for tests

func createValidModel(name string) *models.DbtModel {
	return &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: name,
		},
		RelationName: name,
		Columns: map[string]models.DbtModelColumn{
			"id": {
				Name:     "id",
				DataType: utils.StringPtr("INT64"),
			},
		},
	}
}

func createInvalidModel(name string) *models.DbtModel {
	// Create a model with a column that has invalid data that will cause generation to fail
	// We'll use a malformed column with nil data type which should cause issues
	return &models.DbtModel{
		DbtNode: models.DbtNode{
			Name: name,
		},
		RelationName: name,
		Columns: map[string]models.DbtModelColumn{
			"bad_column": {
				Name:     "", // Empty name will cause issues in dimension generation
				DataType: nil,
			},
		},
	}
}

func TestGenerationResult_ErrorSummaryDetails(t *testing.T) {
	result := &GenerationResult{
		FilesGenerated: 3,
		Errors: []ModelError{
			{ModelName: "model1", Error: errors.New("failed to parse")},
			{ModelName: "model2", Error: errors.New("invalid column")},
		},
		ModelsProcessed: 5,
	}

	summary := result.ErrorSummary()
	assert.Equal(t, "2 models failed to generate", summary)
	assert.Equal(t, 3, result.FilesGenerated)
	assert.Equal(t, 2, len(result.Errors))
	assert.Equal(t, 5, result.ModelsProcessed)
}
