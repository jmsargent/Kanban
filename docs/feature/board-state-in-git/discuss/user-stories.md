<!-- markdownlint-disable MD024 -->
# User Stories: Board State in Git

**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## US-BSG-01: kanban log — Task Transition History

### Problem

Jon Santos is a developer who commits frequently and wants to understand the history of a task at a glance. He finds it frustrating to run `git log -- .kanban/tasks/TASK-007.md` and receive raw commit messages that do not speak in transition terms. He has to mentally reconstruct whether "chore: update TASK-007" means the task went in-progress or just had its description edited. The audit trail exists in git but is not surfaced.

### Who

- Jon Santos | Sole developer, Go CLI project | Wants to reconstruct task history in domain language without git expertise

### Solution

A new `kanban log <TASK-ID>` CLI command that reads git history for the task file and outputs a human-readable transition log. Phase 1 implementation wraps `git log --follow` on the task file and translates commit data into domain language. Status transitions (todo → in-progress → done) are derived from git history of status field changes.

### Domain Examples

#### 1: Happy Path — Complete History

Jon ran `kanban start TASK-007` three days ago, made several commits referencing it, and CI marked it done yesterday. He runs `kanban log TASK-007`. He sees:

```
TASK-007: Add retry logic to CI step

2026-03-15 09:14  todo → in-progress  jon@kanbandev.io
  trigger: kanban start

2026-03-17 16:42  in-progress → done  ci@kanbandev.io
  trigger: kanban ci-done (e91c44d)
```

He immediately knows who moved it and when, without touching git directly.

#### 2: Edge Case — No Transitions Yet

Jon runs `kanban log TASK-003` on a task he created yesterday but has not started. He sees:

```
TASK-003: Add dark mode

No transitions recorded yet.
Current status: todo
```

The command exits 0. No error. He is oriented — the task exists and is in todo.

#### 3: Error Case — Task Does Not Exist

Jon misremembers a task ID and runs `kanban log TASK-999`. He sees:

```
Error: task TASK-999 not found

  Try:
    kanban board      -- see all tasks
    kanban log TASK-007  -- view a specific task
```

Exit code 1. Clear guidance on what to do next.

### UAT Scenarios (BDD)

#### Scenario: View full transition history for a completed task

```gherkin
Given TASK-007 "Add retry logic to CI step" has been started and completed
And git history contains commits that touched .kanban/tasks/TASK-007.md
When Jon runs: kanban log TASK-007
Then the output header shows "TASK-007: Add retry logic to CI step"
And each transition shows timestamp, from-status, to-status, author, and trigger
And entries are sorted chronologically (oldest first)
And the exit code is 0
```

#### Scenario: View history for a task with no commits yet

```gherkin
Given TASK-003 "Add dark mode" exists in .kanban/tasks/
And no commits have referenced or modified .kanban/tasks/TASK-003.md
When Jon runs: kanban log TASK-003
Then the output shows "TASK-003: Add dark mode"
And the output shows "No transitions recorded yet."
And current status is shown as "todo"
And the exit code is 0
```

#### Scenario: Task not found returns actionable error

```gherkin
Given no task with id TASK-999 exists in .kanban/tasks/
When Jon runs: kanban log TASK-999
Then the output shows "Error: task TASK-999 not found"
And the output suggests "kanban board" as next step
And the exit code is 1
```

#### Scenario: Output uses domain language not raw git messages

```gherkin
Given TASK-007 has a git commit with message "kanban: start TASK-007"
When Jon runs: kanban log TASK-007
Then the output shows "todo → in-progress" using domain status terms
And the raw commit SHA is shown as supplementary context
And no raw git plumbing output appears in the result
```

#### Scenario: kanban log run outside a kanban repo

```gherkin
Given the current directory is not inside a git repo with kanban initialized
When Jon runs: kanban log TASK-001
Then the output shows an error about the repository not being initialized
And the output suggests "kanban init" as next step
And the exit code is 1
```

### Acceptance Criteria

- [ ] `kanban log <TASK-ID>` reads git history for the task file and produces formatted output
- [ ] Output header shows task ID and title
- [ ] Each transition entry shows: timestamp (local timezone), from-status, to-status, author, trigger
- [ ] Entries are sorted chronologically, oldest first
- [ ] When no commits exist for the task file, output says "No transitions recorded yet" with current status (todo)
- [ ] When task does not exist, exit code 1 with actionable error message
- [ ] When run outside a kanban repo, exit code 1 with "kanban init" suggestion
- [ ] Output uses domain vocabulary: "todo", "in-progress", "done" — not raw git object types
- [ ] In a local benchmark (5 runs, 95th percentile) response time is under 500ms on a repo with 1,000+ commits — @benchmark, not a CI gate

