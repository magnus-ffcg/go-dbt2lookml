package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magnus-ffcg/go-dbt2lookml/internal/config"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/generators"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/models"
	"github.com/magnus-ffcg/go-dbt2lookml/pkg/parsers"
	pluginMetrics "github.com/magnus-ffcg/go-dbt2lookml/pkg/plugins/metrics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// cliFlags holds all CLI flag values in a single struct
// This avoids package-level variables and makes testing easier
type cliFlags struct {
	cfgFile                     string
	manifestPath                string
	catalogPath                 string
	targetDir                   string
	outputDir                   string
	tag                         string
	logLevel                    string
	logFormat                   string
	selectModel                 string
	exposuresOnly               bool
	exposuresTag                string
	useTableName                bool
	continueOnError             bool
	includeModels               []string
	excludeModels               []string
	timeframes                  []string
	removeSchemaString          string
	reportPath                  string
	flatten                     bool
	nestedViewExplicitReference bool
	useSemanticModels           bool
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
	Example: `  # Generate LookML from dbt target directory (simplest)
  dbt2lookml --target-dir target --output-dir lookml/views

  # Or specify manifest and catalog paths explicitly
  dbt2lookml --manifest-path target/manifest.json --catalog-path target/catalog.json --output-dir lookml/views

  # Use a configuration file
  dbt2lookml --config config.yaml

  # Filter by tag and continue on errors
  dbt2lookml --target-dir target --tag looker --continue-on-error

  # Generate only models referenced in exposures
  dbt2lookml --target-dir target --exposures-only`,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runDbt2Lookml,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Configuration file flag
	rootCmd.PersistentFlags().StringVar(&flags.cfgFile, "config", "", "Path to configuration file (default: ./config.yaml)")

	// Core flags
	rootCmd.Flags().StringVar(&flags.manifestPath, "manifest-path", "", "Path to dbt manifest.json file")
	rootCmd.Flags().StringVar(&flags.catalogPath, "catalog-path", "", "Path to dbt catalog.json file")
	rootCmd.Flags().StringVar(&flags.targetDir, "target-dir", ".", "dbt target directory (looks for manifest.json and catalog.json here)")
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
	rootCmd.Flags().BoolVar(&flags.nestedViewExplicitReference, "nested-view-explicit-reference", false, "Use explicit view_name.column references in nested views instead of ${TABLE}")

	// Semantic Model Options
	rootCmd.Flags().BoolVar(&flags.useSemanticModels, "use-semantic-models", false, "Parse dbt semantic models and generate LookML measures from them (semantic model measures override meta measures with same name)")

	// Error Handling & Logging
	rootCmd.Flags().StringVar(&flags.logLevel, "log-level", "INFO", "Logging level: DEBUG, INFO, WARN, ERROR")
	rootCmd.Flags().StringVar(&flags.logFormat, "log-format", "console", "Log output format: json, console")
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
	_ = viper.BindPFlag("nested_view_explicit_reference", rootCmd.Flags().Lookup("nested-view-explicit-reference"))
	_ = viper.BindPFlag("use_semantic_models", rootCmd.Flags().Lookup("use-semantic-models"))
	_ = viper.BindPFlag("log_level", rootCmd.Flags().Lookup("log-level"))
	_ = viper.BindPFlag("log_format", rootCmd.Flags().Lookup("log-format"))
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
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

// runDbt2Lookml is the main execution function
func runDbt2Lookml(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Load and setup configuration
	cfg, logger, err := loadAndSetupConfig()
	if err != nil {
		return err
	}
	log.Logger = logger // Set global logger

	log.Info().Msg("Starting dbt2lookml conversion")
	log.Info().Str("manifest", cfg.ManifestPath).Str("catalog", cfg.CatalogPath).Str("output", cfg.OutputDir).Msg("Configuration")

	// Parse dbt files
	parser, dbtModels, parseTime, err := parseDBTFiles(cfg)
	if err != nil {
		return err
	}

	if len(dbtModels) == 0 {
		log.Warn().Msg("No models found matching the specified criteria")
		return nil
	}

	// Generate LookML
	_, result, generateTime, err := generateLookML(cfg, parser, dbtModels)
	if err != nil {
		return err
	}

	// Report results
	reportResults(cfg, result, parseTime, generateTime, time.Since(startTime))

	// Generate report if requested
	if cfg.ReportPath != "" {
		if err := generateReport(cfg.ReportPath, dbtModels, result.FilesGenerated, parseTime, generateTime, time.Since(startTime)); err != nil {
			log.Warn().Err(err).Msg("Failed to generate report")
		} else {
			log.Info().Str("path", cfg.ReportPath).Msg("Report generated")
		}
	}

	return nil
}

// loadAndSetupConfig loads configuration and sets up logging
func loadAndSetupConfig() (*config.Config, zerolog.Logger, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, zerolog.Logger{}, fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := cfg.ValidateFilePaths(); err != nil {
		return nil, zerolog.Logger{}, fmt.Errorf("configuration error: %w", err)
	}

	logger := setupLogging(cfg.LogLevel, cfg.LogFormat)
	cfg.SetLogger(logger)

	return cfg, logger, nil
}

// parseDBTFiles loads and parses dbt manifest and catalog files
func parseDBTFiles(cfg *config.Config) (*parsers.DbtParser, []*models.DbtModel, time.Duration, error) {
	log.Info().Msg("Loading dbt files")
	parseStart := time.Now()

	rawManifest, err := loadJSONFile(cfg.ManifestPath)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to load manifest: %w", err)
	}

	rawCatalog, err := loadJSONFile(cfg.CatalogPath)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to load catalog: %w", err)
	}

	parser, err := parsers.NewDbtParser(cfg, rawManifest, rawCatalog)
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to create parser: %w", err)
	}

	dbtModels, err := parser.GetModels()
	if err != nil {
		return nil, nil, 0, fmt.Errorf("failed to parse models: %w", err)
	}

	parseTime := time.Since(parseStart)
	log.Info().Int("count", len(dbtModels)).Dur("time", parseTime).Msg("Parsed models")

	return parser, dbtModels, parseTime, nil
}

