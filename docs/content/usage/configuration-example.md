# Configuration Example

```yaml
# dbt2lookml Configuration Example
# ====================================
# This file shows all available configuration options with explanations.
# Copy this file to `config.yaml` and customize for your project.

# Required: dbt Artifact Paths
# -----------------------------
# Paths to dbt-generated files (usually in target/ directory after `dbt docs generate`)

manifest_path: target/manifest.json  # dbt model definitions and metadata
catalog_path: target/catalog.json    # BigQuery schema information

# Output Configuration
# --------------------
# Where to write generated LookML files

output_dir: lookml/views  # Directory for generated LookML view files
target_dir: .             # dbt target directory (rarely needs changing)

# Model Filtering
# ---------------
# Control which dbt models get converted to LookML

# Filter by dbt tag (only models with this tag)
# tag: looker

# Select a specific model by name
# select: customers

# Include only specific models (comma-separated)
# include_models:
#   - customers
#   - orders
#   - products

# Exclude specific models (comma-separated)
# exclude_models:
#   - staging_customers
#   - temp_orders

# Exposure Filtering
# ------------------
# Generate only models referenced in dbt exposures

# exposures_only: false  # Set to true to only generate exposed models
# exposures_tag: ""      # Filter exposures by tag before processing

# Generation Options
# ------------------
# Customize how LookML is generated

# Use BigQuery table name instead of dbt model name for views
# use_table_name: false

# Custom timeframes for date dimensions
# timeframes:
#   - raw_time
#   - time
#   - date
#   - week
#   - month
#   - quarter
#   - year

# String to remove from schema names in output paths
# Useful for removing prefixes like "prod_" or "dbt_"
# remove_schema_string: "prod_"

# Generate all files in output directory without subdirectories
# flatten: false

# Error Handling
# --------------
# Control how errors are handled during generation

# Continue processing if individual models fail
# continue_on_error: false

# Logging
# -------
# Control logging verbosity

# Logging level: DEBUG, INFO, WARN, ERROR
# log_level: INFO

# Reporting
# ---------
# Generate processing reports

# Path to write processing report (JSON format)
# report: processing_report.json

```