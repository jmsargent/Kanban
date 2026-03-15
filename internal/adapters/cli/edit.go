package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewEditCommand builds the "kanban edit" cobra command.
func NewEditCommand(git ports.GitPort, _ ports.ConfigRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "edit <task-id>",
		Short: "Edit a task in $EDITOR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]

			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			tasks := filesystem.NewTaskRepository()
			uc := usecases.NewEditTask(tasks)
			diff, err := uc.Execute(repoRoot, taskID)
			if err != nil {
				if errors.Is(err, ports.ErrTaskNotFound) {
					fmt.Fprintf(os.Stderr, "task-not-found: %s\n", taskID)
					os.Exit(1)
				}
				if errors.Is(err, ports.ErrInvalidInput) {
					fmt.Fprintf(os.Stderr, "%s\n", unwrapMessage(err))
					os.Exit(2)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if diff.NoChanges {
				fmt.Fprintln(os.Stdout, "No changes")
				return nil
			}

			fmt.Fprintf(os.Stdout, "Updated %s: changed %s\n", taskID, strings.Join(diff.ChangedFields, ", "))
			return nil
		},
	}
}
