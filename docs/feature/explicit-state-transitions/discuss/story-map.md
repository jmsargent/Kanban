# Story Map — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Backbone (User Activities)

```
[Remove auto-hook] → [Remove transitions.log] → [Add kanban done] → [Fix ci-done] → [Fix board]
```

---

## Story Map

| Activity | Walking Skeleton | Notes |
|----------|-----------------|-------|
| **Remove auto-hook** | US-EST-04: Remove commit-msg hook | Delete hook handler, install-hook command, .kanban/hooks/ |
| **Remove transitions.log** | US-EST-05: Remove TransitionLogRepository | Delete adapter, port, domain types; update tests |
| **Add `kanban done`** | US-EST-01: Explicit done via CLI | New command, YAML update only, no commit |
| **Fix `kanban ci-done`** | US-EST-02: CI updates files, no commit | Remove --commit flag behavior |
| **Fix `kanban board`** | US-EST-03: Board reads from YAML | Restore YAML status as state source |

---

## Dependency Order

```
US-EST-05 (remove log/port)
    ↓
US-EST-04 (remove hook) ──── can be parallel with US-EST-05
    ↓
US-EST-03 (fix board — restore YAML state source)
    ↓
US-EST-01 (kanban done)
    ↓
US-EST-02 (fix ci-done — no commit)
```

All five stories are in the walking skeleton — this is a focused reduction/correction feature, not an additive one.

---

## Release Slices

**Release 1 (Walking Skeleton — all stories)**

Deliver all five stories together. They form a coherent atomic change: remove the auto-commit infrastructure and replace with explicit CLI-driven transitions. Shipping them separately would leave the codebase in an intermediate broken state.

Estimated total: 3–4 days
