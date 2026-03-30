package dsl

import (
	"fmt"

	"github.com/jmsargent/kanban/pkg/simpledsl"
	"github.com/jmsargent/kanban/tests/acceptance/backend/driver"
)

var aRepoWithTasksDSL = simpledsl.NewDslParams(
	simpledsl.NewRepeatingGroup("task",
		simpledsl.NewRequiredArg("title"),
		simpledsl.NewOptionalArg("status").SetDefault("todo"),
		simpledsl.NewOptionalArg("assignee"),
		simpledsl.NewOptionalArg("created_at"),
	),
)

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
