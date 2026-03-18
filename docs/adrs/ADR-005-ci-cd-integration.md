# ADR-005: CI/CD Integration Pattern — kanban Binary in CI Step

**Status**: Amended
**Date**: 2026-03-15
**Amended**: 2026-03-18
**Feature**: kanban-tasks; board-state-in-git
**Resolves**: OD-06
**Amendment**: `kanban ci-done` commits only `.kanban/transitions.log`, not task files. The `[skip ci]` annotation and all other decisions are unchanged.

---

## Context

The auto-transition from "in-progress" to "done" must fire when a CI pipeline passes. The CI step must:

1. Run only when all tests pass (exit code 0)
2. Scan commit messages in the pipeline run for task IDs matching `ci_task_pattern`
3. Update matching task files from "in-progress" to "done"
4. Commit the updated task files back to the repository
5. Run in a non-TTY environment with no interactive prompts

Deliverables are staged across releases: a generic shell wrapper first (Walking Skeleton + R1), then platform-specific wrappers for GitHub Actions and GitLab CI (R3).

---

## Decision

The CI step delegates to the kanban binary: `kanban ci-done --since=<base-ref> --commit`.

- `--since=<base-ref>`: scans commit messages from `<base-ref>` to `HEAD` (the commits in this pipeline run)
- `--commit`: after updating task files, runs `git add .kanban/tasks/ && git commit -m "kanban: auto-transition [skip ci]"` non-interactively

The CI step is delivered in three tiers:

**Tier 1 (Walking Skeleton)**: A shell script `kanban-ci.sh` committed to the repository. CI platforms invoke it after the test step. Requires the kanban binary to be present (downloaded in the CI setup step).

**Tier 2 (R3 — GitHub Actions)**: A composite GitHub Action (`kanban-ci` action) that downloads the correct kanban binary for the runner OS, then invokes `kanban ci-done`. Published to the GitHub Actions marketplace.

**Tier 3 (R3 — GitLab CI)**: A `.gitlab-ci.yml` template snippet that mirrors the GitHub Actions pattern. Documented in the repository wiki.

The `[skip ci]` annotation in the commit message prevents the kanban commit from triggering another pipeline run (supported by GitHub Actions and GitLab CI natively).

---

## Alternatives Considered

### Alternative 1: CI Step as a Pure Shell Script (No Binary Dependency)

Implement the CI step entirely in shell, parsing commit messages and updating YAML with `sed`/`awk`. No binary download required.

Rejection rationale: replicates task file parsing and status update logic in shell, creating a second implementation divergent from the Go binary (R-04). A shell-only implementation cannot share the `ci_task_pattern` config parsing logic with the hook, violating the single-implementation principle from ADR-004. Shell fragility around YAML parsing is a known failure mode for the status field update.

### Alternative 2: Separate `kanban-ci` Binary

Produce a separate, minimal Go binary (`kanban-ci`) containing only the CI step logic.

Rejection rationale: two binaries to distribute and version-pin in CI configurations. A subcommand (`kanban ci-done`) achieves the same separation with one binary to manage. Version consistency between the hook and CI step is automatic with a single binary.

### Alternative 3: GitHub App / Webhook Integration

A server-side component receives CI webhooks and updates task files via git API.

Rejection rationale: introduces a server-side dependency, eliminating the git-native, no-external-service property that is a core design constraint (D-02 from DISCUSS wave). Out of scope for the MVP.

---

## CI Environment Compatibility

The `kanban ci-done` subcommand must handle the following CI environment conditions:

| Condition | Handling |
|-----------|---------|
| No TTY | Detected via `os.Stdin` isatty check; all output is plain text, no ANSI, no spinners |
| `NO_COLOR` set | Color output suppressed (inherited from chalk-equivalent in Go: `fatih/color`) |
| Git identity not configured | `kanban ci-done` sets `GIT_AUTHOR_NAME` and `GIT_COMMITTER_NAME` to "kanban-bot" for the commit if not set |
| No tasks to transition | Exits 0 with no output (silent success) |
| `git push` fails on CI commit | Logged as warning; exits 0 (never fails the pipeline) |

---

## Consequences

**Positive**:
- Single binary handles both hook and CI step -- no logic duplication (R-04 mitigated)
- `[skip ci]` annotation prevents infinite CI loops
- Tier 1 shell wrapper works on any CI platform immediately; R3 adds ergonomics
- `kanban ci-done` is testable in isolation via the `CIPipelinePort` in the hexagonal architecture

**Negative**:
- CI setup step must download the kanban binary (adds ~5s to pipeline; within the 5s guardrail NFR)
- The `git commit` inside `kanban ci-done` requires the CI runner to have git credentials configured for push -- standard for most CI setups but requires documentation

**Security note**: the CI step commits only to `.kanban/transitions.log`. It does not read or write outside the `.kanban/` directory. The commit message is hardcoded (`kanban: auto-transition [skip ci]`). No user-supplied input is interpolated into shell commands.

---

## Amendment — 2026-03-18 (board-state-in-git feature)

**Changed behaviour**: `kanban ci-done` no longer updates `status:` fields in task YAML front matter. Instead it calls `TransitionLogRepository.Append` for each in-progress task matching the scanned commit range, then commits only `.kanban/transitions.log`.

**CommitFiles scope change**:

Before amendment: `git add .kanban/tasks/ && git commit`
After amendment: `git add .kanban/transitions.log && git commit`

This narrows the commit surface from all updated task files to a single file. Benefits:
- Fewer merge conflicts (a single append-only file has no conflicts for concurrent appends from different branches merged together)
- Clearer audit trail (the transitions.log commit contains only state changes, not task metadata changes)

**TransitionEntry written by ci-done**:
- `From`: `in-progress` (validated via `TransitionLogRepository.LatestStatus` before appending)
- `To`: `done`
- `Author`: CI git identity (from `GIT_AUTHOR_EMAIL` env or `kanban-bot` fallback)
- `Trigger`: `ci-done:<sha7>` where `<sha7>` is the short SHA of the pipeline's HEAD commit
- `Timestamp`: current UTC time at ci-done execution
