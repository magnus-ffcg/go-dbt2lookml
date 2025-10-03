# ADR 0004: Use dbt Meta Tags for LookML Configuration

**Status:** Accepted

**Date:** 2024

## Context

We need a way for users to customize LookML generation:
- Specify dimension types (string, number, time, etc.)
- Add labels and descriptions
- Configure measures (sum, count, average, etc.)
- Define relationships for explores
- Hide/show fields

We could use:
1. dbt's `meta` tags in model YAML files
2. Separate configuration file
3. Command-line arguments
4. Annotations in SQL comments

## Decision

We will use dbt's `meta` tags under a `looker` namespace for LookML-specific configuration.

## Rationale

**Pros:**
- **Single source of truth** - Configuration lives with dbt models
- **Version controlled** - Changes tracked in git with models
- **Familiar to dbt users** - Uses existing dbt patterns
- **Type-safe** - YAML validation in dbt
- **Model-specific** - Each model can have unique configuration
- **No extra files** - Leverages existing dbt YAML files

**Cons:**
- **Couples to dbt** - Configuration not portable outside dbt
- **YAML verbosity** - Can be verbose for complex configurations

**Alternatives considered:**
- **Separate config file** - Extra file to maintain, can get out of sync
- **SQL comments** - Hard to parse, not structured
- **CLI arguments** - Not scalable for many models

## Consequences

**Positive:**
- Natural fit with dbt workflow
- Easy to understand for dbt users
- Configuration changes reviewed with code
- No separate config files to maintain

**Negative:**
- dbt YAML files can become large
- Requires understanding of both dbt and LookML concepts

## Example

```yaml
models:
  - name: customers
    meta:
      looker:
        label: "Customer Information"
        dimensions:
          - name: customer_id
            type: number
            primary_key: true
        measures:
          - name: total_customers
            type: count
```

## Notes

This approach follows the principle of keeping related configuration together and leveraging existing dbt infrastructure.
