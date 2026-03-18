# Walking Skeleton — board-state-in-git

## What is the walking skeleton

The walking skeleton for this feature is US-BSG-01: `kanban log <TASK-ID>`.

It delivers this observable user value: a developer can type `kanban log TASK-001`
and see the history of state transitions for a task — pulled from the repository's
git history — without any architectural change to the storage model.

The skeleton answers the question: "Can a developer look up what happened to a task
and when?"

## Why kanban log is the right skeleton

Three criteria make US-BSG-01 the correct first slice:

1. **Zero architectural change required.** The existing git adapter already has
   commit history; `kanban log` reads it. This means the skeleton can pass before
   US-BSG-02 (transitions.log) is implemented. The outer loop test fails for the
   right reason — the `log` command doesn't exist yet — not because of missing
   infrastructure.

2. **Surfaces the primary user pain.** The North Star problem this feature solves is
   "developers can't see why a task changed state." `kanban log` directly addresses
   that pain and is demo-able to a stakeholder in under a minute.

3. **Drives the right port.** All tests invoke the compiled binary. No internal
   packages are imported. The wiring from CLI adapter through use cases to the git
   port is exercised by a single `kanban log TASK-001` call — thin but complete.

## What passing means

The walking skeleton is "done" when these 5 tests pass:

| Test | What it validates |
|------|-------------------|
| `TestKanbanLog_ShowsHeader_WhenTaskHasHistory` | The command exists, runs, and identifies the task by ID and title |
| `TestKanbanLog_ShowsNoTransitions_WhenTaskHasNoCommits` | Graceful empty state with helpful message — no crash, no blank output |
| `TestKanbanLog_ExitsOne_WhenTaskNotFound` | Error path: unknown task ID produces exit 1 and "not found" |
| `TestKanbanLog_SuggestsKanbanBoard_WhenTaskNotFound` | Error path: unknown task leads developer to the right next command |
| `TestKanbanLog_ExitsOne_WhenNotInitialised` | Error path: uninitialised repo gives actionable "kanban init" suggestion |

These 5 tests together prove:
- A developer can ask for a task's history and get a meaningful response.
- A developer who makes a mistake gets an actionable error, not silence.
- The binary wires up the CLI adapter, use case, and git port end-to-end.

## What is deferred

The following are explicitly deferred to Milestone 1 (US-BSG-02) and Milestone 2 (US-BSG-03):

**Deferred from walking skeleton (remain as `t.Skip` in kanban_log_test.go):**

- AC-01-2: Transition field display (timestamp, author, trigger) — requires
  transitions.log to be populated, which is US-BSG-02 work.
- AC-01-3: Chronological ordering — same dependency.
- AC-01-9: Domain language formatting — rendering detail, not structural wiring.
- AC-01-10: Commit SHA display — supplementary detail.

**Deferred to Milestone 1 (US-BSG-02 — all 26 tests):**

The entire transitions.log append-only storage model: task creation without YAML
status, kanban start writing to log, commit-msg hook, board deriving status from log,
ci-done log entry, rebase safety, concurrency.

**Deferred to Milestone 2 (US-BSG-03 — all 5 tests):**

The `kanban board --me` filter: developer-scoped board view, unassigned task warnings,
integration of --me filter with log-derived status.

## Demo script for stakeholder

When the walking skeleton passes, this session demonstrates observable user value:

```
$ kanban init
$ kanban add -t "Fix OAuth login bug"
Created TASK-001

$ kanban log TASK-001
TASK-001: Fix OAuth login bug
No transitions recorded yet.

$ kanban start TASK-001
$ kanban log TASK-001
TASK-001: Fix OAuth login bug
  [history will appear here after AC-01-2 is implemented]
```

The first two `kanban log` outputs are fully stakeholder-demonstrable from the
walking skeleton. The detail in the history entries comes from Milestone 1.
