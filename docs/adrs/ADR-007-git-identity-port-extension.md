# ADR-007: Extend GitPort with GetIdentity() for Creator Attribution

**Status**: Accepted
**Date**: 2026-03-17
**Feature**: task-creator-attribution

---

## Context

The `task-creator-attribution` feature requires reading the developer's name from
`git config user.name` at task creation time (`kanban new`). This name is stored as
`created_by` in the task's YAML front matter and displayed on `kanban board`.

The existing `GitPort` interface (ADR-001) already owns all git subprocess concerns.
The question is: where should the identity resolution capability live?

Three options were considered:

1. **Extend `GitPort`** with a `GetIdentity()` method
2. **Separate `IdentityPort`** — a new interface for identity concerns
3. **Resolve identity in the CLI adapter directly** — parse `git config` output in `new.go`
   without a port abstraction

Quality attributes driving this decision:
- **Testability**: use cases must remain testable without real git subprocess calls
- **Maintainability**: git concerns belong in one place
- **Architecture compliance**: no adapter may call another adapter; git concerns must cross
  the hexagon boundary via a port interface (ADR-001)

---

## Decision

Extend `ports.GitPort` with a new method:

```go
GetIdentity() (Identity, error)
```

where `Identity` is a new struct in `internal/ports/git.go`:

```go
type Identity struct {
    Name  string
    Email string
}
```

`GitAdapter.GetIdentity()` implements this by running `git config user.name` and
`git config user.email` as subprocesses.

A new sentinel error `ErrGitIdentityNotConfigured` is added to `internal/ports/errors.go`.

The **CLI adapter** (`cli/new.go`) calls `GetIdentity()`, checks for empty `Name`, and exits
with code 1 + setup instructions **before** invoking the use case. The use case receives
`CreatedBy` as part of `AddTaskInput` — it has no knowledge of git identity resolution.

---

## Alternatives Considered

### Alternative 1: Separate IdentityPort

Define a new interface `IdentityPort` in `internal/ports/`:

```go
type IdentityPort interface {
    GetIdentity() (Identity, error)
}
```

**Rejection rationale**: A separate interface would require a separate adapter implementation
and a new wiring point in `cmd/kanban/main.go`. Since `GetIdentity()` is semantically a git
concern (it reads `git config`), splitting it creates an artificial boundary. The single
`GitAdapter` already handles all other git subprocess calls. Extending `GitPort` is the
minimal, correct solution. A separate `IdentityPort` would only be justified if identity
could be sourced from a non-git backend — which is explicitly out of scope.

### Alternative 2: Inline git config parsing in cli/new.go

Read `git config user.name` directly in the CLI adapter, bypassing the port abstraction:

```go
out, _ := exec.Command("git", "config", "user.name").Output()
name := strings.TrimSpace(string(out))
```

**Rejection rationale**: This violates hexagonal architecture (ADR-001). Adapters must not
call subprocesses directly — all external interactions cross the hexagon via port interfaces.
The CLI adapter would become harder to test (requires a real git subprocess even for unit
tests of the adapter logic), and the pattern would diverge from `git.RepoRoot()` which is
already correctly behind the port.

---

## Consequences

**Positive**:
- All git concerns remain behind a single `GitPort` interface — one place to mock in tests
- `GetIdentity()` is independently testable in the git adapter integration tests
- The use case (`AddTask`) remains completely decoupled from git — `CreatedBy` is injected
  as plain data through `AddTaskInput`, consistent with the existing `Assignee` field pattern
- `cmd/kanban/main.go` wiring is unchanged — no new port to inject

**Negative**:
- Extending `GitPort` requires all existing mock implementations of `GitPort` (in test files)
  to add `GetIdentity()`. This is a one-time cost, caught by the compiler.
- `Email` is read but never stored — minor inefficiency of a second subprocess call.
  Acceptable given the negligible latency of `git config` reads.

**Accepted trade-off**: The one-time mock update cost is worth the architectural cleanliness
of keeping all git concerns behind a single, well-defined port.
