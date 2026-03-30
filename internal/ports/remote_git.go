package ports

// RemoteGitPort is the driven port for remote git operations.
// Implementations live in internal/adapters/git.
type RemoteGitPort interface {
	// Pull fetches and merges changes from the configured remote (origin)
	// into the working repository at repoDir.
	Pull(repoDir string) error
}
