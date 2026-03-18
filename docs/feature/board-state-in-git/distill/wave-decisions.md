# DISTILL Wave Decisions — board-state-in-git

Decisions made during the acceptance test design phase. These supplement
the DESIGN wave decisions already recorded in `design/wave-decisions.md`.

---

## Decision D-01: Test framework — Custom Go DSL

**Decision**: Use the Custom Go DSL (Go standard testing package with Given/When/Then
step functions) that already exists in `tests/acceptance/dsl/`.

**Rationale**: The project already has a mature, consistent DSL used by 8 existing
acceptance test files. Introducing a different framework (e.g., godog/Gherkin) would:
- Create a two-framework test suite that developers must context-switch between.
- Require a new dependency with different execution semantics.
- Break the project's test execution contract (`make acceptance` with `KANBAN_BIN`).

The existing DSL is production-grade, follows the same Given/When/Then vocabulary as
BDD, and its step factories are reusable across test files.

**Consequence**: All new test files are in `package acceptance_test` (actually
`package acceptance` following the existing convention), DSL extensions go in
`tests/acceptance/dsl/kanban_steps.go`.

---

## Decision D-02: Walking skeleton — US-BSG-01 (kanban log)

**Decision**: US-BSG-01 is the walking skeleton, enabled immediately. US-BSG-02 and
US-BSG-03 are fully skipped (Milestone 1 and Milestone 2 respectively).

**Rationale**:
1. US-BSG-01 requires no new ports or adapters — it reads existing git history via
   the existing `GitPort`. This makes it the lowest-risk first slice.
2. US-BSG-01 delivers the most visible user value: "I can see my task's history."
   A stakeholder can watch `kanban log TASK-001` run and confirm value in under 30 seconds.
3. US-BSG-02 (transitions.log) is foundational infrastructure; it must be complete
   before US-BSG-01 scenarios 2, 3, 9, 10 can be enabled (they depend on log entries
   being present). The skip strategy correctly models this dependency.

**Consequence**: 5 tests enabled immediately. 35 tests deferred. Implementation
sequence: walking skeleton → Milestone 1 → Milestone 2.

---

## Decision D-03: Skip/pending strategy

**Decision**: Tests that cannot pass yet are marked with `t.Skip("pending: ...")` as
the first statement after the `KANBAN_BIN` guard. The skip message names the story
(US-BSG-02 or US-BSG-03) to make it easy to locate the next batch of tests to enable.

**Rationale**: `t.Skip` is the idiomatic Go mechanism. It reports tests as SKIPPED
(not failed) in test output, making CI green while clearly communicating "not yet
implemented." This is preferable to build tags or comment-out because:
- Skipped tests appear in `go test -v` output, providing visibility.
- They can be selectively un-skipped with a one-line edit, enforcing "one at a time."
- The skip message is searchable (`grep -r "US-BSG-02"`) for discovery.

**Consequence**: DELIVER wave crafter enables one test, drops to the inner TDD loop,
implements, passes, commits. The outer loop test count decreases by 1 per commit cycle.

---

## Decision D-04: DSL extension in kanban_steps.go (separate from pipeline_steps.go)

**Decision**: All new step factories for the board-state-in-git feature are placed in
`tests/acceptance/dsl/kanban_steps.go`, not in `pipeline_steps.go`.

**Rationale**: `pipeline_steps.go` tests Makefile targets and CI scripts — it operates
on the real project directory and invokes shell scripts. `kanban_steps.go` tests
domain behaviour via the kanban binary in isolated temp repos. These are distinct
concerns. Mixing them would make the DSL harder to navigate and risk setup step
cross-contamination.

The existing DSL files (`setup.go`, `actions.go`, `assertions.go`) cover general
kanban binary invocation. `kanban_steps.go` adds step factories specific to the
transitions.log and board --me capabilities, keeping related steps co-located.

**Consequence**: `kanban_steps.go` is in `package dsl`. It uses `Context`, `Step`,
`Given/When/Then/And`, `run()`, `gitCmd()`, and `taskFilePath()` from the existing
DSL files — no duplication.

---

## Decision D-05: transitions.log direct reads are permitted in US-BSG-02 tests

**Decision**: Tests for US-BSG-02 may read `.kanban/transitions.log` directly from
the filesystem to assert line counts and field format.

**Rationale**: The Port-to-Port principle requires that tests invoke through driving
ports (the binary). However, `transitions.log` is an observable, committed file output
of the system — not an internal implementation detail. Reading it for structural
assertions (line count = 1, fields = 5, no truncation) is analogous to reading a task
file to assert its YAML front matter. The existing DSL already does this for task files
(see `TaskHasStatus` in assertions.go which reads `.kanban/tasks/<id>.md`).

The binary remains the only entry point for all state-changing operations. File reads
are validation only.

**Consequence**: `TransitionsLogLineCountIs`, `TransitionsLogLastLineContains`, and
`TransitionsLogHasNoTruncatedLines` in `kanban_steps.go` read the file directly. These
helpers are clearly named and isolated in the DSL layer.

---

## Decision D-06: AC-01-11 (performance benchmark) excluded from automated suite

**Decision**: The performance criterion for `kanban log` (2s limit on 10,000+ commit
repos) is excluded from the automated acceptance suite and documented as a local-only
benchmark.

**Rationale**: Documented in full in `test-scenarios.md`. Summary: CI timing
variability makes automated perf assertions unreliable; a 10,000-commit repo creation
adds 3-5 minutes per CI run; the criterion is non-functional and not stakeholder-
demonstrable from test output.

**Consequence**: AC-01-11 coverage is a local benchmark, not a CI gate. The benchmark
function signature is documented in `test-scenarios.md`. A developer must run it
manually before a release that touches `kanban log` performance.

---

## Upstream issues inherited from DESIGN wave

### No migration = no backward-compatibility tests

The DESIGN wave decision "no `kanban migrate`, no YAML fallback" means there are no
scenarios testing behaviour on existing repositories with YAML status fields. This
simplifies US-BSG-02 significantly — there is no compatibility mode to test. All
US-BSG-02 scenarios assume a freshly initialised repository.

This is correct given the project is not publicly released. If the project were
publicly released, AC-02-1 through AC-02-3 would require additional scenarios covering
the migration path.

### TransitionLogRepository.LatestStatus missing entry = todo

The DESIGN wave specifies that `LatestStatus` returns `(TaskStatus, error)` and
callers treat a missing entry as `todo`. This is validated by:
- `TestTransitionsLog_NewTaskAppearsInTodoColumn` (AC-02-3)
- `TestTransitionsLog_BoardShowsTodo_WhenNoLogEntries` (AC-02-17)
- `TestBoardMe_WorksWithTransitionsLogStatusStorage` (AC-03-7, indirectly)

No additional test is needed — the existing tests cover this contract from the user's
perspective.
