# ADR-002: Task File Format — Markdown with YAML Front Matter

**Status**: Accepted
**Date**: 2026-03-15
**Feature**: kanban-tasks
**Resolves**: OD-01

---

## Context

Task files are stored as `.kanban/tasks/TASK-NNN.md` in the git repository. Three parties read and write these files:

1. The `kanban` CLI (programmatic reads and writes to all fields)
2. The git commit-msg hook (reads `status` field; writes `status` field; must be fast and require no external parser)
3. Developers directly in their editor (via `kanban edit` which opens `$EDITOR`)

The format must satisfy competing requirements:

- **Machine parseability**: hook and CI step must reliably extract and update the `status` field
- **Human readability**: developers edit these files directly; the format must be legible without tooling
- **Git diff friendliness**: a status-only update (the most common write) should produce a 1-line diff with minimal noise
- **Extensibility**: future fields (description, labels, sub-tasks) must not require a format migration
- **Parse simplicity**: no dependency on external schema validators; standard library or minimal-dependency parsers suffice

Required structured fields (from requirements.md): `title`, `status`, `priority`, `due`, `assignee`.
Optional freeform field: `description` (multi-line prose, referenced in US-06).

---

## Decision

Task files use **Markdown with YAML front matter**.

Structure:
```
---
id: TASK-001
title: Fix OAuth login bug
status: todo
priority: P2
due: 2026-03-20
assignee: Rafael Rodrigues
---

Optional freeform description in Markdown body.
```

The YAML front matter block is delimited by `---` at lines 1 and N. All structured fields live in the front matter. The Markdown body contains the freeform description.

The `status` field is always a single line in the front matter: `status: <value>`. The hook and CI step update it with a line-level replacement (no YAML parser required; a single regex substitution suffices for the status field: `s/^status: .*/status: in-progress/`).

---

## Alternatives Considered

### Alternative 1: Pure YAML

```yaml
id: TASK-001
title: Fix OAuth login bug
status: todo
priority: P2
due: 2026-03-20
assignee: Rafael Rodrigues
description: |
  Reproduces on Chrome and Firefox when using Google OAuth.
```

Evaluation:
- Machine parseability: excellent -- any YAML library handles it
- Human readability: good for structured fields; multi-line description requires block scalar syntax (`|`) which is less natural for prose
- Git diff: clean single-line diff for status updates
- Extensibility: good
- Developer experience: the `.md` extension convention is lost; developers cannot open task files and see syntax highlighting without configuring their editor for YAML

Rejection rationale: The `.md` extension signals to developers that task files are prose-friendly documents. Pure YAML loses this signal and makes the description field awkward. The Markdown body in front-matter format is strictly superior for human authoring while being equally machine-parseable for the structured fields.

### Alternative 2: TOML

```toml
id = "TASK-001"
title = "Fix OAuth login bug"
status = "todo"
priority = "P2"
due = "2026-03-20"
assignee = "Rafael Rodrigues"

[description]
body = """
Reproduces on Chrome and Firefox.
"""
```

Evaluation:
- Machine parseability: requires a TOML parser; fewer standard-library options than YAML
- Human readability: slightly more verbose than YAML for this use case
- Git diff: clean
- Extensibility: adequate
- Developer experience: TOML is less widely known among developers than YAML or Markdown

Rejection rationale: TOML adds a parser dependency without providing advantages over YAML front matter for this use case. The description representation in TOML is more awkward than a Markdown body.

### Alternative 3: Plain JSON

```json
{
  "id": "TASK-001",
  "title": "Fix OAuth login bug",
  "status": "todo",
  "priority": "P2",
  "due": "2026-03-20",
  "assignee": "Rafael Rodrigues"
}
```

Evaluation:
- Machine parseability: excellent
- Human readability: poor for multi-line descriptions; JSON escaping is hostile to prose
- Git diff: every field change produces braces/comma noise; a status update diffs multiple lines
- Developer experience: JSON is not a natural writing format; `kanban edit` opening JSON in `$EDITOR` produces friction

Rejection rationale: JSON git diffs are noisy (confirmed risk R-03 from DISCUSS wave specifically called out the need for minimal diff surface area). JSON description fields require escaping. Ruled out.

---

## Consequences

**Positive**:
- Status updates produce 1-line diffs: `status: todo` -> `status: in-progress` (minimises merge conflict risk, addresses R-03)
- `.md` extension means any editor with Markdown support renders task files readably
- Markdown body enables rich descriptions without format migration
- YAML front matter is parseable by the standard front-matter parsing libraries in any language ecosystem
- Hook and CI step can update `status` with a line-level string replacement, requiring no full YAML parse for the hot path

**Negative**:
- Front matter parsing requires a small parser (e.g., `gray-matter` for TypeScript) -- not zero dependencies, but a single well-maintained library
- Developers editing files directly must understand the `---` delimiter convention; deviating from it corrupts the file

**Enforcement**: The `KanbanFileSystemAdapter` validates front matter structure on read. A corrupted front matter produces a recoverable error with guidance, not a crash.
