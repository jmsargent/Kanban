# Journey: Task Creation with Creator Attribution — Visual Map

## Actor

Developer managing work via the terminal inside a git-initialized kanban workspace.
May be a solo developer or a member of a small team sharing a repository.

---

## Journey Map

| Step | Actor Action | System Response | Emotional State | Shared Artifact |
|------|-------------|-----------------|-----------------|-----------------|
| 1 | Opens terminal in project directory | — | Confident — familiar environment | — |
| 2 | Runs `kanban new "Fix login bug"` | Reads git identity from git config | Anticipating — quick capture expected | `git config user.name` |
| 3 | — | Generates next task ID | — | `TASK-NNN` (ID counter) |
| 4 | — | Writes task file with `created_by: {name}` in front matter | — | `.kanban/tasks/TASK-NNN.md` |
| 5 | — | Prints `Created TASK-NNN: Fix login bug` to stdout | Satisfied — task captured with authorship | — |
| 6 | Runs `kanban board` | Reads all tasks; renders board with Created By column | Confident — can see who raised what | — |

---

## Emotional Arc

```
Confidence
    │                                                    ★  (sees their name on the board)
    │                                  ★  (confirmation message with task ID)
    │              ★  (command accepted, no error)
    │ ★  (types the command — low friction expected)
    └──────────────────────────────────────────────────▶  Steps
       1           2           3           4           6
```

Arc direction: **builds steadily** — no dips on the happy path. The developer's mental model
("I type a command, it records a task") is satisfied without surprising gaps.

---

## Happy Path: Git Identity Is Set

```
Developer  →  kanban new "Fix login bug"
           →  [read git config user.name] → "Jonathan Sargent"
           →  Task{ID: TASK-002, Title: "Fix login bug", CreatedBy: "Jonathan Sargent"} written
           →  stdout: "Created TASK-002: Fix login bug"

Developer  →  kanban board
           →  TASK-002 │ Fix login bug │ -- │ -- │ unassigned │ Jonathan Sargent
```

---

## Error Path 1: Git Identity Not Configured

```
Developer  →  kanban new "Fix login bug"
           →  [read git config user.name] → empty / not set
           →  exit 1
           →  stderr: "git identity not configured — run:
                         git config --global user.name \"Your Name\"
                         git config --global user.email \"you@example.com\""
```

Emotional impact: **brief frustration → rapid recovery**. The message is self-contained;
the developer does not need to search the web or read documentation.

---

## Error Path 2: Pre-Existing Tasks on Board

```
Developer  →  kanban board
           →  TASK-001 (created before this feature, no created_by field in front matter)
           →  TASK-001 │ Old task title │ P1 │ 2026-03-01 │ alice │ --
```

Emotional impact: **neutral**. The `--` placeholder is consistent with how missing
priority and assignee are already displayed — no surprise.

---

## Mental Model Validated

> "When I create a task, my name is captured automatically. I don't need to type it.
>  If git doesn't know who I am, the tool tells me exactly how to fix that.
>  When I look at the board, I can see who created each item."
