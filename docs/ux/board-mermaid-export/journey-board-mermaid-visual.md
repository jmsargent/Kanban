# Journey Map — Board Mermaid Export

**Feature**: board-mermaid-export
**Research depth**: Lightweight
**Date**: 2026-03-23

---

## User Journey: Generate Mermaid Diagram from Kanban Board

```
TRIGGER                  STEP 1                   STEP 2                   STEP 3                   OUTCOME
──────────────────────────────────────────────────────────────────────────────────────────────────────────────
Developer wants          Run kanban board         Inspect Mermaid          Paste/redirect           README or
a shareable view    ──► with --mermaid flag  ──► output on stdout    ──► output to file       ──► doc updated
of the board
```

---

## Step Detail

### Step 1 — Invoke the command

**Action**: `kanban board --mermaid` (optionally with `--me`)
**Pre-condition**: Inside a git repo with kanban initialised
**Expected output**: Mermaid kanban block printed to stdout, exit 0
**Error paths**:
- Not a git repository → stderr "Not a git repository", exit 1
- kanban not initialised → stderr "kanban not initialised — run 'kanban init' first", exit 1
- `--mermaid` and `--json` used together → stderr "flags --mermaid and --json are mutually exclusive", exit 2

### Step 2 — Review output

**Observable**: Developer sees a fenced Mermaid block on stdout
**Format**:
```
\`\`\`mermaid
---
config:
  kanban:
    ticketBaseUrl: ''
---
kanban
  Todo
    TASK-001@{ label: "Fix login bug" }
    TASK-002@{ label: "Write tests" }
  In Progress
    TASK-003@{ label: "Refactor auth" }
  Done
    TASK-004@{ label: "Setup CI" }
\`\`\`
```
**Empty board**: Valid Mermaid block with columns but no task nodes — not an error
**Filtered board** (`--me`): Only tasks assigned to current git user appear as nodes

### Step 3 — Use the output

**Action**: Redirect stdout or copy-paste into README / docs
**Canonical usage**:
```sh
kanban board --mermaid > docs/board.md
kanban board --mermaid | pbcopy
```
**No file-writing performed by the command itself** — stdout only, composable with shell tools

---

## Emotional Arc

```
Step 1  ──────────────────────────────────────────────────── Step 3
        Neutral                 Satisfied              Confident
        (running a             (diagram looks          (README now
         familiar command)      right in preview)       stays current)
```

The arc is flat and positive. No anxiety spikes — the command follows the existing `--json` pattern the developer already knows.

---

## Shared Artifacts

| Artifact | Source | Consumed by |
|----------|--------|-------------|
| `domain.Board` | `GetBoard` use case | `printBoardMermaid` in CLI adapter |
| `domain.Column.Label` | board config | Mermaid section headers |
| `domain.Task.ID` + `domain.Task.Title` | task YAML files | Mermaid task nodes |

---

## Out of Scope (this feature)

- Writing output directly to README or any file
- Custom column order beyond what config defines
- Rendering task metadata (priority, due, assignee) in Mermaid nodes (future enhancement)
- Mermaid graph/flowchart format (kanban diagram only)
