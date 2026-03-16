package usecases

import (
	"errors"
	"fmt"
	"io"
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

// TransitionToInProgress implements the commit-hook use case: scan a commit message
// for task IDs and advance matching todo tasks to in-progress.
// All errors are swallowed internally (hook must always exit 0); callers should
// write errors to hook.log before calling Execute.
type TransitionToInProgress struct {
	config ports.ConfigRepository
	tasks  ports.TaskRepository
	out    io.Writer
}

// NewTransitionToInProgress constructs a TransitionToInProgress use case.
func NewTransitionToInProgress(config ports.ConfigRepository, tasks ports.TaskRepository, out io.Writer) *TransitionToInProgress {
	return &TransitionToInProgress{config: config, tasks: tasks, out: out}
}

// Execute scans message for task ID references, and for each todo task found
// advances it to in-progress, writing a transition line to out.
// Returns nil in all cases — callers must treat a non-nil error from Execute
// as a programming error (it never occurs in practice).
func (u *TransitionToInProgress) Execute(repoRoot, message string) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			retErr = fmt.Errorf("panic in TransitionToInProgress: %v", r)
		}
	}()

	cfg, err := u.config.Read(repoRoot)
	if err != nil {
		if errors.Is(err, ports.ErrNotInitialised) {
			return nil
		}
		return nil // config errors swallowed; logged by caller
	}

	rawPattern := cfg.CITaskPattern
	if rawPattern == "" {
		rawPattern = `TASK-[0-9]+`
	}

	pattern, err := regexp.Compile(rawPattern)
	if err != nil {
		return nil
	}

	ids := pattern.FindAllString(message, -1)
	if len(ids) == 0 {
		return nil
	}

	for _, id := range ids {
		task, findErr := u.tasks.FindByID(repoRoot, id)
		if errors.Is(findErr, ports.ErrTaskNotFound) {
			_, _ = fmt.Fprintf(u.out, "Warning: task %s not found\n", id)
			continue
		}
		if findErr != nil {
			// Non-fatal: continue with remaining IDs
			continue
		}

		if task.Status != domain.StatusTodo {
			continue
		}

		prevStatus := task.Status
		task.Status = domain.StatusInProgress
		if updateErr := u.tasks.Update(repoRoot, task); updateErr != nil {
			// Non-fatal: continue with remaining IDs
			continue
		}

		_, _ = fmt.Fprintf(u.out, "kanban: %s moved  %s -> %s\n", id, prevStatus, task.Status)
	}
	return nil
}
