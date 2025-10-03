# ADR 0013: Tag-Based Model Filtering

**Status:** Accepted

**Date:** 2024

## Context

dbt projects often contain many models:
- Staging models (raw data transformations)
- Intermediate models (business logic)
- Mart models (final analytics-ready tables)
- Test/temporary models

Not all dbt models should become LookML views. We need a way to filter which models get converted.

Options:
- Convert all models
- Manual list of models to include/exclude
- Tag-based filtering
- Folder-based filtering
- Exposure-based filtering

## Decision

We will support multiple filtering approaches with tag-based filtering as the primary recommended pattern:
- `--tag` flag to filter by dbt tags
- `--include-models` for explicit inclusion
- `--exclude-models` for explicit exclusion
- `--exposures-only` to generate only models referenced in dbt exposures

## Rationale

**Pros:**
- **Flexible** - Multiple filtering methods for different use cases
- **dbt-native** - Uses existing dbt tags
- **Maintainable** - Tags managed in dbt YAML
- **Clear intent** - Tag like `looker` signals purpose
- **No duplication** - Reuses dbt metadata
- **Selective conversion** - Convert only what's needed

**Cons:**
- **Multiple options** - Users need to choose approach
- **Requires planning** - Need to tag models appropriately
- **Documentation needed** - Explain filtering strategies

**Alternatives considered:**
- **All models** - Creates too many unnecessary views
- **Folder-based only** - Less flexible than tags
- **Separate config file** - Duplication with dbt

## Consequences

**Positive:**
- Users control which models become views
- Follows dbt conventions
- Easy to maintain model lists
- Supports different workflow patterns
- Prevents view explosion

**Negative:**
- Requires understanding filtering options
- Need to tag models in dbt
- Multiple ways to do same thing

## Filtering Strategies

### Strategy 1: Tag-Based (Recommended)

Tag models in dbt:
```yaml
models:
  - name: customers
    tags: [looker]
  - name: orders
    tags: [looker, analytics]
```

Use tag flag:
```bash
dbt2lookml --tag looker
```

### Strategy 2: Explicit Lists

Include specific models:
```yaml
include_models:
  - customers
  - orders
  - products
```

Or exclude models:
```yaml
exclude_models:
  - staging_*
  - temp_*
```

### Strategy 3: Exposures-Only

Define exposures in dbt:
```yaml
exposures:
  - name: customer_dashboard
    type: dashboard
    depends_on:
      - ref('customers')
      - ref('orders')
```

Generate only those:
```bash
dbt2lookml --exposures-only
```

## Combining Filters

Filters can be combined:
```bash
# Only analytics-tagged models, excluding test models
dbt2lookml --tag analytics --exclude-models "test_*"
```

Filter precedence:
1. Explicit exclusions (highest)
2. Explicit inclusions
3. Tag filters
4. Exposure filters (lowest)

## Best Practices

**For production:**
- Use tags like `looker` or `analytics`
- Tag only mart/final models
- Document tagging convention

**For development:**
- Use `--select` for single model testing
- Use `--exposures-only` for dashboard-specific views

**For large projects:**
- Combine tag + exclude patterns
- Use exposures to manage model sets

## Notes

This multi-pronged approach provides flexibility while maintaining simplicity. The tag-based approach aligns with dbt best practices and scales well for large projects.
