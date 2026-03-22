package usecases_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeGitPort struct {
	repoRootResult string
	repoRootErr    error
	installedHook  bool
	gitignoreEntry string
}

func (f *fakeGitPort) RepoRoot() (string, error) {
	return f.repoRootResult, f.repoRootErr
}

func (f *fakeGitPort) AppendToGitignore(repoRoot, entry string) error {
	f.gitignoreEntry = entry
	return nil
}

func (f *fakeGitPort) CommitMessagesInRange(from, to string) ([]string, error) {
	return nil, nil
}

func (f *fakeGitPort) GetIdentity() (ports.Identity, error) {
	return ports.Identity{}, nil
}

func (f *fakeGitPort) LogFile(repoRoot, filePath string) ([]ports.CommitEntry, error) {
	return nil, nil
}

type fakeConfigRepo struct {
	written    *ports.Config
	readResult ports.Config
	readErr    error
}

// newFreshConfigRepo returns a fake representing a repository with no kanban setup.
func newFreshConfigRepo() *fakeConfigRepo {
	return &fakeConfigRepo{readErr: ports.ErrNotInitialised}
}

func (f *fakeConfigRepo) Read(repoRoot string) (ports.Config, error) {
	return f.readResult, f.readErr
}

func (f *fakeConfigRepo) Write(repoRoot string, config ports.Config) error {
	f.written = &config
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 6)

func tmpRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

func TestInitRepo_CreatesDefaultConfigWithColumns(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &fakeGitPort{repoRootResult: repoRoot}
	cfg := newFreshConfigRepo()
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	err := uc.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.written == nil {
		t.Fatal("expected config to be written, but was not")
	}
	if len(cfg.written.Columns) == 0 {
		t.Error("expected default columns to be written to config")
	}
	if cfg.written.CITaskPattern == "" {
		t.Error("expected ci_task_pattern to be set in config")
	}
}

func TestInitRepo_DoesNotInstallHookOrModifyGitignore(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &fakeGitPort{repoRootResult: repoRoot}
	cfg := newFreshConfigRepo()
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	err := uc.Execute()

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if git.installedHook {
		t.Error("expected InstallHook NOT to be called (C-03)")
	}
	if git.gitignoreEntry != "" {
		t.Errorf("expected AppendToGitignore NOT to be called, got entry: %q", git.gitignoreEntry)
	}
}

func TestInitRepo_PrintsSuccessMessage(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &fakeGitPort{repoRootResult: repoRoot}
	cfg := newFreshConfigRepo()
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	_ = uc.Execute()

	if !strings.Contains(output.String(), "Initialised kanban at .kanban/") {
		t.Errorf("expected success message, got: %q", output.String())
	}
}

func TestInitRepo_ReturnsErrNotGitRepo_WhenNotInGitRepo(t *testing.T) {
	git := &fakeGitPort{repoRootErr: ports.ErrNotGitRepo}
	cfg := newFreshConfigRepo()
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	err := uc.Execute()

	if !errors.Is(err, ports.ErrNotGitRepo) {
		t.Errorf("expected ErrNotGitRepo, got: %v", err)
	}
}

func TestInitRepo_IsIdempotent_WhenAlreadyInitialised(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &fakeGitPort{repoRootResult: repoRoot}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	err := uc.Execute()

	if err != nil {
		t.Fatalf("expected no error on second run, got: %v", err)
	}
	if cfg.written != nil {
		t.Error("expected config NOT to be rewritten when already initialised")
	}
	if !strings.Contains(output.String(), "Already initialised") {
		t.Errorf("expected 'Already initialised' in output, got: %q", output.String())
	}
}

func TestInitRepo_DefaultConfig_HasExpectedColumns(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &fakeGitPort{repoRootResult: repoRoot}
	cfg := newFreshConfigRepo()
	output := &strings.Builder{}

	uc := usecases.NewInitRepo(git, cfg, output)
	_ = uc.Execute()

	expected := []string{"todo", "in-progress", "done"}
	for i, col := range cfg.written.Columns {
		if i >= len(expected) {
			break
		}
		if col.Name != expected[i] {
			t.Errorf("column[%d] expected name %q, got %q", i, expected[i], col.Name)
		}
	}
}
