package ports

// RemoteGitPort is the driven port for remote git operations.
// Implementations live in internal/adapters/git.
type RemoteGitPort interface {
	// Pull fetches and merges changes from the configured remote (origin)
	// into the working repository at repoDir.
	Pull(repoDir string) error

	// Add stages the file at path (relative or absolute) in the repository
	// at repoDir, equivalent to `git add <path>`.
	Add(repoDir, path string) error

	// Commit records a new commit in the repository at repoDir with the
	// given message, equivalent to `git commit -m <message>`.
	Commit(repoDir, message string) error

	// Push sends committed changes to the configured remote (origin),
	// equivalent to `git push`.
	Push(repoDir string) error
}
