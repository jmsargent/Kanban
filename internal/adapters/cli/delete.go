package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewDeleteCommand builds the "kanban delete" cobra command.
func NewDeleteCommand(git ports.GitPort, _ ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <task-id>",
		Short: "Remove a task from the board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			uc := usecases.NewDeleteTask(tasks)
			if err := uc.Execute(repoRoot, taskID, force, os.Stdin, os.Stdout); err != nil {
				if errors.Is(err, ports.ErrTaskNotFound) {
					fmt.Fprintf(os.Stderr, "task-not-found: %s\n", taskID)
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")
	return cmd
}
