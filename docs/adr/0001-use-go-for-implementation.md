# ADR 0001: Use Go for Implementation

**Status:** Accepted

**Date:** 2024

## Context

We need to build a tool that converts dbt models to LookML views. The tool needs to:
- Parse large JSON files (manifest.json, catalog.json)
- Generate LookML files efficiently
- Be distributed as a standalone binary
- Work across different platforms (Linux, macOS, Windows)
- Have minimal runtime dependencies

## Decision

We will implement dbt2lookml in Go.

## Rationale

**Pros:**
- **Single binary distribution** - No runtime dependencies (Python, Node.js, etc.)
- **Fast execution** - Compiled language with excellent JSON parsing performance
- **Cross-platform** - Easy to build for multiple platforms
- **Strong typing** - Catches errors at compile time
- **Good tooling** - Built-in testing, formatting, and dependency management
- **Easy deployment** - Users can download a single binary

**Cons:**
- **Learning curve** - Team may need to learn Go
- **Less common in data space** - Python is more common for data tools

**Alternatives considered:**
- **Python** - Common in data space but requires Python runtime and dependencies
- **Node.js** - Good for JSON but requires Node.js runtime
- **Rust** - Fast but steeper learning curve and longer compile times

## Consequences

**Positive:**
- Users can run the tool without installing dependencies
- Fast performance for large dbt projects
- Easy CI/CD integration
- Simple installation process

**Negative:**
- Contributors need Go knowledge
- Smaller ecosystem for dbt-specific tooling compared to Python

## Notes

This decision aligns with the goal of making dbt2lookml easy to integrate into any workflow, regardless of the environment.
