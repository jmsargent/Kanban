# ADR-016: Separate Binary for Web Server

## Status

Accepted

## Context

kanban-web-view needs a Go HTTP server to render the kanban board in a browser. The existing system has a single binary (`kanban`) that provides CLI commands. We need to decide whether the web server is a subcommand of the existing binary or a separate binary.

The web server has a fundamentally different runtime model: it is a long-running process (HTTP server with background sync), whereas the CLI is a one-shot command executor. The web server also requires additional dependencies (remote git operations, cookie encryption, HTTP handler wiring) that the CLI does not need.

Additionally, the existing CLI binary entry point at `cmd/kanban/` is being refactored to `cmd/kanban-cli/` so both binaries follow the consistent `cmd/{name}/main.go` convention.

## Decision

Create a separate binary `kanban-web` at `cmd/kanban-web/main.go`. Rename the existing CLI entry point from `cmd/kanban/main.go` to `cmd/kanban-cli/main.go`.

Both binaries share `internal/domain`, `internal/ports`, `internal/usecases`, and secondary adapters. Each has its own primary adapter (`internal/adapters/cli/` for the CLI, `internal/adapters/web/` for the web server) and its own wiring in `cmd/`.

## Alternatives Considered

- **Subcommand of existing binary** (`kanban serve`): Simpler build output, single binary to distribute. Rejected because it couples CLI and web dependencies, increases binary size for CLI users, and conflates two different runtime models (one-shot vs long-running). The CLI's non-negotiable "never auto-commits" constraint does not apply to the web server, making shared wiring confusing.

- **Completely separate Go module**: Maximum isolation. Rejected because it would duplicate the domain model, ports, and use cases. The shared internal packages are the primary value of the monorepo layout.

## Consequences

- Positive: Clean separation of concerns. Each binary only wires what it needs. Independent build, test, and deploy cycles. CLI binary stays lean.
- Positive: Both binaries benefit from shared domain logic. Bug fixes in use cases or domain automatically apply to both.
- Negative: Build and release pipeline (goreleaser) must produce two binaries. Makefile and CI config need updates for the `cmd/kanban-cli/` rename.
- Negative: The `cmd/kanban/` rename is a breaking change for existing build scripts. Mitigated by updating Makefile, goreleaser, and CI in the same PR.
