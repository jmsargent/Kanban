# Technology Stack: Internal Go BDD DSL

**Feature**: acceptance-tests
**Wave**: DESIGN
**Date**: 2026-03-16

---

## Stack Decisions

### Go 1.22+ standard `testing` package

| Attribute | Value |
|-----------|-------|
| License | BSD 3-Clause (Go standard library) |
| Source | https://pkg.go.dev/testing |
| Rationale | The project is already a Go 1.22 module. `testing.T`, `t.Fatalf`, `t.Cleanup`, and `t.Helper` provide everything the DSL needs. No additional framework dependency is required for test structure. |

`t.Cleanup` is used in `NewContext` to register temp directory teardown. This is the idiomatic Go pattern since Go 1.14, replacing explicit `defer` calls in test bodies.

`t.Helper` is called at the top of each orchestrator function so that failure output (`t.Fatalf`) reports the line in the test file, not the orchestrator implementation.

### `os/exec` — Subprocess Invocation

| Attribute | Value |
|-----------|-------|
| License | BSD 3-Clause (Go standard library) |
| Source | https://pkg.go.dev/os/exec |
| Rationale | Already used in `kanban_steps_test.go`. The compiled binary-as-subprocess model is a confirmed constraint. `exec.CommandContext` with a 10-second timeout (matching the existing implementation) provides process isolation and prevents test hangs. |

### `github.com/stretchr/testify/assert` — Optional Assertion Clarity

| Attribute | Value |
|-----------|-------|
| License | MIT |
| Source | https://github.com/stretchr/testify |
| GitHub stars | 24k+ |
| Maintenance | Active, regular releases |
| Rationale | testify/assert produces more readable diffs than raw `fmt.Errorf` comparisons when values don't match (e.g., JSON field checks, multi-line output). It is optional — assertion step factories may use it internally where it adds clarity to failure messages. It is NOT exposed in the `dsl.Step` API; step factories return `error` regardless. |
| Addition to go.mod | `require github.com/stretchr/testify v1.9.0` — direct dependency replacing the current indirect reference via godog transitives. |

### godog — Retained During Migration, Then Removed

| Attribute | Value |
|-----------|-------|
| Current role | Sole acceptance test runner |
| Migration role | Runs in parallel with DSL-based tests during incremental porting |
| Removal trigger | All Gherkin scenarios ported to Go DSL |

#### godog removal plan

1. **During migration (Phase 1 and 2)**: godog remains in `go.mod` as a direct dependency. Both test suites run in CI.
2. **After all scenarios are ported (Phase 3)**:
   - Delete `tests/acceptance/kanban-tasks/steps/` and all `.feature` files.
   - Run `go mod tidy` to remove:
     - `github.com/cucumber/godog`
     - `github.com/cucumber/gherkin/go/v26`
     - `github.com/cucumber/messages/go/v21`
     - `github.com/gofrs/uuid`
     - `github.com/hashicorp/go-immutable-radix`
     - `github.com/hashicorp/go-memdb`
     - `github.com/hashicorp/golang-lru`
   - Update `cicd/config.yml` to remove any godog-specific invocation paths.
   - The net result: `go.mod` loses seven dependencies and gains one (`testify`, promoted from indirect to direct).

#### CI invocation before and after

Before (godog + DSL coexisting):
```sh
go test ./tests/acceptance/kanban-tasks/steps/   # godog suite
go test ./tests/acceptance/...                    # DSL-based tests
```

After (godog removed):
```sh
go test ./tests/acceptance/...
```

---

## No New BDD Frameworks

The following were considered and rejected:

| Option | Reason rejected |
|--------|----------------|
| `github.com/onsi/ginkgo` | External BDD framework, adds dependency, introduces separate test runner binary, non-standard `go test` invocation. |
| `github.com/cucumber/godog` (keep) | Regex-based step binding is not compile-time verified; two artefacts per test; six transitive dependencies for a CLI tool. |
| testscript (`golang.org/x/tools/cmd/testscript`) | Evaluated in ADR-006. Uses a custom script language — trades Gherkin indirection for testscript indirection. Closer to Go but still not plain Go. |

The DSL uses nothing beyond the Go standard library and optionally testify/assert. All test behaviour is expressed in plain Go.
