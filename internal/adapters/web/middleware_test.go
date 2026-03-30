package web_test

// Test Budget: 2 distinct behaviors × 2 = 4 max unit tests. Using 2.
// Behavior 1: request without auth cookie → redirect to /auth/token.
// Behavior 2: request with valid auth cookie → passes through to next handler.

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jmsargent/kanban/internal/adapters/web"
)

func TestRequireAuth_NoCookie_RedirectsToTokenEntry(t *testing.T) {
	key := []byte("test-cookie-key-must-be-32bytes!") // 32 bytes
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := web.RequireAuth(key, next)

	req := httptest.NewRequest(http.MethodGet, "/task/new", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("expected redirect (302), got %d", rec.Code)
	}
	location := rec.Header().Get("Location")
	if location != "/auth/token" {
		t.Fatalf("expected redirect to /auth/token, got %q", location)
	}
}

func TestRequireAuth_ValidCookie_PassesThrough(t *testing.T) {
	key := []byte("test-cookie-key-must-be-32bytes!") // 32 bytes
	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})

	handler := web.RequireAuth(key, next)

	// Build a valid encrypted cookie using the package helper.
	cookieValue, err := web.EncryptSession(key, "ghp_testtoken", "Alice")
	if err != nil {
		t.Fatalf("EncryptSession: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/task/new", nil)
	req.AddCookie(&http.Cookie{Name: "kanban_session", Value: cookieValue})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !reached {
		t.Fatal("expected next handler to be reached but it was not")
	}
}
