package ports

import (
	"errors"

	"github.com/jmsargent/kanban/internal/domain"
)

// GitHubAPIPort is the driven port for fetching kanban task data from the
// GitHub REST API. Implementations are injected at the composition root.
type GitHubAPIPort interface {
	// ListTasks fetches all task files from the .kanban/tasks directory of the
	// given public repository and returns them as domain.Task values.
	ListTasks(owner, repo string) ([]domain.Task, error)
}

// Sentinel errors returned by GitHubAPIPort implementations.
var (
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrNoBoardFound       = errors.New("no kanban board in repository")
	ErrRateLimitExceeded  = errors.New("GitHub API rate limit exceeded")
)