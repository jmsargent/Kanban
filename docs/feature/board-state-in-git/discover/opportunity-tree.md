# Opportunity Tree — Board State in Git
**Feature**: board-state-in-git
**Phase**: 2 — Opportunity Mapping
**Status**: COMPLETE
**Date**: 2026-03-18

---

## Desired Outcome (from Phase 1)

> Developers experience kanban-tasks as an extension of their natural git workflow — task *state* derives from git history rather than being written as a side effect into separate files, so the repository stays clean, transitions are auditable, and state is implicitly scoped to the person who committed.

---

## Opportunity Scoring

**Formula**: Score = Importance + Max(0, Importance - Satisfaction)
**Scale**: Importance 1-10 | Satisfaction with current solution 1-10 | Max score 20

Scores derived from Interview 1 signals and codebase analysis.

| ID | Opportunity | Importance | Satisfaction (current) | Score | Priority |
|----|------------|------------|----------------------|-------|----------|
| O1 | Know when each task transitioned (audit trail) | 9 | 2 | 16 | TOP |
| O2 | No hook side-effects writing to task files | 8 | 3 | 13 | TOP |
| O3 | Git log as single source of truth for state | 8 | 2 | 14 | TOP |
| O4 | Per-user board view derived from git authorship | 6 | 2 | 10 | HIGH |
| O5 | Fewer files modified per commit (less noise in diff) | 6 | 4 | 8 | HIGH |
| O6 | Merge conflict elimination on task state | 5 | 6 | 4 | LOW |

**Top opportunities**: O1 (16), O3 (14), O2 (13)

---

## Opportunity Solution Tree

```
Desired Outcome: Task state derives from git history; repository stays clean; transitions auditable

  |
  +-- O1: Know when each task transitioned (score: 16)
  |     +-- S1a: Commit trailers as state events (Option A)
  |     |         Kanban-Task: TASK-001=in-progress in each commit
  |     +-- S1b: Append-only transition log file (.kanban/transitions.log)
  |     |         Git tracks history of this file natively
  |     +-- S1c: State derived from commit message pattern + timestamp
  |               (current partial approach, made explicit)
  |
  +-- O3: Git log as single source of truth for state (score: 14)
  |     +-- S3a: Commit trailers — full event-sourced state (Option A)
  |     +-- S3b: Hybrid — task files retain definition fields only; state field removed
  |     +-- S3c: Separate orphan branch (refs/kanban/state) as state store
  |
  +-- O2: No hook side-effects writing to task files (score: 13)
  |     +-- S2a: Hook becomes read-only (parses message, emits output) — no file writes
  |     +-- S2b: Eliminate hook entirely; state derived lazily on `kanban board`
  |     +-- S2c: Hook writes to a separate append-only log, not to task files
  |
  +-- O4: Per-user board view (score: 10)
  |     +-- S4a: `kanban board --author=<name>` filters by git commit author
  |     +-- S4b: Implicit: if state lives in commits, `board` defaults to current git identity
  |
  +-- O5: Fewer files modified per commit (score: 8)
        +-- S5a: Remove status field from task file YAML entirely
        +-- S5b: Status field kept but marked "derived — do not edit directly"
```

---

## Critical Constraint Analysis

Before scoring solutions, four hard constraints from the existing codebase must be mapped against each solution:

### Constraint 1: The binary never auto-commits (except `kanban ci-done`)
**Source**: CLAUDE.md non-negotiable constraint
**Impact on Option A (commit trailers)**: If task state is recorded *in* commits, then `kanban add` (creating a task as "todo") and `kanban start` (moving to "in-progress") would need to either: (a) make a git commit, or (b) not record state in git at all until the developer's next natural commit. Option (b) means pure-todo tasks have no commit — they cannot exist in the commit-trailer model without a commit. This is a fundamental tension.

### Constraint 2: Tasks can exist in "todo" state before any commit
**Source**: `internal/usecases/add_task.go` — `AddTask.Execute` calls `tasks.Save`, no git involvement
**Impact**: Under Option A, a task created with `kanban add` starts as "todo." If state lives in commits, this task has no commit yet — where does its existence live? It must still live in the task file, meaning task files cannot be fully eliminated as a state store. The task file IS still needed for definition + initial existence.

### Constraint 3: `kanban board` performance — O(commits) vs O(tasks)
**Source**: `internal/usecases/get_board.go` — currently calls `tasks.ListAll` (directory scan, O(n files))
**Impact**: Under Option A, `kanban board` must replay `git log` to derive current state. On a repo with 10,000 commits, this requires parsing every commit message or trailer. Even with `--grep` filtering, this is O(commits) not O(tasks). For a project with 50 tasks and 10,000 commits, this is ~200x more work than the current approach. Mitigation: cache the derived state to a local file (but then you have a cache invalidation problem and a file again).

### Constraint 4: Rebase and amend destroy state history
**Source**: Git fundamentals
**Impact**: Commit trailers are attached to commit SHAs. `git rebase -i` rewrites SHAs. After a rebase, the commit trailer `Kanban-Task: TASK-001=in-progress` is still in the message, but the commit SHA changed — any system that tracks state by SHA is broken. If state is tracked by message content only (not SHA), rebase is safe, but amending a commit message that contained a trailer would alter historical state.

