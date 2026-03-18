# Acceptance Criteria — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

This document consolidates all acceptance criteria derived from UAT scenarios. Each criterion is traceable to a story and a specific scenario. All criteria describe observable, testable user outcomes.

---

## US-BSG-01: kanban log TASK-ID

Source scenarios: 5 UAT scenarios in user-stories.md

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-01-1 | `kanban log <TASK-ID>` prints the task title as a header on the first line | Happy path | Yes — assert stdout line 1 |
| AC-01-2 | Each transition entry in the output includes: timestamp, from-status arrow to-status, author email, and trigger | Happy path | Yes — assert output format per line |
| AC-01-3 | Entries are sorted chronologically, oldest first | Happy path (3 commits) | Yes — assert timestamp ordering |
| AC-01-4 | When no commits reference the task file, output shows "No transitions recorded yet." and current status "todo" | No-commits edge case | Yes — assert exact string and exit 0 |
| AC-01-5 | When the task does not exist, exit code is 1 and output includes "task TASK-999 not found" | Task not found error | Yes — assert exit code and stderr/stdout |
| AC-01-6 | When the task does not exist, output suggests `kanban board` as a next step | Task not found error | Yes — assert suggestion text |
| AC-01-7 | When run outside a kanban-initialized repository, exit code is 1 | Not-in-repo error | Yes — assert exit code |
| AC-01-8 | When run outside a kanban-initialized repository, output suggests `kanban init` | Not-in-repo error | Yes — assert suggestion text |
| AC-01-9 | Output uses "todo", "in-progress", "done" (domain terms); no raw git object IDs in status fields | Domain language | Yes — assert no raw SHA in status position |
| AC-01-10 | Commit SHA is shown as supplementary context (not as the primary transition descriptor) | Domain language | Yes — assert SHA appears after the domain transition |
| AC-01-11 | In a local benchmark of 5 runs on standard developer hardware, 95th percentile response time for `kanban log` on a repo with 1,000+ commits is under 500ms | Performance (benchmark) | @benchmark — local only, NOT executed in CI; time.Now()-based benchmark test |

---

## US-BSG-02: Append-Only Transitions Log

Source scenarios: 7 UAT scenarios in user-stories.md

### Task Creation

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-1 | `kanban add` creates a task file WITHOUT a `status:` field in YAML front matter | Task creation | Yes — assert YAML does not contain "status:" |
| AC-02-2 | The task file contains a comment pointing to `.kanban/transitions.log` | Task creation | Yes — assert comment text present |
| AC-02-3 | `kanban board` shows the newly created task in the TODO column with no log entries | Task creation | Yes — assert board output |

### kanban start

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-4 | `kanban start TASK-007` appends exactly one line to `.kanban/transitions.log` | kanban start | Yes — assert line count delta = 1 |
| AC-02-5 | The appended line contains: TASK-007, `todo->in-progress`, author email, trigger:manual | kanban start | Yes — assert line content |
| AC-02-6 | `.kanban/tasks/TASK-007.md` is NOT modified after `kanban start` | kanban start | Yes — assert file mtime unchanged |
| AC-02-7 | `kanban board` shows the task in IN PROGRESS after `kanban start` | kanban start | Yes — assert board column |
| AC-02-8 | Running `kanban start` on an already in-progress task exits 1 with an "already in-progress" message | Idempotency | Yes — assert exit code and message |
| AC-02-9 | Running `kanban start` on an already in-progress task does NOT add a duplicate log entry | Idempotency | Yes — assert log line count unchanged |

### Commit-Msg Hook

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-10 | Hook appends one line to `.kanban/transitions.log` when a commit references a task | Hook appends | Yes — integration test with real git |
| AC-02-11 | Hook does NOT modify `.kanban/tasks/*.md` files | Hook does not write task files | Yes — assert no task file mtime change |
| AC-02-12 | Hook exits 0 when log write succeeds | Hook normal | Yes — assert exit code |
| AC-02-13 | Hook exits 0 even when the log file is unwritable | Hook failure safety | Yes — make log unwritable, assert exit 0 |
| AC-02-14 | When the hook cannot write the log, it prints a warning to stderr | Hook failure safety | Yes — assert stderr contains "could not record transition" |
| AC-02-15 | When the hook cannot write the log, the commit message is not modified | Hook failure safety | Yes — assert commit message unchanged |

### kanban board

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-16 | `kanban board` places tasks in columns based on their most recent log entry, not YAML status field | Board derives from log | Yes — modify only log, assert board column |
| AC-02-17 | Tasks with no log entries appear in TODO column | Board derives from log (implicit todo) | Yes — task with no entries → assert TODO |

### kanban ci-done

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-18 | `kanban ci-done` appends a done entry to `.kanban/transitions.log` for each in-progress task in the commit range | ci-done commits log | Yes — integration test |
| AC-02-19 | The git commit staged by `kanban ci-done` contains ONLY `.kanban/transitions.log` | ci-done commits log only | Yes — assert `git show` lists only transitions.log |
| AC-02-20 | No `.kanban/tasks/*.md` files appear in the `kanban ci-done` commit | ci-done commits log only | Yes — assert git diff excludes tasks/ |

### Rebase Safety

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-21 | After `git rebase -i` squashing N commits, all log entries for tasks referenced in those commits are still present | Rebase safety | Yes — integration test with real git rebase |
| AC-02-22 | `kanban board` shows correct status after a rebase | Rebase safety | Yes — assert board after rebase |

### Concurrency

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-02-23 | Two simultaneous appends to `.kanban/transitions.log` produce exactly 2 new lines without corruption | Concurrent writes | Yes — goroutine race test |
| AC-02-24 | Each appended line is complete (no truncated lines) after concurrent writes | Concurrent writes | Yes — property-based goroutine race test with sync.WaitGroup; for any N concurrent appends, file has exactly N new lines |
| AC-02-25 | When `.kanban/transitions.log` references a task ID for which no task file exists (task deleted), the task is excluded from `kanban board` without causing an error | Deleted task with log entries | Yes — delete task file, assert board excludes it and exits 0 |
| AC-02-26 | When `kanban ci-done` runs and the commit range contains no in-progress task references, the command exits 0 and outputs "no tasks to transition" | ci-done with no matching tasks | Yes — assert exit code 0 and output message |

---

## US-BSG-03: kanban board --me

Source scenarios: 3 UAT scenarios in user-stories.md

| # | Criterion | Source Scenario | Testable? |
|---|-----------|----------------|-----------|
| AC-03-1 | `kanban board --me` shows only tasks where `assignee` matches current `git config user.email` | Filtered board | Yes — assert only matching tasks appear |
| AC-03-2 | Tasks with different assignees are not shown when `--me` is used | Filtered board | Yes — assert non-matching tasks absent |
| AC-03-3 | When unassigned tasks exist, a warning is shown: "N unassigned task(s) hidden — use 'kanban board' to see all" | Unassigned warning | Yes — assert warning text and count |
| AC-03-4 | Unassigned tasks are not shown in `--me` output (they are in the warning count only) | Unassigned warning | Yes — assert unassigned task not in output |
| AC-03-5 | When no tasks match the current git identity, board shows empty columns with a helpful message | Empty filtered board | Yes — assert message text |
| AC-03-6 | `kanban board` without `--me` is unaffected (shows all tasks) | Regression | Yes — assert all tasks present |
| AC-03-7 | `--me` works regardless of Phase 1 or Phase 2 state storage | Phase compatibility | Yes — test both configurations |
