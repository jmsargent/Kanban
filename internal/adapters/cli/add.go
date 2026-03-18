package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewAddCommand builds the "kanban add" cobra command.
// It is an alternative interface to "kanban new" that accepts the title via
// the required -t/--title flag, enabling scripted and pipeline usage.
func NewAddCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	var title string
	var priority string
	var dueStr string
	var assignee string

	cmd := &cobra.Command{
		Use:           "add",
		Short:         "Create a new task (alias for new, accepts -t flag)",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				osExit(1)
				return nil
			}

			identity, err := git.GetIdentity()
			if err != nil {
				fmt.Fprintln(os.Stderr, "git identity not configured — run: git config --global user.name \"Your Name\"")
				osExit(1)
				return nil
			}

			var due *time.Time
			if dueStr != "" {
				parsed, parseErr := time.Parse("2006-01-02", dueStr)
				if parseErr != nil {
					fmt.Fprintf(os.Stderr, "Invalid due date format (expected YYYY-MM-DD): %s\n", dueStr)
					osExit(2)
					return nil
				}
				due = &parsed
			}

			input := usecases.AddTaskInput{
				Title:     title,
				Priority:  priority,
				Due:       due,
				Assignee:  assignee,
				CreatedBy: identity.Name,
			}

			uc := usecases.NewAddTask(config, tasks)
			task, err := uc.Execute(repoRoot, input)
			if err != nil {
				if errors.Is(err, ports.ErrInvalidInput) {
					fmt.Fprintf(os.Stderr, "%s\n", unwrapMessage(err))
					osExit(2)
					return nil
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					fmt.Fprintln(os.Stderr, "kanban not initialised — run 'kanban init' first")
					osExit(1)
					return nil
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				osExit(1)
				return nil
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created %s: %s\n", task.ID, task.Title)
			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Task title (required)")
	cmd.Flags().StringVar(&priority, "priority", "", "Task priority (e.g. P1, P2)")
	cmd.Flags().StringVar(&dueStr, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assigned team member")

	return cmd
}
