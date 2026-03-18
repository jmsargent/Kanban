# Architecture Design — auto-assign-on-start

**Feature**: auto-assign-on-start
**Date**: 2026-03-18
**Paradigm**: OOP / idiomatic Go (established in CLAUDE.md + ADR-003)
**Pattern**: Hexagonal Architecture — Ports and Adapters (ADR-001)

---

## Overview

This is a targeted brownfield enhancement. No new components, packages, or port interfaces are introduced. The change modifies two existing files:

| File | Change |
|------|--------|
| `internal/adapters/cli/start.go` | Add `git.GetIdentity()` call before use case; hard fail on error; pass `identity.Name` to use case |
| `internal/usecases/start_task.go` | `Execute` gains `assignee string` param; sets `task.Assignee`; captures `PreviousAssignee` in result |

---

## Quality Attributes Driving Design

| Attribute | Priority | Rationale |
|-----------|----------|-----------|
| Architecture compliance | Highest | Identity resolution must stay in CLI adapter per ADR-001 + ADR-007 |
| Testability | High | Use case must remain testable with in-memory fakes only |
| Behavioural consistency | High | Hard-fail on identity error matches `kanban new` (ADR-007) |
| Minimal change surface | High | Brownfield — one story, two file changes |

---

## C4 System Context Diagram

```mermaid
C4Context
  title kanban — System Context

  Person(dev, "Developer", "Uses kanban CLI to manage work items in a git repository")

  System(kanban, "kanban CLI", "Git-native kanban task manager. Manages task files in .kanban/tasks/, hooks into git commit-msg, and integrates with CI.")

  System_Ext(gitrepo, "Git Repository", "Hosts task files (.kanban/tasks/*.md) and git config (user.name, user.email). All state is version-controlled.")

  Rel(dev, kanban, "Runs commands", "CLI (stdin/stdout/stderr)")
  Rel(kanban, gitrepo, "Reads/writes task files; reads git config; installs hooks", "filesystem + git subprocess")
```

---

## C4 Container Diagram

```mermaid
C4Container
  title kanban — Container View

  Person(dev, "Developer")

  Container_Boundary(bin, "kanban binary (single Go binary)") {
    Component(cli, "CLI Adapter", "cobra", "Primary adapter. Parses args, resolves git context + identity, maps use case results to stdout/stderr/exit codes.")
    Component(uc, "Use Cases", "Go", "Application logic. StartTask, AddTask, EditTask, DeleteTask, GetBoard, etc. No I/O dependencies.")
    Component(domain, "Domain", "Go", "Pure business rules and types. Task, Board, Column, Transition. Zero external imports.")
    Component(gitadapter, "Git Adapter", "Go / os/exec", "Secondary adapter. Implements GitPort. Wraps git CLI subprocess calls: RepoRoot, GetIdentity, CommitFiles, InstallHook.")
    Component(fsadapter, "Filesystem Adapter", "Go / os", "Secondary adapter. Implements TaskRepository + ConfigRepository. Atomic file writes via os.Rename.")
  }

  System_Ext(gitconfig, "git config", "user.name, user.email")
  System_Ext(taskfiles, ".kanban/tasks/*.md", "Task markdown files with YAML front matter")

  Rel(dev, cli, "kanban start <id>", "CLI")
  Rel(cli, uc, "Execute(repoRoot, taskID, assignee)", "function call")
  Rel(uc, domain, "CanTransitionTo(), Task{}", "pure function call")
  Rel(cli, gitadapter, "RepoRoot(), GetIdentity()", "GitPort interface")
  Rel(uc, fsadapter, "FindByID(), Update()", "TaskRepository interface")
  Rel(gitadapter, gitconfig, "git config user.name", "subprocess")
  Rel(fsadapter, taskfiles, "read / atomic write", "filesystem")
```

---

## C4 Component Diagram — start command flow

