package backend

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

import . "github.com/jmsargent/kanban/tests/acceptance/backend/dsl"

// TestScaffoldCompiles verifies all driver and DSL types can be constructed.
// This is a compilation smoke test for step 01-01 infrastructure scaffolding.
func TestScaffoldCompiles(t *testing.T) {
	// Verify WebContext construction
	ctx := NewWebContext(t)
	if ctx == nil {
		t.Fatal("expected non-nil WebContext")
	}

	// Verify ServerDriver construction
	sd := driver.NewServerDriver(t)
	if sd == nil {
		t.Fatal("expected non-nil ServerDriver")
	}

	// Verify HTTPDriver construction
	hd := driver.NewHTTPDriver("http://localhost:0")
	if hd == nil {
		t.Fatal("expected non-nil HTTPDriver")
	}

	// Verify RepoDriver construction
	rd := driver.NewRepoDriver(t)
	if rd == nil {
		t.Fatal("expected non-nil RepoDriver")
	}
	if rd.RepoDir() == "" {
		t.Fatal("expected non-empty RepoDir from RepoDriver")
	}

	// Verify GitHubStubDriver construction
	gd := driver.NewGitHubStubDriver()
	if gd == nil {
		t.Fatal("expected non-nil GitHubStubDriver")
	}
	defer gd.Close()
	if gd.URL() == "" {
		t.Fatal("expected non-empty URL from GitHubStubDriver")
	}

	// Verify DSL step types exist
	step := Step{}
	_ = step.Description
	_ = step.Run

	// Verify Given/When/Then/And compile with WebContext
	_ = Given
	_ = When
	_ = Then
	_ = And
}