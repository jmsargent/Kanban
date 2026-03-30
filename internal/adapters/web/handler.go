package web

import (
	"html/template"
	"log"
	"net/http"

	"github.com/jmsargent/kanban/internal/domain"
)

// SessionKey is the 32-byte AES-256 key used to encrypt/decrypt session cookies.
type SessionKey []byte

// TaskProvider is a function that retrieves a task by ID.
type TaskProvider func(id string) (domain.Task, error)

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

// CardDetailHandler serves the card detail page at GET /card/{id}.
type CardDetailHandler struct {
	getTask TaskProvider
	tmpl    *template.Template
}

// NewCardDetailHandler constructs a CardDetailHandler with the given task provider.
func NewCardDetailHandler(getTask TaskProvider) *CardDetailHandler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/card_detail.html"))
	return &CardDetailHandler{getTask: getTask, tmpl: tmpl}
}

// ServeHTTP handles GET /card/{id} requests, rendering the card detail template.
func (h *CardDetailHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	task, err := h.getTask(id)
	if err != nil {
		http.Error(w, "card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "layout", task); err != nil {
		log.Printf("ERROR: render card detail template: %v", err)
	}
}

// TokenEntryHandler serves the authentication form at GET /auth/token.
type TokenEntryHandler struct {
	tmpl *template.Template
}

// NewTokenEntryHandler constructs a TokenEntryHandler.
func NewTokenEntryHandler() *TokenEntryHandler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/token_entry.html"))
	return &TokenEntryHandler{tmpl: tmpl}
}

// ServeHTTP handles GET /auth/token, rendering the token entry form.
func (h *TokenEntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data := struct{ Title string }{Title: "Sign In"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: render token entry template: %v", err)
	}
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
