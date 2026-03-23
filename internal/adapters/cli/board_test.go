package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// Test Budget: 6 behaviors x 2 = 12 max unit tests (using 6)

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

// execBoardJSONOutput runs the board command with --json flag and captures its output.
func execBoardJSONOutput(t *testing.T, tasks []domain.Task) string {
	t.Helper()
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	repo := &fakeTaskRepoBoardCLI{tasks: tasks}

	cmd := cli.NewBoardCommand(git, config, repo)
	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{"board", "--json"})

	return captureStdout(t, func() {
		_ = root.Execute()
	})
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

// Behavior 3: JSON output always includes a created_by field on each task object.
func TestBoardCommand_JSONOutputContainsCreatedByField(t *testing.T) {
	task := domain.Task{
		ID:        "TASK-001",
		Title:     "Fix login bug",
		Status:    domain.StatusTodo,
		CreatedBy: "Test User",
	}

	output := execBoardJSONOutput(t, []domain.Task{task})

	var items []map[string]interface{}
	if err := json.Unmarshal([]byte(output), &items); err != nil {
		t.Fatalf("JSON output is not valid JSON: %v\nOutput: %s", err, output)
	}
	if len(items) == 0 {
		t.Fatalf("expected at least one task in JSON output, got none")
	}
	if _, ok := items[0]["created_by"]; !ok {
		t.Errorf("JSON task object missing created_by field — present keys: %v", keys(items[0]))
	}
}

// execBoardMeOutput runs the board command with the --me flag and captures its text output.
func execBoardMeOutput(t *testing.T, identity ports.Identity, tasks []domain.Task) string {
	t.Helper()
	git := &fakeGitPortCLI{repoRoot: t.TempDir(), identity: identity}
	config := &fakeConfigRepoCLI{}
	repo := &fakeTaskRepoBoardCLI{tasks: tasks}

	cmd := cli.NewBoardCommand(git, config, repo)
	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{"board", "--me"})

	return captureStdout(t, func() {
		_ = root.Execute()
	})
}

// Behavior 4: --me filters board to show only tasks assigned to the current developer.
func TestBoardCommand_MeFlag_ShowsOnlyDeveloperTasks(t *testing.T) {
	myTask := domain.Task{
		ID:       "TASK-001",
		Title:    "My task",
		Status:   domain.StatusTodo,
		Assignee: "dev@example.com",
	}
	otherTask := domain.Task{
		ID:       "TASK-002",
		Title:    "Colleague task",
		Status:   domain.StatusTodo,
		Assignee: "other@example.com",
	}
	identity := ports.Identity{Name: "Dev", Email: "dev@example.com"}

	output := execBoardMeOutput(t, identity, []domain.Task{myTask, otherTask})

	if !strings.Contains(output, "TASK-001") {
		t.Errorf("expected board --me to contain TASK-001 (my task)\nOutput:\n%s", output)
	}
	if strings.Contains(output, "TASK-002") {
		t.Errorf("expected board --me to exclude TASK-002 (colleague's task)\nOutput:\n%s", output)
	}
}

// Behavior 5: --me with no matching tasks shows an informational message and exits cleanly.
func TestBoardCommand_MeFlag_ShowsNoTasksMessage_WhenNothingAssigned(t *testing.T) {
	otherTask := domain.Task{
		ID:       "TASK-001",
		Title:    "Colleague task",
		Status:   domain.StatusTodo,
		Assignee: "other@example.com",
	}
	identity := ports.Identity{Name: "Dev", Email: "dev@example.com"}

	output := execBoardMeOutput(t, identity, []domain.Task{otherTask})

	if !strings.Contains(output, "no tasks") {
		t.Errorf("expected board --me with no matching tasks to output a 'no tasks' message\nOutput:\n%s", output)
	}
}

// captureStderr redirects os.Stderr to a pipe, executes fn, and returns the captured output.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStderr := os.Stderr
	os.Stderr = w

	fn()

	_ = w.Close()
	os.Stderr = origStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String()
}

// Behavior 6: --mermaid and --json together exit 2 with "mutually exclusive" in stderr and no stdout.
func TestBoardCommand_MermaidAndJSON_AreExclusive(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	repo := &fakeTaskRepoBoardCLI{}

	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewBoardCommand(git, config, repo)
	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{"board", "--mermaid", "--json"})

	var stdout string
	stderr := captureStderr(t, func() {
		stdout = captureStdout(t, func() {
			_ = root.Execute()
		})
	})

	if capturedCode != 2 {
		t.Errorf("expected exit code 2, got %d", capturedCode)
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected stderr to contain 'mutually exclusive', got: %q", stderr)
	}
	if stdout != "" {
		t.Errorf("expected stdout to be empty, got: %q", stdout)
	}
}

// Behavior 7: --out without --mermaid exits 2 with "--out requires --mermaid" in stderr.
func TestBoardCommand_OutWithoutMermaid_IsUsageError(t *testing.T) {
	git := &fakeGitPortCLI{repoRoot: t.TempDir()}
	config := &fakeConfigRepoCLI{}
	repo := &fakeTaskRepoBoardCLI{}

	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewBoardCommand(git, config, repo)
	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetArgs([]string{"board", "--out", "README.md"})

	stderr := captureStderr(t, func() {
		_ = root.Execute()
	})

	if capturedCode != 2 {
		t.Errorf("expected exit code 2, got %d", capturedCode)
	}
	if !strings.Contains(stderr, "--out requires --mermaid") {
		t.Errorf("expected stderr to contain '--out requires --mermaid', got: %q", stderr)
	}
}

// keys returns the sorted keys of a map for diagnostic messages.
func keys(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
