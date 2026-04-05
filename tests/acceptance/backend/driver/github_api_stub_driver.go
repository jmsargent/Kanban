package driver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
)

// GitHubAPIStubDriver runs a local httptest server that stubs the GitHub REST
// API endpoints required by the github-api-board-viewer feature:
//   - GET /repos/{owner}/{repo}/contents/.kanban/tasks  (directory listing)
//   - GET /files/{owner}/{repo}/{filename}              (raw file download)
//
// It extends the existing GitHubStubDriver's GET /user stub with these new
// endpoints. Tests configure stub responses per owner/repo before starting the
// server under test.
type GitHubAPIStubDriver struct {
	server   *httptest.Server
	mu       sync.RWMutex
	repos    map[string]repoStub // keyed by "owner/repo"
	baseURL  string
}

// repoStub holds the configured response for a single owner/repo combination.
type repoStub struct {
	tasks       []stubTask // nil means "no .kanban/tasks/ directory" (ErrNoBoardFound)
	notFound    bool       // true -> 404 (ErrRepositoryNotFound)
	rateLimited bool       // true -> 429 (ErrRateLimitExceeded)
}

// stubTask represents a single task file in the stub directory listing.
type stubTask struct {
	filename string
	content  string // raw YAML front matter + body
}

// NewGitHubAPIStubDriver creates and starts a stub server for the GitHub
// contents API. Register repo behaviours via the Add* methods before starting
// the kanban-web server under test.
func NewGitHubAPIStubDriver() *GitHubAPIStubDriver {
	d := &GitHubAPIStubDriver{
		repos: make(map[string]repoStub),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/{owner}/{repo}/contents/.kanban/tasks", d.handleContentsListing)
	mux.HandleFunc("GET /repos/{owner}/{repo}", d.handleRepoCheck)
	mux.HandleFunc("GET /files/{owner}/{repo}/{filename}", d.handleFileDownload)

	d.server = httptest.NewServer(mux)
	d.baseURL = d.server.URL

	return d
}

// URL returns the base URL of the stub server. Inject this via the
// KANBAN_WEB_GITHUB_API_URL environment variable when starting the server
// under test so that the GitHub API adapter calls this stub instead of
// api.github.com.
func (d *GitHubAPIStubDriver) URL() string {
	return d.baseURL
}

// Close shuts down the stub server.
func (d *GitHubAPIStubDriver) Close() {
	d.server.Close()
}

// AddRepoWithTasks registers a valid public repository that contains the given
// task files in its .kanban/tasks/ directory. Each task is provided as raw
// Markdown with YAML front matter (the same format kanban uses on disk).
func (d *GitHubAPIStubDriver) AddRepoWithTasks(owner, repo string, tasks []StubTaskFile) {
	d.mu.Lock()
	defer d.mu.Unlock()

	stub := repoStub{}
	for _, t := range tasks {
		stub.tasks = append(stub.tasks, stubTask{
			filename: t.Filename,
			content:  t.Content,
		})
	}
	d.repos[owner+"/"+repo] = stub
}

// AddRepoWithNoTasks registers a valid public repository that exists but has
// an empty .kanban/tasks/ directory (zero task files, no error).
func (d *GitHubAPIStubDriver) AddRepoWithNoTasks(owner, repo string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.repos[owner+"/"+repo] = repoStub{tasks: []stubTask{}} // non-nil empty slice = directory exists, no files
}

// AddRepoWithNoBoard registers a repository that exists but has no
// .kanban/tasks/ directory at all. The GitHub API returns 404 for the
// directory path, which maps to ErrNoBoardFound.
func (d *GitHubAPIStubDriver) AddRepoWithNoBoard(owner, repo string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.repos[owner+"/"+repo] = repoStub{tasks: nil} // nil tasks = directory absent
}

// AddNotFoundRepo registers an owner/repo combination that the stub will
// respond to with HTTP 404, simulating an invalid owner, invalid repo name,
// or private repository (all indistinguishable by design).
func (d *GitHubAPIStubDriver) AddNotFoundRepo(owner, repo string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.repos[owner+"/"+repo] = repoStub{notFound: true}
}

// AddRateLimitedRepo registers an owner/repo combination that the stub will
// respond to with HTTP 429, simulating GitHub API rate limit exhaustion.
func (d *GitHubAPIStubDriver) AddRateLimitedRepo(owner, repo string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.repos[owner+"/"+repo] = repoStub{rateLimited: true}
}

// StubTaskFile describes a single task file to include in a stub repo listing.
type StubTaskFile struct {
	Filename string // e.g. "TASK-001.md"
	Content  string // raw Markdown with YAML front matter
}

// handleRepoCheck serves GET /repos/{owner}/{repo}.
// Returns 200 if the repo is known and not flagged as notFound, 404 otherwise.
// The adapter calls this endpoint when the contents listing returns 404 to
// distinguish "repo not found" from "repo exists but no .kanban/tasks/ folder".
func (d *GitHubAPIStubDriver) handleRepoCheck(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")

	d.mu.RLock()
	stub, known := d.repos[owner+"/"+repo]
	d.mu.RUnlock()

	if !known || stub.notFound {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, `{"full_name":"%s/%s"}`, owner, repo)
}

// handleContentsListing serves GET /repos/{owner}/{repo}/contents/.kanban/tasks.
// It returns a JSON array of file objects matching the GitHub contents API
// shape that the GitHub API adapter consumes (name + download_url per file).
func (d *GitHubAPIStubDriver) handleContentsListing(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")
	key := owner + "/" + repo

	d.mu.RLock()
	stub, known := d.repos[key]
	d.mu.RUnlock()

	if !known || stub.notFound {
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}
	if stub.rateLimited {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = fmt.Fprint(w, `{"message":"API rate limit exceeded"}`)
		return
	}
	if stub.tasks == nil {
		// Directory does not exist — GitHub returns 404 for the path.
		http.Error(w, `{"message":"Not Found"}`, http.StatusNotFound)
		return
	}

	// Build directory listing — same shape as GitHub contents API with
	// application/vnd.github.object Accept header.
	type fileEntry struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}
	entries := make([]fileEntry, 0, len(stub.tasks))
	for _, t := range stub.tasks {
		entries = append(entries, fileEntry{
			Name:        t.filename,
			Type:        "file",
			DownloadURL: fmt.Sprintf("%s/files/%s/%s/%s", d.baseURL, owner, repo, t.filename),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(entries)
}

// handleFileDownload serves individual task file downloads at the download_url
// path that handleContentsListing embeds in each listing entry.
func (d *GitHubAPIStubDriver) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	owner := r.PathValue("owner")
	repo := r.PathValue("repo")
	filename := r.PathValue("filename")
	key := owner + "/" + repo

	d.mu.RLock()
	stub, known := d.repos[key]
	d.mu.RUnlock()

	if !known {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	for _, t := range stub.tasks {
		if t.filename == filename {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = fmt.Fprint(w, t.content)
			return
		}
	}
	http.Error(w, "not found", http.StatusNotFound)
}
