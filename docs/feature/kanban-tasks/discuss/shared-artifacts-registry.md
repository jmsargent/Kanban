# Shared Artifacts Registry

Feature: kanban-tasks
Journey: Task Management (Git-Native CLI)
Date: 2026-03-15

---

## Purpose

Every value that appears in multiple journey steps is documented here with a single source of
truth and all consumers. In a git-native CLI tool the artifacts are file fields, not UI state --
but the same rule applies: one source, many readers, and any drift is a bug.

---

## Artifacts

### task.id

- Source of truth: The filename of the task file. TASK-001.md -> ID is TASK-001.
- Set by: kanban add (auto-increments from highest existing TASK-NNN)
- Consumers:
  - kanban board (displayed as first column)
  - kanban edit (argument, resolves to file path)
  - kanban delete (argument, resolves to file path)
  - git commit messages (developer references it manually or via hook tip)
  - commit hook and CI step (extracted via ci_task_pattern regex)
- Risk: HIGH -- if two task files ever share the same ID prefix, all ID-based commands break
- Validation: kanban add must check for existing TASK-NNN.md before writing; ID generation must be atomic (no race conditions on concurrent adds)

---

### task.title

- Source of truth: title field in .kanban/tasks/TASK-NNN.md
- Set by: kanban add (first positional argument); updated by kanban edit
- Consumers:
  - kanban board (second column in task line)
  - kanban delete (shown in confirmation prompt and deletion log -- read immediately before display, never cached)
  - CI log messages (shown in transition output)
- Risk: HIGH -- stale title in delete confirmation causes user to question which task they are deleting
- Validation: kanban delete must read the file immediately before printing the confirmation, not from an in-memory cache

---

### task.status

- Source of truth: status field in .kanban/tasks/TASK-NNN.md
- Valid values: todo, in-progress, done (must match values in .kanban/config column definitions)
- Set by: kanban add (initialised to "todo"); commit hook (todo -> in-progress); CI step (in-progress -> done); kanban edit (manual override)
- Consumers:
  - kanban board (groups tasks into column headings)
  - kanban delete (shown in confirmation prompt: "Delete TASK-001 ... [in-progress]?")
  - commit hook (reads current value to decide whether transition applies)
  - CI step (reads current value to decide whether transition applies)
- Risk: CRITICAL -- if task.status value is not in the configured column list, kanban board cannot group it; task becomes invisible
- Validation: any write to task.status must validate against configured column values; kanban board must have a fallback "UNKNOWN STATUS" group for orphaned values

---

### task.priority

- Source of truth: priority field in .kanban/tasks/TASK-NNN.md
- Valid values: P1, P2, P3 (or configurable in .kanban/config)
- Set by: kanban add (--priority flag, optional); kanban edit
- Consumers:
  - kanban board (fourth column in task line)
  - future: sort and filter operations
- Risk: LOW -- priority display is informational; mismatch between file and display is confusing but not blocking
- Validation: kanban board reads priority directly from file on each render; no caching

---

### task.due

- Source of truth: due field in .kanban/tasks/TASK-NNN.md (ISO date: YYYY-MM-DD)
- Set by: kanban add (--due flag, optional); kanban edit
- Consumers:
  - kanban board (fifth column; shows overdue indicator if today > due)
  - kanban add validation (rejects past dates at creation time)
- Risk: MEDIUM -- overdue calculation must use the stored date in consistent timezone; server CI and developer local timezone may differ
- Validation: overdue calculation uses UTC date comparison; due field stored as YYYY-MM-DD (no time component)

---

### task.assignee

- Source of truth: assignee field in .kanban/tasks/TASK-NNN.md (free-text for MVP)
- Set by: kanban add (--assignee flag, optional); kanban edit
- Consumers:
  - kanban board (sixth column in task line; "unassigned" if empty)
  - future: filter by assignee
- Risk: LOW -- free-text means no validation; typos create duplicate assignee names in filters
- Validation: none for MVP; note that upgrade to validated user list is a future release decision

---

### task.description

- Source of truth: description field in .kanban/tasks/TASK-NNN.md
- Set by: kanban add (not available as flag -- edit after creation); kanban edit
- Consumers:
  - kanban edit (shown in editor)
  - future: kanban show TASK-001 (detailed view command)
  - NOT shown in kanban board output (too verbose for summary view)
- Risk: LOW -- single consumer for MVP; no cross-step consistency risk
- Validation: none for MVP; future kanban show command will read and display

---

### repo_root

- Source of truth: output of "git rev-parse --show-toplevel" run at command invocation time
- Set by: resolved by each kanban command at startup
- Consumers:
  - All kanban commands (resolves .kanban/ directory as repo_root/.kanban/)
- Risk: HIGH -- if repo_root is wrong, all file paths are wrong and commands silently operate on wrong directory
- Validation: every kanban command must call "git rev-parse --show-toplevel" first; fail with "Error: Not a git repository." if it exits non-zero

---

### ci_task_pattern

- Source of truth: .kanban/config (key: task_pattern; default: TASK-[0-9]+)
- Set by: kanban init (writes default); user edits .kanban/config to customise
- Consumers:
  - commit hook (extracts task IDs from commit message)
  - CI step (extracts task IDs from commit messages in pipeline run)
  - kanban add tip message (shows pattern example to developer)
- Risk: HIGH -- if commit hook and CI step use different patterns, transitions are inconsistent; hook moves to in-progress but CI does not move to done
- Validation: both commit hook and CI step must read pattern from same .kanban/config; do not hardcode pattern in either

---

## Integration Checkpoints

| Checkpoint | Steps Involved | Risk | Validation |
|---|---|---|---|
| task.status matches kanban board column grouping | 1, 2, 3, 4, 5, 6 | CRITICAL | Board groups by file value; any write error leaves task in wrong column |
| task.id uniqueness | 2 | HIGH | kanban add checks for existing ID before writing |
| ci_task_pattern identical between hook and CI | 3, 4 | HIGH | Both read from .kanban/config; no hardcoded values |
| task.title fresh on delete confirmation | 2, 5, 6 | HIGH | File read immediately before confirmation display |
| repo_root resolved at startup | all | HIGH | All commands fail fast with clear error if not in git repo |

---

## Open Decisions (Dependency Flags)

These decisions must be made before or during DESIGN wave. They are not blocking for requirements
but will constrain the solution.

| Item | Question | Recommended Default | Impact |
|---|---|---|---|
| Task file format | Markdown front matter, plain YAML, TOML, or custom? | Leave to DESIGN wave | Affects how fields are parsed; commit hook and CI step depend on this |
| Assignee validation | Free-text or validated user list? | Free-text for MVP | Free-text unblocks shipping; user list requires user management feature |
| Commit hook delivery | Client-side (git hooks) or server-side (CI only)? | Both -- hook is optional, CI is required | Client hook gives instant feedback; CI is the source of truth |
| Task ID format | Sequential TASK-NNN or content-addressed (hash)? | Sequential TASK-NNN | Easier for humans to reference in commits; hash avoids ID conflicts on parallel adds |
| Column configuration | Fixed (todo/in-progress/done) or configurable per repo? | Configurable via .kanban/config | Configurable is more powerful but adds complexity to status validation |
