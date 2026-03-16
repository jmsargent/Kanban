package main

import (
	"fmt"
	"os"

	"github.com/kanban-tasks/kanban/internal/adapters/cli"
	"github.com/kanban-tasks/kanban/internal/adapters/filesystem"
	gitadapter "github.com/kanban-tasks/kanban/internal/adapters/git"
)

func main() {
	git := gitadapter.NewGitAdapter()
	config := filesystem.NewConfigRepository()
	tasks := filesystem.NewTaskRepository()

	root := cli.NewRootCommand(git, config, tasks, tasks)
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
