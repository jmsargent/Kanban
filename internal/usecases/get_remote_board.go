package usecases

import (
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// GetRemoteBoard implements the view-remote-board use case.
// Driving port entrypoint for "GET /remote/board".
type GetRemoteBoard struct {
	githubAPI ports.GitHubAPIPort
}

// NewGetRemoteBoard constructs a GetRemoteBoard use case with the required port.
func NewGetRemoteBoard(githubAPI ports.GitHubAPIPort) *GetRemoteBoard {
	return &GetRemoteBoard{githubAPI: githubAPI}
}

// Execute fetches tasks for the given public repository and returns them as a
// Board grouped into the three standard columns: Todo, Doing, Done.
func (u *GetRemoteBoard) Execute(owner, repo string) (domain.Board, error) {
	tasks, err := u.githubAPI.ListTasks(owner, repo)
	if err != nil {
		return domain.Board{}, err
	}

	columns := []domain.Column{
		{Name: "todo", Label: "Todo"},
		{Name: "in-progress", Label: "Doing"},
		{Name: "done", Label: "Done"},
	}

	grouped := map[domain.TaskStatus][]domain.Task{
		domain.StatusTodo:       {},
		domain.StatusInProgress: {},
		domain.StatusDone:       {},
	}

	for _, t := range tasks {
		status := t.Status
		if status == "" {
			status = domain.StatusTodo
		}
		if _, ok := grouped[status]; ok {
			grouped[status] = append(grouped[status], t)
		} else {
			grouped[domain.StatusTodo] = append(grouped[domain.StatusTodo], t)
		}
	}

	return domain.Board{
		Columns: columns,
		Tasks:   grouped,
	}, nil
}