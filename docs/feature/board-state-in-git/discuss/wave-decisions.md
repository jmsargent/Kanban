# DISCUSS Wave Decisions — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Status**: COMPLETE
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Executive Summary

This document records the final decisions, tradeoffs, and handoff information from the DISCUSS wave for the feature "board-state-in-git." The wave produced 15 artifacts covering JTBD analysis, journey visualization, story mapping, user stories, acceptance criteria, outcome KPIs, and DoR validation.

**Three stories are ready for the DESIGN wave.** One story carries a size conditional.

---

## What Was Decided

### Decision 1: Walking Skeleton = US-BSG-01 (`kanban log TASK-ID`)

This is a brownfield project. The backbone activities (create task, start work, commit, CI done, view board) all work today. The walking skeleton for this feature is the thinnest new end-to-end slice that delivers primary value: surfacing the audit trail.

`kanban log TASK-ID` delivers:
- The highest-scoring opportunity (OS-1: 16, OS-8: 14, OS-10: 14)
- Zero architectural risk
- 1-day effort
- Validation of the core assumption before the larger architectural investment

**Rationale**: The DISCOVER wave explicitly recommended this sequencing. The DISCUSS wave confirms it.

---

### Decision 2: US-BSG-02 (Concept B) is gated on US-BSG-01 validation

US-BSG-02 is a DoR-ready story but carries a DISCOVER-wave gate: proceed only after US-BSG-01 ships and is evaluated over 2-4 weeks. If the audit trail need is fully satisfied by `kanban log`, the architectural change (Concept B) may not be necessary.

**Gate criteria** (to be evaluated 2-4 weeks after US-BSG-01 ships):
1. Does `kanban log` fully satisfy the audit trail need?
2. Does the hook-writes-files pain still feel significant after seeing Phase 1 in practice?
3. Is the diff-noise pain still observed?

If criteria 2 or 3 are YES, proceed to US-BSG-02. If only criterion 1 is YES (audit trail satisfied), deprioritize US-BSG-02.

---

### Decision 3: US-BSG-03 (`kanban board --me`) ships independently

No dependency on US-BSG-01 or US-BSG-02. Can ship in any order. For a solo developer the urgency is low. For a multi-developer project it becomes "should have." Deferred to after the walking skeleton is validated.

---

### Decision 4: Concept B log format (deferred to DESIGN wave)

The exact format of each transitions.log line is left as an open question (RQ-1) for the DESIGN wave. The product-owner requirements specify:
- Fields required: `<ISO8601_UTC> <TASK-ID> <from>-><to> <author_email> <trigger>`
- Append-only
- Human-readable as plain text (must be `cat`-able and `grep`-able)
- No JSON, no binary, no database

The exact field separator, quoting rules, and trigger format are DESIGN wave decisions.

---

### Decision 5: Migration path (deferred to DESIGN wave)

How existing repos with `status:` in task YAML migrate to the transitions log format is an open question (RQ-4) for the DESIGN wave. The product-owner constraint: migration must not lose any current in-progress task state. The migration story will be created in the DESIGN wave after US-BSG-02 design is complete.

---

## What Was NOT Decided (DESIGN Wave Scope)

| Topic | Why Deferred |
|-------|-------------|
| TransitionLogRepository interface methods and signatures | DESIGN wave — architecture decision |
| Exact transitions.log line format | DESIGN wave — see RQ-1 |
| File locking strategy (advisory vs. mandatory, mutex vs. flock) | DESIGN wave — implementation detail |
| `kanban log` output format (exact columns, widths, color scheme) | DESIGN wave — CLI UX detail |
| Migration command design (`kanban migrate` vs. auto-migration on first run) | DESIGN wave — see RQ-4 |
| Whether `kanban log` gets `--json` output in Phase 1 | DESIGN wave — optional enhancement |

---

## Tradeoffs Documented

### Tradeoff 1: Transitions log vs. no architectural change

The current system's `git log -- .kanban/tasks/TASK-NNN.md` already provides an audit trail. The transitions log (US-BSG-02) adds architectural complexity for a marginal improvement in audit trail quality.

**Resolved by phasing**: Ship `kanban log` first. Evaluate. Proceed to Concept B only if validated.

### Tradeoff 2: Full Concept B vs. partial Concept B

