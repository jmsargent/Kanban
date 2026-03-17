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

// NewCreateCommand builds the "kanban new" cobra command.
// It validates input, invokes the AddTask use case, and maps errors to exit codes.
func NewCreateCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	var priority string
	var dueStr string
	var assignee string

	cmd := &cobra.Command{
		Use:   "new <title>",
		Short: "Create a new task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			title := ""
			if len(args) > 0 {
				title = args[0]
			}

			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			identity, err := git.GetIdentity()
			if err != nil {
				fmt.Fprintln(os.Stderr, "git identity not configured — run: git config --global user.name \"Your Name\"")
				os.Exit(1)
			}

			var due *time.Time
			if dueStr != "" {
				parsed, parseErr := time.Parse("2006-01-02", dueStr)
				if parseErr != nil {
					fmt.Fprintf(os.Stderr, "Invalid due date format (expected YYYY-MM-DD): %s\n", dueStr)
					os.Exit(2)
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
					os.Exit(2)
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					fmt.Fprintln(os.Stderr, "kanban not initialised — run 'kanban init' first")
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Created %s: %s\n", task.ID, task.Title)
			fmt.Printf("Hint: reference %s in your next commit to start tracking\n", task.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&priority, "priority", "", "Task priority (e.g. P1, P2)")
	cmd.Flags().StringVar(&dueStr, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assigned team member")

	return cmd
}

// unwrapMessage extracts the human-readable message from a wrapped error.
func unwrapMessage(err error) string {
	// ErrInvalidInput is wrapped as "invalid input: <message>"
	// Extract the domain message for display.
	inner := errors.Unwrap(err)
	if inner != nil {
		return inner.Error()
	}
	return err.Error()
}
