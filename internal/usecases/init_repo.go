package usecases

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// InitRepo implements the repository initialisation use case.
// It is the driving port entrypoint for "kanban init".
type InitRepo struct {
	git    ports.GitPort
	config ports.ConfigRepository
	out    io.Writer
}

// NewInitRepo constructs an InitRepo use case with its required collaborators.
func NewInitRepo(git ports.GitPort, config ports.ConfigRepository, out io.Writer) *InitRepo {
	return &InitRepo{git: git, config: config, out: out}
}

// Execute runs the initialisation use case:
//  1. Resolves the repository root via GitPort.
//  2. Returns ErrNotGitRepo when not inside a git repository.
//  3. Skips initialisation if already done (idempotent).
//  4. Writes the default config, appends the hook log to .gitignore,
//     and installs the commit-msg hook.
func (u *InitRepo) Execute() error {
	repoRoot, err := u.git.RepoRoot()
	if err != nil {
		return ports.ErrNotGitRepo
	}

	_, readErr := u.config.Read(repoRoot)
	if readErr == nil {
		_, _ = fmt.Fprintln(u.out, "Already initialised at .kanban/ -- no changes made.")
		return nil
	}
	if !errors.Is(readErr, ports.ErrNotInitialised) {
		return fmt.Errorf("read config: %w", readErr)
	}

	tasksDir := filepath.Join(repoRoot, ".kanban", "tasks")
	if err = os.MkdirAll(tasksDir, 0o755); err != nil {
		return fmt.Errorf("create tasks dir: %w", err)
	}

	defaultConfig := ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "TODO"},
			{Name: "in-progress", Label: "IN PROGRESS"},
			{Name: "done", Label: "DONE"},
		},
		CITaskPattern: `TASK-[0-9]+`,
	}

	if err = u.config.Write(repoRoot, defaultConfig); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if err = u.git.AppendToGitignore(repoRoot, ".kanban/hook.log"); err != nil {
		return fmt.Errorf("append to .gitignore: %w", err)
	}

	if err = u.git.InstallHook(repoRoot); err != nil {
		return fmt.Errorf("install hook: %w", err)
	}

	_, _ = fmt.Fprintln(u.out, "Initialised kanban at .kanban/")
	return nil
}
