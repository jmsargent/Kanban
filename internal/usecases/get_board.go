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
	log    ports.TransitionLogRepository
}

// NewGetBoard constructs a GetBoard use case with its required collaborators.
func NewGetBoard(config ports.ConfigRepository, tasks ports.TaskRepository, log ports.TransitionLogRepository) *GetBoard {
	return &GetBoard{config: config, tasks: tasks, log: log}
}

// Execute reads the configuration and all tasks, then groups them into a Board.
// filterAssignee restricts the board to tasks assigned to that value; pass an
// empty string to show all tasks. Status is resolved from TransitionLogRepository,
// not from the task's YAML status field.
func (u *GetBoard) Execute(repoRoot string, filterAssignee string) (domain.Board, error) {
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

	unassignedCount := 0
	for _, t := range allTasks {
		if filterAssignee != "" && t.Assignee != filterAssignee {
			if t.Assignee == "" {
				unassignedCount++
			}
			continue
		}

		status, err := u.log.LatestStatus(repoRoot, t.ID)
		if err != nil {
			return domain.Board{}, fmt.Errorf("get status for %s: %w", t.ID, err)
		}

		grouped[status] = append(grouped[status], t)
	}

	return domain.Board{
		Columns:         cfg.Columns,
		Tasks:           grouped,
		UnassignedCount: unassignedCount,
	}, nil
}
