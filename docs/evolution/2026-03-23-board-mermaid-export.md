# Evolution: board-mermaid-export

**Date**: 2026-03-23
**Feature ID**: board-mermaid-export
**Duration**: Single session — 2026-03-23T06:39:30Z to 2026-03-23T08:14:44Z
**Waves completed**: DISCUSS, DESIGN, DISTILL, DELIVER

---

## Feature Summary

Added `--mermaid` and `--out FILE` flags to the existing `kanban board` command. When `--mermaid` is set, the command emits a GitHub-renderable Mermaid v11 kanban diagram to stdout instead of the standard text board. When `--out FILE` is also set, the diagram is written directly into the named file — either creating it (if absent) or replacing an existing fenced Mermaid kanban block in-place while preserving surrounding content. This is a brownfield CLI adapter extension: no new domain types, use cases, or port interfaces were introduced.

**Business context**: Developers using the kanban CLI can now keep their README or documentation continuously updated with a visual board representation using a single command (`kanban board --mermaid --out README.md`), with no manual copy-paste and no external tooling.

---

## Delivery Execution — All 12 Steps

Steps are grouped by milestone. Each step followed the RED-GREEN-COMMIT cycle.

### Milestone 1 — Core Rendering (Phase 01)

| Step | Name | RED Acceptance | RED Unit | GREEN | Commit | Duration |
|------|------|----------------|----------|-------|--------|----------|
| 01-01 | Walking skeleton: --mermaid flag wired to stdout | PASS | PASS | PASS | PASS | 06:39:30 - 06:42:06 |
| 01-02 | All columns appear as Mermaid sections | PASS | APPROVED_SKIP: acceptance test passed after fixing default column labels; implementation already correct | PASS | PASS | 06:46:13 - 06:54:45 |
| 01-03 | Tasks appear as nodes under their column | PASS | APPROVED_SKIP: acceptance test passed immediately; renderBoardMermaid already emits task nodes from step 01-01 | PASS | PASS | 06:56:25 - 07:00:28 |
| 01-04 | Empty board produces valid Mermaid block | PASS | APPROVED_SKIP: acceptance test passed immediately — no code change required | PASS | PASS | 07:03:26 - 07:05:31 |

### Milestone 2 — Composability and Sanitisation (Phase 02)

| Step | Name | RED Acceptance | RED Unit | GREEN | Commit |
|------|------|----------------|----------|-------|--------|
| 02-01 | --mermaid and --json are mutually exclusive | PASS | PASS | PASS | PASS |
| 02-02 | --mermaid --me filters to current git user | APPROVED_SKIP: test passed immediately — filtered board already passed to renderBoardMermaid | NOT_APPLICABLE | PASS | PASS |
| 02-03 | Task titles with unsafe chars are sanitised | PASS | PASS | PASS | PASS |
| 02-04 | Column labels with special chars are safe | PASS | APPROVED_SKIP: acceptance test passed without unit test — colon in label does not break binary output | PASS | PASS |

### Milestone 3 — File Output (Phase 03)

| Step | Name | RED Acceptance | RED Unit | GREEN | Commit |
|------|------|----------------|----------|-------|--------|
| 03-01 | --out without --mermaid is a usage error | PASS | PASS | PASS | PASS |
| 03-02 | --out creates file when it does not exist | PASS | PASS | PASS | PASS |
| 03-03 | --out errors when file exists but has no kanban block | PASS | PASS | PASS | PASS |
| 03-04 | --out replaces existing kanban block in-place | PASS | PASS | PASS | PASS |

### Post-Implementation Refactoring

A final refactoring pass (refactor-L1-L4) was executed after all 12 acceptance criteria passed. RED/ACCEPTANCE phases not applicable for refactoring. GREEN PASS confirmed all tests still passing after cleanup.

---

## Key Decisions

### DISCUSS Wave

- **D1 — Output to stdout only**: Consistent with Unix composability — user redirects with `>`. Avoids overwrite-vs-append ambiguity. Later superseded by D6/D7 which added the `--out` file-write path.
- **D2 — `--mermaid` flag (not `--format mermaid`)**: Parallel to existing `--json` flag — both are simple booleans. Mirrors established project pattern.
- **D3 — Sanitise unsafe chars rather than reject**: Silent sanitisation is friendlier than rejection; users whose task titles happen to contain quotes still get a usable diagram.
- **D4 — `--mermaid` and `--json` mutually exclusive (exit 2)**: Undefined output format is a usage error. Exit 2 matches project convention.
- **D5 — All 7 stories shipped together**: US-05 and US-06 (sanitisation) are correctness requirements — a diagram with broken Mermaid syntax silently fails on GitHub.
- **D6 — In-place update uses block detection**: The command finds and replaces an existing `\`\`\`mermaid` / `kanban` fenced block. If the block is absent in an existing file, the command errors with a descriptive message rather than appending to an unknown position.
- **D7 — `--out` requires `--mermaid`**: Makes intent explicit; avoids ambiguity about what format gets written.

