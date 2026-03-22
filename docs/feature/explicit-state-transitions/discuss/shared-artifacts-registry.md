# Shared Artifacts Registry — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

| Artifact | Location | Written By | Read By | Notes |
|----------|----------|------------|---------|-------|
| Task YAML `status:` field | `.kanban/tasks/<TASK-ID>.md` | `kanban done`, `kanban ci-done`, `kanban start` | `kanban board` | Single source of board state |
| Task YAML `created_by:` field | `.kanban/tasks/<TASK-ID>.md` | `kanban add` | `kanban board` | Populated via `GitPort.GetIdentity()` |
| Task YAML `assignee:` field | `.kanban/tasks/<TASK-ID>.md` | `kanban start` | `kanban board` | Populated via `GitPort.GetIdentity()` |
| Git commit messages | git history | Developer / CI | `kanban ci-done --since` | Scanned for task ID pattern |
| `git config user.name/email` | git config | Developer tooling | `GitPort.GetIdentity()` | Read-only; never modified by kanban |

## Removed Artifacts (from board-state-in-git)

| Artifact | Former Location | Status |
|----------|----------------|--------|
| `transitions.log` | `.kanban/transitions.log` | **Removed** — no longer written or read |
| Hook log | `.kanban/hook.log` | **Removed** — hook is gone |
| Hook script | `.kanban/hooks/commit-msg` | **Removed** — directory deleted |
