# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for go-dbt2lookml.

## What is an ADR?

An Architecture Decision Record (ADR) captures an important architectural decision made along with its context and consequences.

## Format

Each ADR follows this structure:
- **Status** - Proposed, Accepted, Deprecated, Superseded
- **Date** - When the decision was made
- **Context** - The issue motivating this decision
- **Decision** - The change being proposed or made
- **Rationale** - Why this decision was made (pros/cons)
- **Consequences** - What becomes easier or harder as a result

## Index

- [ADR 0001](0001-use-go-for-implementation.md) - Use Go for Implementation
- [ADR 0002](0002-parse-dbt-artifacts-not-database.md) - Parse dbt Artifacts Instead of Querying Database
- [ADR 0003](0003-support-bigquery-nested-types.md) - Support BigQuery Nested Types (ARRAY and STRUCT)
- [ADR 0004](0004-use-dbt-meta-for-lookml-config.md) - Use dbt Meta Tags for LookML Configuration
- [ADR 0005](0005-cli-first-with-config-file-support.md) - CLI-First Design with Configuration File Support
- [ADR 0006](0006-generate-explores-from-dbt-relationships.md) - Generate Explores from dbt Relationships
- [ADR 0007](0007-use-cobra-for-cli.md) - Use Cobra for CLI Framework
- [ADR 0008](0008-one-file-per-view-output.md) - Generate One File Per LookML View
- [ADR 0009](0009-continue-on-error-strategy.md) - Continue-on-Error Strategy for Partial Success
- [ADR 0010](0010-testing-strategy.md) - Testing Strategy with Unit and Integration Tests
- [ADR 0011](0011-bigquery-only-initial-scope.md) - BigQuery-Only Initial Scope
- [ADR 0012](0012-auto-generate-default-count-measure.md) - Auto-Generate Default Count Measure
- [ADR 0013](0013-tag-based-model-filtering.md) - Tag-Based Model Filtering
- [ADR 0014](0014-support-dbt-semantic-models.md) - Support dbt Semantic Models for Measure Generation

## Creating New ADRs

When making significant architectural decisions:

1. Copy the template structure
2. Number sequentially (0015, 0016, etc.)
3. Use descriptive filename: `NNNN-brief-description.md`
4. Fill in all sections
5. Update this README index
6. Commit with the code changes

## References

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
