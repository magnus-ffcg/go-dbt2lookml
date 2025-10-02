package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigStructure tests the config structure matches expected fields
func TestConfigStructure(t *testing.T) {
	// Test that we can create a config with all the expected fields
	cfg := &config.Config{
		ManifestPath:       "test_manifest.json",
		CatalogPath:        "test_catalog.json",
		TargetDir:          "test_target",
		OutputDir:          "test_output",
		Tag:                "analytics",
		Select:             "specific_model",
		IncludeModels:      []string{"model1", "model2"},
		ExcludeModels:      []string{"test_model"},
		ExposuresOnly:      true,
		ExposuresTag:       "dashboard",
		UseTableName:       true,
		Timeframes:         []string{"day", "week", "month"},
		RemoveSchemaString: "schema_prefix",
		LogLevel:           "INFO",
		ContinueOnError:    true,
		ReportPath:         "report.json",
	}

	// Verify all fields are accessible
	assert.Equal(t, "test_manifest.json", cfg.ManifestPath)
	assert.Equal(t, "test_catalog.json", cfg.CatalogPath)
	assert.Equal(t, "test_target", cfg.TargetDir)
	assert.Equal(t, "test_output", cfg.OutputDir)
	assert.Equal(t, "analytics", cfg.Tag)
	assert.Equal(t, "specific_model", cfg.Select)
	assert.Equal(t, []string{"model1", "model2"}, cfg.IncludeModels)
	assert.Equal(t, []string{"test_model"}, cfg.ExcludeModels)
	assert.True(t, cfg.ExposuresOnly)
	assert.Equal(t, "dashboard", cfg.ExposuresTag)
	assert.True(t, cfg.UseTableName)
	assert.Equal(t, []string{"day", "week", "month"}, cfg.Timeframes)
	assert.Equal(t, "schema_prefix", cfg.RemoveSchemaString)
	assert.Equal(t, "INFO", cfg.LogLevel)
	assert.True(t, cfg.ContinueOnError)
	assert.Equal(t, "report.json", cfg.ReportPath)
}

// TestConfigFileParsing tests configuration file parsing with viper
func TestConfigFileParsing(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "config.yaml")

	configContent := `
manifest_path: "config_manifest.json"
catalog_path: "config_catalog.json"
output_dir: "config_output"
use_table_name: true
continue_on_error: false
select: "config_model"
tag: "config_tag"
include_models:
  - "model1"
  - "model2"
exclude_models:
  - "exclude1"
include_iso_fields: false
timeframes:
  - "date"
  - "week"
remove_schema_string: "schema_prefix"
exposures_only: false
exposures_tag: "dashboard"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	// Reset viper for clean test
	viper.Reset()
	viper.SetConfigFile(configFile)

	err = viper.ReadInConfig()
	require.NoError(t, err)

	// Test loading config using the actual LoadConfig function
	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify config values
	assert.Equal(t, "config_manifest.json", cfg.ManifestPath)
	assert.Equal(t, "config_catalog.json", cfg.CatalogPath)
	assert.Equal(t, "config_output", cfg.OutputDir)
	assert.True(t, cfg.UseTableName)
	assert.False(t, cfg.ContinueOnError)
	assert.Equal(t, "config_model", cfg.Select)
	assert.Equal(t, "config_tag", cfg.Tag)
	assert.Equal(t, []string{"model1", "model2"}, cfg.IncludeModels)
	assert.Equal(t, []string{"exclude1"}, cfg.ExcludeModels)
	assert.Equal(t, []string{"date", "week"}, cfg.Timeframes)
	assert.Equal(t, "schema_prefix", cfg.RemoveSchemaString)
	assert.False(t, cfg.ExposuresOnly)
	assert.Equal(t, "dashboard", cfg.ExposuresTag)
}

// TestConfigValidation tests the config validation functionality
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &config.Config{
				ManifestPath: "manifest.json",
				CatalogPath:  "catalog.json",
				LogLevel:     "INFO",
				LogFormat:    "console",
			},
			expectError: false,
		},
		{
			name: "missing manifest path",
			config: &config.Config{
				CatalogPath: "catalog.json",
				LogLevel:    "INFO",
				LogFormat:   "console",
			},
			expectError: true,
			errorMsg:    "manifest_path is required",
		},
		{
			name: "missing catalog path",
			config: &config.Config{
				ManifestPath: "manifest.json",
				LogLevel:     "INFO",
				LogFormat:    "console",
			},
			expectError: true,
			errorMsg:    "catalog_path is required",
		},
		{
			name: "invalid log level",
			config: &config.Config{
				ManifestPath: "manifest.json",
				CatalogPath:  "catalog.json",
				LogLevel:     "INVALID",
				LogFormat:    "console",
			},
			expectError: true,
			errorMsg:    "invalid log_level",
		},
		{
			name: "invalid log format",
			config: &config.Config{
				ManifestPath: "manifest.json",
				CatalogPath:  "catalog.json",
				LogLevel:     "INFO",
				LogFormat:    "INVALID",
			},
			expectError: true,
			errorMsg:    "invalid log_format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestConfigDefaults tests that viper sets appropriate default values
func TestConfigDefaults(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create minimal config file with only required fields
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "minimal_config.yaml")

	configContent := `
manifest_path: "manifest.json"
catalog_path: "catalog.json"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	viper.SetConfigFile(configFile)
	err = viper.ReadInConfig()
	require.NoError(t, err)

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Check default values are set
	assert.Equal(t, ".", cfg.TargetDir, "Should have default target directory")
	assert.Equal(t, ".", cfg.OutputDir, "Should have default output directory")
	assert.Equal(t, "INFO", cfg.LogLevel, "Should have default log level")
	assert.False(t, cfg.ExposuresOnly, "ExposuresOnly should default to false")
	assert.False(t, cfg.UseTableName, "UseTableName should default to false")
	assert.False(t, cfg.ContinueOnError, "ContinueOnError should default to false")
	assert.Empty(t, cfg.IncludeModels, "IncludeModels should be empty by default")
	assert.Empty(t, cfg.ExcludeModels, "ExcludeModels should be empty by default")
	assert.Empty(t, cfg.Timeframes, "Timeframes should be empty by default")
}
