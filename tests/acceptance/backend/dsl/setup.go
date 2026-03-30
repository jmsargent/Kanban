package dsl

import (
	"fmt"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

var aUserPushesTaskDSL = simpledsl.NewDslParams(
	simpledsl.NewRequiredArg("task"),
	simpledsl.NewOptionalArg("status").SetDefault("todo"),
	simpledsl.NewOptionalArg("assignee"),
)

var aRepoWithTasksDSL = simpledsl.NewDslParams(
	simpledsl.NewRepeatingGroup("task",
		simpledsl.NewRequiredArg("title"),
		simpledsl.NewOptionalArg("status").SetDefault("todo"),
		simpledsl.NewOptionalArg("assignee"),
		simpledsl.NewOptionalArg("created_at"),
	),
)

// ARepoWithNoTasks sets up an empty temporary repo (no task files) and stores
// the repo dir on the context so IVisitTheBoard can start the server with --repo.
func ARepoWithNoTasks(params ...string) Step {
	return Step{
		Description: "a repo with no tasks",
		Run: func(ctx *WebContext) error {
			rd := driver.NewRepoDriver(ctx.T)
			ctx.RepoDir = rd.RepoDir()
			return nil
		},
	}
}

// ARepoWithRemote sets up a temporary local repo backed by a bare remote
// (simulating a shared git server). ctx.RepoDir is set to the local clone and
// ctx.RemoteDir is set to the bare repo path so AUserPushesTask can push to it.
func ARepoWithRemote(params ...string) Step {
	return Step{
		Description: "a repo with a remote",
		Run: func(ctx *WebContext) error {
			rd := driver.NewRepoDriver(ctx.T)
			rd.SetupBareRemote()
			ctx.RepoDir = rd.RepoDir()
			ctx.RemoteDir = rd.BareDir()
			return nil
		},
	}
}

// AUserPushesTask simulates another user pushing a new task to the shared
// remote. It clones the bare remote into a fresh temp dir, writes a task file,
// commits, and pushes back. Required param: "task: <title>".
// Optional params: "status: <status>", "assignee: <assignee>".
func AUserPushesTask(params ...string) Step {
	return Step{
		Description: fmt.Sprintf("a user pushes a task (%v)", params),
		Run: func(ctx *WebContext) error {
			if ctx.RemoteDir == "" {
				return fmt.Errorf("AUserPushesTask: RemoteDir not set; call ARepoWithRemote first")
			}
			vals, err := aUserPushesTaskDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("AUserPushesTask: %w", err)
			}
			title := vals.Value("task")
			status := vals.Value("status")
			assignee := vals.Value("assignee")

			// Use a fresh RepoDriver as the "other user's" working copy.
			other := driver.NewRepoDriverFromRemote(ctx.T, ctx.RemoteDir)
			id := "REMOTE-001"
			other.SeedTask(id, title, status, assignee)
			other.PushToOrigin()
			return nil
		},
	}
}

// ARepoWithTasks seeds a temporary repo with the given tasks and stores the
// repo dir on the context so IVisitTheBoard can start the server with --repo.
// Each task is introduced by "task: <title>" followed by optional named params.
// Optional "created_at: <RFC3339>" param controls creation date ordering.
func ARepoWithTasks(params ...string) Step {
	return Step{
		Description: "a repo with tasks",
		Run: func(ctx *WebContext) error {
			vals, err := aRepoWithTasksDSL.Parse(params)
			if err != nil {
				return fmt.Errorf("ARepoWithTasks: %w", err)
			}
			rd := driver.NewRepoDriver(ctx.T)
			for i, taskVals := range vals.Group("task") {
				id := fmt.Sprintf("TASK-%03d", i+1)
				rd.SeedTaskWithDate(id, taskVals.Value("title"), taskVals.Value("status"), taskVals.Value("assignee"), taskVals.Value("created_at"))
			}
			ctx.RepoDir = rd.RepoDir()
			return nil
		},
	}
}
