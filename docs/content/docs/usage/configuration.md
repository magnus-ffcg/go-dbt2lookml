---
title: Configuration
weight: 20
---

# Configuration Guide

Complete reference for configuring dbt2lookml.

## Configuration Methods

dbt2lookml can be configured in three ways (in order of precedence):

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file** (lowest priority)

### Example

```bash
# Using flags
dbt2lookml --manifest-path target/manifest.json --catalog-path target/catalog.json

# Using config file
dbt2lookml --config config.yaml

# Flags override config file
dbt2lookml --config config.yaml --continue-on-error
```

---

## Configuration File

### Location

Default: `./config.yaml`  
Custom: Use `--config path/to/config.yaml`

### Format

YAML format. See [`example.config.yaml`](../../example.config.yaml) for a complete template.

**Basic example:**

```yaml
manifest_path: target/manifest.json
catalog_path: target/catalog.json
output_dir: lookml/views
tag: looker
continue_on_error: true
```

---

## Configuration Options

### Required Options

#### `manifest_path` (string)

Path to dbt `manifest.json` file.

```yaml
manifest_path: target/manifest.json
```

```bash
--manifest-path target/manifest.json
```

#### `catalog_path` (string)

Path to dbt `catalog.json` file.

```yaml
catalog_path: target/catalog.json
```

```bash
--catalog-path target/catalog.json
```

---

### Output Options

#### `output_dir` (string)

Directory for generated LookML files.

**Default:** `.`

```yaml
output_dir: lookml/views
```

```bash
--output-dir lookml/views
```

**Output structure:**

```
lookml/views/
├── schema_name/
│   ├── model1.view.lkml
│   ├── model2.view.lkml
│   └── ...
└── another_schema/
    └── ...
```

#### `target_dir` (string)

dbt target directory (rarely needs changing).

**Default:** `.`

```yaml
target_dir: .
```

```bash
--target-dir .
```

#### `flatten` (boolean)

Generate all files in output directory without subdirectories.

**Default:** `false`

```yaml
flatten: true
```

```bash
--flatten
```

**With flatten:**

```
lookml/views/
├── model1.view.lkml
├── model2.view.lkml
└── model3.view.lkml
```

---

### Model Filtering

Filter which dbt models get converted to LookML.

#### `tag` (string)

Filter models by dbt tag.

```yaml
tag: looker
```

```bash
--tag looker
```

**Only models with this tag will be processed.**

#### `select` (string)

Select a specific model by name.

```yaml
select: customers
```

```bash
--select customers
```

#### `include_models` (array/string)

Include only specific models.

```yaml
include_models:
  - customers
  - orders
  - products
```

```bash
--include-models customers,orders,products
```

#### `exclude_models` (array/string)

Exclude specific models.

```yaml
exclude_models:
  - staging_customers
  - temp_orders
```

```bash
--exclude-models staging_customers,temp_orders
```

---

### Exposure Filtering

#### `exposures_only` (boolean)

Generate only models referenced in dbt exposures.

**Default:** `false`

```yaml
exposures_only: true
```

```bash
--exposures-only
```

#### `exposures_tag` (string)

Filter exposures by tag before processing.

```yaml
exposures_tag: looker
```

```bash
--exposures-tag looker
```

**Use case:** You have multiple exposure types (dashboards, reports) but only want to generate LookML for those tagged as Looker exposures.

---

### Generation Options

#### `use_table_name` (boolean)

Use BigQuery table name instead of dbt model name for view names.

**Default:** `false`

```yaml
use_table_name: true
```

```bash
--use-table-name
```

**Example:**
- dbt model: `stg_customers`
- BigQuery table: `staging_customers`
- Without flag: `stg_customers.view.lkml`
- With flag: `staging_customers.view.lkml`

#### `generate_locale` (boolean)

Generate locale-specific number formatting in LookML.

**Default:** `false`

```yaml
generate_locale: true
```

```bash
--generate-locale
```

#### `include_iso_fields` (boolean)

Include ISO 8601 formatted date/time fields.

**Default:** `false`

```yaml
include_iso_fields: true
```

```bash
--include-iso-fields
```

Adds dimensions like `created_at_iso` alongside `created_at`.

#### `timeframes` (array/string)

Custom timeframes for date dimensions.

**Default:** `[raw_time, time, date, week, month, quarter, year]`

```yaml
timeframes:
  - date
  - week
  - month
  - quarter
  - year
```

```bash
--timeframes date,week,month,quarter,year
```

