package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/jmsargent/kanban/internal/adapters/cli"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// Test Budget: 6 behaviors x 2 = 12 max unit tests (using 7)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeGitPortCLI struct {
	repoRoot       string
	repoRootErr    error
	identity       ports.Identity
	getIdentityErr error
}

func (f *fakeGitPortCLI) RepoRoot() (string, error) {
	return f.repoRoot, f.repoRootErr
}

func (f *fakeGitPortCLI) AppendToGitignore(_, _ string) error                 { return nil }
func (f *fakeGitPortCLI) CommitMessagesInRange(_, _ string) ([]string, error) { return nil, nil }
func (f *fakeGitPortCLI) GetIdentity() (ports.Identity, error) {
	return f.identity, f.getIdentityErr
}

func (f *fakeGitPortCLI) LogFile(_ string, _ string) ([]ports.CommitEntry, error) {
	return nil, nil
}

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
// with the given task ID, and returns the captured stdout, stderr, and the exit
// code captured via the osExit override (0 = no exit called).
func execStart(t *testing.T, git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, taskID string) (stdout, stderr string, exitCode int) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer
	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewStartCommand(git, config, tasks)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"start", taskID})

	_ = root.Execute()
	return outBuf.String(), errBuf.String(), capturedCode
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestStartCommand_PrintsStartedMessage_WhenTaskIsInTodo(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo}
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Dev", Email: "dev@example.com"}}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, exitCode := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stdout, "Started TASK-001: Fix login bug") {
		t.Errorf("expected stdout to contain 'Started TASK-001: Fix login bug', got: %q", stdout)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestStartCommand_PrintsAlreadyInProgress_WhenTaskIsInProgress(t *testing.T) {
	task := domain.Task{ID: "TASK-002", Title: "Running task", Status: domain.StatusInProgress}
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Dev", Email: "dev@example.com"}}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, exitCode := execStart(t, git, config, tasks, "TASK-002")

	if !strings.Contains(stdout, "Task TASK-002 is already in progress") {
		t.Errorf("expected stdout to contain 'Task TASK-002 is already in progress', got: %q", stdout)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0 when task is already in progress, got %d", exitCode)
	}
}

func TestStartCommand_PrintsAlreadyFinished_WhenTaskIsDone(t *testing.T) {
	task := domain.Task{ID: "TASK-003", Title: "Completed task", Status: domain.StatusDone}
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Dev", Email: "dev@example.com"}}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	_, stderr, exitCode := execStart(t, git, config, tasks, "TASK-003")

	if !strings.Contains(stderr, "Task TASK-003 is already finished") {
		t.Errorf("expected stderr to contain 'Task TASK-003 is already finished', got: %q", stderr)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when task is done, got %d", exitCode)
	}
}

func TestStartCommand_PrintsNotFound_WhenTaskIDDoesNotExist(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Dev", Email: "dev@example.com"}}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI() // empty

	_, stderr, exitCode := execStart(t, git, config, tasks, "TASK-999")

	if !strings.Contains(stderr, "Task TASK-999 not found") {
		t.Errorf("expected stderr to contain 'Task TASK-999 not found', got: %q", stderr)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when task not found, got %d", exitCode)
	}
}

func TestStartCommand_PrintsNotInitialised_WhenRepoNotInitialised(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Dev", Email: "dev@example.com"}}
	config := &fakeConfigRepoCLI{readErr: ports.ErrNotInitialised}
	tasks := newFakeTaskRepoCLI()

	_, stderr, exitCode := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stderr, "kanban not initialised") {
		t.Errorf("expected stderr to contain 'kanban not initialised', got: %q", stderr)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when repo not initialised, got %d", exitCode)
	}
}

func TestStartCommand_ExitsWithError_WhenGitIdentityNotConfigured(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo}
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), getIdentityErr: ports.ErrGitIdentityNotConfigured}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	_, stderr, exitCode := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stderr, "git identity not configured") {
		t.Errorf("expected stderr to contain 'git identity not configured', got: %q", stderr)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when git identity not configured, got %d", exitCode)
	}
}

func TestStartCommand_PrintsPreviouslyAssignedWarning_WhenReassignedFromAnotherDeveloper(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo, Assignee: "Alice"}
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: ports.Identity{Name: "Bob", Email: "bob@example.com"}}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, exitCode := execStart(t, git, config, tasks, "TASK-001")

	if !strings.Contains(stdout, "Note: task was previously assigned to Alice") {
		t.Errorf("expected stdout to contain reassignment note, got: %q", stdout)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0 on reassignment, got %d", exitCode)
	}
}
