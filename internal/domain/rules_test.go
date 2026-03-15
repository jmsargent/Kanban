package domain_test

import (
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
)

func TestCanTransitionTo_AllowsValidTransitions(t *testing.T) {
	validTransitions := []struct {
		from domain.TaskStatus
		to   domain.TaskStatus
	}{
		{domain.StatusTodo, domain.StatusInProgress},
		{domain.StatusInProgress, domain.StatusDone},
	}
	for _, tc := range validTransitions {
		if !domain.CanTransitionTo(tc.from, tc.to) {
			t.Errorf("expected transition %s->%s to be allowed", tc.from, tc.to)
		}
	}
}

func TestCanTransitionTo_RejectsInvalidTransitions(t *testing.T) {
	invalidTransitions := []struct {
		from domain.TaskStatus
		to   domain.TaskStatus
	}{
		{domain.StatusTodo, domain.StatusDone},
		{domain.StatusInProgress, domain.StatusTodo},
		{domain.StatusDone, domain.StatusTodo},
		{domain.StatusDone, domain.StatusInProgress},
		{domain.StatusDone, domain.StatusDone},
	}
	for _, tc := range invalidTransitions {
		if domain.CanTransitionTo(tc.from, tc.to) {
			t.Errorf("expected transition %s->%s to be rejected", tc.from, tc.to)
		}
	}
}

func TestIsOverdue_ReturnsTrueForTasksPastDue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{
		ID:     "TASK-001",
		Title:  "Overdue task",
		Status: domain.StatusTodo,
		Due:    &past,
	}
	if !domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue to return true for task with past due date")
	}
}

func TestIsOverdue_ReturnsFalseForTasksNotYetDue(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	task := domain.Task{
		ID:     "TASK-002",
		Title:  "Future task",
		Status: domain.StatusTodo,
		Due:    &future,
	}
	if domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue to return false for task with future due date")
	}
}

func TestIsOverdue_ReturnsFalseForTaskWithNoDueDate(t *testing.T) {
	task := domain.Task{
		ID:     "TASK-003",
		Title:  "No due date task",
		Status: domain.StatusTodo,
		Due:    nil,
	}
	if domain.IsOverdue(task, time.Now()) {
		t.Error("expected IsOverdue to return false for task with no due date")
	}
}

// asValidationError is a helper to type-assert errors to *domain.ValidationError.
// Defined here (not in task_test.go) to avoid duplicate declarations.
func asValidationError(err error, target **domain.ValidationError) bool {
	if e, ok := err.(*domain.ValidationError); ok {
		*target = e
		return true
	}
	return false
}
