package web

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jmsargent/kanban/internal/domain"
)

// BoardProvider is a function that returns the current board state.
// It is the driving-side abstraction used by BoardHandler.
type BoardProvider func() (domain.Board, error)

// StaticBoardProvider returns a BoardProvider that always yields the given board.
// Useful for unit tests.
func StaticBoardProvider(b domain.Board) BoardProvider {
	return func() (domain.Board, error) { return b, nil }
}

// BoardHandler serves the kanban board page at GET /board.
type BoardHandler struct {
	getBoard BoardProvider
	tmpl     *template.Template
}

// NewBoardHandler constructs a BoardHandler with the given board provider.
func NewBoardHandler(getBoard BoardProvider) *BoardHandler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/board.html"))
	return &BoardHandler{getBoard: getBoard, tmpl: tmpl}
}

// ServeHTTP handles GET /board requests, rendering the board template.
func (h *BoardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	board, err := h.getBoard()
	if err != nil {
		log.Printf("ERROR: get board: %v", err)
		http.Error(w, "failed to load board", http.StatusInternalServerError)
		return
	}

	// Build a view-friendly structure: slice of column+tasks pairs in order.
	type columnView struct {
		Label  string
		Column string // display name used in data-column attribute
		Tasks  []domain.Task
	}

	cols := make([]columnView, 0, len(board.Columns))
	for _, col := range board.Columns {
		tasks := board.Tasks[domain.TaskStatus(col.Name)]
		cols = append(cols, columnView{
			Label:  col.Label,
			Column: col.Label,
			Tasks:  tasks,
		})
	}

	// When no columns configured (e.g. empty board in tests), emit defaults.
	if len(cols) == 0 {
		cols = []columnView{
			{Label: "Todo", Column: "Todo", Tasks: nil},
			{Label: "Doing", Column: "Doing", Tasks: nil},
			{Label: "Done", Column: "Done", Tasks: nil},
		}
	}

	data := struct {
		Title   string
		Columns []columnView
	}{
		Title:   "Kanban Board",
		Columns: cols,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: render board template: %v", err)
	}
}
