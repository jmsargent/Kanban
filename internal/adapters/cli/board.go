package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/kanban-tasks/kanban/internal/domain"
	"github.com/kanban-tasks/kanban/internal/ports"
	"github.com/kanban-tasks/kanban/internal/usecases"
)

// NewBoardCommand builds the "kanban board" cobra command.
// It retrieves all tasks, groups them by status, and prints a columnar display.
func NewBoardCommand(git ports.GitPort, config ports.ConfigRepository, tasks ports.TaskRepository) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "board",
		Short: "Display the kanban board",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, err := git.RepoRoot()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Not a git repository")
				os.Exit(1)
			}

			uc := usecases.NewGetBoard(config, tasks)
			board, err := uc.Execute(repoRoot)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if jsonOutput {
				printBoardJSON(board)
				return nil
			}

			printBoard(board)
			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Emit tasks as a JSON array")

	return cmd
}

// printBoard renders the board as a human-readable columnar display.
func printBoard(board domain.Board) {
	totalTasks := 0
	for _, tasks := range board.Tasks {
		totalTasks += len(tasks)
	}

	if totalTasks == 0 {
		fmt.Println("No tasks found in .kanban/tasks/")
		fmt.Println("Hint: run 'kanban new <title>' to create your first task")
		return
	}

	for _, col := range board.Columns {
		status := domain.TaskStatus(col.Name)
		columnTasks := board.Tasks[status]
		count := len(columnTasks)

		fmt.Printf("\n%s (%d)\n", col.Label, count)
		fmt.Println(strings.Repeat("-", len(col.Label)+10))

		for _, t := range columnTasks {
			priority := t.Priority
			if priority == "" {
				priority = "--"
			}
			due := "--"
			if t.Due != nil {
				due = t.Due.Format("2006-01-02")
			}
			assignee := t.Assignee
			if assignee == "" {
				assignee = "unassigned"
			}
			createdBy := t.CreatedBy
			if createdBy == "" {
				createdBy = "--"
			}
			fmt.Printf("  %-12s  %-40s  %-4s  %-10s  %-20s  %s\n",
				t.ID, t.Title, priority, due, assignee, createdBy)
		}
	}
	fmt.Println()
}

// boardTaskJSON is the JSON serialisation shape for a task in the board output.
type boardTaskJSON struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	Status   string  `json:"status"`
	Priority string  `json:"priority"`
	Due      *string `json:"due"`
	Assignee string  `json:"assignee"`
}

// printBoardJSON emits all tasks as a flat JSON array.
func printBoardJSON(board domain.Board) {
	var result []boardTaskJSON
	for _, col := range board.Columns {
		status := domain.TaskStatus(col.Name)
		for _, t := range board.Tasks[status] {
			item := boardTaskJSON{
				ID:       t.ID,
				Title:    t.Title,
				Status:   string(t.Status),
				Priority: t.Priority,
				Assignee: t.Assignee,
			}
			if t.Due != nil {
				due := t.Due.Format("2006-01-02")
				item.Due = &due
			}
			result = append(result, item)
		}
	}
	if result == nil {
		result = []boardTaskJSON{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(result) //nolint:errcheck
}
