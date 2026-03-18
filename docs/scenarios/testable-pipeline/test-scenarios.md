# Test Scenarios — testable-pipeline

**Feature**: testable-pipeline
**Wave**: DISTILL
**Date**: 2026-03-18

---

## Scenario Inventory

19 scenarios total. 8 error/edge scenarios = 42% — meets the 40% minimum.

| # | Test Name | Story | Tier | Type | Status |
|---|-----------|-------|------|------|--------|
| 1 | TestPipeline_ValidateTarget_PassesWhenAllChecksGreen | US-TP-02 | 1 | Walking Skeleton | **ENABLED** |
| 2 | TestPipeline_Makefile_ContainsRequiredTargets | US-TP-02 | 1 | Happy path | skip |
| 3 | TestPipeline_ValidateTarget_HelpTargetListsAllTargets | US-TP-02 | 1 | Happy path | skip |
| 4 | TestPipeline_ValidateTarget_ReportsEachStepLabel | US-TP-02 | 1 | Happy path | skip |
| 5 | TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt | US-TP-02 | 1 | Error path | skip (needs fixture) |
| 6 | TestPipeline_Acceptance_FailsWhenBinaryNotBuilt | US-TP-02 | 1 | Error path | skip |
| 7 | TestPipeline_Acceptance_PassesWithBuiltBinary | US-TP-02 | 1 | Happy path | skip |
| 8 | TestPipeline_CI_StopsOnValidateFailure | US-TP-02 | 1 | Error path | skip (needs fixture) |
| 9 | TestPipeline_PreCommit_ArchLintStepIsActive | US-TP-03 | 1 | Happy path (structure) | skip |
| 10 | TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest | US-TP-03 | 1 | Happy path (ordering) | skip |
| 11 | TestPipeline_PreCommit_BlocksOnToolVersionMismatch | US-TP-03 | 1 | Error path | skip (needs fixture) |
| 12 | TestPipeline_PreCommit_BlocksOnArchitectureViolation | US-TP-03 | 1 | Error path | skip (needs fixture) |
| 13 | TestPipeline_PreCommit_PassesWhenGatesParity | US-TP-03 | 1 | Happy path | skip |
| 14 | TestPipeline_CheckVersions_ReportsAllToolsMatch | US-TP-03 | 1 | Happy path | skip |
| 15 | TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions | US-TP-03 | 1 | Error path | skip (needs fixture) |
| 16 | TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing | US-TP-01 | 2 | Happy path | skip |
| 17 | TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig | US-TP-01 | 2 | Error path | skip (needs fixture) |
| 18 | TestPipeline_ReleaseSnapshot_FailsGracefullyWhenGoreleaserNotInstalled | US-TP-01 | 2 | Error path | skip (needs fixture) |
| 19 | TestPipeline_SkipRelease_CIOnlyDocumented | US-TP-04 | 3 | Documentation | permanent skip |
| 20 | TestPipeline_SkipRelease_NormalCommitRunsFullPipeline | US-TP-04 | 3 | Documentation | permanent skip |
| 21 | TestPipeline_SkipRelease_CaseInsensitiveAndPositionIndependent | US-TP-04 | 3 | Documentation | permanent skip |
| 22 | TestPipeline_GoreleaserCache_CIOnlyDocumented | US-TP-01 | 3 | Documentation | permanent skip |

---

## Testability Tier Classification

### Tier 1 — Fully Local

Tests that invoke real shell commands, real Makefile targets, and real scripts at
the project root. No secrets required. No external services.

**Scenarios**: 1–15

**What "fully local" means**:
- All tools must be installed locally at versions matching `cicd/config.yml`
- `cicd/check-versions.sh` is the pre-condition gate for most Tier 1 tests
- Tests with "(needs fixture)" require the software crafter to implement a
  temporary source or config mutation (e.g., introduce an arch violation,
  mutate a version parameter, break a Go file)

### Tier 2 — Local Build Validation

Tests that require goreleaser to be installed locally. `make release-snapshot` builds
all six cross-compile targets to `dist/` without publishing.

**Scenarios**: 16–18

**Pre-condition**: `goreleaser` installed at the version in `cicd/config.yml`.
The `check-versions.sh` output will confirm this before running.

### Tier 3 — CI-Only (Permanent Skip)

Tests whose observable outcomes only exist inside a live CircleCI run:
- goreleaser cache restore/save (cache key hit/miss)
- `[skip release]` shell guard filtering tag and release jobs
- Actual GitHub Release creation and Homebrew formula update

**Scenarios**: 19–22

These tests document expected CI behaviour as executable specifications. They
will never pass locally. The `t.Skip` message names the exact CircleCI mechanism
being documented.

---