// Deprecated helper functions below - no longer used with new plugin architecture
// Kept temporarily for reference, will be removed in next cleanup

// loadSemanticMeasures loads and processes semantic models if enabled
// Deprecated: Plugins now parse manifest internally
func loadSemanticMeasures(cfg *config.Config, parser *parsers.DbtParser, dbtModels []*models.DbtModel) map[string][]models.DbtSemanticMeasure {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing semantic models")
	semanticModelParser := parser.GetSemanticModelParser()

	if !semanticModelParser.HasSemanticModels() {
		log.Info().Msg("No semantic models found in manifest")
		return nil
	}

	semanticModelMap, err := semanticModelParser.LinkSemanticModelToModel(dbtModels)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to link semantic models to dbt models")
		return nil
	}

	measureMap := make(map[string][]models.DbtSemanticMeasure)
	for modelName, semanticModels := range semanticModelMap {
		var allMeasures []models.DbtSemanticMeasure
		for _, sm := range semanticModels {
			allMeasures = append(allMeasures, sm.Measures...)
		}
		if len(allMeasures) > 0 {
			measureMap[modelName] = allMeasures
			log.Debug().
				Str("model", modelName).
				Int("measures", len(allMeasures)).
				Msg("Added semantic measures for model")
		}
	}

	log.Info().Int("models_with_semantic_measures", len(measureMap)).Msg("Loaded semantic models")
	return measureMap
}

// loadRatioMetrics loads and processes ratio metrics if enabled
// Deprecated: Plugins now parse manifest internally
func loadRatioMetrics(cfg *config.Config, parser *parsers.DbtParser) []models.DbtMetric {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing ratio metrics")
	metricParser := parser.GetMetricParser()

	if !metricParser.HasMetrics() {
		log.Info().Msg("No metrics found in manifest")
		return nil
	}

	ratioMetrics, err := metricParser.GetRatioMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get ratio metrics")
		return nil
	}

	if len(ratioMetrics) == 0 {
		log.Debug().Msg("No ratio metrics found")
		return nil
	}

	log.Info().Int("ratio_metrics_count", len(ratioMetrics)).Msg("Loaded ratio metrics")
	return ratioMetrics
}

// loadDerivedMetrics loads and processes derived metrics if enabled
// Deprecated: Plugins now parse manifest internally
func loadDerivedMetrics(cfg *config.Config, parser *parsers.DbtParser) []models.DbtMetric {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing derived metrics")
	metricParser := parser.GetMetricParser()

	if !metricParser.HasMetrics() {
		log.Info().Msg("No metrics found in manifest")
		return nil
	}

	derivedMetrics, err := metricParser.GetDerivedMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get derived metrics")
		return nil
	}

	if len(derivedMetrics) == 0 {
		log.Debug().Msg("No derived metrics found")
		return nil
	}

	log.Info().Int("derived_metrics_count", len(derivedMetrics)).Msg("Loaded derived metrics")
	return derivedMetrics
}

// loadSimpleMetrics loads and processes simple metrics with filters if enabled
// Deprecated: Plugins now parse manifest internally
func loadSimpleMetrics(cfg *config.Config, parser *parsers.DbtParser) []models.DbtMetric {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing simple metrics with filters")
	metricParser := parser.GetMetricParser()

	if !metricParser.HasMetrics() {
		log.Info().Msg("No metrics found in manifest")
		return nil
	}

	simpleMetrics, err := metricParser.GetSimpleMetricsWithFilters()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get simple metrics")
		return nil
	}

	if len(simpleMetrics) == 0 {
		log.Debug().Msg("No simple metrics with filters found")
		return nil
	}

	log.Info().Int("simple_metrics_count", len(simpleMetrics)).Msg("Loaded simple metrics with filters")
	return simpleMetrics
}

