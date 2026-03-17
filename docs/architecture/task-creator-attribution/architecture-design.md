# Architecture Design — task-creator-attribution

## Overview

The `task-creator-attribution` feature adds automatic creator capture to `kanban new` and
surfaces the creator on `kanban board`. It integrates into the existing hexagonal architecture
(ADR-001) without introducing new packages or new architectural patterns. All changes are
additive: a new port method, a new domain field, extended serialization, and updated display.

**Paradigm**: Object-Oriented (Go idiomatic — existing project standard per CLAUDE.md)
**Pattern**: Hexagonal Architecture / Ports-and-Adapters (existing — ADR-001)
**No new packages required.**

---

## C4 Level 1 — System Context

```mermaid
C4Context
  title System Context — kanban CLI (task-creator-attribution)

  Person(dev, "Developer", "Creates and manages tasks from the terminal")

  System(kanban, "kanban CLI", "Git-native kanban task manager. Tracks work items in Markdown files under .kanban/tasks/")

  System_Ext(gitrepo, "Git Repository", "Stores task files and git identity (user.name via git config)")
  System_Ext(filesystem, "Local Filesystem", "Persists .kanban/tasks/*.md files and .kanban/config.yml")

  Rel(dev, kanban, "Runs commands", "CLI (stdin/stdout)")
  Rel(kanban, gitrepo, "Reads git identity, repo root, commit history; installs hook", "git CLI subprocess")
  Rel(kanban, filesystem, "Reads and writes task files atomically", "os package")

  UpdateRelStyle(dev, kanban, $offsetX="-40")
  UpdateRelStyle(kanban, gitrepo, $offsetX="10")
```

---

## C4 Level 2 — Container Diagram

```mermaid
C4Container
  title Container Diagram — kanban binary (task-creator-attribution)

  Person(dev, "Developer")

  Container_Boundary(binary, "kanban binary (single Go binary)") {
    Component(cobra, "Cobra CLI Adapter", "Go / cobra", "Primary adapter. Routes commands: new, board, edit, start, delete, init, hook, ci-done. Resolves git identity before calling use cases.")
    Component(usecases, "Use Cases", "Go", "Application logic: AddTask, GetBoard, EditTask, StartTask, DeleteTask, InitRepo. AddTask accepts CreatedBy via AddTaskInput.")
    Component(domain, "Domain Core", "Go", "Task, Board, Column, Transition, ValidationError. Task.CreatedBy string field. Zero external imports.")
    Component(gitadapter, "Git Adapter", "Go / os/exec", "Secondary adapter. Implements GitPort: RepoRoot, GetIdentity, CommitFiles, InstallHook, AppendToGitignore, CommitMessagesInRange.")
    Component(fsadapter, "Filesystem Adapter", "Go / os", "Secondary adapter. Implements TaskRepository, ConfigRepository, EditFilePort. Serializes Task including created_by. Excludes created_by from editable temp file.")
  }

  System_Ext(gitrepo, "Git Repository", "git config, .git/hooks/, commit history")
  System_Ext(fs, "Local Filesystem", ".kanban/tasks/*.md")

  Rel(dev, cobra, "kanban new / board / edit / ...", "CLI args, stdin")
  Rel(cobra, usecases, "Execute(repoRoot, AddTaskInput{..., CreatedBy})", "Go method call")
  Rel(usecases, domain, "constructs / validates domain.Task", "Go struct")
  Rel(usecases, fsadapter, "Save / Update / FindByID / ListAll", "ports.TaskRepository")
  Rel(cobra, gitadapter, "RepoRoot() / GetIdentity()", "ports.GitPort")
  Rel(gitadapter, gitrepo, "git rev-parse / git config user.name", "subprocess")
  Rel(fsadapter, fs, "atomic write (.tmp → rename) / read", "os package")

  UpdateRelStyle(dev, cobra, $offsetX="-40")
```

---

## C4 Level 3 — Component: kanban new with Creator Attribution

This diagram shows the precise execution path for `kanban new` after this feature ships.

```mermaid
C4Component
  title Component: kanban new — Creator Attribution Flow

  Person(dev, "Developer", "runs kanban new")

  Component(newcmd, "NewCreateCommand", "cli/new.go", "Cobra command handler. Calls GetIdentity(), validates non-empty, builds AddTaskInput with CreatedBy.")
  Component(gitadapter, "GitAdapter.GetIdentity()", "adapters/git/git_adapter.go", "Runs: git config user.name. Returns Identity{Name} or error if empty.")
  Component(addtask, "AddTask.Execute()", "usecases/add_task.go", "Validates title/due, reads config, generates ID, constructs Task{CreatedBy}, calls TaskRepository.Save.")
  Component(taskrepo, "TaskRepository.Save()", "adapters/filesystem/task_repository.go", "Marshals Task (including created_by) to Markdown+YAML. Writes atomically via .tmp → os.Rename.")

  Rel(dev, newcmd, "kanban new 'title'", "CLI args")
  Rel(newcmd, gitadapter, "GetIdentity()", "ports.GitPort")
  Rel(newcmd, addtask, "Execute(repoRoot, AddTaskInput{Title, CreatedBy, ...})", "use case call")
  Rel(addtask, taskrepo, "Save(repoRoot, Task{..., CreatedBy})", "ports.TaskRepository")
```

