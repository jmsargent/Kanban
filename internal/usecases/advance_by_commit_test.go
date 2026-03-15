package usecases_test

import (
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// Test Budget: 3 behaviors x 2 = 6 max unit tests (using 3)
// Behaviors:
//   1. AdvanceByCommitMessage transitions a todo task to in-progress when referenced
//   2. AdvanceByCommitMessage skips tasks that are not in todo status
//   3. AdvanceByCommitMessage is a no-op when no task IDs appear in the message

func TestAdvanceByCommitMessage_TransitionsTodoTask_WhenReferenced(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}

	uc := usecases.NewTransitionTask(cfg, tasks)
	err := uc.AdvanceByCommitMessage(repoRoot, "TASK-001: start OAuth work")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated == nil {
		t.Fatal("expected task to be updated via TaskRepository")
	}
	if tasks.updated.Status != domain.StatusInProgress {
		t.Errorf("expected status in-progress, got: %s", tasks.updated.Status)
	}
}

func TestAdvanceByCommitMessage_SkipsTask_WhenNotTodo(t *testing.T) {
	repoRoot := tmpRepo(t)
	inProgress := domain.Task{ID: "TASK-002", Title: "Active work", Status: domain.StatusInProgress}
	done := domain.Task{ID: "TASK-003", Title: "Finished", Status: domain.StatusDone}
	tasks := newTransitionTaskRepo(inProgress, done)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}

	uc := usecases.NewTransitionTask(cfg, tasks)
	err := uc.AdvanceByCommitMessage(repoRoot, "TASK-002: more work TASK-003: polish")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Errorf("expected no update for non-todo tasks, but got: %+v", *tasks.updated)
	}
}

func TestAdvanceByCommitMessage_IsNoOp_WhenNoTaskIDsInMessage(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Status: domain.StatusTodo}
	tasks := newTransitionTaskRepo(task)
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}

	uc := usecases.NewTransitionTask(cfg, tasks)
	err := uc.AdvanceByCommitMessage(repoRoot, "chore: fix typo in README")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if tasks.updated != nil {
		t.Error("expected no update when message has no task reference")
	}
}
