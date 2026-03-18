# Technology Stack — auto-assign-on-start

## Stack

Unchanged from the established project stack. No new dependencies.

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| Language | Go 1.22+ | Project language (ADR-003) |
| CLI framework | cobra | Already in use for all commands (ADR-003) |
| Git identity | `git config user.name` via `GitAdapter.GetIdentity()` | Port already exists (ADR-007) |
| Test framework | stdlib `testing` package + godog (acceptance) | Established test strategy (CLAUDE.md) |

## New Dependencies

None. `GitPort.GetIdentity()` and `ports.ErrGitIdentityNotConfigured` are already compiled into the binary.

## Development Paradigm

**OOP / idiomatic Go** — established in CLAUDE.md. Use `@nw-software-crafter` for implementation.
