# ADR 0003: Support BigQuery Nested Types (ARRAY and STRUCT)

**Status:** Accepted

**Date:** 2024

## Context

BigQuery supports complex nested data types:
- **ARRAY** - Repeated fields
- **STRUCT** - Nested records with multiple fields
- **ARRAY<STRUCT>** - Repeated nested records

LookML has specific syntax for handling nested fields. We need to decide how to handle these in the conversion.

## Decision

We will fully support BigQuery's ARRAY and STRUCT types by generating appropriate LookML dimensions with proper SQL references.

## Rationale

**Pros:**
- **Complete BigQuery support** - Handles real-world BigQuery schemas
- **Preserves data structure** - Users can query nested fields in Looker
- **Competitive advantage** - Many tools don't handle nested types well
- **Matches BigQuery best practices** - Nested types are common in BigQuery

**Cons:**
- **Implementation complexity** - Requires recursive parsing and generation
- **LookML complexity** - Generated views can be large with many nested fields
- **Testing complexity** - Need comprehensive test cases for nested structures

**Alternatives considered:**
- **Flatten all nested types** - Loses structure and creates many columns
- **Ignore nested types** - Incomplete solution for BigQuery users
- **Manual configuration only** - Puts burden on users

## Consequences

**Positive:**
- Works with real-world BigQuery schemas
- Users can explore nested data in Looker
- Handles complex analytics use cases
- Differentiates from simpler tools

**Negative:**
- More complex codebase
- Generated LookML files can be large
- Requires good documentation for users

## Implementation Notes

- Use recursive parsing for nested structures
- Generate proper SQL paths (e.g., `${TABLE}.user.address.city`)
- Support UNNEST for ARRAY types
- Preserve field descriptions through nesting levels
