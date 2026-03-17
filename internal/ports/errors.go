package ports

import "errors"

// Sentinel errors returned by port implementations.
// Callers use errors.Is to distinguish them.
var (
	ErrTaskNotFound               = errors.New("task not found")
	ErrNotInitialised             = errors.New("kanban not initialised in this repository")
	ErrNotGitRepo                 = errors.New("not a git repository")
	ErrInvalidTransition          = errors.New("invalid status transition")
	ErrInvalidInput               = errors.New("invalid input")
	ErrGitIdentityNotConfigured   = errors.New("git identity not configured")
)
