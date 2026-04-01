package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jmsargent/kanban/internal/adapters/web"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
	"github.com/jmsargent/kanban/internal/usecases"
)

// Test Budget: 3 behaviors × 2 = 6 max unit tests. Using 3.
// Behavior 1: GET /board returns 200 with non-empty HTML body.
// Behavior 2: BoardHandler renders tasks in correct columns (Todo/Doing/Done).
// Behavior 3: BoardHandler renders three empty columns when no tasks exist.

func TestBoardHandler_Returns200WithHTMLBody(t *testing.T) {
	handler := web.NewBoardHandler(web.StaticBoardProvider(domain.Board{}))

	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.TrimSpace(body) == "" {
		t.Fatal("expected non-empty response body")
	}
}

func TestBoardHandler_RendersTasksInCorrectColumns(t *testing.T) {
	board := domain.Board{
		Columns: []domain.Column{
			{Name: "todo", Label: "Todo"},
			{Name: "in-progress", Label: "Doing"},
			{Name: "done", Label: "Done"},
		},
		Tasks: map[domain.TaskStatus][]domain.Task{
			domain.StatusTodo:       {{ID: "TASK-002", Title: "Write docs", Status: domain.StatusTodo}},
			domain.StatusInProgress: {{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusInProgress}},
			domain.StatusDone:       {{ID: "TASK-003", Title: "Deploy v1", Status: domain.StatusDone}},
		},
	}

	handler := web.NewBoardHandler(web.StaticBoardProvider(board))
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	assertColumnContains(t, body, "Todo", "Write docs")
	assertColumnContains(t, body, "Doing", "Fix login bug")
	assertColumnContains(t, body, "Done", "Deploy v1")
}

func TestBoardHandler_RendersEmptyColumns(t *testing.T) {
	board := domain.Board{
		Columns: []domain.Column{
			{Name: "todo", Label: "Todo"},
			{Name: "in-progress", Label: "Doing"},
			{Name: "done", Label: "Done"},
		},
		Tasks: map[domain.TaskStatus][]domain.Task{},
	}

	handler := web.NewBoardHandler(web.StaticBoardProvider(board))
	req := httptest.NewRequest(http.MethodGet, "/board", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()
	for _, col := range []string{"Todo", "Doing", "Done"} {
		marker := fmt.Sprintf(`data-column="%s"`, col)
		if !strings.Contains(body, marker) {
			t.Errorf("expected column %q in HTML", col)
		}
	}
}

// Test budget for CardDetailHandler: 2 behaviors × 2 = 4 max. Using 2.
// Behavior 1: card found → 200 with title, assignee, status fields.
// Behavior 2: card not found → 404.

func TestCardDetailHandler_RendersCardFields(t *testing.T) {
	task := domain.Task{ID: "TASK-001", Title: "Fix login bug", Status: domain.StatusInProgress, Assignee: "alice"}
	provider := func(id string) (domain.Task, error) {
		if id == "TASK-001" {
			return task, nil
		}
		return domain.Task{}, fmt.Errorf("not found")
	}

	handler := web.NewCardDetailHandler(provider)
	req := httptest.NewRequest(http.MethodGet, "/card/TASK-001", nil)
	req.SetPathValue("id", "TASK-001")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`data-field="title"`, `data-field="status"`, `data-field="assignee"`} {
		if !strings.Contains(body, want) {
			t.Errorf("expected %q in response body", want)
		}
	}
}

func TestCardDetailHandler_NotFound(t *testing.T) {
	provider := func(id string) (domain.Task, error) {
		return domain.Task{}, fmt.Errorf("not found")
	}

	handler := web.NewCardDetailHandler(provider)
	req := httptest.NewRequest(http.MethodGet, "/card/TASK-999", nil)
	req.SetPathValue("id", "TASK-999")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// Test Budget: 2 behaviors × 2 = 4 max unit tests. Using 2.
// Behavior 1: valid token → 303 redirect to /board + kanban_session cookie set.
// Behavior 2: invalid token → re-renders token-entry form (no redirect, no cookie).

func TestTokenSubmitHandler_ValidToken_RedirectsToBoardWithCookie(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer valid-token" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprintln(w, `{"login":"alice","name":"Alice"}`)
			return
		}
		http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
	}))
	defer stub.Close()

	sessionKey := make([]byte, 32)
	handler := web.NewTokenSubmitHandler(sessionKey, stub.URL)

	form := strings.NewReader("token=valid-token&display_name=Alice")
	req := httptest.NewRequest(http.MethodPost, "/auth/token", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 See Other, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/board" {
		t.Fatalf("expected redirect to /board, got %q", loc)
	}
	var sessionCookie *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == "kanban_session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected kanban_session cookie to be set")
	}
	if !sessionCookie.HttpOnly {
		t.Error("expected kanban_session cookie to be HttpOnly")
	}
}

