# Journey Map: Explicit State Transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22
**Persona**: Developer + CI pipeline

---

## Developer Journey: Moving a Task to Done

```
DEVELOPER JOURNEY: "I finished work on a task and want to mark it done"

Phases:     [ Work on Task ]──────────────[ Mark Done ]──────────────[ Commit ]
                                                │
                                     kanban done TASK-NNN

Steps:
  1. Dev completes work
  2. Dev runs: kanban done TASK-NNN
  3. kanban updates task YAML: status: done (atomic write)
  4. kanban prints: "kanban: TASK-NNN moved in-progress -> done"
  5. Dev runs: git add .kanban/tasks/TASK-NNN.md
  6. Dev commits: git commit -m "feat: ... closes TASK-NNN"
  7. kanban board shows TASK-NNN in DONE column

Emotional arc:
  [ neutral ]──[ immediate feedback ]──[ control ]──[ satisfaction ]
                      ↑                     ↑
               Output is instant       Dev owns the commit
```

---

## CI Pipeline Journey: Auto-transitioning tasks on pipeline pass

```
CI PIPELINE JOURNEY: "Pipeline passes — mark referenced tasks as done"

Phases:     [ Tests Pass ]────────[ kanban ci-done ]────────[ Explicit Commit ]────[ Push ]
                                         │                         │
                              No commit by kanban        CI config owns this
                              Just file updates           (visible, auditable)

Steps:
  1. CI: all tests pass (exit 0)
  2. CI runs: kanban ci-done --since=$CIRCLE_SHA1^
  3. kanban scans commits for task IDs
  4. For each task found: updates status: done in YAML (atomic write)
  5. kanban prints transitions, exits 0
  6. CI runs: git add .kanban/tasks/ && git commit -m "kanban: mark tasks done [skip ci]"
  7. CI runs: git push
  8. kanban board shows updated state

Emotional arc (team reviewing CI log):
  [ expectation ]──[ transparency ]──[ trust ]──[ confidence ]
                         ↑                 ↑
                  Commit step visible   Team sees exactly
                  in CI config          what kanban changed

Key difference from old behavior:
  BEFORE: kanban ci-done --commit (kanban owns the commit — invisible, surprising)
  AFTER:  kanban ci-done (file update only) + explicit CI step (visible, team-owned)
```

---

## Shared Artifacts

| Artifact | Source | Consumers |
|----------|--------|-----------|
| Task YAML `status:` field | `kanban done`, `kanban ci-done`, `kanban start` | `kanban board` |
| Commit message | Developer / CI | `kanban ci-done --since` scan |
| `git config user.name/email` | `GitPort.GetIdentity()` | `kanban add` (creator), `kanban start` (assignee) |

---

## Error Paths

| Scenario | Behaviour |
|----------|-----------|
| `kanban done NONEXISTENT` | Exit 1, stderr: task not found |
| `kanban done` on already-done task | Exit 0, stdout: already done (idempotent) |
| CI: `kanban ci-done` finds no tasks | Exit 0, silent |
| Leftover `.git/hooks/commit-msg` on dev machine | Delegates to `kanban _hook commit-msg` → safe no-op, exit 0 |
| Task file unwritable | Exit 1, stderr: permission error |