### Outcome KPIs

- **Who**: Jon Santos (developer using kanban-tasks)
- **Does what**: Views task transition history in domain language without running git plumbing commands
- **By how much**: 100% of history queries answered in 1 command (down from 0% — no `kanban log` exists today)
- **Measured by**: Developer self-report after 2-4 weeks of use; observation of whether `git log -- .kanban/tasks/` usage decreases
- **Baseline**: 0 — no `kanban log` command exists; developers must run git plumbing manually

### Technical Notes

- Driving port: CLI adapter (`internal/adapters/cli/kanban_log.go`)
- New use case: `GetTaskHistory.Execute(repoRoot, taskID string)` in `internal/usecases/`
- Git port: uses existing `GitPort` — requires a new method `LogFile(repoRoot, filePath string) ([]CommitEntry, error)`
- **Constraint**: `internal/domain` must not import GitPort. History is assembled in the use case layer.
- Phase 1: wraps `git log --follow --format='...' -- .kanban/tasks/<id>.md`
- Phase 2 (US-BSG-02): switches to reading `.kanban/transitions.log` instead of git log
- No new ADRs required for Phase 1; Phase 2 requires ADR update

---

## US-BSG-02: Append-Only Transitions Log (Concept B)

### Problem

Jon Santos is a developer who values clean architecture. He finds it frustrating that the commit-msg hook mutates `.kanban/tasks/*.md` files as a side effect of his commits. When he runs `git diff`, status field rewrites appear alongside real code changes — task-state housekeeping noise mixed with meaningful work. He also knows that running `git rebase -i` on branches could theoretically corrupt the state if it were stored in commit trailers. He wants task state to live in a dedicated place, not scattered across task definition files.

### Who

- Jon Santos | Developer who commits frequently and rebases regularly | Wants state and definition concerns cleanly separated

### Solution

Replace `status:` field in task YAML front matter with an append-only `.kanban/transitions.log` file. Each line records one transition: `<ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>`. The hook appends to this file instead of mutating task files. `kanban board` derives current status from the log. Task files store definition only: id, title, priority, due, assignee, description.

### Domain Examples

#### 1: Happy Path — Starting Work

Jon runs `kanban start TASK-007`. He observes:
- `.kanban/tasks/TASK-007.md` is NOT modified
- `.kanban/transitions.log` gains one line: `2026-03-15T09:14:23Z TASK-007 todo->in-progress jon@kanbandev.io manual`
- `kanban board` shows TASK-007 in IN PROGRESS
- `git diff` after the next commit shows one log line appended, no YAML mutation

#### 2: Edge Case — Rebase Safety

Jon rebases three commits (`git rebase -i HEAD~3`), squashing them. He then checks `kanban board`. TASK-007 is still in IN PROGRESS because the transition was recorded in `transitions.log` — a separate file that the rebase did not touch. His state history is intact.

#### 3: Error Case — Hook Failure

The `.kanban/transitions.log` file has been accidentally made unwritable (permissions issue). Jon commits referencing TASK-007. The commit succeeds (exit code 0). He sees a warning on stderr: `kanban: could not record transition for TASK-007: permission denied`. The commit is not blocked. He resolves the permissions issue and manually records the transition or reruns.

### UAT Scenarios (BDD)

#### Scenario: Task created without status field

```gherkin
Given kanban is initialized with Concept B enabled
When Jon runs: kanban add -t "Add retry logic to CI step" -p high
Then .kanban/tasks/TASK-007.md is created
And the YAML front matter does NOT contain a "status:" field
And .kanban/tasks/TASK-007.md contains a comment pointing to transitions.log
And kanban board shows TASK-007 in the TODO column
And the exit code is 0
```

#### Scenario: kanban start appends to log without modifying task file

```gherkin
Given TASK-007 exists with no entries in .kanban/transitions.log
When Jon runs: kanban start TASK-007
Then .kanban/transitions.log gains exactly one new line
And the line contains: TASK-007, todo->in-progress, jon@kanbandev.io, trigger:manual
And .kanban/tasks/TASK-007.md is NOT modified
And kanban board shows TASK-007 in the IN PROGRESS column
And the exit code is 0
```

