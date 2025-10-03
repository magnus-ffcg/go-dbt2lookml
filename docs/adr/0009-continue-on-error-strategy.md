# ADR 0009: Continue-on-Error Strategy for Partial Success

**Status:** Accepted

**Date:** 2024

## Context

When processing multiple dbt models, errors can occur:
- Invalid model configuration
- Missing required fields
- Type conversion errors
- Invalid LookML generation

We need to decide how to handle errors:
- Stop on first error (fail fast)
- Continue processing remaining models
- Collect errors and report at end

## Decision

We will support both behaviors with a `--continue-on-error` flag, defaulting to fail-fast for safety.

## Rationale

**Pros:**
- **Flexible** - Users choose based on their needs
- **Development-friendly** - Continue-on-error helps during development
- **Production-safe** - Fail-fast default ensures quality
- **Debugging** - See all errors at once with continue-on-error
- **Partial success** - Generate what's possible, flag what's not

**Cons:**
- **Complexity** - Need to track errors across processing
- **Incomplete output** - Continue-on-error produces partial results
- **User confusion** - Need to understand when to use each mode

**Alternatives considered:**
- **Always fail-fast** - Too rigid for development
- **Always continue** - Hides problems in production
- **Separate validate command** - Extra complexity

## Consequences

**Positive:**
- Development mode: see all issues at once
- Production mode: ensure complete, valid output
- Error reports show what succeeded/failed
- Users can choose based on workflow

**Negative:**
- Need clear error reporting
- Partial success can be confusing
- Must validate output carefully with continue-on-error

## Implementation

Default behavior (fail-fast):
```bash
dbt2lookml --config config.yaml
# Stops on first error
```

Continue-on-error mode:
```bash
dbt2lookml --config config.yaml --continue-on-error
# Processes all models, reports errors at end
```

Error reporting includes:
- Which models succeeded
- Which models failed (with reasons)
- Total success/failure count
- Exit code reflects errors (non-zero if any failed)

## Usage Recommendations

**Use fail-fast (default) when:**
- Running in CI/CD
- Production deployments
- Quality is critical

**Use continue-on-error when:**
- Local development
- Debugging configuration issues
- Migrating large projects
- Want to see all problems at once

## Notes

This pattern is common in build tools (make -k, go test -failfast=false) and provides the right balance of safety and flexibility.
