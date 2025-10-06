# Sample DBT Project for Testing

This is a minimal dbt project used to generate **real dbt artifacts** for testing go-dbt2lookml, especially semantic model support.

## Purpose

- Generate authentic `manifest.json` and `catalog.json` files
- Test semantic models integration
- Provide reproducible test fixtures
- Document example usage

## Features

✅ **No external database required** - Uses DuckDB in-memory  
✅ **Semantic models included** - Tests semantic layer integration  
✅ **Simple setup** - One command to generate artifacts  
✅ **Reproducible** - Anyone can regenerate test data

## Prerequisites

```bash
pip install dbt-duckdb
```

That's it! No database setup needed.

## Quick Start

### Generate Artifacts

```bash
cd tests/fixtures/sample_dbt_project
chmod +x generate_artifacts.sh
./generate_artifacts.sh
```

This will:
1. ✅ Run dbt to create in-memory tables
2. ✅ Generate `manifest.json` (includes semantic models)
3. ✅ Generate `catalog.json` (includes column types)
4. ✅ Copy artifacts to `../data/` for testing

### Generated Files

```
../data/
├── manifest_semantic_generated.json  ← Real manifest with semantic models
└── catalog_semantic_generated.json   ← Real catalog with column metadata
```

## Project Structure

```
sample_dbt_project/
├── dbt_project.yml          # Project configuration
├── profiles.yml             # DuckDB in-memory profile
├── models/
│   ├── orders.sql          # Sample fact table
│   └── schema.yml          # Model + semantic model definitions
├── generate_artifacts.sh   # Script to generate artifacts
└── README.md
```

## Semantic Models Included

The `orders` model includes a semantic model with:

**Measures:**
- `total_revenue` - Sum aggregation
- `order_count` - Count aggregation
- `avg_order_value` - Average aggregation
- `unique_customers` - Count distinct aggregation

**Dimensions:**
- `order_date` - Time dimension
- `status` - Categorical dimension

**Entities:**
- `order` - Primary entity
- `customer` - Foreign entity

## Manual Generation

If you prefer to run commands manually:

```bash
# Set profiles directory
export DBT_PROFILES_DIR=$(pwd)

# Run models
dbt run --profiles-dir .

# Generate documentation
dbt docs generate --profiles-dir .

# Artifacts are in target/manifest.json and target/catalog.json
```

## Testing with go-dbt2lookml

After generating artifacts, test semantic models:

```bash
cd ../../..  # Back to repo root
go run cmd/dbt2lookml/main.go \
  --manifest-path tests/fixtures/data/manifest_semantic_generated.json \
  --catalog-path tests/fixtures/data/catalog_semantic_generated.json \
  --output-dir output_test \
  --use-semantic-models
```

## Updating Test Data

When dbt releases new versions or semantic model schema changes:

1. Update `models/schema.yml` with new semantic model features
2. Run `./generate_artifacts.sh`
3. Review generated artifacts
4. Commit updated test fixtures

## Notes

- **DuckDB in-memory** - No persistent database files created
- **Deterministic data** - Sample data is hardcoded in SQL
- **Version testing** - Can test against different dbt versions
- **CI/CD friendly** - No credentials or external services needed

## References

- [dbt Semantic Layer](https://docs.getdbt.com/docs/build/about-metricflow)
- [dbt-duckdb adapter](https://github.com/duckdb/dbt-duckdb)
- [Semantic Models](https://docs.getdbt.com/docs/build/semantic-models)
