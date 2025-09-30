package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all configuration options for dbt2lookml
type Config struct {
	// Core paths
	ManifestPath string `mapstructure:"manifest_path"`
	CatalogPath  string `mapstructure:"catalog_path"`
	TargetDir    string `mapstructure:"target_dir"`
	OutputDir    string `mapstructure:"output_dir"`

	// Filtering options
	Tag           string   `mapstructure:"tag"`
	Select        string   `mapstructure:"select"`
	IncludeModels []string `mapstructure:"include_models"`
	ExcludeModels []string `mapstructure:"exclude_models"`

	// Exposure options
	ExposuresOnly bool   `mapstructure:"exposures_only"`
	ExposuresTag  string `mapstructure:"exposures_tag"`

	// Generation options
	UseTableName       bool     `mapstructure:"use_table_name"`
	GenerateLocale     bool     `mapstructure:"generate_locale"`
	IncludeISOFields   bool     `mapstructure:"include_iso_fields"`
	Timeframes         []string `mapstructure:"timeframes"`
	RemoveSchemaString string   `mapstructure:"remove_schema_string"`
	Flatten            bool     `mapstructure:"flatten"`

	// Utility options
	LogLevel        string `mapstructure:"log_level"`
	ContinueOnError bool   `mapstructure:"continue_on_error"`
	ReportPath      string `mapstructure:"report"`
}

// LoadConfig loads configuration from viper (which handles CLI flags, config files, and env vars)
func LoadConfig() (*Config, error) {
	// Set defaults
	setDefaults()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	viper.SetDefault("target_dir", ".")
	viper.SetDefault("output_dir", ".")
	viper.SetDefault("log_level", "INFO")
	viper.SetDefault("exposures_only", false)
	viper.SetDefault("use_table_name", false)
	viper.SetDefault("generate_locale", false)
	viper.SetDefault("include_iso_fields", false)
	viper.SetDefault("flatten", false)
	viper.SetDefault("continue_on_error", false)
	viper.SetDefault("include_models", []string{})
	viper.SetDefault("exclude_models", []string{})
	viper.SetDefault("timeframes", []string{})
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate required fields
	if c.ManifestPath == "" {
		return fmt.Errorf("manifest_path is required")
	}
	if c.CatalogPath == "" {
		return fmt.Errorf("catalog_path is required")
	}

	// Validate log level
	validLogLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	logLevel := strings.ToUpper(c.LogLevel)
	valid := false
	for _, level := range validLogLevels {
		if logLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log_level: %s (must be one of: %v)", c.LogLevel, validLogLevels)
	}
	c.LogLevel = logLevel

	// Validate timeframes if provided
	if len(c.Timeframes) > 0 {
		validTimeframes := []string{"raw", "date", "week", "month", "quarter", "year", "time"}
		for _, tf := range c.Timeframes {
			valid := false
			for _, validTf := range validTimeframes {
				if strings.ToLower(tf) == validTf {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid timeframe: %s (must be one of: %v)", tf, validTimeframes)
			}
		}
	}

	return nil
}

// GetFilteredModels returns the list of models to include/exclude based on configuration
func (c *Config) GetFilteredModels() (include []string, exclude []string) {
	return c.IncludeModels, c.ExcludeModels
}

// ShouldFilterByExposures returns true if models should be filtered by exposures
func (c *Config) ShouldFilterByExposures() bool {
	return c.ExposuresOnly || c.ExposuresTag != ""
}

// GetExposureTag returns the exposure tag to filter by, or empty string if none
func (c *Config) GetExposureTag() string {
	return c.ExposuresTag
}

// IsDebugMode returns true if debug logging is enabled
func (c *Config) IsDebugMode() bool {
	return c.LogLevel == "DEBUG"
}

// GetOutputPath returns the full output path for a given filename
func (c *Config) GetOutputPath(filename string) string {
	if c.OutputDir == "" || c.OutputDir == "." {
		return filename
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.OutputDir, "/"), filename)
}

// GetTargetPath returns the full target path for a given filename
func (c *Config) GetTargetPath(filename string) string {
	if c.TargetDir == "" || c.TargetDir == "." {
		return filename
	}
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.TargetDir, "/"), filename)
}

// Clone creates a copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c
	
	// Deep copy slices
	if len(c.IncludeModels) > 0 {
		clone.IncludeModels = make([]string, len(c.IncludeModels))
		copy(clone.IncludeModels, c.IncludeModels)
	}
	
	if len(c.ExcludeModels) > 0 {
		clone.ExcludeModels = make([]string, len(c.ExcludeModels))
		copy(clone.ExcludeModels, c.ExcludeModels)
	}
	
	if len(c.Timeframes) > 0 {
		clone.Timeframes = make([]string, len(c.Timeframes))
		copy(clone.Timeframes, c.Timeframes)
	}
	
	return &clone
}

// String returns a string representation of the configuration (excluding sensitive data)
func (c *Config) String() string {
	return fmt.Sprintf("Config{ManifestPath: %s, CatalogPath: %s, OutputDir: %s, LogLevel: %s}", 
		c.ManifestPath, c.CatalogPath, c.OutputDir, c.LogLevel)
}
