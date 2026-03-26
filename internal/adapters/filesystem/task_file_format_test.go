// acceptance: Created task file contains valid YAML front matter
// Driving port: TaskRepository.Save
// Asserts: written file begins with "---\n", contains id/title/priority/due/assignee fields,
// does NOT contain status: (status is tracked in .kanban/transitions.log),
// and DOES contain a comment referencing transitions.log.
package filesystem_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmsargent/kanban/internal/adapters/filesystem"
	"github.com/jmsargent/kanban/internal/domain"
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

	// Must contain all required non-status fields
	for _, field := range []string{"id:", "title:", "priority:", "due:", "assignee:"} {
		if !strings.Contains(content, field) {
			t.Errorf("front matter missing field %q", field)
		}
	}

	// Status is tracked in transitions.log — must NOT appear in the task file
	if strings.Contains(content, "status:") {
		t.Errorf("task file must not contain 'status:' field (status tracked in transitions.log)\nContent:\n%s", content)
	}

	// Must contain comment directing readers to transitions.log
	if !strings.Contains(content, "transitions.log") {
		t.Errorf("task file must contain 'transitions.log' comment\nContent:\n%s", content)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
