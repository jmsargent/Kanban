package web

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
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
	data := struct {
		Title string
		Error string
	}{Title: "Sign In"}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: render token entry template: %v", err)
	}
}

// TokenSubmitHandler handles POST /auth/token — validates the submitted GitHub
// token, sets an encrypted session cookie, and redirects to /board on success.
type TokenSubmitHandler struct {
	sessionKey       []byte
	githubAPIBaseURL string
	tmpl             *template.Template
}

// NewTokenSubmitHandler constructs a TokenSubmitHandler.
// sessionKey must be 32 bytes. githubAPIBaseURL is the GitHub API base (e.g.
// "https://api.github.com" in production, or a stub URL in tests).
func NewTokenSubmitHandler(sessionKey []byte, githubAPIBaseURL string) *TokenSubmitHandler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/token_entry.html"))
	return &TokenSubmitHandler{
		sessionKey:       sessionKey,
		githubAPIBaseURL: githubAPIBaseURL,
		tmpl:             tmpl,
	}
}

// ServeHTTP handles POST /auth/token. It validates the token against the GitHub
// API, sets an AES-256-GCM encrypted HttpOnly session cookie on success, and
// redirects to /board with 303 See Other. On failure it re-renders the form.
func (h *TokenSubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}
	token := r.FormValue("token")
	displayName := r.FormValue("display_name")

	if !h.validateGitHubToken(token) {
		data := struct {
			Title string
			Error string
		}{Title: "Sign In", Error: "Invalid token. Please check your GitHub personal access token and try again."}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
			log.Printf("ERROR: render token entry template: %v", err)
		}
		return
	}

	cookieValue, err := EncryptSession(h.sessionKey, token, displayName)
	if err != nil {
		log.Printf("ERROR: encrypt session: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "kanban_session",
		Value:    cookieValue,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/board", http.StatusSeeOther)
}

// validateGitHubToken calls GET /user on the GitHub API with the provided token.
// Returns true if the API responds with 200 OK.
func (h *TokenSubmitHandler) validateGitHubToken(token string) bool {
	req, err := http.NewRequest(http.MethodGet, h.githubAPIBaseURL+"/user", nil)
	if err != nil {
		log.Printf("ERROR: create GitHub validation request: %v", err)
		return false
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("ERROR: call GitHub API: %v", err)
		return false
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}

// AddTaskHandler handles GET /task/new (render form) and POST /task (create task).
// It wraps RequireAuth so both routes require an authenticated session.
type AddTaskHandler struct {
	sessionKey []byte
	addTask    usecases.TaskExecutor
	repoDir    string
	tmpl       *template.Template
}

// NewAddTaskHandler constructs an AddTaskHandler.
// sessionKey is the 32-byte AES-256 key used to decrypt the session cookie.
// addTask is the wired task executor (AddTask or AddTaskAndPush). repoDir is the repository root.
func NewAddTaskHandler(sessionKey []byte, addTask usecases.TaskExecutor, repoDir string) http.Handler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/add_task.html"))
	h := &AddTaskHandler{sessionKey: sessionKey, addTask: addTask, repoDir: repoDir, tmpl: tmpl}
	return RequireAuth(sessionKey, h)
}

// ServeHTTP dispatches GET to the form renderer and POST to the task creator.
func (h *AddTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.renderForm(w, r, "")
	case http.MethodPost:
		h.handlePost(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *AddTaskHandler) renderForm(w http.ResponseWriter, r *http.Request, errMsg string) {
	h.renderFormWithStatus(w, r, errMsg, http.StatusOK)
}

func (h *AddTaskHandler) renderFormWithStatus(w http.ResponseWriter, _ *http.Request, errMsg string, status int) {
	data := struct {
		Title string
		Error string
	}{Title: "Add Task", Error: errMsg}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: render add_task template: %v", err)
	}
}

func (h *AddTaskHandler) handlePost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	// Extract created_by from the session cookie — RequireAuth already validated it.
	cookie, _ := r.Cookie("kanban_session")
	session, err := decryptSession(h.sessionKey, cookie.Value)
	if err != nil {
		log.Printf("ERROR: decrypt session in AddTaskHandler: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	title := r.FormValue("title")
	if strings.TrimSpace(title) == "" {
		h.renderFormWithStatus(w, r, "title is required", http.StatusUnprocessableEntity)
		return
	}

	input := usecases.AddTaskInput{
		Title:       title,
		Description: r.FormValue("description"),
		Priority:    r.FormValue("priority"),
		Assignee:    r.FormValue("assignee"),
		CreatedBy:   session.DisplayName,
	}

	if _, err := h.addTask.Execute(h.repoDir, input); err != nil {
		log.Printf("ERROR: AddTask.Execute: %v", err)
		h.renderForm(w, r, "Failed to create task: "+err.Error())
		return
	}

	http.Redirect(w, r, "/board", http.StatusSeeOther)
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
		Title    string
		Columns  []columnView
		ReadOnly bool
		PushURL  string
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

// RemoteBoardExecutor is the driving-side abstraction used by RemoteBoardHandler.
type RemoteBoardExecutor interface {
	Execute(owner, repo string) (domain.Board, error)
}

// RemoteBoardHandler serves GET /remote/board?owner=X&repo=Y.
// It is read-only: no add/edit controls are rendered.
type RemoteBoardHandler struct {
	getRemoteBoard RemoteBoardExecutor
	tmpl           *template.Template
}

// validOwnerRepo matches alphanumeric characters, hyphens, and dots (max 100 chars).
var validOwnerRepo = regexp.MustCompile(`^[a-zA-Z0-9._-]{1,100}$`)

// NewRemoteBoardHandler constructs a RemoteBoardHandler.
func NewRemoteBoardHandler(getRemoteBoard RemoteBoardExecutor) *RemoteBoardHandler {
	tmpl := template.Must(template.ParseFS(templateFS, "templates/layout.html", "templates/board.html"))
	return &RemoteBoardHandler{getRemoteBoard: getRemoteBoard, tmpl: tmpl}
}

// ServeHTTP handles GET /remote/board?owner=X&repo=Y.
func (h *RemoteBoardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	owner := r.URL.Query().Get("owner")
	repo := r.URL.Query().Get("repo")

	if !validOwnerRepo.MatchString(owner) || !validOwnerRepo.MatchString(repo) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "<p>Invalid owner or repository name.</p>")
		return
	}

	board, err := h.getRemoteBoard.Execute(owner, repo)
	if err != nil {
		h.renderError(w, err)
		return
	}

	type columnView struct {
		Label  string
		Column string
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

	data := struct {
		Title    string
		Columns  []columnView
		ReadOnly bool
		PushURL  string
		Owner    string
		Repo     string
	}{
		Title:    fmt.Sprintf("%s/%s", owner, repo),
		Columns:  cols,
		ReadOnly: true,
		PushURL:  fmt.Sprintf("?owner=%s&repo=%s", owner, repo),
		Owner:    owner,
		Repo:     repo,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := h.tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		log.Printf("ERROR: render remote board template: %v", err)
	}
}

func (h *RemoteBoardHandler) renderError(w http.ResponseWriter, err error) {
	var msg string
	switch {
	case errors.Is(err, ports.ErrRepositoryNotFound):
		msg = "Repository not found"
	case errors.Is(err, ports.ErrNoBoardFound):
		msg = "This repository has no kanban board"
	case errors.Is(err, ports.ErrRateLimitExceeded):
		msg = "GitHub API rate limit exceeded. Try again later."
	default:
		msg = "Failed to load board: " + err.Error()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "<p>%s</p>", msg)
}
