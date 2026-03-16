package usecases_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 5)
// Behaviors:
//  1. Edit task returns diff with changed fields when fields are modified
//  2. Edit task returns ErrTaskNotFound when task does not exist
//  3. Edit task rejects empty title after editor clears it

// ─── Fakes ───────────────────────────────────────────────────────────────────

// editTaskRepo is a fake TaskRepository for EditTask tests.
// It pre-populates FindByID to return a known task and records Update calls.
type editTaskRepo struct {
	fakeTaskRepo
	findResult domain.Task
	findErr    error
	updated    *domain.Task
	updateErr  error
}

func (f *editTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	return f.findResult, f.findErr
}

func (f *editTaskRepo) Update(repoRoot string, task domain.Task) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	f.updated = &task
	return nil
}

// mockEditorScript creates a temporary shell script that replaces title in the
// temp file. Returns the script path and a cleanup function.
func mockEditorScript(t *testing.T, newTitle string) string {
	t.Helper()
	dir := t.TempDir()
	script := filepath.Join(dir, "editor.sh")
	content := fmt.Sprintf("#!/bin/sh\nsed -i.bak 's/^title: .*/title: %s/' \"$1\"\n", newTitle)
	if err := os.WriteFile(script, []byte(content), 0755); err != nil {
		t.Fatalf("write mock editor: %v", err)
	}
	return script
}

// editFileStub is a fake EditFilePort that performs real temp file I/O so that
// shell script editors used in tests can read and modify the YAML file.
type editFileStub struct{}

type editFieldsYAML struct {
	Title       string `yaml:"title"`
	Priority    string `yaml:"priority"`
	Due         string `yaml:"due"`
	Assignee    string `yaml:"assignee"`
	Description string `yaml:"description"`
}

func (s *editFileStub) WriteTemp(task domain.Task) (string, error) {
	f, err := os.CreateTemp("", "kanban-edit-*.yaml")
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()
	due := ""
	if task.Due != nil {
		due = task.Due.Format("2006-01-02")
	}
	ef := editFieldsYAML{
		Title:       task.Title,
		Priority:    task.Priority,
		Due:         due,
		Assignee:    task.Assignee,
		Description: task.Description,
	}
	data, err := yaml.Marshal(ef)
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	return f.Name(), err
}

func (s *editFileStub) ReadTemp(path string) (ports.EditSnapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ports.EditSnapshot{}, err
	}
	var ef editFieldsYAML
	if err := yaml.Unmarshal(data, &ef); err != nil {
		return ports.EditSnapshot{}, err
	}
	return ports.EditSnapshot{
		Title:       ef.Title,
		Priority:    ef.Priority,
		Due:         ef.Due,
		Assignee:    ef.Assignee,
		Description: ef.Description,
	}, nil
}

// ensure editFileStub satisfies the interface at compile time
var _ ports.EditFilePort = (*editFileStub)(nil)

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestEditTask_ReturnsDiffWithChangedField_WhenTitleIsUpdated(t *testing.T) {
	repoRoot := tmpRepo(t)
	original := domain.Task{
		ID:    "TASK-001",
		Title: "Old title",
	}
	repo := &editTaskRepo{findResult: original}
	editorScript := mockEditorScript(t, "New title")

	t.Setenv("EDITOR", editorScript)
	uc := usecases.NewEditTask(repo, &editFileStub{})
	diff, err := uc.Execute(repoRoot, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.updated == nil {
		t.Fatal("expected Update to be called with changed task")
	}
	if repo.updated.Title != "New title" {
		t.Errorf("expected updated title 'New title', got: %q", repo.updated.Title)
	}
	if diff.Before.Title != "Old title" {
		t.Errorf("expected diff.Before.Title 'Old title', got: %q", diff.Before.Title)
	}
	if diff.After.Title != "New title" {
		t.Errorf("expected diff.After.Title 'New title', got: %q", diff.After.Title)
	}
}

func TestEditTask_ReturnsNoChangesResult_WhenEditorMakesNoChange(t *testing.T) {
	repoRoot := tmpRepo(t)
	original := domain.Task{
		ID:    "TASK-001",
		Title: "Same title",
	}
	repo := &editTaskRepo{findResult: original}

	// Editor that does nothing (cat / true)
	dir := t.TempDir()
	script := filepath.Join(dir, "noop.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n: # no-op\n"), 0755); err != nil {
		t.Fatalf("write noop editor: %v", err)
	}
	t.Setenv("EDITOR", script)

	uc := usecases.NewEditTask(repo, &editFileStub{})
	diff, err := uc.Execute(repoRoot, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !diff.NoChanges {
		t.Error("expected NoChanges true when editor makes no change")
	}
	if repo.updated != nil {
		t.Error("expected Update NOT to be called when there are no changes")
	}
}

func TestEditTask_ReturnsErrTaskNotFound_WhenTaskDoesNotExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	repo := &editTaskRepo{findErr: ports.ErrTaskNotFound}

	uc := usecases.NewEditTask(repo, &editFileStub{})
	_, err := uc.Execute(repoRoot, "TASK-999")

	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
}

func TestEditTask_ReturnsError_WhenEditorClearsTitle(t *testing.T) {
	repoRoot := tmpRepo(t)
	original := domain.Task{ID: "TASK-001", Title: "Has a title"}
	repo := &editTaskRepo{findResult: original}

	dir := t.TempDir()
	script := filepath.Join(dir, "clear-title.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsed -i.bak 's/^title: .*/title: /' \"$1\"\n"), 0755); err != nil {
		t.Fatalf("write clear-title editor: %v", err)
	}
	t.Setenv("EDITOR", script)

	uc := usecases.NewEditTask(repo, &editFileStub{})
	_, err := uc.Execute(repoRoot, "TASK-001")

	if err == nil {
		t.Fatal("expected error when editor clears the title, got nil")
	}
	if repo.updated != nil {
		t.Error("expected Update NOT to be called when title is invalid")
	}
}

// Verify that the output of EditTask.ChangedFields lists the field name.
func TestEditTask_ChangedFields_IncludesTitleWhenTitleChanged(t *testing.T) {
	repoRoot := tmpRepo(t)
	original := domain.Task{ID: "TASK-001", Title: "Before"}
	repo := &editTaskRepo{findResult: original}
	editorScript := mockEditorScript(t, "After")

	t.Setenv("EDITOR", editorScript)
	uc := usecases.NewEditTask(repo, &editFileStub{})
	diff, err := uc.Execute(repoRoot, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(strings.Join(diff.ChangedFields, " "), "title") {
		t.Errorf("expected changed fields to include 'title', got: %v", diff.ChangedFields)
	}
}
