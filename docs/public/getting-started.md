# Getting Started with dbt2lookml

Generate LookML views from your dbt models with BigQuery support for complex nested structures.

## Prerequisites

- **dbt project** with BigQuery models
- **dbt artifacts:** `manifest.json` and `catalog.json`
- **Go 1.21+** (if building from source)

## Installation

### Option 1: Pre-built Binaries (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/magnus-ffcg/go-dbt2lookml/releases):

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/magnus-ffcg/go-dbt2lookml/releases/latest/download/dbt2lookml_Darwin_arm64.tar.gz | tar xz
sudo mv dbt2lookml /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/magnus-ffcg/go-dbt2lookml/releases/latest/download/dbt2lookml_Darwin_x86_64.tar.gz | tar xz
sudo mv dbt2lookml /usr/local/bin/
```

**Linux (amd64):**
```bash
curl -L https://github.com/magnus-ffcg/go-dbt2lookml/releases/latest/download/dbt2lookml_Linux_x86_64.tar.gz | tar xz
sudo mv dbt2lookml /usr/local/bin/
```

**Windows (amd64):**
Download `dbt2lookml_Windows_x86_64.zip` from releases and extract to your PATH.

### Option 2: Install via Go

```bash
go install github.com/magnus-ffcg/go-dbt2lookml/cmd/dbt2lookml@latest
```

### Verify Installation

```bash
dbt2lookml version
```

## Quick Start

### 1. Generate dbt Artifacts

First, ensure you have the required dbt artifacts:

```bash
cd your-dbt-project
dbt docs generate
```

This creates:
- `target/manifest.json` - dbt model definitions
- `target/catalog.json` - BigQuery schema information

### 2. Basic Usage

Generate LookML views from all models:

```bash
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views
```

### 3. Check Output

Your LookML files are now in `lookml/views/`:

```
lookml/views/
├── schema_name/
│   ├── model1.view.lkml
│   ├── model2.view.lkml
│   └── ...
└── another_schema/
    └── ...
```

## Common Use Cases

### Generate for Specific Models

```bash
# By tag
dbt2lookml --tag looker --manifest-path ... --catalog-path ...

# By model name
dbt2lookml --select my_model --manifest-path ... --catalog-path ...

# Multiple models
dbt2lookml --include-models model1,model2,model3 --manifest-path ...
```

### Use Configuration File

Create `config.yaml`:

```yaml
manifest_path: target/manifest.json
catalog_path: target/catalog.json
output_dir: lookml/views
tag: looker
use_table_name: false
continue_on_error: true
```

Run with config:

```bash
dbt2lookml --config config.yaml
```

### Handle Errors Gracefully

```bash
# Continue even if some models fail
dbt2lookml --continue-on-error --manifest-path ... --catalog-path ...

# The tool will report which models failed and continue processing others
```

## Features

### ✅ Supported BigQuery Types

- **Primitives:** INT64, FLOAT64, STRING, BOOL, DATE, DATETIME, TIMESTAMP
- **Nested:** STRUCT (nested objects)
- **Arrays:** ARRAY (repeated fields)
- **Complex:** ARRAY<STRUCT> (repeated nested objects)

### ✅ Generated LookML Elements

- **Views:** One per dbt model
- **Dimensions:** All column types with appropriate types
- **Dimension Groups:** Automatic for DATE/DATETIME/TIMESTAMP columns
- **Measures:** Count measure + custom measures from dbt meta
- **Explores:** If defined in dbt meta
- **Nested Views:** Automatic for ARRAY fields

### ✅ Advanced Features

- **Deep Nesting:** Supports up to 3 levels of nested arrays
- **Conflict Resolution:** Automatically handles dimension/dimension_group conflicts
- **Metadata Support:** Reads dbt meta tags for custom LookML
- **Error Handling:** Flexible error strategies (fail fast, fail at end, continue)
- **Context Support:** Cancellable operations with timeout support

## Next Steps

- **[Configuration Guide](configuration.md)** - Learn all configuration options
- **[CLI Reference](cli-reference.md)** - Complete command-line reference
- **[Examples](examples/)** - Real-world usage examples
- **[Error Handling](error-handling.md)** - Understanding error strategies
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions

## Getting Help

- **Documentation:** [docs/public/](.)
- **Issues:** [GitHub Issues](https://github.com/magnus-ffcg/go-dbt2lookml/issues)
- **Discussions:** [GitHub Discussions](https://github.com/magnus-ffcg/go-dbt2lookml/discussions)

## Contributing

Interested in contributing? See our [Contributing Guide](../../CONTRIBUTING.md).

---

**Next:** [Configuration Guide](configuration.md) →
