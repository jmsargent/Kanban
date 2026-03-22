package usecases_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// ─── Fakes ───────────────────────────────────────────────────────────────────

// historyFakeTaskRepo satisfies ports.TaskRepository for GetTaskHistory tests.
type historyFakeTaskRepo struct {
	byID    map[string]domain.Task
	findErr error
}

func newHistoryFakeTaskRepo(tasks ...domain.Task) *historyFakeTaskRepo {
	repo := &historyFakeTaskRepo{byID: make(map[string]domain.Task)}
	for _, t := range tasks {
		repo.byID[t.ID] = t
	}
	return repo
}

func (f *historyFakeTaskRepo) FindByID(repoRoot, taskID string) (domain.Task, error) {
	if f.findErr != nil {
		return domain.Task{}, f.findErr
	}
	t, ok := f.byID[taskID]
	if !ok {
		return domain.Task{}, ports.ErrTaskNotFound
	}
	return t, nil
}

func (f *historyFakeTaskRepo) Save(repoRoot string, task domain.Task) error  { return nil }
func (f *historyFakeTaskRepo) ListAll(repoRoot string) ([]domain.Task, error) { return nil, nil }
func (f *historyFakeTaskRepo) Update(repoRoot string, task domain.Task) error { return nil }
func (f *historyFakeTaskRepo) Delete(repoRoot, taskID string) error           { return nil }
func (f *historyFakeTaskRepo) NextID(repoRoot string) (string, error)         { return "", nil }

// historyFakeLog satisfies ports.TransitionLogRepository for GetTaskHistory tests.
type historyFakeLog struct {
	entries []domain.TransitionEntry
}

func (f *historyFakeLog) Append(_ string, _ domain.TransitionEntry) error { return nil }
func (f *historyFakeLog) LatestStatus(_, _ string) (domain.TaskStatus, error) {
	return domain.StatusTodo, nil
}
func (f *historyFakeLog) History(_, _ string) ([]domain.TransitionEntry, error) {
	return f.entries, nil
}

// historyFakeGitPort satisfies ports.GitPort for GetTaskHistory tests.
// It records LogFile calls and returns configured entries.
type historyFakeGitPort struct {
	logFileEntries []ports.CommitEntry
	logFileErr     error
	logFilePath    string // records the filePath passed to LogFile
}

func (f *historyFakeGitPort) RepoRoot() (string, error) { return "", nil }
func (f *historyFakeGitPort) CommitMessagesInRange(from, to string) ([]string, error) {
	return nil, nil
}
func (f *historyFakeGitPort) AppendToGitignore(repoRoot, entry string) error { return nil }
func (f *historyFakeGitPort) GetIdentity() (ports.Identity, error) {
	return ports.Identity{}, nil
}
func (f *historyFakeGitPort) LogFile(repoRoot, filePath string) ([]ports.CommitEntry, error) {
	f.logFilePath = filePath
	return f.logFileEntries, f.logFileErr
}

// ─── Tests ───────────────────────────────────────────────────────────────────

// Test Budget: 4 behaviors x 2 = 8 max unit tests (using 4)

func TestGetTaskHistory_ReturnsHeaderAndEntries_WhenTaskHasCommits(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-001", Title: "Fix OAuth login bug", Status: domain.StatusInProgress}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newHistoryFakeTaskRepo(task)
	ts := time.Date(2026, 3, 18, 10, 0, 0, 0, time.UTC)
	git := &historyFakeGitPort{
		logFileEntries: []ports.CommitEntry{
			{SHA: "abc1234", Timestamp: ts, Author: "alice@example.com", Message: "TASK-001: start work"},
		},
	}

	uc := usecases.NewGetTaskHistory(cfg, tasks, git)
	result, err := uc.Execute(repoRoot, "TASK-001")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.TaskID != "TASK-001" {
		t.Errorf("expected TaskID TASK-001, got: %s", result.TaskID)
	}
	if result.Title != "Fix OAuth login bug" {
		t.Errorf("expected title 'Fix OAuth login bug', got: %s", result.Title)
	}
	if len(result.Entries) != 1 {
		t.Fatalf("expected 1 entry, got: %d", len(result.Entries))
	}
	if result.Entries[0].SHA != "abc1234" {
		t.Errorf("expected entry SHA abc1234, got: %s", result.Entries[0].SHA)
	}
}

func TestGetTaskHistory_ReturnsEmptyEntries_WhenTaskHasNoCommits(t *testing.T) {
	repoRoot := tmpRepo(t)
	task := domain.Task{ID: "TASK-002", Title: "Write release notes", Status: domain.StatusTodo}
	cfg := &fakeConfigRepo{readResult: ports.Config{
		Columns: []domain.Column{{Name: "todo", Label: "TODO"}},
	}}
	tasks := newHistoryFakeTaskRepo(task)
	git := &historyFakeGitPort{logFileEntries: nil}

	uc := usecases.NewGetTaskHistory(cfg, tasks, git)
	result, err := uc.Execute(repoRoot, "TASK-002")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if result.TaskID != "TASK-002" {
		t.Errorf("expected TaskID TASK-002, got: %s", result.TaskID)
	}
	if len(result.Entries) != 0 {
		t.Errorf("expected 0 entries, got: %d", len(result.Entries))
	}
}

func TestGetTaskHistory_ReturnsError_ForInvalidPreconditions(t *testing.T) {
	// Two error paths share the same assertion shape: Execute returns a sentinel error.
	// Parametrized as input variations of one behavior (Mandate 5).
	cases := []struct {
		name    string
		cfg     ports.ConfigRepository
		task    *domain.Task
		taskID  string
		wantErr error
	}{
		{
			name:    "missing task returns ErrTaskNotFound",
			cfg:     &fakeConfigRepo{readResult: ports.Config{Columns: []domain.Column{{Name: "todo", Label: "TODO"}}}},
			task:    nil,
			taskID:  "TASK-999",
			wantErr: ports.ErrTaskNotFound,
		},
		{
			name:    "uninitialised repo returns ErrNotInitialised",
			cfg:     newFreshConfigRepo(),
			task:    nil,
			taskID:  "TASK-001",
			wantErr: ports.ErrNotInitialised,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repoRoot := tmpRepo(t)
			var tasks *historyFakeTaskRepo
			if tc.task != nil {
				tasks = newHistoryFakeTaskRepo(*tc.task)
			} else {
				tasks = newHistoryFakeTaskRepo()
			}
			git := &historyFakeGitPort{}

			uc := usecases.NewGetTaskHistory(tc.cfg, tasks, git)
			_, err := uc.Execute(repoRoot, tc.taskID)

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("expected %v, got: %v", tc.wantErr, err)
			}
		})
	}
}
