# Outcome KPIs: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15

---

## Feature Objective

Make a developer team's kanban board accurate by default -- not through discipline, but because
board state is driven by work events (commits, CI passes) rather than manual updates.
Outcome: developers trust the board and check it daily because it reflects reality.

---

## Outcome KPIs

| # | Who | Does What | By How Much | Baseline | Measured By | Type |
|---|---|---|---|---|---|---|
| 1 | Developer on shared team repo | Checks kanban board instead of asking a teammate what they are working on | Daily board checks per developer per week: target 5 (once per working day) | 0 (new tool) | Shell history frequency survey at 30 days post-release | Leading |
| 2 | Developer making a commit | Task status updates to in-progress without a manual command | 90% of in-progress transitions are auto (commit hook), not manual (kanban edit) | 0% auto (no tool exists) | Ratio of auto-transitions to manual status edits in .kanban/tasks/ git history | Leading |
| 3 | Developer whose CI pipeline passed | Task status updates to done without a manual command | 95% of done transitions are auto (CI step), not manual | 0% auto (no tool exists) | Ratio of auto-transitions to manual done edits in git history | Leading |
| 4 | New developer joining the team | Sets up kanban (init + install-hook + CI integration) without reading documentation beyond kanban --help | 100% of new team members complete setup in under 5 minutes | N/A | Time-on-task: team onboarding observation | Leading |
| 5 | Developer looking for their own tasks | Finds their assigned tasks on a 10+ task board in under 10 seconds | 100% success rate, under 10 seconds | N/A | Time-on-task: usability test with 10-task board | Leading |

---

## Metric Hierarchy

- North Star: percentage of total status transitions that are automatic (not from kanban edit)
  Target: 90%+ automatic by 60 days post-release
  Rationale: if developers are manually updating status, the auto-transition model has failed

- Leading Indicators:
  - Daily active users of "kanban board" (proxy: shell history)
  - Commit message task reference rate (% of commits that include a TASK-NNN reference)
  - CI step installation rate (% of teams using kanban who have CI step configured)

- Guardrail Metrics (must NOT degrade):
  - Git commit time: p95 commit time must not increase by more than 500ms after hook installation
  - CI pipeline duration: kanban CI step must add less than 5 seconds to total pipeline time
  - Hook failure rate: kanban hook errors (logged to hook.log) must be below 1% of all commits

---

## Measurement Plan

| KPI | Data Source | Collection Method | Frequency | Owner |
|---|---|---|---|---|
| Daily board checks | Developer shell history (opt-in survey) | 30-day survey at release + 60 days | Monthly | Product owner |
| Auto-transition ratio | .kanban/tasks/ git log (status field change author: hook vs human) | Script against repo history | Weekly | Developer |
| CI integration rate | .kanban/config presence of CI step config | Repo scan at 30 and 60 days | Monthly | Developer |
| Setup time | Team onboarding observation sessions | At each new team member onboarding | Per event | Developer |
| Git commit overhead | p95 commit time before and after hook install | git commit timing measurements | Before/after | Developer |

---

## Hypothesis

We believe that auto-transitioning task status on git commits and CI passes for developer teams
on shared git repositories will achieve a 90%+ automatic transition rate.
We will know this is true when developers on a team check "kanban board" daily and
fewer than 10% of status transitions are manual edits.

---

## OKR Mapping

Objective: Make team work status accurate by default, measured at 60 days post-release

Key Results:
- KR1: 90% of in-progress and done transitions are automatic (commit hook + CI step)
- KR2: Developers on using teams check "kanban board" at least once per working day (5x/week)
- KR3: New team setup (init to first CI integration) completes in under 5 minutes for 100% of teams

These are committed KRs. The walking skeleton must be shipped to begin measurement.
