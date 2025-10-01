## Description

<!-- Provide a brief description of the changes in this PR -->

## Type of Change

<!-- Mark the relevant option with an 'x' -->

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Refactoring (no functional changes, code improvements)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Test improvements

## Related Issues

<!-- Link to related issues -->

Fixes #(issue)
Relates to #(issue)

## Changes Made

<!-- Describe the changes in detail -->

- Change 1
- Change 2
- Change 3

## Testing

<!-- Describe the tests you ran and how to reproduce them -->

### Test Coverage

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] All tests pass locally

### Manual Testing

```bash
# Commands used for manual testing
go test ./...
go build ./cmd/dbt2lookml
./dbt2lookml --help
```

**Test scenarios:**
1. Scenario 1: [Description]
2. Scenario 2: [Description]

## Performance Impact

<!-- If applicable, describe performance implications -->

- [ ] No performance impact
- [ ] Performance improved: [describe]
- [ ] Performance regression: [describe and justify]

## Breaking Changes

<!-- List any breaking changes and migration steps -->

- [ ] No breaking changes
- [ ] Breaking changes (describe below)

**Migration guide:**
```
Steps for users to migrate
```

## Documentation

- [ ] Code comments updated
- [ ] README.md updated
- [ ] CHANGELOG.md updated
- [ ] docs/ folder updated (if applicable)
- [ ] Examples added/updated

## Checklist

<!-- Mark completed items with an 'x' -->

### Code Quality

- [ ] Code follows the project's style guidelines
- [ ] Self-review of code completed
- [ ] Comments added for complex logic
- [ ] No unnecessary console logs or debug code
- [ ] Error handling is appropriate

### Testing

- [ ] All existing tests pass
- [ ] New tests added for new functionality
- [ ] Tests cover edge cases
- [ ] Integration tests pass
- [ ] No test warnings or flaky tests

### Security

- [ ] No sensitive information in code/commits
- [ ] Input validation added where needed
- [ ] Security implications considered
- [ ] Dependencies are up to date

### CI/CD

- [ ] CI pipeline passes
- [ ] golangci-lint passes
- [ ] Race detector passes
- [ ] Build succeeds for all platforms

### Documentation

- [ ] User-facing changes documented
- [ ] API changes documented
- [ ] Migration guide provided (if breaking changes)

## Screenshots/Examples

<!-- If applicable, add screenshots or example output -->

### Before
```
Example of old behavior
```

### After
```
Example of new behavior
```

## Additional Notes

<!-- Any additional information for reviewers -->

## Reviewer Notes

<!-- Optional: specific areas you'd like reviewers to focus on -->

- Please pay special attention to: [...]
- Questions I have: [...]
- Known limitations: [...]

---

**By submitting this PR, I confirm that:**

- [ ] I have read and followed the [CONTRIBUTING.md](../CONTRIBUTING.md) guidelines
- [ ] My code follows the project's coding standards
- [ ] I have added tests that prove my fix/feature works
- [ ] All new and existing tests pass
