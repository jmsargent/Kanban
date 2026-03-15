package domain_test

import (
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
)

// Test Budget: 5 behaviors x 2 = 10 max unit tests
// Behaviors:
// 1. ValidateNewTask accepts valid task (title + future due date)
// 2. ValidateNewTask rejects empty title
// 3. ValidateNewTask rejects past due date
// 4. CanTransitionTo enforces valid/invalid transitions
// 5. IsOverdue returns true only for tasks past due

func TestValidateNewTask_AcceptsValidTitleAndFutureDueDate(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	err := domain.ValidateNewTask("Fix OAuth login bug", &future)
	if err != nil {
		t.Errorf("expected no error for valid task, got: %v", err)
	}
}

func TestValidateNewTask_AcceptsValidTitleWithNilDueDate(t *testing.T) {
	err := domain.ValidateNewTask("Fix OAuth login bug", nil)
	if err != nil {
		t.Errorf("expected no error for valid task without due date, got: %v", err)
	}
}

func TestValidateNewTask_RejectsEmptyTitle(t *testing.T) {
	err := domain.ValidateNewTask("", nil)
	if err == nil {
		t.Fatal("expected error for empty title, got nil")
	}
	var validationErr *domain.ValidationError
	if !asValidationError(err, &validationErr) {
		t.Errorf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}

func TestValidateNewTask_RejectsPastDueDate(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	err := domain.ValidateNewTask("Valid title", &past)
	if err == nil {
		t.Fatal("expected error for past due date, got nil")
	}
	var validationErr *domain.ValidationError
	if !asValidationError(err, &validationErr) {
		t.Errorf("expected *domain.ValidationError, got %T: %v", err, err)
	}
}
