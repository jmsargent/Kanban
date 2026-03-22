package cli

import (
	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
)

// NewRootCommand builds the root "kanban" cobra command and registers all
// sub-commands. Collaborators are injected so this function is testable.
func NewRootCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository, editor ports.EditFilePort) *cobra.Command {
	root := &cobra.Command{
		Use:   "kanban",
		Short: "Kanban task manager for git repositories",
		// Silence usage output on RunE errors — errors are reported by sub-commands.
		SilenceUsage: true,
	}

	root.AddCommand(NewInitCommand(git, config))
	root.AddCommand(NewCreateCommand(git, config, tasks, editor))
	root.AddCommand(NewAddCommand(git, config, tasks))
	root.AddCommand(NewBoardCommand(git, config, tasks))
	root.AddCommand(NewHookCommand())
	root.AddCommand(NewInstallHookCommand())
	root.AddCommand(NewCIDoneCommand(git, config, tasks))
	root.AddCommand(NewEditCommand(git, config, tasks, editor))
	root.AddCommand(NewDeleteCommand(git, config, tasks))
	root.AddCommand(NewStartCommand(git, config, tasks))
	root.AddCommand(NewDoneCommand(git, tasks))
	root.AddCommand(NewLogCommand(git, config, tasks))

	return root
}
