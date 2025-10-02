---
title: CLI Reference
weight: 30
---

# CLI Reference

Complete command-line reference for dbt2lookml.

## Synopsis

```
dbt2lookml [flags]
dbt2lookml [command]
```

## Description

dbt2lookml generates LookML views from BigQuery via dbt models. It parses dbt manifest and catalog files to create comprehensive LookML views with dimensions, measures, and explores for use in Looker.

## Quick Start

```bash
# Basic usage
dbt2lookml --manifest-path target/manifest.json --catalog-path target/catalog.json --output-dir lookml/views

# Using config file
dbt2lookml --config config.yaml

# With filtering
dbt2lookml --tag looker --config config.yaml
```

---

## Commands

### `dbt2lookml`

Main command - generates LookML files.

### `dbt2lookml version`

Show version information.

```bash
$ dbt2lookml version
dbt2lookml version v1.0.0
  commit: abc123
  built: 2025-10-01
  go: go1.23.0
  platform: darwin/arm64
```

### `dbt2lookml completion`

Generate shell completion scripts.

```bash
# Bash
dbt2lookml completion bash > /etc/bash_completion.d/dbt2lookml

# Zsh
dbt2lookml completion zsh > "${fpath[1]}/_dbt2lookml"

# Fish
dbt2lookml completion fish > ~/.config/fish/completions/dbt2lookml.fish
```

---

## Global Flags

### `--config` (string)

Path to configuration file.

**Default:** `./config.yaml`

```bash
dbt2lookml --config config.yaml
dbt2lookml --config /path/to/custom-config.yaml
```

---

## Core Flags

### `--manifest-path` (string) **[Required]**

Path to dbt `manifest.json` file.

```bash
--manifest-path target/manifest.json
--manifest-path /full/path/to/manifest.json
```

### `--catalog-path` (string) **[Required]**

Path to dbt `catalog.json` file.

```bash
--catalog-path target/catalog.json
--catalog-path /full/path/to/catalog.json
```

### `--target-dir` (string)

dbt target directory.

**Default:** `.`

```bash
--target-dir target
--target-dir /path/to/dbt/target
```

### `--output-dir` (string)

Output directory for generated LookML files.

**Default:** `.`

```bash
--output-dir lookml/views
--output-dir /path/to/lookml
```

---

## Model Filtering Flags

### `--tag` (string)

Filter models by dbt tag.

```bash
--tag looker
--tag production
```

### `--select` (string)

Select a specific model by name.

```bash
--select customers
--select mart_customers
```

### `--include-models` (comma-separated)

Include only specific models.

```bash
--include-models customers,orders,products
--include-models "customer_*,order_*"
```

### `--exclude-models` (comma-separated)

Exclude specific models.

```bash
--exclude-models staging_*,temp_*
--exclude-models test_model,debug_model
```

---

## Exposure Filtering Flags

### `--exposures-only`

Generate only models referenced in dbt exposures.

```bash
--exposures-only
```

### `--exposures-tag` (string)

Filter exposures by tag.

```bash
--exposures-tag looker
--exposures-tag dashboard
```

---

## Generation Options Flags

### `--use-table-name`

Use BigQuery table name instead of dbt model name.

```bash
--use-table-name
```

### `--generate-locale`

Generate locale-specific number formatting.

```bash
--generate-locale
```

### `--include-iso-fields`

Include ISO 8601 formatted date/time fields.

```bash
--include-iso-fields
```

### `--timeframes` (comma-separated)

Custom timeframes for date dimensions.

```bash
--timeframes date,week,month,quarter,year
--timeframes day,month,quarter
```

### `--remove-schema-string` (string)

String to remove from schema names.

```bash
--remove-schema-string "prod_"
--remove-schema-string "staging_"
```

### `--flatten`

Generate all files in output directory without subdirectories.

```bash
--flatten
```

---

## Error Handling & Logging Flags

### `--continue-on-error`

Continue processing if models fail.

```bash
--continue-on-error
```

### `--log-level` (string)

Logging level.

**Options:** `DEBUG`, `INFO`, `WARN`, `ERROR`  
**Default:** `INFO`

```bash
--log-level DEBUG
--log-level ERROR
```

### `--report` (string)

Path to write processing report (JSON).

```bash
--report report.json
--report /path/to/generation-report.json
```

---

## Examples

### Basic Usage

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views
```

### Using Config File

```bash
dbt2lookml --config config.yaml
```

### Filter by Tag

```bash
dbt2lookml \
  --config config.yaml \
  --tag looker
