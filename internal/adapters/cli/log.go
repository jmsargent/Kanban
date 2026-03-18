package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewLogCommand builds the "kanban log <task-id>" cobra command.
// It displays the task header and its commit history derived from git log.
func NewLogCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "log <task-id>",
		Short:         "Show the git commit history for a task",
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
				return exitError("Not a git repository. Run 'kanban init' first.")
			}

			uc := usecases.NewGetTaskHistory(config, tasks, git)
			result, err := uc.Execute(repoRoot, taskID)
			if err != nil {
				if errors.Is(err, ports.ErrTaskNotFound) {
					return exitError(fmt.Sprintf("Task %s not found. Run 'kanban board' to see valid task IDs.", taskID))
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					return exitError("kanban not initialised. Run 'kanban init' first.")
				}
				return exitError(fmt.Sprintf("Error: %v", err))
			}

			writeLine(out, fmt.Sprintf("%s: %s", result.TaskID, result.Title))

			if len(result.Entries) == 0 {
				writeLine(out, "No transitions recorded yet.")
				return nil
			}

			for _, entry := range result.Entries {
				writeLine(out, fmt.Sprintf("  %s %s %s", entry.SHA[:7], entry.Author, entry.Message))
			}
			return nil
		},
	}

	return cmd
}
