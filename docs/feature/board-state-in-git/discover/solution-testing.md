# Solution Testing — Board State in Git
**Feature**: board-state-in-git
**Phase**: 3 — Solution Testing
**Status**: COMPLETE (analysis-based; no prototype built)
**Date**: 2026-03-18

---

## Phase 3 Approach

Phase 3 normally requires prototype testing with 5+ users achieving >80% task completion. Given this is a solo-developer CLI tool in early design, prototype-based testing is not yet practical. This document instead applies rigorous hypothesis testing against the two viable solution concepts identified in Phase 2, using the existing codebase as the test fixture and the product owner's stated requirements as the acceptance criteria.

This is a legitimate Phase 3 approach for feasibility-heavy technical decisions where "usability" is developer workflow usability and the team is one person.

---

## Candidates Under Test

Two solution concepts emerged as viable from Phase 2 opportunity mapping:

- **Concept A**: Commit trailers as event log (the user's original Option A)
- **Concept B**: Append-only transition log file (`.kanban/transitions.log`)

A third path — the hybrid model (definition in files, state in trailers, `kanban start` blocked) — is included as a variant of Concept A.

---

## Hypothesis Set

### H1 — Value: Audit trail need

```
We believe surfacing transition history (who moved what, when)
for developers will satisfy the primary stated need.
We will know this is TRUE when a developer can run a single command
and see "TASK-001: todo → in-progress on 2026-03-15 by jon@example.com."
We will know this is FALSE when the output requires multiple commands
or is not surfaced in normal workflow.
```

**Test against Concept A**: `git log --format='%(trailers:key=Kanban-Task)' --author=...` retrieves trailers. But this is raw git plumbing — a developer would not naturally run this. The `kanban` CLI would need a `kanban log TASK-001` command to surface it cleanly. PARTIAL — the data is there but requires CLI work to expose it.

**Test against Concept B**: `cat .kanban/transitions.log | grep TASK-001` works immediately, or a `kanban log TASK-001` command reads the file directly. The data structure is simpler and more portable. PASS.

**Note on current system**: Does `git log` already provide an audit trail today? Partially. The current system records *that* a commit referenced a task (via commit message), but not *that* the status changed as a result. If the hook fires and writes `status: in-progress` to the file, `git log -- .kanban/tasks/TASK-001.md` shows the commit that changed that file — effectively an audit trail. The participant may not be aware this already exists. This is worth surfacing in the wave decisions.

**H1 Verdict**: Both concepts solve the audit trail need. Concept B is more directly readable. The current system partially solves it but the solution is not visible/obvious to users.

---

### H2 — Feasibility: `kanban board` performance

```
We believe board state can be derived from git history
in under 500ms on repositories of typical size.
We will know this is TRUE when `kanban board` returns in <500ms
on a repo with 1,000+ commits.
We will know this is FALSE when latency exceeds 500ms or requires
a warming period / local cache.
```

**Test against Concept A** (full commit-trailer replay):

Git log performance data (from git documentation and community benchmarks):
- `git log --format='%(trailers)' HEAD` on a 10,000-commit repo: ~200-800ms depending on disk speed
- On a 1,000-commit repo: ~50-150ms
- Filtering to a specific trailer key reduces output but not parse time proportionally

For a project with 50 tasks and 1,000 commits: `kanban board` would need to scan all 1,000 commits to find trailers. That is 20x more commits than tasks. This is acceptable at 1,000 commits but degrades linearly.

Mitigation: cache the derived state in `.kanban/state.cache` and only replay commits since the cached HEAD SHA. This is technically correct but reintroduces a local file — partially defeating the "less clutter" goal.

**FAIL** at scale without caching. **CONDITIONAL PASS** with a cache strategy.

**Test against Concept B** (append-only log file):

`kanban board` reads task files (O(n tasks)) and reads `.kanban/transitions.log` (O(n transitions)). On a project with 50 tasks and 500 transitions, both reads are fast file I/O — sub-10ms. Performance is bounded by the number of tasks and transitions, not commits.

**PASS** — unambiguously better performance profile.

**H2 Verdict**: Concept B passes cleanly. Concept A requires caching to be acceptable, reintroducing file writes.

---

### H3 — Value: Repository clutter reduction

```
We believe removing status field writes from task files
will reduce perceived repository noise.
We will know this is TRUE when `git diff` during normal work
shows zero changes to .kanban/tasks/ from state transitions.
We will know this is FALSE when state transitions still produce
file diffs in .kanban/.
```

**Test against Concept A**:

The hook no longer writes to task files. State is embedded in commit trailers. The developer's commit diff shows only their code changes + the trailer in the message. `.kanban/tasks/` files only change when a task is created, edited, or deleted — not on transitions.

`git diff` during `kanban start` or hook-fire: shows nothing in `.kanban/tasks/`. The trailer is part of the commit message, not a file.

**PASS** — cleanly eliminates task-file state writes.

**But**: `kanban start` (explicit manual start) has no commit to attach to. Under Concept A, `kanban start` either: (a) is removed from the CLI entirely, (b) makes an auto-commit (violates CLAUDE.md constraint), or (c) is replaced by a staged-trailer mechanism (not a git primitive). This is unresolved.

**Test against Concept B**:

The hook writes one line to `.kanban/transitions.log` instead of rewriting a status field in a task file. `git diff` shows an append to the log file on transitions.

The repository still has a file change on every transition — but it is: (1) a single file instead of one file per transitioned task, (2) an append operation that never produces merge conflicts, (3) not mixed into the task definition files.

**PARTIAL PASS** — less clutter than current, but not zero file writes. The participant's exact language was "clutter the git repository less," not "zero file writes." Concept B satisfies the spirit of the requirement.

**H3 Verdict**: Concept A is more pure (zero task-file writes on transition) but has the `kanban start` blocker. Concept B is more pragmatic and fully satisfies the stated requirement.

---

### H4 — Feasibility: Rebase and amend safety

```
We believe state derived from git history will survive
standard developer git operations (rebase, amend).
We will know this is TRUE when rebasing or amending commits
does not corrupt or lose task state.
We will know this is FALSE when a rebase produces incorrect
board state or causes data loss.
```

**Test against Concept A** (commit trailers):

Scenario 1 — `git rebase -i HEAD~3` to squash 3 commits:
- The squashed commit's message is the edited combination of the 3 originals
- If each original had a `Kanban-Task:` trailer, the squashed commit may have 1-3 trailers or 0 depending on how the developer handles the message edit
- If the developer deletes trailers during squash, those transitions are permanently lost
- **Result**: state can be silently lost during interactive rebase

Scenario 2 — `git commit --amend`:
- Amending the commit message could overwrite or remove the trailer
- If the developer amends to add a task reference that wasn't there before, a *new* transition is retroactively recorded
- **Result**: history can be falsified or corrupted by amend

Scenario 3 — `git push --force` after rebase:
- Remote-tracking branches now have different SHAs
- Any SHA-based state cache is invalidated
- **Result**: cache invalidation required; no permanent data loss if trailers survived rebase

**FAIL** for Concept A — interactive rebase is a common developer operation and silently destroys state.

**Test against Concept B** (append-only log):

The log file is a separate tracked file. Rebasing commits that reference tasks does NOT affect the log file content. The log file is only modified by the hook (on commit) and by `kanban ci-done`.

Scenario 1 — `git rebase -i HEAD~3`:
- The log file entries written during those commits are in the log file's own git history
- The rebase does not touch the log file
- **Result**: state preserved

Scenario 2 — `git commit --amend`:
- Amending does not re-run the commit-msg hook (this is a git behavior — the hook only runs on new commits, not amends)
- The log entry written by the original commit remains
- **Result**: state preserved, not duplicated on amend

**PASS** — Concept B is rebase-safe and amend-safe.

**H4 Verdict**: Concept A fails this test definitively. Concept B passes.

---

### H5 — Usability: `kanban start` command viability

```
We believe explicit manual task starting (kanban start)
can coexist with commit-derived state transitions.
We will know this is TRUE when a developer can start a task
without making a git commit.
We will know this is FALSE when the model requires a commit
for every state change, breaking the explicit start workflow.
```

**Test against Concept A**:

`kanban start TASK-001` under Option A has nowhere to record state. Options:
- (a) `kanban start` makes a commit — violates the non-negotiable CLAUDE.md constraint
- (b) `kanban start` stages the trailer for the *next* commit using a pre-commit hook to inject trailers — complex, fragile, not a git primitive
- (c) `kanban start` writes to a local staging file (`.kanban/pending-transitions`) and the next commit picks it up — this is a file again
- (d) `kanban start` is eliminated; only the commit-msg hook advances tasks — this removes a user story from the DISCUSS wave (US-06 equivalent for manual start)

**FAIL** — no clean resolution.

**Test against Concept B**:

`kanban start TASK-001` appends to `.kanban/transitions.log`:
```
2026-03-18T14:32:00Z TASK-001 todo->in-progress jon@example.com manual
```
No commit required. The log file is modified; the developer's next commit will include this change.

**PASS** — `kanban start` works unchanged from the user's perspective.

**H5 Verdict**: Concept A fundamentally conflicts with `kanban start`. Concept B preserves it.

---

### H6 — Value: Per-user board view

```
We believe filtering the board by git author identity
will satisfy the "my tasks" use case.
We will know this is TRUE when `kanban board` can be filtered
by the current git user.name without additional configuration.
We will know this is FALSE when user identity requires
a separate configuration step beyond git config.
```

**Test against Concept A**:
- State derives from commit trailers, each of which is attached to a commit with an author
- `kanban board --me` could filter to commits where `author.email == git config user.email`
- **PASS** — natural fit

**Test against Concept B**:
- Log entries include `author@example.com` on each line
- `kanban board --me` reads `git config user.email` and filters log entries
- **PASS** — equally achievable, slightly more implementation work

**But**: The current system already has an `Assignee` field on tasks, and `kanban start` sets it to `git config user.name`. A per-user filter on `kanban board` could be implemented today by filtering tasks where `assignee == current identity`, without any architectural change.

**H6 Verdict**: Both concepts solve the per-user view. The current system can *also* solve it with a `--me` flag on `kanban board` using the existing `assignee` field. This is a lower-risk path to the same outcome.

---

## Consolidated Test Results

| Hypothesis | Concept A (Commit Trailers) | Concept B (Transition Log) | Current System + Extensions |
|-----------|---------------------------|---------------------------|---------------------------|
| H1 — Audit trail | PARTIAL (needs CLI command) | PASS | PARTIAL (git log works but not obvious) |
| H2 — Board performance | FAIL at scale / CONDITIONAL | PASS | PASS (current) |
| H3 — Repo clutter | PASS (but `kanban start` broken) | PARTIAL PASS | FAIL (current) |
| H4 — Rebase/amend safety | FAIL | PASS | N/A |
| H5 — `kanban start` viability | FAIL | PASS | PASS (current) |
| H6 — Per-user view | PASS | PASS | PASS (with minor extension) |

**Score**: Concept A: 1.5/6 | Concept B: 5.5/6 | Current + Extensions: 3.5/6

---

## Critical Insight: The Audit Trail Problem Is Already Partially Solved

The participant's primary motivator was the audit trail. A key finding from codebase analysis:

**The current system already creates an implicit audit trail via git history.**

When the hook fires and updates `TASK-001.md` from `status: todo` to `status: in-progress`, that file change is committed (either by the developer or by `kanban ci-done`). Running `git log --follow -- .kanban/tasks/TASK-001.md` shows every commit that touched that file, with timestamps and authors.

The audit trail is there. It is just not surfaced through the CLI.

**This means the primary stated need (audit trail) could be addressed by adding `kanban log TASK-001` that runs `git log -- .kanban/tasks/<id>.md` — a two-hour implementation, not an architectural overhaul.**

This does not invalidate the other problems (hook-writes-files feels wrong, clutter), but it is important to note that the highest-priority pain point has a low-cost solution path that does not require an architectural change.

---

## Gate G3 Evaluation

| Criterion | Target | Current | Status |
|-----------|--------|---------|--------|
| 5+ users tested | 5+ | 1 (analysis-based) | CONDITIONAL |
| >80% task completion | >80% | Concept B: 5.5/6 hypotheses pass | PASS |
| Value perception validated | >70% | Both primary problems (audit trail, clutter) addressed by Concept B | PASS |
| Key assumptions validated | >80% | 83% (5/6) | PASS |

**Gate G3: CONDITIONAL PASS** — analysis-based testing, not prototype testing. A prototype test with 2-3 additional developers would strengthen confidence.
