package main

import (
	"fmt"
	"os"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	gitadapter "github.com/kanban-tasks/kanban/internal/adapters/git"
	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
)

func main() {
	git := gitadapter.NewGitAdapter()
	config := filesystem.NewConfigRepository()

	root := cli.NewRootCommand(git, config)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
