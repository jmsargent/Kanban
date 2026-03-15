# ADR-004: Git Hook Strategy — commit-msg Hook, Exit 0 Guarantee

**Status**: Accepted
**Date**: 2026-03-15
**Feature**: kanban-tasks
**Resolves**: OD-05

---

## Context

The auto-transition from "todo" to "in-progress" must fire when a developer makes a git commit referencing a task ID. This requires a git hook. Three concerns shape the design:

1. **Hook type**: which git hook fires at the right moment with access to the commit message
2. **Delivery mechanism**: how the hook file reaches developer machines (`.git/hooks/` is local-only and not committed)
3. **Safety guarantee**: the hook must never block a git commit under any circumstance (D-05 from DISCUSS wave; BR-6)

---

## Decision

**Hook type**: `commit-msg` hook. Receives the path to the commit message file as `$1`. Reads the message, extracts task IDs matching `ci_task_pattern`, updates matching task files.

**Delivery mechanism**: The hook is installed to `.git/hooks/commit-msg` by `kanban install-hook`. The source hook script lives at `.kanban/hooks/commit-msg` in the repository (version-controlled, shared with team). `kanban install-hook` creates a `.git/hooks/commit-msg` that delegates to the kanban binary: `exec kanban _hook commit-msg "$@"`. This means the hook logic lives in the Go binary, not in a fragile shell script.

**Exit 0 guarantee**: The hook entry point in the Go binary wraps all execution in a top-level recover. On any panic or error, it logs to `.kanban/hook.log` and exits 0. The hook never propagates a non-zero exit to git.

---

## Alternatives Considered

### Alternative 1: post-commit Hook

Fires after the commit is written. Does not receive the commit message as an argument; must read it from `git log -1 --format=%s`.

Rejection rationale: `post-commit` fires after the commit is already recorded. The transition output cannot appear inline in the `git commit` terminal session alongside the standard git output -- it appears as a separate shell invocation after git finishes. This breaks the AC-04-2 requirement that "commit output includes `kanban: TASK-NNN moved todo -> in-progress`". The `commit-msg` hook fires before the commit is finalised and its stdout appears inline.

### Alternative 2: Shell Script Committed to `.kanban/hooks/`

Deliver the hook as a pure shell script that is version-controlled in `.kanban/hooks/commit-msg` and symlinked (or copied) to `.git/hooks/commit-msg` by `kanban install-hook`.

Rejection rationale: A shell script hook must duplicate the `ci_task_pattern` parsing logic and YAML status update logic that already lives in the Go binary. Two implementations of the same logic diverge (R-04 from DISCUSS wave). Additionally, shell scripts are harder to test and have no type safety. The delegation pattern (`exec kanban _hook commit-msg "$@"`) keeps a single implementation in the Go binary while still allowing the hook to be version-controlled and shared.

### Alternative 3: Husky / Git Hook Manager

Use a third-party hook manager (Husky, pre-commit, lefthook) to manage hook installation.

Rejection rationale: these tools are ecosystem-specific (Husky is npm-only; pre-commit is Python). Adding a dependency on an ecosystem-specific hook manager defeats the purpose of a self-contained Go CLI. `kanban install-hook` provides the same one-command installation experience without a secondary tool.

---

## Hook Delivery Architecture

```
.kanban/hooks/commit-msg     (version-controlled; thin wrapper)
    |
    v (kanban install-hook copies/links to)
.git/hooks/commit-msg        (local to developer's machine; not committed)
    |
    v (delegates to)
kanban _hook commit-msg $1   (Go binary handles all logic)
```

The `.kanban/hooks/commit-msg` file contains only:
```sh
#!/bin/sh
exec kanban _hook commit-msg "$1"
```

This is the minimal shell surface area. All logic (pattern extraction, file update, log on error) is in Go.

---

## Consequences

**Positive**:
- Single implementation of hook logic (Go binary); no shell/Go duplication (mitigates R-04)
- `.kanban/hooks/` is committed to the repo; new team members get the hook by cloning, then run `kanban install-hook`
- Exit 0 guarantee is enforced at the Go entry point level; no shell script can accidentally propagate a non-zero exit
- Hook logic is unit-testable via the `GitHookPort` interface in the hexagonal core

**Negative**:
- Developers must run `kanban install-hook` after cloning; the hook is not automatically active
- If the kanban binary is not on `$PATH`, the hook silently no-ops (logs to hook.log) -- mitigated by the exit 0 guarantee and the logged warning

**Mitigation for installation friction (R-01)**: `kanban init` output will include a reminder to run `kanban install-hook`. This is documented in the onboarding flow.
