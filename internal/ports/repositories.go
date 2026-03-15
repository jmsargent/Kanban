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
