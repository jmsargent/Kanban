# Requirements: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15
Status: Ready for DESIGN wave

---

## 1. Business Context

Small development teams use external project management tools (Jira, Trello, Linear) that exist
outside the codebase. These tools go stale because updating them requires a separate context
switch. Developers forget, prioritise coding over admin, and the board stops reflecting reality.

This feature delivers a kanban board whose state IS the git repository. Tasks are files.
State transitions are git events. There is no separate tool to maintain.

---

## 2. Personas

### Primary: Rafael Rodrigues

Backend developer, 3 years experience, works from the terminal 90% of the time. Uses git
daily. Frustrated by Jira because "it's always wrong -- nobody updates it." Comfortable
with markdown and YAML files. Works on a 4-person team sharing a single git repository.

Pain: After every sprint, the board is a graveyard of stale cards that nobody updated
because updating Jira requires leaving the code and logging into a browser.

Goal: Know what the team is working on by running one command in the terminal.

### Secondary: Priya Nair

Frontend developer on the same team. Less CLI-fluent than Rafael but comfortable with git.
Uses the kanban board primarily to see what her teammates are doing, not to manage her own tasks.

Pain: Asks Rafael "what are you working on?" because the board is never right.

---

## 3. Functional Requirements

### FR-1: Repository Initialisation

The CLI must initialise the .kanban/ directory structure in an existing git repository.
- Creates .kanban/tasks/ directory
- Creates .kanban/config with default column definitions and ci_task_pattern
- Idempotent: running kanban init on an already-initialised repo is a no-op with a status message

### FR-2: Task Creation

A developer can create a task with a title (required) and optional metadata.
- Required: title (non-empty string)
- Optional: priority (P1/P2/P3), due date (ISO date, today or future), assignee (free-text)
- Task ID is auto-generated in TASK-NNN format, incrementing from highest existing
- Task file written to .kanban/tasks/TASK-NNN.md
- Initial status is always "todo"
- Output confirms: ID, file path, status, all provided fields, and a commit message tip

### FR-3: Board Display

A developer can view all tasks grouped by status in the terminal.
- Reads all .md files in .kanban/tasks/
- Groups by status field value
- Within each group, shows: ID, title, priority, due date, assignee
- Shows count per status group in the heading
- Shows overdue indicator when today's date is after task.due
- Supports --json flag for machine-parseable output
- Empty .kanban/tasks/ shows empty state with onboarding tip

### FR-4: Automatic Status Transition -- In Progress

When a developer makes a git commit whose message contains a task ID, that task's status
is automatically updated from "todo" to "in-progress".
- Transition fires on first commit referencing the task
- Does not fire if task is already in-progress or done (no-op)
- If task ID is not found, a warning is shown but the commit is NOT blocked
- Commit output shows the transition line: "kanban: TASK-NNN moved  todo -> in-progress"
- Delivered as a git commit-msg hook installed via kanban install-hook

### FR-5: Automatic Status Transition -- Done

When a CI pipeline passes, tasks referenced in commits in that pipeline run are automatically
updated from "in-progress" to "done".
- Transition fires only when pipeline exit code is 0 (all tests pass)
- Does not fire if pipeline fails (non-zero exit code)
- Does not fire if task is already done (no-op)
- CI step commits updated task file(s) back to repo
- CI log shows individual transition lines per task

### FR-6: Task Editing

A developer can edit any task field using their default editor ($EDITOR).
- Command: kanban edit TASK-NNN
- Displays current field values before opening editor
- Opens task file in $EDITOR
- After save and exit, confirms which fields changed
- status field can be manually overridden via edit (escape hatch for corrections)

### FR-7: Task Deletion

A developer can delete a task by removing its file from the repository.
- Command: kanban delete TASK-NNN
- Shows confirmation prompt with task title and current status
- On confirm: removes task file
- On abort: no change
- --force flag skips confirmation (for scripting and CI use)
- After deletion, developer must commit the removal (kanban does not auto-commit on delete)
- Output includes a git commit command suggestion

