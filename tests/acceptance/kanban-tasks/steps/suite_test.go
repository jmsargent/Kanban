package steps

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
)

// TestFeatures is the entry point for the godog acceptance test suite.
// It discovers and runs all .feature files in the features/ directory
// relative to this package.
//
// Usage:
//
//	go test ./tests/acceptance/kanban-tasks/steps/
//
// To specify a compiled binary:
//
//	KANBAN_BIN=/path/to/kanban go test ./tests/acceptance/kanban-tasks/steps/
//
// To run a specific feature file:
//
//	go test ./tests/acceptance/kanban-tasks/steps/ --godog.paths=../walking-skeleton.feature
//
// The compiled binary path defaults to ../../bin/kanban relative to this
// directory. Set the KANBAN_BIN environment variable to override.
func TestFeatures(t *testing.T) {
	paths := []string{"../"}

	// Allow overriding feature file paths via environment for CI
	if envPaths := os.Getenv("GODOG_PATHS"); envPaths != "" {
		paths = []string{envPaths}
	}

	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    paths,
			TestingT: t,
			NoColors: true,
			// Run only the first (non-skipped) scenario to support one-at-a-time TDD.
			// Remove or increase this once additional scenarios are enabled.
			Tags: "~@skip",
		},
	}

	if suite.Run() != 0 {
		t.Fatal("godog acceptance test suite failed")
	}
}