### DESIGN Wave

- **D1 — `board_mermaid.go` (new file, package `cli`)**: `board.go` was 178 lines; adding ~150 lines of rendering and file-writing logic would harm readability. Dedicated file in the same package follows existing multi-file convention.
- **D2 — No port or interface for file writing**: `writeMermaidToFile` is presentation-layer I/O equivalent to writing to stdout. Tested with `t.TempDir()` at the adapter layer; a `FileWriterPort` would add abstraction without benefit.
- **D3 — Block detection by line scanning (not regex)**: Two-marker line scan is sufficient; regex adds no correctness advantage for this use case.
- **D4 — Sanitisation replaces rather than strips characters**: Stripping `"` would silently lose meaning from titles like `Fix "login" bug`. Replacing with `'` preserves readability.
- **D5 — Mutual exclusion checked before use case execution**: Validation errors (exit 2) should not require disk reads. Checking flags first is cheaper and clearer.

### DISTILL Wave

- **D1 — Custom Go DSL (not Gherkin)**: Consistent with every other acceptance test in the project. A BDD framework would be disproportionate for a one-file CLI extension.
- **D2 — `BOARD.md` for AC-09**: `InAGitRepo()` creates `README.md`, so using it for the "file does not exist" scenario would require deleting it first. `BOARD.md` is a clean alternative.
- **D3 — `README.md` for AC-10**: `InAGitRepo()` creates `README.md` with `# test\n` (no Mermaid block), making AC-10 a natural consequence of the test setup without extra steps.
- **D4 — Walking skeleton is the only test enabled on delivery start**: All 11 remaining tests use `t.Skip` to enforce one-at-a-time RED-GREEN cycle.
- **D5 — `TASK-OLD` as placeholder node in `FileExistsWithKanbanBlock`**: Provides a stable, predictable token to assert has been removed after in-place replacement (AC-11).

---

## Issues Encountered and Resolutions

### Review BLOCKER — Fixed

During the DISTILL wave, an acceptance-review issue was identified: one of the test helper function names had an inconsistency between the test-scenarios.md specification and the actual DSL implementation. This was identified before DELIVER commenced and resolved by aligning the DSL step function names to the documented spec. No production code was affected.

### Approved RED Unit Skips

Four steps had their RED_UNIT phase approved-skipped because the acceptance test already passed after implementation from a previous step — the unit test would have been testing behaviour already covered at the acceptance level with no additional value. These were documented in the execution log with `APPROVED_SKIP` justification.

### Step 02-02 Acceptance Skip

Step 02-02 (`--me` filter) had its RED_ACCEPTANCE phase approved-skipped: the filtered board was already being passed correctly to `renderBoardMermaid` as a consequence of how the existing `--me` flag interacted with the GetBoard use case. No code change was required; the test passed immediately.

---

## Migrated Permanent Artifacts

| Artifact | Permanent Location |
|----------|--------------------|
| Architecture Design | `docs/architecture/board-mermaid-export/architecture-design.md` |
| Component Boundaries | `docs/architecture/board-mermaid-export/component-boundaries.md` |
| Technology Stack | `docs/architecture/board-mermaid-export/technology-stack.md` |
| Data Models | `docs/architecture/board-mermaid-export/data-models.md` |
| ADR-015 (Mermaid rendering in CLI adapter) | `docs/adrs/ADR-015-mermaid-render-in-cli-adapter.md` |
| Test Scenarios | `docs/scenarios/board-mermaid-export/test-scenarios.md` |
| Walking Skeleton | `docs/scenarios/board-mermaid-export/walking-skeleton.md` |
| User Journey (YAML) | `docs/ux/board-mermaid-export/journey-board-mermaid.yaml` |
| User Journey (Visual) | `docs/ux/board-mermaid-export/journey-board-mermaid-visual.md` |

---

## Implementation Files Delivered

| File | Status |
|------|--------|
| `internal/adapters/cli/board_mermaid.go` | New — renderer, sanitisation helpers, file writer |
| `internal/adapters/cli/board.go` | Modified — `--mermaid` and `--out` flags, dispatch logic, mutual exclusion checks |
| `tests/acceptance/board_mermaid_test.go` | New — 12 acceptance scenarios |
| `tests/acceptance/dsl/mermaid_steps.go` | New — DSL steps for Mermaid acceptance tests |
| `docs/adrs/ADR-015-mermaid-render-in-cli-adapter.md` | New — architectural decision record |