#### Scenario: Hook appends to log and does not write task files

```gherkin
Given TASK-007 is in-progress per .kanban/transitions.log
When Jon commits with message "refactor: add retry logic TASK-007"
Then .kanban/transitions.log gains exactly one new line for TASK-007
And .kanban/tasks/TASK-007.md is NOT modified by the hook
And the commit exits with code 0
```

#### Scenario: Hook exits 0 even when log write fails

```gherkin
Given .kanban/transitions.log is not writable due to permissions
When Jon commits with message "fix: something TASK-007"
Then the commit exits with code 0
And stderr contains a warning about the failed transition record
And the commit message is unchanged
```

#### Scenario: kanban board derives status from transitions log

```gherkin
Given .kanban/transitions.log has entries showing TASK-007 as done and TASK-005 as in-progress
And TASK-003 has no entries in the log
When Jon runs: kanban board
Then TASK-007 appears in the DONE column
And TASK-005 appears in the IN PROGRESS column
And TASK-003 appears in the TODO column
```

#### Scenario: State is preserved after git rebase

```gherkin
Given TASK-007 has a transition entry in .kanban/transitions.log
When Jon performs git rebase -i HEAD~3 squashing the three most recent commits
Then .kanban/transitions.log still contains the original TASK-007 entry
And kanban board still shows TASK-007 in the correct column
And kanban log TASK-007 still shows the transition history
```

#### Scenario: kanban ci-done commits only transitions.log

```gherkin
Given TASK-007 is in-progress per .kanban/transitions.log
And the CI commit range contains commits referencing TASK-007
When kanban ci-done runs
Then .kanban/transitions.log gains a done entry for TASK-007
And the git commit staged by ci-done contains only .kanban/transitions.log
And no .kanban/tasks/*.md files appear in the ci-done commit
And the exit code is 0
```

### Acceptance Criteria

- [ ] `.kanban/transitions.log` is created by `kanban init` if it does not exist
- [ ] Each line in the log follows the format: `<ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>`
- [ ] `kanban start <TASK-ID>` appends to log; does NOT write to the task file
- [ ] The commit-msg hook appends a commit reference to the log; does NOT write to task files
- [ ] `kanban board` derives all task statuses from the log (not from YAML front matter)
- [ ] Tasks with no log entries are displayed in the TODO column
- [ ] `kanban ci-done` commits only `.kanban/transitions.log` (no task files in the commit)
- [ ] The hook exits 0 in all cases; log write failure emits a stderr warning only
- [ ] Concurrent appends to the log by two processes are safe (file locking)
- [ ] The log is rebase-safe: rebasing commits does not alter log entries
- [ ] New task files created after Concept B ships do NOT contain a `status:` field
- [ ] A comment in each new task file points to `.kanban/transitions.log` for state

### Outcome KPIs

- **Who**: Jon Santos (developer committing to a kanban-tasks repo)
- **Does what**: Commits that reference tasks produce zero task file mutations; the diff shows only code changes and a single log append
- **By how much**: Files modified by a state transition: 1 (transitions.log only), down from 1 per transitioned task file
- **Measured by**: `git show <commit>` on any hook-fired commit; count of `.kanban/tasks/` files in the diff
- **Baseline**: Current — every hook-fired commit mutates the task file (1 file per transitioned task)

### Technical Notes

- New secondary port: `TransitionLogRepository` in `internal/ports/` with methods: `Append(repoRoot string, entry TransitionEntry) error`, `LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error)`, `History(repoRoot, taskID string) ([]TransitionEntry, error)`
- New domain type: `TransitionEntry` in `internal/domain/` — fields: Timestamp, TaskID, From, To, Author, Trigger
- New filesystem adapter: implements `TransitionLogRepository` — atomic appends using `os.OpenFile` with `O_APPEND|O_CREATE|O_WRONLY` + advisory file locking
- Modify `GetBoard.Execute`: replaces `task.Status` (from YAML) with `log.LatestStatus(taskID)` for each task
- Modify `StartTask.Execute`: append to log instead of `tasks.Update()`
- Modify commit-msg hook handler: append to log instead of `tasks.Update()`
- Modify `TransitionToDone.Execute` (ci-done): append to log; commit log file only (not task files)
- Remove `status:` field from `taskFrontMatter` struct in filesystem adapter
- **Dependencies**: US-BSG-01 must be shipped and validated before US-BSG-02 begins
- **ADRs**: ADR-002, ADR-004, ADR-005 require updates; new ADR for transitions log format required

