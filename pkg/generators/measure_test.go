package generators

import (
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeasureGenerator_GenerateDefaultCountMeasure(t *testing.T) {
	cfg := &config.Config{}
	generator := NewMeasureGenerator(cfg)

	tests := []struct {
		name           string
		model          *models.DbtModel
		expectedResult *models.LookMLMeasure
	}{
		{
			name: "simple model without existing count measure",
			model: &models.DbtModel{
				DbtNode: models.DbtNode{
					Name: "test_model",
				},
				Meta: nil, // No existing measures
			},
			expectedResult: &models.LookMLMeasure{
				Name: "count",
				Type: enums.MeasureCount,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.GenerateDefaultCountMeasure(tt.model)

			if tt.expectedResult == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Name, result.Name)
				assert.Equal(t, tt.expectedResult.Type, result.Type)
				// Default count measure should be minimal - only name and type
				assert.Nil(t, result.SQL)
				assert.Nil(t, result.Label)
				assert.Nil(t, result.Description)
			}
		})
	}
}

// Helper functions
func measureStringPtr(s string) *string {
	return &s
}

func measureBoolPtr(b bool) *bool {
	return &b
}

func measureIntPtr(i int) *int {
	return &i
}

func valueFormatPtr(vf enums.LookerValueFormatName) *enums.LookerValueFormatName {
	return &vf
}
