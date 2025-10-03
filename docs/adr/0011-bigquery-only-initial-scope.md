# ADR 0011: BigQuery-Only Initial Scope

**Status:** Accepted

**Date:** 2024

## Context

dbt supports multiple data warehouses:
- BigQuery
- Snowflake
- Redshift
- Databricks
- PostgreSQL
- Others

We need to decide whether to support:
- All dbt-supported warehouses
- Multiple warehouses (selected subset)
- Single warehouse (BigQuery)

## Decision

We will initially support BigQuery only, with architecture designed to support other warehouses in the future.

## Rationale

**Pros:**
- **Focused scope** - Ship faster with quality BigQuery support
- **Complex type support** - BigQuery's ARRAY/STRUCT are unique challenges
- **Clear use case** - Team's immediate need is BigQuery
- **Quality over breadth** - Better to excel at one than be mediocre at many
- **Learn from usage** - Understand patterns before expanding
- **Extensible design** - Architecture allows future warehouse support

**Cons:**
- **Limited audience** - Excludes Snowflake, Redshift users
- **Missed opportunities** - Other warehouse users might contribute
- **Rework needed** - Adding warehouses later requires changes

**Alternatives considered:**
- **All warehouses from start** - Too ambitious, delays shipping
- **Multi-warehouse from start** - Increases complexity significantly
- **Warehouse-agnostic** - Misses BigQuery-specific optimizations

## Consequences

**Positive:**
- Ship working product faster
- Deep BigQuery expertise
- Excellent nested type support
- Clear feature set
- Maintainable codebase

**Negative:**
- Limited to BigQuery users initially
- Marketing limited to BigQuery community
- Feature requests for other warehouses

## Future Expansion Path

Architecture supports future warehouses through:
- Abstraction layer for type mapping
- Warehouse-specific generators
- Pluggable type converters
- Configuration for warehouse selection

Example future usage:
```bash
dbt2lookml --warehouse snowflake --config config.yaml
```

## Type Mapping Abstraction

Design allows warehouse-specific type mapping:
```go
type TypeMapper interface {
    MapToDimensionType(warehouseType string) string
}

type BigQueryTypeMapper struct {}
type SnowflakeTypeMapper struct {}  // Future
```

## Notes

This "start small, expand later" approach balances shipping value quickly with maintaining flexibility for future growth. BigQuery's unique features (nested types) make it a good proving ground.
