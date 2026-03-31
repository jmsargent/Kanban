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
	cookieKeyStr := os.Getenv("KANBAN_SESSION_KEY")
	githubAPIURL := os.Getenv("KANBAN_WEB_GITHUB_API_URL")

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
			if i+1 < len(os.Args) {
				cookieKeyStr = os.Args[i+1]
				i++
			}
		}
	}

	// Session key must be exactly 32 bytes for AES-256-GCM.
	// Use KANBAN_SESSION_KEY or --cookie-key with a securely generated 32-byte value.
	sessionKey := []byte(cookieKeyStr)
	if len(sessionKey) != 32 {
		log.Fatalf("session key must be exactly 32 bytes (got %d); set KANBAN_SESSION_KEY to a 32-byte value", len(sessionKey))
	}

	taskRepo := filesystem.NewTaskRepository()
	configRepo := filesystem.NewConfigRepository()
	getBoardUC := usecases.NewGetBoard(configRepo, taskRepo)
	remoteGit := gitadapter.NewGitAdapter()

	// When a repo directory is provided, use AddTaskAndPush so that new tasks
	// are committed and pushed to the remote. Without a repo (dev/preview mode)
	// fall back to AddTask which only persists locally.
	var addTaskUC usecases.TaskExecutor
	if repoDir != "" {
		addTaskUC = usecases.NewAddTaskAndPush(configRepo, taskRepo, remoteGit)
	} else {
		addTaskUC = usecases.NewAddTask(configRepo, taskRepo)
	}

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

	server := web.NewServer(addr, boardProvider, taskProvider, sessionKey, githubAPIURL, addTaskUC, repoDir)
	fmt.Fprintf(os.Stderr, "kanban-web listening on %s\n", addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
