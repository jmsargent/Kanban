package usecases_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 4)
// Behaviors:
//  1. Delete task calls Delete on repository after 'y' confirmation
//  2. Delete task returns ErrTaskNotFound when task does not exist
//  3. Delete task with force flag skips confirmation and deletes

// ─── Fakes ───────────────────────────────────────────────────────────────────

// deleteTaskRepo is a fake TaskRepository for DeleteTask tests.
type deleteTaskRepo struct {
	fakeTaskRepo
	findResult domain.Task
	findErr    error
	deleted    string
	deleteErr  error
}

func (f *deleteTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	return f.findResult, f.findErr
}

func (f *deleteTaskRepo) Delete(repoRoot, taskID string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = taskID
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestDeleteTask_DeletesTaskAndSuggestsGitCommit_WhenConfirmedWithY(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Some work"}
	repo := &deleteTaskRepo{findResult: task}
	stdin := strings.NewReader("y\n")
	var output strings.Builder

	uc := usecases.NewDeleteTask(repo)
	err := uc.Execute(repoRoot, "TASK-001", false, stdin, &output)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != "TASK-001" {
		t.Errorf("expected Delete called with TASK-001, got: %q", repo.deleted)
	}
	if !strings.Contains(output.String(), "git commit") {
		t.Errorf("expected output to suggest a git commit command, got: %q", output.String())
	}
}

func TestDeleteTask_AbortsDeletion_WhenConfirmationIsN(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Some work"}
	repo := &deleteTaskRepo{findResult: task}
	stdin := strings.NewReader("n\n")
	var output strings.Builder

	uc := usecases.NewDeleteTask(repo)
	err := uc.Execute(repoRoot, "TASK-001", false, stdin, &output)

	if err != nil {
		t.Fatalf("expected no error on abort, got: %v", err)
	}
	if repo.deleted != "" {
		t.Errorf("expected Delete NOT to be called on 'n' confirmation, but deleted: %q", repo.deleted)
	}
	if !strings.Contains(output.String(), "cancelled") {
		t.Errorf("expected output to contain 'cancelled', got: %q", output.String())
	}
}

func TestDeleteTask_DeletesWithoutPrompt_WhenForceIsTrue(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Force delete work"}
	repo := &deleteTaskRepo{findResult: task}
	stdin := strings.NewReader("") // no input provided
	var output strings.Builder

	uc := usecases.NewDeleteTask(repo)
	err := uc.Execute(repoRoot, "TASK-001", true, stdin, &output)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if repo.deleted != "TASK-001" {
		t.Errorf("expected Delete called with TASK-001, got: %q", repo.deleted)
	}
}

func TestDeleteTask_ReturnsErrTaskNotFound_WhenTaskDoesNotExist(t *testing.T) {
	repoRoot := tmpRepo(t)
	repo := &deleteTaskRepo{findErr: ports.ErrTaskNotFound}
	stdin := strings.NewReader("y\n")
	var output strings.Builder

	uc := usecases.NewDeleteTask(repo)
	err := uc.Execute(repoRoot, "TASK-999", false, stdin, &output)

	if !errors.Is(err, ports.ErrTaskNotFound) {
		t.Errorf("expected ErrTaskNotFound, got: %v", err)
	}
	if repo.deleted != "" {
		t.Error("expected Delete NOT to be called when task not found")
	}
}
