# Outcome KPIs — new-editor-mode

## Feature: new-editor-mode

### Objective

Make task creation as frictionless as task editing — developers capture a task with full metadata in a single `kanban new` invocation, without a follow-up `kanban edit` step, within 30 days of release.

---

### Outcome KPIs

| # | Who | Does What | By How Much | Baseline | Measured By | Type |
|---|-----|-----------|-------------|----------|-------------|------|
| 1 | Developers running `kanban new` interactively | Create a task with title AND at least one optional field (priority, assignee, or description) in a single command | 30% of task creation events include at least one optional field | 0% (not currently possible without a second `kanban edit` step) | Count of task files created with non-empty priority, assignee, or description fields / total task files created | Leading |
| 2 | Developers running `kanban new` interactively | Avoid the "new then immediate edit" pattern (edit within 60s of create) | Reduce same-session new→edit sequences by 50% | Establish in first 14 days post-release | Compare timestamps of task creation vs first edit per task file | Leading |

---

### Metric Hierarchy

- **North Star**: rate of task files created with at least one optional field populated (KPI-1). This is the direct signal that the editor mode is being used for richer task capture, not just as an alternative title-entry method.
- **Leading Indicators**:
  - KPI-2: reduction in same-session new→edit sequences (proxy for "had to go back to add details")
  - Editor mode invocation rate vs title-arg invocation rate (tracks adoption split)
- **Guardrail Metrics**:
  - `kanban new <title>` success rate must not degrade (zero regressions on existing path)
  - Exit code 2 rate for empty-title must remain stable (not spike, indicating UX confusion in the template)

---

### Measurement Plan

| KPI | Data Source | Collection Method | Frequency | Owner |
|-----|------------|-------------------|-----------|-------|
| KPI-1: optional field population rate | `.kanban/tasks/*.md` file content | Script: parse YAML front matter, count non-empty optional fields | Weekly | Developer / team |
| KPI-2: new→edit sequence rate | `.kanban/tasks/*.md` mtime + git log timestamps | Script: compare task creation time to first `kanban edit` time | Weekly | Developer / team |
| Guardrail: `kanban new <title>` regression | CI acceptance test suite | 100% pass rate | Every commit | CI |

Note: kanban operates as a local CLI tool with no telemetry. All measurement is done by inspecting task files and git history in the repository. No instrumentation changes are required in the binary itself.

---

### Hypothesis

We believe that opening `$EDITOR` when `kanban new` is invoked with no arguments will enable developers to capture task metadata in a single step. We will know this is true when 30% of newly created task files include at least one optional field populated, measured within 30 days of release.

---

### Baseline Establishment

Before measuring KPI-1 and KPI-2, a 14-day baseline period is needed post-release to establish the "editor mode adoption" rate. During this period, count all task creation events and classify by input path (editor mode vs. title-arg). Baseline for KPI-1 is expected to be 0% (optional fields are not accessible in a single step today without flags, and flag usage for optional fields is assumed to be low based on the feature request context).
