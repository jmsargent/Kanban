<!-- markdownlint-disable MD024 -->
# User Stories: new-editor-mode

---

## US-01: kanban new launches editor when invoked with no arguments

### Problem

Alex is a developer who uses kanban to track tasks inside their git repositories. They find it friction-heavy to compose a task title, description, and priority all on one command line — especially mid-flow during a coding session. Today, running `kanban new` with no arguments silently fails with a validation error rather than opening the editor the way `kanban edit` does.

### Who

- Developers using kanban as a personal task tracker
- Working interactively at a terminal (TTY context)
- Familiar with `$EDITOR`-based workflows (git commit, `kanban edit`)

### Solution

When `kanban new` is invoked with no positional arguments, open `$EDITOR` with a blank task template — the same editor pattern already used by `kanban edit`. After the user saves and quits, validate the title is non-empty and create the task. If the title is empty, exit with code 2 and a clear error message. Existing `kanban new <title>` behaviour is unchanged.

### Domain Examples

#### Example 1: Alex captures a high-priority bug mid-debug session

Alex is investigating a production crash. They want to track the fix without losing their current thought. They type `kanban new`, their editor opens with a blank template. They fill in:

```
title: "Fix nil pointer in auth handler"
priority: "P1"
description: "Crashes on logout when session token is nil"
```

They save and quit. The terminal prints:

```
Created TASK-042: Fix nil pointer in auth handler
Hint: reference TASK-042 in your next commit to start tracking
```

#### Example 2: Priya captures a backlog item with minimal detail

Priya remembers she needs to update the README examples. She types `kanban new`, fills in only the title field (`Update README examples`), leaves all other fields blank, saves and quits. The task is created as TASK-007 with no priority, no assignee, and no description.

#### Example 3: Jordan opens the editor, then quits without filling in the title

Jordan runs `kanban new` by accident. The editor opens. They realise they don't have a task to create yet. They quit without saving (`:q!` in vim) or save with an empty title. The terminal prints `title cannot be empty` to stderr and exits with code 2. No task file is created.

### UAT Scenarios (BDD)

#### Scenario: Editor opens when kanban new has no arguments
Given Alex is in a git repository with kanban initialised
And `$EDITOR` is set to a test double that records the temp file path
When Alex runs `kanban new` via CLI with no arguments
Then the editor process is invoked with a temp file path
And the temp file contains empty `title`, `priority`, `assignee`, `description` fields
And the temp file contains comment lines explaining that title is required

#### Scenario: Task created with title and optional fields
Given the editor has been opened with a blank template
When Alex sets `title` to "Fix nil pointer in auth handler" and `priority` to "P1"
And Alex saves and quits the editor
Then the process exits with code 0
And stdout contains "Created TASK-042: Fix nil pointer in auth handler"
And stdout contains "Hint: reference TASK-042 in your next commit to start tracking"
And a task file exists in `.kanban/tasks/` with title "Fix nil pointer in auth handler" and priority "P1"

#### Scenario: Task created with title only (optional fields blank)
Given Priya runs `kanban new` with no arguments
When Priya sets `title` to "Update README examples" and leaves all other fields blank
And Priya saves and quits the editor
Then the process exits with code 0
And the created task has no priority, no assignee, and no description

#### Scenario: Empty title after editor save is rejected
Given Jordan opens the editor via `kanban new`
When Jordan saves the file without filling in the title field
Then the process prints to stderr "title cannot be empty"
And the process exits with code 2
And no task file is created in `.kanban/tasks/`

#### Scenario: Existing kanban new <title> behaviour unchanged
When Alex runs `kanban new` with argument "Quick hotfix for prod"
Then no editor is opened
And the process exits with code 0
And stdout contains "Created TASK-043: Quick hotfix for prod"

#### Scenario: $EDITOR not set and vi unavailable
Given `$EDITOR` is not set
And `vi` is not present in `PATH`
When Alex runs `kanban new` with no arguments
Then the process prints to stderr a message containing "open editor"
And the process exits with code 1

#### Scenario: kanban not initialised before editor opens
Given Alex is in a git repository where `kanban init` has not been run
When Alex runs `kanban new` with no arguments
Then the process prints to stderr "kanban not initialised — run 'kanban init' first"
And the process exits with code 1
And no editor is opened

### Acceptance Criteria

- [ ] `kanban new` with no arguments opens `$EDITOR` instead of printing a validation error
- [ ] The temp file presented to the editor contains empty `title`, `priority`, `assignee`, `description` fields plus comment guidance
- [ ] After editor exits with non-empty title, a task is created and success output matches `kanban new <title>` format exactly
- [ ] Empty title after editor exits → stderr "title cannot be empty" + exit code 2
- [ ] No task file is created when the title is empty
- [ ] `kanban new <title>` with a positional argument continues to work without opening the editor
- [ ] Pre-flight checks (git repo, git identity, kanban init) run before the editor opens; failures produce existing error messages + appropriate exit codes
- [ ] Editor resolution uses the same `openEditor()` function as `kanban edit` (no duplication)
- [ ] Temp file is cleaned up after editor exits, regardless of success or error

### Outcome KPIs

- **Who**: developers running `kanban new` interactively
- **Does what**: create tasks with title and optional metadata in a single command invocation, without a second `kanban edit` step
- **By how much**: reduce two-step task creation (new + edit) to one step for at least 80% of task creation events (measured via task creation frequency vs edit-immediately-after-create frequency)
- **Measured by**: count of `kanban edit` invocations occurring within 60 seconds of `kanban new` in the same repository — proxy for "new then edit" pattern
- **Baseline**: not currently measured (establish baseline in first 30 days post-release)

### Technical Notes

- **Port constraint**: `openEditor()` is currently a private function in `internal/usecases/edit_task.go`. The solution-architect must decide how to share it with the new creation path — options include: extract to a package-level exported function, move to an `EditorPort` interface, or accept a light duplication if the function is trivial. Duplication is the least preferred option.
- **WriteTemp with zero-value task**: `EditFilePort.WriteTemp(domain.Task{})` must produce a blank template with empty strings, not a template with default field values. The filesystem adapter must be verified to handle this without panicking.
- **Due date not in editor template**: the `due` field is intentionally omitted from the blank template in this story. The `--due` flag remains available on the `kanban new <title>` path. Adding `due` to the editor template is a future enhancement.
- **Dependency**: `EditFilePort` interface (`ports/repositories.go:56`) and its filesystem adapter must be unchanged.
- **Exit code contract**: empty title → exit 2 (usage error, consistent with how `AddTask.Execute` returning `ErrInvalidInput` is handled in `new.go:66-69`). The validation happens in the CLI adapter before calling `AddTask.Execute`, not inside the use case, so the exit code is controlled correctly.
