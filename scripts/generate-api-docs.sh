#!/bin/bash
set -e

# Generate API documentation from Go packages
echo "🔍 Generating API documentation from Go packages..."

# Create API docs directory
mkdir -p docs/content/docs/api

# Generate API index
cat > docs/content/docs/api/_index.md << 'EOF'
---
title: API Reference
weight: 50
bookCollapseSection: true
---

# API Reference

Complete Go package documentation generated from source code.

## Packages

- **[models](models)** - Core data models for dbt and LookML
- **[parsers](parsers)** - Parsing dbt manifest and catalog files  
- **[generators](generators)** - LookML generation from dbt models
- **[enums](enums)** - Enumeration types and constants
- **[utils](utils)** - Utility functions

---

*Documentation generated from Go source code using godoc comments.*
EOF

# Generate documentation for each package
echo "📦 Generating docs for pkg/models..."
gomarkdoc --output docs/content/docs/api/models.md ./pkg/models

echo "📦 Generating docs for pkg/parsers..."
gomarkdoc --output docs/content/docs/api/parsers.md ./pkg/parsers

echo "📦 Generating docs for pkg/generators..."
gomarkdoc --output docs/content/docs/api/generators.md ./pkg/generators

echo "📦 Generating docs for pkg/enums..."
gomarkdoc --output docs/content/docs/api/enums.md ./pkg/enums

echo "📦 Generating docs for pkg/utils..."
gomarkdoc --output docs/content/docs/api/utils.md ./pkg/utils

echo "✅ API documentation generated successfully!"
