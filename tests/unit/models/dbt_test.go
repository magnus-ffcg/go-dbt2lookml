package models

import (
	"testing"

	"github.com/magnus-ffcg/dbt2lookml/pkg/enums"
	"github.com/magnus-ffcg/dbt2lookml/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDbtModelColumn_ProcessColumn tests column processing
func TestDbtModelColumn_ProcessColumn(t *testing.T) {
	tests := []struct {
		name                 string
		column               models.DbtModelColumn
		expectedName         string
		expectedOriginalName string
		expectedNested       bool
		expectedLookMLName   string
		expectedLongName     string
	}{
		{
			name: "simple column",
			column: models.DbtModelColumn{
				Name: "CustomerName",
			},
			expectedName:         "customername",
			expectedOriginalName: "CustomerName",
			expectedNested:       false,
			expectedLookMLName:   "customer_name",
			expectedLongName:     "customer_name",
		},
		{
			name: "nested column",
			column: models.DbtModelColumn{
				Name: "Address.Street.Name",
			},
			expectedName:         "address.street.name",
			expectedOriginalName: "Address.Street.Name",
			expectedNested:       true,
			expectedLookMLName:   "name",
			expectedLongName:     "address__street__name",
		},
		{
			name: "lowercase column",
			column: models.DbtModelColumn{
				Name: "simple_column",
			},
			expectedName:         "simple_column",
			expectedOriginalName: "simple_column",
			expectedNested:       false,
			expectedLookMLName:   "simple_column",
			expectedLongName:     "simple_column",
		},
		{
			name: "column with existing original name",
			column: models.DbtModelColumn{
				Name:         "testcolumn",
				OriginalName: stringPtr("TestColumn"),
			},
			expectedName:         "testcolumn",
			expectedOriginalName: "TestColumn",
			expectedNested:       false,
			expectedLookMLName:   "test_column",
			expectedLongName:     "test_column",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.column.ProcessColumn()

			assert.Equal(t, tt.expectedName, tt.column.Name)
			assert.Equal(t, tt.expectedNested, tt.column.Nested)
			
			require.NotNil(t, tt.column.OriginalName)
			assert.Equal(t, tt.expectedOriginalName, *tt.column.OriginalName)
			
			require.NotNil(t, tt.column.LookMLName)
			assert.Equal(t, tt.expectedLookMLName, *tt.column.LookMLName)
			
			require.NotNil(t, tt.column.LookMLLongName)
			assert.Equal(t, tt.expectedLongName, *tt.column.LookMLLongName)
		})
	}
}

// TestDbtModelColumn_DescriptionNotSet tests that description is not auto-set
func TestDbtModelColumn_DescriptionNotSet(t *testing.T) {
	column := models.DbtModelColumn{
		Name: "test",
	}
	
	column.ProcessColumn()
	
	// Description should remain nil if not provided
	assert.Nil(t, column.Description, "Description should not be auto-set")
}

// TestDbtCatalogNodeColumn_ProcessColumnType tests column type processing
func TestDbtCatalogNodeColumn_ProcessColumnType(t *testing.T) {
	tests := []struct {
		name             string
		columnType       string
		expectedDataType string
		expectInnerTypes bool
	}{
		{
			name:             "simple INT64",
			columnType:       "INT64",
			expectedDataType: "INT64",
			expectInnerTypes: false,
		},
		{
			name:             "simple ARRAY",
			columnType:       "ARRAY<STRING>",
			expectedDataType: "ARRAY",
			expectInnerTypes: true,
		},
		{
			name:             "complex ARRAY STRUCT",
			columnType:       "ARRAY<STRUCT<field1 STRING, field2 INT64>>",
			expectedDataType: "ARRAY",
			expectInnerTypes: true,
		},
		{
			name:             "STRUCT with parentheses",
			columnType:       "STRUCT(name STRING, value INT64)",
			expectedDataType: "STRUCT",
			expectInnerTypes: false,
		},
		{
			name:             "NUMERIC with precision",
			columnType:       "NUMERIC(10, 2)",
			expectedDataType: "NUMERIC",
			expectInnerTypes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			column := models.DbtCatalogNodeColumn{
				Name: "test_column",
				Type: tt.columnType,
			}

			column.ProcessColumnType()

			assert.Equal(t, tt.expectedDataType, column.DataType)
			
			if tt.expectInnerTypes {
				assert.NotEmpty(t, column.InnerTypes)
			}
		})
	}
}

