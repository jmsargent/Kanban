package cli

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// NewCreateCommand builds the "kanban new" cobra command.
// With no positional argument it opens $EDITOR with a blank task template.
// With a positional argument it creates the task immediately.
func NewCreateCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, editor ports.EditFilePort) *cobra.Command {
	var priority string
	var dueStr string
	var assignee string

	cmd := &cobra.Command{
		Use:   "new [title]",
		Short: "Create a new task",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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

			if len(args) == 0 {
				return runEditorMode(repoRoot, identity.Name, config, tasks, editor)
			}

			due, parseErr := parseDueDate(dueStr)
			if parseErr != nil {
				fmt.Fprintln(os.Stderr, parseErr)
				os.Exit(2)
			}

			input := usecases.AddTaskInput{
				Title:     args[0],
				Priority:  priority,
				Due:       due,
				Assignee:  assignee,
				CreatedBy: identity.Name,
			}
			return runTitleMode(repoRoot, input, config, tasks)
		},
	}

	cmd.Flags().StringVar(&priority, "priority", "", "Task priority (e.g. P1, P2)")
	cmd.Flags().StringVar(&dueStr, "due", "", "Due date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&assignee, "assignee", "", "Assigned team member")

	return cmd
}

// runEditorMode handles "kanban new" with no positional argument:
// opens $EDITOR with a blank task template and persists the result.
func runEditorMode(repoRoot, createdBy string, config ports.ConfigRepository, tasks ports.TaskRepository, editor ports.EditFilePort) error {
	uc := usecases.NewEditorTaskUseCase(config, tasks, editor)
	task, err := uc.Execute(repoRoot, createdBy)
	if err != nil {
		exitOnNewCommandError(err)
	}
	printTaskCreated(task.ID, task.Title)
	return nil
}

// runTitleMode handles "kanban new <title>" with a positional title argument:
// creates the task immediately without opening an editor.
func runTitleMode(repoRoot string, input usecases.AddTaskInput, config ports.ConfigRepository, tasks ports.TaskRepository) error {
	uc := usecases.NewAddTask(config, tasks)
	task, err := uc.Execute(repoRoot, input)
	if err != nil {
		exitOnNewCommandError(err)
	}
	printTaskCreated(task.ID, task.Title)
	return nil
}

// parseDueDate parses an optional due date string in YYYY-MM-DD format.
// Returns nil when dueStr is empty.
func parseDueDate(dueStr string) (*time.Time, error) {
	if dueStr == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", dueStr)
	if err != nil {
		return nil, fmt.Errorf("invalid due date format (expected YYYY-MM-DD): %s", dueStr)
	}
	return &parsed, nil
}

// exitOnNewCommandError maps domain errors to the appropriate stderr message
// and exit code for the "kanban new" command.
func exitOnNewCommandError(err error) {
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

// printTaskCreated prints the standard task creation confirmation lines.
func printTaskCreated(id, title string) {
	fmt.Printf("Created %s: %s\n", id, title)
	fmt.Printf("Hint: reference %s in your next commit to start tracking\n", id)
}

// unwrapMessage extracts the human-readable message from a wrapped error.
// Returns the full error string so the domain message (e.g. "title cannot be
// empty") is always present in the output.
func unwrapMessage(err error) string {
	return err.Error()
}
