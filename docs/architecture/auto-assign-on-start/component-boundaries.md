# Component Boundaries — auto-assign-on-start

## Dependency Rule Compliance

All changes respect the hexagonal dependency rule: dependencies point inward.

```
internal/adapters/cli/start.go
  imports: internal/ports, internal/usecases   ✓ (no new imports)

internal/usecases/start_task.go
  imports: internal/domain, internal/ports     ✓ (no new imports)

internal/domain/
  imports: stdlib only                         ✓ (no change)
```

No adapter imports another adapter. No use case imports an adapter. The domain imports nothing external.

---

## Changed Interfaces

### `StartTask.Execute` — signature change

**Before**:
```go
func (u *StartTask) Execute(repoRoot, taskID string) (StartTaskResult, error)
```

**After**:
```go
func (u *StartTask) Execute(repoRoot, taskID, assignee string) (StartTaskResult, error)
```

The `assignee` value is the caller-supplied git identity name. The use case does not know how it was obtained — it is plain data injected through the driving port.

### `StartTaskResult` — new field

**Before**:
```go
type StartTaskResult struct {
    Transitioned      bool
    AlreadyInProgress bool
    Task              domain.Task
}
```

**After**:
```go
type StartTaskResult struct {
    Transitioned      bool
    AlreadyInProgress bool
    Task              domain.Task
    PreviousAssignee  string  // non-empty when task had a different assignee before this start
}
```

`PreviousAssignee` is the value of `task.Assignee` **before** the update, captured only when `Transitioned = true` and the old value differed from the new. The CLI adapter uses this to decide whether to emit the `Note:` warning line.

---

## Unchanged Interfaces

| Interface | Status |
|-----------|--------|
| `ports.GitPort` | No change — `GetIdentity()` already exists |
| `ports.TaskRepository` | No change |
| `ports.ConfigRepository` | No change |
| `ports.Identity` | No change |
| `ports.ErrGitIdentityNotConfigured` | No change |
| `domain.Task` | No change — `Assignee` field already exists |

---

## Call Site Inventory

### `start.go` — the only caller of `StartTask.Execute`

Order of operations after change:
1. `git.RepoRoot()` — unchanged
2. `git.GetIdentity()` — **new call**, hard fail on `ErrGitIdentityNotConfigured`
3. `usecases.NewStartTask(config, tasks)` — unchanged
4. `uc.Execute(repoRoot, taskID, identity.Name)` — **new third argument**
5. Result handling — **new**: check `result.PreviousAssignee` to emit optional warning

### `start_test.go` — unit tests for `start.go` (CLI adapter)

All `uc.Execute(repoRoot, taskID)` calls become `uc.Execute(repoRoot, taskID, "")`. Existing fake `GitPort` implementations gain a `GetIdentity()` stub returning a configured identity.

### `start_task_test.go` — unit tests for `StartTask` use case

All `uc.Execute(repoRoot, taskID)` calls become `uc.Execute(repoRoot, taskID, "")` or with a specific assignee where the test exercises the new behaviour.

---

## Architecture Lint Compliance

`go-arch-lint` rules (from CLAUDE.md) are not affected:
- `internal/domain` gains no new imports ✓
- `internal/usecases` gains no imports from `internal/adapters` ✓
- No adapter package imports another adapter ✓
- All secondary port dependencies cross via interfaces ✓
