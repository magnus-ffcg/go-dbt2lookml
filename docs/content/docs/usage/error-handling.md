---
title: Error Handling
weight: 40
---

# Error Handling Guide

Understanding and configuring how dbt2lookml handles errors during generation.

## Error Strategies

dbt2lookml offers flexible error handling through the `--continue-on-error` flag, which maps to different error strategies internally.

### Default Behavior (Fail Fast)

**When to use:** Production environments, CI/CD pipelines

```bash
dbt2lookml --manifest-path ... --catalog-path ...
```

**Behavior:**
- Stops immediately on first error
- Returns error code
- No partial output

**Output:**
```
Generating LookML files...
Error: generation failed on model customer_orders: invalid column type
```

### Continue on Error

**When to use:** Development, debugging, large projects

```bash
dbt2lookml --continue-on-error --manifest-path ... --catalog-path ...
```

**Behavior:**
- Processes all models even if some fail
- Collects all errors
- Reports success with warnings
- Generates files for successful models

**Output:**
```
Generating LookML files...
Warning: 3 models failed to generate:
  - model customer_orders: invalid column type
  - model product_analytics: missing required field
  - model sales_summary: schema mismatch
Files generated: 97/100
Total time: 2.5s
```

## Error Types

### 1. Configuration Errors

**Cause:** Invalid configuration or missing required parameters

**Examples:**
- Missing manifest or catalog path
- Invalid output directory
- Conflicting options

**Solution:**
```bash
# Check your configuration
dbt2lookml --config config.yaml

# Validate paths exist
ls -l target/manifest.json target/catalog.json
```

### 2. Parsing Errors

**Cause:** Invalid or corrupted dbt artifacts

**Examples:**
- Malformed JSON
- Missing required fields
- Incompatible dbt version

**Solution:**
```bash
# Regenerate dbt artifacts
dbt docs generate

# Verify JSON is valid
cat target/manifest.json | jq . > /dev/null
```

### 3. Generation Errors

**Cause:** Model-specific issues during LookML generation

**Examples:**
- Unsupported column type
- Missing metadata
- Invalid nested structure (4+ levels)

**Solution:**
- Fix model definition in dbt
- Add missing metadata
- Simplify nested structures
- Use `--continue-on-error` to generate other models

### 4. File System Errors

**Cause:** Permission or file system issues

**Examples:**
- No write permission to output directory
- Disk full
- Invalid characters in file names

**Solution:**
```bash
# Check permissions
ls -ld lookml/views/

# Create output directory if needed
mkdir -p lookml/views

# Check disk space
df -h
```

## Understanding Error Messages

### Per-Model Errors

When using `--continue-on-error`, you'll see detailed per-model errors:

```
Warning: 3 models failed to generate:
  - model customer_orders: column 'nested.deep.field' exceeds maximum nesting depth (level 4)
  - model product_analytics: missing required field 'name'
  - model sales_summary: STRUCT parent 'address' has no children
```

**Each error includes:**
1. Model name that failed
2. Specific reason for failure
3. Actionable information

### Progress Reporting

```
Files generated: 97/100
```

**Interpretation:**
- **97** models successfully generated LookML
- **100** models attempted
- **3** models failed (100 - 97)

## Best Practices

### 1. Development Workflow

Use `--continue-on-error` during development:

```bash
#!/bin/bash
# dev-generate.sh
dbt2lookml \
  --continue-on-error \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views
```

### 2. CI/CD Pipeline

Use fail-fast (default) in CI/CD:

```yaml
# .github/workflows/generate-lookml.yml
- name: Generate LookML
  run: |
    dbt2lookml \
      --manifest-path target/manifest.json \
      --catalog-path target/catalog.json \
      --output-dir lookml/views
  # Fails the pipeline if any model fails
```

### 3. Monitoring Production

Log errors for monitoring:

```bash
dbt2lookml \
  --continue-on-error \
  --manifest-path ... \
  --catalog-path ... \
  2>&1 | tee generation.log

# Check for failures
if grep -q "failed to generate" generation.log; then
  echo "⚠️  Some models failed, check generation.log"
  exit 1
fi
```

### 4. Debugging Failures

When models fail:

```bash
# 1. Isolate the failing model
dbt2lookml --select failing_model --manifest-path ... --catalog-path ...

# 2. Check model definition in dbt
cat models/path/to/failing_model.sql

# 3. Verify column types in catalog
cat target/catalog.json | jq '.nodes["model.project.failing_model"]'

# 4. Fix the issue and regenerate
dbt docs generate
dbt2lookml --select failing_model --manifest-path ... --catalog-path ...
```

## Common Errors and Solutions

### Error: "exceeds maximum nesting depth (level 4)"

**Cause:** Array nested more than 3 levels deep

**Solution:**
```sql
-- Simplify your dbt model to reduce nesting
-- Or flatten the structure in your SQL
```

### Error: "missing required field 'name'"

**Cause:** Model missing required metadata

**Solution:**
```yaml
# models/schema.yml
models:
  - name: my_model
    description: "Model description"  # Required
```

### Error: "STRUCT parent has no children"

**Cause:** STRUCT type without nested fields

**Solution:**
- Check BigQuery schema
- Ensure STRUCT has nested fields
- Regenerate dbt catalog: `dbt docs generate`

### Error: "dimension name conflict"

**Cause:** Dimension and dimension_group have the same name

**Solution:**
Don't worry! This is automatically handled:
- Conflicting dimensions are renamed with `_conflict` suffix
- They are hidden by default
- Dimension groups take precedence

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success (all models generated) |
| 1 | Error (configuration, parsing, or generation failure) |

**Note:** With `--continue-on-error`, exit code is 0 even if some models fail. Check the warning message for failures.

## Programmatic Error Handling

If using dbt2lookml as a library:

```go
result, err := generator.GenerateAllWithOptions(ctx, models, opts)

if err != nil {
    // Handle error
}

// Check for partial success
fmt.Printf("Generated: %d/%d\n", result.FilesGenerated, result.ModelsProcessed)

// Report per-model errors
if result.HasErrors() {
    for _, modelErr := range result.Errors {
        log.Printf("Failed: %s\n", modelErr.String())
    }
}
```

## Getting Help

If you encounter errors not covered here:

1. **Check logs** for detailed error messages
2. **Search issues** on GitHub
3. **Ask in discussions** or create an issue
4. **Include:**
   - Full error message
   - dbt2lookml version
   - dbt version
   - Relevant model definition

---

**Related:**
- [CLI Reference](cli-reference.md)
- [Configuration Guide](configuration.md)
- [Troubleshooting](troubleshooting.md)
