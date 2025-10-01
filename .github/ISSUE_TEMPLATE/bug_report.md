---
name: Bug Report
about: Create a report to help us improve
title: '[BUG] '
labels: bug
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

Steps to reproduce the behavior:

1. Run command: `dbt2lookml ...`
2. With config: `...`
3. See error: `...`

## Expected Behavior

A clear and concise description of what you expected to happen.

## Actual Behavior

What actually happened instead.

## Environment

**dbt2lookml version:** (run `dbt2lookml version`)
```
paste output here
```

**Operating System:**
- [ ] macOS
- [ ] Linux
- [ ] Windows
- Version: [e.g., macOS 14.0, Ubuntu 22.04, Windows 11]

**Go version** (if building from source):
```
go version output
```

## Configuration

**Config file** (if applicable):
```yaml
# Paste relevant parts of your config.yaml here
# Remove sensitive information
```

**Command used:**
```bash
dbt2lookml --manifest-path ... --catalog-path ...
```

## dbt Information

**dbt version:**
```
dbt --version output
```

**Number of models:** [e.g., 500]

**BigQuery specifics:**
- [ ] Uses STRUCT types
- [ ] Uses ARRAY types
- [ ] Deeply nested structures (3+ levels)

## Logs/Error Output

<details>
<summary>Full error output</summary>

```
Paste full error output here
```

</details>

## Additional Context

Add any other context about the problem here. Screenshots, generated files, or links to your dbt project (if public) can be helpful.

## Possible Solution

If you have suggestions on how to fix the issue, please describe them here.

## Checklist

- [ ] I have searched existing issues to avoid duplicates
- [ ] I have provided all requested information
- [ ] I have removed any sensitive information from logs/config
- [ ] I am using the latest version of dbt2lookml
