# Component Boundaries: Internal Go BDD DSL

**Feature**: acceptance-tests
**Wave**: DESIGN
**Date**: 2026-03-16

---

## Overview

Two units of code exist: the `dsl` package (shared infrastructure) and the individual feature test files (one per feature area). Neither unit imports from `internal/`.

---

## dsl Package — `tests/acceptance/dsl/`

### Ownership

The `dsl` package owns all mutable state for a single test scenario and all mechanics for driving the kanban binary as a subprocess.

### `dsl.Context`

Owns:

| Field | Type | Description |
|-------|------|-------------|
| `t` | `*testing.T` | Test handle. Used for `t.Fatalf`, `t.Helper`, and `t.Cleanup` registration. |
| `repoDir` | `string` | Absolute path to the per-test temporary git repository. Created by `InAGitRepo` setup step. |
| `binPath` | `string` | Absolute path to the kanban binary. Resolved once at construction time: `KANBAN_BIN` env var, then project-root-relative fallback. |
| `env` | `[]string` | Environment slice passed to every subprocess. Starts as a copy of `os.Environ()`. Modified by `EnvVarSet` steps. |
| `lastStdout` | `string` | Captured stdout from the most recent subprocess invocation. |
| `lastStderr` | `string` | Captured stderr from the most recent subprocess invocation. |
| `lastOutput` | `string` | `lastStdout + lastStderr`. Available to assertion steps via `OutputContains` etc. |
| `lastExit` | `int` | Exit code from the most recent subprocess invocation. |
| `lastTaskID` | `string` | Most recently captured `TASK-NNN` pattern from output. Available to step factories that need a dynamic task ID (e.g., `ctx.LastTaskID()`). |

Cleanup registration: `NewContext(t)` registers a `t.Cleanup` callback that calls `os.RemoveAll(ctx.repoDir)` when the test ends. No explicit `defer` is needed in test files.

`Context` does NOT own git process management directly. The internal `runner` function (unexported, lives in `dsl/runner.go`) is the single point for subprocess invocation.

### `Step` type

```
Step struct {
    Description string       // human-readable, used in Fatalf messages
    Run         func(*Context) error
}
```

Step factories are plain Go functions that return a `Step` value. Factories that need parameters close over them. Example shape:

```
ATaskWithStatus(title, status string) Step  ->  Step{Description: "a task "+title+" with status "+status, Run: func(*Context) error {...}}
```

### Orchestrators: `Given`, `When`, `Then`, `And`

All four have the same signature:

```
func Given(ctx *Context, step Step)
func When(ctx *Context, step Step)
func Then(ctx *Context, step Step)
func And(ctx *Context, step Step)
```

Behaviour:

1. Call `ctx.t.Helper()` so failure lines point to the test file, not the orchestrator.
2. Invoke `step.Run(ctx)`.
3. If error is non-nil: `ctx.t.Fatalf("<Phase>: %s: %v", step.Description, err)`.
4. `And` is a direct alias for `Then` — identical implementation.

The phase label in Fatalf is the function name (`Given`, `When`, `Then`). This maps to the requirement: `t.Fatalf("Given: a task with status todo: %v", err)`.

### Step Factory Files

Step factories are grouped by category into separate files within the `dsl` package. They are not sub-packages — grouping is by filename only, for navigability.

| File | Category | Factory functions |
|------|----------|-------------------|
| `dsl/setup.go` | Setup | `InAGitRepo`, `KanbanInitialised`, `NoKanbanSetup`, `NotAGitRepo`, `ATaskWithStatus`, `ATaskWithStatusAs`, `ATaskExists`, `NoTasksExist`, `TaskFileExists`, `CommitHookInstalled`, `EnvVarSet`, `PipelineCommitWith` |
| `dsl/actions.go` | Action | `IRunKanban`, `IRunKanbanNew`, `IRunKanbanNewWithOptions`, `IRunKanbanBoard`, `IRunKanbanBoardJSON`, `IRunKanbanStart`, `IRunKanbanStartOnThatTask`, `IRunKanbanEdit`, `IRunKanbanEditTitle`, `IRunKanbanDelete`, `IRunKanbanDeleteForce`, `ICommitWithMessage`, `ICommitWithTaskID`, `CIStepRunsPass`, `CIStepRunsFail` |
| `dsl/assertions.go` | Assertion | `ExitCodeIs`, `StdoutContains`, `OutputContains`, `OutputMatchesNone`, `OutputIsValidJSON`, `JSONHasFields`, `TaskHasStatus`, `TaskStatusRemains`, `TaskFileExists`, `TaskFileRemoved`, `BoardShowsTaskUnder`, `BoardNotListsTask`, `GitCommitExitCodeIs`, `WorkspaceReady`, `ConfigFileHasDefaults`, `HookLogInGitignore`, `NoTempFilesRemain`, `UpdatedTaskCommitted`, `NoAutoCommitFromDelete`, `NoKanbanOutputLines`, `NoTransitionLines`, `NoANSIEscapeCodes`, `NoSpinnerChars` |
| `dsl/runner.go` | Internal | `run(ctx, args...)` — unexported subprocess executor |
| `dsl/context.go` | Core | `Context` struct, `NewContext(t)` constructor, `LastTaskID()` accessor |
| `dsl/step.go` | Core | `Step` type, `Given`, `When`, `Then`, `And` orchestrators |

### Import Rule

`tests/acceptance/dsl/` imports:

- Standard library only: `os`, `os/exec`, `path/filepath`, `strings`, `regexp`, `bytes`, `context`, `time`, `testing`, `encoding/json`, `fmt`.
- Zero imports from `github.com/jmsargent/kanban/internal/`.
- Zero imports from any third-party package (testify is available but not needed in the DSL itself; it may be used optionally in assertion step implementations where its error messages add clarity).

---

## Feature Test Files — `tests/acceptance/`

### Location and naming

One `_test.go` file per feature area, alongside the `dsl` package directory:

```
tests/acceptance/
  dsl/
    context.go
    step.go
    runner.go
    setup.go
    actions.go
    assertions.go
  init_test.go
  task_crud_test.go
  auto_transitions_test.go
  start_command_test.go
```

### What lives in test files

- `package acceptance` declaration (external test package — exercises the binary, not internal types).
- One `TestXxx` function per scenario.
- Composition of step factories into `Given / When / Then / And` calls.
- Use of `ctx.LastTaskID()` where the test needs to reference a dynamically generated ID.

### What does NOT live in test files

- Subprocess invocation logic (lives in `dsl/runner.go`).
- Temp directory creation or cleanup (lives in `dsl/setup.go` and `NewContext`).
- Assertion helpers beyond calling the step factories (assertion logic lives in `dsl/assertions.go`).
- Any import from `internal/`.

### Import rule

Test files import:

```go
import (
    "testing"
    "github.com/jmsargent/kanban/tests/acceptance/dsl"
)
```

No other imports are expected for standard scenarios. Tests that need to assert raw file content may additionally import `os` and `path/filepath`, but this should be the exception — prefer assertion step factories.

---

## Boundary Summary

```
tests/acceptance/*_test.go
    |
    | imports
    v
tests/acceptance/dsl/         (Step factories, Context, orchestrators)
    |
    | imports stdlib only
    v
os/exec --> kanban binary --> .kanban/tasks/ in t.TempDir()

NEVER:
tests/acceptance/dsl/ --> internal/domain/
tests/acceptance/dsl/ --> internal/usecases/
tests/acceptance/*_test.go --> internal/
```

This boundary enforces the hexagonal architecture rule from CLAUDE.md: acceptance tests exercise the CLI driving port (binary subprocess), not the application core directly.
