package domain_test

import (
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
)

// CanTransitionTo has a stable public interface and is a standalone transition
// algorithm — exception to the no-direct-domain-test rule.

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 6)
// Behaviors:
//   1. CanTransitionTo allows valid transitions (todo→in-progress, in-progress→done)
//   2. CanTransitionTo rejects invalid transitions
//   3. IsOverdue: nil due date is never overdue; past due date is overdue; future is not

func TestCanTransitionTo_AllowsTodoToInProgress(t *testing.T) {
	if !domain.CanTransitionTo(domain.StatusTodo, domain.StatusInProgress) {
		t.Error("expected todo->in-progress to be allowed")
	}
}

func TestCanTransitionTo_AllowsInProgressToDone(t *testing.T) {
	if !domain.CanTransitionTo(domain.StatusInProgress, domain.StatusDone) {
		t.Error("expected in-progress->done to be allowed")
	}
}

func TestCanTransitionTo_RejectsInvalidTransitions(t *testing.T) {
	cases := []struct {
		from domain.TaskStatus
		to   domain.TaskStatus
	}{
		{domain.StatusTodo, domain.StatusDone},
		{domain.StatusInProgress, domain.StatusTodo},
		{domain.StatusDone, domain.StatusTodo},
		{domain.StatusDone, domain.StatusInProgress},
		{domain.StatusDone, domain.StatusDone},
	}
	for _, tc := range cases {
		if domain.CanTransitionTo(tc.from, tc.to) {
			t.Errorf("expected transition %s->%s to be rejected", tc.from, tc.to)
		}
	}
}

func TestIsOverdue_ReturnsFalse_WhenNoDueDateSet(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "No due date", Status: domain.StatusTodo, Due: nil}
	if domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue false when Due is nil")
	}
}

func TestIsOverdue_ReturnsTrue_WhenDueDateIsInThePast(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{ID: "TASK-002", Title: "Overdue", Status: domain.StatusTodo, Due: &past}
	if !domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue true for task with past due date")
	}
}

func TestIsOverdue_ReturnsFalse_WhenDueDateIsInTheFuture(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	task := domain.Task{ID: "TASK-003", Title: "Future", Status: domain.StatusTodo, Due: &future}
	if domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue false for task with future due date")
	}
}
