package usecases

import (
	"fmt"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// GetBoard implements the view-board use case.
// Driving port entrypoint for "kanban board".
type GetBoard struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
}

// NewGetBoard constructs a GetBoard use case with its required collaborators.
func NewGetBoard(config ports.ConfigRepository, tasks ports.TaskRepository) *GetBoard {
	return &GetBoard{config: config, tasks: tasks}
}

// Execute reads the configuration and all tasks, then groups them into a Board.
func (u *GetBoard) Execute(repoRoot string) (domain.Board, error) {
	cfg, err := u.config.Read(repoRoot)
	if err != nil {
		return domain.Board{}, fmt.Errorf("read config: %w", err)
	}

	allTasks, err := u.tasks.ListAll(repoRoot)
	if err != nil {
		return domain.Board{}, fmt.Errorf("list tasks: %w", err)
	}

	grouped := make(map[domain.TaskStatus][]domain.Task)
	for _, col := range cfg.Columns {
		grouped[domain.TaskStatus(col.Name)] = []domain.Task{}
	}
	for _, t := range allTasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	return domain.Board{
		Columns: cfg.Columns,
		Tasks:   grouped,
	}, nil
}
