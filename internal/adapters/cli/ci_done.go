package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewCIDoneCommand builds the "kanban ci-done" command.
// It advances all in-progress tasks referenced by commits in the pipeline range to done
// and commits the updated files back with [skip ci] to prevent CI recursion.
func NewCIDoneCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	var fromRef string
	var toRef string

	cmd := &cobra.Command{
		Use:   "ci-done",
		Short: "Advance in-progress tasks to done after CI pipeline succeeds",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			from := resolveFrom(fromRef)
			to := resolveTo(toRef)

			uc := usecases.NewTransitionToDone(git, tasks, config, os.Stdout)
			if execErr := uc.Execute(repoRoot, from, to); execErr != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", execErr)
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&fromRef, "from", "", "Start of commit range (defaults to GITHUB_BASE_SHA or CI_COMMIT_BEFORE_SHA env vars, then HEAD^)")
	cmd.Flags().StringVar(&toRef, "to", "HEAD", "End of commit range (default: HEAD)")

	return cmd
}

// resolveFrom returns the --from value or falls back to CI environment variables,
// then to HEAD^ when nothing is configured.
func resolveFrom(flag string) string {
	if flag != "" {
		return flag
	}
	for _, envVar := range []string{"GITHUB_BASE_SHA", "CI_COMMIT_BEFORE_SHA"} {
		if val := os.Getenv(envVar); val != "" {
			return val
		}
	}
	return "HEAD^"
}

// resolveTo returns the --to value, defaulting to HEAD.
func resolveTo(flag string) string {
	if flag != "" {
		return flag
	}
	return "HEAD"
}
