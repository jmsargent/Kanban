# Acceptance Criteria: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15

All criteria are derived from UAT scenarios in user-stories.md.
Format: observable, testable, implementation-neutral.

---

## US-01: Repository Initialisation

- AC-01-1: "kanban init" in a git repository creates .kanban/tasks/ directory
- AC-01-2: "kanban init" creates .kanban/config with default ci_task_pattern (TASK-[0-9]+) and column list (todo, in-progress, done)
- AC-01-3: "kanban init" adds .kanban/hook.log to .gitignore
- AC-01-4: Running "kanban init" in an already-initialised repo produces "Already initialised" message and makes no changes to existing files
- AC-01-5: Running "kanban init" outside a git repository exits with code 1 and message containing "Not a git repository"

---

## US-02: Create Task

- AC-02-1: "kanban add" with a non-empty title creates .kanban/tasks/TASK-NNN.md with status field set to "todo"
- AC-02-2: Task IDs auto-increment: if TASK-001.md exists, next add creates TASK-002.md
- AC-02-3: Creation output shows: ID, title, file path, status, and a commit tip with the correct task ID
- AC-02-4: --priority flag value is stored in the task file and shown by "kanban board"
- AC-02-5: --due flag value is stored in the task file and shown by "kanban board"
- AC-02-6: --assignee flag value is stored in the task file and shown by "kanban board"
- AC-02-7: "kanban add" with empty title exits with code 2 and message "Task title is required" with no file created
- AC-02-8: "kanban add" with a past due date exits with code 2 and message "Due date must be today or in the future" with no file created
- AC-02-9: "kanban add" outside a git repository exits with code 1 and message "Not a git repository"

---

## US-03: View Board

- AC-03-1: "kanban board" output groups tasks under headings: TODO, IN PROGRESS, DONE
- AC-03-2: Each task line shows: ID, title, priority, due date, assignee
- AC-03-3: Fields with no value show "--" for priority/due or "unassigned" for assignee
- AC-03-4: Each group heading shows the count of tasks in that group
- AC-03-5: A task whose due date is before today shows a distinct overdue indicator on its line
- AC-03-6: Empty .kanban/tasks/ shows "No tasks found" message and "kanban add" suggestion
- AC-03-7: "kanban board --json" outputs a valid JSON array where each element contains fields: id, title, status, priority, due, assignee
- AC-03-8: When NO_COLOR environment variable is set, output contains no ANSI escape codes
- AC-03-9: When output is piped (non-TTY), output contains no ANSI escape codes and no spinner characters
- AC-03-10: First visible output appears within 100ms for a repository with up to 500 task files

---

## US-04: Auto-move to In Progress (Commit Hook)

- AC-04-1: A git commit message matching ci_task_pattern causes the matched task file status to be updated to "in-progress"
- AC-04-2: Commit output includes "kanban: TASK-NNN moved  todo -> in-progress" when a transition occurs
- AC-04-3: The git commit itself always succeeds (exit code 0) regardless of kanban hook outcome
- AC-04-4: A commit referencing a task already "in-progress" or "done" produces no transition and no kanban output
- AC-04-5: A commit referencing an unknown task ID shows a warning and does not block the commit
- AC-04-6: A commit message with no pattern match produces no kanban output
- AC-04-7: The hook reads ci_task_pattern from .kanban/config (not hardcoded)
- AC-04-8: The hook completes within 500ms

---

## US-05: Auto-move to Done (CI Pipeline)

- AC-05-1: CI step executes only when the test suite exits with code 0
- AC-05-2: CI step scans all commit messages in the pipeline run and extracts IDs matching ci_task_pattern
- AC-05-3: Each matched task with status "in-progress" has its file updated to "done"
- AC-05-4: CI step commits updated task files back to the repository
- AC-05-5: CI log contains an individual transition line for each moved task
- AC-05-6: When tests fail (non-zero exit), no task files are modified
- AC-05-7: A task already in "done" status: no-op, no output, no file modification
- AC-05-8: CI step reads ci_task_pattern from .kanban/config in the checked-out repo
- AC-05-9: CI step runs without interactive prompts or color codes (non-TTY compatible)

---

## US-06: Edit Task

- AC-06-1: "kanban edit TASK-NNN" displays all current field values before opening the editor
- AC-06-2: The editor opened is $EDITOR; falls back to vi when $EDITOR is unset
- AC-06-3: After save and exit, output lists the names of changed fields
- AC-06-4: If the user exits with no changes, output shows "No changes made."
- AC-06-5: "kanban board" reflects updated field values immediately after edit
- AC-06-6: "kanban edit" for a non-existent task exits with code 1 with an actionable error message

---

## US-07: Delete Task

- AC-07-1: "kanban delete TASK-NNN" shows a confirmation prompt with task title and current status
- AC-07-2: Entering "y" removes .kanban/tasks/TASK-NNN.md from the filesystem
- AC-07-3: Entering "n" or pressing Enter aborts with "Aborted. TASK-NNN unchanged." and no file change
- AC-07-4: "kanban delete TASK-NNN --force" removes the file immediately without prompting
- AC-07-5: Deletion output includes task title, "Deleted" confirmation, and a suggested git commit command
- AC-07-6: "kanban board" run after deletion no longer lists the deleted task
- AC-07-7: "kanban delete" for a non-existent task exits with code 1 with an actionable error message
- AC-07-8: kanban does not auto-commit the deletion; developer commits manually

---

## Cross-Cutting Criteria

- AC-X-1: Every kanban subcommand supports --help with usage, flags, and at least one example
- AC-X-2: Exit codes consistent across all commands: 0=success, 1=runtime error, 2=usage error
- AC-X-3: All commands run outside a git repository exit with code 1 and message "Not a git repository"
- AC-X-4: Error messages state what happened, why it happened, and what to do next
- AC-X-5: kanban --version outputs the installed version string
