# JTBD Job Stories — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)
**Source**: Derived from DISCOVER wave interviews, problem-validation.md, and wave-decisions.md

---

## Persona

**Jon Santos** — sole developer on a Go open-source CLI tool. Works from a single laptop, commits frequently using topic branches that he rebases before merging. He uses `kanban-tasks` to track his own work; his motivation is a clean, auditable git history that tells the whole story of a feature without requiring a separate system.

Jon's git identity: `jon@kanbandev.io`
His task IDs follow the pattern: `TASK-001`, `TASK-002`, etc.

---

## Job Story 1 — Audit Trail (Primary)

**When** I finish a session of work and look back at my git log, I want to see not just what code changed but when each task transitioned and who triggered it, so I can reconstruct the exact sequence of events on a feature without opening a separate system.

### Functional Job
Surface the full transition history for any task (who moved it, when, from what status, to what status) from a single CLI command.

### Emotional Job
Feel confident that my work history is complete and trustworthy — that nothing is lost even if I switch branches or rebase.

### Social Job
Demonstrate to collaborators (or future self) that the project was run with discipline: tasks moved deliberately, not haphazardly.

### Forces Analysis

**Demand-Generating:**
- **Push**: Running `git log -- .kanban/tasks/TASK-001.md` shows file-mutation commits but does not clearly articulate state changes. The developer must mentally reconstruct what "edited this file" means in terms of status transitions.
- **Pull**: A single `kanban log TASK-001` command that surfaces `todo → in-progress at 2026-03-15 14:23 by jon@kanbandev.io` delivers the audit trail in the developer's domain language.

**Demand-Reducing:**
- **Anxiety**: Will the command show me the full history or just recent transitions? What happens if I already have a long git log?
- **Habit**: The current system does not have a `kanban log` command. Developers who have internalized `git log` may just keep using it.

**Switch likelihood**: HIGH — this is the stated primary motivator; the habit (git log) is non-optimal.

---

## Job Story 2 — Hook Feels Wrong (Architectural Integrity)

**When** I watch the commit-msg hook write `status: in-progress` into a `.kanban/tasks/` file as a side effect of my commit, I want the hook to record the transition in a dedicated place without mutating task definition files, so I can trust that task files represent only what I intentionally put there.

### Functional Job
Decouple task state storage from task definition storage. The hook should not write to the same file that holds the task's title, priority, and description.

### Emotional Job
Feel that the system has a clean architecture — that each file has one purpose and the hook is not a hidden file-mutation agent operating behind the scenes.

### Social Job
Be seen as someone whose tools are well-designed — a developer who chose a CLI that respects separation of concerns.

### Forces Analysis

**Demand-Generating:**
- **Push**: Direct quote from DISCOVER: "the hook writing back to files feels like the wrong place." This is an architectural intuition signal, not a functional failure.
- **Pull**: Hook appends one line to `.kanban/transitions.log` — a single-purpose file with a clear contract. Task files only change when the developer explicitly edits them.

**Demand-Reducing:**
- **Anxiety**: Will the new log file get out of sync with task files? What if the log is corrupted?
- **Habit**: The current system has worked without incident. The pain is architectural, not operational.

**Switch likelihood**: MEDIUM — the push is real but the habit (current system working) reduces urgency.

---

## Job Story 3 — Repository Cleanliness

**When** I run `git diff` after making a commit that references a task, I want to see only my code changes in the diff (not task file status rewrites), so I can review my work without noise from task-state housekeeping.

### Functional Job
Eliminate task file status-field writes from the developer's commit diffs. State transitions should not appear as file mutations in `git status` or `git diff`.

### Emotional Job
Feel that my git repository is clean and purposeful — every file change represents intentional work, not mechanical side effects.

### Social Job
Produce pull requests and commit histories that colleagues can review without filtering out kanban housekeeping noise.

### Forces Analysis

**Demand-Generating:**
- **Push**: Status field rewrites show up in `git show` and `git diff` mixed with code changes. The commit message references a task; the diff shows both code and a `status: in-progress` change in a YAML file.
- **Pull**: Commit diffs that show only code changes. The transition log records the state change separately, committed as its own file change (or by `kanban ci-done`).

**Demand-Reducing:**
- **Anxiety**: Concept B still writes to `.kanban/transitions.log` — so there is still a non-code file in the diff. The noise is reduced but not eliminated.
- **Habit**: Developer has normalized seeing the status field change in diffs.

**Switch likelihood**: MEDIUM-HIGH — the clutter is real, the partial improvement of Concept B satisfies the "clutter less" language from DISCOVER.

---

## Job Story 4 — Per-User Board View

**When** I open `kanban board` in a project where multiple contributors are working, I want to see a filtered view of only the tasks assigned to me (via my git identity), so I can focus on my work without manually scanning through tasks owned by others.

### Functional Job
Filter the board display by the current `git config user.email`, showing only tasks where the assignee matches the local git identity.

### Emotional Job
Feel immediately oriented when opening the board — "these are my tasks" without scanning a global list.

### Social Job
Demonstrate that the tool supports multi-developer workflows naturally through git identity, not manual user management.

### Forces Analysis

**Demand-Generating:**
- **Push**: On a project with 20+ tasks across contributors, the board shows everything. The developer must scan and filter mentally.
- **Pull**: `kanban board --me` instantly filters to "my tasks" using the git identity already configured.

**Demand-Reducing:**
- **Anxiety**: What if `assignee` was not set correctly on some tasks? Will tasks be silently hidden?
- **Habit**: Reading the full board and mentally filtering has become automatic.

**Switch likelihood**: MEDIUM — the pain scales with team size; for a solo developer it is low urgency.

---

## Job Story Summary

| # | Job Story | Primary Force | Importance (DISCOVER score) | Switch Likelihood |
|---|-----------|--------------|---------------------------|-------------------|
| JS-1 | Audit trail — `kanban log TASK-ID` | Push: history not surfaced | O1: 16 | HIGH |
| JS-2 | Hook should not write task files | Push: architectural intuition | O2: 13 | MEDIUM |
| JS-3 | Clean diffs without state writes | Push: repo noise | O3+O5: 14+8 | MEDIUM-HIGH |
| JS-4 | Per-user board filter | Push: global board is noisy at scale | O4: 10 | MEDIUM |
