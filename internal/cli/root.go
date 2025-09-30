package cli

import (
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

	"github.com/magnus-ffcg/dbt2lookml/internal/config"
	"github.com/magnus-ffcg/dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/dbt2lookml/pkg/parsers"
)

var (
	cfgFile      string
	manifestPath string
	catalogPath  string
	targetDir    string
	outputDir    string
	tag          string
	logLevel     string
	selectModel  string
	exposuresOnly bool
	exposuresTag string
	useTableName bool
	continueOnError bool
	includeModels []string
	excludeModels []string
	timeframes   []string
	includeISOFields bool
	generateLocale bool
	removeSchemaString string
	reportPath string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbt2lookml",
	Short: "Convert dbt models to LookML views",
	Long: `dbt2lookml is a tool that generates LookML views from BigQuery via dbt models.
	
It parses dbt manifest and catalog files to create comprehensive LookML views
with dimensions, measures, and explores for use in Looker.`,
	RunE: runDbt2Lookml,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Core flags
	rootCmd.Flags().StringVar(&manifestPath, "manifest-path", "", "Path to dbt manifest.json file")
	rootCmd.Flags().StringVar(&catalogPath, "catalog-path", "", "Path to dbt catalog.json file")
	rootCmd.Flags().StringVar(&targetDir, "target-dir", ".", "Target directory for output files")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", ".", "Output directory for LookML files")

	// Filtering flags
	rootCmd.Flags().StringVar(&tag, "tag", "", "Filter models by tag")
	rootCmd.Flags().StringVar(&selectModel, "select", "", "Select a specific model")
	rootCmd.Flags().StringSliceVar(&includeModels, "include-models", []string{}, "Include specific models")
	rootCmd.Flags().StringSliceVar(&excludeModels, "exclude-models", []string{}, "Exclude specific models")

	// Exposure flags
	rootCmd.Flags().BoolVar(&exposuresOnly, "exposures-only", false, "Generate only models referenced in exposures")
	rootCmd.Flags().StringVar(&exposuresTag, "exposures-tag", "", "Filter exposures by tag")

	// Generation options
	rootCmd.Flags().BoolVar(&useTableName, "use-table-name", false, "Use table name instead of model name")
	rootCmd.Flags().BoolVar(&generateLocale, "generate-locale", false, "Generate locale-specific formatting")
	rootCmd.Flags().BoolVar(&includeISOFields, "include-iso-fields", false, "Include ISO date/time fields")
	rootCmd.Flags().StringSliceVar(&timeframes, "timeframes", []string{}, "Custom timeframes for date dimensions")
	rootCmd.Flags().StringVar(&removeSchemaString, "remove-schema-string", "", "String to remove from schema names")

	// Utility flags
	rootCmd.Flags().StringVar(&logLevel, "log-level", "INFO", "Log level (DEBUG, INFO, WARN, ERROR)")
	rootCmd.Flags().BoolVar(&continueOnError, "continue-on-error", false, "Continue processing on errors")
	rootCmd.Flags().StringVar(&reportPath, "report", "", "Generate processing report to file")

	// Bind flags to viper
	viper.BindPFlag("manifest_path", rootCmd.Flags().Lookup("manifest-path"))
	viper.BindPFlag("catalog_path", rootCmd.Flags().Lookup("catalog-path"))
	viper.BindPFlag("target_dir", rootCmd.Flags().Lookup("target-dir"))
	viper.BindPFlag("output_dir", rootCmd.Flags().Lookup("output-dir"))
	viper.BindPFlag("tag", rootCmd.Flags().Lookup("tag"))
	viper.BindPFlag("select", rootCmd.Flags().Lookup("select"))
	viper.BindPFlag("include_models", rootCmd.Flags().Lookup("include-models"))
	viper.BindPFlag("exclude_models", rootCmd.Flags().Lookup("exclude-models"))
	viper.BindPFlag("exposures_only", rootCmd.Flags().Lookup("exposures-only"))
	viper.BindPFlag("exposures_tag", rootCmd.Flags().Lookup("exposures-tag"))
	viper.BindPFlag("use_table_name", rootCmd.Flags().Lookup("use-table-name"))
	viper.BindPFlag("generate_locale", rootCmd.Flags().Lookup("generate-locale"))
	viper.BindPFlag("include_iso_fields", rootCmd.Flags().Lookup("include-iso-fields"))
	viper.BindPFlag("timeframes", rootCmd.Flags().Lookup("timeframes"))
	viper.BindPFlag("remove_schema_string", rootCmd.Flags().Lookup("remove-schema-string"))
	viper.BindPFlag("log_level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("continue_on_error", rootCmd.Flags().Lookup("continue-on-error"))
	viper.BindPFlag("report", rootCmd.Flags().Lookup("report"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
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

	// Validate required paths
	if cfg.ManifestPath == "" {
		return fmt.Errorf("manifest-path is required")
	}
	if cfg.CatalogPath == "" {
		return fmt.Errorf("catalog-path is required")
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
	filesGenerated, err := generator.GenerateAll(models)
	if err != nil {
		if cfg.ContinueOnError {
			log.Printf("Warning: Generation completed with errors: %v", err)
		} else {
			return fmt.Errorf("failed to generate LookML: %w", err)
		}
	}

	generateTime := time.Since(generateStart)
	totalTime := time.Since(startTime)

	// Report results
	log.Printf("Generation completed successfully!")
	log.Printf("Files generated: %d", filesGenerated)
	log.Printf("Parsing time: %v", parseTime)
	log.Printf("Generation time: %v", generateTime)
	log.Printf("Total time: %v", totalTime)

	// Generate report if requested
	if cfg.ReportPath != "" {
		if err := generateReport(cfg.ReportPath, models, filesGenerated, parseTime, generateTime, totalTime); err != nil {
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
func generateReport(reportPath string, models interface{}, filesGenerated int, parseTime, generateTime, totalTime time.Duration) error {
	report := map[string]interface{}{
		"timestamp":        time.Now().Format(time.RFC3339),
		"models_processed": len(models.([]interface{})), // This would need proper typing
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
