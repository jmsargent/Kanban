# Evolution: task-creator-attribution

**Date**: 2026-03-17
**Feature ID**: task-creator-attribution
**Status**: IMPLEMENTED

---

## Feature Summary

When `kanban new` is run, the creator's git identity (user.name) is automatically captured and
stored as `created_by` in the task's YAML front matter. The board displays it in a new rightmost
column; missing identity triggers exit 1 with guidance; edit cannot overwrite it.

---

## Business Context

Developer attribution for audit and accountability. Tasks created with `kanban new` now carry
the name of the developer who created them, readable from the task file and visible on
`kanban board`. This supports team-level visibility into work ownership and provides an
immutable audit record at creation time — without requiring any manual input from the developer.

Classification: completeness feature. The gap (tasks have no creator field) was identified by the
feature owner. There is no support ticket or usage metric driving this; adoption is automatic and
opt-out-impossible by design.

---

## Steps Completed

The feature was delivered in 5 phases across 9 roadmap steps using the DELIVER wave TDD cycle
(RED_ACCEPTANCE → RED_UNIT → GREEN → COMMIT per step).

| Phase | Steps | Status | Description |
|-------|-------|--------|-------------|
| 01 Foundation | 01-01 | EXECUTED / PASS | Domain and ports layer: added `CreatedBy` to `domain.Task`, `Identity` struct and `GetIdentity()` to `ports.GitPort`, `ErrGitIdentityNotConfigured` sentinel |
| 02 Walking Skeleton | 02-01 | EXECUTED / PASS | Core task creation with identity wiring: `GitAdapter.GetIdentity()`, `AddTaskInput.CreatedBy`, filesystem serialization, CLI wiring in `new.go` |
| 02 Walking Skeleton | 02-02 | SKIPPED (pre-passing) | Front matter completeness verification: already passing from 02-01 |
| 02 Walking Skeleton | 02-03 | SKIPPED (pre-passing) | Atomic write safety: already passing from 02-01 |
| 03 Board Display | 03-01 | EXECUTED / PASS | Board text output shows creator name: Created By column added to text renderer |
| 03 Board Display | 03-02 | SKIPPED (pre-passing) | Pre-existing task shows -- on board: already passing from 03-01 |
| 03 Board Display | 03-03 | EXECUTED / PASS | JSON output includes `created_by` field: `boardTaskJSON.CreatedBy` added |
| 04 Error Path | 04-01 | EXECUTED / PASS | Missing git identity fails with guidance: `ErrGitIdentityNotConfigured` guard in `new.go`, exit 1 with corrective instructions |
| 05 Immutability | 05-01 | EXECUTED / PASS | Edit does not change creator: structural exclusion confirmed, no production code changes required |

Roadmap total: 9 steps, 5 phases. Created at 2026-03-17T20:46:10Z. Validation approved by
solution-architect-reviewer at 2026-03-17T20:46:10Z.

---

## Key Decisions

### DISCUSS Wave

- **[D1] Identity source: git config user.name (name only)**. Git identity is already required
  for the project; no new credentials or setup. Name only — not email — to minimize personal
  data in task files.

- **[D2] Failure mode: exit 1 with setup instructions (not interactive prompt)**. Interactive
  prompts during `kanban new` would be surprising. A clear error with the exact `git config`
  commands lets the developer fix the root cause once.

- **[D3] Storage: YAML front matter field `created_by`**. Consistent with all other task
  metadata. Grep-able, diffable, human-readable.

- **[D4] Immutability: excluded from kanban edit temp file**. Creator is an audit field, not an
  editable property. Excluding it from the edit workflow is the simplest way to guarantee
  immutability without domain-level enforcement.

- **[D5] Port design: extend GitPort with GetIdentity()**. Keeps all git concerns behind a
  single interface. CLI adapter calls GetIdentity(), resolves the name, and injects it into
  AddTaskInput.CreatedBy. The use case remains unaware of git config.

- **[D6] Pre-existing tasks: display "--" (no retroactive migration)**. Backfilling from git
  blame introduces risk and complexity disproportionate to benefit.

- **[D7] Board display: Created By column added as rightmost column**. Least disruptive to
  existing layout.

### DESIGN Wave

- **[D1] Extend GitPort with GetIdentity() — not a separate IdentityPort**. All git subprocess
  concerns belong behind a single interface. Adding one method to the existing interface is the
  minimal correct solution.

- **[D2] Identity struct carries both Name and Email; only Name is stored**. Reading both avoids
  a second subprocess call if email is ever needed by a future feature.

- **[D3] ErrGitIdentityNotConfigured sentinel defined in ports/errors.go**. Consistent with
  existing sentinels (ErrTaskNotFound, ErrNotInitialised). Handled in CLI adapter — never
  propagated to the use case.

- **[D4] Immutability by structural exclusion — not domain-level enforcement**. Excluding
  `created_by` from `editFields` and `EditSnapshot` is the simplest mechanism. Zero
  domain-layer changes beyond adding the field.

- **[D5] No new packages — all changes in existing packages**. The feature is additive. No new
  package boundaries warranted for a field addition, one new port method, and two display changes.

- **[D6] `created_by` YAML field uses no omitempty tag**. Pre-existing files parse cleanly as
  empty string. omitempty would silently omit an empty `created_by` if the guard were ever
  bypassed — making the bug invisible.

### DISTILL Wave

- **[D1] Internal Go BDD DSL — not Gherkin/godog**. ADR-006 replaced godog with the internal
  Go DSL. All acceptance tests follow the `TestXxx(t)` + `dsl.Given/When/Then` pattern.

