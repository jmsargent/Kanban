package cli

import (
	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
)

// NewRootCommand builds the root "kanban" cobra command and registers all
// sub-commands. Collaborators are injected so this function is testable.
func NewRootCommand(git ports.GitPort, config ports.ConfigRepository) *cobra.Command {
	root := &cobra.Command{
		Use:   "kanban",
		Short: "Kanban task manager for git repositories",
		// Silence usage output on RunE errors — errors are reported by sub-commands.
		SilenceUsage: true,
	}

	root.AddCommand(NewInitCommand(git, config))
	root.AddCommand(NewAddCommand(git, config))
	root.AddCommand(NewBoardCommand(git, config))
	root.AddCommand(NewHookCommand(git, config))
	root.AddCommand(NewCIDoneCommand(git, config))
	root.AddCommand(NewEditCommand(git, config))
	root.AddCommand(NewDeleteCommand(git, config))

	return root
}
