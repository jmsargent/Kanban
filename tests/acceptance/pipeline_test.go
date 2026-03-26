package acceptance

// Pipeline acceptance tests — testable-pipeline feature
//
// Testability tiers:
//   Tier 1 (fully local): Makefile targets exist, pre-commit hook content,
//                         check-versions.sh exits correctly, go-arch-lint runs.
//   Tier 2 (local build): make release-snapshot builds dist/ without publishing.
//   Tier 3 (CI-only):     goreleaser cache hit/miss, [skip release] CI behaviour,
//                         actual GitHub Release creation.
//
// Tier 3 tests are marked t.Skip("requires CircleCI context: ...").
//
// Running order follows Outside-In TDD: only the first test in each group is
// enabled. Mark subsequent tests with t.Skip to enable them one at a time as
// implementation progresses.
//
// Walking skeleton:
//   TestPipeline_ValidateTarget_PassesWhenAllChecksGreen
//
// Enable sequence (implementation order):
//   1. TestPipeline_ValidateTarget_PassesWhenAllChecksGreen        ← walking skeleton, enable first
//   2. TestPipeline_ValidateTarget_HelpTargetListsAllTargets
//   3. TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt
//   4. TestPipeline_Acceptance_FailsWhenBinaryNotBuilt
//   5. TestPipeline_Acceptance_PassesWithBuiltBinary
//   6. TestPipeline_CI_StopsOnValidateFailure
//   7. TestPipeline_PreCommit_ArchLintStepIsActive
//   8. TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest
//   9. TestPipeline_CheckVersions_ReportsAllToolsMatch
//   10. TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions
//   11. TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing  (Tier 2)
//   12. TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig    (Tier 2)
//   13. TestPipeline_SkipRelease_CIOnlyDocumented                     (Tier 3, skip)
//   14. TestPipeline_GoreleaserCache_CIOnlyDocumented                 (Tier 3, skip)

import (
	"testing"

	dsl "github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// ============================================================
// Walking Skeleton — US-TP-02
// ============================================================

// TestPipeline_ValidateTarget_PassesWhenAllChecksGreen is the walking skeleton.
// It validates the minimum observable developer outcome: running `make validate`
// in a correctly configured project completes successfully and reports each step.
//
// This is the first test to enable. All others begin as t.Skip.
func TestPipeline_ValidateTarget_PassesWhenAllChecksGreen(t *testing.T) {
	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineGiven(pc, dsl.AllToolVersionsMatchPipeline())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("validate"))
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[0/4] check-versions"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[1/4]"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[2/4]"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[3/4] go-arch-lint"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[4/4]"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("PASS"))
}

// ============================================================
// Makefile target existence — US-TP-02
// ============================================================

// TestPipeline_Makefile_ContainsRequiredTargets asserts the Makefile declares
// all targets the developer needs to mirror CI jobs locally.
func TestPipeline_Makefile_ContainsRequiredTargets(t *testing.T) {
	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineThen(pc, dsl.MakefileContainsTarget("pre-commit"))
	dsl.PipelineAnd(pc, dsl.MakefileContainsTarget("release-snapshot"))
	dsl.PipelineAnd(pc, dsl.MakefileContainsTarget("tag-dry"))
	dsl.PipelineAnd(pc, dsl.MakefileContainsTarget("help"))
}

