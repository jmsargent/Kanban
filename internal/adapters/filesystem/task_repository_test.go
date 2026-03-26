package filesystem_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmsargent/kanban/internal/adapters/filesystem"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

func setupTasksDir(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()
	tasksDir := filepath.Join(repoRoot, ".kanban", "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	return repoRoot
}

func newTask(id string) domain.Task {
	due := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	return domain.Task{
		ID:          id,
		Title:       "Fix login bug",
		Status:      domain.StatusTodo,
		Priority:    "medium",
		Due:         &due,
		Assignee:    "alice",
		Description: "Optional description body.",
	}
}

// Behavior 1: Save creates the task file in the tasks directory
func TestSave_CreatesTaskFile(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	if err := repo.Save(repoRoot, newTask("TASK-001")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	path := filepath.Join(repoRoot, ".kanban", "tasks", "TASK-001.md")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected task file at %s: %v", path, err)
	}
}

// Behavior 2: No temp files remain after write
func TestSave_NoTempFilesRemainAfterWrite(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	if err := repo.Save(repoRoot, newTask("TASK-001")); err != nil {
		t.Fatalf("Save: %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(repoRoot, ".kanban", "tasks"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("temp file remains after write: %s", e.Name())
		}
	}
}

// Behavior 3: FindByID round-trips all field values
func TestFindByID_RoundTripsAllFields(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()
	original := newTask("TASK-001")

	if err := repo.Save(repoRoot, original); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := repo.FindByID(repoRoot, "TASK-001")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}

	if got.ID != original.ID {
		t.Errorf("ID: got %q, want %q", got.ID, original.ID)
	}
	if got.Title != original.Title {
		t.Errorf("Title: got %q, want %q", got.Title, original.Title)
	}
	if got.Status != original.Status {
		t.Errorf("Status: got %q, want %q", got.Status, original.Status)
	}
	if got.Priority != original.Priority {
		t.Errorf("Priority: got %q, want %q", got.Priority, original.Priority)
	}
	if got.Assignee != original.Assignee {
		t.Errorf("Assignee: got %q, want %q", got.Assignee, original.Assignee)
	}
	if got.Due == nil || !got.Due.Equal(*original.Due) {
		t.Errorf("Due: got %v, want %v", got.Due, original.Due)
	}
	if got.Description != original.Description {
		t.Errorf("Description: got %q, want %q", got.Description, original.Description)
	}
}

// Behavior 3b: FindByID returns ErrTaskNotFound for missing task
func TestFindByID_ReturnsErrTaskNotFound_WhenMissing(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	_, err := repo.FindByID(repoRoot, "TASK-099")
	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

// Behavior 4: Second Save with same ID fails (O_EXCL / ErrTaskAlreadyExists)
func TestSave_FailsWithDuplicateID(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()
	task := newTask("TASK-001")

	if err := repo.Save(repoRoot, task); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	err := repo.Save(repoRoot, task)
	if err == nil {
		t.Fatal("expected error on duplicate ID, got nil")
	}

	// File must not be corrupted — original content must still be readable
	got, readErr := repo.FindByID(repoRoot, "TASK-001")
	if readErr != nil {
		t.Fatalf("FindByID after duplicate save: %v", readErr)
	}
	if got.Title != task.Title {
		t.Errorf("original file overwritten: got title %q, want %q", got.Title, task.Title)
	}
}

// Behavior 5: ListAll returns all tasks
func TestListAll_ReturnsAllTasks(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("TASK-%03d", i)
		task := newTask(id)
		task.Title = fmt.Sprintf("Task %d", i)
		if err := repo.Save(repoRoot, task); err != nil {
			t.Fatalf("Save %s: %v", id, err)
		}
	}

	tasks, err := repo.ListAll(repoRoot)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(tasks) != 3 {
		t.Errorf("ListAll: got %d tasks, want 3", len(tasks))
	}
}

// Behavior 5b: ListAll returns empty slice for empty directory
func TestListAll_ReturnsEmptySlice_WhenNoTasks(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	tasks, err := repo.ListAll(repoRoot)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

// Behavior 6: Delete removes the task file
func TestDelete_RemovesTaskFile(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	if err := repo.Save(repoRoot, newTask("TASK-001")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := repo.Delete(repoRoot, "TASK-001"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByID(repoRoot, "TASK-001")
	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound after delete, got %v", err)
	}
}

// Behavior 6b: Delete returns ErrTaskNotFound for non-existent task
func TestDelete_ReturnsErrTaskNotFound_WhenMissing(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	err := repo.Delete(repoRoot, "TASK-099")
	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got %v", err)
	}
}

// Behavior 7: NextID returns sequential IDs
func TestNextID_ReturnsSequentialIDs(t *testing.T) {
	repoRoot := setupTasksDir(t)
	repo := filesystem.NewTaskRepository()

	// No existing tasks — first ID should be TASK-001
	id1, err := repo.NextID(repoRoot)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id1 != "TASK-001" {
		t.Errorf("first NextID: got %q, want %q", id1, "TASK-001")
	}

	// After saving TASK-001, next should be TASK-002
	task := newTask(id1)
	if err := repo.Save(repoRoot, task); err != nil {
		t.Fatalf("Save: %v", err)
	}

	id2, err := repo.NextID(repoRoot)
	if err != nil {
		t.Fatalf("NextID: %v", err)
	}
	if id2 != "TASK-002" {
		t.Errorf("second NextID: got %q, want %q", id2, "TASK-002")
	}
}