---

## 4. Non-Functional Requirements

### NFR-1: Performance
- kanban board must produce first output within 100ms for repos with up to 500 task files
- commit hook must complete within 500ms (must not visibly slow down git commit)

### NFR-2: Reliability
- All file writes are atomic (write to temp file, rename) to prevent partial writes on interrupt
- kanban commands are idempotent where possible: running the same command twice produces the same result
- commit hook failure must never block git operations (hook exits 0 regardless of kanban errors; logs to .kanban/hook.log)

### NFR-3: CI/CD Compatibility
- CI step must work in non-TTY environments (no interactive prompts, no color codes)
- CI step must be deliverable as a standalone script requiring only git and the kanban binary
- Exit codes: 0=success, 1=error, 2=usage error (consistent across all commands)

### NFR-4: Output
- Human-readable (color, aligned columns) in TTY by default
- NO_COLOR environment variable respected (no ANSI codes)
- Color is never the only means of conveying information (all status info available via text label)
- --json flag available on kanban board and kanban list for machine consumption
- --json schema is a versioned contract (breaking changes require major version bump)

### NFR-5: Developer Experience
- Every subcommand supports --help with usage, flags, and at least one example
- Did-you-mean suggestions for typos in command names
- Shell completion scripts available (bash, zsh, fish)
- First output within 100ms (print something immediately before any file I/O)

### NFR-6: Git Compatibility
- Works with any git repository (no dependency on GitHub, GitLab, or any specific remote)
- .kanban/ directory should be committed to the repo (shared state for the team)
- .kanban/hook.log should be added to .gitignore by kanban init

---

## 5. Business Rules

- BR-1: A task's status may only be one of the configured column values (default: todo, in-progress, done)
- BR-2: Task IDs must be unique within a repository; kanban add must reject if collision detected
- BR-3: A task cannot be created with a due date in the past
- BR-4: Automatic transitions move status in one direction only: todo -> in-progress -> done. A commit cannot move a task backwards. Manual edit can override.
- BR-5: The CI transition to "done" fires only on pipeline pass. Pipeline failure is a no-op.
- BR-6: Commit hook failure does not block the git commit. Developer workflow is never interrupted by kanban errors.
- BR-7: kanban delete does not auto-commit. The developer owns the commit history.

---

## 6. Constraints and Dependencies

| Item | Type | Notes |
|---|---|---|
| Git must be installed | Hard constraint | kanban requires git for repo_root detection and hook installation |
| Task file format | Open decision | Leave to DESIGN wave; must support structured fields (title, status, priority, due, assignee) |
| Assignee field | Design decision | Free-text for MVP; validated user list deferred |
| CI/CD platform | Integration | GitHub Actions and GitLab CI are primary targets for R3; generic shell script first |
| $EDITOR | Environment | kanban edit uses $EDITOR; falls back to vi if unset; must document |
| .kanban/config format | Open decision | Leave to DESIGN wave; must support ci_task_pattern and column definitions |

---

## 7. Domain Glossary

| Term | Definition |
|---|---|
| Task | A unit of work tracked as a file in .kanban/tasks/ |
| Board | The output of "kanban board" -- tasks grouped by status |
| Status | The current workflow state of a task: todo, in-progress, or done |
| Column | A status group displayed in the board output |
| Task ID | The unique identifier of a task, derived from its filename (TASK-NNN) |
| ci_task_pattern | A regex configured in .kanban/config used to extract task IDs from commit messages |
| Commit hook | A git commit-msg hook that auto-transitions tasks on commit |
| CI step | A script run in the CI/CD pipeline that auto-transitions tasks on test pass |
| Auto-transition | A status change triggered by a git event (commit or CI pass), not a manual command |
| repo_root | The root directory of the git repository (git rev-parse --show-toplevel) |
