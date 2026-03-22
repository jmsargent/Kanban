# Acceptance Criteria — new-editor-mode

## US-01: kanban new launches editor when invoked with no arguments

These criteria are derived directly from the UAT scenarios in user-stories.md. Each criterion is traceable to a specific scenario and names the driving port.

---

### AC-01: No-argument invocation routes to editor mode

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Editor opens when kanban new has no arguments"

When the developer invokes `kanban new` via CLI with no positional arguments, the `$EDITOR` process is launched. No validation error is printed to stderr before the editor opens.

**Pass condition**: a test double editor is invoked; no output appears on stderr before the editor exits.

---

### AC-02: Blank task template presented in editor

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Editor opens when kanban new has no arguments"

The temp file opened in the editor contains:
- A `title` field with an empty string value
- A `priority` field with an empty string value
- An `assignee` field with an empty string value
- A `description` field with an empty string value
- At least one comment line stating that `title` is required

**Pass condition**: contents of the temp file match the above structure before the editor modifies it.

---

### AC-03: Task created and confirmed after valid editor session

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Task created with title and optional fields"

After the editor exits with a non-empty title field, the process:
1. Creates a task file in `.kanban/tasks/`
2. Prints to stdout: `Created ${TASK_ID}: ${TASK_TITLE}`
3. Prints to stdout: `Hint: reference ${TASK_ID} in your next commit to start tracking`
4. Exits with code 0

The stdout format is byte-for-byte identical to the format produced by `kanban new <title>`.

---

### AC-04: Optional fields are persisted when filled

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Task created with title and optional fields"

When the developer sets `priority`, `assignee`, or `description` in the editor, those values are persisted in the task file. When any optional field is left blank, the task file omits or leaves blank that field — no default values are injected.

---

### AC-05: Empty title after editor exits → exit code 2

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Empty title after editor save is rejected"

When the editor exits and the `title` field is empty (or whitespace only):
1. The process prints to stderr: `title cannot be empty`
2. The process exits with code 2
3. No task file is created in `.kanban/tasks/`

**Exit code 2** signals a usage error (not a runtime failure), consistent with the existing `ErrInvalidInput` handling in `new.go`.

---

### AC-06: Existing kanban new <title> behaviour unchanged

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "Existing kanban new <title> behaviour unchanged"

When `kanban new` is invoked with a positional argument, no editor is launched. The existing behaviour (validate title, create task, print success) is preserved with no change to output or exit codes.

---

### AC-07: Runtime error when $EDITOR unavailable → exit code 1

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "$EDITOR not set and vi unavailable"

When `$EDITOR` is not set and `vi` is not in `PATH`:
1. The process prints to stderr a message containing "open editor"
2. The process exits with code 1

**Exit code 1** signals a runtime error, consistent with `new.go` error handling for system-level failures.

---

### AC-08: Pre-flight checks run before editor opens

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Scenario "kanban not initialised before editor opens"

If any pre-flight check fails (not a git repo, git identity not configured, kanban not initialised), the process exits with the existing error message and exit code without opening the editor.

Specific messages and exit codes:
- Not a git repo → stderr `Not a git repository` + exit 1
- Git identity missing → stderr `git identity not configured — run: git config --global user.name "Your Name"` + exit 1
- Not initialised → stderr `kanban not initialised — run 'kanban init' first` + exit 1

---

### AC-09: Temp file is cleaned up after editor exits

**Driving port**: `kanban new` subcommand (CLI primary adapter)
**Derived from**: Non-functional requirement — temp file cleanup

The temp file created by `EditFilePort.WriteTemp` is removed after the editor exits, regardless of whether the task was created or an error occurred. This applies to all exit paths (success, empty title, editor not found).

---

## Acceptance Criteria Traceability

| AC | UAT Scenario | Exit Code | Error Path? |
|---|---|---|---|
| AC-01 | Editor opens when kanban new has no arguments | — | No |
| AC-02 | Editor opens when kanban new has no arguments | — | No |
| AC-03 | Task created with title and optional fields | 0 | No |
| AC-04 | Task created with title only (optional fields blank) | 0 | No |
| AC-05 | Empty title after editor save is rejected | 2 | Yes |
| AC-06 | Existing kanban new <title> behaviour unchanged | 0 | No |
| AC-07 | $EDITOR not set and vi unavailable | 1 | Yes |
| AC-08 | kanban not initialised before editor opens | 1 | Yes |
| AC-09 | (NFR — all scenarios) | — | Both |
