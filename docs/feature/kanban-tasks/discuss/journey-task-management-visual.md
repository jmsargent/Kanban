# Journey: Task Management (Git-Native CLI) -- Visual Map

Feature: kanban-tasks
Persona: Rafael Rodrigues -- backend developer on a 4-person shared team repo
Goal: Track team work without leaving the terminal or maintaining a separate tool
Platform: CLI tool with CI/CD integration hooks
Date: 2026-03-15

Prior wave artifacts: none found (greenfield project)
Codebase: empty git repository

---

## Mental Model

Rafael thinks of the board as a mirror of the repository. Tasks are files committed alongside
code. The board is never wrong because state transitions are driven by real events -- commits
and pipeline results -- not by someone remembering to update a card.

    Code events  -->  board state changes  (automatic, driven by git)
    Task files   -->  tracked by git       (same repo, same history)

No external tool. No manual sync. The board is the repo.

---

## Emotional Arc

```
START                STEP 2            STEP 3            STEP 4            STEP 5-6          END
Skeptical /          Grounded          Focused           Confident         Clean             Trusting
context-switching    (task exists      (board updated    (CI moved it,     (corrections      "The board
from code to tool    in the repo,      itself from       no manual         and cleanup       is the repo.
                     no external       a commit,         step needed)      done without      It stays
"Another Jira?"      sync needed)      zero effort)                        ceremony)         accurate."
```

---

## Journey Flow

```
[Trigger: Rafael is about to start a new piece of work]
  Feels: Skeptical -- "will this stay in sync or go stale like Jira?"
         |
         v
[Step 1: View Board  --  kanban board]
  Sees: Tasks grouped by status in the terminal
  Feels: Orienting -- what is the team working on right now?
         |
         v
[Step 2: Create Task  --  kanban add "Fix OAuth login bug"]
  Sees: Task file created in .kanban/tasks/TASK-001.md and committed
  Feels: Grounded -- task lives in the codebase, not in a foreign system
         |
         v
[Step 3: Start Work  --  first commit referencing TASK-001]
  Event: git commit hook or CI step detects task ID in commit message
  Effect: task.status auto-updated to "in-progress" in task file
  Feels: Focused -- the board updated itself, zero extra effort
         |
         v
[Step 4: Complete Work  --  CI pipeline passes]
  Event: pipeline runs, all tests pass, TASK-001 referenced in commit
  Effect: task.status auto-updated to "done" in task file
  Feels: Confident -- done means done, tests prove it
         |
         v
[Step 5: Edit Task  --  kanban edit TASK-001]
  Sees: Task fields in editor ($EDITOR), saves updates task file
  Feels: In control -- can correct metadata without ceremony
         |
         v
[Step 6: Delete Task  --  kanban delete TASK-001]
  Sees: Confirmation prompt; confirms; task file removed
  Feels: Clean -- board reflects only live work
         |
         v
[End: Board always reflects codebase reality]
  Feels: Trusting -- "kanban board tells me exactly what is happening"
```

---

## Step-by-Step CLI Mockups

### Step 1: View Board (kanban board)

```
$ kanban board

Sprint board  .kanban/tasks/  (12 tasks)

TODO (4)
  TASK-001  Fix OAuth login bug           P2  due 2026-03-18  Rafael Rodrigues
  TASK-002  Write unit tests auth module  P2  due 2026-03-19  unassigned
  TASK-004  Update API docs               P3  due 2026-03-25  Priya Nair
  TASK-007  Refactor config loader        P3  --              unassigned

IN PROGRESS (2)
  TASK-003  API rate limiting             P1  due 2026-03-20  Alex Kim
  TASK-005  Database migration script     P2  due 2026-03-22  Rafael Rodrigues

DONE (6)
  TASK-006  Deploy v1.2 to staging        P1  done 2026-03-14
  TASK-008  Fix typo in README            P3  done 2026-03-13
  ...

Run `kanban board --help` for filter and format options.
```

### Step 1b: Empty Board

```
$ kanban board

No tasks found in .kanban/tasks/

Get started:
  kanban add "Your first task"

Run `kanban --help` for all commands.
```

### Step 2: Create Task (kanban add)

```
$ kanban add "Fix OAuth login bug" --priority=P2 --due=2026-03-18 --assignee="Rafael Rodrigues"

Created  TASK-001  Fix OAuth login bug
  File:      .kanban/tasks/TASK-001.md
  Status:    todo
  Priority:  P2
  Due:       2026-03-18
  Assignee:  Rafael Rodrigues
  Created:   2026-03-15

Tip: reference TASK-001 in commit messages to auto-move to in-progress.
  git commit -m "TASK-001: begin investigation"
```

