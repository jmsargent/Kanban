package githubapi_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jmsargent/kanban/internal/adapters/githubapi"
	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// Test Budget: 2 behaviors × 2 = 4 max unit tests. Using 2.
// Behavior 1: stub returns task files → adapter maps them to domain.Task slice.
// Behavior 2: stub returns 404 → adapter returns ErrRepositoryNotFound.

func TestGitHubAPIAdapter_ShouldReturnTasksFromStub(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/contents/.kanban/tasks":
			entries := []map[string]string{
				{"name": "TASK-001.md", "type": "file", "download_url": "http://" + r.Host + "/files/TASK-001.md"},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(entries)
		case "/files/TASK-001.md":
			_, _ = fmt.Fprint(w, "---\nid: TASK-001\ntitle: Fix scheduler bug\nstatus: in-progress\n---\n")
		default:
			http.NotFound(w, r)
		}
	}))
	defer stub.Close()

	adapter := githubapi.NewAdapter(stub.URL)
	tasks, err := adapter.ListTasks("owner", "repo")

	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "TASK-001", tasks[0].ID)
	assert.Equal(t, "Fix scheduler bug", tasks[0].Title)
	assert.Equal(t, domain.StatusInProgress, tasks[0].Status)
}

func TestGitHubAPIAdapter_ShouldReturnErrRepositoryNotFoundOn404(t *testing.T) {
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
	}))
	defer stub.Close()

	adapter := githubapi.NewAdapter(stub.URL)
	_, err := adapter.ListTasks("no-such", "repo")

	assert.ErrorIs(t, err, ports.ErrRepositoryNotFound)
}
