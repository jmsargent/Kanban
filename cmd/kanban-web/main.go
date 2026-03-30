package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jmsargent/kanban/internal/adapters/filesystem"
	gitadapter "github.com/jmsargent/kanban/internal/adapters/git"
	"github.com/jmsargent/kanban/internal/adapters/web"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/usecases"
)

func main() {
	addr := ":8080"
	repoDir := ""

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--port":
			if i+1 < len(os.Args) {
				addr = ":" + os.Args[i+1]
				i++
			}
		case "--addr":
			if i+1 < len(os.Args) {
				addr = os.Args[i+1]
				i++
			}
		case "--repo":
			if i+1 < len(os.Args) {
				repoDir = os.Args[i+1]
				i++
			}
		case "--cookie-key":
			// accepted for forward-compat; unused in this step
			if i+1 < len(os.Args) {
				i++
			}
		}
	}

	taskRepo := filesystem.NewTaskRepository()
	configRepo := filesystem.NewConfigRepository()
	getBoardUC := usecases.NewGetBoard(configRepo, taskRepo)
	remoteGit := gitadapter.NewGitAdapter()

	// Capture repoDir in the closure; it may be empty for backward-compat
	// (board will return empty columns in that case).
	boardProvider := func() (domain.Board, error) {
		if repoDir == "" {
			return domain.Board{
				Columns: []domain.Column{
					{Name: "todo", Label: "Todo"},
					{Name: "in-progress", Label: "Doing"},
					{Name: "done", Label: "Done"},
				},
				Tasks: map[domain.TaskStatus][]domain.Task{},
			}, nil
		}
		// Pull latest changes from remote before rendering so the board
		// reflects commits pushed by other users. Failure is non-fatal:
		// the board serves stale data rather than returning an error.
		if err := remoteGit.Pull(repoDir); err != nil {
			log.Printf("WARN: git pull failed (serving cached state): %v", err)
		}
		return getBoardUC.Execute(repoDir, "")
	}

	taskProvider := func(id string) (domain.Task, error) {
		if repoDir == "" {
			return domain.Task{}, fmt.Errorf("no repo configured")
		}
		return taskRepo.FindByID(repoDir, id)
	}

	server := web.NewServer(addr, boardProvider, taskProvider)
	fmt.Fprintf(os.Stderr, "kanban-web listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