### Step 2b: Create Task -- Validation Error (empty title)

```
$ kanban add ""

Error: Task title is required.

  Provide a non-empty title:
    kanban add "Description of the work"

Run `kanban add --help` for all options.
```

### Step 3: Auto-move to In Progress (commit hook)

```
$ git commit -m "TASK-001: reproduce OAuth bug on Chrome and Firefox"

[main 3f2a1c4] TASK-001: reproduce OAuth bug on Chrome and Firefox
 1 file changed, 3 insertions(+)

kanban: TASK-001 moved  todo -> in-progress
  Task:    Fix OAuth login bug
  Trigger: commit 3f2a1c4
```

### Step 4: Auto-move to Done (CI pipeline)

```
-- CI output (GitHub Actions / GitLab CI / etc.) --

[kanban] Scanning commits for task references...
[kanban] Found: TASK-001 in commit 9ab34e1 by Rafael Rodrigues
[kanban] Tests passed.
[kanban] TASK-001 moved  in-progress -> done
         Task:     Fix OAuth login bug
         Commit:   9ab34e1
         Pipeline: build #47
```

### Step 5: Edit Task (kanban edit)

```
$ kanban edit TASK-001

Editing  TASK-001  Fix OAuth login bug

  title       Fix OAuth login bug
  description (empty)
  priority    P2
  due         2026-03-18
  assignee    Rafael Rodrigues
  status      in-progress  [auto-managed -- edit to override manually]

Open in editor? [Y/n]: Y

-- $EDITOR opens .kanban/tasks/TASK-001.md --

Saved  TASK-001
  Updated: description, priority
```

### Step 5b: Edit -- Task Not Found

```
$ kanban edit TASK-099

Error: Task TASK-099 not found.

  No file at .kanban/tasks/TASK-099.md

  List all tasks:
    kanban board
```

### Step 6: Delete Task (kanban delete)

```
$ kanban delete TASK-001

Delete  TASK-001  Fix OAuth login bug  [in-progress]?

  This will remove .kanban/tasks/TASK-001.md from the repository.
  Commit the deletion to share with your team.

Confirm [y/N]: y

Deleted  TASK-001  Fix OAuth login bug
  Removed: .kanban/tasks/TASK-001.md

  Next step:
    git add -A && git commit -m "Remove TASK-001: Fix OAuth login bug"
```

### Step 6b: Delete -- User Aborts

```
$ kanban delete TASK-001

Delete  TASK-001  Fix OAuth login bug  [in-progress]?
Confirm [y/N]: n

Aborted. TASK-001 unchanged.
```

### Error: Not a git repository

```
$ kanban board

Error: Not a git repository.

  kanban stores task files in the current git repository.

  Initialise one here:
    git init
    kanban init
```

### Error: Commit references unknown task ID

```
-- CI output --

[kanban] Found task reference in commit: TASK-099
[kanban] Warning: TASK-099 not found in .kanban/tasks/
         Skipping status transition.
         Commit: 8cd22f0  by Rafael Rodrigues

  To create this task:
    kanban add "Task description" --id=TASK-099
```

---

## Integration Checkpoints

| Event | Trigger | Effect | Verified By |
|---|---|---|---|
| kanban add | CLI command | Task file created in .kanban/tasks/ | File present; kanban board shows task in todo |
| First commit with task ID | git commit message | task.status updated to in-progress in file | kanban board shows task in in-progress |
| CI pipeline passes with task ID in commit | CI hook/step | task.status updated to done in file | kanban board shows task in done |
| kanban edit | CLI command + $EDITOR | Task file fields updated, committed | kanban board shows updated metadata |
| kanban delete + confirm | CLI command | Task file removed | kanban board no longer shows task |

---

## Shared Artifact Sources

| Variable | Source of Truth | Consumers |
|---|---|---|
| task.id | Auto-generated on kanban add (e.g., TASK-001) | Filename, commit message references, all kanban commands |
| task.title | Task file field (set on add, updated on edit) | kanban board display, delete confirmation, CI log messages |
| task.status | Task file field (set by CLI or auto-updated by hooks) | kanban board column grouping |
| task.priority | Task file field (set on add, updated on edit) | kanban board display, sort/filter |
| task.due | Task file field | kanban board display, overdue indicator |
| task.assignee | Task file field | kanban board display, filter by assignee |
| task.description | Task file field | kanban edit / show only (not in board summary) |
| repo_root | git rev-parse --show-toplevel | .kanban/tasks/ path resolution for all commands |
| ci_task_pattern | Configurable regex in .kanban/config (default: TASK-[0-9]+) | Commit hook, CI step -- task ID extraction from commit messages |
