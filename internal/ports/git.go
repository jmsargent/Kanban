package ports

// Identity holds the git author identity (name and email) for a repository user.
// This is an infrastructure concern sourced from git config — it belongs in ports.
type Identity struct {
	Name  string
	Email string
}

// GitPort is the driven port for all git operations required by use cases.
// Implementations live in internal/adapters; this package declares contract only.
type GitPort interface {
	// RepoRoot returns the absolute path to the root of the current git repository.
	// Returns ErrNotGitRepo when the working directory is not inside a git repo.
	RepoRoot() (string, error)

	// CommitMessagesInRange returns the commit messages for all commits
	// reachable from `to` but not from `from` (equivalent to `git log from..to`).
	CommitMessagesInRange(from, to string) ([]string, error)

	// CommitFiles stages the given paths and creates a commit with the supplied
	// message inside the repository at repoRoot.
	CommitFiles(repoRoot, message string, paths []string) error

	// InstallHook writes the kanban commit-msg hook into the .git/hooks directory
	// of the repository at repoRoot, making it executable.
	InstallHook(repoRoot string) error

	// AppendToGitignore adds entry to the .gitignore file at the root of repoRoot,
	// creating the file if it does not exist.
	AppendToGitignore(repoRoot, entry string) error

	// GetIdentity returns the git author identity (user.name and user.email)
	// configured in the current git context.
	// Returns ErrGitIdentityNotConfigured when user.name is not set.
	GetIdentity() (Identity, error)
}
