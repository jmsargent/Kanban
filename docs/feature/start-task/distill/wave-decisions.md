# Wave Decisions — kanban start

## Framework

**BDD framework**: godog (Go implementation of Cucumber/Gherkin)
**Feature file format**: Gherkin `.feature` files, organized by milestone
**Step definitions**: Go, in `tests/acceptance/kanban-tasks/steps/kanban_steps_test.go`

## Integration Approach

All acceptance tests invoke the real compiled binary as a subprocess (`exec.Command`). There are no mocks at the acceptance level. The binary path resolves to `tests/acceptance/bin/kanban` (relative to the steps package directory at runtime). Override via `KANBAN_BIN` environment variable.

The `kanbanCtx` struct holds per-scenario state: repo directory, binary path, last stdout/stderr/exit code, and last captured task ID. Each scenario runs in a fresh `os.MkdirTemp` directory with a real `git init`.

## Driving Port

The driving port for all `kanban start` scenarios is the CLI command: `kanban start <task-id>`.

Tests enter exclusively through this port. Internal components (use case, task repository, filesystem adapter) are exercised indirectly as a consequence of invoking the CLI. No internal packages are imported.

## Mandate Compliance

- **CM-A**: Step definitions import no internal packages. The only external import is `github.com/cucumber/godog`. The binary is invoked via `os/exec`.
- **CM-B**: All Gherkin uses business language. Zero technical terms (no HTTP, no JSON, no repository, no adapter) appear in `.feature` files.
- **CM-C**: Every scenario validates a complete user journey: a user action via CLI, the resulting business state change (or rejection), and an observable outcome (stdout/stderr message and exit code).

## Constraints

- `osExit` in `start.go` is a package-level variable to allow unit-test override without affecting acceptance tests, which measure the real process exit code via `exec.ExitError`.
- The "not initialised" scenario requires no Background `kanban init` call. This is handled by per-scenario `Given the repository has no kanban setup` after the Background's `Given I am working in a git repository`.
- The "already in-progress" scenario asserts both exit code 0 and no state change — this is an idempotence guarantee, not an error path.
