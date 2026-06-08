---
name: testing
description: Go testing patterns for this repository. Use when writing or running tests. Covers test framework, conventions, and available utilities.
---

# Testing

## Framework

Standard `testing` package with [testify](https://github.com/stretchr/testify) (`github.com/stretchr/testify v1.11.1`).

- `require` — fails immediately on assertion failure
- `assert` — continues execution after failure

Use `require` for preconditions (test setup, data shape), `assert` for the actual checks.

## Running Tests

```bash
go test ./path/to/package/ -v
go test ./... -v              # all tests
```

## Conventions

- Test files: `<file>_test.go` in the same package (white-box testing)
- Table-driven tests for multiple cases of the same function
- Test helpers use `t.Helper()` for clean stack traces
- Test names: `TestFunctionName_Scenario` (e.g., `TestMigrate_EmptyRegistry_NoOp`)

## Examples

- `migrations/main_test.go` — generic engine tests with inline registries
- `migrations/route_test.go` — domain-specific migration tests with helpers
- `services/expedition_test.go` — testify suites with setup/teardown and `TestLogger`