```

### Continue on Errors

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --continue-on-error \
  --report errors.json
```

### Exposures Only

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --exposures-only \
  --exposures-tag dashboard
```

### Custom Timeframes

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --timeframes date,week,month,quarter,year \
  --include-iso-fields
```

### Exclude Models

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --exclude-models staging_*,temp_*,test_*
```

### Debug Mode

```bash
dbt2lookml \
  --config config.yaml \
  --log-level DEBUG \
  --report debug-report.json
```

### Flat Output Structure

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views \
  --flatten
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success - all models generated |
| 1 | Error - configuration, parsing, or generation failure |

**Note:** With `--continue-on-error`, exit code is 0 even if some models fail. Check warning messages or report file for failures.

---

## Environment Variables

dbt2lookml respects standard Go environment variables:

- `GOOS` - Target operating system
- `GOARCH` - Target architecture  
- `TMPDIR` - Temporary directory

---

## Shell Completion

Enable tab completion for faster CLI usage:

### Bash

```bash
# Install
dbt2lookml completion bash > /etc/bash_completion.d/dbt2lookml

# Or for current user
dbt2lookml completion bash > ~/.bash_completion

# Reload
source ~/.bashrc
```

### Zsh

```bash
# Install
dbt2lookml completion zsh > "${fpath[1]}/_dbt2lookml"

# Reload
autoload -U compinit && compinit
```

### Fish

```bash
# Install
dbt2lookml completion fish > ~/.config/fish/completions/dbt2lookml.fish

# Reload
source ~/.config/fish/config.fish
```

---

## Tips & Tricks

### 1. Use Config Files for Consistency

```bash
# Save time with repeated runs
dbt2lookml --config production.yaml
dbt2lookml --config development.yaml
```

### 2. Combine Flags and Config

```bash
# Override specific options
dbt2lookml --config base.yaml --tag looker --log-level DEBUG
```

### 3. Use Shell Variables

```bash
DBT_DIR=target
OUTPUT_DIR=lookml/views

dbt2lookml \
  --manifest-path $DBT_DIR/manifest.json \
  --catalog-path $DBT_DIR/catalog.json \
  --output-dir $OUTPUT_DIR
```

### 4. Create Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias d2l='dbt2lookml --config config.yaml'
alias d2l-dev='dbt2lookml --config dev-config.yaml --log-level DEBUG'

# Usage
d2l
d2l-dev
```

### 5. Check What Will Be Generated

```bash
# Use dry-run concept with report
dbt2lookml --config config.yaml --report preview.json
cat preview.json | jq '.models_processed'
```

### 6. Debug Specific Model

```bash
dbt2lookml \
  --config config.yaml \
  --select problematic_model \
  --log-level DEBUG
```

### 7. Incremental Development

```bash
# Start with small set
dbt2lookml --include-models customer,order --log-level DEBUG

# Expand once working
dbt2lookml --tag looker
```

---

## Common Patterns

### CI/CD Pipeline

```bash
#!/bin/bash
set -e

# Generate dbt artifacts
dbt docs generate

# Generate LookML
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views \
  --tag production \
  --log-level ERROR

# Check for generated files
ls -l lookml/views/
```

### Development Workflow

```bash
#!/bin/bash

# Regenerate with error tolerance
dbt2lookml \
  --config dev-config.yaml \
  --continue-on-error \
  --log-level DEBUG \
  --report dev-report.json

# Show errors
cat dev-report.json | jq '.errors'
```

### Tagged Release

```bash
#!/bin/bash

VERSION=$(dbt2lookml version | grep version | awk '{print $3}')

dbt2lookml \
  --config config.yaml \
  --tag release-$VERSION \
  --report release-report.json
```

---

## Troubleshooting

### Command Not Found

```bash
# Check installation
which dbt2lookml

# Add to PATH
export PATH=$PATH:/path/to/dbt2lookml
```

### Permission Denied

```bash
# Make executable
chmod +x /path/to/dbt2lookml

# Or use go run
go run ./cmd/dbt2lookml --help
```

### Config Not Found

```bash
# Specify full path
dbt2lookml --config /full/path/to/config.yaml

# Check current directory
ls -la config.yaml
```

---

**Related:**
- [Getting Started](getting-started.md)
- [Configuration Guide](configuration.md)
- [Error Handling](error-handling.md)
- [Example Config](../../example.config.yaml)
