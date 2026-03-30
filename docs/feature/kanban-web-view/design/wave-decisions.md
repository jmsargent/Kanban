# DESIGN Decisions: kanban-web-view

**Date**: 2026-03-29
**Wave**: DESIGN
**Status**: Final (user feedback incorporated)

---

## Design Decisions

### DD-01: Separate Binary for Web Server

**Decision**: `kanban-web` is a separate binary in `cmd/kanban-web/`, not a subcommand of the existing `kanban` CLI.

**User feedback**: "Separate binary is the way to go."

**Rationale**: The web server has a fundamentally different runtime lifecycle (long-running process vs one-shot CLI command). Separate binaries allow independent compilation, deployment, and dependency wiring. The CLI binary does not need web or remote git dependencies.

**Consequence**: Both binaries share `internal/domain`, `internal/ports`, `internal/usecases`, and secondary adapters. Each has its own primary adapter and wiring in `cmd/`.

**ADR**: ADR-016

---

### DD-02: CLI Binary Moved to cmd/kanban-cli/

**Decision**: The existing CLI entry point moves from `cmd/kanban/` to `cmd/kanban-cli/`.

**User feedback**: "Yes, but also refactor so that the cli binary has its own folder as well."

**Rationale**: Consistency. Both binaries follow the same `cmd/{name}/main.go` convention. Makes the `cmd/` directory self-documenting.

**Consequence**: The `go build` command for the CLI changes from `./cmd/kanban` to `./cmd/kanban-cli`. Makefile, CI config, and goreleaser config need updating. The produced binary name remains `kanban` (configured in goreleaser).

---

### DD-03: Web Server Permitted to Commit and Push

**Decision**: The `kanban-web` binary is allowed to commit task files and push to the remote repository on behalf of authenticated users.

**User feedback**: "Yes, the web server is allowed to commit + push."

**Rationale**: Non-technical users cannot use git. The web server acts as a proxy, committing with the user's display name as the author and pushing with their GitHub PAT for authentication.

**Consequence**: A new `RemoteGitPort` interface is needed. The commit message format should match the CLI's format for consistency. The non-negotiable constraint "kanban binary never auto-commits" applies only to the CLI binary, not the web server (different execution context).

**ADR**: ADR-017

---

### DD-04: Client-Side Cookie for Token Storage

**Decision**: GitHub PAT and display name are stored in an encrypted `HttpOnly`, `Secure`, `SameSite=Strict` cookie.

**User feedback**: "Yes this sounds like a good solution."

**Rationale**: No server-side session state. The server is stateless (easy to blue/green deploy). The token is encrypted with AES-256-GCM so it is not readable even if the cookie is intercepted. `HttpOnly` prevents JavaScript access. `SameSite=Strict` prevents CSRF via cross-origin requests.

**Trade-off**: Cookie size is larger than a session ID approach, but the simplicity of stateless server outweighs this for a single-VM deployment.

---

### DD-05: htmx for Client Interactivity (Not Vanilla JS)

**Decision**: Use htmx instead of vanilla JavaScript for partial page updates.

**User feedback**: "Im considering if htmx would be a good choice, then we do not even have to use explicit javascript."

**Rationale**: htmx enables AJAX-like behavior entirely through HTML attributes. The server returns HTML fragments, which aligns perfectly with Go `html/template` rendering. No JavaScript build toolchain, no bundler, no npm. The entire client-side interactivity is declared in HTML.

**Trade-off**: htmx is a runtime dependency (14KB gzipped JS file loaded from CDN or self-hosted). Accepted because it eliminates the need for any custom JavaScript code.

**ADR**: ADR-018

---

### DD-06: Configurable Git Sync Interval

**Decision**: The interval between `git pull` operations is configurable via CLI flag (`--sync-interval`) and environment variable (`KANBAN_WEB_SYNC_INTERVAL`), defaulting to 60 seconds.

**User feedback**: "I think this could be an environment variable or an argument given to the program at start, which allows for experimenting with the best frequency."

**Rationale**: The optimal sync frequency depends on team activity, e2-micro CPU budget, and GitHub rate limits. Making it configurable allows experimentation without code changes. The 60-second default balances freshness with resource constraints.

---

### DD-07: nginx Reverse Proxy with Blue/Green Deployments

**Decision**: Use nginx (not Caddy) as the reverse proxy, designed for blue/green deployments.

**User feedback**: "How come caddy over nginx? Also I would like to be able to support blue/green deployments, I think nginx has good support for this."

**Rationale**: nginx has mature blue/green deployment support via upstream switching and zero-downtime reload (`nginx -s reload`). The user has a preference for nginx, and it is the more widely known tool. Caddy's automatic HTTPS is convenient but not essential when running behind GCP's infrastructure.

**Blue/green pattern**: Two instances of `kanban-web` run on different ports. nginx routes to the active one. Deployment switches the upstream and reloads nginx. The inactive instance is then updated.

---

### DD-08: Templates Named Per Page

**Decision**: Template files are named by their page purpose, not generically.

**User feedback**: "I think that templates is a bit of a generic name, maybe have one template per page, and name it to the page."

**Files**: `board.html`, `card_detail.html`, `add_task.html`, `token_entry.html`, `layout.html`.

**Rationale**: Self-documenting file names. A developer can find the template for any page without reading code. `layout.html` is the shared structure (HTML head, navigation, footer) included by all page templates.

---

### DD-09: Git Remote Operations in Existing Git Adapter

**Decision**: Remote git operations (Clone, Pull, Add, Commit, Push) are implemented in `internal/adapters/git/`, not in a separate adapter package.

**User feedback**: "I think the adapter is more suited in existing internal/adapters/git."

**Rationale**: All git CLI operations belong in the same adapter. Adding a separate `gitcli/` adapter would violate the "no adapter imports another adapter" rule if it needed to share utility functions. A new `RemoteGitPort` interface provides interface segregation -- the CLI binary only wires `GitPort`, the web binary wires both `GitPort` and `RemoteGitPort`.

---

## Constraints Carried Forward

| Constraint | Source | Impact on Design |
|-----------|--------|-----------------|
| Zero budget (GCP free tier) | Founder (student) | Single e2-micro VM, all components co-located |
| Go 1.22+ | Existing codebase | Shared domain/ports/usecases layers |
| Hexagonal architecture | Existing codebase | Web adapter follows ports-and-adapters pattern |
| Public repos first | DISCUSS D2 | No auth needed for read operations |
| Task file format | Existing CLI | Markdown + YAML front matter, unchanged |
| Atomic file writes | Non-negotiable #2 | Filesystem adapter already handles this |
| Architecture enforcement | Non-negotiable | go-arch-lint rules extended for new packages |
