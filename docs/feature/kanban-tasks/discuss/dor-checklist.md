# Definition of Ready Checklist: kanban-tasks

Feature: Git-Native Kanban Task Management CLI
Date: 2026-03-15
Validator: Luna (nw-product-owner)

---

## DoR Validation Results

### US-01: Repository Initialisation

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "Rafael finds it frustrating to manually create directories and config files before he can do anything useful" |
| User/persona with specific characteristics | PASS | "Backend developer, in an existing git repository, wants to start tracking work" |
| 3+ domain examples with real data | PASS | Three examples with Rafael, Priya, and Sam; real commands and file paths |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 3 scenarios covering happy path, idempotent, and error |
| AC derived from UAT | PASS | 5 acceptance criteria, each traceable to a UAT scenario |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 3 UAT scenarios; estimated 0.5-1 day |
| Technical notes: constraints/dependencies | PASS | git rev-parse dependency noted; .kanban/ committed to repo; no overwrite on re-init |
| Dependencies resolved or tracked | PASS | No external dependencies; depends only on git binary presence |
| Outcome KPIs defined | PASS | Setup time KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-02: Create Task

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "wants to add it without leaving the terminal ... with the confidence that the task will be tracked in the same history as the code fix" |
| User/persona with specific characteristics | PASS | "Backend developer, in terminal mid-work session, wants to capture work without context switch" |
| 3+ domain examples with real data | PASS | Three examples: Rafael minimal, Rafael full metadata, Rafael validation error; real commands and outputs |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 5 scenarios covering happy path, full fields, ID increment, empty title, past due date |
| AC derived from UAT | PASS | 9 acceptance criteria (AC-02-1 through AC-02-9) each traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 5 UAT scenarios; estimated 1-2 days |
| Technical notes: constraints/dependencies | PASS | Atomic ID generation noted; file format deferred to DESIGN; kanban init dependency noted |
| Dependencies resolved or tracked | PASS | Depends on US-01 (kanban init); tracked in story map |
| Outcome KPIs defined | PASS | Task capture time KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-03: View Board

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "She does not want to open a browser, log into Jira, or send a Slack message" -- concrete pain |
| User/persona with specific characteristics | PASS | Priya Nair, frontend developer, start of day, wants current team status |
| 3+ domain examples with real data | PASS | Three examples: Rafael board before standup, Priya --json with jq, Sam empty board |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 5 scenarios covering grouped display, overdue indicator, empty state, JSON output, NO_COLOR |
| AC derived from UAT | PASS | 10 acceptance criteria (AC-03-1 through AC-03-10) traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 5 UAT scenarios; estimated 1-2 days |
| Technical notes: constraints/dependencies | PASS | 100ms performance requirement noted; non-TTY handling documented; UTC date comparison noted |
| Dependencies resolved or tracked | PASS | Depends on US-02 (tasks must exist); tracked in story map |
| Outcome KPIs defined | PASS | Daily board checks KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-04: Auto-move to In Progress (Commit Hook)

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "He wants the board to show the task as in-progress without making him type a separate command. Any extra step after 'git commit' means he might forget" |
| User/persona with specific characteristics | PASS | "Developer making first commit on a task, in terminal coding session" |
| 3+ domain examples with real data | PASS | Three examples: Rafael first commit auto-transition, unknown task ID warning, clean commit with no reference |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 4 scenarios covering transition, no-op, unknown ID, no reference |
| AC derived from UAT | PASS | 8 acceptance criteria (AC-04-1 through AC-04-8) traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 4 UAT scenarios; estimated 1-2 days |
| Technical notes: constraints/dependencies | PASS | 500ms performance requirement; hook-never-blocks-git rule; ci_task_pattern from config; hook.log fallback |
| Dependencies resolved or tracked | PASS | Depends on US-02 (task files must exist); depends on kanban install-hook (R3, tracked as release dependency) |
| Outcome KPIs defined | PASS | 90% auto-transition rate KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-05: Auto-move to Done (CI Pipeline)

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "The passing tests are the definition of done, so no human should have to move the card. A manual step here is exactly the kind of thing that makes boards go stale" |
| User/persona with specific characteristics | PASS | "Alex Kim, developer whose CI pipeline just passed, in CI/CD environment" |
| 3+ domain examples with real data | PASS | Three examples: Alex single task, Rafael+Alex multiple tasks, pipeline fail no transition |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 4 scenarios: pass transition, fail no-op, multiple tasks, already-done no-op |
| AC derived from UAT | PASS | 9 acceptance criteria (AC-05-1 through AC-05-9) traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 4 UAT scenarios; estimated 1-2 days for script + 1 day for CI platform wrappers |
| Technical notes: constraints/dependencies | PASS | Non-TTY requirement; git commit-back noted; ci_task_pattern from config; platform targets noted |
| Dependencies resolved or tracked | PASS | Depends on US-04 (same config and pattern logic); CI platform integration is R3 |
| Outcome KPIs defined | PASS | 95% auto-done-transition KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-06: Edit Task

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "wants to add it without creating a new task or deleting and recreating" -- concrete workaround described |
| User/persona with specific characteristics | PASS | "Developer with an existing task needing updated metadata, wants to correct without ceremony" |
| 3+ domain examples with real data | PASS | Three examples: Rafael adds description, Rafael fixes typo, Rafael wrong task ID |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 4 scenarios: add description, show current values, edit title, non-existent task |
| AC derived from UAT | PASS | 6 acceptance criteria (AC-06-1 through AC-06-6) traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 4 UAT scenarios; estimated 0.5-1 day |
| Technical notes: constraints/dependencies | PASS | $EDITOR variable documented; file mtime check to detect no-save noted; status override as escape hatch |
| Dependencies resolved or tracked | PASS | Depends on US-02 (task files must exist) |
| Outcome KPIs defined | PASS | 100% in-terminal edit completion KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