func TestTokenSubmitHandler_InvalidToken_RerendersForm(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Bad credentials"}`, http.StatusUnauthorized)
	}))
	defer stub.Close()

	sessionKey := make([]byte, 32)
	handler := web.NewTokenSubmitHandler(sessionKey, stub.URL)

	form := strings.NewReader("token=bad-token&display_name=Bob")
	req := httptest.NewRequest(http.MethodPost, "/auth/token", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusSeeOther {
		t.Fatal("expected no redirect on invalid token")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "token-entry") {
		t.Errorf("expected token-entry form in response, got:\n%s", body)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "kanban_session" {
			t.Error("expected no session cookie on invalid token")
		}
	}
}

// Test Budget: 2 behaviors × 2 = 4 max unit tests. Using 2.
// Behavior 1: authenticated POST with valid fields → 303 redirect to /board + task saved.
// Behavior 2: unauthenticated POST → 302 redirect to /auth/token (RequireAuth middleware).

// fakeTaskRepo is a minimal in-memory TaskRepository for unit tests.
type fakeTaskRepo struct {
	saved []domain.Task
}

func (r *fakeTaskRepo) Save(_ string, task domain.Task) error {
	r.saved = append(r.saved, task)
	return nil
}
func (r *fakeTaskRepo) FindByID(_ string, _ string) (domain.Task, error) { return domain.Task{}, nil }
func (r *fakeTaskRepo) ListAll(_ string) ([]domain.Task, error)          { return nil, nil }
func (r *fakeTaskRepo) Update(_ string, _ domain.Task) error             { return nil }
func (r *fakeTaskRepo) Delete(_ string, _ string) error                  { return nil }
func (r *fakeTaskRepo) NextID(_ string) (string, error)                  { return "TASK-001", nil }

// fakeConfigRepo is a minimal in-memory ConfigRepository for unit tests.
type fakeConfigRepo struct{}

func (r *fakeConfigRepo) Read(_ string) (ports.Config, error) {
	return ports.Config{
		Columns: []domain.Column{
			{Name: "todo", Label: "Todo"},
		},
	}, nil
}
func (r *fakeConfigRepo) Write(_ string, _ ports.Config) error { return nil }

func TestAddTaskHandler_AuthenticatedPost_RedirectsAndSavesTask(t *testing.T) {
	sessionKey := []byte("test-cookie-key-must-be-32bytes!")
	taskRepo := &fakeTaskRepo{}
	configRepo := &fakeConfigRepo{}
	addTaskUC := usecases.NewAddTask(configRepo, taskRepo)

	handler := web.NewAddTaskHandler(sessionKey, addTaskUC, "")

	cookieValue, err := web.EncryptSession(sessionKey, "ghp_token", "Alice")
	if err != nil {
		t.Fatalf("EncryptSession: %v", err)
	}

	form := strings.NewReader("title=Fix+the+bug&description=Needs+attention&priority=high&assignee=bob")
	req := httptest.NewRequest(http.MethodPost, "/task", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "kanban_session", Value: cookieValue})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("expected 303 See Other, got %d\nbody: %s", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); loc != "/board" {
		t.Fatalf("expected redirect to /board, got %q", loc)
	}
	if len(taskRepo.saved) != 1 {
		t.Fatalf("expected 1 saved task, got %d", len(taskRepo.saved))
	}
	saved := taskRepo.saved[0]
	if saved.Title != "Fix the bug" {
		t.Errorf("expected title %q, got %q", "Fix the bug", saved.Title)
	}
	if saved.CreatedBy != "Alice" {
		t.Errorf("expected created_by %q (from session), got %q", "Alice", saved.CreatedBy)
	}
}

func TestAddTaskHandler_Unauthenticated_RedirectsToAuth(t *testing.T) {
	sessionKey := []byte("test-cookie-key-must-be-32bytes!")
	taskRepo := &fakeTaskRepo{}
	configRepo := &fakeConfigRepo{}
	addTaskUC := usecases.NewAddTask(configRepo, taskRepo)

	handler := web.NewAddTaskHandler(sessionKey, addTaskUC, "")

	form := strings.NewReader("title=Fix+the+bug")
	req := httptest.NewRequest(http.MethodPost, "/task", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected 302 Found, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/auth/token" {
		t.Fatalf("expected redirect to /auth/token, got %q", loc)
	}
}

