# JTBD Opportunity Scores — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Scoring Method

**Formula**: Opportunity Score = Importance + max(0, Importance - Satisfaction)
**Scale**: Importance 1-10 (importance of the outcome) | Satisfaction 1-10 (satisfaction with current solution)
**Source**: DISCOVER wave interview + codebase analysis (1 participant; solo-developer product — team-estimate confidence per jtbd-opportunity-scoring methodology)
**Data quality**: Team estimate (solo developer product owner). Treat scores as directional rankings, not absolute values.

---

## Outcome Statements and Scores

Outcome statements derived from the 8-step Universal Job Map (Ulwick) applied to the developer's task-management workflow.

### Job: "Manage and track task state transitions as part of my git workflow"

| # | Outcome Statement | Importance | Satisfaction | Score | Priority |
|---|-------------------|-----------|--------------|-------|----------|
| OS-1 | Minimize the time to see when a specific task last transitioned | 9 | 2 | 16 | Extremely Underserved |
| OS-2 | Minimize the likelihood of task state being silently lost during a git rebase | 8 | 3 | 13 | Underserved |
| OS-3 | Minimize the number of files modified in a commit by task-state side effects | 8 | 3 | 13 | Underserved |
| OS-4 | Minimize the likelihood that `kanban board` reflects stale or incorrect state | 9 | 6 | 12 | Appropriately Served |
| OS-5 | Minimize the time for `kanban board` to display current state | 8 | 7 | 9 | Overserved |
| OS-6 | Maximize the likelihood that any developer's board shows their own tasks first | 6 | 2 | 10 | Appropriately Served |
| OS-7 | Minimize the likelihood of merge conflicts when two developers transition tasks simultaneously | 5 | 6 | 4 | Overserved |
| OS-8 | Maximize the readability of task transition history without git expertise | 8 | 2 | 14 | Extremely Underserved |
| OS-9 | Minimize the architectural coupling between task state and task definition | 7 | 3 | 11 | Appropriately Served |
| OS-10 | Minimize the time to understand what happened to a task when something went wrong | 8 | 2 | 14 | Extremely Underserved |

---

## Top Opportunities (Score >= 12)

| Rank | Outcome Statement | Score | Solution Direction | Story |
|------|-------------------|-------|-------------------|-------|
| 1 | Minimize time to see when a task last transitioned (OS-1) | 16 | `kanban log TASK-ID` quick win | US-BSG-01 |
| 2 | Maximize readability of transition history without git expertise (OS-8) | 14 | `kanban log TASK-ID` CLI output format | US-BSG-01 |
| 3 | Minimize time to understand what happened to a task (OS-10) | 14 | `kanban log TASK-ID` + Concept B log | US-BSG-01, US-BSG-02 |
| 4 | Minimize likelihood of state lost during rebase (OS-2) | 13 | Concept B — separate log file (rebase-safe) | US-BSG-02 |
| 5 | Minimize files modified by state side effects (OS-3) | 13 | Concept B — hook writes to log, not task files | US-BSG-02 |
| 6 | Minimize likelihood of stale board state (OS-4) | 12 | Both paths (existing + log) | Guardrail KPI |

---

## Overserved Areas (Score < 10)

| Outcome Statement | Score | Implication |
|-------------------|-------|-------------|
| Minimize `kanban board` latency (OS-5) | 9 | Current O(tasks) scan is already fast; do not over-engineer performance |
| Minimize merge conflicts on simultaneous transitions (OS-7) | 4 | Current conflict risk is low; append-only log makes it near-zero — do not prioritize conflict resolution further |

---

## Solution Mapping by Score

### Quick Win Bucket (OS-1, OS-8, OS-10) — Score 14-16

All three are directly addressed by `kanban log TASK-ID`. This single command:
- Shows when the task last transitioned (OS-1)
- Presents history in human-readable form without git knowledge (OS-8)
- Explains the full sequence of events when something went wrong (OS-10)

**Implementation**: 1 day. Wraps `git log --follow -- .kanban/tasks/<id>.md` with domain-language formatting.

### Architectural Redesign Bucket (OS-2, OS-3) — Score 13

Both are addressed only by Concept B (append-only transition log):
- Rebase safety: log file is separate from commit graph (OS-2)
- No task file state writes from the hook (OS-3)

`kanban log` (quick win) does NOT address these. They require the full Concept B implementation.

### Per-User Board (OS-6) — Score 10

`kanban board --me` flag addresses OS-6 independently of the other two buckets. No dependency on OS-1, OS-2, or OS-3. Can ship as a standalone story.

---

## Opportunity Scoring Influence on Prioritization

The scoring confirms the DISCOVER wave sequencing recommendation:

1. **Phase 1**: Build `kanban log TASK-ID` — addresses the 3 highest-scoring outcomes (16, 14, 14) with 1 day of work
2. **Phase 2**: Build Concept B — addresses outcomes 13, 13 with 8-10 days of work; only proceed after Phase 1 validates the audit trail need
3. **Phase 3**: Build `kanban board --me` — addresses outcome 10 with 1 day of work; independent of Phases 1 and 2

---

## Data Quality Notes

- **Source**: Single participant (product owner / sole developer) — DISCOVER wave Interview 1
- **Confidence**: Low for absolute values; directional rankings are sound given strong qualitative signals
- **Gap**: 2-3 additional developers using kanban-tasks in their own projects would strengthen confidence significantly before committing to Phase 2 (Concept B)
- **Validation plan**: Ship `kanban log` (Phase 1). After 2-4 weeks of use, re-evaluate whether OS-2 and OS-3 (architectural coupling, rebase safety) remain high-priority based on actual experience rather than anticipated frustration
