# Component Boundaries — new-editor-mode

**Wave**: DESIGN
**Date**: 2026-03-22

---

## Boundary Summary

This is a brownfield feature. The boundary table below describes every component touched, the nature of each change, and the explicit justification for each decision.

---

## Components That Change

### 1. `internal/adapters/cli/new.go` — `NewCreateCommand`

**Layer**: Primary adapter (CLI)
**Change type**: Additive (new RunE branch)

**Responsibilities after change**:
- Route `kanban new <title>` (one arg) to the existing inline path — unchanged
- Route `kanban new` (zero args) to the new editor-mode path:
  1. Call `git.RepoRoot()` — pre-flight
  2. Call `git.GetIdentity()` — pre-flight
  3. Call `config.Read(repoRoot)` — pre-flight (ErrNotInitialised → exit 1)
  4. Call `editor.WriteTempNew()` — produce blank template
  5. Defer `os.Remove(tmpFile)`
  6. Call `usecases.OpenEditor(tmpFile)` — block until editor exits
  7. Call `editor.ReadTemp(tmpFile)` — parse YAML
  8. Validate `title != ""` — empty → `os.Exit(2)`
  9. Call `usecases.NewAddTask(config, tasks).Execute(repoRoot, input)` — create task
  10. Print success output (identical to inline path)

**Constructor signature change**: `NewCreateCommand` gains an `editor ports.EditFilePort` parameter. The wiring site in `cmd/kanban/` passes the `TaskRepository` (which implements both `TaskRepository` and `EditFilePort`).

**What this component does NOT do**:
- Does not call `git add` or `git commit` (C-01)
- Does not write files directly — delegates to ports (C-02)
- Does not import any adapter package (C-04)

---

### 2. `internal/ports/repositories.go` — `EditFilePort`

**Layer**: Port (driven/secondary)
**Change type**: Interface extension

**Addition**: One new method added to the `EditFilePort` interface:

```
WriteTempNew() (string, error)
```

Contract:
- Produces a temporary YAML file containing a blank task template
- Template includes YAML comment lines providing field guidance (at minimum: `# title is required`)
- Template omits the `due` field (WD-01)
- Returns the absolute path to the temp file
- Caller is responsible for deletion

**`WriteTemp(task domain.Task) (string, error)`**: Unchanged. Still used by `EditTask` use case for the edit-existing-task workflow.

**`ReadTemp(path string) (EditSnapshot, error)`**: Unchanged. Shared by both edit and new-editor paths. Already comment-safe (YAML unmarshal ignores comment lines).

**Why a new method rather than a flag parameter**: A flag parameter on `WriteTemp` (e.g., `newMode bool`) would force `EditTask` to pass `false` on every call — coupling edit to awareness of a new feature. A separate method has a single clear purpose, is independently testable, and keeps the existing `WriteTemp` contract stable.

---

### 3. `internal/adapters/filesystem/task_repository.go` — `TaskRepository`

**Layer**: Secondary adapter (filesystem)
**Change type**: Additive (new method)

**Addition**: `WriteTempNew() (string, error)` implementation.

Behaviour:
- Creates a temporary file via `os.CreateTemp`
- Writes a YAML document with comment-annotated blank fields: `title`, `priority`, `assignee`, `description`
- Does NOT include `due` field
- Returns temp file path

**`WriteTemp` is unchanged.** The `editFields` struct, YAML marshal logic, and `due` handling in `WriteTemp` are not modified.

**Compile-time interface check**: The existing `var _ ports.EditFilePort = (*TaskRepository)(nil)` will enforce that `WriteTempNew` is implemented. No new enforcement line required — the existing line catches the omission.

---

### 4. `internal/usecases/editor.go` — `OpenEditor` (NEW FILE)

**Layer**: Use case
**Change type**: New file (extraction)

**Contains**: One exported function: `OpenEditor(filePath string) error`

**Behaviour** (extracted verbatim from `edit_task.go openEditor`):
- Reads `$EDITOR` environment variable
- Falls back to `vi` when `$EDITOR` is empty
- Launches editor via `exec.Command` with `cmd.Stdin`, `cmd.Stdout`, `cmd.Stderr` wired to `os.Stdin`/`os.Stdout`/`os.Stderr`
- Returns the command exit error (non-nil on non-zero exit)

**Imports**: `os`, `os/exec` (stdlib only). No violation of architecture rules.

**Why this is a use-case concern and not a port**: `OpenEditor` is application orchestration — it resolves an environment variable and launches a subprocess as part of a use case flow. It is not domain logic (no business rules) and not an adapter (it does not implement a port interface). The use cases package is the correct home.

---

### 5. `internal/usecases/edit_task.go` — `EditTask`

**Layer**: Use case
**Change type**: Minimal — one line changed

**Change**: The unexported `openEditor()` function is removed from this file. The call site `openEditor(tmpFile)` is replaced with `OpenEditor(tmpFile)` (calling the new exported function in the same package). No behavioural change.

---

## Components That Do NOT Change

| Component | Location | Reason |
|-----------|----------|--------|
| `AddTask` use case | `internal/usecases/add_task.go` | Called with a fully-populated `AddTaskInput` by the CLI adapter. Its internal `config.Read` is redundant for the editor-mode path but harmless and preserves the existing unit test coverage. No change warranted. |
| `domain/` (all files) | `internal/domain/` | No new domain types required. The feature is UI-layer behaviour using existing domain constructs. |
| `GitPort` interface | `internal/ports/` | No new git operations. |
| `ConfigRepository` interface | `internal/ports/repositories.go` | No new config operations. |
| `TaskRepository` interface | `internal/ports/repositories.go` | No new persistence operations for this feature. |
| `config_repository.go` | `internal/adapters/filesystem/` | No change. |
| `edit.go` | `internal/adapters/cli/edit.go` | No change. |

---

## Dependency Graph (delta only)

```
cmd/kanban/ (wiring)
  └─ passes EditFilePort (TaskRepository) to NewCreateCommand

internal/adapters/cli/new.go
  ├─ imports internal/ports (EditFilePort, GitPort, ConfigRepository)
  ├─ imports internal/usecases (AddTask, OpenEditor)
  └─ [no adapter imports — C-04 preserved]

internal/usecases/editor.go
  └─ imports os, os/exec (stdlib only)

internal/usecases/edit_task.go
  └─ calls OpenEditor() from same package [no new import]

internal/ports/repositories.go
  └─ EditFilePort gains WriteTempNew()

internal/adapters/filesystem/task_repository.go
  └─ implements WriteTempNew()
  └─ [no new external imports]
```

---

## Architecture Rule Compliance

| Rule | Compliance |
|------|-----------|
| `internal/domain` zero imports from non-stdlib | Maintained — no domain changes |
| `internal/usecases` zero imports from `internal/adapters` | Maintained — `editor.go` imports stdlib only |
| No adapter imports another adapter | Maintained — `cli/new.go` imports ports and usecases only |
| All secondary dependencies cross via port interfaces | Maintained — `EditFilePort` is the boundary |
| Enforced by `go-arch-lint` | No new rules required; existing config covers the new file |
