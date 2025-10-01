package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/parsers"
)

// cliFlags holds all CLI flag values in a single struct
// This avoids package-level variables and makes testing easier
type cliFlags struct {
	cfgFile            string
	manifestPath       string
	catalogPath        string
	targetDir          string
	outputDir          string
	tag                string
	logLevel           string
	selectModel        string
	exposuresOnly      bool
	exposuresTag       string
	useTableName       bool
	continueOnError    bool
	includeModels      []string
	excludeModels      []string
	timeframes         []string
	removeSchemaString string
	reportPath         string
	flatten            bool
}

// flags is the single instance holding CLI flag values
var flags = &cliFlags{}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbt2lookml",
	Short: "Convert dbt models to LookML views",
	Long: `dbt2lookml generates LookML views from BigQuery via dbt models.

It parses dbt manifest and catalog files to create comprehensive LookML views
with dimensions, measures, and explores for use in Looker. Supports complex
nested structures (STRUCT, ARRAY) and provides flexible error handling.`,
	Example: `  # Generate LookML from dbt artifacts
  dbt2lookml --manifest-path target/manifest.json --catalog-path target/catalog.json --output-dir lookml/views

  # Use a configuration file
  dbt2lookml --config config.yaml

  # Filter by tag and continue on errors
  dbt2lookml --tag looker --continue-on-error --manifest-path target/manifest.json --catalog-path target/catalog.json

  # Generate only models referenced in exposures
  dbt2lookml --exposures-only --manifest-path target/manifest.json --catalog-path target/catalog.json`,
	RunE: runDbt2Lookml,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&flags.cfgFile, "config", "", "Path to configuration file (default: ./config.yaml)")

	// Core flags (Required)
	rootCmd.Flags().StringVar(&flags.manifestPath, "manifest-path", "", "Path to dbt manifest.json file (required)")
	rootCmd.Flags().StringVar(&flags.catalogPath, "catalog-path", "", "Path to dbt catalog.json file (required)")
	rootCmd.Flags().StringVar(&flags.targetDir, "target-dir", ".", "dbt target directory (default: .)")
	rootCmd.Flags().StringVar(&flags.outputDir, "output-dir", ".", "Output directory for generated LookML files (default: .)")

	// Model Filtering
	rootCmd.Flags().StringVar(&flags.tag, "tag", "", "Filter models by dbt tag (e.g., 'looker')")
	rootCmd.Flags().StringVar(&flags.selectModel, "select", "", "Select a specific model by name")
	rootCmd.Flags().StringSliceVar(&flags.includeModels, "include-models", []string{}, "Comma-separated list of models to include")
	rootCmd.Flags().StringSliceVar(&flags.excludeModels, "exclude-models", []string{}, "Comma-separated list of models to exclude")

	// Exposure Filtering
	rootCmd.Flags().BoolVar(&flags.exposuresOnly, "exposures-only", false, "Generate only models referenced in dbt exposures")
	rootCmd.Flags().StringVar(&flags.exposuresTag, "exposures-tag", "", "Filter exposures by tag before processing")

	// Generation Options
	rootCmd.Flags().BoolVar(&flags.useTableName, "use-table-name", false, "Use BigQuery table name instead of dbt model name for view names")
	rootCmd.Flags().StringSliceVar(&flags.timeframes, "timeframes", []string{}, "Custom timeframes for date dimensions (e.g., 'day,week,month')")
	rootCmd.Flags().StringVar(&flags.removeSchemaString, "remove-schema-string", "", "String to remove from schema names in output paths")
	rootCmd.Flags().BoolVar(&flags.flatten, "flatten", false, "Generate all LookML files in output directory without subdirectories")

	// Error Handling & Logging
	rootCmd.Flags().StringVar(&flags.logLevel, "log-level", "INFO", "Logging level: DEBUG, INFO, WARN, ERROR")
	rootCmd.Flags().BoolVar(&flags.continueOnError, "continue-on-error", false, "Continue processing remaining models if errors occur")
	rootCmd.Flags().StringVar(&flags.reportPath, "report", "", "Path to write processing report (JSON format)")

	// Bind flags to viper
	// These errors are safe to ignore as they only fail if the flag doesn't exist (which is a programmer error caught in testing)
	_ = viper.BindPFlag("manifest_path", rootCmd.Flags().Lookup("manifest-path"))
	_ = viper.BindPFlag("catalog_path", rootCmd.Flags().Lookup("catalog-path"))
	_ = viper.BindPFlag("target_dir", rootCmd.Flags().Lookup("target-dir"))
	_ = viper.BindPFlag("output_dir", rootCmd.Flags().Lookup("output-dir"))
	_ = viper.BindPFlag("tag", rootCmd.Flags().Lookup("tag"))
	_ = viper.BindPFlag("select", rootCmd.Flags().Lookup("select"))
	_ = viper.BindPFlag("include_models", rootCmd.Flags().Lookup("include-models"))
	_ = viper.BindPFlag("exclude_models", rootCmd.Flags().Lookup("exclude-models"))
	_ = viper.BindPFlag("exposures_only", rootCmd.Flags().Lookup("exposures-only"))
	_ = viper.BindPFlag("exposures_tag", rootCmd.Flags().Lookup("exposures-tag"))
	_ = viper.BindPFlag("use_table_name", rootCmd.Flags().Lookup("use-table-name"))
	_ = viper.BindPFlag("timeframes", rootCmd.Flags().Lookup("timeframes"))
	_ = viper.BindPFlag("remove_schema_string", rootCmd.Flags().Lookup("remove-schema-string"))
	_ = viper.BindPFlag("flatten", rootCmd.Flags().Lookup("flatten"))
	_ = viper.BindPFlag("log_level", rootCmd.Flags().Lookup("log-level"))
	_ = viper.BindPFlag("continue_on_error", rootCmd.Flags().Lookup("continue-on-error"))
	_ = viper.BindPFlag("report", rootCmd.Flags().Lookup("report"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if flags.cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(flags.cfgFile)
	} else {
		// Search for config in current directory
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s", viper.ConfigFileUsed())
	}
}

