package filesystem_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	"github.com/kanban-tasks/kanban/internal/domain"
)

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 3)

// Behavior 1: Append writes a correctly formatted 5-field line to transitions.log
func TestTransitionLogAdapter_Append_WritesFormattedLine(t *testing.T) {
	repoRoot := t.TempDir()
	adapter := filesystem.NewTransitionLogAdapter()

	entry := domain.TransitionEntry{
		Timestamp: time.Date(2026, 3, 15, 9, 14, 23, 0, time.UTC),
		TaskID:    "TASK-007",
		From:      domain.StatusTodo,
		To:        domain.StatusInProgress,
		Author:    "jon@example.com",
		Trigger:   "manual",
	}

	if err := adapter.Append(repoRoot, entry); err != nil {
		t.Fatalf("Append: %v", err)
	}

	logPath := filepath.Join(repoRoot, ".kanban", "transitions.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read transitions.log: %v", err)
	}

	line := strings.TrimSpace(string(data))
	wantContains := []string{"TASK-007", "todo->in-progress", "jon@example.com", "manual"}
	for _, want := range wantContains {
		if !strings.Contains(line, want) {
			t.Errorf("expected line to contain %q\nLine: %s", want, line)
		}
	}

	fields := strings.Fields(line)
	if len(fields) != 5 {
		t.Errorf("expected exactly 5 fields in log line, got %d\nLine: %s", len(fields), line)
	}
}

// Behavior 2: LatestStatus returns StatusTodo (no error) when log has no entry for taskID
func TestTransitionLogAdapter_LatestStatus_ReturnsStatusTodo_WhenNoEntry(t *testing.T) {
	repoRoot := t.TempDir()
	adapter := filesystem.NewTransitionLogAdapter()

	status, err := adapter.LatestStatus(repoRoot, "TASK-001")
	if err != nil {
		t.Fatalf("LatestStatus: expected nil error for missing entry, got %v", err)
	}
	if status != domain.StatusTodo {
		t.Errorf("expected StatusTodo for missing entry, got %q", status)
	}
}

// Behavior 4: Concurrent Append calls for the same task+transition produce exactly 1 line
// (idempotency maintained under concurrency — the lock must cover read-check-write)
func TestTransitionLogAdapter_Append_ConcurrentSameTransition_ProducesOneLine(t *testing.T) {
	repoRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(repoRoot, ".kanban"), 0o755); err != nil {
		t.Fatalf("mkdir .kanban: %v", err)
	}
	adapter := filesystem.NewTransitionLogAdapter()

	entry := domain.TransitionEntry{
		Timestamp: time.Now().UTC(),
		TaskID:    "TASK-001",
		From:      domain.StatusTodo,
		To:        domain.StatusInProgress,
		Author:    "dev@example.com",
		Trigger:   "manual",
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = adapter.Append(repoRoot, entry)
		}()
	}
	wg.Wait()

	logPath := filepath.Join(repoRoot, ".kanban", "transitions.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read transitions.log: %v", err)
	}
	lines := 0
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if strings.TrimSpace(line) != "" {
			lines++
		}
	}
	if lines != 1 {
		t.Errorf("expected exactly 1 line after 10 concurrent appends of same transition, got %d\nContent:\n%s", lines, string(data))
	}
}

// Behavior 3: Save (task file) omits status field and includes transitions.log comment
func TestTaskRepository_Save_OmitsStatusFieldAndIncludesTransitionsLogComment(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	task := domain.Task{
		ID:     "TASK-001",
		Title:  "Fix OAuth login bug",
		Status: domain.StatusTodo,
	}

	if err := repo.Save(repoRoot, task); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(repoRoot, ".kanban", "tasks", "TASK-001.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read task file: %v", err)
	}

	content := string(data)
	if strings.Contains(content, "status:") {
		t.Errorf("task file must not contain 'status:' field\nContent:\n%s", content)
	}
	if !strings.Contains(content, "transitions.log") {
		t.Errorf("task file must contain 'transitions.log' comment\nContent:\n%s", content)
	}
}
