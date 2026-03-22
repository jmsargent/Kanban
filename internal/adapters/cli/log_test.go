package cli_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// Test Budget: 2 behaviors x 2 = 4 max unit tests (using 2)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeGitPortLog struct {
	repoRoot string
	entries  []ports.CommitEntry
}

func (f *fakeGitPortLog) RepoRoot() (string, error)                           { return f.repoRoot, nil }
func (f *fakeGitPortLog) AppendToGitignore(_, _ string) error                 { return nil }
func (f *fakeGitPortLog) CommitMessagesInRange(_, _ string) ([]string, error) { return nil, nil }
func (f *fakeGitPortLog) GetIdentity() (ports.Identity, error)                { return ports.Identity{}, nil }
func (f *fakeGitPortLog) LogFile(_ string, _ string) ([]ports.CommitEntry, error) {
	return f.entries, nil
}

type fakeConfigRepoLog struct{}

func (f *fakeConfigRepoLog) Read(_ string) (ports.Config, error) {
	return ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}, nil
}
func (f *fakeConfigRepoLog) Write(_ string, _ ports.Config) error { return nil }

type fakeTaskRepoLog struct {
	task domain.Task
}

func (f *fakeTaskRepoLog) FindByID(_, _ string) (domain.Task, error) { return f.task, nil }
func (f *fakeTaskRepoLog) Save(_ string, _ domain.Task) error        { return nil }
func (f *fakeTaskRepoLog) ListAll(_ string) ([]domain.Task, error)   { return nil, nil }
func (f *fakeTaskRepoLog) Update(_ string, _ domain.Task) error      { return nil }
func (f *fakeTaskRepoLog) Delete(_, _ string) error                  { return nil }
func (f *fakeTaskRepoLog) NextID(_ string) (string, error)           { return "", nil }

type fakeTransitionLogLog struct{}

func (f *fakeTransitionLogLog) Append(_ string, _ domain.TransitionEntry) error { return nil }
func (f *fakeTransitionLogLog) LatestStatus(_, _ string) (domain.TaskStatus, error) {
	return domain.StatusTodo, nil
}
func (f *fakeTransitionLogLog) History(_, _ string) ([]domain.TransitionEntry, error) {
	return nil, nil
}

// ─── Helper ──────────────────────────────────────────────────────────────────

func execLog(t *testing.T, git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, taskID string) (stdout string, exitCode int) {
	t.Helper()

	var outBuf, errBuf bytes.Buffer
	capturedCode := 0
	cli.SetOsExit(func(code int) { capturedCode = code })
	t.Cleanup(func() { cli.SetOsExit(nil) })

	cmd := cli.NewLogCommand(git, config, tasks)
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)

	root := &cobra.Command{Use: "kanban", SilenceUsage: true}
	root.AddCommand(cmd)
	root.SetOut(&outBuf)
	root.SetErr(&errBuf)
	root.SetArgs([]string{"log", taskID})

	_ = root.Execute()
	return outBuf.String(), capturedCode
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// TestLogCommand_RendersTransitionFields verifies that each log entry shows
// a timestamp with "T" (ISO 8601), a "->" arrow, author email, and trigger label.
func TestLogCommand_RendersTransitionFields(t *testing.T) {
	ts := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	task := domain.Task{ID: "TASK-001", Title: "Fix OAuth login bug", Status: domain.StatusInProgress}
	git := &fakeGitPortLog{
		repoRoot: t.TempDir(),
		entries: []ports.CommitEntry{
			{
				SHA:       "abc123def456",
				Timestamp: ts,
				Author:    "dev@example.com",
				Message:   "TASK-001: todo->in-progress [trigger:manual]",
			},
		},
	}
	config := &fakeConfigRepoLog{}
	tasks := &fakeTaskRepoLog{task: task}

	stdout, exitCode := execLog(t, git, config, tasks, "TASK-001")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\nOutput: %s", exitCode, stdout)
	}
	// ISO 8601 timestamp contains date+time separator "T" e.g. "2026-03-18T10:00:00Z"
	if !strings.Contains(stdout, "2026-03-18T") {
		t.Errorf("expected output to contain ISO 8601 timestamp '2026-03-18T', got: %q", stdout)
	}
	if !strings.Contains(stdout, "->") {
		t.Errorf("expected output to contain '->' transition arrow, got: %q", stdout)
	}
	if !strings.Contains(stdout, "dev@example.com") {
		t.Errorf("expected output to contain author email, got: %q", stdout)
	}
	if !strings.Contains(stdout, "manual") {
		t.Errorf("expected output to contain trigger label 'manual', got: %q", stdout)
	}
}

// TestLogCommand_RendersCommitSHAAsSupplementaryContext verifies that the
// commit SHA appears in the output as "commit:<sha>" supplementary context.
func TestLogCommand_RendersCommitSHAAsSupplementaryContext(t *testing.T) {
	ts := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	task := domain.Task{ID: "TASK-001", Title: "Deploy to staging", Status: domain.StatusInProgress}
	git := &fakeGitPortLog{
		repoRoot: t.TempDir(),
		entries: []ports.CommitEntry{
			{
				SHA:       "abc123def456",
				Timestamp: ts,
				Author:    "dev@example.com",
				Message:   "TASK-001: todo->in-progress [trigger:manual]",
			},
		},
	}
	config := &fakeConfigRepoLog{}
	tasks := &fakeTaskRepoLog{task: task}

	stdout, exitCode := execLog(t, git, config, tasks, "TASK-001")

	if exitCode != 0 {
		t.Errorf("expected exit 0, got %d\nOutput: %s", exitCode, stdout)
	}
	if !strings.Contains(stdout, "commit:") {
		t.Errorf("expected output to contain 'commit:' supplementary context, got: %q", stdout)
	}
}