Option: implement transitions log for the hook only (remove task file writes from the hook) without changing `kanban start` or `kanban board`. This partial approach:
- Pros: smaller scope, immediate diff cleanliness improvement
- Cons: inconsistent — `kanban start` still writes to task file, creating two sources of truth

**Resolved**: The DISCUSS wave recommends the full Concept B implementation. Partial implementations create inconsistency that is harder to reason about than the current clean state. If scope must be reduced, use the Elephant Carpaccio split documented in story-map.md (split into US-BSG-02a/02b/02c).

### Tradeoff 3: Transitions log performance vs. current YAML approach

Current `kanban board`: O(n tasks) — reads n task files.
Phase 2 `kanban board`: O(n tasks) + O(n transitions) — reads n task files + 1 log file.

This is a performance concern only at extreme scale (10,000+ transitions). For the expected project lifetime (50-200 tasks, 500-2000 transitions), performance is equivalent or better than the current approach.

**Resolved**: Not a performance concern within expected product usage range. Guardrail metric: `kanban board` response time must not degrade below current baseline.

---

## Artifacts Produced

| Artifact | Path | Status |
|---------|------|--------|
| JTBD Job Stories | discuss/jtbd-job-stories.md | Complete |
| JTBD Four Forces | discuss/jtbd-four-forces.md | Complete |
| JTBD Opportunity Scores | discuss/jtbd-opportunity-scores.md | Complete |
| Journey Visual | discuss/journey-task-state-transitions-visual.md | Complete |
| Journey YAML Schema | discuss/journey-task-state-transitions.yaml | Complete |
| Journey Gherkin | discuss/journey-task-state-transitions.feature | Complete |
| Shared Artifacts Registry | discuss/shared-artifacts-registry.md | Complete |
| Story Map | discuss/story-map.md | Complete |
| Prioritization | discuss/prioritization.md | Complete |
| Requirements | discuss/requirements.md | Complete |
| User Stories | discuss/user-stories.md | Complete |
| Acceptance Criteria | discuss/acceptance-criteria.md | Complete |
| DoR Checklist | discuss/dor-checklist.md | Complete |
| Outcome KPIs | discuss/outcome-kpis.md | Complete |
| Wave Decisions | discuss/wave-decisions.md | This document |

---

## Handoff Package for DESIGN Wave

### Stories Ready for Design

| Story | Effort Estimate | DoR Status | Priority |
|-------|----------------|-----------|---------|
| US-BSG-01: `kanban log TASK-ID` | 1 day | PASSED | P1 — build first |
| US-BSG-02: Append-only transitions log | 8-10 days | PASSED (sizing conditional) | P2 — after US-BSG-01 validated |
| US-BSG-03: `kanban board --me` | 1 day | PASSED | P3 — independent |

### Key Constraints for Solution Architect

1. `internal/domain` has zero imports from non-stdlib (CLAUDE.md)
2. The commit-msg hook always exits 0 — wrapped in `recover()` (CLAUDE.md)
3. All file writes are atomic (CLAUDE.md)
4. The binary never auto-commits except `kanban ci-done` (CLAUDE.md)
5. Exit codes: 0=success, 1=runtime error, 2=usage error
6. Architecture: hexagonal. New `TransitionLogRepository` is a new secondary port (ports.go + new filesystem adapter)
7. ADR-002, ADR-004, ADR-005 require updates. A new ADR for the transitions log format is required.
8. `kanban board` output format must remain unchanged (user-visible behavior preserved)

### Open Questions for DESIGN Wave

| RQ | Question | Impact |
|----|----------|--------|
| RQ-1 | Exact transitions.log line format | Parsing + LatestStatus/History implementations |
| RQ-2 | `git log --follow` reliability on renamed task files | Phase 1 edge case documentation |
| RQ-3 | `kanban log --json` output | Optional Phase 1 enhancement |
| RQ-4 | Migration path from YAML status to transitions log | Prerequisite for US-BSG-02 on existing repos |

### Outcome KPIs for Platform Architect

KPI-1 and KPI-4 are developer self-report — no instrumentation infrastructure needed.
KPI-2 is measured by `git show` — no new infrastructure needed; can be an acceptance test assertion.
KPI-3 is measured by automated integration test — no new infrastructure needed.

No observability infrastructure changes required for this feature.
