# ADR 0002: Parse dbt Artifacts Instead of Querying Database

**Status:** Accepted

**Date:** 2024

## Context

To generate LookML views, we need information about:
- Table schemas (column names, types)
- dbt model metadata (tags, descriptions, custom meta)
- Relationships between models

We could either:
1. Parse dbt's generated artifacts (manifest.json, catalog.json)
2. Query the database directly (BigQuery information_schema)

## Decision

We will parse dbt's manifest.json and catalog.json files instead of querying the database.

## Rationale

**Pros:**
- **No database credentials needed** - Works offline, no security concerns
- **Faster** - No network calls to database
- **Works with dbt's metadata** - Access to tags, descriptions, custom meta
- **Consistent with dbt workflow** - Users already run `dbt docs generate`
- **Version controlled** - Artifacts can be committed to git
- **No database permissions needed** - Works in restricted environments

**Cons:**
- **Requires dbt docs generate** - Extra step in workflow
- **Stale data possible** - If artifacts not regenerated after schema changes

**Alternatives considered:**
- **Query BigQuery directly** - Would require credentials and permissions
- **Hybrid approach** - Complex and adds dependencies

## Consequences

**Positive:**
- Simple deployment - no database drivers needed
- Fast execution - all data is local
- Secure - no credentials to manage
- Works in CI/CD without database access

**Negative:**
- Users must run `dbt docs generate` first
- Artifacts must be kept in sync with database

## Notes

This decision makes dbt2lookml a pure transformation tool that fits naturally into the dbt workflow. Users already run `dbt docs generate` for documentation, so this is a minimal additional burden.
