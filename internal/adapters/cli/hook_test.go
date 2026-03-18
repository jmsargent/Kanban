package cli_test

// Test Budget: 1 behavior x 2 = 2 max unit tests (using 1)
// Behavior: hook writes "warning" to stderr and returns nil when log.Append fails.

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// failingTransitionLog always returns an error from Append to simulate an unwritable log.
type failingTransitionLog struct {
	appendErr error
}

func (f *failingTransitionLog) Append(_ string, _ domain.TransitionEntry) error {
	return f.appendErr
}
func (f *failingTransitionLog) LatestStatus(_, _ string) (domain.TaskStatus, error) {
	return domain.StatusTodo, nil
}
func (f *failingTransitionLog) History(_, _ string) ([]domain.TransitionEntry, error) {
	return nil, nil
}

// execHookCommitMsg wires the hook command to in-memory fakes and runs it with
// the given commit-message file path. Returns the captured stdout, stderr, and
// the RunE return value (nil = exit 0).
func execHookCommitMsg(t *testing.T, git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository, msgFile string) (stdout, stderr string) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer
	cmd := cli.NewHookCommand(git, config, tasks, log)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true, SilenceErrors: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"_hook", "commit-msg", msgFile})

	_ = root.Execute()
	return outBuf.String(), errBuf.String()
}

func TestHookCommand_WritesWarningToStderr_WhenLogAppendFails(t *testing.T) {
	// Arrange: a task file with a real commit-msg file referencing TASK-001.
	repoDir := t.TempDir()
	msgFile := repoDir + "/COMMIT_EDITMSG"
	if err := os.WriteFile(msgFile, []byte("TASK-001: add feature"), 0o644); err != nil {
		t.Fatalf("setup: create commit-msg file: %v", err)
	}

	task := domain.Task{ID: "TASK-001", Title: "Add feature", Status: domain.StatusTodo}
	git := &fakeGitPortCLI{repoRoot: repoDir}
	config := &fakeConfigRepoCLI{}
	tasks := newFakeTaskRepoCLI(task)
	log := &failingTransitionLog{appendErr: errors.New("permission denied")}

	// Act: run the hook through its driving port (the _hook commit-msg subcommand).
	_, stderr := execHookCommitMsg(t, git, config, tasks, log, msgFile)

	// Assert: stderr contains "warning" (the observable business outcome).
	if !strings.Contains(strings.ToLower(stderr), "warning") {
		t.Errorf("expected stderr to contain 'warning' when log.Append fails, got: %q", stderr)
	}
}
