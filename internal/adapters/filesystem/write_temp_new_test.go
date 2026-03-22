package filesystem_test

import (
	"os"
	"strings"
	"testing"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
)

// Test Budget: 1 behavior x 2 = 2 max integration tests (using 2)
// Behavior: WriteTempNew creates a temp file with a blank task template

// TestWriteTempNew_CreatesFileWithBlankTitleField verifies that WriteTempNew
// returns a valid file path containing a blank title field. Observable outcome:
// file exists and contains `title: ""`.
func TestWriteTempNew_CreatesFileWithBlankTitleField(t *testing.T) {
	repo := filesystem.NewTaskRepository()

	path, err := repo.WriteTempNew()
	if err != nil {
		t.Fatalf("WriteTempNew: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	if path == "" {
		t.Fatal("expected non-empty file path")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}

	if !strings.Contains(string(content), `title: ""`) {
		t.Errorf("expected temp file to contain blank title field, got:\n%s", string(content))
	}
}

// TestWriteTempNew_FileContainsTitleRequiredComment verifies that the template
// contains a comment indicating that the title field is required.
func TestWriteTempNew_FileContainsTitleRequiredComment(t *testing.T) {
	repo := filesystem.NewTaskRepository()

	path, err := repo.WriteTempNew()
	if err != nil {
		t.Fatalf("WriteTempNew: %v", err)
	}
	t.Cleanup(func() { _ = os.Remove(path) })

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#") &&
			strings.Contains(strings.ToLower(trimmed), "required") {
			return
		}
	}
	t.Errorf("expected a comment containing 'required' in template, got:\n%s", string(content))
}
