# Walking Skeleton — new-editor-mode

**Wave**: DISTILL
**Date**: 2026-03-22

---

## Scenario

`TestNewEditorMode_WalkingSkeleton_TaskCreated`

**User goal**: A developer wants to create a new task without pre-knowing its title — they invoke `kanban new` with no arguments to get an editor where they can compose the task interactively.

**Covers**: AC-01 (invocation routes to editor) and AC-03 (task created and confirmed).

---

## Prose Description

Given a developer is working in a git repository with kanban initialised, when they run `kanban new` with no arguments and an editor opens presenting a blank task template, and they fill in a title ("Implement user authentication") and save, then the command exits 0, prints a creation confirmation line ("Created TASK-NNN: Implement user authentication"), prints a hint to reference the task ID in the next commit, and the task file is present in `.kanban/tasks/`.

---

## Stakeholder Litmus Test

"A developer ran `kanban new` with no arguments, typed a title in their editor, saved, and the tool confirmed the task was created and told them the ID to reference in their commit."

A non-technical stakeholder can confirm: yes, that is what developers need.

---

## Why This Is the Right Walking Skeleton

The feature's single user story (US-01) has one core journey: no-argument invocation -> editor opens -> title filled -> task created -> confirmation. This skeleton traces that complete journey end to end, touching every new component (CLI routing branch, WriteTempNew, EDITOR env var launch, ReadTemp, AddTask, success output) as a consequence of the user journey — not as a design goal.

It is demo-able: run the test, read the stdout, show the task file. Done.

---

## What the Skeleton Does NOT Validate

The walking skeleton deliberately excludes:

- Template structure details (AC-02) — covered by focused scenario 2
- Optional fields (AC-04) — covered by focused scenario 3
- Error paths (AC-05, AC-07, AC-08) — covered by focused scenarios 4, 6, 7
- Temp file cleanup (AC-09) — covered by focused scenarios 8 and 9

These are left to focused scenarios so the skeleton stays minimal and fast to make green.
