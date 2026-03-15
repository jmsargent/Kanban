package domain

import "time"

// ValidateNewTask enforces creation invariants: title must be non-empty and
// due date (when provided) must not be in the past.
func ValidateNewTask(title string, due *time.Time) error {
	if title == "" {
		return &ValidationError{
			Field:   "title",
			Message: "Task title is required",
		}
	}
	if due != nil && due.Before(time.Now().Truncate(24*time.Hour)) {
		return &ValidationError{
			Field:   "due",
			Message: "Due date must be today or in the future",
		}
	}
	return nil
}

// CanTransitionTo reports whether moving a task from one status to another
// is permitted by board rules.
// Allowed: todo->in-progress, in-progress->done.
// All other transitions are invalid.
func CanTransitionTo(from, to TaskStatus) bool {
	switch from {
	case StatusTodo:
		return to == StatusInProgress
	case StatusInProgress:
		return to == StatusDone
	default:
		return false
	}
}

// IsOverdue reports true when the task has a due date that is strictly
// before the given reference time (now).
func IsOverdue(task Task, now time.Time) bool {
	if task.Due == nil {
		return false
	}
	return task.Due.Before(now)
}
