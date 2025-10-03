# ADR 0008: Generate One File Per LookML View

**Status:** Accepted

**Date:** 2024

## Context

When generating LookML views, we need to decide the output file structure:
- Single file with all views
- One file per view
- Group views by schema/folder
- User-configurable structure

LookML projects typically organize views in separate files for maintainability.

## Decision

We will generate one LookML file per dbt model/view, organized by schema with an option to flatten the structure.

## Rationale

**Pros:**
- **Standard LookML practice** - Matches how Looker developers work
- **Easy to navigate** - Find specific view files easily
- **Git-friendly** - Changes to one view don't affect others
- **Parallel editing** - Multiple developers can work simultaneously
- **Easier code review** - Smaller, focused PRs
- **Looker IDE compatible** - Works well with Looker's file browser

**Cons:**
- **Many files** - Large projects generate many files
- **Import complexity** - Need to import all files in model file
- **Folder structure** - Need to organize by schema

**Alternatives considered:**
- **Single file** - Hard to maintain, poor git diffs
- **Schema-based grouping** - Still creates large files
- **Custom structure** - Too complex to configure

## Consequences

**Positive:**
- Familiar structure for Looker developers
- Easy to find and edit specific views
- Clean git history
- Works well with Looker's IDE

**Negative:**
- Many files for large projects
- Need directory structure management
- Model file must include all views

## Implementation

Default structure:
```
lookml/
  views/
    schema_name/
      model_name.view.lkml
      another_model.view.lkml
```

Flattened structure (with `--flatten` flag):
```
lookml/
  views/
    model_name.view.lkml
    another_model.view.lkml
```

## Configuration

Users can control output structure:
- `--output-dir` - Base directory for views
- `--flatten` - Skip schema subdirectories
- `--remove-schema-string` - Remove prefix from schema names

## Notes

This decision aligns with Looker best practices and makes the generated LookML easy to integrate into existing Looker projects.
