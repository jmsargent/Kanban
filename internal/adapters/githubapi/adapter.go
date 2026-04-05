package githubapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/jmsargent/kanban/internal/domain"
	"github.com/jmsargent/kanban/internal/ports"
)

// Adapter implements ports.GitHubAPIPort by calling the GitHub REST API.
// It caches responses per "owner/repo" key with a configurable TTL.
type Adapter struct {
	baseURL    string
	httpClient *http.Client
	ttl        time.Duration

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

type cacheEntry struct {
	tasks     []domain.Task
	fetchedAt time.Time
}

// NewAdapter constructs an Adapter that calls the given GitHub API base URL.
// Pass "https://api.github.com" in production; use an httptest stub URL in tests.
// Default TTL is 60 seconds.
func NewAdapter(baseURL string) *Adapter {
	return &Adapter{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 15 * time.Second},
		ttl:        60 * time.Second,
		cache:      make(map[string]cacheEntry),
	}
}

// WithTTL sets the cache TTL and returns the adapter for chaining.
func (a *Adapter) WithTTL(ttl time.Duration) *Adapter {
	a.ttl = ttl
	return a
}

// ListTasks fetches all task files from the .kanban/tasks directory of the
// given repository. Results are cached for the configured TTL.
func (a *Adapter) ListTasks(owner, repo string) ([]domain.Task, error) {
	key := owner + "/" + repo

	a.mu.RLock()
	entry, found := a.cache[key]
	a.mu.RUnlock()

	if found && time.Since(entry.fetchedAt) < a.ttl {
		return entry.tasks, nil
	}

	tasks, err := a.fetchTasks(owner, repo)
	if err != nil {
		return nil, err
	}

	a.mu.Lock()
	// Double-checked locking: re-check under write lock before overwriting.
	if existing, ok := a.cache[key]; !ok || time.Since(existing.fetchedAt) >= a.ttl {
		a.cache[key] = cacheEntry{tasks: tasks, fetchedAt: time.Now()}
	}
	a.mu.Unlock()

	return tasks, nil
}

// fileEntry is the shape of a single entry returned by the GitHub contents API.
type fileEntry struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	DownloadURL string `json:"download_url"`
}

// taskFrontMatter is the YAML front matter shape for reading task files.
type taskFrontMatter struct {
	ID        string `yaml:"id"`
	Title     string `yaml:"title"`
	Status    string `yaml:"status"`
	Priority  string `yaml:"priority"`
	Assignee  string `yaml:"assignee"`
	CreatedBy string `yaml:"created_by"`
}

func (a *Adapter) fetchTasks(owner, repo string) ([]domain.Task, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/.kanban/tasks", a.baseURL, owner, repo)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create contents request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github.object")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("contents request: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ports.ErrRepositoryNotFound
	case http.StatusForbidden:
		if resp.Header.Get("X-RateLimit-Remaining") == "0" {
			return nil, ports.ErrRateLimitExceeded
		}
		return nil, ports.ErrRepositoryNotFound
	case http.StatusTooManyRequests:
		return nil, ports.ErrRateLimitExceeded
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d from GitHub contents API", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read contents response: %w", err)
	}

	var entries []fileEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, fmt.Errorf("decode contents response: %w", err)
	}

	tasks := make([]domain.Task, 0, len(entries))
	for _, entry := range entries {
		if entry.Type != "file" || !strings.HasSuffix(entry.Name, ".md") {
			continue
		}
		task, err := a.downloadTask(entry.DownloadURL)
		if err != nil {
			return nil, fmt.Errorf("download task %s: %w", entry.Name, err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (a *Adapter) downloadTask(downloadURL string) (domain.Task, error) {
	resp, err := a.httpClient.Get(downloadURL)
	if err != nil {
		return domain.Task{}, fmt.Errorf("download: %w", err)
	}
	defer func() { _, _ = io.Copy(io.Discard, resp.Body); _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return domain.Task{}, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Task{}, fmt.Errorf("read task file: %w", err)
	}

	return parseTaskFile(content)
}

// parseTaskFile parses a Markdown file with YAML front matter into a domain.Task.
func parseTaskFile(data []byte) (domain.Task, error) {
	content := string(data)
	if !strings.HasPrefix(content, "---\n") {
		return domain.Task{}, fmt.Errorf("missing YAML front matter delimiter")
	}

	rest := content[4:]
	parts := strings.SplitN(rest, "\n---", 2)
	if len(parts) < 2 {
		return domain.Task{}, fmt.Errorf("unclosed YAML front matter")
	}

	var fm taskFrontMatter
	if err := yaml.Unmarshal([]byte(parts[0]), &fm); err != nil {
		return domain.Task{}, fmt.Errorf("parse front matter: %w", err)
	}

	status := domain.TaskStatus(fm.Status)
	if status == "" {
		status = domain.StatusTodo
	}

	return domain.Task{
		ID:        fm.ID,
		Title:     fm.Title,
		Status:    status,
		Priority:  fm.Priority,
		Assignee:  fm.Assignee,
		CreatedBy: fm.CreatedBy,
	}, nil
}

// compile-time interface check
var _ ports.GitHubAPIPort = (*Adapter)(nil)