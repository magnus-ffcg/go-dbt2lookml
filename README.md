# go-dbt2lookml

**Convert dbt models to LookML views for Looker**

go-dbt2lookml is a CLI tool that generates LookML views from BigQuery via dbt models. It parses dbt manifest and catalog files to create comprehensive LookML views with dimensions, measures, and explores.

## Features

- Parse dbt manifest and catalog files
- Generate LookML views, dimensions, measures, and explores
- Support for complex nested BigQuery structures (ARRAY, STRUCT)
- Flexible CLI and YAML configuration
- Comprehensive validation and error handling
- Continue-on-error mode for partial generation
- Structured logging with JSON and console output

## Installation

```bash
go install github.com/magnus-ffcg/go-dbt2lookml/cmd/dbt2lookml@latest
```

Or download the latest binary from [Releases](https://github.com/magnus-ffcg/go-dbt2lookml/releases).

## Quick Start

```bash
# Using dbt target directory
dbt2lookml --target-dir target --output-dir lookml/views

# Using explicit paths
dbt2lookml \
  --manifest-path target/manifest.json \
  --catalog-path target/catalog.json \
  --output-dir lookml/views

# With configuration file
dbt2lookml --config config.yaml
```

## Documentation

Full documentation is available at **https://magnus-ffcg.github.io/go-dbt2lookml/**

## Contributing

Contributions are welcome! See the [documentation](https://magnus-ffcg.github.io/go-dbt2lookml/) for development setup, contributing guidelines, and API reference.

## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.
