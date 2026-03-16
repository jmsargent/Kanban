package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// Test Budget: 5 behaviors x 2 = 10 max unit tests (using 6)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeGitPortCLI struct {
	repoRoot    string
	repoRootErr error
}

func (f *fakeGitPortCLI) RepoRoot() (string, error) {
	return f.repoRoot, f.repoRootErr
}

func (f *fakeGitPortCLI) InstallHook(_ string) error                          { return nil }
func (f *fakeGitPortCLI) AppendToGitignore(_, _ string) error                 { return nil }
func (f *fakeGitPortCLI) CommitMessagesInRange(_, _ string) ([]string, error) { return nil, nil }
func (f *fakeGitPortCLI) CommitFiles(_, _ string, _ []string) error           { return nil }

type fakeConfigRepoCLI struct {
	readErr error
}

func (f *fakeConfigRepoCLI) Read(_ string) (ports.Config, error) {
	return ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}, f.readErr
}

func (f *fakeConfigRepoCLI) Write(_ string, _ ports.Config) error { return nil }

type fakeTaskRepoCLI struct {
	byID map[string]domain.Task
}

func newFakeTaskRepoCLI(tasks ...domain.Task) *fakeTaskRepoCLI {
	repo := &fakeTaskRepoCLI{byID: make(map[string]domain.Task)}
	for _, t := range tasks {
		repo.byID[t.ID] = t
	}
	return repo
}

func (f *fakeTaskRepoCLI) FindByID(_, taskID string) (domain.Task, error) {
	t, ok := f.byID[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (f *fakeTaskRepoCLI) Update(_ string, task domain.Task) error {
	f.byID[task.ID] = task
	return nil
}

func (f *fakeTaskRepoCLI) Save(_ string, _ domain.Task) error      { return nil }
func (f *fakeTaskRepoCLI) ListAll(_ string) ([]domain.Task, error) { return nil, nil }
func (f *fakeTaskRepoCLI) Delete(_, _ string) error                { return nil }
func (f *fakeTaskRepoCLI) NextID(_ string) (string, error)         { return "", nil }

// ─── Helper ──────────────────────────────────────────────────────────────────

// execStart constructs the start command wired to in-memory fakes, executes it
// with the given task ID, and returns the captured stdout, stderr, and whether
// RunE returned an error (used as a proxy for non-zero exit intent).
func execStart(t *testing.T, git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, taskID string) (stdout, stderr string, hadError bool) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer

	cmd := cli.NewStartCommand(git, config, tasks)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"start", taskID})

	err := root.Execute()
	return outBuf.String(), errBuf.String(), err != nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestStartCommand_PrintsStartedMessage_WhenTaskIsInTodo(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo}
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, hadError := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stdout, "Started TASK-001: Fix login bug") {
		t.Errorf("expected stdout to contain 'Started TASK-001: Fix login bug', got: %q", stdout)
	}
	if hadError {
		t.Error("expected no error for successful start")
	}
}

func TestStartCommand_PrintsAlreadyInProgress_WhenTaskIsInProgress(t *testing.T) {
	task := domain.Task{ID: "TASK-002", Title: "Running task", Status: domain.StatusInProgress}
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, hadError := execStart(t, git, config, tasks, "TASK-002")

	if !strings.Contains(stdout, "Task TASK-002 is already in progress") {
		t.Errorf("expected stdout to contain 'Task TASK-002 is already in progress', got: %q", stdout)
	}
	if hadError {
		t.Error("expected no error (exit 0) when task is already in progress")
	}
}

func TestStartCommand_PrintsAlreadyFinished_WhenTaskIsDone(t *testing.T) {
	task := domain.Task{ID: "TASK-003", Title: "Completed task", Status: domain.StatusDone}
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	_, stderr, hadError := execStart(t, git, config, tasks, "TASK-003")

	if !strings.Contains(stderr, "Task TASK-003 is already finished") {
		t.Errorf("expected stderr to contain 'Task TASK-003 is already finished', got: %q", stderr)
	}
	if !hadError {
		t.Error("expected error (exit 1) when task is done")
	}
}

func TestStartCommand_PrintsNotFound_WhenTaskIDDoesNotExist(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI() // empty

	_, stderr, hadError := execStart(t, git, config, tasks, "TASK-999")

	if !strings.Contains(stderr, "Task TASK-999 not found") {
		t.Errorf("expected stderr to contain 'Task TASK-999 not found', got: %q", stderr)
	}
	if !hadError {
		t.Error("expected error (exit 1) when task not found")
	}
}

func TestStartCommand_PrintsNotInitialised_WhenRepoNotInitialised(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{readErr: ports.ErrNotInitialised}
	tasks := newFakeTaskRepoCLI()

	_, stderr, hadError := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stderr, "kanban not initialised") {
		t.Errorf("expected stderr to contain 'kanban not initialised', got: %q", stderr)
	}
	if !hadError {
		t.Error("expected error (exit 1) when repo not initialised")
	}
}

func TestStartCommand_AppearsInHelpOutput(t *testing.T) {
	git := &fakeGitPortCLI{}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI()

	cmd := cli.NewStartCommand(git, config, tasks)

	if !strings.HasPrefix(cmd.Use, "start") {
		t.Errorf("expected command Use to start with 'start', got: %q", cmd.Use)
	}
	if cmd.Short == "" {
		t.Error("expected command to have a non-empty Short description")
	}
}
