# Wave Decisions: Internal Go BDD DSL

**Feature**: acceptance-tests
**Wave**: DESIGN
**Date**: 2026-03-16

---

## Decision Summary

| # | Decision | Rationale |
|---|----------|-----------|
| D-01 | Replace godog/Gherkin with an internal Go DSL | Gherkin files are a layer of indirection for a Go project. Regex step binding is not compile-time verified. Six transitive dependencies add no value to a single-binary CLI tool. See ADR-006. |
| D-02 | DSL lives in `tests/acceptance/dsl/` | Shared package importable by all feature test files. Keeps test infra separate from application source. |
| D-03 | `Step` is a struct (`Description string` + `Run func(*Context) error`) | Embeds human-readable text for Fatalf messages. Makes step factories self-describing without a registration map. |
| D-04 | Orchestrators stop on first failing step (`t.Fatalf`) | Prevents cascading failures from an invalid precondition obscuring the actual error. Consistent with the failure mode in the existing godog steps. |
| D-05 | `t.Cleanup` for temp directory teardown | Idiomatic Go since 1.14. No explicit `defer` in test bodies. Cleanup runs even when `t.Fatalf` is called. |
| D-06 | Binary resolution via `KANBAN_BIN` env var, then project-root-relative fallback | Preserves the existing CI and local developer workflow. No change required to `cicd/config.yml` during migration. |
| D-07 | Migration is incremental — godog coexists until all scenarios are ported | Zero-disruption to the existing green CI pipeline. Each ported scenario is independently verifiable before the Gherkin equivalent is removed. |
| D-08 | `dsl/` must not import `internal/` | Enforces the hexagonal boundary: acceptance tests exercise the CLI driving port (the compiled binary), never the application core directly. Enforced by `go-arch-lint`. |
| D-09 | testify/assert is optional within step factory implementations | Provides richer diff messages for equality assertions. Not exposed in the `Step` API. Does not change the DSL contract. |
| D-10 | `And` is an alias for `Then` | Preserves BDD readability (`Then(ctx, X); And(ctx, Y)`) without introducing a separate code path. |

---

## Quality Attribute Priorities

| Quality Attribute | Priority | How Addressed |
|-------------------|----------|---------------|
| Maintainability | High | Plain Go: rename, extract, find-usages all work. No string-to-function mapping to maintain. |
| Testability (of the tests) | High | Step factories are pure functions returning `Step` values. They can be unit-tested in isolation. |
| Compilability | High | Every step binding is a Go function call. Broken step factories are caught at `go build` time. |
| Readability | High | `Given / When / Then / And` structure mirrors Gherkin scenarios. Test bodies read as narrative. |
| Migration safety | High | Incremental porting. Existing godog tests are unaffected. CI stays green throughout. |
| Dependency minimisation | Medium | Removes six godog transitive dependencies. Adds one (testify, already present transitively). |

---

## Open Decisions (None)

All decisions were confirmed in the requirements gathering phase. No open decisions remain for the crafter.

---

## Constraints Carried Forward to Crafter

1. Step factories return `Step`, not `error`. The orchestrator owns the error-to-Fatalf mapping.
2. `Context` fields are unexported. Test files access state only through exported methods (`LastTaskID()`). Additional accessor methods may be added if test files require them — the crafter should prefer adding an accessor over exporting a field.
3. The `runner` function is unexported from `dsl/`. It is the single subprocess invocation path. Action and setup steps that invoke the binary must route through it.
4. No `.feature` files, `.yaml` test specs, or other external test definition files are created. The Go source files are the specification.
5. The `package` declaration in test files must be `package acceptance` (external test package), not `package dsl`.

---

## Migration Checklist for Crafter

- [ ] Create `tests/acceptance/dsl/` package with all files listed in `component-boundaries.md`.
- [ ] Verify `go build ./tests/acceptance/dsl/` passes.
- [ ] Add `go-arch-lint` rule forbidding `internal/` imports from `tests/acceptance/dsl/`.
- [ ] Port milestone-3-start-command.feature first (smallest, already partially active).
- [ ] Port milestone-2-auto-transitions.feature second.
- [ ] Port milestone-1-task-crud.feature last (largest, most steps).
- [ ] Tag each ported Gherkin scenario `@ported` or move to `@skip` group.
- [ ] After all ported: delete `.feature` files, delete `steps/` directory, run `go mod tidy`.
- [ ] Update `cicd/config.yml` to remove godog-specific invocation.
- [ ] Verify CI passes with `go test ./tests/acceptance/...` as sole acceptance command.