// TestPipeline_ValidateTarget_HelpTargetListsAllTargets asserts `make help`
// prints a target list that fits in 80 columns.
func TestPipeline_ValidateTarget_HelpTargetListsAllTargets(t *testing.T) {
	t.Skip("enable after TestPipeline_Makefile_ContainsRequiredTargets passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("help"))
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("validate"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("acceptance"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("release-snapshot"))
}

// ============================================================
// make validate — step sequence — US-TP-02
// ============================================================

// TestPipeline_ValidateTarget_ReportsEachStepLabel asserts `make validate`
// prints the expected step labels in order, mirroring the CI job steps.
func TestPipeline_ValidateTarget_ReportsEachStepLabel(t *testing.T) {
	t.Skip("enable after walking skeleton passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineGiven(pc, dsl.AllToolVersionsMatchPipeline())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("validate"))
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	// Step labels must appear — their presence proves commands ran in declared order
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[0/4] check-versions"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[1/4]"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[2/4]"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[3/4] go-arch-lint"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("[4/4]"))
}

// TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt asserts that when
// `make validate` encounters a build failure, it exits non-zero and reports
// which step failed — so the developer knows exactly what to fix.
func TestPipeline_ValidateTarget_FailsWhenBinaryCannotBeBuilt(t *testing.T) {
	t.Skip("enable after TestPipeline_ValidateTarget_ReportsEachStepLabel passes")

	// This scenario requires injecting a deliberate build error. Implemented by
	// the software crafter when the make validate step-label mechanism is in place.
	// The observable behaviour under test: exit non-zero + "FAIL" in output.
	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	// Precondition: a deliberately broken build is outside test scope here.
	// The software crafter implements this test body using a temp source file mutation.
	_ = pc
	t.Skip("requires source mutation fixture — implement in delivery")
}

// ============================================================
// make acceptance — US-TP-02
// ============================================================

// TestPipeline_Acceptance_FailsWhenBinaryNotBuilt asserts that `make acceptance`
// exits non-zero with an actionable message when the kanban binary does not exist.
func TestPipeline_Acceptance_FailsWhenBinaryNotBuilt(t *testing.T) {
	t.Skip("enable after TestPipeline_ValidateTarget_PassesWhenAllChecksGreen passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineGiven(pc, dsl.TheKanbanBinaryIsAbsent())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("acceptance"))
	dsl.PipelineThen(pc, dsl.PipelineExitsWithFailure())
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("make validate"))
}

// TestPipeline_Acceptance_PassesWithBuiltBinary asserts that `make acceptance`
// exits 0 when the binary is present and all acceptance tests pass.
func TestPipeline_Acceptance_PassesWithBuiltBinary(t *testing.T) {
	t.Skip("enable after TestPipeline_Acceptance_FailsWhenBinaryNotBuilt passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineGiven(pc, dsl.TheKanbanBinaryIsBuilt())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("acceptance"))
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
}

// ============================================================
// make ci — US-TP-02
// ============================================================

// TestPipeline_CI_StopsOnValidateFailure asserts that `make ci` does not run
// the acceptance tests when `make validate` fails — protecting the developer
// from a misleading green acceptance result on a broken build.
func TestPipeline_CI_StopsOnValidateFailure(t *testing.T) {
	t.Skip("enable after TestPipeline_Acceptance_PassesWithBuiltBinary passes")

	// Observable outcome: acceptance step output does not appear when validate fails.
	// The software crafter implements via source mutation fixture.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires source mutation fixture — implement in delivery")
}

// ============================================================
// Pre-commit hook — US-TP-03
// ============================================================


// TestPipeline_PreCommit_BlocksOnToolVersionMismatch asserts that when a tool
// version does not match cicd/config.yml, the pre-commit hook exits 1 and
// identifies the mismatched tool — before any test or lint steps run.
func TestPipeline_PreCommit_BlocksOnToolVersionMismatch(t *testing.T) {
	t.Skip("enable after TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest passes")

	// Observable outcome: exit 1 + tool name + expected version in output.
	// The software crafter implements via a temporary config.yml mutation that
	// sets an impossible expected version for one tool.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires cicd/config.yml mutation fixture — implement in delivery")
}

// TestPipeline_PreCommit_BlocksOnArchitectureViolation asserts the pre-commit
// hook exits 1 and identifies the forbidden import when go-arch-lint detects
// an architecture violation.
func TestPipeline_PreCommit_BlocksOnArchitectureViolation(t *testing.T) {
	t.Skip("enable after TestPipeline_PreCommit_BlocksOnToolVersionMismatch passes")

	// Observable outcome: exit 1 + FAIL [3/5] go-arch-lint in output.
	// The software crafter implements via a temp source file with a forbidden import.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires source mutation fixture — implement in delivery")
}

// TestPipeline_PreCommit_PassesWhenGatesParity asserts that when all tools match
// the pipeline and no violations exist, the pre-commit hook exits 0.
func TestPipeline_PreCommit_PassesWhenGatesParity(t *testing.T) {
	t.Skip("enable after TestPipeline_PreCommit_BlocksOnArchitectureViolation passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.ThePreCommitHookScript())
	dsl.PipelineGiven(pc, dsl.AllToolVersionsMatchPipeline())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsPreCommitHook())
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("PASS [0/5] check-versions"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("PASS [3/5] go-arch-lint"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("all quality gates passed"))
}

// ============================================================
// cicd/check-versions.sh — US-TP-03
// ============================================================

// TestPipeline_CheckVersions_ReportsAllToolsMatch asserts that check-versions.sh
// exits 0 and prints an OK line for each tool when local versions match
// the pipeline parameters in cicd/config.yml.
func TestPipeline_CheckVersions_ReportsAllToolsMatch(t *testing.T) {
	t.Skip("enable after TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest passes")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheCheckVersionsScript())
	dsl.PipelineWhen(pc, dsl.DeveloperRunsCheckVersions())
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("OK  go"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("OK  golangci-lint"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("OK  go-arch-lint"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("All"))
	dsl.PipelineAnd(pc, dsl.PipelineOutputContains("match"))
}

// TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions asserts that when
// a mismatch exists, check-versions.sh exits 1 and the output names the tool and
// both the local version and the pipeline-expected version.
func TestPipeline_CheckVersions_OutputIdentifiesExpectedVersions(t *testing.T) {
	t.Skip("enable after TestPipeline_CheckVersions_ReportsAllToolsMatch passes")

	// Observable outcome: "FAIL golangci-lint: local=v2.10.0  pipeline=v2.11.3"
	// The software crafter implements via a temporary config.yml parameter mutation.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires cicd/config.yml mutation fixture — implement in delivery")
}

// ============================================================
// make release-snapshot — US-TP-01 (Tier 2)
// ============================================================

// TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing asserts that
// `make release-snapshot` builds all cross-compile targets to dist/ and exits 0,
// without requiring GITHUB_TOKEN or creating any git tag or GitHub Release.
func TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing(t *testing.T) {
	t.Skip("enable after TestPipeline_ValidateTarget_PassesWhenAllChecksGreen passes — Tier 2: requires goreleaser installed locally")

	pc := dsl.NewPipelineContext(t)
	dsl.PipelineGiven(pc, dsl.TheProjectMakefile())
	dsl.PipelineGiven(pc, dsl.MakefileContainsTarget("release-snapshot"))
	dsl.PipelineWhen(pc, dsl.DeveloperRunsMakeTarget("release-snapshot"))
	dsl.PipelineThen(pc, dsl.PipelineExitsSuccessfully())
	dsl.PipelineAnd(pc, dsl.PipelineOutputDoesNotContain("GITHUB_TOKEN"))
	// Goreleaser snapshot produces binaries in dist/ — presence of dist/ signals success
	dsl.PipelineAnd(pc, dsl.PipelineStep{
		Description: "dist/ directory contains build artifacts",
		Run: func(ppc *dsl.PipelineContext) error {
			return nil // software crafter implements dist/ content assertion
		},
	})
}

// TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig asserts that when
// cicd/goreleaser.yml contains a syntax error, `make release-snapshot` exits
// non-zero and the output identifies the config error before any build runs.
func TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig(t *testing.T) {
	t.Skip("enable after TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing passes — Tier 2: requires goreleaser installed locally")

	// Observable outcome: exit non-zero + config validation error in output.
	// The software crafter implements via a temp goreleaser.yml mutation.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires cicd/goreleaser.yml mutation fixture — implement in delivery")
}

// TestPipeline_ReleaseSnapshot_FailsGracefullyWhenGoreleaserNotInstalled asserts
// that `make release-snapshot` fails with an actionable installation message when
// goreleaser is not in PATH, rather than a confusing "command not found" error.
func TestPipeline_ReleaseSnapshot_FailsGracefullyWhenGoreleaserNotInstalled(t *testing.T) {
	t.Skip("enable after TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig passes — Tier 2: requires PATH manipulation")

	// Observable outcome: exit non-zero + "goreleaser" + install instructions in output.
	// The software crafter implements via PATH override excluding goreleaser.
	_ = dsl.NewPipelineContext(t)
	t.Skip("requires PATH mutation fixture — implement in delivery")
}

// ============================================================
// [skip release] convention — US-TP-04 (Tier 3 — CI-only)
// ============================================================

// TestPipeline_SkipRelease_CIOnlyDocumented documents the expected CircleCI
// behaviour when a commit message contains [skip release].
//
// Tier 3: This behaviour cannot be tested locally. It requires a live CircleCI
// run. This test exists as executable documentation of the expected observable
// outcome in CI.
//
// Expected CI outcome:
//   - validate-and-build job: runs and passes
//   - acceptance job: runs and passes
//   - tag job: runs, detects [skip release], exits 0 without creating a tag
//   - release job: runs, detects [skip release], exits 0 without publishing
//   - No git tag created in the repository
//   - No GitHub Release published
func TestPipeline_SkipRelease_CIOnlyDocumented(t *testing.T) {
	t.Skip("requires CircleCI context: [skip release] guard in tag/release job scripts — cannot be asserted locally")
}

// TestPipeline_SkipRelease_NormalCommitRunsFullPipeline documents that a commit
// without [skip release] runs all four CI jobs (existing behaviour preserved).
//
// Tier 3: CI-only observable behaviour.
func TestPipeline_SkipRelease_NormalCommitRunsFullPipeline(t *testing.T) {
	t.Skip("requires CircleCI context: full pipeline run — cannot be asserted locally")
}

// TestPipeline_SkipRelease_CaseInsensitiveAndPositionIndependent documents that
// [SKIP RELEASE], [skip release] at end of message, and mixed case all suppress
// the tag and release jobs.
//
// Tier 3: CI-only observable behaviour.
func TestPipeline_SkipRelease_CaseInsensitiveAndPositionIndependent(t *testing.T) {
	t.Skip("requires CircleCI context: grep -qi pattern in job script — cannot be asserted locally")
}

// ============================================================
// goreleaser CI caching — US-TP-01 (Tier 3 — CI-only)
// ============================================================

// TestPipeline_GoreleaserCache_CIOnlyDocumented documents the expected CircleCI
// cache behaviour for the goreleaser install-goreleaser command.
//
// Tier 3: Cache hit/miss behaviour is only observable in CircleCI.
//
// Expected CI outcome on cache hit:
//   - restore_cache step finds key "goreleaser-<version>"
//   - install guard "if [ ! -f ... ]" skips installation
//   - release job proceeds without downloading goreleaser
//
// Expected CI outcome on version bump (cache miss):
//   - restore_cache step finds no matching key
//   - goreleaser installed via go install at new version
//   - new binary saved to cache key "goreleaser-<new-version>"
func TestPipeline_GoreleaserCache_CIOnlyDocumented(t *testing.T) {
	t.Skip("requires CircleCI context: cache restore/save behaviour — cannot be asserted locally")
}
