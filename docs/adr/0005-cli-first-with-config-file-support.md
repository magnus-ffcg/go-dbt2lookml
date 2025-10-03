# ADR 0005: CLI-First Design with Configuration File Support

**Status:** Accepted

**Date:** 2024

## Context

Users need to configure how dbt2lookml runs:
- Input file paths (manifest.json, catalog.json)
- Output directory
- Model filtering (tags, includes, excludes)
- Generation options (timeframes, error handling)

We need to decide the primary interface for configuration.

## Decision

We will implement a CLI-first design with optional YAML configuration file support. Command-line flags take precedence over config file settings.

## Rationale

**Pros:**
- **Flexible** - Works for both one-off runs and automated workflows
- **CI/CD friendly** - Easy to override settings in pipelines
- **Discoverable** - `--help` shows all options
- **Scriptable** - Easy to use in shell scripts
- **Config file for complex setups** - Reduces command-line verbosity
- **Standard pattern** - Common in CLI tools

**Cons:**
- **Two ways to configure** - Users need to understand precedence
- **Documentation overhead** - Need to document both methods

**Alternatives considered:**
- **Config file only** - Less flexible for CI/CD
- **CLI only** - Verbose for complex configurations
- **Environment variables** - Hard to discover and document

## Consequences

**Positive:**
- Easy to get started with simple commands
- Powerful for complex configurations
- Works well in automation
- Standard tool behavior

**Negative:**
- Need to maintain both CLI and config parsing
- Precedence rules must be clear

## Precedence Order

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

## Example

```bash
# Simple CLI usage
dbt2lookml --manifest-path target/manifest.json --output-dir lookml/

# With config file
dbt2lookml --config config.yaml

# Override config file setting
dbt2lookml --config config.yaml --continue-on-error
```

## Notes

This design pattern is used by successful tools like kubectl, docker, and terraform.
