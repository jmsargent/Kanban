# Observability Design — kanban-tasks

**Wave**: DEVOPS
**Date**: 2026-03-16
**Feature**: kanban-tasks (Git-Native Kanban Task Management CLI)

---

## Decision: Observability Deferred

kanban-tasks is a developer CLI tool distributed as a static binary. It has no server-side runtime, no API endpoints, and no persistent service to monitor. Standard observability infrastructure (metrics servers, distributed tracing, alerting pipelines, dashboards) is not applicable and has been deferred indefinitely.

This is not a gap — it is the correct design for the deployment model.

---

## What Is Observable (Without Infrastructure)

### 1. Local Hook Error Log: `.kanban/hook.log`

The commit-msg hook (installed by `kanban install-hook`) writes error events to `.kanban/hook.log` within the user's repository. This is the only runtime observability artifact.

Format: append-only, one error per line, timestamp + error message.

This log enables:
- Debugging hook failures on a specific machine
- Measuring hook failure rate (see guardrail KPI below)

The hook always exits 0 (per ADR-004 and non-negotiable constraint #1), so failures are silent to git but logged here.

### 2. Git History as Observability Data

The `.kanban/tasks/` directory is tracked in git. Every status transition is a commit. This makes the git log the primary data source for all outcome KPI measurement. No telemetry infrastructure needed.

Queries against git history can answer:
- Auto-transition ratio (commit hook vs. manual `kanban edit`)
- CI step adoption rate (presence of `kanban ci-done` in CI config files)
- Board activity patterns (frequency of `kanban board` commands, inferred from git history timing)

---

## Guardrail KPI Measurement (from outcome-kpis.md)

These are the three guardrail metrics that must not degrade. Each can be measured without infrastructure.

### Guardrail 1: Git Commit Time (p95 < +500ms after hook installation)

**What**: p95 commit wall-clock time must not increase by more than 500ms after installing the kanban commit-msg hook.

**How to measure** (no infrastructure required):

```sh
# Before hook installation — baseline
for i in $(seq 1 20); do
  time git commit --allow-empty -m "timing test $i"
done 2>&1 | grep real

# After hook installation — delta
for i in $(seq 1 20); do
  time git commit --allow-empty -m "timing test $i [kanban: TASK-001]"
done 2>&1 | grep real
```

Take p95 of each set. Delta must be < 500ms.

**Cadence**: measured once at initial rollout per team. Re-measured if hook implementation changes.

### Guardrail 2: CI Step Duration (< 5 seconds added)

**What**: `kanban ci-done` must add less than 5 seconds to the total CI pipeline time.

**How to measure**: CircleCI step-level timing is reported in the pipeline UI. Compare total pipeline time with and without the `kanban ci-done` step.

**Cadence**: validated once at CI integration setup. Monitored via CircleCI timing data in the pipeline dashboard (no external tooling needed).

### Guardrail 3: Hook Failure Rate (< 1% of commits)

**What**: kanban hook errors (written to `.kanban/hook.log`) must be below 1% of all commits.

**How to measure**:

```sh
# Count hook error events in the log
error_count=$(wc -l < .kanban/hook.log)

# Count total commits to main since hook installation
total_commits=$(git log --oneline --since="<install-date>" | wc -l)

# Failure rate
echo "scale=4; $error_count / $total_commits * 100" | bc
```

**Cadence**: weekly review during the first 60 days post-release, then monthly.

---

## Outcome KPI Measurement Scripts (from outcome-kpis.md)

These are measurement methods for the 5 outcome KPIs, all git-history-based.

### KPI 2 + 3: Auto-Transition Ratio

```sh
# In a team's repo, after 30 days of use:
# Count transitions where author = hook (commit message contains [hook] marker)
git log --all --oneline -- .kanban/tasks/ | grep -c "auto-transition"

# Count total status transitions
git log --all --oneline -- .kanban/tasks/ | wc -l
```

The exact markers depend on how the commit-msg hook formats its commits. The hook's commit message format should include a machine-readable marker (e.g., `[kanban-hook]` or `[kanban-ci]`) to distinguish auto-transitions from manual edits. This is a design constraint for the hook implementation.

### KPI 4: Setup Time

Measured via team onboarding observation. No tooling. Stopwatch + observer.
Target: < 5 minutes from `kanban init` to first CI integration.

### KPI 5: Board Query Time

Measured via usability test with a pre-seeded 10-task board.
Target: developer finds their assigned tasks in < 10 seconds.

---

## Summary: No Monitoring Infrastructure Required

| Concern | Decision |
|---------|----------|
| Metrics server | Not needed — no service to instrument |
| Distributed tracing | Not applicable — single binary, no network calls |
| Alerting pipeline | Not needed — no SLOs to alert on |
| Dashboards | Not needed — git history + CircleCI UI provide sufficient visibility |
| Log aggregation | Not needed — `.kanban/hook.log` is local and sufficient |
| Error tracking (Sentry, etc.) | Deferred — revisit if a server-side component is added |
