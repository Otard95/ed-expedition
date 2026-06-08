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

## Suites

Use `github.com/stretchr/testify/suite` when tests share setup/teardown or state.

```go
type MyTestSuite struct {
    suite.Suite
    // shared state
}

func (s *MyTestSuite) SetupTest() { /* runs before each test */ }
func (s *MyTestSuite) TearDownTest() { /* runs after each test */ }
func (s *MyTestSuite) SetupSuite() { /* runs once before all tests */ }

func (s *MyTestSuite) TestSomething() {
    s.Equal(expected, actual)   // assert methods available directly on suite
    s.Require().NoError(err)    // require available via s.Require()
}

// Entry point — required for go test to discover the suite
func TestMyTestSuite(t *testing.T) {
    suite.Run(t, new(MyTestSuite))
}
```

**Note:** Suites do not support parallel tests.

## Examples

- `migrations/main_test.go` — table-driven tests with inline registries
- `migrations/route_test.go` — domain-specific tests with helper functions
- `services/expedition_test.go` — testify suites with setup/teardown and `TestLogger`
