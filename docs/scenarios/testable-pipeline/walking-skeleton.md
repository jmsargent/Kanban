# Walking Skeleton — testable-pipeline

**Feature**: testable-pipeline
**Wave**: DISTILL
**Date**: 2026-03-18

---

## The Walking Skeleton

**Test**: `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen`
**File**: `tests/acceptance/pipeline_test.go`

### User goal it answers

"Can a developer run `make validate` and receive a clear green/red signal before pushing to main?"

This is the minimum observable slice: a Makefile exists, it runs the same steps as CI's `validate-and-build` job, and it exits 0 with step labels that confirm each gate passed. The developer can immediately stop using live CircleCI pushes as the test environment for the most common failure class.

### Why this is the walking skeleton

Per the DISCUSS wave story map, the walking skeleton connects all five backbone activities:

| Activity | Walking Skeleton Task |
|----------|-----------------------|
| Detect | Developer knows a pipeline change is needed (pre-existing) |
| Reproduce Locally | `make validate` runs the same commands as CI validate-and-build |
| Fix | Developer edits config locally |
| Validate Locally | `make validate` gives green/red signal in < 2 minutes |
| Push with Confidence | Pre-commit gate includes go-arch-lint (restored parity) |

The pre-commit parity (US-TP-03) is the companion half of the skeleton. The `TestPipeline_PreCommit_PassesWhenGatesParity` test documents that outcome, but starts as `t.Skip` per the one-at-a-time implementation rule.

### Stakeholder demo statement

"Watch: I run `make validate` in under 2 minutes and see every quality gate pass locally — the same gates CircleCI runs. I do not need to push to main to find out."

### Litmus test (from test-design-mandates)

1. Title describes user goal ("ValidateTarget_PassesWhenAllChecksGreen") — not technical flow. PASS
2. Given/When describe developer actions, not system state setup. PASS
3. Then describes developer observations (step labels, exit code, "PASS" output) — not internal side effects. PASS
4. Non-technical stakeholder can confirm: "yes, that is what the developer needs." PASS

---

## First Test to Enable

`TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` — the only test without `t.Skip` in `pipeline_test.go`.

**What it requires to pass**:
1. `Makefile` exists at the project root
2. `make validate` target runs the command sequence in order
3. Each step prints a `[N/5]` label and a `PASS` marker
4. All tools match the versions in `cicd/config.yml` (pre-condition checked by `AllToolVersionsMatchPipeline`)
5. The go-arch-lint step is labeled `[3/5]` (confirming US-TP-03 is in place)

**Expected failure reason before implementation**: `make: *** No rule to make target 'validate'. Stop.` — a Makefile does not exist yet.

---

## Enable Sequence

Enable tests one at a time as each implementation increment is committed:

```
1. TestPipeline_ValidateTarget_PassesWhenAllChecksGreen      ← walking skeleton (enabled now)
2. TestPipeline_Makefile_ContainsRequiredTargets
3. TestPipeline_ValidateTarget_HelpTargetListsAllTargets
4. TestPipeline_ValidateTarget_ReportsEachStepLabel
5. TestPipeline_Acceptance_FailsWhenBinaryNotBuilt
6. TestPipeline_Acceptance_PassesWithBuiltBinary
7. TestPipeline_PreCommit_ArchLintStepIsActive
8. TestPipeline_PreCommit_CheckVersionsRunsBeforeGoTest
9. TestPipeline_PreCommit_PassesWhenGatesParity
10. TestPipeline_CheckVersions_ReportsAllToolsMatch
11. TestPipeline_ReleaseSnapshot_BuildsArtifactsWithoutPublishing  (Tier 2 — goreleaser required)
12. TestPipeline_ReleaseSnapshot_FailsOnInvalidGoreleaserConfig    (Tier 2)
```

Tests 13–14 (`SkipRelease_*`, `GoreleaserCache_*`) remain `t.Skip("requires CircleCI context")` permanently.
