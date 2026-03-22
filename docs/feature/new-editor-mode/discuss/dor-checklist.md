# Definition of Ready Checklist — new-editor-mode

## Story: US-01 — kanban new launches editor when invoked with no arguments

| DoR Item | Status | Evidence |
|----------|--------|---------|
| 1. Problem statement clear, domain language | PASS | "Alex finds it friction-heavy to compose title, description, and priority on one command line mid-coding session. Running `kanban new` with no arguments silently fails instead of opening the editor." Domain terms: task, editor, git repository, kanban init. |
| 2. User/persona with specific characteristics | PASS | Alex Chen — developer using kanban as personal git-native task tracker, working interactively at a TTY, familiar with $EDITOR-based workflows (git commit, kanban edit). Three named personas used in domain examples: Alex, Priya, Jordan. |
| 3. At least 3 domain examples with real data | PASS | Example 1: Alex, TASK-042, "Fix nil pointer in auth handler", priority P1, description present. Example 2: Priya, TASK-007, "Update README examples", all optional fields blank. Example 3: Jordan, quits without title, exit 2, no task created. |
| 4. UAT scenarios in Given/When/Then (3-7 scenarios) | PASS | 7 scenarios in user-stories.md covering: editor opens, full field creation, title-only creation, empty title rejection, existing path unchanged, editor unavailable, pre-flight failure. |
| 5. Acceptance criteria derived from UAT | PASS | 9 AC items in acceptance-criteria.md, each citing the UAT scenario it derives from and naming the driving port (kanban new subcommand). |
| 6. Story right-sized (1-3 days, 3-7 scenarios) | PASS | 1 story, 7 scenarios, estimated 1-2 days effort. Brownfield; reuses existing EditFilePort, openEditor(), AddTask.Execute. Only the CLI adapter branch is new. |
| 7. Technical notes: constraints and dependencies | PASS | Technical Notes in user-stories.md: openEditor() sharing strategy, WriteTemp zero-value task, due field exclusion from template, EditFilePort dependency, exit code routing for empty title. Constraints documented in requirements.md (C-01 through C-05). |
| 8. Dependencies resolved or tracked | PASS | Dependencies: EditFilePort interface (exists, unchanged), AddTask use case (exists, unchanged), openEditor() function (exists in usecases — sharing strategy flagged as design decision for solution-architect, risk documented in prioritization.md). |
| 9. Outcome KPIs defined with measurable targets | PASS | outcome-kpis.md: KPI-1 (30% of task files include optional field within 30 days), KPI-2 (50% reduction in new→edit sequences). Baseline, measurement method, and hypothesis documented. |

---

### DoR Status: PASSED

All 9 items pass. Story is ready for handoff to the DESIGN wave (solution-architect).

---

## Anti-Pattern Check

| Anti-Pattern | Check | Result |
|---|---|---|
| Implement-X | Story title and problem statement | PASS — "kanban new launches editor when invoked with no arguments" starts from user pain, not technical task |
| Generic data | Domain examples | PASS — Alex Chen, Priya, Jordan; TASK-042, TASK-007; real task titles used |
| Technical AC | Acceptance criteria | PASS — all AC describe observable user outcomes (editor opens, file created, stdout content, exit codes). No AC prescribes JWT, database, or implementation technology |
| Oversized story | 7 scenarios, 1-2 days | PASS — at the upper boundary (7 scenarios) but all are essential; no scenario is decorative |
| Abstract requirements | Domain examples and AC | PASS — 3 concrete examples with named personas and specific field values |

---

## Scope Assessment: PASS

1 story, 2 bounded contexts touched (cli adapter, usecases), estimated 1-2 days, 0 new integration points (all reuse existing ports and use cases).
