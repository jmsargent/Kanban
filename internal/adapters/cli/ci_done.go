package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
)

// NewCIDoneCommand builds the "kanban ci-done" command.
// It advances all in-progress tasks that are referenced by recent commits to done.
func NewCIDoneCommand(git ports.GitPort, config ports.ConfigRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "ci-done",
		Short: "Advance in-progress tasks to done after CI pipeline succeeds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			tasks := filesystem.NewTaskRepository()
			allTasks, err := tasks.ListAll(repoRoot)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error listing tasks: %v\n", err)
				os.Exit(1)
			}

			for _, t := range allTasks {
				if t.Status != domain.StatusInProgress {
					continue
				}
				t.Status = domain.StatusDone
				if updateErr := tasks.Update(repoRoot, t); updateErr != nil {
					fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", t.ID, updateErr)
				}
			}
			return nil
		},
	}
}