```mermaid
C4Component
  title kanban start — Component Interaction (auto-assign-on-start change)

  Person(dev, "Developer")

  Component(startcmd, "start.go (CLI Adapter)", "cobra RunE", "Orchestrates: resolves repo root → resolves identity → invokes StartTask → maps result to output.")
  Component(gitport, "GitPort", "interface", "Port contract: RepoRoot() + GetIdentity() + ...")
  Component(gitimpl, "GitAdapter", "git subprocess", "Implements GitPort. GetIdentity() calls 'git config user.name'.")
  Component(starttask, "StartTask (use case)", "Go struct", "Execute(repoRoot, taskID, assignee string). Sets task.Assignee, captures PreviousAssignee, persists via TaskRepository.")
  Component(taskrepo, "TaskRepository", "interface", "Port contract: FindByID() + Update() + ...")
  Component(fsimpl, "FilesystemAdapter", "os + atomic write", "Implements TaskRepository. Reads/writes .kanban/tasks/*.md.")

  Rel(dev, startcmd, "kanban start TASK-001", "CLI")
  Rel(startcmd, gitport, "RepoRoot()", "port call")
  Rel(startcmd, gitport, "GetIdentity() → Identity{Name}", "port call — NEW")
  Rel(gitport, gitimpl, "implements", "")
  Rel(startcmd, starttask, "Execute(repoRoot, taskID, identity.Name)", "NEW: assignee param")
  Rel(starttask, taskrepo, "FindByID() + Update(task)", "port call")
  Rel(taskrepo, fsimpl, "implements", "")
```

---

## Data Flow — happy path (AC-09-1)

```
Developer
  │ kanban start TASK-001
  ▼
start.go
  ├─ git.RepoRoot()            → "/path/to/repo"
  ├─ git.GetIdentity()         → Identity{Name: "Jon", Email: "jon@…"}
  │   └─ ErrGitIdentityNotConfigured → stderr error, osExit(1) ◄─── hard fail (FR-3)
  └─ StartTask.Execute("/path/to/repo", "TASK-001", "Jon")
        ├─ config.Read()       → validates kanban is initialised
        ├─ tasks.FindByID()    → Task{Assignee: "", Status: "todo"}
        ├─ CanTransitionTo()   → true
        ├─ task.Assignee = "Jon"        ◄─── NEW
        ├─ task.Status = "in-progress"
        └─ tasks.Update(task)  → atomic write to .kanban/tasks/TASK-001.md
             └─ returns StartTaskResult{Transitioned: true, PreviousAssignee: ""}
start.go
  ├─ stdout: "Started TASK-001: Fix login bug"
  └─ (no warning — PreviousAssignee is empty)
exit 0
```

---

## Data Flow — reassignment path (AC-09-2)

```
git.GetIdentity()    → Identity{Name: "Bob"}
tasks.FindByID()     → Task{Assignee: "Alice", Status: "todo"}
task.Assignee = "Bob"    ◄─── overwrite
task.Status = "in-progress"
tasks.Update(task)
returns StartTaskResult{Transitioned: true, PreviousAssignee: "Alice"}

start.go:
  stdout: "Started TASK-002: Update docs"
  stdout: "Note: task was previously assigned to Alice"   ◄─── FR-2
exit 0
```

---

## Integration Points

| Component | Change | Impact |
|-----------|--------|--------|
| `start.go` | Add `GetIdentity()` call + hard fail + pass to use case | Mirrors existing pattern in `new.go` |
| `StartTask.Execute` | New `assignee string` param | 1 call site to update (`start.go`) |
| `StartTaskResult` | New `PreviousAssignee string` field | All existing callers unaffected (zero-value is empty string) |
| Existing unit tests | Signature update only | 5 tests need `assignee ""` added to `Execute` calls |
| Acceptance tests | New milestone file for this feature | No changes to existing milestone files |

---

## Reuse vs New

| Decision | Choice | Rationale |
|----------|--------|-----------|
| `GitPort.GetIdentity()` | **Reuse** | Already implemented in `GitAdapter`, already on the port (ADR-007) |
| `ports.ErrGitIdentityNotConfigured` | **Reuse** | Already declared in `ports/errors.go` |
| `domain.Task.Assignee` | **Reuse** | Field already exists in domain model |
| `ports.Identity` struct | **Reuse** | Already in `ports/git.go` |
| Hard-fail pattern on identity error | **Reuse** | Mirrors `new.go` L38-42 verbatim |
| New port interface | **Not needed** | No new port methods required |
| New package | **Not needed** | Change is confined to existing packages |
