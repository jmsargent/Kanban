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

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 3)

// ─── Helper ──────────────────────────────────────────────────────────────────

// execDone constructs the done command wired to in-memory fakes, executes it
// with the given task ID, and returns the captured stdout, stderr, and the exit
// code captured via the osExit override (0 = no exit called).
func execDone(t *testing.T, git ports.GitPort, tasks ports.TaskRepository, taskID string) (stdout, stderr string, exitCode int) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer
	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewDoneCommand(git, tasks)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"done", taskID})

	_ = root.Execute()
	return outBuf.String(), errBuf.String(), capturedCode
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestDoneCommand_PrintsMovedMessage_WhenTaskTransitionsToDone(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix OAuth login bug", Status: domain.StatusInProgress}
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, exitCode := execDone(t, git, tasks, "TASK-001")

	if !strings.Contains(stdout, "moved in-progress -> done") {
		t.Errorf("expected stdout to contain 'moved in-progress -> done', got: %q", stdout)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
}

func TestDoneCommand_PrintsAlreadyDone_WhenTaskIsAlreadyDone(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Deploy to production", Status: domain.StatusDone}
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	tasks := newFakeTaskRepoCLI(task)

	stdout, _, exitCode := execDone(t, git, tasks, "TASK-001")

	if !strings.Contains(stdout, "already done") {
		t.Errorf("expected stdout to contain 'already done', got: %q", stdout)
	}
	if exitCode != 0 {
		t.Errorf("expected exit code 0 for idempotent done, got %d", exitCode)
	}
}

func TestDoneCommand_ExitsWithError_WhenTaskIDDoesNotExist(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	tasks := newFakeTaskRepoCLI() // empty

	_, stderr, exitCode := execDone(t, git, tasks, "TASK-999")

	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected stderr to contain 'not found', got: %q", stderr)
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1 when task not found, got %d", exitCode)
	}
}
