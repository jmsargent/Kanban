# Walking Skeleton: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DISTILL
**Date**: 2026-03-15

---

## Purpose

The walking skeleton is the thinnest vertical slice that delivers observable user value end-to-end. It is the first scenario the software-crafter enables and implements. Every other scenario builds on it.

The test for the skeleton is demo-able to stakeholders: run the commands, point at the terminal output. No explanation of implementation required.

---

## Skeleton Scope

**Scenario**: Developer completes the full task lifecycle from creation to done
**File**: `tests/acceptance/kanban-tasks/walking-skeleton.feature`
**Tag**: `@walking_skeleton`

**User journey covered**:

```
kanban init → kanban add → kanban board → git commit (hook) → CI pass → kanban board
```

This covers all five backbone activities from the story map:
- Create: `kanban add`
- View: `kanban board`
- Work: commit hook moves task to in-progress
- Complete: CI step moves task to done
- (Maintain: covered by the third walking skeleton scenario)

---

## Litmus Test

The walking skeleton passes all four litmus tests from the test-design-mandates skill:

1. **Title describes user goal**: "Developer completes the full task lifecycle from creation to done" — not "end-to-end flow through all layers."
2. **Given/When describe user actions**: "I run kanban init", "I commit with message containing the task ID" — not "database contains record."
3. **Then describes user observations**: "the board shows the task under DONE" — not "task file status field updated."
4. **Non-technical stakeholder confirmation**: Yes. A product owner can read the scenario and confirm "that is what developers need."

---

## Three Walking Skeletons

### WS-1: Full lifecycle (init → add → board → commit → CI → board)
The primary skeleton. Proves the complete developer loop end-to-end. All three driving ports exercised: CLIAdapter, GitHookAdapter, CIPipelineAdapter.

### WS-2: Init, add, view
The minimal "can a developer capture and see a task" slice. Proves CLIAdapter alone. Enables smoke-testing the binary before hook and CI infrastructure exist.

### WS-3: Edit and delete (maintenance)
Proves the maintenance loop. A developer can fix a mistake and remove cancelled work. Simpler than WS-1 but covers the Maintain backbone activity.

---

## Implementation Sequence

The software-crafter implements in this order:

1. Enable WS-2 (simplest — CLIAdapter only, no hooks or CI)
   - `kanban init` compiles and runs
   - `kanban add "title"` creates TASK-001.md
   - `kanban board` reads and displays it
2. Enable WS-1 (adds GitHookAdapter and CIPipelineAdapter)
3. Enable WS-3 (adds edit and delete paths)
4. Then enable focused scenarios one at a time, milestone by milestone

---

## Driving Ports Used

| Scenario | Driving Port | Invocation |
|----------|-------------|------------|
| WS-1 (full lifecycle) | CLIAdapter + GitHookAdapter + CIPipelineAdapter | `kanban <cmd>`, `kanban _hook commit-msg`, `kanban ci-done` |
| WS-2 (init/add/view) | CLIAdapter | `kanban init`, `kanban add`, `kanban board` |
| WS-3 (edit/delete) | CLIAdapter | `kanban edit`, `kanban delete` |

All three driving ports from the architecture design are represented.

---

## Observable Outcomes

At completion of WS-1, a stakeholder can:
- Watch a developer run four terminal commands
- See a task move from TODO to IN PROGRESS to DONE without any manual status update
- Confirm: "yes, the board tracks real work automatically"

This is the definition of a user-centric walking skeleton.
