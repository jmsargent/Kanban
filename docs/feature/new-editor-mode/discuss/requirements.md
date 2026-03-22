# Requirements: new-editor-mode

## Business Context

The `kanban new` command currently requires a task title as a positional argument. Developers who want to add a description, set priority, or assign a task must either use flags inline or run `kanban edit <id>` as a second step. This two-step flow interrupts coding context.

The `kanban edit` command already demonstrates the editor pattern: open `$EDITOR` with a temp YAML file, let the user fill in fields, read back the result. This feature extends that pattern to task creation.

---

## Scope

**In scope:**
- `kanban new` with no arguments opens `$EDITOR` with a blank task template
- User fills in title (required) and optionally priority, assignee, description
- Empty title after editor save → stderr message + exit 2
- Success output identical to existing `kanban new <title>` format
- Existing `kanban new <title>` (with argument) behaviour unchanged

**Out of scope:**
- `--due` flag support in editor mode (due dates are already supported via flag on the existing path; editor mode does not add a `due` field to the template in this story — see wave-decisions.md for rationale)
- New subcommands or aliases
- Changes to `kanban edit` behaviour
- Changes to `kanban add` behaviour

---

## Business Rules

| Rule | Source | Notes |
|------|--------|-------|
| BR-1: title is required | `domain.ValidateNewTask` — repoRoot/internal/domain | Applies to all task creation paths |
| BR-2: exit 2 for invalid input (title empty) | CLAUDE.md exit code contract | Usage error, not runtime error |
| BR-3: exit 1 for runtime errors (editor not found, not in git repo) | CLAUDE.md exit code contract | |
| BR-4: binary never auto-commits | CLAUDE.md C-03 | Task file saved, not committed |
| BR-5: task file writes are atomic | CLAUDE.md C-02 | Write to .tmp then os.Rename |
| BR-6: editor resolution: $EDITOR, fallback vi | `openEditor()` in usecases/edit_task.go | Reuse existing function; do not duplicate |
| BR-7: success format identical to `kanban new <title>` | Consistency contract | See shared-artifacts-registry.md |

---

## Functional Requirements

### FR-1: No-argument detection

When `kanban new` is invoked with zero positional arguments, the command routes to editor mode rather than immediately attempting to create a task with an empty title.

### FR-2: Pre-flight checks before opening editor

Before opening the editor, the binary validates:
- The working directory is a git repository (`git.RepoRoot()`)
- Git identity is configured (`git.GetIdentity()`)
- kanban is initialised in the repository (`config.Read()`)

Pre-flight failure → appropriate stderr message + exit code (per existing error messages in `new.go`) without opening the editor.

### FR-3: Blank task template

The editor opens with a temp file containing:
- Comment lines explaining usage (title required, other fields optional, comment lines ignored)
- Empty `title`, `priority`, `assignee`, `description` fields in the same YAML format used by `kanban edit`

The template is produced by calling `EditFilePort.WriteTemp` with a zero-value `domain.Task`.

### FR-4: Same editor resolution as `kanban edit`

Editor is resolved by reading `$EDITOR` environment variable, falling back to `vi`. This must use the same `openEditor()` function already in `usecases/edit_task.go` rather than a new independent implementation.

### FR-5: Title validation after editor exits

After the editor exits, `EditFilePort.ReadTemp` parses the temp file. If `title` is empty after trimming whitespace:
- Print to stderr: `title cannot be empty`
- Exit with code 2
- Do not create any task file

### FR-6: Task creation on valid input

If title is non-empty, call `AddTask.Execute` with all parsed fields (`Title`, `Priority`, `Assignee`, `Description`) and the git identity as `CreatedBy`.

### FR-7: Success output

Print the same two-line success message as `kanban new <title>`:

```
Created ${TASK_ID}: ${TASK_TITLE}
Hint: reference ${TASK_ID} in your next commit to start tracking
```

Exit with code 0.

---

## Non-Functional Requirements

| NFR | Threshold | Notes |
|-----|-----------|-------|
| Editor launch latency | < 100ms from command invocation to editor open | Pre-flight checks must not block perceptibly |
| Temp file cleanup | Always — even on error paths | Defer `os.Remove(tmpFile)` as in edit_task.go |
| TTY detection | Not required for this story | Editor mode inherits stdout/stdin/stderr from process (same as edit) |

---

## Constraints

- C-01: Do not auto-commit on task creation (CLAUDE.md C-03)
- C-02: Task file write must be atomic (CLAUDE.md C-02)
- C-03: `internal/domain` has zero external imports (CLAUDE.md architectural rule)
- C-04: No adapter package may import another adapter package
- C-05: `openEditor()` is currently a package-level function in `internal/usecases` — solution-architect must decide whether to export it, extract it to a shared helper, or duplicate it (duplication is explicitly the wrong choice)
