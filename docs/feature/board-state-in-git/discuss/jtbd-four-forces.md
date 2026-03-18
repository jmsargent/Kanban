# JTBD Four Forces — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Overview

The Four Forces model (Klement) explains why a developer would switch from the current file-based status system to the transition log approach (Concept B). Both demand-generating and demand-reducing forces are mapped across the two proposed solution paths: `kanban log` quick-win and Concept B full redesign.

```
        PROGRESS (switching happens)
             ^
             |
Push of  ----+---- Pull of New
Current       |     Solution
Situation     |
             |
        NO PROGRESS (staying put)
             ^
             |
Anxiety  ----+---- Habit of
of New        |     Present
Solution      |
```

---

## Forces Analysis: `kanban log TASK-ID` (Quick Win)

### Demand-Generating

**Push** — Frustrations driving change:
- Running `git log -- .kanban/tasks/TASK-001.md` works but the output is generic commit messages, not task-state language. The developer sees "chore: update TASK-001" not "TASK-001: todo → in-progress."
- No CLI command exists to surface the audit trail. The data is in git history but inaccessible without knowing the right git plumbing command.
- Primary motivator (DISCOVER P5): "The audit trail feels natural with commits" — the user wants the history surfaced, not buried in git internals.

**Pull** — Attractiveness of the alternative:
- `kanban log TASK-001` outputs a clean, human-readable transition history in domain language.
- Zero architectural change — uses existing data already in git history.
- Can be built in ~1 day with zero migration risk.
- Immediately validates whether the audit trail need is the core driver or a proxy for deeper discomfort.

### Demand-Reducing

**Anxiety** — Fears about the new approach:
- "Will it show me complete history or only recent commits?" (bounded by `git log` depth)
- "What if the task file was moved or renamed? Does `--follow` work reliably?"
- "If a commit only modified code (no task file change), does that commit appear in the log?" (Yes — only file-touching commits appear, which is correct behavior)
- "Is the output format stable enough to trust?"

**Habit** — Inertia of current behavior:
- Developer already knows `git log -- .kanban/tasks/TASK-001.md` works (even if awkward).
- No new command to learn — low habit disruption, but also low differentiation from current approach.

### Assessment

| Factor | Detail |
|--------|--------|
| Switch likelihood | HIGH |
| Key blocker | Anxiety about `--follow` reliability on renamed files |
| Key enabler | Zero migration risk; instant gratification of the audit trail |
| Design implication | `kanban log` output must use domain language (todo, in-progress, done), not git commit message verbatim |

---

## Forces Analysis: Concept B — Append-Only Transition Log

### Demand-Generating

**Push** — Frustrations driving change:
- DISCOVER direct quote: "the hook writing back to files feels like the wrong place" — architectural intuition signal.
- Status field in task YAML means two separate concerns (definition + state) are collapsed into one file. When the hook fires, it is a side-effect mutation of a definitional artifact.
- `git diff` on any hook-fired commit shows a YAML status change mixed with real code changes — clutter in the diff.
- Merge conflicts on task files (P1 in DISCOVER) — low-frequency but non-zero risk on collaboratively edited repos.

**Pull** — Attractiveness of Concept B:
- Hook becomes append-only to `.kanban/transitions.log` — one line per transition, no task file touched.
- Task files become definition-only: title, priority, due, assignee, description. Status is no longer stored there.
- `kanban board` derives current state from the log — same output, different source.
- `kanban log TASK-001` reads from the log file directly — O(transitions), not O(commits).
- Rebase-safe: the log file is separate from the commit graph; rebasing commits does not affect the log.
- `kanban start` continues to work: it appends to the log without requiring a git commit.
- Merge conflicts on the log are trivially resolved: both appended lines are valid; keep both.

**Pull — Identity signal**: This change makes the product's "git-native" identity more accurate. Currently the product is "git-native for storage but file-mutation for state." After Concept B it is "git-native for storage AND state-aware via git history." The product more fully delivers on its value proposition.

### Demand-Reducing

**Anxiety** — Fears about Concept B:
- "What happens to existing task files that have `status:` in their YAML? Do I need to migrate?" (Yes — migration path required; no transition log entries means tasks are implicitly todo)
- "What if `.kanban/transitions.log` grows very large?" (O(transitions): 500 transitions = fast file read; not a concern for typical project lifetime)
- "Will `kanban board` be slower now that it reads two sources?" (No — file read on O(tasks) + O(transitions) is faster than the current approach at scale)
- "Can I still see current status without running `kanban log`?" (Yes — `kanban board` still shows current status per column; status just derives from the log now)
- "The status field disappearing from task files will confuse me when I open a task file directly." (Mitigated by a comment in the template pointing to transitions.log)

**Habit** — Inertia of current system:
- Current system works. Tasks transition correctly. The hook fires on commit. `kanban board` shows the right state.
- The developer is accustomed to seeing `status: in-progress` in the task YAML. Its absence may feel like missing information.
- `kanban edit` will open a task file without a status field — the developer must know to look at `kanban board` or `kanban log` for current state.
- Migration required: existing `status:` fields must be stripped or converted to log entries before Concept B goes live.

### Assessment

| Factor | Detail |
|--------|--------|
| Switch likelihood | MEDIUM-HIGH |
| Key blocker | Habit: `status:` field in YAML is visible and familiar; its removal requires trust in the log |
| Key enabler | Push: "hook writing back to files feels like the wrong place" — this is an architectural identity motivation, not just functional |
| Design implication | Migration path must be smooth (one command: `kanban migrate` or handled in init). In-file comment must orient developers who open task files directly. The log format must be readable as plain text (it will be read with `cat` or `grep`) |

---

## Forces Analysis: Per-User Board Filter (`--me` flag)

### Demand-Generating

**Push**: Full board on a multi-developer project shows all tasks; developers must mentally filter to find their own.

**Pull**: `kanban board --me` filters instantly using `git config user.email` — no additional configuration, git identity is the natural key.

### Demand-Reducing

**Anxiety**: Tasks where `assignee` was not set (e.g., tasks created by `kanban add` without explicit `--assignee`) will not appear in `--me` view even if the developer is the implicit owner. Silently hidden tasks are dangerous.

**Habit**: Scanning the full board manually has become fast for small teams. The pain scales with team size.

### Assessment

| Factor | Detail |
|--------|--------|
| Switch likelihood | LOW-MEDIUM (solo/small team scenario) |
| Key blocker | Anxiety: silently hidden unassigned tasks |
| Key enabler | Zero configuration; git identity is already set |
| Design implication | `kanban board --me` must warn when unassigned tasks exist in the repo: "3 unassigned tasks hidden — use `kanban board` to see all" |

---

## Force Balance Summary

| Solution | Push | Pull | Anxiety | Habit | Net Verdict |
|----------|------|------|---------|-------|-------------|
| `kanban log` quick win | Strong | High | Low | Low | POSITIVE — build now |
| Concept B full redesign | Strong | High | Medium | Medium | POSITIVE — build after validation |
| `kanban board --me` | Weak-Medium | Medium | Medium | Low | NEUTRAL — build independently, low priority |
