package ports

import "github.com/kanban-tasks/kanban/internal/domain"

// TaskRepository is the driven port for persisting and retrieving Task aggregates.
// All implementations must be wired at the composition root; no concrete types
// may be imported into internal/ports or internal/domain.
type TaskRepository interface {
	// FindByID retrieves a task by its ID within the given repository root.
	// Returns ErrTaskNotFound when no task matches.
	FindByID(repoRoot, taskID string) (domain.Task, error)

	// Save persists the task, creating or updating as needed.
	Save(repoRoot string, task domain.Task) error

	// ListAll returns every task stored under repoRoot.
	ListAll(repoRoot string) ([]domain.Task, error)

	// Update overwrites the persisted task. Returns ErrTaskNotFound when no
	// file exists for the task ID.
	Update(repoRoot string, task domain.Task) error

	// Delete removes the task identified by taskID.
	Delete(repoRoot, taskID string) error

	// NextID returns the next available unique task identifier.
	NextID(repoRoot string) (string, error)
}

// Config holds the board configuration persisted alongside tasks.
type Config struct {
	Columns       []domain.Column
	CITaskPattern string
}

// ConfigRepository is the driven port for reading and writing board configuration.
type ConfigRepository interface {
	// Read loads the configuration for the given repository root.
	// Returns ErrNotInitialised when no configuration exists.
	Read(repoRoot string) (Config, error)

	// Write stores the configuration for the given repository root.
	Write(repoRoot string, config Config) error
}

// EditSnapshot is the editable subset of task fields for the edit workflow.
type EditSnapshot struct {
	Title       string
	Priority    string
	Due         string
	Assignee    string
	Description string
}

// EditFilePort is the driven port for reading and writing temp edit files.
type EditFilePort interface {
	// WriteTemp writes editable fields of the task to a temporary YAML file.
	// Returns the path to the temp file.
	WriteTemp(task domain.Task) (string, error)

	// WriteTempNew writes a blank task template to a temporary file for use by
	// the interactive "kanban new" flow. Returns the path to the temp file.
	WriteTempNew() (string, error)

	// ReadTemp reads the YAML temp file at path and returns the snapshot.
	ReadTemp(path string) (EditSnapshot, error)
}