---

## US-BSG-03: kanban board --me Filter

### Problem

Jon Santos is working on a project where he and two colleagues each have kanban-tasks initialized. When he runs `kanban board`, he sees 15 tasks across all three developers. He has to scan the board mentally to find his 4 tasks. As the project grows, this scanning becomes slower. He wants a "show me my tasks" view that uses his existing git identity without additional configuration.

### Who

- Jon Santos (and any developer on a multi-contributor project) | Daily `kanban board` user | Wants to focus on their own work without manual filtering

### Solution

Add a `--me` flag to `kanban board` that filters displayed tasks to those where the `assignee` field matches the current `git config user.email`. When unassigned tasks exist in the repo, a warning is shown so hidden tasks are never silently lost.

### Domain Examples

#### 1: Happy Path — Filtered Board

Jon runs `kanban board --me`. His git identity is `jon@kanbandev.io`. TASK-005 and TASK-007 have `assignee: jon@kanbandev.io`. The board shows only those two tasks. Other tasks are not shown.

#### 2: Edge Case — Unassigned Tasks Warning

Maria Santos (`maria@kanbandev.io`) runs `kanban board --me`. TASK-012 has no assignee field. The board shows Maria's tasks and a warning: `1 unassigned task hidden — use 'kanban board' to see all`. TASK-012 is not silently lost.

#### 3: Error Case — No Matching Tasks

Jon runs `kanban board --me` on a fresh repo where he has not started any tasks. The board shows three empty columns and a note: `No tasks assigned to jon@kanbandev.io. Use 'kanban board' to see all tasks.` Exit code 0 — not an error condition.

### UAT Scenarios (BDD)

#### Scenario: Board filters to current developer's tasks

```gherkin
Given the board has 5 tasks total
And TASK-005 and TASK-007 have assignee: jon@kanbandev.io
And TASK-001, TASK-003, TASK-006 have different assignees
When Jon runs: kanban board --me
Then the output shows only TASK-005 and TASK-007
And TASK-001, TASK-003, TASK-006 are not shown
And the exit code is 0
```

#### Scenario: --me warns about unassigned tasks

```gherkin
Given TASK-008 exists with no assignee field
When Jon runs: kanban board --me
Then the output shows a warning: "1 unassigned task hidden — use 'kanban board' to see all"
And TASK-008 is not shown in the filtered output
```

#### Scenario: --me shows empty board gracefully when no tasks are assigned

```gherkin
Given no tasks in the repo have assignee: jon@kanbandev.io
When Jon runs: kanban board --me
Then the board shows empty columns
And a message says: "No tasks assigned to jon@kanbandev.io"
And suggests: "use 'kanban board' to see all tasks"
And the exit code is 0
```

### Acceptance Criteria

- [ ] `kanban board --me` filters tasks by `assignee` field matching `git config user.email`
- [ ] Unassigned tasks trigger a warning count: "N unassigned task(s) hidden — use 'kanban board' to see all"
- [ ] When no tasks match the current identity, output shows empty board with a helpful message
- [ ] `kanban board` (without --me) is unaffected — shows all tasks as before
- [ ] The flag works with both Phase 1 (YAML status) and Phase 2 (transitions log) state sources

### Outcome KPIs

- **Who**: Developers on multi-contributor projects (2+ assignees)
- **Does what**: View their own tasks without scanning the full board
- **By how much**: Time to locate own tasks: from O(total tasks scan) to O(1) visual scan
- **Measured by**: Developer self-report; feature adoption rate (% of `kanban board` invocations using `--me`)
- **Baseline**: 0% — `--me` flag does not exist today

### Technical Notes

- Modify `GetBoard.Execute`: accept optional `filterAssignee string` parameter
- CLI adapter: add `--me` flag to `board` command; populate `filterAssignee` from `GitPort.GetIdentity().Email`
- No new ports required — uses existing `TaskRepository` and `GitPort`
- Works with both Phase 1 (status from YAML) and Phase 2 (status from transitions log) — the board derivation is orthogonal to the filtering
- Dependency: NONE — can ship independently of US-BSG-01 and US-BSG-02
