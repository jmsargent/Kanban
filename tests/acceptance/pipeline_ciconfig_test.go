package acceptance

// pipeline_ciconfig_test.go — structural tests for cicd/config.yml
//
// These tests assert the structure of cicd/config.yml directly.
// They are enabled (not t.Skip) and must pass after GREEN.

import (
	"strings"
	"testing"
)

// TestPipeline_CIConfig_ContainsInstallToolsCommand asserts cicd/config.yml
// declares a reusable install-tools command that delegates to the Makefile,
// replacing the previous per-tool install commands.
func TestPipeline_CIConfig_ContainsInstallToolsCommand(t *testing.T) {
	driver := NewPipelineDriver(t)
	config := driver.ReadCIConfig()

	if !strings.Contains(config, "install-tools:") {
		t.Error("cicd/config.yml does not contain the reusable 'install-tools:' command")
	}

	if !strings.Contains(config, "make install-tools") {
		t.Error("cicd/config.yml install-tools command does not delegate to 'make install-tools'")
	}
}

// TestPipeline_CIConfig_InstallToolsCommandUsesChecksumCache asserts the
// install-tools command caches tools using a checksum of .tool-versions,
// so any version change automatically invalidates the CI cache.
func TestPipeline_CIConfig_InstallToolsCommandUsesChecksumCache(t *testing.T) {
	driver := NewPipelineDriver(t)
	config := driver.ReadCIConfig()

	if !strings.Contains(config, `checksum "cicd/tool-versions"`) {
		t.Error("cicd/config.yml does not use '{{ checksum \"cicd/tool-versions\" }}' as the tools cache key")
	}
}

// TestPipeline_CIConfig_TagJobHasSkipReleaseGuard asserts the [skip release]
// shell guard exists in the Makefile ci-tag target (called by the CI tag-and-release job).
// The guard skips tagging when the commit message contains [skip release].
func TestPipeline_CIConfig_TagJobHasSkipReleaseGuard(t *testing.T) {
	driver := NewPipelineDriver(t)
	makefile := driver.ReadMakefile()

	if !strings.Contains(makefile, "[skip release]") {
		t.Error("Makefile ci-tag target does not contain '[skip release]' guard")
	}

	// grep -qi is the case-insensitive, position-independent pattern required by ADR-009.
	if !strings.Contains(makefile, "grep -qi") {
		t.Error("Makefile ci-tag target does not use 'grep -qi' for case-insensitive [skip release] detection")
	}
}

// TestPipeline_CIConfig_JobsUseInstallToolsCommand asserts that CI jobs
// invoke the reusable install-tools command rather than per-tool install commands.
func TestPipeline_CIConfig_JobsUseInstallToolsCommand(t *testing.T) {
	driver := NewPipelineDriver(t)
	config := driver.ReadCIConfig()

	if !strings.Contains(config, "- install-tools") {
		t.Error("cicd/config.yml jobs do not invoke '- install-tools' command")
	}

	// Per-tool commands should no longer appear as job steps.
	if strings.Contains(config, "- install-goreleaser") {
		t.Error("cicd/config.yml still references '- install-goreleaser' as a job step — should use '- install-tools'")
	}
}

func TestCIConfigCommandsShouldBeMakeCommands(t *testing.T){
	driver := NewPipelineDriver(t)
	
	cmds := driver.ReadCommands()
	for _, cmd := range cmds {
		for name, command := range cmd {
			if !strings.HasPrefix(strings.TrimSpace(command), "make") {
				t.Errorf("command %q has non-make line: %q", name, command)
			}
		}
	}
}

func TestPreCommitShouldCallSameMakeTargetsAsPipeline(t *testing.T) {
	driver := NewPipelineDriver(t)

	pipelineTargets := driver.ReadPipelineMakeTargets(
		[]string{"tag-and-release"},
		[]string{"ci-set-env"},
	)
	preCommitSteps := driver.ReadMakeTargetSteps("pre-commit")

	if len(preCommitSteps) == 0 {
		t.Fatal("Makefile pre-commit target has no make sub-steps")
	}

	if len(pipelineTargets) != len(preCommitSteps) {
		t.Fatalf("pipeline has %d make targets %v but pre-commit has %d steps %v",
			len(pipelineTargets), pipelineTargets, len(preCommitSteps), preCommitSteps)
	}

	for i := range pipelineTargets {
		if pipelineTargets[i] != preCommitSteps[i] {
			t.Errorf("step %d: pipeline has %q but pre-commit has %q", i, pipelineTargets[i], preCommitSteps[i])
		}
	}
}