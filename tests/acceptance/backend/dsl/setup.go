package dsl

import (
	"fmt"
	"strings"

	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

// TaskSpec holds the parameters for a single task to be seeded.
type TaskSpec struct {
	Title  string
	Status string
}

// Task builds a TaskSpec from a title and optional "key: value" params.
// Supported params: "status: <value>", "assignee: <value>" (assignee is stored
// in the task file but not used for column placement).
func Task(title string, params ...string) TaskSpec {
	spec := TaskSpec{
		Title:  title,
		Status: "todo",
	}
	for _, p := range params {
		parts := strings.SplitN(p, ": ", 2)
		if len(parts) != 2 {
			continue
		}
		switch strings.TrimSpace(parts[0]) {
		case "status":
			spec.Status = strings.TrimSpace(parts[1])
		}
	}
	return spec
}

// ARepoWithTasks seeds a temporary repo with the given tasks and stores the
// repo dir on the context so IVisitTheBoard can start the server with --repo.
func ARepoWithTasks(tasks ...TaskSpec) Step {
	return Step{
		Description: "a repo with tasks",
		Run: func(ctx *WebContext) error {
			rd := driver.NewRepoDriver(ctx.T)
			for i, task := range tasks {
				id := fmt.Sprintf("TASK-%03d", i+1)
				rd.SeedTask(id, task.Title, task.Status)
			}
			ctx.RepoDir = rd.RepoDir()
			return nil
		},
	}
}
