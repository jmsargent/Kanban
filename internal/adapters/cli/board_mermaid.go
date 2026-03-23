package cli

import (
	"strings"

	"github.com/kanban-tasks/kanban/internal/domain"
)

// mermaidTitleReplacer replaces characters that are unsafe in Mermaid node labels.
var mermaidTitleReplacer = strings.NewReplacer(
	`"`, `'`,
	`[`, `(`,
	`]`, `)`,
	"\n", " ",
	"\r", " ",
)

// sanitiseMermaidTitle replaces Mermaid-unsafe characters in a task title with
// safe substitutes. It is a pure function: no I/O, no side effects.
func sanitiseMermaidTitle(s string) string {
	return mermaidTitleReplacer.Replace(s)
}

// mermaidLabelReplacer replaces characters that are unsafe in Mermaid section labels.
var mermaidLabelReplacer = strings.NewReplacer(
	`:`, ` `,
	`"`, `'`,
	`[`, `(`,
	`]`, `)`,
	"\n", " ",
	"\r", " ",
)

// sanitiseMermaidLabel replaces Mermaid-unsafe characters in a column label with
// safe substitutes. It is a pure function: no I/O, no side effects.
func sanitiseMermaidLabel(s string) string {
	return mermaidLabelReplacer.Replace(s)
}

// renderBoardMermaid returns a fenced Mermaid kanban block representing the board.
// It is a pure function: no I/O, no side effects.
func renderBoardMermaid(board domain.Board) string {
	var sb strings.Builder
	sb.WriteString("```mermaid\n")
	sb.WriteString("kanban\n")
	for _, col := range board.Columns {
		sb.WriteString("  section ")
		sb.WriteString(sanitiseMermaidLabel(col.Label))
		sb.WriteString("\n")
		status := domain.TaskStatus(col.Name)
		for _, task := range board.Tasks[status] {
			sb.WriteString("    ")
			sb.WriteString(task.ID)
			sb.WriteString("@{ label: \"")
			sb.WriteString(sanitiseMermaidTitle(task.Title))
			sb.WriteString("\" }\n")
		}
	}
	sb.WriteString("```\n")
	return sb.String()
}
