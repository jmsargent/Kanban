# Peer Review — Board State in Git DISCUSS Wave
**Review ID**: req_rev_20260318_001
**Reviewer**: Luna (product-owner, review mode)
**Artifacts**: All 15 DISCUSS wave artifacts
**Iteration**: 1
**Date**: 2026-03-18

---

## Review Output

```yaml
review_id: "req_rev_20260318_001"
reviewer: "product-owner (review mode)"
artifact: "docs/feature/board-state-in-git/discuss/"
iteration: 1

strengths:
  - "Phased delivery well-justified: walking skeleton (US-BSG-01) separates the quick win from architectural investment. DISCOVER wave recommendation is honored."
  - "All three stories have concrete domain examples with real persona (Jon Santos), real email (jon@kanbandev.io), real task IDs (TASK-007, TASK-003) — no generic data."
  - "Shared artifacts registry covers all ${variables} in TUI mockups with documented single sources of truth."
  - "CLAUDE.md non-negotiables explicitly addressed in every story's technical notes — hook exit-0 guarantee, atomic writes, no auto-commit."
  - "Elephant Carpaccio gate applied: US-BSG-02 at upper limit flagged with split options documented."
  - "Error paths documented in journey visual: task not found, no transitions, uninitialized repo."
  - "Outcome KPIs follow the Who/Does what/By how much template and include baselines."
  - "Decision gate after US-BSG-01 explicitly defined with measurable criteria — avoids premature architectural commitment."

issues_identified:
  confirmation_bias:
    - issue: "Happy path bias: US-BSG-02 scenarios cover the happy path and rebase safety, but do not cover what happens when a developer deletes a task file after transitions have been recorded for it. The log has entries for a task that no longer exists."
      severity: "high"
      location: "US-BSG-02 UAT scenarios, AC-02"
      recommendation: "Add scenario: kanban board when transitions.log references a task file that has been deleted. Expected behavior: task is excluded from board with optional warning. Add AC-02-25."

    - issue: "Missing error scenario for US-BSG-01: what if the git repo has no commits at all (empty repo, just initialized)? git log --follow returns nothing."
      severity: "medium"
      location: "US-BSG-01 UAT scenarios"
      recommendation: "Edge case is covered implicitly by the 'no transitions recorded' scenario (AC-01-4), but should be explicit. The empty-repo case may also affect kanban init sequencing. Low priority — add as a note in AC-01-4."

  completeness_gaps:
    - issue: "Non-functional requirement missing: what is the maximum size of .kanban/transitions.log before performance degrades? The requirements mention 1,000 transitions as acceptable but do not define an upper threshold or indexing strategy."
      severity: "high"
      location: "requirements.md — Non-Functional Requirements, Performance"
      recommendation: "Add explicit NFR: transitions.log performance is acceptable for projects with up to 10,000 transitions (O(n) file read remains fast). If log exceeds 10,000 entries, recommend DESIGN wave investigation of binary search or index file. Add to requirements.md."

    - issue: "Migration story is noted as a dependency for US-BSG-02 on existing repos, but no acceptance criteria or user story exist for it. If this story is needed before US-BSG-02 ships, its absence is a dependency gap."
      severity: "medium"
      location: "prioritization.md — Release 2: Migration"
      recommendation: "Document the migration story as a known future story (not DoR-ready yet). Add a stub in user-stories.md as 'US-BSG-04: migrate existing repo to transitions log (Draft — to be crafted in DESIGN wave)'. This makes the dependency explicit without blocking handoff."

    - issue: "US-BSG-02 does not define behavior when kanban ci-done has no in-progress tasks in the commit range. Is it a no-op (exit 0, output 'no tasks to transition') or an error?"
      severity: "high"
      location: "US-BSG-02 UAT scenarios"
      recommendation: "Add scenario: kanban ci-done with no in-progress tasks referenced in commit range. Expected: exit 0, output 'no tasks transitioned'. Add AC-02-26."

  clarity_issues:
    - issue: "RQ-1 (transitions.log line format) is left entirely open for the DESIGN wave, but the Gherkin scenarios in journey-task-state-transitions.feature use a specific format in the Background section. If the DESIGN wave chooses a different format, the Gherkin scenarios become misleading."
      severity: "medium"
      location: "journey-task-state-transitions.feature — Gherkin docstring in Scenario: kanban board derives status"
      recommendation: "Add a note to the Gherkin file: 'NOTE: exact log line format is subject to RQ-1 resolution in DESIGN wave. This Gherkin uses an illustrative format.' The requirement is the FIELDS present, not the exact delimiter."

    - issue: "The phrase 'Concept B enabled' in the Gherkin Background is not a defined concept in the codebase. It is used shorthand in scenarios but will confuse the acceptance-designer."
      severity: "low"
      location: "journey-task-state-transitions.feature — @phase:2 scenarios"
      recommendation: "Replace 'Concept B enabled' with 'the kanban repository uses transitions log for state storage' or 'kanban is initialized with transitions log enabled'. The feature flag mechanism is a DESIGN wave decision."

  testability_concerns:
    - issue: "AC-01-11 (performance: 500ms on 1,000+ commits) uses wall-clock time, which is environment-dependent and flaky in CI. CI machines with slow I/O may fail this test inconsistently."
      severity: "high"
      location: "acceptance-criteria.md — AC-01-11"
      recommendation: "Rewrite as a benchmark test with tolerance: 'In a benchmark of 5 runs, the 95th percentile response time is under 500ms on standard developer hardware.' Flag as a benchmark test (@benchmark tag in Gherkin), not an acceptance test. CI can skip benchmark tests; developers run them locally."

    - issue: "AC-02-23 (concurrent writes) uses the phrase 'two simultaneous appends' which is hard to guarantee in a test. True simultaneity requires goroutine synchronization."
      severity: "medium"
      location: "acceptance-criteria.md — AC-02-23"
      recommendation: "This is best expressed as a property-based test: for any sequence of N concurrent appends, the resulting file has exactly N new lines. The acceptance-designer should implement this as a goroutine-based race test with sync.WaitGroup. Note this in AC-02-23."

  priority_validation:
    q1_largest_bottleneck: "YES — audit trail (OS-1: 16) is the primary pain confirmed by DISCOVER; walking skeleton targets it directly"
    q2_simple_alternatives: "ADEQUATE — Concept A (commit trailers) was evaluated and rejected with documented reasoning; current system extension was considered and validated as Phase 1"
    q3_constraint_prioritization: "CORRECT — CLAUDE.md non-negotiables (hook exit-0, no auto-commit) are treated as hard constraints, not tradeoffs"
    q4_data_justified: "JUSTIFIED — opportunity scores derived from DISCOVER wave interview; CONDITIONAL because N=1 participant"
    verdict: "PASS"

approval_status: "conditionally_approved"
critical_issues_count: 0
high_issues_count: 3
medium_issues_count: 3
low_issues_count: 1

condition_for_full_approval: "Address the 3 high-severity issues before DESIGN wave handoff. Medium and low issues may be addressed in DESIGN wave."
```

---

## Remediation Actions

The following high-severity issues require remediation before handoff is complete:

### HIGH-1: Missing scenario — deleted task file with log entries

**Action**: Add to US-BSG-02 acceptance criteria:
> AC-02-25: When `kanban board` runs and `.kanban/transitions.log` references a task ID for which no task file exists (task was deleted), the deleted task is excluded from the board display and does not cause an error. Exit code 0.

### HIGH-2: Missing scenario — `kanban ci-done` with no in-progress tasks

**Action**: Add to US-BSG-02 acceptance criteria:
> AC-02-26: When `kanban ci-done` runs and the commit range contains no references to in-progress tasks, the command exits 0 and outputs "no tasks to transition." No entries are added to the transitions log.

### HIGH-3: Performance criterion (AC-01-11) is flaky in CI

**Action**: Reclassify AC-01-11 as a benchmark criterion:
> AC-01-11 (revised): In a local benchmark of 5 runs on standard developer hardware, the 95th percentile response time for `kanban log` on a repo with 1,000+ commits is under 500ms. This is a benchmark criterion tagged @benchmark — it is NOT executed in CI by default.