---

## Integration Points with Existing Components

| Component | Change | Nature |
|-----------|--------|--------|
| `internal/domain/task.go` | Add `CreatedBy string` to `Task` struct | Additive — plain field, no imports |
| `internal/ports/git.go` | Add `Identity` type + `GetIdentity() (Identity, error)` to `GitPort` | Additive — new method on existing interface |
| `internal/usecases/add_task.go` | Add `CreatedBy string` to `AddTaskInput`; set `task.CreatedBy` in `Execute` | Additive — existing callers unaffected (zero-value = empty string) |
| `internal/adapters/filesystem/task_repository.go` | Add `created_by` to `taskFrontMatter`; update `marshalTask`/`unmarshalTask` | Additive — backward compatible (missing field → empty string) |
| `internal/adapters/git/git_adapter.go` | Implement `GetIdentity()` | Additive — compile-time check enforces coverage |
| `internal/adapters/cli/new.go` | Call `GetIdentity()` before use case; guard on empty name | Modification — error path added before existing flow |
| `internal/adapters/cli/board.go` | Add `CreatedBy` to board row + JSON output | Modification — display-only, no logic change |

---

## Dependency Rule Compliance

The dependency rule (all arrows point inward toward domain) is preserved:

```
cli/new.go
  → ports.GitPort.GetIdentity()          [existing boundary]
  → usecases.AddTaskInput.CreatedBy      [existing boundary]

usecases/add_task.go
  → domain.Task.CreatedBy                [existing boundary]
  → ports.TaskRepository.Save            [existing boundary]

adapters/filesystem
  → domain.Task.CreatedBy                [existing boundary, read direction]

adapters/git
  → ports.Identity                       [new type, lives in ports]
```

**`internal/domain` has zero new imports.** `CreatedBy string` is a plain field.
**`internal/usecases` has zero new imports from `internal/adapters`.**
**No adapter imports another adapter.**

---

## Identity Validation Rule

**Pre-condition**: `git config user.name` must return a non-empty, non-whitespace-only string.

- `GitAdapter.GetIdentity()` trims the output of `git config user.name`. If the result is empty,
  it returns `ErrGitIdentityNotConfigured` (not an `Identity` with empty `Name`).
- The CLI adapter (`cli/new.go`) checks for this error and exits 1 with the setup instructions
  message **before** any use case call or file write.
- There is zero tolerance for an empty `created_by` on a newly created task. The error path
  (AC-03-1, AC-03-2) is the enforcement mechanism.

This rule lives in the adapter, not the domain. The use case trusts that any non-empty
`AddTaskInput.CreatedBy` it receives is valid.

---

## Identity Resolution: Pre-Condition Guard Pattern

The CLI adapter applies a pre-condition guard before invoking the use case — a pattern
already established in `new.go` for the `RepoRoot` check. Identity resolution follows
the same pattern:

```
1. git.RepoRoot()         → error → exit 1 "Not a git repository"
2. git.GetIdentity()      → error/empty → exit 1 "git identity not configured — run: ..."
3. uc.Execute(...)        → error → exit per error type
```

This keeps the use case clean (no git concerns) and the error handling centralized in the
adapter layer where it belongs under hexagonal architecture.

---

## Immutability Mechanism

Creator immutability is enforced by **structural exclusion** — the simplest possible
mechanism that satisfies the requirement:

- `editFields` struct in `filesystem/task_repository.go` does **not** include `created_by`
- `WriteTemp` writes only `editFields` — `created_by` never enters the temp file
- `applyEditFields` in `edit_task.go` copies only the fields in `EditSnapshot` back to the task
- Since `EditSnapshot` has no `CreatedBy` field, the value from the original `FindByID` call
  is carried through `applyEditFields` unchanged and written back by `Update`

No domain-level enforcement is needed. The architecture boundary guarantees immutability.

---

## Backward Compatibility

- Task files without `created_by` in front matter are parsed by `unmarshalTask` with `CreatedBy = ""`
- `kanban board` renders `--` for any task where `CreatedBy == ""`
- `kanban board --json` emits `"created_by": ""` for such tasks
- No migration step required. No data loss.
