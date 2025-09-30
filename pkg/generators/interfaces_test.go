package generators

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
)

// TestInterfaceImplementations verifies that all generator types implement their interfaces
func TestInterfaceImplementations(t *testing.T) {
	cfg := &config.Config{
		UseTableName: false,
	}

	t.Run("DimensionGenerator implements DimensionGeneratorInterface", func(t *testing.T) {
		var _ DimensionGeneratorInterface = NewDimensionGenerator(cfg)
	})

	t.Run("ViewGenerator implements ViewGeneratorInterface", func(t *testing.T) {
		var _ ViewGeneratorInterface = NewViewGenerator(cfg)
	})

	t.Run("MeasureGenerator implements MeasureGeneratorInterface", func(t *testing.T) {
		var _ MeasureGeneratorInterface = NewMeasureGenerator(cfg)
	})

	t.Run("ExploreGenerator implements ExploreGeneratorInterface", func(t *testing.T) {
		var _ ExploreGeneratorInterface = NewExploreGenerator(cfg)
	})

	t.Run("LookMLGenerator implements LookMLGeneratorInterface", func(t *testing.T) {
		var _ LookMLGeneratorInterface = NewLookMLGenerator(cfg)
	})
}

// MockDimensionGenerator is a mock implementation for testing
type MockDimensionGenerator struct {
	GenerateDimensionFunc      func(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimension, error)
	GenerateDimensionGroupFunc func(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimensionGroup, error)
	GetDimensionNameFunc       func(column *models.DbtModelColumn) string
	GetDimensionGroupLabelFunc func(column *models.DbtModelColumn) *string
}

func (m *MockDimensionGenerator) GenerateDimension(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimension, error) {
	if m.GenerateDimensionFunc != nil {
		return m.GenerateDimensionFunc(model, column)
	}
	return nil, nil
}

func (m *MockDimensionGenerator) GenerateDimensionGroup(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimensionGroup, error) {
	if m.GenerateDimensionGroupFunc != nil {
		return m.GenerateDimensionGroupFunc(model, column)
	}
	return nil, nil
}

func (m *MockDimensionGenerator) GetDimensionName(column *models.DbtModelColumn) string {
	if m.GetDimensionNameFunc != nil {
		return m.GetDimensionNameFunc(column)
	}
	return column.Name
}

func (m *MockDimensionGenerator) GetDimensionGroupLabel(column *models.DbtModelColumn) *string {
	if m.GetDimensionGroupLabelFunc != nil {
		return m.GetDimensionGroupLabelFunc(column)
	}
	return nil
}

// Compile-time check to ensure MockDimensionGenerator implements DimensionGeneratorInterface
var _ DimensionGeneratorInterface = (*MockDimensionGenerator)(nil)

// TestMockDimensionGenerator demonstrates using a mock for testing
func TestMockDimensionGenerator(t *testing.T) {
	t.Run("mock dimension generator can be used in place of real implementation", func(t *testing.T) {
		expectedName := "custom_dimension_name"

		mock := &MockDimensionGenerator{
			GetDimensionNameFunc: func(column *models.DbtModelColumn) string {
				return expectedName
			},
		}

		// Use mock as the interface
		var dimGen DimensionGeneratorInterface = mock

		// Test
		column := &models.DbtModelColumn{
			Name: "original_name",
		}

		result := dimGen.GetDimensionName(column)

		if result != expectedName {
			t.Errorf("expected %s, got %s", expectedName, result)
		}
	})

	t.Run("mock can track calls and verify behavior", func(t *testing.T) {
		callCount := 0
		expectedDimensionType := "string"

		mock := &MockDimensionGenerator{
			GenerateDimensionFunc: func(model *models.DbtModel, column *models.DbtModelColumn) (*models.LookMLDimension, error) {
				callCount++
				return &models.LookMLDimension{
					Name: column.Name,
					Type: expectedDimensionType,
					SQL:  "${TABLE}." + column.Name,
				}, nil
			},
		}

		var dimGen DimensionGeneratorInterface = mock

		// Call the function
		model := &models.DbtModel{}
		column := &models.DbtModelColumn{Name: "test_column"}

		result, err := dimGen.GenerateDimension(model, column)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if callCount != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.Type != expectedDimensionType {
			t.Errorf("expected type %s, got %s", expectedDimensionType, result.Type)
		}
	})
}

// TestInterfaceDecoupling demonstrates how interfaces enable decoupling
func TestInterfaceDecoupling(t *testing.T) {
	t.Run("functions can accept interface types instead of concrete types", func(t *testing.T) {
		// This function accepts any DimensionGeneratorInterface implementation
		processDimension := func(gen DimensionGeneratorInterface, model *models.DbtModel, column *models.DbtModelColumn) string {
			return gen.GetDimensionName(column)
		}

		// Can use real implementation
		cfg := &config.Config{}
		realGen := NewDimensionGenerator(cfg)
		column := &models.DbtModelColumn{Name: "test"}

		result1 := processDimension(realGen, nil, column)
		if result1 != "test" {
			t.Errorf("expected 'test', got %s", result1)
		}

		// Can use mock implementation
		mockGen := &MockDimensionGenerator{
			GetDimensionNameFunc: func(column *models.DbtModelColumn) string {
				return "mocked_" + column.Name
			},
		}

		result2 := processDimension(mockGen, nil, column)
		if result2 != "mocked_test" {
			t.Errorf("expected 'mocked_test', got %s", result2)
		}
	})
}
