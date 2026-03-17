package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// osExit is a variable so tests can override it to capture exit-code intent
// without terminating the test process. Defaults to os.Exit in production.
var osExit = os.Exit

// SetOsExit replaces the exit function used by the start command. Pass nil to
// restore the default os.Exit. Intended for use in tests only.
func SetOsExit(fn func(int)) {
	if fn == nil {
		osExit = os.Exit
		return
	}
	osExit = fn
}

// NewStartCommand builds the "kanban start" cobra command.
// It accepts a single task ID, delegates to the StartTask use case, and maps
// each result or error to the correct stdout/stderr message and exit code.
func NewStartCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "start <task-id>",
		Short:         "Start working on a task (transitions todo -> in-progress)",
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

			identity, err := git.GetIdentity()
			if err != nil {
				return exitError("git identity not configured — run: git config --global user.name \"Your Name\"")
			}

			uc := usecases.NewStartTask(config, tasks)
			result, err := uc.Execute(repoRoot, taskID, identity.Name)
			if err != nil {
				if errors.Is(err, ports.ErrInvalidTransition) {
					return exitError(fmt.Sprintf("Task %s is already finished", taskID))
				}
				if errors.Is(err, ports.ErrTaskNotFound) {
					return exitError(fmt.Sprintf("Task %s not found", taskID))
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					return exitError("kanban not initialised — run 'kanban init' first")
				}
				return exitError(fmt.Sprintf("Error: %v", err))
			}

			if result.AlreadyInProgress {
				writeLine(out, fmt.Sprintf("Task %s is already in progress", taskID))
				return nil
			}

			writeLine(out, fmt.Sprintf("Started %s: %s", taskID, result.Task.Title))
			if result.PreviousAssignee != "" {
				writeLine(out, fmt.Sprintf("Note: task was previously assigned to %s", result.PreviousAssignee))
			}
			return nil
		},
	}

	return cmd
}

// writeLine writes a message followed by a newline to the given writer,
// discarding any write error (CLI output failures are non-actionable).
func writeLine(w io.Writer, msg string) {
	_, _ = fmt.Fprintln(w, msg)
}
