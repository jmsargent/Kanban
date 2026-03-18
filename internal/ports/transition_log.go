package ports

import (
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
)

// CommitEntry represents a single commit in git history for a file.
// It is returned by GitPort.LogFile.
type CommitEntry struct {
	SHA       string
	Timestamp time.Time
	Author    string
	Message   string
}

// TransitionLogRepository is the driven port for reading and appending to
// the transitions audit log. Implementations live in internal/adapters.
type TransitionLogRepository interface {
	// Append records a new transition entry for the task.
	Append(repoRoot string, entry domain.TransitionEntry) error

	// LatestStatus returns the most recent TaskStatus recorded for taskID.
	// Returns ErrTaskNotFound when no entries exist for the task.
	LatestStatus(repoRoot, taskID string) (domain.TaskStatus, error)

	// History returns all recorded transitions for taskID, oldest first.
	// Returns an empty slice (not an error) when no entries exist.
	History(repoRoot, taskID string) ([]domain.TransitionEntry, error)
}
