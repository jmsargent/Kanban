package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// NewInitCommand builds the "kanban init" cobra command.
// It wires the InitRepo use case and maps errors to appropriate exit codes.
func NewInitCommand(git ports.GitPort, config ports.ConfigRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialise kanban in the current git repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			uc := usecases.NewInitRepo(git, config, os.Stdout)
			if err := uc.Execute(); err != nil {
				if errors.Is(err, ports.ErrNotGitRepo) {
					fmt.Fprintln(os.Stderr, "Not a git repository")
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			return nil
		},
	}
}