---

## Feasibility Assessment by Solution

### Option A — Full commit-trailer event log (S1a + S3a combined)

**Model**: Every state transition appends a trailer to the commit that triggered it. State is derived by replaying `git log --format='%(trailers:key=Kanban-Task)'`.

| Question | Answer | Severity |
|----------|--------|----------|
| How does `kanban add` create a todo task? | Must still write a task file (definition) — commit not required for creation | Manageable |
| How does `kanban start` record in-progress without auto-committing? | It cannot — either `kanban start` makes a commit (violates constraint) or state is embedded in the *developer's next commit* via the hook | CRITICAL |
| How does `kanban board` derive state? | Replay `git log` for all trailers — O(commits) | CRITICAL at scale |
| What happens on `git rebase`? | Trailers survive if message is preserved — safe for rebase that keeps messages; broken by interactive rebase that edits messages | HIGH risk |
| What happens on `git commit --amend`? | If trailer is in the amended message, history is silently altered | HIGH risk |
| Where do pure-todo tasks live? | Still in task files — task files cannot be eliminated | Manageable |
| How does `kanban ci-done` work? | `ci-done` currently writes task files and makes a commit. Under Option A, the "done" transition would already be recorded in the CI commit's trailer — the auto-commit from `ci-done` could be eliminated | POSITIVE — simplification |
| What is the `kanban board` latency on 10k commits? | Potentially 500ms-2s without caching — unacceptable for interactive use | HIGH risk |

**Verdict**: Option A as a *complete replacement* for file-based state is not viable. Too many operations have no commit to attach to. However, Option A as a *complement* — trailers recording transition events on commits that reference tasks — is viable and already partially implemented (the hook fires on commit-msg).

### Hybrid Approach (S3b — most viable)

**Model**: Task files retain definition fields (id, title, priority, due, assignee, description) and the `status` field is **removed** from task files. State is derived from git log for any task that has commits. Tasks with no commits are implicitly "todo." The hook records transitions via trailers; no file writes occur from the hook.

| Question | Answer | Severity |
|----------|--------|----------|
| How does `kanban add` work? | Creates task file without `status` field — todo is implicit | Clean |
| How does `kanban start` work? | Writes a commit with trailer `Kanban-Task: TASK-001=in-progress` — but this requires an auto-commit | CRITICAL — violates constraint |
| How does `kanban board` work? | Scans task files for definitions + replays git log for state — O(tasks + commits) | Manageable with caching |
| How does the hook work? | On commit with TASK-NNN ref, adds trailer automatically — no file writes | Clean |
| What about `kanban ci-done`? | CI commit gets trailer `Kanban-Task: TASK-001=done` — no separate status commit needed | POSITIVE |
| Audit trail? | `git log --format='%(trailers:key=Kanban-Task)'` gives full history | SOLVED |
| Per-user view? | `git log --author=<name>` gives all transitions by that user | SOLVED |

**Remaining blocker**: `kanban start` (explicit manual start) has no commit to attach a trailer to. Under this model, `kanban start` would need to either (a) make a git commit, (b) stage a trailer for the *next* commit, or (c) be removed in favor of "state is implicit until you commit."

### Append-Only Log Approach (S1b)

**Model**: A single `.kanban/transitions.log` file is append-only. Each line: `2026-03-18T14:32:00Z TASK-001 todo->in-progress author@example.com`. The hook appends to this file. Git tracks history via `git log -- .kanban/transitions.log`.

| Question | Answer | Severity |
|----------|--------|----------|
| Audit trail? | Fully solved — git blame + log on the file | SOLVED |
| `kanban board` performance? | Read task files (O(n tasks)) + read log file (O(n transitions)) | Excellent |
| Hook behavior? | Appends one line — simpler than current status field rewrite | Improvement |
| Merge conflicts? | Two developers appending to the same file simultaneously — conflict risk exists but is trivially resolved (both lines are valid, keep both) | Low risk |
| Rebase safety? | Log file is a separate file; not affected by commit rebase | SAFE |
| `kanban start`? | Appends to log file — no commit required | SOLVED |
| `kanban ci-done`? | Appends to log + commits the log file — simpler than committing all task files | Improvement |
| Does this "clutter the repository"? | One file instead of n status-field rewrites — significantly less noise | IMPROVEMENT |

**Verdict**: S1b is architecturally simpler than Option A and solves all stated problems without any of the hard blockers.

---

## Prioritized Opportunities Summary

| Rank | Opportunity | Best Solution | Feasibility |
|------|------------|---------------|-------------|
| 1 | Audit trail (O1) | S1b (transition log) or commit trailers via hook | High |
| 2 | Git log as state authority (O3) | Hybrid: files for definition, log/trailers for state | Medium |
| 3 | No hook file-writes (O2) | S2c (hook writes to log, not task files) | High |
| 4 | Per-user view (O4) | S4a (--author filter on board command) | High |
| 5 | Less repo noise (O5) | S5a (remove status field from task files) | Medium |