### US-07: Delete Task

| DoR Item | Status | Evidence |
|---|---|---|
| Problem statement clear, domain language | PASS | "A task was added for a feature that was descoped mid-sprint. He needs to remove it cleanly, with the removal tracked in git history" |
| User/persona with specific characteristics | PASS | "Developer with a cancelled task, wants clean removal without accidental data loss" |
| 3+ domain examples with real data | PASS | Three examples: Rafael deletes cancelled task, Priya aborts on wrong title, CI script --force |
| UAT in Given/When/Then, 3-7 scenarios | PASS | 4 scenarios: confirm delete, abort, --force, non-existent task |
| AC derived from UAT | PASS | 8 acceptance criteria (AC-07-1 through AC-07-8) traceable to UAT |
| Right-sized: 1-3 days, 3-7 scenarios | PASS | 4 UAT scenarios; estimated 0.5-1 day |
| Technical notes: constraints/dependencies | PASS | Immediate file read before confirmation; default "N"; no auto-commit; developer owns git history |
| Dependencies resolved or tracked | PASS | Depends on US-02 (task files must exist) |
| Outcome KPIs defined | PASS | Zero accidental deletions KPI defined in outcome-kpis.md |

### DoR Status: PASSED

---

## Summary

| Story | DoR Status | Estimated Effort |
|---|---|---|
| US-01: Repository Initialisation | PASSED | 0.5-1 day |
| US-02: Create Task | PASSED | 1-2 days |
| US-03: View Board | PASSED | 1-2 days |
| US-04: Auto-move to In Progress | PASSED | 1-2 days |
| US-05: Auto-move to Done (CI) | PASSED | 2-3 days |
| US-06: Edit Task | PASSED | 0.5-1 day |
| US-07: Delete Task | PASSED | 0.5-1 day |

All 7 stories pass DoR. Total walking skeleton estimated at 6-12 days.
Ready for handoff to DESIGN wave (solution-architect).
