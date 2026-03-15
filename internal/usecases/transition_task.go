package usecases

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// TransitionTask implements the status-transition use case driven by git events.
type TransitionTask struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
}

// NewTransitionTask constructs a TransitionTask use case.
func NewTransitionTask(config ports.ConfigRepository, tasks ports.TaskRepository) *TransitionTask {
	return &TransitionTask{config: config, tasks: tasks}
}

// AdvanceByCommitMessage scans the commit message for TASK-NNN references,
// finds those tasks, and advances each eligible task one status step.
// Tasks already at the target status or in an invalid transition are skipped.
func (u *TransitionTask) AdvanceByCommitMessage(repoRoot, message string) error {
	cfg, err := u.config.Read(repoRoot)
	if err != nil {
		if errors.Is(err, ports.ErrNotInitialised) {
			return nil // not initialised; hook is a no-op
		}
		return fmt.Errorf("read config: %w", err)
	}
	_ = cfg

	pattern := regexp.MustCompile(`TASK-\d+`)
	ids := pattern.FindAllString(message, -1)
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		task, findErr := u.tasks.FindByID(repoRoot, id)
		if errors.Is(findErr, ports.ErrTaskNotFound) {
			continue
		}
		if findErr != nil {
			return fmt.Errorf("find task %s: %w", id, findErr)
		}

		// The commit-msg hook only handles todo → in-progress.
		// In-progress → done is managed exclusively by the CI pipeline (kanban ci-done).
		if task.Status != domain.StatusTodo {
			continue
		}
		next := domain.StatusInProgress

		if !domain.CanTransitionTo(task.Status, next) {
			continue
		}

		task.Status = next
		if updateErr := u.tasks.Update(repoRoot, task); updateErr != nil {
			return fmt.Errorf("update task %s: %w", id, updateErr)
		}
	}
	return nil
}
