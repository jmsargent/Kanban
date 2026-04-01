package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmsargent/kanban/internal/domain"
)

func TestNewServer_NilSessionKey_UsesZeroKey(t *testing.T) {
	s := NewServer(":0", StaticBoardProvider(domain.Board{}), nil, nil, "", nil, "")
	if s.sessionKey == nil {
		t.Fatal("expected sessionKey to be non-nil after nil input")
	}
	if len(s.sessionKey) != 32 {
		t.Errorf("expected 32-byte session key, got %d bytes", len(s.sessionKey))
	}
	for _, b := range s.sessionKey {
		if b != 0 {
			t.Fatal("expected all-zero key when nil is passed")
		}
	}
}

func TestNewServer_EmptyGitHubAPIURL_UsesDefault(t *testing.T) {
	s := NewServer(":0", StaticBoardProvider(domain.Board{}), nil, make([]byte, 32), "", nil, "")
	if s.githubAPIBaseURL != "https://api.github.com" {
		t.Errorf("expected default GitHub API URL, got %q", s.githubAPIBaseURL)
	}
}

func TestNewServer_NilAddTask_AuthenticatedRequestReturns501(t *testing.T) {
	sessionKey := []byte("test-cookie-key-must-be-32bytes!")
	s := NewServer(":0", StaticBoardProvider(domain.Board{}), nil, sessionKey, "", nil, "")

	cookieValue, err := EncryptSession(sessionKey, "ghp_token", "Alice")
	if err != nil {
		t.Fatalf("EncryptSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/task/new", nil)
	req.AddCookie(&http.Cookie{Name: "kanban_session", Value: cookieValue})
	rec := httptest.NewRecorder()

	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented when addTask is nil, got %d", rec.Code)
	}
}
