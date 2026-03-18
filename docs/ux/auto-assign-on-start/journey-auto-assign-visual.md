# Journey Map — Auto-Assign on Start (Visual)

**Feature**: auto-assign-on-start
**Journey name**: Developer starts a task and is automatically assigned
**Type**: Backend CLI
**Research depth**: Lightweight

---

## Journey Overview

```
Developer                      kanban CLI                    Git / Filesystem
    |                               |                               |
    |-- kanban start TASK-001 ----> |                               |
    |                               |-- git.RepoRoot() -----------> |
    |                               |<-- repoRoot ------------------|
    |                               |-- git.GetIdentity() --------> |
    |                               |<-- Identity{Name:"Jon"} ------|
    |                               |-- tasks.FindByID() ---------->|
    |                               |<-- Task{Assignee:""} ---------|
    |                               |                               |
    |                               | [set assignee = "Jon"]        |
    |                               | [set status = in-progress]    |
    |                               |                               |
    |                               |-- tasks.Update(task) -------> |
    |                               |<-- ok ------------------------|
    |                               |                               |
    |<-- "Started TASK-001: ..." ---|                               |
```

---

## Emotional Arc

| Step | Action | Feeling |
|------|--------|---------|
| 1 | Runs `kanban start TASK-001` | Focused — wants to claim task quickly |
| 2 | Sees "Started TASK-001: Fix bug" | Confident — board is accurate without extra steps |
| 3 (variant) | Sees warning about previous assignee | Mildly surprised — but understands they've claimed it |

**Emotional arc**: Neutral → Confident. No friction. The automation removes a manual step the user would otherwise forget.

---

## Happy Path

```
GIVEN: task exists in todo, no assignee, git identity configured
WHEN:  developer runs "kanban start TASK-001"
THEN:  task transitions todo → in-progress
       task.Assignee set to git user.name
       stdout: "Started TASK-001: <title>"
       exit code: 0
```

---

## Variant: Task Already Assigned to Someone Else

```
GIVEN: task exists in todo, assignee = "Alice"
WHEN:  developer "Bob" runs "kanban start TASK-001"
THEN:  task transitions todo → in-progress
       task.Assignee overwritten to "Bob"
       stdout: "Started TASK-001: <title>"
       stdout: "Note: task was previously assigned to Alice"
       exit code: 0
```

---

## Variant: Git Identity Not Configured

```
GIVEN: task exists in todo, git user.name is not set
WHEN:  developer runs "kanban start TASK-001"
THEN:  task remains in todo (no state change)
       stderr: "git identity not configured — run: git config --global user.name ..."
       exit code: 1
```

---

## Variant: Task Already In-Progress (Idempotence Preserved)

```
GIVEN: task is already in-progress, assignee = "Alice"
WHEN:  developer runs "kanban start TASK-001"
THEN:  no state change (existing behaviour preserved)
       stdout: "Task TASK-001 is already in progress"
       exit code: 0
```

> Rationale: assignee is not updated on idempotent re-starts to preserve the
> guarantee established in US-08 (start-task feature) that already-in-progress
> is a no-op. Re-assignment would require an explicit mechanism.

---

## Shared Artifacts

| Artifact | Source | Consumer |
|----------|--------|----------|
| `task.Assignee` | `git config user.name` via `GitPort.GetIdentity()` | Task file, board display |
| `task.Status` | Business rule: todo → in-progress | Board, commit hook |
| `StartTaskResult.PreviousAssignee` | Previous value of task.Assignee before update | CLI warning output |
