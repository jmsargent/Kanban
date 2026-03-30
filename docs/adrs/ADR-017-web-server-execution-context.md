# ADR-017: Web Server Execution Context (Commit and Push)

## Status

Accepted

## Context

The kanban CLI has a non-negotiable constraint: "The kanban binary never auto-commits on behalf of the developer" (except `kanban ci-done` which is explicitly a CI step). Non-technical users need to add tasks via the web interface, which requires creating a task file, committing it, and pushing to the remote repository.

We need to decide whether the web server is allowed to perform git commit and push operations, and how this relates to the existing non-negotiable constraint.

## Decision

The `kanban-web` binary is a distinct execution context that is explicitly permitted to commit and push on behalf of authenticated users. This does not violate the existing constraint, which applies to the interactive CLI execution context.

The three execution contexts are:

1. **Interactive CLI** (`kanban-cli`): Never auto-commits. User controls git workflow.
2. **Commit-msg hook**: Read-only. Always exits 0. (ADR-004)
3. **CI pipeline** (`kanban ci-done`): Explicitly commits as a CI step. (ADR-005)
4. **Web server** (`kanban-web`): Commits and pushes on behalf of authenticated users. (This ADR)

A new `RemoteGitPort` interface provides `Clone`, `Pull`, `Add`, `Commit`, and `Push` operations. It is implemented in the existing `internal/adapters/git/` package. Only `kanban-web` wires this port; the CLI binary does not.

The `Push` method accepts the user's GitHub PAT as a parameter, injecting it into the HTTPS remote URL for authentication. The token is never stored on disk.

## Alternatives Considered

- **Web server creates task files but does not commit**: Users would need to manually commit/push from the CLI or a separate process. Rejected because the target users are non-technical and cannot use git. This defeats the purpose of the web interface.

- **Separate background worker for commit/push**: A cron job or daemon periodically commits and pushes uncommitted changes. Rejected because it adds complexity (two processes to manage), introduces delay between task creation and visibility in GitHub, and makes error attribution harder (which user's token to use?).

## Consequences

- Positive: Non-technical users can add tasks end-to-end without any git knowledge.
- Positive: Each commit is attributed to the specific user (via display name) and authenticated with their token.
- Negative: The web server needs write access to the local git clone. Concurrent writes must be serialized (mutex). Accepted because write volume is expected to be low.
- Negative: If a push fails (network, permissions), the local clone diverges from remote. Mitigated by returning a clear error to the user and not advancing the local branch on push failure.