// loadCumulativeMetrics loads and processes cumulative metrics if enabled
func loadCumulativeMetrics(cfg *config.Config, parser *parsers.DbtParser) []models.DbtMetric {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing cumulative metrics")
	metricParser := parser.GetMetricParser()

	if !metricParser.HasMetrics() {
		log.Info().Msg("No metrics found in manifest")
		return nil
	}

	cumulativeMetrics, err := metricParser.GetCumulativeMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get cumulative metrics")
		return nil
	}

	if len(cumulativeMetrics) == 0 {
		log.Debug().Msg("No cumulative metrics found")
		return nil
	}

	log.Info().Int("cumulative_metrics_count", len(cumulativeMetrics)).Msg("Loaded cumulative metrics")
	return cumulativeMetrics
}

// loadConversionMetrics loads and processes conversion metrics if enabled
func loadConversionMetrics(cfg *config.Config, parser *parsers.DbtParser) []models.DbtMetric {
	if !cfg.UseSemanticModels {
		return nil
	}

	log.Info().Msg("Parsing conversion metrics")
	metricParser := parser.GetMetricParser()

	if !metricParser.HasMetrics() {
		log.Info().Msg("No metrics found in manifest")
		return nil
	}

	conversionMetrics, err := metricParser.GetConversionMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get conversion metrics")
		return nil
	}

	if len(conversionMetrics) == 0 {
		log.Debug().Msg("No conversion metrics found")
		return nil
	}

	log.Info().Int("conversion_metrics_count", len(conversionMetrics)).Msg("Loaded conversion metrics")
	return conversionMetrics
}

// generateLookML generates LookML files from dbt models
func generateLookML(cfg *config.Config, parser *parsers.DbtParser, dbtModels []*models.DbtModel) (*generators.LookMLGenerator, *generators.GenerationResult, time.Duration, error) {
	log.Info().Msg("Generating LookML files")
	generateStart := time.Now()

	generator := generators.NewLookMLGenerator(cfg)

	// Register metrics plugin if semantic models enabled
	if cfg.UseSemanticModels {
		metricsPlugin := pluginMetrics.NewMetricsPlugin(cfg)
		generator.RegisterPlugin(metricsPlugin)

		// Pass manifest to plugin for parsing
		// Plugin will parse semantic models and metrics internally
		manifest := parser.GetManifest()
		generator.LoadManifest(manifest)

		log.Info().Msg("Semantic measures and metrics will be parsed by plugin")
	}

	// Configure error handling strategy
	errorStrategy := generators.FailFast
	if cfg.ContinueOnError {
		errorStrategy = generators.ContinueOnError
	}

	opts := generators.GenerationOptions{
		ErrorStrategy: errorStrategy,
		MaxErrors:     0, // No limit
	}

	result, err := generator.GenerateAllWithOptions(context.Background(), dbtModels, opts)
	if err != nil {
		if cfg.ContinueOnError {
			log.Warn().Err(err).Msg("Generation completed with errors")
		} else {
			return nil, nil, 0, fmt.Errorf("failed to generate LookML: %w", err)
		}
	}

	return generator, result, time.Since(generateStart), nil
}

// reportResults logs generation results and any errors
func reportResults(cfg *config.Config, result *generators.GenerationResult, parseTime, generateTime, totalTime time.Duration) {
	if result.HasErrors() {
		log.Warn().Int("failed", len(result.Errors)).Msg("Models failed to generate")
		for _, modelErr := range result.Errors {
			log.Warn().Str("model", modelErr.ModelName).Err(modelErr.Error).Msg("Model generation failed")
		}
	}

	if !result.HasErrors() || cfg.ContinueOnError {
		log.Info().Msg("Generation completed")
	}

	log.Info().Int("files", result.FilesGenerated).Int("models", result.ModelsProcessed).Msg("Results")
	log.Info().Dur("parse", parseTime).Dur("generation", generateTime).Dur("total", totalTime).Msg("Timing")
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

// setupLogging configures zerolog and returns a logger
func setupLogging(level, format string) zerolog.Logger {
	var logger zerolog.Logger

	// Choose output format
	if strings.ToLower(format) == config.LogFormatJSON {
		// JSON output - direct to stderr
		logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	} else {
		// Console output - human-readable
		output := zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: "15:04:05",
			NoColor:    false,
		}
		logger = zerolog.New(output).With().Timestamp().Logger()
	}

	// Set global log level
	switch strings.ToUpper(level) {
	case config.LogLevelDebug:
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case config.LogLevelInfo:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case config.LogLevelWarn:
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case config.LogLevelError:
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	return logger
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
