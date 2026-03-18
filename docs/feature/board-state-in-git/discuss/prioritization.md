# Prioritization: Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Release Priority

**Prioritization formula**: Value (1-5) x Urgency (1-5) / Effort (1-5) = Priority Score
**Tie-breaking**: Walking Skeleton > Riskiest Assumption > Highest Value

| Priority | Release | Target Outcome | Stories | Value | Urgency | Effort | Score | Rationale |
|----------|---------|---------------|---------|-------|---------|--------|-------|-----------|
| 1 | Walking Skeleton | Audit trail surfaced from existing git history | US-BSG-01 | 5 | 5 | 1 | 25.0 | Highest-scoring opportunities (16, 14, 14); 1-day effort; zero risk; validates core assumption before architectural investment |
| 2 | Release 1: Architectural Audit Trail | Hook no longer writes task files; log is the state source | US-BSG-02 | 5 | 3 | 4 | 3.75 | Only build after WS validates that audit trail alone is insufficient; addresses OS-2 (13), OS-3 (13); higher effort |
| 3 | Release 3: Per-User Board | Developer sees only their own tasks with --me | US-BSG-03 | 3 | 2 | 1 | 6.0 | Independent; low urgency for solo developer; OS-6 score: 10 (appropriately served) |
| 4 | Release 2: Migration | Existing repos can migrate to transitions log | (future story) | 4 | 2 | 3 | 2.67 | Prerequisite for Concept B on existing repos; not needed until R1 ships |

---

## Rationale by Story

### US-BSG-01: `kanban log TASK-ID` (Walking Skeleton)

**Build now. Unconditionally.**

The three highest-scoring outcomes (OS-1: 16, OS-8: 14, OS-10: 14) are all addressed by this single 1-day command. The DISCOVER wave explicitly recommended this as the first delivery: "build `kanban log TASK-001` first — 2 hours implementation, directly validates the primary pain."

The riskiest assumption for this feature is: "Does surfacing the existing git history satisfy the audit trail need, or does the developer need the structured transitions log format?" This assumption can only be validated by shipping US-BSG-01 and using it for 2-4 weeks.

**Decision gate**: After shipping US-BSG-01, evaluate:
- Is the audit trail need satisfied? If YES: deprioritize US-BSG-02.
- Does the hook-writes-files pain remain? If YES: proceed to US-BSG-02.
- Is the diff-noise pain still felt? If YES: proceed to US-BSG-02.

### US-BSG-02: Concept B Transitions Log (Release 1)

**Build after US-BSG-01 validation.**

This story addresses OS-2 (rebase safety: 13) and OS-3 (no task file writes: 13). It has an 8-10 day implementation scope and changes the internal architecture of how state is stored. The DISCOVER wave rated this "CONDITIONAL — only proceed if Phase 1 does not fully resolve the audit trail need."

Building US-BSG-02 without first shipping US-BSG-01 would be premature investment in architectural change when the primary pain (audit trail) might be solved by the simpler approach.

### US-BSG-03: `kanban board --me` (Release 3)

**Build independently, any time.**

This story has zero dependency on US-BSG-01 or US-BSG-02. It addresses OS-6 (per-user view: 10) — an "appropriately served" opportunity. The effort is 1 day (filter tasks by assignee field). It can be shipped before, after, or in parallel with the other stories without affecting them.

For a solo developer, the urgency is low. For a 2+ developer team, the urgency rises. Priority is deferred until the WS and R1 work are validated.

---

## Backlog Suggestions

> **Note**: Story IDs (US-BSG-01 through US-BSG-03) are confirmed Phase 4 identifiers. Dependencies reflect Phase 2.5 analysis — verify in Phase 4 after full requirements are crafted.

| Story | Release | Priority | Outcome Link | Dependencies | MoSCoW |
|-------|---------|----------|-------------|--------------|--------|
| US-BSG-01: kanban log TASK-ID | Walking Skeleton | P1 | OS-1 (16), OS-8 (14), OS-10 (14) | None | Must Have |
| US-BSG-02: Append-only transitions log | Release 1 | P2 | OS-2 (13), OS-3 (13) | US-BSG-01 validated | Should Have |
| US-BSG-03: kanban board --me | Release 3 | P3 | OS-6 (10) | None | Could Have |
| Migration story (unnamed) | Release 2 | P4 | Enables US-BSG-02 on existing repos | US-BSG-02 | Should Have (when R1 ships) |

---

## Risk Register

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| US-BSG-01 fully satisfies audit trail; US-BSG-02 not needed | Medium | Positive — saves 8-10 days | Ship WS first; evaluate explicitly after 2-4 weeks |
| US-BSG-02 implementation reveals hidden complexity (>10 days) | Medium | Medium | Elephant Carpaccio split available (see story-map.md) |
| `git log --follow` unreliable on task files that have been moved | Low | Medium | Test with moved files in integration tests; document limitation in US-BSG-01 AC |
| Developer confusion when status disappears from task YAML | Medium | Low | In-file comment in task template; migration guide |
| Transition log merge conflict on concurrent appends | Low | Low | File locking in TransitionLogRepository adapter; document resolution strategy |
