# ADR 0006: Generate Explores from dbt Relationships

**Status:** Accepted

**Date:** 2024

## Context

LookML explores define how views can be joined together for analysis. We need to decide how to generate explores:
- Manually configured by users
- Automatically from dbt relationships
- Not generated at all

dbt models can define relationships in their YAML:
```yaml
models:
  - name: orders
    columns:
      - name: customer_id
        relationships:
          - to: ref('customers')
            field: customer_id
```

## Decision

We will automatically generate LookML explores based on dbt relationships defined in model YAML files, with the ability to customize via meta tags.

## Rationale

**Pros:**
- **Leverage existing dbt metadata** - Relationships already defined
- **Automatic join generation** - Less manual work for users
- **Consistent with dbt** - Joins match dbt's understanding
- **Reduces duplication** - Don't redefine relationships in LookML
- **Validates relationships** - dbt tests ensure relationships are valid

**Cons:**
- **Limited control** - Auto-generation may not fit all use cases
- **Complex explores** - Some explores need manual tuning
- **dbt relationship limitations** - Not all join types supported

**Alternatives considered:**
- **Manual explore configuration only** - More work, duplication
- **No explores** - Users must create manually in Looker
- **Separate explore config file** - Another file to maintain

## Consequences

**Positive:**
- Quick start for users - explores work out of the box
- Relationships stay in sync between dbt and LookML
- Less configuration needed
- Follows single source of truth principle

**Negative:**
- Complex explores may need manual override
- Users need to understand dbt relationships
- May generate explores that aren't needed

## Implementation

- Parse dbt relationships from model YAML
- Generate explores with appropriate joins
- Allow meta tags to customize join types, relationships
- Support both `ref()` and direct table references
- Generate one explore per base model by default

## Customization Example

```yaml
models:
  - name: orders
    meta:
      looker:
        explore:
          label: "Order Analysis"
          joins:
            - view: customers
              type: left_outer
              relationship: many_to_one
```

## Notes

This approach balances automation with flexibility, making it easy to get started while allowing customization when needed.
