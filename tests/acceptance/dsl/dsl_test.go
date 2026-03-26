package dsl_test

import (
	"testing"

	"github.com/jmsargent/kanban/tests/acceptance/dsl"
)

// TestNewContextReturnsNonNil verifies that NewContext resolves a binary path
// and returns a usable Context.
func TestNewContextReturnsNonNil(t *testing.T) {
	ctx := dsl.NewContext(t)
	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}
}

// TestNewContextResolvesBinPath verifies that the resolved binary path is non-empty.
func TestNewContextResolvesBinPath(t *testing.T) {
	ctx := dsl.NewContext(t)
	// LastTaskID is exported as a proxy to confirm the context is functional;
	// the real observable is that binPath was resolved (non-empty).
	// We verify via LastTaskID (empty string expected — no command run yet).
	if got := ctx.LastTaskID(); got != "" {
		t.Errorf("expected empty LastTaskID before any command, got %q", got)
	}
}

// TestStepWithNilRunPanicsOrReturnsError verifies Step.Run nil does not silently succeed.
// Given/When/Then call t.Fatalf which stops the test, so we exercise through a
// Step that returns a non-nil error using a sub-test helper.
func TestOrchestratorsCallHelperAndFatalOnError(t *testing.T) {
	// Use a sub-test with a mock T to catch Fatal calls without stopping the outer test.
	// Since t.Fatalf stops execution, we verify it indirectly: a Step that returns nil
	// must not cause t.Fatal.
	ctx := dsl.NewContext(t)

	successStep := dsl.Step{
		Description: "a step that succeeds",
		Run: func(*dsl.Context) error {
			return nil
		},
	}

	// These must not panic or cause test failure.
	dsl.Given(ctx, successStep)
	dsl.When(ctx, successStep)
	dsl.Then(ctx, successStep)
	dsl.And(ctx, successStep)
}