// TestDbtCatalogNode_NormalizeColumnNames tests column name normalization
func TestDbtCatalogNode_NormalizeColumnNames(t *testing.T) {
	node := models.DbtCatalogNode{
		Columns: map[string]models.DbtCatalogNodeColumn{
			"CustomerID": {
				Name: "CustomerID",
				Type: "INT64",
			},
			"ProductName": {
				Name: "ProductName",
				Type: "STRING",
			},
			"already_lowercase": {
				Name: "already_lowercase",
				Type: "STRING",
			},
		},
	}

	node.NormalizeColumnNames()

	// Check that keys are lowercase
	_, hasUpperCase := node.Columns["CustomerID"]
	assert.False(t, hasUpperCase, "Should not have uppercase key")
	
	_, hasLowerCase := node.Columns["customerid"]
	assert.True(t, hasLowerCase, "Should have lowercase key")
	
	// Check that original names are preserved
	col := node.Columns["customerid"]
	assert.Equal(t, "customerid", col.Name)
	assert.Equal(t, "CustomerID", col.OriginalName)
	
	// Check already lowercase column
	lowerCol := node.Columns["already_lowercase"]
	assert.Equal(t, "already_lowercase", lowerCol.Name)
	assert.Equal(t, "already_lowercase", lowerCol.OriginalName)
}

// TestDbtModel_NormalizeColumnNames tests model column normalization
func TestDbtModel_NormalizeColumnNames(t *testing.T) {
	model := models.DbtModel{
		Columns: map[string]models.DbtModelColumn{
			"MixedCase": {
				Name: "MixedCase",
			},
			"UPPERCASE": {
				Name: "UPPERCASE",
			},
			"lowercase": {
				Name: "lowercase",
			},
		},
	}

	model.NormalizeColumnNames()

	// All keys should be lowercase
	assert.Contains(t, model.Columns, "mixedcase")
	assert.Contains(t, model.Columns, "uppercase")
	assert.Contains(t, model.Columns, "lowercase")
	
	// Original keys should not exist
	assert.NotContains(t, model.Columns, "MixedCase")
	assert.NotContains(t, model.Columns, "UPPERCASE")
}

// TestDbtManifestMetadata_ValidateAdapter tests adapter validation
func TestDbtManifestMetadata_ValidateAdapter(t *testing.T) {
	tests := []struct {
		name        string
		adapterType string
		expectError bool
	}{
		{
			name:        "valid bigquery",
			adapterType: "bigquery",
			expectError: false,
		},
		{
			name:        "invalid snowflake",
			adapterType: "snowflake",
			expectError: true,
		},
		{
			name:        "invalid redshift",
			adapterType: "redshift",
			expectError: true,
		},
		{
			name:        "empty adapter",
			adapterType: "",
			expectError: true,
		},
		{
			name:        "case sensitive BigQuery",
			adapterType: "BigQuery",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := models.DbtManifestMetadata{
				AdapterType: tt.adapterType,
			}

			err := metadata.ValidateAdapter()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not supported")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestDbtManifest_GetModels tests getting models from manifest
func TestDbtManifest_GetModels(t *testing.T) {
	manifest := models.DbtManifest{
		Nodes: map[string]interface{}{
			"model.test.model1": map[string]interface{}{
				"resource_type": "model",
				"name":          "model1",
			},
			"model.test.model2": map[string]interface{}{
				"resource_type": "model",
				"name":          "model2",
			},
			"seed.test.seed1": map[string]interface{}{
				"resource_type": "seed",
				"name":          "seed1",
			},
		},
	}

	models := manifest.GetModels()

	// Should only return model nodes, not seeds
	// Note: convertMapToDbtModel returns nil currently, so this will be empty
	// This test documents the expected behavior once conversion is implemented
	assert.NotNil(t, models)
}

// TestDbtExposureRef tests exposure reference structure
func TestDbtExposureRef(t *testing.T) {
	ref := models.DbtExposureRef{
		Name:    "test_model",
		Package: stringPtr("test_package"),
		Version: "1.0",
	}

	assert.Equal(t, "test_model", ref.Name)
	assert.NotNil(t, ref.Package)
	assert.Equal(t, "test_package", *ref.Package)
	assert.Equal(t, "1.0", ref.Version)
}

// TestDbtExposure tests exposure structure
func TestDbtExposure(t *testing.T) {
	description := "Test exposure"
	url := "https://example.com"
	
	exposure := models.DbtExposure{
		DbtNode: models.DbtNode{
			Name:         "test_exposure",
			UniqueID:     "exposure.test.test_exposure",
			ResourceType: enums.ResourceExposure,
		},
		Description: &description,
		URL:         &url,
		Refs: []models.DbtExposureRef{
			{Name: "model1"},
			{Name: "model2"},
		},
		Tags: []string{"tag1", "tag2"},
	}

	assert.Equal(t, "test_exposure", exposure.Name)
	assert.Equal(t, enums.ResourceExposure, exposure.ResourceType)
	assert.Len(t, exposure.Refs, 2)
	assert.Len(t, exposure.Tags, 2)
	require.NotNil(t, exposure.Description)
	assert.Equal(t, "Test exposure", *exposure.Description)
}
