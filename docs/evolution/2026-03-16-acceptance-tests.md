# Evolution: acceptance-tests

**Date**: 2026-03-16
**Feature ID**: acceptance-tests
**Wave**: DELIVER

---

## Summary

Replaced the godog/Gherkin BDD test framework with an internal Go DSL for acceptance testing. The project had 5 `.feature` files driving acceptance tests via regex-bound step functions (godog). These were replaced by a pure-Go DSL in `tests/acceptance/dsl/` that provides the same Given/When/Then/And narrative structure but with compile-time verification, no regex step binding, and zero external BDD dependencies.

---

## Business Context

The kanban-tasks CLI is a Go project. Gherkin `.feature` files are a layer of indirection that adds no value for a single-language, single-binary project: step bindings are string-matched at runtime, six transitive dependencies were added for godog, and broken step factories were invisible until test execution. The internal DSL eliminates all three issues while preserving the BDD readability that makes acceptance tests self-documenting. See ADR-006.

---

## Steps Completed

| Step | Description | Status |
|------|-------------|--------|
| 01-01 | Create `tests/acceptance/dsl/` package with Context, Step, orchestrators | DONE |
| 01-02 | Implement setup step factories | DONE |
| 01-03 | Implement action and assertion step factories | DONE |
| 02-01 | Port milestone-3-start-command.feature (5 scenarios) | DONE |
| 02-02 | Port walking-skeleton and milestone-2-auto-transitions features | DONE |
| 02-03 | Port milestone-1-task-crud feature | DONE |
| 02-04 | Remove godog, delete feature files, clean up tooling config | DONE |

Total: 7 steps across 2 phases (DSL Infrastructure + Migration)

---

## Key Design Decisions

| # | Decision | Rationale |
|---|----------|-----------|
| D-01 | Replace godog with internal Go DSL | No compile-time verification for regex steps; six transitive dependencies for no net benefit |
| D-02 | DSL in `tests/acceptance/dsl/` | Shared package importable by all feature test files; separate from application source |
| D-03 | `Step` = `{Description, Run func(*Context) error}` | Human-readable for Fatalf messages; step factories are self-describing |
| D-04 | Orchestrators stop at first failure via `t.Fatalf` | Prevents cascading failures from invalid preconditions |
| D-05 | `t.Cleanup` for teardown | Idiomatic Go 1.14+; runs even when `t.Fatalf` called |
| D-06 | Binary via `KANBAN_BIN` env, then project-root fallback | Preserves CI and developer workflow unchanged |
| D-07 | Incremental migration — godog coexisted until all ported | Zero CI disruption throughout |
| D-08 | `dsl/` must not import `internal/` | Hexagonal boundary: acceptance tests drive the CLI binary, not the domain core |
| D-10 | `And` is an alias for `Then` | BDD readability without a separate code path |

---

## DSL Package Structure

```
tests/acceptance/dsl/
  context.go    — Context struct, NewContext(t), LastTaskID(), RepoDir()
  step.go       — Step type, Given/When/Then/And orchestrators
  runner.go     — unexported run() subprocess executor
  setup.go      — 13 setup factories (InAGitRepo, KanbanInitialised, WithLastOutput, ...)
  actions.go    — 15 action factories (IRunKanbanStart, CIStepRunsFail, ...)
  assertions.go — 22 assertion factories (ExitCodeIs, TaskHasStatus, NoANSIEscapeCodes, ...)

tests/acceptance/
  start_command_test.go      — 5 scenarios (todo→in-progress, already-in-progress, done, not-found, uninitialised)
  walking_skeleton_test.go   — 2 ported + stubs
  auto_transitions_test.go   — 2 ported + stubs
  init_test.go               — 1 ported + stubs
  task_crud_test.go          — 1 ported + stubs
```

---

## Lessons Learned

1. **`WithLastOutput` factory needed for failure-path testing** — assertion factories that check `ctx.lastOutput` could not be failure-tested without a way to inject output without a subprocess. `WithLastOutput(output string) Step` was added to `setup.go` during the adversarial review revision pass.

2. **`And` phase label bug found in refactor** — The `And` orchestrator was emitting `"Then:"` in Fatalf messages. Found during L1 refactoring. Fixed to emit `"And:"`.

3. **`CIStepRunsFail` tests actual binary behavior** — `KANBAN_TEST_EXIT=1` is appended to env but not handled by the binary. `CIStepRunsFail` therefore exercises the same transition path as `CIStepRunsPass`. Test asserts `exit 0` and `TaskHasStatus("done")` to reflect actual binary behavior.

4. **Adversarial review found Testing Theater** — Three tests were flagged: no assertions after `CIStepRunsFail`, and two `_Fail` tests that only checked factory description fields. All three fixed in the revision pass.

---

## Permanent Artifacts

- `docs/architecture/acceptance-tests/` — Design docs (architecture, component boundaries, data models, tech stack)
- `docs/adrs/ADR-006-internal-bdd-dsl.md` — Decision to replace godog
- `tests/acceptance/dsl/` — DSL package (production artifact)
- `tests/acceptance/*_test.go` — Feature test files (production artifact)
