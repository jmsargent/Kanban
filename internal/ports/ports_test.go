package ports_test

import (
	"errors"
	"testing"

	"github.com/jmsargent/kanban/internal/ports"
)

// Test budget: 1 distinct behavior (sentinel errors are defined and distinct) x 2 = 2 unit tests.

func TestSentinelErrors_AreDistinct(t *testing.T) {
	sentinels := []error{
		ports.ErrTaskNotFound,
		ports.ErrNotInitialised,
		ports.ErrNotGitRepo,
		ports.ErrInvalidTransition,
		ports.ErrInvalidInput,
	}
	for i, a := range sentinels {
		for j, b := range sentinels {
			if i != j && errors.Is(a, b) {
				t.Errorf("sentinel errors %d and %d are not distinct", i, j)
			}
		}
	}
}

func TestSentinelErrors_AreNonNil(t *testing.T) {
	for _, err := range []error{ports.ErrTaskNotFound, ports.ErrNotInitialised, ports.ErrNotGitRepo, ports.ErrInvalidTransition, ports.ErrInvalidInput} {
		if err == nil {
			t.Error("expected non-nil sentinel error")
		}
	}
}
