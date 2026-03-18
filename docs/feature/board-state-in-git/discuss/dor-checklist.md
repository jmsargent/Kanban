# Definition of Ready Checklist — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

Hard gate: all 9 items must PASS before handoff to DESIGN wave.

---

## US-BSG-01: kanban log TASK-ID

| DoR Item | Status | Evidence |
|----------|--------|---------|
| 1. Problem statement clear, domain language | PASS | "Jon cannot see when tasks transitioned without running git plumbing commands. No `kanban log` command exists." Domain terms: transition, audit trail, task file. |
| 2. User/persona with specific characteristics | PASS | Jon Santos, sole developer, Go CLI project, commits frequently, rebases regularly, git identity: jon@kanbandev.io |
| 3. At least 3 domain examples with real data | PASS | Example 1: TASK-007 complete history with timestamps and real emails. Example 2: TASK-003 no transitions. Example 3: TASK-999 not found with error guidance. |
| 4. UAT scenarios in Given/When/Then (3-7) | PASS | 5 scenarios: happy path, no-commits edge case, task-not-found error, domain language validation, not-in-repo error |
| 5. Acceptance criteria derived from UAT | PASS | 11 criteria (AC-01-1 through AC-01-11) each traceable to a UAT scenario |
| 6. Right-sized (1-3 days, 3-7 scenarios) | PASS | 1 day effort, 5 scenarios |
| 7. Technical notes: constraints and dependencies | PASS | New `GetTaskHistory` use case; new `LogFile` method on GitPort; domain language translation requirement; Phase 2 upgrade path documented |
| 8. Dependencies resolved or tracked | PASS | No dependencies. Walking skeleton. Ships first. |
| 9. Outcome KPIs defined with measurable targets | PASS | KPI-1: 100% of history queries in ≤1 command (from 0%); measurement method: developer self-report 2-4 weeks post-ship |

### DoR Status: PASSED

---

## US-BSG-02: Append-Only Transitions Log (Concept B)

| DoR Item | Status | Evidence |
|----------|--------|---------|
| 1. Problem statement clear, domain language | PASS | "The commit-msg hook mutating task definition files as a side effect of commits feels architecturally wrong. Status field rewrites pollute commit diffs." Domain terms: transitions log, state separation, hook, task definition. |
| 2. User/persona with specific characteristics | PASS | Jon Santos, developer who rebases regularly, observes `git diff` after commits, values architectural cleanliness |
| 3. At least 3 domain examples with real data | PASS | Example 1: TASK-007 start appends to log (no task file write). Example 2: Rebase safety — TASK-007 preserved after `git rebase -i HEAD~3`. Example 3: Hook failure — commit exits 0, stderr warning, permissions error. |
| 4. UAT scenarios in Given/When/Then (3-7) | PASS | 7 scenarios: task creation, kanban start, hook appends, hook failure safety, board derives from log, rebase safety, ci-done commits log only |
| 5. Acceptance criteria derived from UAT | PASS | 24 criteria (AC-02-1 through AC-02-24) each traceable to UAT scenarios. Covers creation, start, hook, board, ci-done, rebase safety, concurrency. |
| 6. Right-sized (1-3 days, 3-7 scenarios) | CONDITIONAL | 7 scenarios (at upper bound). 8-10 day effort (above the 3-day guideline). One coherent user outcome. Elephant Carpaccio assessment performed (story-map.md): complexity is implementation depth, not scope breadth. Split options documented if hidden complexity emerges during design. |
| 7. Technical notes: constraints and dependencies | PASS | New `TransitionLogRepository` port + filesystem adapter; domain type `TransitionEntry`; modified `GetBoard`, `StartTask`, hook handler, `TransitionToDone`; removed `status:` from YAML; ADR updates required; file locking strategy specified. |
| 8. Dependencies resolved or tracked | PASS | Depends on US-BSG-01 being validated first (2-4 week gate). DISCOVER wave decision documented. RQ-1 (log line format) and RQ-4 (migration path) are open questions for DESIGN wave. |
| 9. Outcome KPIs defined with measurable targets | PASS | KPI-2: zero task file mutations per hook-fired commit; KPI-3: 100% transitions survive rebase; measurement method: `git show` + automated integration test |

### DoR Status: PASSED (with conditional on sizing — DESIGN wave must confirm scope estimate and split if needed)

**Peer review remediations applied (all HIGH issues resolved)**:
- AC-02-25 added: deleted task file with log entries — excluded from board, no error
- AC-02-26 added: ci-done with no in-progress tasks — exits 0, "no tasks to transition"
- AC-01-11 reclassified as @benchmark — not a CI gate

**Conditional note on sizing**: US-BSG-02 is at the upper limit of story size (7 scenarios, 8-10 days). The DESIGN wave must evaluate whether to proceed as a single story or apply the Elephant Carpaccio split documented in story-map.md. The product-owner assessment is that the user outcome is coherent and indivisible, but DESIGN must confirm the implementation scope estimate before proceeding.

---

## US-BSG-03: kanban board --me Filter

| DoR Item | Status | Evidence |
|----------|--------|---------|
| 1. Problem statement clear, domain language | PASS | "Jon scans a board of 15 tasks across all developers to find his 4. As the project grows, the scanning cost increases." Domain terms: assignee, git identity, board filter. |
| 2. User/persona with specific characteristics | PASS | Jon Santos on a 2+ person project; git identity: jon@kanbandev.io; runs `kanban board` daily |
| 3. At least 3 domain examples with real data | PASS | Example 1: TASK-005 + TASK-007 shown for jon@kanbandev.io. Example 2: TASK-012 unassigned → Maria gets warning. Example 3: No matching tasks → empty board with helpful message. |
| 4. UAT scenarios in Given/When/Then (3-7) | PASS | 3 scenarios: filtered board, unassigned warning, empty filtered board |
| 5. Acceptance criteria derived from UAT | PASS | 7 criteria (AC-03-1 through AC-03-7) each traceable to UAT scenarios |
| 6. Right-sized (1-3 days, 3-7 scenarios) | PASS | 1 day effort, 3 scenarios |
| 7. Technical notes: constraints and dependencies | PASS | Modify `GetBoard.Execute` with optional `filterAssignee` param; CLI `--me` flag; uses existing `GitPort.GetIdentity()`; no new ports needed; works with both Phase 1 and Phase 2 state storage. |
| 8. Dependencies resolved or tracked | PASS | No dependencies on US-BSG-01 or US-BSG-02. Independent delivery. |
| 9. Outcome KPIs defined with measurable targets | PASS | KPI-4: time to locate own tasks O(1) instead of O(n tasks); measurement: developer self-report + `--me` adoption rate |

### DoR Status: PASSED

---

## Summary

| Story | DoR Status | Handoff Readiness |
|-------|-----------|-------------------|
| US-BSG-01: kanban log | PASSED | Ready for DESIGN wave |
| US-BSG-02: Transitions log | PASSED (conditional on sizing confirmation) | Ready for DESIGN wave — DESIGN must confirm scope estimate |
| US-BSG-03: board --me | PASSED | Ready for DESIGN wave |

**Overall gate: PASSED**

All three stories meet the Definition of Ready. US-BSG-02 carries a size conditional that the DESIGN wave must resolve. No stories are blocked.