- **[D2] New step factories in a separate file `dsl/creator_steps.go`**. Keeps creator-specific
  steps isolated without touching existing setup.go and assertions.go.

- **[D3] `InAGitRepoWithoutGitIdentity()` uses isolated HOME + GIT_CONFIG_NOSYSTEM**. Most
  portable way to ensure `git config user.name` returns empty regardless of the developer's
  global ~/.gitconfig.

- **[D4] `APreExistingTaskWithoutCreator()` writes the file directly (no binary call)**. The
  only way to produce a file in the legacy format, accurately representing the
  backward-compatibility scenario.

- **[D5] AC-04-1 tested via AC-04-2 behavioral outcome**. AC-04-1 (edit temp file does not
  expose `created_by`) cannot be verified at acceptance level; AC-04-2 proves the behavioral
  outcome end-to-end.

---

## Architecture Decisions

### Hexagonal Boundary Respected

The dependency rule (all arrows inward toward domain) is fully preserved. No new package
boundaries were introduced. All changes are additive within existing packages.

- `internal/domain` has zero new imports. `CreatedBy string` is a plain field.
- `internal/usecases` has zero new imports from `internal/adapters`.
- No adapter imports another adapter.

### GitPort Extended with GetIdentity()

`ports.GitPort` gained one new method: `GetIdentity() (Identity, error)`. A new `Identity`
struct `{Name string, Email string}` is defined in `internal/ports/git.go`. The `GitAdapter`
implements this via two `git config` subprocess calls following the existing `runGitIn` pattern.
Compile-time interface compliance is enforced by the existing `var _ ports.GitPort = (*GitAdapter)(nil)`
guard.

### ErrGitIdentityNotConfigured Sentinel

Defined in `internal/ports/errors.go` alongside `ErrTaskNotFound` and `ErrNotInitialised`.
Returned by `GitAdapter.GetIdentity()` when `git config user.name` produces an empty or
whitespace-only result. Handled exclusively in `cli/new.go` — the use case never sees it.

### Structural Exclusion for Immutability

Creator immutability is enforced by structural exclusion: `editFields` struct in
`filesystem/task_repository.go` does not include `created_by`. `WriteTemp` writes only
`editFields`; `applyEditFields` reads only `EditSnapshot` fields; `Task.CreatedBy` from the
original `FindByID` call passes through the edit cycle unchanged and is written back by `Update`.
No domain-level enforcement is needed.

### ADR Reference

ADR-007 (`docs/adrs/ADR-007-git-identity-port-extension.md`) records the formal decision to
extend `GitPort` with `GetIdentity()` and defines the `Identity` type, error sentinel, and
immutability mechanism.

---

## Acceptance Tests

9 acceptance tests, all passing. Verified by adversarial review — APPROVED.

| # | Test Function | Acceptance Criterion | Category |
|---|--------------|---------------------|---------|
| 1 | TestTaskCreator_CreatorRecordedOnKanbanNew | AC-01-1 | Walking Skeleton / Happy Path |
| 2 | TestTaskCreator_FrontMatterContainsCreatorField | AC-01-2 | Happy Path |
| 3 | TestTaskCreator_AtomicWriteNoTempFiles | AC-01-3 | Non-Functional / Safety |
| 4 | TestTaskCreator_BoardShowsCreatorName | AC-02-1 | Happy Path |
| 5 | TestTaskCreator_PreExistingTaskShowsDashOnBoard | AC-02-2 | Edge Case / Backward Compatibility |
| 6 | TestTaskCreator_JSONOutputHasCreatedByField | AC-02-3 | Happy Path |
| 7 | TestTaskCreator_MissingIdentityFailsWithGuidance | AC-03-1 | Error Path |
| 8 | TestTaskCreator_MissingIdentityWritesNoFile | AC-03-2 | Error Path / Safety |
| 9 | TestTaskCreator_EditDoesNotChangeCreator | AC-04-2 | Edge Case / Audit Integrity |

Test file: `tests/acceptance/task_creator_test.go`
DSL step factories: `tests/acceptance/dsl/creator_steps.go`
Framework: Internal Go BDD DSL (ADR-006) — real compiled binary as subprocess, no mocks at
acceptance level.

---

## Mutation Testing

Disabled per project configuration (CLAUDE.md: "Mutation testing is disabled. Test quality
validated through code review and CI coverage.").

---

## Files Changed

| File | Change |
|------|--------|
| `internal/domain/task.go` | Added `CreatedBy string` to `Task` struct |
| `internal/ports/git.go` | Added `Identity` type + `GetIdentity() (Identity, error)` to `GitPort` |
| `internal/ports/errors.go` | Added `ErrGitIdentityNotConfigured` sentinel |
| `internal/usecases/add_task.go` | Added `CreatedBy string` to `AddTaskInput`; assigned in `Execute` |
| `internal/adapters/git/git_adapter.go` | Implemented `GetIdentity()` |
| `internal/adapters/filesystem/task_repository.go` | Added `created_by` to `taskFrontMatter`; excluded from `editFields` |
| `internal/adapters/cli/new.go` | Added `GetIdentity()` call and error guard before use case |
| `internal/adapters/cli/board.go` | Added Created By column (text) and `created_by` field (JSON) |
| `tests/acceptance/task_creator_test.go` | 9 new acceptance tests |
| `tests/acceptance/dsl/creator_steps.go` | New DSL step factories for creator scenarios |

---

## Permanent Artifacts

- Architecture: `docs/architecture/task-creator-attribution/`
- Test scenarios: `docs/scenarios/task-creator-attribution/`
- UX journeys: `docs/ux/task-creator-attribution/`
- ADR: `docs/adrs/ADR-007-git-identity-port-extension.md`
