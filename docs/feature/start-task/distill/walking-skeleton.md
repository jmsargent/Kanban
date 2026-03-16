# Walking Skeleton — kanban start

## Identified Skeleton

**Scenario**: "Developer starts a todo task and sees it move to in-progress"
(milestone-3-start-command.feature, tagged `@walking_skeleton`)

## Litmus Test

1. Title describes a user goal — "Developer starts a task and sees it move" — not a technical flow. Pass.
2. Given/When describe user context (initialised repo, existing todo task) and a user action (running a command). Pass.
3. Then describes what the user observes: the task status on disk has changed to in-progress, the command reports success with the task ID. Pass.
4. A non-technical stakeholder can confirm: "Yes, a developer should be able to start a task and see it reflected." Pass.

## What It Proves

A developer can transition a task from the backlog to active work in a single command. The board state updates atomically, and the command confirms the change with a human-readable message. This is the simplest complete user journey for the start feature.

## Implementation Sequence

This scenario is first in the feature file and carries no `@skip` tag. All five scenarios are enabled because the feature is fully implemented. The implementation sequence (one at a time) is:

1. Scenario 1 (walking skeleton) — establishes the core path
2. Scenario 2 (already in-progress idempotence)
3. Scenario 3 (done task rejection)
4. Scenario 4 (task not found)
5. Scenario 5 (not initialised)
