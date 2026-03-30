package usecases

import (
	"fmt"
	"sort"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
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
// filterAssignee restricts the board to tasks assigned to that value; pass an
// empty string to show all tasks. Status is read directly from the task's YAML
// status field — YAML is the authoritative state source.
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

		status := t.Status
		if status == "" {
			status = domain.StatusTodo
		}

		grouped[status] = append(grouped[status], t)
	}

	for status, tasks := range grouped {
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
		})
		grouped[status] = tasks
	}

	return domain.Board{
		Columns:         cfg.Columns,
		Tasks:           grouped,
		UnassignedCount: unassignedCount,
	}, nil
}
