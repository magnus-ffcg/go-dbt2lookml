---
title: go-dbt2lookml Documentation
type: docs
bookToc: false
---

# go-dbt2lookml

**Convert dbt models to LookML views for Looker**

go-dbt2lookml is a powerful CLI tool that generates LookML views from BigQuery via dbt models. It parses dbt manifest and catalog files to create comprehensive LookML views with dimensions, measures, and explores.

## Features

- **Parse dbt artifacts** - Works with manifest.json and catalog.json
- **Generate LookML views** - Complete with dimensions, measures, and explores
- **Complex structures** - Full support for BigQuery ARRAY and STRUCT types
- **Rich configuration** - Flexible CLI options and YAML config
- **Error handling** - Comprehensive validation and error strategies
- **Fast & efficient** - Written in Go for optimal performance

## Quick Start

### Installation

{{< tabs "install" >}}
{{< tab "Go Install" >}}

go install github.com/magnus-ffcg/go-dbt2lookml/cmd/dbt2lookml@latest

{{< /tab >}}
{{< tab "Download Binary" >}}
Download the latest release from [GitHub Releases](https://github.com/magnus-ffcg/go-dbt2lookml/releases)
{{< /tab >}}
{{< /tabs >}}

### Basic Usage

```bash
# Using dbt target directory
dbt2lookml --target-dir target --output-dir lookml/views

# Using explicit paths
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views
```

## Documentation

### Usage
- **[Getting Started](docs/usage/getting-started)** - Installation and first steps
- **[Configuration](docs/usage/configuration)** - All configuration options
- **[CLI Reference](docs/usage/cli-reference)** - Complete command-line reference
- **[Error Handling](docs/usage/error-handling)** - Error strategies and troubleshooting

### Development
- **[Development Setup](docs/development/setup)** - Set up your development environment
- **[Contributing](docs/development/contributing)** - How to contribute
- **[Testing](docs/development/testing)** - Testing guide and practices
- **[API Reference](docs/api)** - Go package documentation

## Contributing

Contributions are welcome! Please see our [Contributing Guide](docs/development/contributing) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](https://github.com/magnus-ffcg/go-dbt2lookml/blob/main/LICENSE) file for details.