// runDbt2Lookml is the main execution function
func runDbt2Lookml(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate file paths exist before proceeding
	if err := cfg.ValidateFilePaths(); err != nil {
		return fmt.Errorf("configuration error: %w", err)
	}

	// Set up logging
	setupLogging(cfg.LogLevel)

	log.Printf("Starting dbt2lookml conversion...")
	log.Printf("Manifest: %s", cfg.ManifestPath)
	log.Printf("Catalog: %s", cfg.CatalogPath)
	log.Printf("Output: %s", cfg.OutputDir)

	// Load manifest and catalog files
	log.Printf("Loading dbt files...")
	parseStart := time.Now()

	rawManifest, err := loadJSONFile(cfg.ManifestPath)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	rawCatalog, err := loadJSONFile(cfg.CatalogPath)
	if err != nil {
		return fmt.Errorf("failed to load catalog: %w", err)
	}

	// Parse dbt data
	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	models, err := parser.GetModels()
	if err != nil {
		return fmt.Errorf("failed to parse models: %w", err)
	}

	parseTime := time.Since(parseStart)
	log.Printf("Parsed %d models in %v", len(models), parseTime)

	if len(models) == 0 {
		log.Printf("No models found matching the specified criteria")
		return nil
	}

	// Generate LookML
	log.Printf("Generating LookML files...")
	generateStart := time.Now()

	generator := generators.NewLookMLGenerator(cfg)

	// Use new error strategy system for better error handling
	var errorStrategy generators.ErrorStrategy
	if cfg.ContinueOnError {
		errorStrategy = generators.ContinueOnError
	} else {
		errorStrategy = generators.FailFast
	}

	opts := generators.GenerationOptions{
		ErrorStrategy: errorStrategy,
		MaxErrors:     0,    // No limit
		Verbose:       true, // Enable verbose logging
	}

	result, err := generator.GenerateAllWithOptions(context.Background(), models, opts)
	if err != nil {
		if cfg.ContinueOnError {
			log.Printf("Warning: Generation completed with errors: %v", err)
		} else {
			return fmt.Errorf("failed to generate LookML: %w", err)
		}
	}

	// Report any model-specific errors
	if result.HasErrors() {
		log.Printf("Warning: %d models failed to generate:", len(result.Errors))
		for _, modelErr := range result.Errors {
			log.Printf("  - %s", modelErr.String())
		}
	}

	generateTime := time.Since(generateStart)
	totalTime := time.Since(startTime)

	// Report results
	if !result.HasErrors() || cfg.ContinueOnError {
		log.Printf("Generation completed!")
	}
	log.Printf("Files generated: %d/%d", result.FilesGenerated, result.ModelsProcessed)
	log.Printf("Parsing time: %v", parseTime)
	log.Printf("Generation time: %v", generateTime)
	log.Printf("Total time: %v", totalTime)

	// Generate report if requested
	if cfg.ReportPath != "" {
		if err := generateReport(cfg.ReportPath, models, result.FilesGenerated, parseTime, generateTime, totalTime); err != nil {
			log.Printf("Warning: Failed to generate report: %v", err)
		} else {
			log.Printf("Report generated: %s", cfg.ReportPath)
		}
	}

	return nil
}

// loadJSONFile loads and parses a JSON file
func loadJSONFile(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", path, err)
	}

	return result, nil
}

// setupLogging configures the logging level
func setupLogging(level string) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	case "INFO":
		log.SetFlags(log.LstdFlags)
	case "WARN", "ERROR":
		log.SetFlags(log.LstdFlags)
	default:
		log.SetFlags(log.LstdFlags)
	}
}

// generateReport creates a processing report
func generateReport(reportPath string, models []*models.DbtModel, filesGenerated int, parseTime, generateTime, totalTime time.Duration) error {
	report := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"models_processed": len(models),
		"files_generated":  filesGenerated,
		"timing": map[string]string{
			"parsing":    parseTime.String(),
			"generation": generateTime.String(),
			"total":      totalTime.String(),
		},
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(reportPath), 0755); err != nil {
		return err
	}

	// Write report
	data, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	return os.WriteFile(reportPath, data, 0644)
}
