package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// NewDoneCommand builds the "kanban done" cobra command.
// It accepts a single task ID, delegates to the CompleteTask use case, and maps
// each result or error to the correct stdout/stderr message and exit code.
// The command never invokes git commit or git add (C-03 compliance).
func NewDoneCommand(git ports.GitPort, tasks ports.TaskRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "done <task-id>",
		Short:         "Mark a task as done (transitions any status -> done)",
		Args:          cobra.ExactArgs(1),
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()

			exitError := func(msg string) error {
				writeLine(errOut, msg)
				osExit(1)
				return nil
			}

			repoRoot, err := git.RepoRoot()
			if err != nil {
				return exitError("Not a git repository")
			}

			uc := usecases.NewCompleteTask(tasks)
			result, err := uc.Execute(repoRoot, taskID)
			if err != nil {
				if errors.Is(err, ports.ErrTaskNotFound) {
					return exitError(fmt.Sprintf("kanban: %s not found", taskID))
				}
				return exitError(fmt.Sprintf("kanban: %v", err))
			}

			if result.AlreadyDone {
				writeLine(out, fmt.Sprintf("kanban: %s already done", taskID))
				return nil
			}

			writeLine(out, fmt.Sprintf("kanban: %s moved %s -> done", taskID, result.From))
			return nil
		},
	}

	return cmd
}
