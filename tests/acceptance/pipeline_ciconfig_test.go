package acceptance

// pipeline_ciconfig_test.go — structural tests for cicd/config.yml
//
// Step 01-04: goreleaser caching and [skip release] shell guard
//
// These tests assert the structure of cicd/config.yml directly.
// They are enabled (not t.Skip) and must pass after GREEN.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	dsl "github.com/kanban-tasks/kanban/tests/acceptance/dsl"
)

// readCIConfig reads cicd/config.yml and returns its content.
func readCIConfig(t *testing.T) string {
	t.Helper()
	root, err := dsl.ProjectRoot()
	if err != nil {
		t.Fatalf("locate project root: %v", err)
	}
	path := filepath.Join(root, "cicd", "config.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cicd/config.yml: %v", err)
	}
	return string(content)
}

// TestPipeline_CIConfig_ContainsInstallGoreleaserCommand asserts cicd/config.yml
// declares a reusable install-goreleaser command — the observable outcome is that
// goreleaser is installed via go install (not curl) with restore_cache/save_cache.
func TestPipeline_CIConfig_ContainsInstallGoreleaserCommand(t *testing.T) {
	config := readCIConfig(t)

	if !strings.Contains(config, "install-goreleaser:") {
		t.Error("cicd/config.yml does not contain the reusable 'install-goreleaser:' command")
	}

	if !strings.Contains(config, "go install github.com/goreleaser/goreleaser") {
		t.Error("cicd/config.yml does not install goreleaser via 'go install github.com/goreleaser/goreleaser'")
	}
}

// TestPipeline_CIConfig_GoreleaserCommandUsesCache asserts the install-goreleaser
// command uses restore_cache and save_cache with a version-keyed cache key.
func TestPipeline_CIConfig_GoreleaserCommandUsesCache(t *testing.T) {
	config := readCIConfig(t)

	if !strings.Contains(config, "goreleaser-<< pipeline.parameters.goreleaser-version >>") {
		t.Error("cicd/config.yml does not contain version-keyed goreleaser cache key")
	}
}

// TestPipeline_CIConfig_TagJobHasSkipReleaseGuard asserts the tag job contains
// the [skip release] shell guard so CI skips tagging when the commit message
// contains [skip release].
func TestPipeline_CIConfig_TagJobHasSkipReleaseGuard(t *testing.T) {
	config := readCIConfig(t)

	if !strings.Contains(config, "[skip release]") {
		t.Error("cicd/config.yml does not contain '[skip release]' guard")
	}

	// grep -qi is the case-insensitive, position-independent pattern required by ADR-009.
	if !strings.Contains(config, "grep -qi") {
		t.Error("cicd/config.yml does not use 'grep -qi' for case-insensitive [skip release] detection")
	}
}

// TestPipeline_CIConfig_ReleaseJobUsesInstallGoreleaserCommand asserts the release
// job invokes the install-goreleaser reusable command instead of the inline curl|bash.
func TestPipeline_CIConfig_ReleaseJobUsesInstallGoreleaserCommand(t *testing.T) {
	config := readCIConfig(t)

	if strings.Contains(config, "curl -sfL https://goreleaser.com/static/run") {
		t.Error("cicd/config.yml still uses the inline curl|bash goreleaser install — should use install-goreleaser command")
	}

	if !strings.Contains(config, "- install-goreleaser") {
		t.Error("cicd/config.yml release job does not invoke '- install-goreleaser' command")
	}
}
