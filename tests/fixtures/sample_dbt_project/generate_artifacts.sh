#!/bin/bash

# Script to generate dbt artifacts (manifest.json and catalog.json)
# Uses dbt-duckdb with in-memory database - no external dependencies needed!

set -e

echo "🔧 Generating dbt artifacts..."
echo ""

# Check if dbt is installed
if ! command -v dbt &> /dev/null; then
    echo "❌ dbt is not installed"
    echo "Install with: pip install dbt-duckdb"
    exit 1
fi

# Check if dbt-duckdb is available
if ! dbt --version | grep -q "duckdb"; then
    echo "⚠️  dbt-duckdb adapter not found"
    echo "Install with: pip install dbt-duckdb"
    exit 1
fi

echo "✅ dbt-duckdb found"
echo ""

# Set profiles directory to current directory
export DBT_PROFILES_DIR=$(pwd)

# Clean previous artifacts
echo "🧹 Cleaning previous artifacts..."
rm -rf target/
echo ""

# Run dbt to create tables in in-memory DuckDB
echo "📊 Running dbt (creates in-memory tables)..."
dbt run --profiles-dir .
echo ""

# Generate documentation (creates both manifest.json and catalog.json)
echo "📝 Generating documentation..."
dbt docs generate --profiles-dir .
echo ""

# Copy artifacts to test data directory
echo "📋 Copying artifacts to test data directory..."
cp target/manifest.json ../data/manifest_semantic_generated.json
cp target/catalog.json ../data/catalog_semantic_generated.json
echo ""

echo "✅ Artifacts generated successfully!"
echo ""
echo "Generated files:"
echo "  - ../data/manifest_semantic_generated.json"
echo "  - ../data/catalog_semantic_generated.json"
echo ""
echo "💡 You can now use these files for integration testing"
