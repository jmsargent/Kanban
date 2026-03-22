# Shared Artifacts Registry — new-editor-mode

## Purpose

Every `${variable}` that appears in the journey mockups and Gherkin scenarios is tracked here with its single source of truth and all consumers. Untracked variables are the primary cause of integration failures.

---

## Registry

### EDITOR

```yaml
EDITOR:
  source_of_truth: "os.Getenv(\"EDITOR\") in openEditor() — usecases/edit_task.go:85-89"
  fallback: "\"vi\" when env var is empty"
  consumers:
    - "openEditor(tmpFile) — existing function in usecases/edit_task.go, shared by kanban edit"
    - "kanban new (editor mode) — must call the same openEditor() helper, not a new implementation"
  owner: "usecases layer — openEditor() function"
  integration_risk: "MEDIUM — if new-editor-mode duplicates the resolution logic instead of reusing openEditor(), the two commands may behave differently when EDITOR contains arguments (e.g. 'code --wait')"
  validation: "Acceptance test: set EDITOR to a script that records its invocation path; verify both 'kanban edit' and 'kanban new' invoke the same path"
```

### TMP_FILE

```yaml
TMP_FILE:
  source_of_truth: "EditFilePort.WriteTemp(task) — ports/repositories.go:57"
  consumers:
    - "openEditor(tmpFile) — path passed to editor process as first argument"
    - "EditFilePort.ReadTemp(tmpFile) — parsed after editor process exits"
    - "os.Remove(tmpFile) — cleanup in defer"
  owner: "EditFilePort secondary adapter (filesystem)"
  integration_risk: "HIGH — if WriteTemp path and ReadTemp path diverge (e.g. different temp dir logic), the binary reads stale or wrong content"
  validation: "Unit test: WriteTemp returns a path; ReadTemp of that exact path returns the expected snapshot"
```

### TASK_TITLE

```yaml
TASK_TITLE:
  source_of_truth: "temp file title field, parsed by EditFilePort.ReadTemp"
  consumers:
    - "AddTask.Execute — input.Title field"
    - "domain.ValidateNewTask — must be non-empty"
    - "stdout success message: 'Created ${TASK_ID}: ${TASK_TITLE}'"
    - "task file content: YAML front matter title field"
  owner: "usecases/add_task.go via AddTaskInput"
  integration_risk: "HIGH — title is the required field; empty string must be caught before AddTask.Execute is called, not inside it (to match exit code 2, not 1)"
  validation: "Acceptance test: editor session sets title to empty; assert exit code 2 and stderr 'title cannot be empty'"
```

### TASK_ID

```yaml
TASK_ID:
  source_of_truth: "TaskRepository.NextID(repoRoot) — called inside AddTask.Execute"
  consumers:
    - "stdout line 1: 'Created ${TASK_ID}: ${TASK_TITLE}'"
    - "stdout line 2: 'Hint: reference ${TASK_ID} in your next commit...'"
    - "task file name: .kanban/tasks/${TASK_ID}.md"
    - "task file YAML front matter: id field"
  owner: "usecases/add_task.go — returned from Execute"
  integration_risk: "LOW — ID generation is already tested via existing 'kanban new <title>' path; new-editor-mode reuses the same AddTask.Execute call"
  validation: "Acceptance test: verify TASK_ID in stdout matches filename of created task file"
```

### SUCCESS_MESSAGE_FORMAT

```yaml
SUCCESS_MESSAGE_FORMAT:
  source_of_truth: "internal/adapters/cli/new.go lines 77-78"
  value: |
    Created ${TASK_ID}: ${TASK_TITLE}
    Hint: reference ${TASK_ID} in your next commit to start tracking
  consumers:
    - "kanban new <title> path (existing)"
    - "kanban new (editor mode) path (new)"
  owner: "cli/new.go — NewCreateCommand RunE"
  integration_risk: "MEDIUM — if editor mode prints a different success format, users who rely on parsing this output will break"
  validation: "Acceptance test: assert stdout byte-for-byte matches the existing path output"
```

---

## Integration Checkpoints

| Checkpoint | Risk | Validation |
|---|---|---|
| Editor resolution uses existing `openEditor()` | MEDIUM | Both commands invoke same binary path |
| WriteTemp → ReadTemp path consistency | HIGH | Unit test round-trip |
| Empty title exits code 2 (not code 1) | HIGH | Acceptance test: assert exit code = 2 |
| Success format identical to `kanban new <title>` | MEDIUM | Byte-comparison in acceptance test |
| No task file created when editor aborts | MEDIUM | Assert `.kanban/tasks/` unchanged after abort |

---

## Variables With No Integration Risk

These appear in mockups but are defined entirely within a single step:

| Variable | Reason |
|---|---|
| `TMP_FILE path` (e.g. `/tmp/kanban-new-7f3a.yaml`) | Ephemeral — created and removed within single command invocation |
| `EDITOR` value (e.g. `"nvim"`) | Read-only environment input; never written by the binary |
