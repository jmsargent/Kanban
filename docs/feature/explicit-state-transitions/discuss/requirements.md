# Requirements — explicit-state-transitions

**Feature**: Explicit State Transitions (Decouple kanban from git commits)
**Date**: 2026-03-22
**Status**: DISCUSS complete

---

## Problem Statement

Users reported negative feedback from kanban autonomously performing git commits. The tool's current design has two paths that produce commits without developer intent:

1. `kanban ci-done --commit` — appends to transitions.log and commits it inside the CI pipeline
2. The commit-msg hook writes to transitions.log on each developer commit (file change, though no commit)

Beyond the commit concern, the transitions.log architecture was found to be unnecessary: the `.kanban/tasks/` folder already carries state via YAML front matter, and that state naturally changes across commits like any other project file.

---

## Goals

- **G-01**: Kanban never initiates a git commit on behalf of the developer or CI pipeline.
- **G-02**: Task state is readable from task file YAML front matter — no separate log file required.
- **G-03**: State transitions are explicit, initiated by the developer via CLI commands.
- **G-04**: Git identity recognition (`git config user.name/email`) is preserved for creator and assignee attribution.
- **G-05**: `kanban board` accurately reflects current task state from task files alone.

---

## Non-Goals

- Removing automatic state detection entirely from CI (CI may still identify tasks from commit messages; it just doesn't commit).
- Changing task file format or YAML front matter schema.
- Removing git identity reading — that stays.
- Changing `kanban add`, `kanban edit`, `kanban delete`, or `kanban board` display logic beyond the state-source change.

---

## Functional Requirements

### FR-01: Remove transitions.log as state source
`kanban board` MUST read task state from the `status:` field in each task file's YAML front matter. The transitions.log file and TransitionLogRepository port are removed from the architecture.

### FR-02: Remove commit-msg hook auto-transition
The commit-msg hook MUST be removed or made a no-op. It no longer intercepts commits to auto-transition tasks. `kanban install-hook` is deprecated or removed.

### FR-03: Add `kanban done` command
A new explicit CLI command `kanban done <TASK-ID>` MUST be added. It sets `status: done` in the task's YAML front matter and prints a confirmation. The developer then commits the changed file themselves.

### FR-04: `kanban ci-done` does not commit
`kanban ci-done` MUST scan commit messages in the given range, identify referenced task IDs, and update their `status:` to `done` in YAML front matter. It MUST NOT run any git commit. It MUST NOT stage files. Exit 0 when no tasks to transition (silent success). The CI config is responsible for staging and committing the updated task files.

### FR-05: `kanban start` updates task file only, does not commit
`kanban start <TASK-ID>` MUST update `status: in-progress` in the task's YAML front matter. It MUST NOT commit. (This is already the existing behaviour — confirm no auto-commit exists.)

### FR-06: Git identity preserved
`GitPort.GetIdentity()` is retained. Creator attribution (`created_by:`) and assignee auto-population on `kanban start` remain unaffected.

---

## Non-Functional Requirements

- **NFR-01**: `kanban done` executes in < 200ms on a repo with 500 task files.
- **NFR-02**: `kanban ci-done` output is CI-friendly: no ANSI, no spinners when `NO_COLOR` is set or no TTY.
- **NFR-03**: All task file writes remain atomic (write to `.tmp`, then `os.Rename`) — unchanged from existing constraint.

---

## Constraints

- **C-01**: Architecture remains hexagonal — no adapter imports another adapter.
- **C-02**: `internal/domain` has zero non-stdlib imports.
- **C-03**: The Go binary never calls `git commit` or `git add` anywhere in its implementation.
- **C-04**: Exit codes remain consistent: 0=success, 1=runtime error, 2=usage error.

---

## Impact on Existing Architecture

| Component | Change |
|-----------|--------|
| `transitions.log` | **Removed** — no longer written or read |
| `TransitionLogAdapter` | **Removed** from `internal/adapters/filesystem/` |
| `TransitionLogRepository` port | **Removed** from `internal/ports/` |
| `kanban _hook commit-msg` | **Removed** — hook handler deleted |
| `kanban install-hook` | **Removed** or deprecated |
| `kanban ci-done` | **Changed** — no longer commits; updates YAML only |
| `kanban done` | **New** explicit command |
| `kanban board` | **Changed** — reads `status:` from YAML (was already true pre-board-state-in-git) |
| `GitPort.GetIdentity()` | **Unchanged** |
| `kanban start` | **Unchanged** (already does not commit) |

---

## Migration Notes

The project is **not yet publicly released** (confirmed in project memory). No backward compatibility required. The transitions.log file, if present in any repo, can be deleted manually. No migration command is needed.
