package usecases_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// Test Budget: 2 behaviors × 2 = 4 max unit tests. Using 4.
// Behavior 1: tasks returned from port are grouped into three-column board.
//   - should return three columns
//   - should place each task in the correct column by status
// Behavior 2: port error propagates as use-case error.
//   - should propagate ErrRepositoryNotFound unchanged

// stubGitHubAPIPort is a test double for the GitHubAPIPort driven port.
type stubGitHubAPIPort struct {
	tasks []domain.Task
	err   error
}

func (s *stubGitHubAPIPort) ListTasks(_, _ string) ([]domain.Task, error) {
	return s.tasks, s.err
}

func threeTaskStub() *stubGitHubAPIPort {
	return &stubGitHubAPIPort{
		tasks: []domain.Task{
			{ID: "TASK-001", Title: "Write release notes", Status: domain.StatusTodo},
			{ID: "TASK-002", Title: "Fix scheduler bug", Status: domain.StatusInProgress},
			{ID: "TASK-003", Title: "Close milestone", Status: domain.StatusDone},
		},
	}
}

func TestGetRemoteBoard_ShouldReturnThreeColumns(t *testing.T) {
	uc := usecases.NewGetRemoteBoard(threeTaskStub())

	board, err := uc.Execute("torvalds", "linux")

	require.NoError(t, err)
	assert.Len(t, board.Columns, 3)
}

func TestGetRemoteBoard_ShouldPlaceEachTaskInCorrectColumnByStatus(t *testing.T) {
	uc := usecases.NewGetRemoteBoard(threeTaskStub())

	board, err := uc.Execute("torvalds", "linux")

	require.NoError(t, err)
	require.Len(t, board.Tasks[domain.StatusTodo], 1)
	assert.Equal(t, "Write release notes", board.Tasks[domain.StatusTodo][0].Title)
	require.Len(t, board.Tasks[domain.StatusInProgress], 1)
	assert.Equal(t, "Fix scheduler bug", board.Tasks[domain.StatusInProgress][0].Title)
	require.Len(t, board.Tasks[domain.StatusDone], 1)
	assert.Equal(t, "Close milestone", board.Tasks[domain.StatusDone][0].Title)
}

func TestGetRemoteBoard_ShouldPropagatePortError(t *testing.T) {
	stub := &stubGitHubAPIPort{err: ports.ErrRepositoryNotFound}
	uc := usecases.NewGetRemoteBoard(stub)

	_, err := uc.Execute("no-such", "repo")

	assert.ErrorIs(t, err, ports.ErrRepositoryNotFound)
}
