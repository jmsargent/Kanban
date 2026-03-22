# Wave Decisions — new-editor-mode

## DISCUSS Wave — Decisions and Open Items

These are decisions made during the DISCUSS wave that the DESIGN wave must carry forward, plus open items requiring solution-architect resolution.

---

## Decisions Made (DISCUSS Wave)

### WD-01: Due date field excluded from editor template

**Decision**: The blank task template presented to the editor does NOT include a `due` field.

**Rationale**: The `due` field requires a specific format (`YYYY-MM-DD`) that is harder to validate from a YAML text field than from a flag with an explicit format error message. The `--due` flag is already available on the `kanban new <title>` path and works well for date entry. Adding `due` to the editor template would require additional validation logic in the `ReadTemp` → parse → convert pipeline. This is a future enhancement, not in scope for this story.

**Impact on DESIGN**: The `EditFilePort.WriteTemp` template for a zero-value task should not include a `due` field. If the filesystem adapter currently includes `due` in the edit template (for `kanban edit`), the new-task template must use a different template or omit the due field conditionally.

---

### WD-02: Empty title after editor exits → exit code 2 (usage error, not runtime error)

**Decision**: When the user saves the editor with an empty title, the exit code is 2 (usage error), not 1 (runtime error).

**Rationale**: An empty title is a user input error, equivalent to running `kanban new ""`. The existing `new.go` already maps `ErrInvalidInput` to exit 2. For consistency, the editor-mode empty title must produce the same exit code. This means the CLI adapter must validate the title before calling `AddTask.Execute` and issue `os.Exit(2)` directly, not let the use case return `ErrInvalidInput` (which would also produce exit 2, but the routing logic should be explicit and testable).

**Impact on DESIGN**: Validation of title non-emptiness happens in the CLI adapter (new.go RunE) after `ReadTemp`, before `AddTask.Execute`. Do not rely on `AddTask.Execute` returning `ErrInvalidInput` for this case.

---

### WD-03: openEditor() must be shared, not duplicated

**Decision**: The new editor-mode path must reuse the existing `openEditor()` function from `internal/usecases/edit_task.go`, not implement a separate editor-launch function.

**Rationale**: Duplication of the editor-launch logic would mean two places to maintain `$EDITOR` resolution, fallback behaviour, and stdin/stdout/stderr wiring. This creates a maintenance burden and risks divergence.

**Open item for solution-architect**: `openEditor()` is currently a package-level unexported function in `internal/usecases`. The solution-architect must decide the sharing strategy:

- Option A: Export `openEditor()` as `OpenEditor()` from `internal/usecases` (simple, but leaks a utility function into the package API)
- Option B: Extract to a new `internal/usecases/editor.go` or a shared helper package (cleaner separation)
- Option C: Make it a method on an `EditorPort` interface injected into both use cases (most architecturally pure, adds interface overhead for a simple function)

Duplication (Option D) is explicitly rejected by this wave decision.

---

### WD-04: Pre-flight checks run before editor opens

**Decision**: Git repo check, git identity check, and kanban-init check all run before the editor is opened.

**Rationale**: Opening an editor, letting the user fill in a task, then discovering kanban is not initialised is a frustrating flow. All state-checks that could prevent task creation must happen before the editor launches. This is consistent with how `new.go` already runs `git.RepoRoot()` and `git.GetIdentity()` before doing any work.

**Impact on DESIGN**: The config.Read() check (for `ErrNotInitialised`) must also move to before the editor opens in the new code path. In the existing `new.go`, this check happens inside `AddTask.Execute`. For editor mode, this must be pulled out and run pre-editor.

---

### WD-05: Success output format must be identical to kanban new <title>

**Decision**: The two-line success output (`Created ${TASK_ID}: ${TASK_TITLE}` + `Hint: reference ${TASK_ID}...`) must be byte-for-byte identical regardless of which input path (inline title vs. editor mode) was used.

**Rationale**: Scripts or users that parse this output must not need to distinguish between the two paths.

**Impact on DESIGN**: The success-printing code should be extracted to a shared function in `new.go` (or inlined identically) so both branches produce the same output. An acceptance test must assert byte equality.

---

## Open Items for Solution-Architect (DESIGN Wave)

| ID | Item | Context |
|----|------|---------|
| OI-01 | Sharing strategy for `openEditor()` | WD-03 — options A/B/C documented above |
| OI-02 | WriteTemp with zero-value domain.Task | Does the filesystem adapter handle empty fields correctly? Does it produce comment lines in the template? The filesystem adapter may need a separate "new task template" function, or WriteTemp may need to be updated to accept a flag/parameter indicating "new mode". |
| OI-03 | Config pre-flight check before editor | WD-04 — `config.Read()` must run before `editor.WriteTemp()` in the new code path. Confirm this is consistent with hexagonal architecture rules (config port call is in the use case today). |
| OI-04 | Template comment lines | The blank template must contain comment guidance. Determine whether WriteTemp produces comments today (for edit mode, comments would be noise), or whether a separate `WriteTempForNew()` method is needed on `EditFilePort`. |

---

## Constraints Passed to DESIGN Wave

All constraints from CLAUDE.md are active:
- C-01: Binary never auto-commits (no git add/commit in new code path)
- C-02: Task file writes are atomic (write to .tmp, then os.Rename — enforced by TaskRepository.Save)
- C-03: `internal/domain` zero external imports
- C-04: No adapter imports another adapter
- C-05: Exit codes: 0=success, 1=runtime error, 2=usage error (enforced by AC-05, AC-07, AC-08)
