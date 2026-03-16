package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// errCommandFailed is a sentinel returned by RunE to signal a non-zero exit
// without printing an additional error message (cobra would otherwise print it).
var errCommandFailed = errors.New("exit1")

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

			repoRoot, err := git.RepoRoot()
			if err != nil {
				writeLine(errOut, "Not a git repository")
				return errCommandFailed
			}

			uc := usecases.NewStartTask(config, tasks)
			result, err := uc.Execute(repoRoot, taskID)
			if err != nil {
				if errors.Is(err, ports.ErrInvalidTransition) {
					writeLine(errOut, fmt.Sprintf("Task %s is already finished", taskID))
					return errCommandFailed
				}
				if errors.Is(err, ports.ErrTaskNotFound) {
					writeLine(errOut, fmt.Sprintf("Task %s not found", taskID))
					return errCommandFailed
				}
				if errors.Is(err, ports.ErrNotInitialised) {
					writeLine(errOut, "kanban not initialised — run 'kanban init' first")
					return errCommandFailed
				}
				writeLine(errOut, fmt.Sprintf("Error: %v", err))
				return errCommandFailed
			}

			if result.AlreadyInProgress {
				writeLine(out, fmt.Sprintf("Task %s is already in progress", taskID))
				return nil
			}

			writeLine(out, fmt.Sprintf("Started %s: %s", taskID, result.Task.Title))
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
