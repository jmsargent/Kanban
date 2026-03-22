# Prioritization — explicit-state-transitions

**Feature**: explicit-state-transitions
**Date**: 2026-03-22

---

## Priority Order

All stories are P0 (walking skeleton). Deliver as a single cohesive change.

| Story | Priority | Effort | Rationale |
|-------|----------|--------|-----------|
| US-EST-05: Remove transitions.log | P0 | S (1d) | Foundational — removes the port/adapter that other stories depend on removing |
| US-EST-04: Remove hook | P0 | XS (0.5d) | Removes the auto-transition trigger; parallel with US-EST-05 |
| US-EST-03: Fix board | P0 | XS (0.5d) | Restores YAML as state source — needed before done/ci-done can be meaningful |
| US-EST-01: `kanban done` | P0 | XS (0.5d) | New explicit transition command; replaces what hook used to automate |
| US-EST-02: Fix `kanban ci-done` | P0 | S (1d) | Removes auto-commit; keeps CI-driven identification of done tasks |

**Total estimate**: 3–4 days

---

## Rationale for Atomic Delivery

Shipping these stories separately would leave the board in an inconsistent state:
- Board reading from transitions.log (which nothing writes to anymore) → blank board
- Hook removed but board still reading log → no visible state changes

All five changes should ship in a single PR.
