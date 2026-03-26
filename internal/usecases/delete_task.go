package usecases

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/jmsargent/kanban/internal/ports"
)

// DeleteTask implements the delete-task use case.
// Driving port entrypoint for "kanban delete".
type DeleteTask struct {
	tasks ports.TaskRepository
}

// NewDeleteTask constructs a DeleteTask use case.
func NewDeleteTask(tasks ports.TaskRepository) *DeleteTask {
	return &DeleteTask{tasks: tasks}
}

// Execute finds the task, optionally prompts for confirmation, deletes the task
// file, and prints a suggested git commit command.
// Returns ErrTaskNotFound when the task does not exist.
// When force is false, reads a line from stdin and aborts unless the user enters 'y' or 'Y'.
func (u *DeleteTask) Execute(repoRoot, taskID string, force bool, stdin io.Reader, output io.Writer) error {
	task, err := u.tasks.FindByID(repoRoot, taskID)
	if err != nil {
		return err
	}

	if !force {
		_, _ = fmt.Fprintf(output, "Delete %s: %s? [y/N] ", task.ID, task.Title)
		reader := bufio.NewReader(stdin)
		line, _ := reader.ReadString('\n')
		answer := strings.TrimSpace(line)
		if answer != "y" && answer != "Y" {
			_, _ = fmt.Fprintln(output, "Deletion cancelled")
			return nil
		}
	}

	if err := u.tasks.Delete(repoRoot, taskID); err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	_, _ = fmt.Fprintf(output, "Deleted %s\n", task.ID)
	_, _ = fmt.Fprintf(output, "Suggested: git commit -m \"chore: remove %s (%s)\"\n", task.ID, task.Title)
	return nil
}
