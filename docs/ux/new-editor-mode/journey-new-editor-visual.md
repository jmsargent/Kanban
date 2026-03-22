# Journey Visual: kanban new (editor mode)

## Persona

**Alex Chen** — a developer using kanban as their personal task tracker within a git repo.
Alex is mid-flow on a coding session and wants to capture a new task without breaking their mental context.

## Emotional Arc

```
Start: Slightly interrupted (needs to context-switch to create a task)
  |
  v
Invoke: Familiar — types a single command they already know
  |
  v
Editor open: Focused — the editor is their home, they know exactly what to do
  |
  v
Save + quit: Confident — the tool confirms what was created, no ambiguity
  |
  v
End: Relieved — task is captured, back to coding without friction
```

Arc pattern: **Problem Relief** (interrupted -> focused -> relieved)

---

## Happy Path Flow

```
[Alex types: kanban new]
         |
         | no title arg detected
         v
[Binary checks: EDITOR env var set?]
         |
         | yes ($EDITOR = "nvim")
         v
[WriteTemp: blank task template to /tmp/kanban-new-XXXX.yaml]
         |
         v
[Editor opens with blank template]
         |
         | Alex fills in: title, optional priority/assignee/description
         | Alex saves and quits editor (:wq)
         v
[ReadTemp: parse fields from temp file]
         |
         v
[Validate: title non-empty?]
         |
         | yes
         v
[AddTask.Execute: persist task to .kanban/tasks/TASK-042.md]
         |
         v
[stdout: "Created TASK-042: Fix nil pointer in auth handler"]
[stdout: "Hint: reference TASK-042 in your next commit to start tracking"]
         |
         v
[Exit 0]
```

---

## TUI Mockups

### Step 1: Invocation

```
$ kanban new
```

No output yet — editor opens immediately (< 100ms).

---

### Step 2: Editor Buffer (what Alex sees in their $EDITOR)

```
+-- /tmp/kanban-new-7f3a.yaml ----------------------------------------+
|                                                                       |
|  # New task — fill in the fields below, then save and quit.          |
|  # Lines beginning with # are ignored.                               |
|  # title is required; all other fields are optional.                 |
|                                                                       |
|  title: ""                                                            |
|  priority: ""                                                         |
|  assignee: ""                                                         |
|  description: ""                                                      |
|                                                                       |
+-----------------------------------------------------------------------+
```

Alex fills in:

```
+-- /tmp/kanban-new-7f3a.yaml ----------------------------------------+
|                                                                       |
|  # New task — fill in the fields below, then save and quit.          |
|  # Lines beginning with # are ignored.                               |
|  # title is required; all other fields are optional.                 |
|                                                                       |
|  title: "Fix nil pointer in auth handler"                             |
|  priority: "P1"                                                       |
|  assignee: ""                                                         |
|  description: "Crashes on logout when session token is nil"           |
|                                                                       |
+-----------------------------------------------------------------------+
```

Alex saves and quits. Control returns to the binary.

---

### Step 3: Success Output

```
$ kanban new
Created TASK-042: Fix nil pointer in auth handler
Hint: reference TASK-042 in your next commit to start tracking
```

---

## Error Path: Editor exits with empty title

```
[Alex opens editor, deletes title line, saves and quits]
         |
         v
[ReadTemp: title field is empty string]
         |
         v
[Validate: title is empty]
         |
         v
[stderr: "title cannot be empty"]
[Exit 2]  (usage error — user provided no title)
```

TUI mockup:

```
$ kanban new
title cannot be empty
```

---

## Error Path: $EDITOR not set, vi not available

```
[EDITOR env var is empty string]
         |
         v
[openEditor falls back to "vi"]
         |
         v
[exec.Command("vi", tmpFile).Run() returns error]
         |
         v
[stderr: "Error: open editor: <system error>"]
[Exit 1]
```

TUI mockup:

```
$ kanban new
Error: open editor: exec: "vi": executable file not found in $PATH
```

---

## Error Path: Not in a git repository

```
[git.RepoRoot() returns error]
         |
         v
[stderr: "Not a git repository"]
[Exit 1]
```

---

## Error Path: kanban not initialised

```
[config.Read() returns ErrNotInitialised]
         |
         v
[stderr: "kanban not initialised — run 'kanban init' first"]
[Exit 1]
```

---

## Integration with Existing `kanban new <title>` Behaviour

When a title argument IS provided, behaviour is unchanged:

```
$ kanban new "Fix nil pointer"
Created TASK-042: Fix nil pointer
Hint: reference TASK-042 in your next commit to start tracking
```

The editor mode is activated only when zero arguments are given.

---

## Shared Artifacts

| Variable | Source | Appears In |
|---|---|---|
| `${TASK_ID}` | `tasks.NextID(repoRoot)` | success stdout, task file name |
| `${TASK_TITLE}` | editor temp file, parsed by ReadTemp | success stdout, task file content |
| `${EDITOR}` | `os.Getenv("EDITOR")`, fallback `"vi"` | openEditor invocation |
| `${TMP_FILE}` | `editor.WriteTemp(emptyTask)` | editor invocation path, ReadTemp path |
