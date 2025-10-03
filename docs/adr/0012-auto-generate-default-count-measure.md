# ADR 0012: Auto-Generate Default Count Measure

**Status:** Accepted

**Date:** 2024

## Context

Every LookML view should have at least one measure for basic aggregation. Users can define custom measures in dbt meta, but many views might not have any measures defined.

We need to decide:
- Require users to define all measures
- Auto-generate a basic count measure
- Generate multiple default measures (count, sum, avg, etc.)
- Make it optional via configuration

## Decision

We will automatically generate a default `count` measure for every view that doesn't already have one defined, with an option to disable this behavior.

## Rationale

**Pros:**
- **Usability** - Views work immediately in Looker
- **Common pattern** - Count is the most common measure
- **Minimal** - Just one measure, not overwhelming
- **Overridable** - Users can define their own count if needed
- **Looker best practice** - Every view should have measures
- **Quick start** - Users can explore data right away

**Cons:**
- **Opinionated** - Assumes users want this
- **Extra code** - Generates measure even if not needed
- **Namespace pollution** - Adds measure user might not want

**Alternatives considered:**
- **No default measures** - Views would be dimension-only
- **Multiple default measures** - Too many unused measures
- **Required user configuration** - More work for users
- **Prompt user** - Not suitable for CLI tool

## Consequences

**Positive:**
- Views are immediately useful in Looker
- Users can count records without configuration
- Follows Looker conventions
- Reduces initial configuration burden

**Negative:**
- Views might have unused count measure
- Users must know they can override it
- Slightly larger generated files

## Implementation

Default count measure:
```lkml
measure: count {
  type: count
  drill_fields: [detail*]
}
```

Only generated if:
- No measures defined in dbt meta
- No existing measure with type `count`

Can be disabled:
```yaml
# config.yaml
generate_default_count: false
```

## Override Example

Users can define their own count measure in dbt meta:
```yaml
models:
  - name: customers
    meta:
      looker:
        measures:
          - name: customer_count
            type: count
            label: "Number of Customers"
```

## Notes

This decision prioritizes user experience and follows the principle of "convention over configuration." Users get working views immediately while retaining full control through dbt meta configuration.
