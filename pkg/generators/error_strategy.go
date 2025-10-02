package generators

import "fmt"

// ErrorStrategy defines how the generator should handle errors during generation.
type ErrorStrategy int

const (
	// FailFast stops generation immediately on the first error.
	// This is the safest strategy for production use.
	FailFast ErrorStrategy = iota

	// FailAtEnd collects all errors and continues processing all models.
	// Returns an error at the end if any models failed.
	// Useful for seeing all problems at once.
	FailAtEnd

	// ContinueOnError logs errors but does not fail the generation.
	// Returns success even if some models failed.
	// Use with caution - only for non-critical generation tasks.
	ContinueOnError
)

// String returns a human-readable name for the error strategy.
func (e ErrorStrategy) String() string {
	switch e {
	case FailFast:
		return "FailFast"
	case FailAtEnd:
		return "FailAtEnd"
	case ContinueOnError:
		return "ContinueOnError"
	default:
		return fmt.Sprintf("Unknown(%d)", e)
	}
}

// GenerationOptions configures the behavior of the generation process.
type GenerationOptions struct {
	// ErrorStrategy determines how errors are handled
	ErrorStrategy ErrorStrategy

	// MaxErrors limits the number of errors before stopping (0 = unlimited)
	// Only applies when ErrorStrategy is FailAtEnd
	MaxErrors int
}

// DefaultGenerationOptions returns the default options (FailFast).
func DefaultGenerationOptions() GenerationOptions {
	return GenerationOptions{
		ErrorStrategy: FailFast,
		MaxErrors:     0,
	}
}

// ModelError represents an error that occurred while generating a specific model.
type ModelError struct {
	ModelName string
	Error     error
}

// String returns a formatted error message.
func (e ModelError) String() string {
	return fmt.Sprintf("model %s: %s", e.ModelName, e.Error.Error())
}

// GenerationResult contains the results of a generation operation.
type GenerationResult struct {
	// FilesGenerated is the number of files successfully generated
	FilesGenerated int

	// Errors contains all errors that occurred during generation
	Errors []ModelError

	// ModelsProcessed is the total number of models attempted
	ModelsProcessed int
}

// HasErrors returns true if any errors occurred during generation.
func (r *GenerationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ErrorSummary returns a summary of the errors.
func (r *GenerationResult) ErrorSummary() string {
	if !r.HasErrors() {
		return "no errors"
	}

	if len(r.Errors) == 1 {
		return r.Errors[0].String()
	}

	return fmt.Sprintf("%d models failed to generate", len(r.Errors))
}

// Success returns true if generation completed successfully according to the strategy.
func (r *GenerationResult) Success() bool {
	return !r.HasErrors()
}
