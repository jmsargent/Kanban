# Outcome KPIs — Board State in Git
**Feature**: board-state-in-git
**Wave**: DISCUSS
**Date**: 2026-03-18
**Author**: Luna (product-owner)

---

## Feature: board-state-in-git

### Objective

By the end of the DISCUSS wave deliveries, a developer using kanban-tasks can reconstruct the full history of any task in domain language from a single command, and task state transitions no longer pollute task definition files or commit diffs.

---

## Outcome KPIs

| # | Who | Does What | By How Much | Baseline | Measured By | Type |
|---|-----|-----------|-------------|----------|-------------|------|
| KPI-1 | Developer using kanban-tasks | Views task transition history in domain language with one command | 100% of history queries completed in ≤1 command (from 0% — no command exists) | 0 — no `kanban log` command exists | Developer self-report; observation of git plumbing command usage in shell history | Leading — Outcome |
| KPI-2 | Developer committing code that references a task | Makes commits that produce zero task file mutations from state side effects | Files modified by state transition: 1 (transitions.log), down from 1 per transitioned task file | 1 task file modified per hook-fired commit | `git show <commit>` — count of `.kanban/tasks/*.md` in commit diff | Leading — Outcome |
| KPI-3 | Developer who runs `git rebase -i` | Has task state preserved after rebase | 100% of transitions survive rebase (up from "not guaranteed" — risk exists with commit trailer approach) | Partial — current system does not use commit trailers, so rebase is technically safe, but this is not enforced | Integration test: rebase + assert board state matches pre-rebase | Leading — Secondary |
| KPI-4 | Developer on 2+ person project | Finds their own tasks without scanning the full board | Time to locate own tasks: O(1) instead of O(n tasks) | Full board scan required — O(n tasks) | Developer self-report; `--me` flag adoption rate | Leading — Secondary |

---

## Metric Hierarchy

### North Star
**KPI-1**: Developers view task transition history in domain language from one command.

This is the most direct measure of the feature's primary value proposition. It addresses the highest-scoring opportunity (OS-1: 16) and was the developer's explicitly stated primary motivator ("The audit trail feels natural with commits").

### Leading Indicators (predicting north star)

- **`kanban log` usage rate**: % of kanban-tasks users who run `kanban log` within 7 days of it shipping
- **Reduction in `git log -- .kanban/tasks/` invocations**: proxy for users switching from manual git to the new command (hard to measure directly; developer self-report)

### Guardrail Metrics (must NOT degrade)

| Metric | Threshold | Rationale |
|--------|-----------|-----------|
| `kanban board` display accuracy | 100% — board column must match actual task status at all times | Core product reliability |
| Hook exit code | 100% exits 0 — zero commit-blocking incidents | CLAUDE.md non-negotiable |
| `kanban board` response time | <500ms on projects with 50+ tasks | Current performance baseline must be maintained or improved |
| Commit diff correctness | Zero task files appearing in a state-transition-only commit (Phase 2) | Regression would invalidate KPI-2 |

---

## Measurement Plan

| KPI | Data Source | Collection Method | Frequency | Owner |
|-----|------------|-------------------|-----------|-------|
| KPI-1 | Developer self-report | 2-4 week check-in after US-BSG-01 ships; question: "Does `kanban log` satisfy your audit trail need?" | Once (2-4 weeks post-ship) | Product owner |
| KPI-2 | Git history | `git show <commit>` on hook-fired commits; count task file mutations | On each tagged release; spot check during development | Developer (automated assertion in acceptance test) |
| KPI-3 | Integration test | Automated: create task, make transitions, rebase, assert state | CI run on every commit | CI pipeline |
| KPI-4 | Developer self-report | After US-BSG-03 ships; question: "Does `--me` reduce the time you spend scanning the board?" | Once (2-4 weeks post-ship) | Product owner |

---

## Hypothesis

We believe that adding `kanban log <TASK-ID>` for Jon Santos and other developers using kanban-tasks will achieve KPI-1: developers view transition history with one command (up from 0%).

We will know this is true when Jon reports that `kanban log TASK-007` satisfies his audit trail need without requiring additional `git log` commands.

We believe that implementing Concept B (append-only transitions log) for Jon Santos will achieve KPI-2: zero task file mutations per hook-fired commit.

We will know this is true when `git show` on any hook-fired commit shows only `.kanban/transitions.log` in the diff, with no `.kanban/tasks/*.md` modifications.

---

## KPI Decision Gate

After US-BSG-01 ships, evaluate:

| Question | If YES | If NO |
|----------|--------|-------|
| Is KPI-1 met? Does `kanban log` fully satisfy the audit trail need? | Re-evaluate necessity of US-BSG-02 — the primary pain may be resolved | Proceed with US-BSG-02 — the transitions log is needed |
| Does KPI-2 remain important? Do hook-writes-files and diff-noise still bother the developer? | Proceed with US-BSG-02 | Deprioritize US-BSG-02 — secondary pains are tolerable |

This decision gate is the explicit validation checkpoint recommended by the DISCOVER wave before committing to the 8-10 day US-BSG-02 investment.

---

## North Star KPI for the Full Feature (DISCOVER alignment)

The DISCOVER wave identified a North Star KPI:
> "Automatic transitions feel invisible and the board always reflects current reality."

The stories in this wave contribute to it as follows:

| Story | Contribution to North Star |
|-------|--------------------------|
| US-BSG-01 | Surfacing history makes the past transitions visible and trustworthy |
| US-BSG-02 | Making current transitions invisible (no file mutations) and reliable (rebase-safe) |
| US-BSG-03 | Scoping "current reality" to the developer's own perspective |
