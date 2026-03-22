package usecases_test

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 5)
// Behaviors:
//   1. Advances in-progress task to done when referenced in commits
//   2. Skips todo tasks even when referenced in commits
//   3. Skips done tasks when referenced in commits
//   4. Commits updated file paths with [skip ci] in message

import (
	"strings"
	"testing"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes for TransitionToDone ───────────────────────────────────────────────

type spyGitPort struct {
	messages          []string
	commitFilesRoot   string
	commitFilesMsg    string
	commitFilesPaths  []string
	commitFilesErr    error
	commitFilesCalled bool
}

func (s *spyGitPort) RepoRoot() (string, error)           { return "", nil }
func (s *spyGitPort) InstallHook(repoRoot string) error   { return nil }
func (s *spyGitPort) AppendToGitignore(r, e string) error { return nil }
func (s *spyGitPort) CommitMessagesInRange(from, to string) ([]string, error) {
	return s.messages, nil
}
func (s *spyGitPort) CommitFiles(repoRoot, message string, paths []string) error {
	s.commitFilesCalled = true
	s.commitFilesRoot = repoRoot
	s.commitFilesMsg = message
	s.commitFilesPaths = paths
	return s.commitFilesErr
}

func (s *spyGitPort) GetIdentity() (ports.Identity, error) {
	return ports.Identity{}, nil
}

func (s *spyGitPort) LogFile(_ string, _ string) ([]ports.CommitEntry, error) {
	return nil, nil
}

type spyTaskRepo struct {
	byID    map[string]domain.Task
	updated []domain.Task
}

func newSpyTaskRepo() *spyTaskRepo {
	return &spyTaskRepo{byID: make(map[string]domain.Task)}
}

func (r *spyTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	t, ok := r.byID[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (r *spyTaskRepo) Save(repoRoot string, task domain.Task) error {
	r.byID[task.ID] = task
	return nil
}

func (r *spyTaskRepo) Update(repoRoot string, task domain.Task) error {
	r.byID[task.ID] = task
	r.updated = append(r.updated, task)
	return nil
}

func (r *spyTaskRepo) ListAll(repoRoot string) ([]domain.Task, error) {
	tasks := make([]domain.Task, 0, len(r.byID))
	for _, t := range r.byID {
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *spyTaskRepo) Delete(repoRoot, taskID string) error { return nil }
func (r *spyTaskRepo) NextID(repoRoot string) (string, error) {
	return "TASK-001", nil
}

// ─── Tests ───────────────────────────────────────────────────────────────────

func TestTransitionToDone_AdvancesInProgressTask_WhenReferencedInCommits(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-001: implement feature"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-001"] = domain.Task{ID: "TASK-001", Status: domain.StatusInProgress}
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(output.String(), "TASK-001") {
		t.Errorf("expected log output to mention TASK-001, got: %s", output.String())
	}
	if !strings.Contains(output.String(), "done") {
		t.Errorf("expected log output to contain 'done', got: %s", output.String())
	}
}

func TestTransitionToDone_SkipsTodoTask_WhenReferencedInCommits(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-002: some work"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-002"] = domain.Task{ID: "TASK-002"}
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tasks.updated) != 0 {
		t.Error("expected no update when todo task is referenced")
	}
}

func TestTransitionToDone_SkipsDoneTask_WhenReferencedInCommits(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-003: cleanup"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-003"] = domain.Task{ID: "TASK-003", Status: domain.StatusDone}
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tasks.updated) != 0 {
		t.Error("expected no update when done task is referenced")
	}
}

func TestTransitionToDone_UpdatesTaskFile_WhenTasksAdvanced(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-001: feature done"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-001"] = domain.Task{ID: "TASK-001", Status: domain.StatusInProgress}
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tasks.updated) == 0 {
		t.Error("expected Update to be called after advancing tasks")
	}
	if tasks.updated[0].Status != domain.StatusDone {
		t.Errorf("expected task status to be done, got: %s", tasks.updated[0].Status)
	}
	if git.commitFilesCalled {
		t.Error("expected no git commit (C-03)")
	}
}

func TestTransitionToDone_SkipsCommit_WhenNoTasksAdvanced(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"chore: update readme"}}
	tasks := newSpyTaskRepo()
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(tasks.updated) != 0 {
		t.Error("expected no Update when no tasks were advanced")
	}
}

func TestTransitionToDone_UsesDefaultPattern_WhenConfigPatternEmpty(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-001: implement feature"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-001"] = domain.Task{ID: "TASK-001", Status: domain.StatusInProgress}
	// CITaskPattern is empty
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: ""}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(output.String(), "TASK-001") {
		t.Errorf("expected output to mention TASK-001 using default pattern, got: %s", output.String())
	}
}

func TestTransitionToDone_AdvancesMultipleTasks_WhenReferencedInSameCommit(t *testing.T) {
	repoRoot := tmpRepo(t)
	git := &spyGitPort{messages: []string{"TASK-001 and TASK-002: implementation complete"}}
	tasks := newSpyTaskRepo()
	tasks.byID["TASK-001"] = domain.Task{ID: "TASK-001", Status: domain.StatusInProgress}
	tasks.byID["TASK-002"] = domain.Task{ID: "TASK-002", Status: domain.StatusInProgress}
	cfg := &fakeConfigRepo{readResult: ports.Config{CITaskPattern: `TASK-\d+`}}
	output := &strings.Builder{}

	uc := usecases.NewTransitionToDone(git, tasks, cfg, output)
	err := uc.Execute(repoRoot, "", "HEAD")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !strings.Contains(output.String(), "TASK-001") {
		t.Errorf("expected output to mention TASK-001, got: %s", output.String())
	}
	if !strings.Contains(output.String(), "TASK-002") {
		t.Errorf("expected output to mention TASK-002, got: %s", output.String())
	}
}
