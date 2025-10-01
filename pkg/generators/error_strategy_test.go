package generators

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorStrategy_String(t *testing.T) {
	tests := []struct {
		name     string
		strategy ErrorStrategy
		expected string
	}{
		{
			name:     "FailFast",
			strategy: FailFast,
			expected: "FailFast",
		},
		{
			name:     "FailAtEnd",
			strategy: FailAtEnd,
			expected: "FailAtEnd",
		},
		{
			name:     "ContinueOnError",
			strategy: ContinueOnError,
			expected: "ContinueOnError",
		},
		{
			name:     "Unknown strategy",
			strategy: ErrorStrategy(99),
			expected: "Unknown(99)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultGenerationOptions(t *testing.T) {
	opts := DefaultGenerationOptions()

	assert.Equal(t, FailFast, opts.ErrorStrategy)
	assert.Equal(t, 0, opts.MaxErrors)
	assert.False(t, opts.Verbose)
}

func TestModelError_String(t *testing.T) {
	err := ModelError{
		ModelName: "test_model",
		Error:     errors.New("something went wrong"),
	}

	result := err.String()
	assert.Equal(t, "model test_model: something went wrong", result)
}

func TestGenerationResult_HasErrors(t *testing.T) {
	t.Run("no errors", func(t *testing.T) {
		result := &GenerationResult{
			FilesGenerated:  5,
			Errors:          []ModelError{},
			ModelsProcessed: 5,
		}

		assert.False(t, result.HasErrors())
		assert.True(t, result.Success())
	})

	t.Run("with errors", func(t *testing.T) {
		result := &GenerationResult{
			FilesGenerated: 3,
			Errors: []ModelError{
				{ModelName: "model1", Error: errors.New("error1")},
				{ModelName: "model2", Error: errors.New("error2")},
			},
			ModelsProcessed: 5,
		}

		assert.True(t, result.HasErrors())
		assert.False(t, result.Success())
	})
}

func TestGenerationResult_ErrorSummary(t *testing.T) {
	tests := []struct {
		name     string
		result   *GenerationResult
		expected string
	}{
		{
			name: "no errors",
			result: &GenerationResult{
				FilesGenerated:  5,
				Errors:          []ModelError{},
				ModelsProcessed: 5,
			},
			expected: "no errors",
		},
		{
			name: "single error",
			result: &GenerationResult{
				FilesGenerated: 4,
				Errors: []ModelError{
					{ModelName: "failing_model", Error: errors.New("failed to generate")},
				},
				ModelsProcessed: 5,
			},
			expected: "model failing_model: failed to generate",
		},
		{
			name: "multiple errors",
			result: &GenerationResult{
				FilesGenerated: 2,
				Errors: []ModelError{
					{ModelName: "model1", Error: errors.New("error1")},
					{ModelName: "model2", Error: errors.New("error2")},
					{ModelName: "model3", Error: errors.New("error3")},
				},
				ModelsProcessed: 5,
			},
			expected: "3 models failed to generate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := tt.result.ErrorSummary()
			assert.Equal(t, tt.expected, summary)
		})
	}
}

func TestGenerationOptions_CustomValues(t *testing.T) {
	opts := GenerationOptions{
		ErrorStrategy: FailAtEnd,
		MaxErrors:     10,
		Verbose:       true,
	}

	assert.Equal(t, FailAtEnd, opts.ErrorStrategy)
	assert.Equal(t, 10, opts.MaxErrors)
	assert.True(t, opts.Verbose)
}

func TestGenerationResult_Metrics(t *testing.T) {
	result := &GenerationResult{
		FilesGenerated: 8,
		Errors: []ModelError{
			{ModelName: "model1", Error: errors.New("error1")},
			{ModelName: "model2", Error: errors.New("error2")},
		},
		ModelsProcessed: 10,
	}

	assert.Equal(t, 8, result.FilesGenerated)
	assert.Equal(t, 2, len(result.Errors))
	assert.Equal(t, 10, result.ModelsProcessed)
	assert.True(t, result.HasErrors())
}
