package domain

import "time"

// TaskStatus is the lifecycle state of a Task.
type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in-progress"
	StatusDone       TaskStatus = "done"
)

// Task is the core work item tracked on the board.
type Task struct {
	ID          string
	Title       string
	Status      TaskStatus
	Priority    string
	Due         *time.Time
	Assignee    string
	Description string
	CreatedBy   string
}

// Column represents a vertical lane on the board.
type Column struct {
	Name  string
	Label string
}

// Board groups columns and the tasks within them.
type Board struct {
	Columns []Column
	Tasks   map[TaskStatus][]Task
}

// Transition records a status change for a task.
type Transition struct {
	TaskID     string
	FromStatus TaskStatus
	ToStatus   TaskStatus
}

// ValidationError is a typed error returned for business rule violations.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
