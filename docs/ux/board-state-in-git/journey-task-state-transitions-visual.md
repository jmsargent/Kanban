# Journey: Task State Transitions — Visual Map
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Journey Overview

**Persona**: Jon Santos — sole developer, Go CLI tool, git-native workflow
**Goal**: Move a task from idea to done while keeping a clean, auditable git history
**Platform**: CLI / Terminal (cobra commands)

---

## Current-State Journey (Before Feature)

The current journey has three pain points marked with [!]:

```
Jon has an idea for a new feature
          |
          v
    kanban add "Add retry logic to CI step"
          |    Creates TASK-007 in .kanban/tasks/TASK-007.md
          |    status: todo written to YAML front matter
          |
          v
    [emotional state: neutral — task created, board shows it]
          |
          v
    Jon opens code editor, starts working
          |
          v
    kanban start TASK-007
          |    Hook writes status: in-progress to TASK-007.md [!]
          |    "This feels like the wrong place to write state"
          |
          v
    [emotional state: mild friction — side-effect mutation noticed]
          |
          v
    Jon makes commits referencing TASK-007
          |    commit-msg hook fires, re-reads status field
          |    git diff shows: status: todo -> in-progress change [!]
          |    "Why is a YAML field in my code diff?"
          |
          v
    [emotional state: frustration — diff noise]
          |
          v
    Jon finishes. kanban ci-done runs.
          |    CI writes status: done back to task file
          |    Makes a commit of the task file
          |
          v
    Jon reviews history weeks later:
    git log -- .kanban/tasks/TASK-007.md [!]
          |    Output: raw commit messages, not state transitions
          |    "When did this actually go in-progress? Was it yesterday or Tuesday?"
          |
          v
    [emotional state: uncertain — audit trail exists but is not readable]
```

**Pain summary**: Hook writes to wrong place [!] | Diff noise [!] | Audit trail not surfaced [!]

---

## Desired-State Journey — Phase 1: `kanban log` Quick Win

The Phase 1 journey eliminates pain point [!3] immediately. Pain points [!1] and [!2] are unchanged but the developer now has a usable audit trail command.

```
Emotional Arc: Neutral → Mild Friction → Relief
```

```
Jon has an idea
          |
          v
    kanban add "Add retry logic to CI step"
          |    Creates TASK-007 (no change from today)
          |
          v
    kanban start TASK-007  [still writes to task file — Phase 1 unchanged]
          |
          v
    Jon makes commits referencing TASK-007
          |    git diff still shows YAML status change [still present in Phase 1]
          |
          v
    kanban ci-done  [still commits task file]
          |
          v
    Jon wants to know when TASK-007 went in-progress:

    +-- kanban log TASK-007 ----------------------------------------+
    | TASK-007: Add retry logic to CI step                          |
    |                                                               |
    | 2026-03-15 09:14  todo → in-progress  jon@kanbandev.io       |
    |   a3f12bc  kanban start: TASK-007                            |
    |                                                               |
    | 2026-03-17 16:42  in-progress → done  jon@kanbandev.io       |
    |   e91c44d  ci: add retry logic TASK-007                      |
    +---------------------------------------------------------------+

    [emotional state: RELIEF — clear history in domain language]
```

**Phase 1 value**: Audit trail is now surfaced. One command. Zero architectural change.

---

## Desired-State Journey — Phase 2: Concept B (Full Transition Log)

The Phase 2 journey eliminates all three pain points. The emotional arc shifts to full confidence.

```
Emotional Arc: Curious → Focused → Confident/Satisfied
```

### Step 1: Task Creation

```
Jon has an idea
          |
          v
  +-- kanban add "Add retry logic to CI step" ----------------------+
  |   kanban add -t "Add retry logic to CI step" -p high           |
  |                                                                  |
  |   Task created: TASK-007                                        |
  |   File: .kanban/tasks/TASK-007.md                               |
  |   (No status: field in YAML — definition only)                  |
  +------------------------------------------------------------------+
          |
          v
  [emotional state: CURIOUS — "no status field? where is it?"]
          |
  (discovers: .kanban/tasks/TASK-007.md has comment:
   "# state is tracked in .kanban/transitions.log — see kanban board")
          |
          v
  [emotional state: NEUTRAL-POSITIVE — system is clean, task visible on board as TODO]
```

**Shared artifact**: `TASK-007` (task ID) sourced from `tasks.NextID()` — consumed by transitions.log, kanban log, kanban board.

### Step 2: Starting Work

```
          |
          v
  +-- kanban start TASK-007 ----------------------------------------+
  |                                                                   |
  |   Appends to .kanban/transitions.log:                            |
  |   2026-03-15T09:14:23Z TASK-007 todo->in-progress               |
  |       jon@kanbandev.io manual                                     |
  |                                                                   |
  |   Board now shows TASK-007 in IN PROGRESS column.                |
  |   No task file was modified.                                      |
  +-------------------------------------------------------------------+
          |
          v
  [emotional state: FOCUSED — "the log is separate, the task file is untouched"]
```

**Integration checkpoint**: `kanban board` must read transitions.log to derive status. If log is missing or unreadable, board falls back to "todo" for all tasks.

### Step 3: Committing Work (Hook Fires)

