package usecases_test

import (
	"testing"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

type fakeCommitGitPort struct {
	addCalls    []string
	commitCalls []string
	pushCalls   int
	addErr      error
	commitErr   error
	pushErr     error
}

func (f *fakeCommitGitPort) Add(repoDir, path string) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.addCalls = append(f.addCalls, path)
	return nil
}

func (f *fakeCommitGitPort) Commit(repoDir, message string) error {
	if f.commitErr != nil {
		return f.commitErr
	}
	f.commitCalls = append(f.commitCalls, message)
	return nil
}

func (f *fakeCommitGitPort) Push(repoDir string) error {
	if f.pushErr != nil {
		return f.pushErr
	}
	f.pushCalls++
	return nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 1 behavior x 2 = 2 max unit tests (using 1)

// TestAddTaskAndPush_CreatesTaskAndCallsGitAddCommitPush verifies that the use
// case orchestrates: save task → git add → git commit → git push in sequence.
func TestAddTaskAndPush_CreatesTaskAndCallsGitAddCommitPush(t *testing.T) {
	repoRoot := tmpRepo(t)
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := &fakeTaskRepo{nextID: "TASK-001"}
	git := &fakeCommitGitPort{}
	input := usecases.AddTaskInput{Title: "Push me to remote", CreatedBy: "Alice"}

	uc := usecases.NewAddTaskAndPush(cfg, tasks, git)
	task, err := uc.Execute(repoRoot, input)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if task.ID != "TASK-001" {
		t.Errorf("expected task ID TASK-001, got: %s", task.ID)
	}
	if tasks.saved == nil {
		t.Error("expected task to be saved via TaskRepository")
	}
	if len(git.addCalls) != 1 {
		t.Errorf("expected 1 git add call, got: %d", len(git.addCalls))
	}
	if len(git.commitCalls) != 1 {
		t.Errorf("expected 1 git commit call, got: %d", len(git.commitCalls))
	}
	if git.pushCalls != 1 {
		t.Errorf("expected 1 git push call, got: %d", git.pushCalls)
	}
}
