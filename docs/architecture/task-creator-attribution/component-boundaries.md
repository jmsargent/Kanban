# Component Boundaries — task-creator-attribution

## Boundary Map

Each row shows which layer owns the concern and what crosses the boundary.

| Concern | Owner | Boundary Crossing |
|---------|-------|-------------------|
| Reading git identity | `adapters/git` (GitAdapter) | Returns `ports.Identity` via `ports.GitPort` |
| Validating non-empty name | `adapters/cli` (new.go) | Error exit before use case call |
| Carrying name into use case | `usecases` (AddTaskInput) | `AddTaskInput.CreatedBy string` |
| Domain representation | `domain` (Task) | `Task.CreatedBy string` |
| Serializing to file | `adapters/filesystem` (TaskRepository) | `taskFrontMatter.CreatedBy` yaml tag |
| Excluding from edit | `adapters/filesystem` (TaskRepository) | Absent from `editFields` struct |
| Displaying on board | `adapters/cli` (board.go) | Reads `Task.CreatedBy`, renders `--` if empty |
| Error sentinel | `ports` (errors.go) | `ErrGitIdentityNotConfigured` |

---

## Package Dependency Graph (feature-relevant)

```
cmd/kanban/main.go
  └── internal/adapters/cli/          [primary adapters — wired here]
        ├── new.go
        │     → internal/ports/git.go (GitPort.GetIdentity)
        │     → internal/usecases/add_task.go (AddTask.Execute)
        └── board.go
              → internal/usecases/get_board.go (GetBoard.Execute)

internal/usecases/add_task.go
  → internal/domain/task.go           (Task, TaskStatus)
  → internal/ports/repositories.go    (TaskRepository, ConfigRepository)

internal/adapters/git/git_adapter.go
  → internal/ports/git.go             (GitPort, Identity)
  [implements GitPort — wired at cmd/kanban]

internal/adapters/filesystem/task_repository.go
  → internal/domain/task.go           (Task)
  → internal/ports/repositories.go    (TaskRepository, EditFilePort, EditSnapshot)
  [implements TaskRepository and EditFilePort — wired at cmd/kanban]
```

**No new edges added.** The feature extends existing packages, not wiring.

---

## Immutability Boundary: Edit Flow

The edit flow (kanban edit) preserves `CreatedBy` through structural exclusion:

```
FindByID(repoRoot, id)
  → Task{CreatedBy: "Jonathan Sargent", ...}   [full task loaded from disk]
       │
       ▼
editor.WriteTemp(task)
  → temp.yaml contains: title, priority, due, assignee, description
  → created_by is NOT in editFields struct → NOT written to temp file
       │
       ▼ [user edits temp file]
editor.ReadTemp(path)
  → EditSnapshot{Title, Priority, Due, Assignee, Description}
  → created_by absent from EditSnapshot → NOT read back
       │
       ▼
applyEditFields(task, updated)
  → task.Title = updated.Title
  → task.Priority = updated.Priority
  → task.Assignee = updated.Assignee
  → task.Description = updated.Description
  → task.CreatedBy unchanged (not in EditSnapshot, not touched)
       │
       ▼
TaskRepository.Update(repoRoot, task)
  → marshalTask includes task.CreatedBy → written back to file unchanged
```

The original `CreatedBy` value from `FindByID` survives the entire edit cycle unchanged.

---

## Compile-Time Interface Compliance

The existing pattern of `var _ ports.GitPort = (*GitAdapter)(nil)` in `git_adapter.go`
will catch any missing `GetIdentity` implementation at compile time.

Adding `GetIdentity()` to `ports.GitPort` will cause a **compile error** in the
test helper mocks (e.g. `ports_test.go`) until those also implement the new method.
This is a known consequence of extending an existing interface and is addressed in the
implementation guide (DISTILL wave).

---

## Test Isolation Boundaries

| Layer | Test Type | Identity concern |
|-------|-----------|-----------------|
| Domain | Pure unit | No — `CreatedBy` is a plain string field |
| Use cases | In-memory mock of `TaskRepository` | No — `CreatedBy` passed as `AddTaskInput.CreatedBy`, no git call |
| Git adapter | Integration (real git subprocess) | Yes — `TestGetIdentity` configures git user.name in temp repo |
| CLI (end-to-end) | Compiled binary subprocess | Yes — acceptance tests set `GIT_AUTHOR_NAME` or run `git config` |