// Test Budget (AddTaskHandler validation): 1 behavior × 2 = 2 max. Using 1.
// Behavior 3: POST with empty title → re-renders add-task form with "title is required" error, no task saved.

func TestAddTaskHandler_EmptyTitle_RerendersFormWithError(t *testing.T) {
	sessionKey := []byte("test-cookie-key-must-be-32bytes!")
	taskRepo := &fakeTaskRepo{}
	configRepo := &fakeConfigRepo{}
	addTaskUC := usecases.NewAddTask(configRepo, taskRepo)

	handler := web.NewAddTaskHandler(sessionKey, addTaskUC, "")

	cookieValue, err := web.EncryptSession(sessionKey, "ghp_token", "Alice")
	if err != nil {
		t.Fatalf("EncryptSession: %v", err)
	}

	form := strings.NewReader("title=&description=Some+description")
	req := httptest.NewRequest(http.MethodPost, "/task", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "kanban_session", Value: cookieValue})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusSeeOther {
		t.Fatal("expected no redirect when title is empty")
	}
	body := rec.Body.String()
	// The handler must produce exactly "title is required" — not a use-case error message.
	if !strings.Contains(body, "title is required") {
		t.Errorf("expected 'title is required' error in response body, got:\n%s", body)
	}
	if strings.Contains(body, "Failed to create task") {
		t.Errorf("handler must validate before calling use case; got use-case error in body:\n%s", body)
	}
	if !strings.Contains(body, "add-task-form") {
		t.Errorf("expected add-task form to be re-rendered, got:\n%s", body)
	}
	if len(taskRepo.saved) != 0 {
		t.Errorf("expected no task saved when title is empty, got %d", len(taskRepo.saved))
	}
}

// fakeErrorTaskRepo is an in-memory TaskRepository whose Save always returns an error.
type fakeErrorTaskRepo struct{}

func (r *fakeErrorTaskRepo) Save(_ string, _ domain.Task) error {
	return fmt.Errorf("disk full")
}
func (r *fakeErrorTaskRepo) FindByID(_ string, _ string) (domain.Task, error) {
	return domain.Task{}, nil
}
func (r *fakeErrorTaskRepo) ListAll(_ string) ([]domain.Task, error) { return nil, nil }
func (r *fakeErrorTaskRepo) Update(_ string, _ domain.Task) error    { return nil }
func (r *fakeErrorTaskRepo) Delete(_ string, _ string) error         { return nil }
func (r *fakeErrorTaskRepo) NextID(_ string) (string, error)         { return "TASK-001", nil }

func TestAddTaskHandler_ExecuteFailure_RerendersFormWithError(t *testing.T) {
	sessionKey := []byte("test-cookie-key-must-be-32bytes!")
	configRepo := &fakeConfigRepo{}
	addTaskUC := usecases.NewAddTask(configRepo, &fakeErrorTaskRepo{})

	handler := web.NewAddTaskHandler(sessionKey, addTaskUC, "")

	cookieValue, err := web.EncryptSession(sessionKey, "ghp_token", "Alice")
	if err != nil {
		t.Fatalf("EncryptSession: %v", err)
	}

	form := strings.NewReader("title=Fix+the+bug")
	req := httptest.NewRequest(http.MethodPost, "/task", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "kanban_session", Value: cookieValue})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code == http.StatusSeeOther {
		t.Fatal("expected no redirect when task executor fails")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Failed to create task") {
		t.Errorf("expected error message in response body, got:\n%s", body)
	}
}

// assertColumnContains checks that the named column in body contains a card
// with the given title.
func assertColumnContains(t *testing.T, body, column, title string) {
	t.Helper()
	colMarker := fmt.Sprintf(`data-column="%s"`, column)
	colIdx := strings.Index(body, colMarker)
	if colIdx < 0 {
		t.Errorf("column %q not found in HTML", column)
		return
	}
	section := extractSection(body, colIdx)
	cardMarker := fmt.Sprintf(`data-title="%s"`, title)
	if !strings.Contains(section, cardMarker) {
		t.Errorf("card %q not found in column %q\nSection:\n%s", title, column, section)
	}
}

func extractSection(html string, startIdx int) string {
	openTag := startIdx
	for openTag > 0 && html[openTag] != '<' {
		openTag--
	}
	depth := 0
	i := openTag
	n := len(html)
	for i < n {
		if strings.HasPrefix(html[i:], "<div") {
			depth++
			i += 4
			continue
		}
		if strings.HasPrefix(html[i:], "</div>") {
			depth--
			i += 6
			if depth == 0 {
				return html[openTag:i]
			}
			continue
		}
		i++
	}
	return html[openTag:]
}
