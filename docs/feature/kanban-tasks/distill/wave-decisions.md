# Wave Decisions: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DISTILL (acceptance-designer)
**Date**: 2026-03-15

---

## Prior Wave Artifacts Consumed

| Artifact | Status | Notes |
|---------|--------|-------|
| discuss/acceptance-criteria.md | Read | 43 ACs across 7 stories + 5 cross-cutting. All traced to scenarios. |
| discuss/story-map.md | Read | Walking skeleton and release slices informed milestone structure. |
| discuss/user-stories.md | Read | 7 stories, all DoR-passed. UAT scenarios used as scenario seeds. |
| discuss/wave-decisions.md | Read | D-05 (hook always exits 0), D-06 (CI fail = no transition) directly expressed in scenarios. |
| design/architecture-design.md | Read | Three driving ports identified. Port-to-port principle applied to all scenarios. |
| design/component-boundaries.md | Read | Exit code contract (0/1/2) and error types informed error path coverage. |
| design/wave-decisions.md | Read | ADR-004 (hook exit-0), ADR-005 (ci-done subcommand) directly expressed in milestone 2. |
| devops/platform-architecture.md | Not found | DEVOPS wave not yet run. CI smoke test written based on DESIGN ADR-005. |
| devops/ci-cd-pipeline.md | Not found | Not blocking. CI step uses kanban ci-done per DESIGN wave spec. |

---

## Decisions Made in DISTILL Wave

### DT-01: Binary subprocess invocation as driving port

**Decision**: All acceptance tests invoke the compiled `kanban` binary as a subprocess via `exec.Command`. No Go package from the `internal/` tree is imported by test code.

**Rationale**: The driving port for a CLI tool is the binary's command interface — flags, arguments, stdout, stderr, and exit codes. Importing internal packages would bypass the port and test implementation details. Subprocess invocation is the only approach that proves the binary works end-to-end, including cobra wiring, adapter instantiation, and output formatting.

**Impact**: Tests require a compiled binary. The `KANBAN_BIN` environment variable allows CI to inject the built binary path. Default path is `../../bin/kanban` relative to the steps package.

---

### DT-02: Real filesystem and real git repository per scenario

**Decision**: Each scenario uses a fresh `os.MkdirTemp` directory with `git init`. No mocks, no fake filesystem, no stubbed git.

**Rationale**: The DESIGN wave explicitly specified this approach (design/wave-decisions.md: "Adapter integration tests: real filesystem in t.TempDir(), real git repo initialised with git init"). Using real infrastructure catches integration wiring bugs that mocks cannot detect — file permission issues, git configuration requirements, atomic write behaviour.

**Impact**: Tests are slightly slower than mock-based tests (10-50ms per scenario for filesystem and git operations). Acceptable at acceptance level. Inner loop unit tests use mocks.

---

### DT-03: Milestone feature files aligned to release slices

**Decision**: Feature files are organised by milestone rather than by user story or component:
- `walking-skeleton.feature`: spans all stories, covers the full developer loop
- `milestone-1-task-crud.feature`: US-01, US-02, US-03, US-06, US-07
- `milestone-2-auto-transitions.feature`: US-04, US-05
- `integration-checkpoints.feature`: cross-cutting ACs, exit codes, CI smoke tests

**Rationale**: The story map defines two release slices. Organising by milestone means the crafter can work through `milestone-1-task-crud.feature` completely before touching hook and CI infrastructure. Cross-cutting error handling is concentrated in `integration-checkpoints.feature` to avoid repetition across other files.

---

### DT-04: @skip tags for one-at-a-time TDD

**Decision**: Only the first scenario in `walking-skeleton.feature` (WS-1, full lifecycle) and the first scenario in `milestone-1-task-crud.feature` (kanban init) start without `@skip`. All 48 other scenarios are tagged `@skip`.

**Wait** — walking skeleton WS-1 is the most complex scenario (requires CLIAdapter + GitHookAdapter + CIPipelineAdapter). A simpler first scenario is WS-2 (CLIAdapter only). The crafter should enable WS-2 first, then WS-1, then focused scenarios.

**Revised sequence**:
1. Enable WS-2 (walking-skeleton.feature, scenario 2: init + add + board)
2. Enable focused init scenario (milestone-1, scenario 1: kanban init setup)
3. Enable focused add scenarios one at a time
4. Enable board scenarios one at a time
5. Enable WS-1 (walking-skeleton.feature, scenario 1: full lifecycle)
6. Enable milestone-2 scenarios one at a time
7. Enable integration-checkpoints scenarios one at a time

The `suite_test.go` uses `Tags: "~@skip"` to run only non-skipped scenarios.

---

### DT-05: Editor simulation in edit step definitions

