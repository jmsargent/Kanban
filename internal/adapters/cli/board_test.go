package cli_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// Test Budget: 2 behaviors x 2 = 4 max unit tests (using 2)

// fakeTaskRepoBoardCLI is a task repository fake for board tests that supports ListAll.
type fakeTaskRepoBoardCLI struct {
	tasks []domain.Task
}

func (f *fakeTaskRepoBoardCLI) ListAll(_ string) ([]domain.Task, error) { return f.tasks, nil }
func (f *fakeTaskRepoBoardCLI) FindByID(_, _ string) (domain.Task, error) {
	return domain.Task{}, ports.ErrTaskNotFound
}
func (f *fakeTaskRepoBoardCLI) Save(_ string, _ domain.Task) error   { return nil }
func (f *fakeTaskRepoBoardCLI) Update(_ string, _ domain.Task) error { return nil }
func (f *fakeTaskRepoBoardCLI) Delete(_, _ string) error             { return nil }
func (f *fakeTaskRepoBoardCLI) NextID(_ string) (string, error)      { return "", nil }

// captureStdout redirects os.Stdout to a pipe, executes fn, and returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// execBoardOutput runs the board command and captures its text output.
func execBoardOutput(t *testing.T, tasks []domain.Task) string {
	t.Helper()
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	repo := &fakeTaskRepoBoardCLI{tasks: tasks}

	cmd := cli.NewBoardCommand(git, config, repo)
	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{"board"})

	return captureStdout(t, func() {
		_ = root.Execute()
	})
}

// Behavior 1: task with non-empty CreatedBy shows creator name in board row.
func TestBoardCommand_ShowsCreatorName_WhenCreatedByIsSet(t *testing.T) {
	task := domain.Task{
		ID:        "TASK-001",
		Title:     "Fix login bug",
		Status:    domain.StatusTodo,
		CreatedBy: "Test User",
	}

	output := execBoardOutput(t, []domain.Task{task})

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "TASK-001") {
			if !strings.Contains(line, "Test User") {
				t.Errorf("board row for TASK-001 does not contain creator name\nRow: %q\nFull output:\n%s", line, output)
			}
			return
		}
	}
	t.Errorf("board row for TASK-001 not found in output:\n%s", output)
}

// Behavior 2: task with empty CreatedBy shows "--" in the creator column.
func TestBoardCommand_ShowsDash_WhenCreatedByIsEmpty(t *testing.T) {
	task := domain.Task{
		ID:        "TASK-002",
		Title:     "Legacy task",
		Status:    domain.StatusTodo,
		CreatedBy: "",
	}

	output := execBoardOutput(t, []domain.Task{task})

	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, "TASK-002") {
			if !strings.Contains(line, "--") {
				t.Errorf("board row for TASK-002 does not contain \"--\" for empty creator\nRow: %q\nFull output:\n%s", line, output)
			}
			return
		}
	}
	t.Errorf("board row for TASK-002 not found in output:\n%s", output)
}
