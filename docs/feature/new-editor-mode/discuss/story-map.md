# Story Map: new-editor-mode

## User

**Alex Chen** — developer using kanban as a personal git-native task tracker

## Goal

Capture a new task interactively via the editor, without breaking coding flow by composing a title inline

---

## Backbone

| Invoke command | Open editor | Fill in details | Validate & create | Confirm |
|----------------|-------------|-----------------|-------------------|---------|
| Run `kanban new` with no args | Editor launches with blank template | Alex fills title (required), optional fields | Binary validates title non-empty, calls AddTask | Success message printed |
| Detect zero-arg case | Resolve $EDITOR or fallback to vi | Comment guidance visible in buffer | Empty title rejected with exit 2 | Hint message printed |
| Pre-flight: git repo + identity check | WriteTemp with blank task | Title trimmed of whitespace | Task file written atomically | Exit 0 |

---

### Walking Skeleton

The thinnest end-to-end slice that connects every activity:

1. **Invoke**: `kanban new` with no args routes to editor mode (not empty-title failure)
2. **Open editor**: editor opens with a blank template via `EditFilePort.WriteTemp` + `openEditor()`
3. **Fill in details**: user fills title field and saves
4. **Validate and create**: title validated non-empty; `AddTask.Execute` called
5. **Confirm**: `"Created TASK-042: Fix nil pointer in auth handler"` printed; exit 0

Walking skeleton = **US-01** (single story — this feature is a thin, isolated slice)

---

### Release 1: Working editor mode (the full feature)

All behaviour fits in one story. The walking skeleton IS the release.

Tasks included:
- Zero-arg detection routing to editor mode
- WriteTemp with blank task (zero-value `domain.Task`)
- openEditor() reuse (same helper as `kanban edit`)
- ReadTemp + title validation (exit 2 on empty)
- AddTask.Execute call with editor-sourced fields
- Success output identical to `kanban new <title>`

Outcome KPI targeted: developers capture tasks without leaving their editor mental model

---

## Scope Assessment: PASS

3 user stories in total (1 primary, 2 error-path stories embedded as scenarios) — fits within a single story with 7 UAT scenarios. Touches 2 modules (cli adapter + usecases). Estimated 1-2 days effort.

Note: this feature is brownfield and isolated. Walking skeleton is marked "No" in configuration because the binary already has a walking skeleton (existing `kanban new` and `kanban edit`). This story adds a new input path to an existing, working end-to-end flow.

---

## Story Summary

| Story | Activity Coverage | Notes |
|---|---|---|
| US-01: kanban new editor mode | All 5 activities | Single right-sized story; all scenarios fit within 7 UAT limit |