#### `remove_schema_string` (string)

String to remove from schema names in output paths.

```yaml
remove_schema_string: "prod_"
```

```bash
--remove-schema-string "prod_"
```

**Example:**
- Schema: `prod_analytics`
- Output: `lookml/views/analytics/` (not `lookml/views/prod_analytics/`)

---

### Error Handling

#### `continue_on_error` (boolean)

Continue processing if individual models fail.

**Default:** `false` (stop on first error)

```yaml
continue_on_error: true
```

```bash
--continue-on-error
```

**Output with this flag:**

```
Warning: 3 models failed to generate:
  - model customer_orders: invalid column type
  - model product_analytics: missing required field
  - model sales_summary: schema mismatch
Files generated: 97/100
```

**See:** [Error Handling Guide](error-handling.md)

---

### Logging

#### `log_level` (string)

Logging verbosity level.

**Options:** `DEBUG`, `INFO`, `WARN`, `ERROR`  
**Default:** `INFO`

```yaml
log_level: DEBUG
```

```bash
--log-level DEBUG
```

**Levels:**
- `DEBUG` - Detailed information for diagnosing issues
- `INFO` - General informational messages
- `WARN` - Warning messages for potential issues
- `ERROR` - Error messages only

---

### Reporting

#### `report` (string)

Path to write processing report (JSON format).

```yaml
report: processing_report.json
```

```bash
--report processing_report.json
```

**Report includes:**
- Models processed
- Files generated
- Errors encountered
- Processing time
- Statistics

---

## Complete Example

```yaml
# dbt artifacts
manifest_path: target/manifest.json
catalog_path: target/catalog.json

# Output
output_dir: lookml/views

# Filtering
tag: looker
exclude_models:
  - staging_*
  - temp_*

# Generation
use_table_name: false
include_iso_fields: true
timeframes:
  - date
  - week
  - month
  - quarter
  - year

# Error handling
continue_on_error: true
log_level: INFO
report: generation_report.json
```

---

## Environment Variables

Use environment variables in your config with `${VAR_NAME}` syntax:

```yaml
manifest_path: ${DBT_TARGET_DIR}/manifest.json
catalog_path: ${DBT_TARGET_DIR}/catalog.json
output_dir: ${LOOKML_OUTPUT_DIR}
```

**Example:**

```bash
export DBT_TARGET_DIR=target
export LOOKML_OUTPUT_DIR=lookml/views
dbt2lookml --config config.yaml
```

---

## Common Scenarios

### Development Environment

```yaml
manifest_path: target/manifest.json
catalog_path: target/catalog.json
output_dir: lookml/views
continue_on_error: true
log_level: DEBUG
report: dev_report.json
```

### Production CI/CD

```yaml
manifest_path: ${DBT_TARGET_DIR}/manifest.json
catalog_path: ${DBT_TARGET_DIR}/catalog.json
output_dir: ${LOOKML_OUTPUT_DIR}
tag: production
log_level: ERROR
```

### Tagged Models Only

```yaml
manifest_path: target/manifest.json
catalog_path: target/catalog.json
output_dir: lookml/views
tag: looker
```

### Exposures Only

```yaml
manifest_path: target/manifest.json
catalog_path: target/catalog.json
output_dir: lookml/views
exposures_only: true
exposures_tag: dashboard
```

---

## Precedence Rules

When the same option is specified in multiple places:

```
Command-line flag  >  Environment variable  >  Config file  >  Default
```

**Example:**

```yaml
# config.yaml
log_level: INFO
```

```bash
export LOG_LEVEL=WARN
dbt2lookml --config config.yaml --log-level DEBUG
```

**Result:** `DEBUG` (flag wins)

---

## Validation

dbt2lookml validates your configuration on startup:

- Required fields present
- File paths exist and are readable
- Values are correct types
- Conflicting options detected

**Invalid configuration:**

```bash
$ dbt2lookml --config config.yaml
Error: manifest_path is required
Error: catalog_path not found: /path/to/catalog.json
```

---

## Tips

1. **Start simple** - Use only required options first
2. **Use config files** - Easier than long command lines
3. **Version control** - Keep config files in git (exclude secrets)
4. **Document scenarios** - Add comments for different use cases
5. **Test locally** - Use `continue_on_error` during development
6. **Enable reporting** - Track what's being generated

---

**Related:**
- [Getting Started](getting-started.md)
- [CLI Reference](cli-reference.md)
- [Error Handling](error-handling.md)
- [Example Config](../../example.config.yaml)
