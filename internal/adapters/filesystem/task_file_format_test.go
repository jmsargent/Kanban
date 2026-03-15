// acceptance: Created task file contains valid YAML front matter
// Driving port: TaskRepository.Save
// Asserts: written file begins with "---\n" and contains id, title, status, priority, due, assignee fields
package filesystem_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	"github.com/kanban-tasks/kanban/internal/domain"
)

func TestSave_CreatesFileWithValidYAMLFrontMatter(t *testing.T) {
	repoRoot := t.TempDir()
	if err := exec.Command("git", "init", repoRoot).Run(); err != nil {
		t.Fatalf("git init: %v", err)
	}

	due := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	task := domain.Task{
		ID:       "TASK-001",
		Title:    "Format validation task",
		Status:   domain.StatusTodo,
		Priority: "medium",
		Due:      &due,
		Assignee: "alice",
	}

	repo := filesystem.NewTaskRepository()
	if err := repo.Save(repoRoot, task); err != nil {
		t.Fatalf("Save: %v", err)
	}

	taskFile := filepath.Join(repoRoot, ".kanban", "tasks", "TASK-001.md")
	data, err := os.ReadFile(taskFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	content := string(data)

	// Must begin with YAML front matter delimiter
	if !strings.HasPrefix(content, "---\n") {
		t.Errorf("file does not begin with YAML front matter delimiter '---\\n': %q", content[:min(len(content), 20)])
	}

	// Must contain all required fields
	for _, field := range []string{"id:", "title:", "status:", "priority:", "due:", "assignee:"} {
		if !strings.Contains(content, field) {
			t.Errorf("front matter missing field %q", field)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