**Decision**: The `kanban edit` step definitions simulate the editor by writing directly to the task file (bypassing $EDITOR subprocess invocation) for scenarios that test post-edit outcomes (changed fields, board reflects update).

**Rationale**: $EDITOR interaction is an integration concern. The core observable outcome — "which fields changed, board reflects update" — is testable without launching a real editor. A dedicated scenario for AC-06-2 (editor selection) will use an EDITOR env var pointing to a shell script that modifies the file.

**Impact**: Edit scenarios test outcomes through the CLI port (`kanban edit` output + board state), not $EDITOR process management. The AC-06-2 gap scenario is documented in acceptance-review.md.

---

### DT-06: CI smoke test without DEVOPS wave artifacts

**Decision**: The `integration-checkpoints.feature` includes CI/CD smoke test scenarios based on DESIGN wave ADR-005 (`kanban ci-done` subcommand). No DEVOPS platform-specific scenarios (GitHub Actions YAML, GitLab CI templates) are included.

**Rationale**: DEVOPS wave has not run. Platform-specific CI integration tests require the pipeline configuration files that DEVOPS wave produces. The smoke tests validate: binary builds, `--version` responds, full pipeline sequence produces correct board state.

**Impact**: When DEVOPS wave runs, additional platform-specific scenarios should be added to `integration-checkpoints.feature`.

---

## Risks Carried to DELIVER Wave

| ID | Risk | Mitigation |
|----|-----|-----------|
| DT-R-01 | AC-06-2 (editor selection) has no implementation scenario | Gap documented in acceptance-review.md. Crafter adds scenario when implementing `kanban edit`. |
| DT-R-02 | Hook timing step is a placeholder | Crafter must implement `time.Now()` measurement. Scenario exists and is correctly expressed. |
| DT-R-03 | Board performance (100ms/500 tasks) has no scenario | Deferred to Go benchmark test in inner loop. Crafter owns this. |
| DT-R-04 | CI smoke tests assume kanban ci-done subcommand | If CIPipelineAdapter entry point changes, scenario steps need updating. Low risk — DESIGN ADR-005 is definitive. |

---

## Handoff Package for DELIVER Wave (software-crafter)

### Deliverables Produced

| Artifact | Path |
|---------|------|
| Walking skeleton feature | `tests/acceptance/kanban-tasks/walking-skeleton.feature` |
| Milestone 1 feature (CRUD) | `tests/acceptance/kanban-tasks/milestone-1-task-crud.feature` |
| Milestone 2 feature (auto-transitions) | `tests/acceptance/kanban-tasks/milestone-2-auto-transitions.feature` |
| Integration checkpoints feature | `tests/acceptance/kanban-tasks/integration-checkpoints.feature` |
| godog step definitions | `tests/acceptance/kanban-tasks/steps/kanban_steps_test.go` |
| godog suite setup | `tests/acceptance/kanban-tasks/steps/suite_test.go` |
| Test scenario inventory | `docs/feature/kanban-tasks/distill/test-scenarios.md` |
| Walking skeleton rationale | `docs/feature/kanban-tasks/distill/walking-skeleton.md` |
| Acceptance review + gap analysis | `docs/feature/kanban-tasks/distill/acceptance-review.md` |
| Wave decisions (this document) | `docs/feature/kanban-tasks/distill/wave-decisions.md` |

### Implementation Sequence for Crafter

1. Compile the binary: `go build -o bin/kanban ./cmd/kanban`
2. Enable WS-2 in `walking-skeleton.feature` (remove `@skip` from scenario 2)
3. Run `KANBAN_BIN=./bin/kanban go test ./tests/acceptance/kanban-tasks/steps/` — expect failure for business logic reasons (binary does not exist yet)
4. Implement `kanban init`, `kanban add`, `kanban board` until WS-2 passes
5. Enable kanban init focused scenario, implement until it passes, commit
6. Continue one scenario at a time through milestone-1, then milestone-2, then integration-checkpoints
7. Enable WS-1 (full lifecycle) after all milestone scenarios pass

### Mandate Compliance Evidence

| Mandate | Evidence |
|---------|---------|
| CM-A (driving ports only) | step imports: only stdlib + godog. Zero `internal/` imports. |
| CM-B (business language) | grep for database/HTTP/REST/struct/interface in .feature files: 0 hits. |
| CM-C (walking skeleton + focused counts) | 3 walking skeletons. 49 focused scenarios. All titles describe user goals. |

### Peer Review Status

**Conditionally approved** (see acceptance-review.md).

Conditions before story is demonstrable end-to-end:
1. Add editor selection scenario (AC-06-2)
2. Implement timing assertion in hook performance step

Both are low-risk additions that do not affect the walking skeleton or the primary happy paths.