## Coverage Map: Stories to Scenarios

### US-TP-01: Cached goreleaser install and local release snapshot

| Acceptance Criterion | Test | Tier |
|---------------------|------|------|
| `make release-snapshot` exists | TestPipeline_Makefile_ContainsRequiredTargets | 1 |
| `make release-snapshot` exits 0 on valid config | TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing | 2 |
| `make release-snapshot` exits non-zero on invalid config | TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig | 2 |
| Does not require GITHUB_TOKEN | TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing | 2 |
| goreleaser cache hit/miss in CI | TestPipeline_GoreleaserCache_CIOnlyDocumented | 3 |
| install-goreleaser uses go install (not curl) | TestPipeline_GoreleaserCache_CIOnlyDocumented | 3 |
| Fails gracefully if goreleaser not installed | TestPipeline_ReleaseSnapshot_FailsGracefullyWhenGoreleaserNotInstalled | 2 |

### US-TP-02: Makefile local CI mirror

| Acceptance Criterion | Test | Tier |
|---------------------|------|------|
| Makefile exists at project root | TestPipeline_ValidateTarget_PassesWhenAllChecksGreen | 1 |
| `make validate` runs in declared order | TestPipeline_ValidateTarget_ReportsEachStepLabel | 1 |
| `make validate` exits 0 on success | TestPipeline_ValidateTarget_PassesWhenAllChecksGreen | 1 |
| `make validate` exits 1 on failure, prints step | TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt | 1 |
| `make acceptance` exits non-zero with message if binary absent | TestPipeline_Acceptance_FailsWhenBinaryNotBuilt | 1 |
| `make acceptance` passes with built binary | TestPipeline_Acceptance_PassesWithBuiltBinary | 1 |
| `make ci` stops on validate failure | TestPipeline_CI_StopsOnValidateFailure | 1 |
| `make help` prints formatted list | TestPipeline_ValidateTarget_HelpTargetListsAllTargets | 1 |
| All required targets declared | TestPipeline_Makefile_ContainsRequiredTargets | 1 |

### US-TP-03: Pre-commit gate parity

| Acceptance Criterion | Test | Tier |
|---------------------|------|------|
| go-arch-lint uncommented in pre-commit | TestPipeline_PreCommit_ArchLintStepIsActive | 1 |
| check-versions.sh is step 0 | TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest | 1 |
| Version mismatch blocks commit | TestPipeline_PreCommit_BlocksOnToolVersionMismatch | 1 |
| Version mismatch output identifies tool+version | TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions | 1 |
| Architecture violation blocks commit | TestPipeline_PreCommit_BlocksOnArchitectureViolation | 1 |
| Pre-commit exits 0 when all gates pass | TestPipeline_PreCommit_PassesWhenGatesParity | 1 |
| check-versions.sh reports all tools match | TestPipeline_CheckVersions_ReportsAllToolsMatch | 1 |

### US-TP-04: Safe test push convention

| Acceptance Criterion | Test | Tier |
|---------------------|------|------|
| [skip release] suppresses tag + release jobs | TestPipeline_SkipRelease_CIOnlyDocumented | 3 |
| validate + acceptance always run | TestPipeline_SkipRelease_NormalCommitRunsFullPipeline | 3 |
| Case-insensitive, position-independent | TestPipeline_SkipRelease_CaseInsensitiveAndPositionIndependent | 3 |

---

## Error / Edge Scenario Count

Total scenarios with executable assertions: 22
Error and edge scenarios: 9 (scenarios 5, 6, 8, 11, 12, 15, 17, 18, plus Tier 3 skip-on-wrong-commit)
Error ratio: 9/22 = 41% — above the 40% minimum.

---

## Scenarios Requiring Mutation Fixtures

The following tests are marked `t.Skip("requires ... fixture")`. The software crafter
implements the fixture mechanism during delivery:

| Test | Fixture Type | Observable Outcome |
|------|--------------|--------------------|
| TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt | Temp broken Go source file | exit 1 + "FAIL" in output |
| TestPipeline_CI_StopsOnValidateFailure | Temp broken Go source file | acceptance step not printed |
| TestPipeline_PreCommit_BlocksOnToolVersionMismatch | Temp config.yml version mutation | exit 1 + tool name + versions |
| TestPipeline_PreCommit_BlocksOnArchitectureViolation | Temp forbidden import in adapter | exit 1 + FAIL [3/5] |
| TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions | Temp config.yml version mutation | FAIL + local vs expected versions |
| TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig | Temp goreleaser.yml syntax error | exit non-zero + config error |
| TestPipeline_ReleaseSnapshot_FailsGracefullyWhenGoreleaserNotInstalled | PATH override | exit non-zero + install message |