```
          |
          v
  git commit -m "refactor: add retry logic TASK-007"
          |
          v
  +-- commit-msg hook -----------------------------------------------+
  |   Parses message: finds "TASK-007"                               |
  |   Current status from log: in-progress                           |
  |   (no file write — hook is read-only on task files)             |
  |   Appends to .kanban/transitions.log:                            |
  |   2026-03-15T14:32:00Z TASK-007 in-progress->in-progress        |
  |       jon@kanbandev.io commit:a3f12bc                            |
  |   (transition recorded; state unchanged — still in-progress)     |
  +-------------------------------------------------------------------+
          |
          v
  git diff output:
  +-- git diff HEAD~1 -----------------------------------------------+
  |   diff --git a/internal/usecases/ci.go b/internal/usecases/ci.go|
  |   ...                                                            |
  |   diff --git a/.kanban/transitions.log                           |
  |   + 2026-03-15T14:32:00Z TASK-007 in-progress->...              |
  +------------------------------------------------------------------+
          |
          v
  [emotional state: FOCUSED — diff shows code + one log append, not YAML status rewrite]
```

**Key UX improvement**: The transitions.log diff is a single-line append (additive). The previous system showed YAML mutations (modify). Append is visually distinct and semantically clearer in diffs.

### Step 4: CI Completion

```
          |
          v
  kanban ci-done (runs in CI pipeline)
          |
          v
  +-- CI pipeline: kanban ci-done --from=abc123 --to=HEAD -----------+
  |   Reads commit range: finds TASK-007 references                  |
  |   Current status: in-progress                                    |
  |   Appends to .kanban/transitions.log:                            |
  |   2026-03-17T16:42:11Z TASK-007 in-progress->done               |
  |       ci@kanbandev.io ci-done:e91c44d                            |
  |   Commits .kanban/transitions.log only (one file, not all tasks) |
  +-------------------------------------------------------------------+
          |
          v
  [emotional state: CONFIDENT — CI committed one file, the audit trail is permanent]
```

### Step 5: Reviewing the Audit Trail

```
          |
          v
  +-- kanban log TASK-007 ----------------------------------------+
  | TASK-007: Add retry logic to CI step                          |
  |                                                               |
  | 2026-03-15 09:14  todo → in-progress  jon@kanbandev.io       |
  |   trigger: manual (kanban start)                              |
  |                                                               |
  | 2026-03-15 14:32  in-progress → in-progress  jon@kanbandev.io|
  |   trigger: commit a3f12bc                                     |
  |   msg: "refactor: add retry logic TASK-007"                  |
  |                                                               |
  | 2026-03-17 16:42  in-progress → done  ci@kanbandev.io        |
  |   trigger: ci-done e91c44d                                    |
  +---------------------------------------------------------------+
          |
          v
  [emotional state: SATISFIED — complete, readable audit trail in domain language]
```

### Step 6: Board View (Current State)

```
          |
          v
  +-- kanban board -------------------------------------------------+
  | TODO                IN PROGRESS          DONE                   |
  | ----                -----------          ----                   |
  | TASK-003            TASK-005             TASK-007               |
  | Add dark mode       Fix CI timeout       Add retry logic        |
  |                                          (done 2026-03-17)      |
  |                     TASK-006                                    |
  |                     Update docs                                 |
  +------------------------------------------------------------------+
          |
          v
  [emotional state: CONFIDENT — board reflects log-derived state, matches expectation]
```

---

## Emotional Arc Summary

### Phase 1 Journey (Quick Win)
| Step | Emotional State | Transition Type |
|------|----------------|-----------------|
| Task creation | Neutral | Baseline |
| kanban start | Mild friction (hook writes file) | Small negative |
| Commits | Mild frustration (YAML in diff) | Sustained negative |
| kanban log | Relief | Positive resolution |

**Arc**: Neutral → Mild Friction → Relief (P1 addressed, P2/P3 remain)

### Phase 2 Journey (Concept B)
| Step | Emotional State | Transition Type |
|------|----------------|-----------------|
| Task creation | Curious (no status field) | Mild uncertainty |
| Discovery of transition log | Neutral-positive | Understanding |
| kanban start | Focused (log append, no file write) | Confirmation |
| Commit (hook) | Focused (clean diff) | Satisfaction |
| CI done | Confident (one file committed) | Trust built |
| kanban log | Satisfied (complete history) | Full resolution |

**Arc**: Curious → Focused → Confident → Satisfied (all three pain points resolved)

---

## Error Path: Task Not Found

```
kanban log TASK-999
          |
          v
  +-- Error output -------------------------------------------------+
  | Error: task TASK-999 not found                                  |
  |                                                                  |
  |   No task file exists at .kanban/tasks/TASK-999.md             |
  |                                                                  |
  |   Try:                                                          |
  |     kanban board          -- see all tasks                      |
  |     kanban log TASK-007   -- view a specific task               |
  +------------------------------------------------------------------+
  exit code: 1
```

## Error Path: No Transitions Recorded

```
kanban log TASK-003
          |
          v
  +-- Output (Phase 1: git log fallback) --------------------------+
  | TASK-003: Add dark mode                                        |
  |                                                                 |
  | No commits reference this task file yet.                       |
  | Task created 2026-03-16, current status: todo                  |
  +----------------------------------------------------------------+

  +-- Output (Phase 2: transition log approach) --------------------+
  | TASK-003: Add dark mode                                        |
  |                                                                 |
  | No transitions recorded yet.                                   |
  | Current status: todo (implicit — no transitions in log)        |
  +----------------------------------------------------------------+
  exit code: 0
```

## Error Path: Uninitialized Repository

```
kanban log TASK-001  (run outside a kanban repo)
          |
          v
  +-- Error output -------------------------------------------------+
  | Error: not a kanban repository                                  |
  |                                                                  |
  |   No .kanban/ directory found in current path or parents.      |
  |                                                                  |
  |   Try:                                                          |
  |     kanban init            -- initialize kanban in this repo   |
  +------------------------------------------------------------------+
  exit code: 1
```
