# Acceptance Tests: kanban-web-view

## Test Architecture

Following the existing ATDD pattern (LMAX SimpleDSL style), adapted for the web backend.

### Layer Structure

```
tests/acceptance/backend/
  *_test.go                    # Test layer: Given/When/Then scenarios
  dsl/
    context.go                 # WebContext: server URL, HTTP client, cookies, last response
    step.go                    # Given/When/Then/And runners (reuse or mirror existing)
    board_steps.go             # DSL: view board, check columns, check cards
    card_steps.go              # DSL: view card detail, check fields
    auth_steps.go              # DSL: authenticate, check auth state
    task_steps.go              # DSL: add task, check creation
    setup.go                   # DSL: start server, configure repo, seed tasks
    assertions.go              # DSL: response assertions
  driver/
    server_driver.go           # Driver: start/stop kanban-web binary as subprocess
    http_driver.go             # Driver: HTTP client (GET, POST, parse HTML responses)
    repo_driver.go             # Driver: set up git repo with .kanban/tasks/, verify pushed commits
    github_stub_driver.go      # Driver: stub GitHub API for token validation (httptest server)
```

### Key Design Decisions

1. **Separate DSL and driver packages** — DSL speaks domain language ("view board", "add task"), driver speaks protocol ("HTTP GET /board", "parse HTML")
2. **WebContext** replaces the CLI Context — holds server URL, HTTP client, cookies, last HTTP response, last parsed HTML
3. **Server as subprocess** — the `kanban-web` binary is compiled and started before tests, just like the CLI tests compile and use the `kanban` binary
4. **GitHub API stub** — a local `httptest` server that stubs `GET /user` for token validation, so tests don't hit real GitHub
5. **Real git repo** — tests create a real temp git repo with `t.TempDir()` + `git init`, seed `.kanban/tasks/` files, and verify the server reads/writes them
6. **Server online is implied** — `NewWebContext` starts the server automatically; individual tests do not need a `ServerIsRunning` step
7. **No technology leakage** — tests never mention URLs, cookies, HTTP status codes, htmx, or HTML. The DSL hides all protocol details. Tests read like use case descriptions.
8. **All parameters optional with defaults** — `Task("title")` creates a todo task with sensible defaults. Only specify what the test cares about.

---

## US-WV-01: Hello World Go Server on GCP

### Test: Server starts and responds

```go
func TestHelloWorld_ServerResponds(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.When(ctx, dsl.IVisitTheBoard())
    dsl.Then(ctx, dsl.BoardIsVisible())
}
```

### Test: Server responds within acceptable time

```go
func TestHelloWorld_ResponseTime(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.When(ctx, dsl.IVisitTheBoard())
    dsl.Then(ctx, dsl.BoardLoadsWithin(500 * time.Millisecond))
}
```

---

## US-WV-02: View Board for Public Repository

### Test: Board shows three columns with tasks

```go
func TestViewBoard_ThreeColumnsWithTasks(t *testing.T) {
    ctx := dsl.NewWebContext(t)

    // dsl should only have string parameters
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Fix login bug", "status: in-progress", "assignee: alice@example.com"),
        dsl.Task("Write docs"),
        dsl.Task("Deploy v1", "status: done"),
    ))
    dsl.When(ctx, dsl.IVisitTheBoard())
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Write docs"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Doing", "title: Fix login bug"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Done", "title: Deploy v1"))
}
```

### Test: Cards show summary info

```go
func TestViewBoard_CardShowsSummary(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Fix login bug", "status: in-progress", "assignee: alice@example.com"),
    ))
    dsl.When(ctx, dsl.IViewCard("Fix login bug"))
    dsl.Then(ctx, dsl.CardShows("title: Fix login bug", "assignee: alice@example.com", "status: in-progress"))
}
```

### Test: Cards sorted by date within columns

```go
func TestViewBoard_CardsSortedByDate(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Older task", "created_at: 2026-03-01T00:00:00Z"),
        dsl.Task("Newer task", "created_at: 2026-03-15T00:00:00Z"),
    ))
    dsl.When(ctx, dsl.IVisitTheBoard())
    dsl.Then(ctx, dsl.CardAppearsBeforeInColumn("column: Todo", "first: Older task", "second: Newer task"))
}
```

### Test: Empty board shows three empty columns

```go
func TestViewBoard_EmptyBoard(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.When(ctx, dsl.IVisitTheBoard())
    dsl.Then(ctx, dsl.BoardHasColumns("Todo", "Doing", "Done"))
    dsl.Then(ctx, dsl.ColumnIsEmpty("Todo"))
    dsl.Then(ctx, dsl.ColumnIsEmpty("Doing"))
    dsl.Then(ctx, dsl.ColumnIsEmpty("Done"))
}
```

### Test: Board reflects changes from another user

```go
func TestViewBoard_ReflectsChangesFromAnotherUser(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.Users("developer", "stakeholder"))
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Initial task"),
    ))
    dsl.When(ctx, dsl.UserPushesNewTask("name: developer", "title: Late addition"))
    dsl.When(ctx, dsl.UserVisitsBoard("name: stakeholder"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Initial task", "title: Late addition"))
}
```

### Test: Unauthenticated user can view public board

```go
func TestViewBoard_UnauthenticatedUserCanView(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Public task"),
    ))
    dsl.Given(ctx, dsl.User("visitor"))
    dsl.When(ctx, dsl.UserVisitsBoard("name: visitor"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Public task"))
}
```

---

## US-WV-03: View Card Details

### Test: Card shows full details

```go
func TestCardDetail_ShowsFullInfo(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Fix login bug",
            "status: in-progress",
            "description: Users cannot log in with SSO",
            "priority: high",
            "assignee: alice@example.com",
            "created_by: Jonathan Sargent",
        ),
    ))
    dsl.When(ctx, dsl.IViewCard("Fix login bug"))
    dsl.Then(ctx, dsl.CardShows(
        "title: Fix login bug",
        "description: Users cannot log in with SSO",
        "priority: high",
        "assignee: alice@example.com",
        "created_by: Jonathan Sargent",
        "status: in-progress",
    ))
}
```

### Test: Missing optional fields handled gracefully

```go
func TestCardDetail_MissingOptionalFields(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Minimal task"),
    ))
    dsl.Given(ctx, dsl.User("visitor"))
    dsl.When(ctx, dsl.UserViewsCard("name: visitor", "title: Minimal task"))
    dsl.Then(ctx, dsl.CardShows("title: Minimal task"))
    dsl.Then(ctx, dsl.CardDoesNotShow("assignee", "priority"))
}
```

### Test: Task ID visible on card

```go
func TestCardDetail_TaskIDVisible(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Some task"),
    ))
    dsl.When(ctx, dsl.IViewCard("Some task"))
    dsl.Then(ctx, dsl.CardShowsTaskID())
}
```

---

## US-WV-04: Enter GitHub Token for Write Access

### Test: First-time user prompted to authenticate when adding task

```go
func TestAuth_FirstTimeUserPromptedOnAddTask(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("visitor"))
    dsl.When(ctx, dsl.UserTriesToAddTask("name: visitor"))
    dsl.Then(ctx, dsl.PromptedToAuthenticate())
}
```

### Test: User authenticates successfully

```go
func TestAuth_UserAuthenticatesSuccessfully(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Alice Johnson"))
    dsl.When(ctx, dsl.UserAuthenticates("name: Alice Johnson"))
    dsl.Then(ctx, dsl.IAmOnTheBoard())
    dsl.Then(ctx, dsl.ICanAddTasks())
}
```

### Test: Invalid credentials rejected

```go
func TestAuth_InvalidCredentialsRejected(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("bad-actor", "auth: invalid"))
    dsl.When(ctx, dsl.UserAttemptsToAuthenticate("name: bad-actor"))
    dsl.Then(ctx, dsl.AuthenticationIsRejected())
    dsl.Then(ctx, dsl.ICannotAddTasks())
}
```

### Test: Authenticated user stays authenticated

```go
func TestAuth_UserStaysAuthenticated(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Alice", "auth: valid"))
    dsl.When(ctx, dsl.UserTriesToAddTask("name: Alice"))
    dsl.Then(ctx, dsl.AddTaskFormIsShown())
}
```

### Test: Unauthenticated user can view but not add

```go
func TestAuth_UnauthenticatedCanViewNotAdd(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithTasks(
        dsl.Task("Visible task"),
    ))
    dsl.Given(ctx, dsl.User("visitor"))
    dsl.When(ctx, dsl.UserVisitsBoard("name: visitor"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Visible task"))
    dsl.Then(ctx, dsl.AddTaskOptionIsVisible())
}
```

---

## US-WV-05: Add Task via Web Interface

### Test: Add task with all fields

```go
func TestAddTask_AllFields(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Alice Johnson", "auth: valid"))
    dsl.When(ctx, dsl.UserAddsTask(
        "name: Alice Johnson",
        "title: Update onboarding docs",
        "description: The current docs are outdated",
        "priority: medium",
        "assignee: bob@example.com",
    ))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Update onboarding docs"))
    dsl.Then(ctx, dsl.TaskExistsInRepo("title: Update onboarding docs", "created_by: Alice Johnson"))
}
```

### Test: Add task with only title

```go
func TestAddTask_TitleOnly(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Bob", "auth: valid"))
    dsl.When(ctx, dsl.UserAddsTask("name: Bob", "title: Quick bug fix"))
    dsl.Then(ctx, dsl.ColumnContainsCards("column: Todo", "title: Quick bug fix"))
    dsl.Then(ctx, dsl.TaskExistsInRepo("title: Quick bug fix", "created_by: Bob", "status: todo"))
}
```

### Test: Title is required

```go
func TestAddTask_TitleRequired(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Alice", "auth: valid"))
    dsl.When(ctx, dsl.UserAddsTask("name: Alice", "title: "))
    dsl.Then(ctx, dsl.TaskCreationFails("title is required"))
    dsl.Then(ctx, dsl.NoNewTaskInRepo())
}
```

### Test: Task file follows existing format

```go
func TestAddTask_FileFollowsFormat(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("Alice", "auth: valid"))
    dsl.When(ctx, dsl.UserAddsTask("name: Alice", "title: Format check task", "priority: high"))
    dsl.Then(ctx, dsl.TaskExistsInRepo("title: Format check task"))
    dsl.Then(ctx, dsl.TaskFileIsValidFormat())
    dsl.Then(ctx, dsl.TaskHasSequentialID())
}
```

### Test: Task is committed and pushed to remote

```go
func TestAddTask_CommittedAndPushed(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithRemote())
    dsl.Given(ctx, dsl.User("Alice", "auth: valid"))
    dsl.When(ctx, dsl.UserAddsTask("name: Alice", "title: Pushed task"))
    dsl.Then(ctx, dsl.RemoteRepoContainsTask("Pushed task"))
}
```

### Test: Unauthenticated user cannot add tasks

```go
func TestAddTask_UnauthenticatedRejected(t *testing.T) {
    ctx := dsl.NewWebContext(t)
    dsl.Given(ctx, dsl.ARepoWithNoTasks())
    dsl.Given(ctx, dsl.User("visitor"))
    dsl.When(ctx, dsl.UserAddsTask("name: visitor", "title: Sneaky task"))
    dsl.Then(ctx, dsl.TaskCreationIsRejected())
    dsl.Then(ctx, dsl.NoNewTaskInRepo())
}
```

---

## Driver Layer Detail

### ServerDriver

```go
// server_driver.go
// Compiles and starts kanban-web as a subprocess.
// Manages lifecycle: start, health check, stop.
// Provides: ServerURL(), Start(), Stop(), WaitForHealthy()

type ServerDriver struct {
    cmd          *exec.Cmd
    url          string
    binPath      string
    repoDir      string
    syncInterval time.Duration
    cookieKey    string
}
```

**Responsibilities**:
- Compile `kanban-web` binary (or use pre-compiled via env var `KANBAN_WEB_BIN`)
- Start as subprocess with configured repo, sync interval, cookie key
- Wait for `/healthz` to return 200 before tests proceed
- Stop on test cleanup

### HTTPDriver

```go
// http_driver.go
// Wraps net/http.Client with cookie jar.
// Provides: GET, POST, ParseHTML, ExtractElements

type HTTPDriver struct {
    client  *http.Client
    baseURL string
}
```

**Responsibilities**:
- HTTP GET/POST with automatic cookie handling
- Parse HTML responses (use `golang.org/x/net/html` or similar)
- Extract elements by CSS-like selectors for assertions
- Track response time, status code, body

### RepoDriver

```go
// repo_driver.go
// Creates and manages temporary git repos for tests.
// Provides: CreateRepo, SeedTasks, VerifyTaskFile, VerifyPushed

type RepoDriver struct {
    t       *testing.T
    repoDir string
    bareDir string  // bare remote for push verification
}
```

**Responsibilities**:
- `git init` temp repo + bare remote
- Write `.kanban/` structure and task files
- Verify task files after server writes
- Verify commits pushed to bare remote

### GitHubStubDriver

```go
// github_stub_driver.go
// httptest server that stubs GitHub API endpoints.
// Provides: StubTokenValidation, RejectToken

type GitHubStubDriver struct {
    server       *httptest.Server
    validTokens  map[string]string  // token -> username
}
```

**Responsibilities**:
- Stub `GET /user` — return user info for valid tokens, 401 for invalid
- Configurable per-test: which tokens are valid
- Server URL injected into `kanban-web` via env var (e.g., `KANBAN_WEB_GITHUB_API_URL`)

---

## DSL Method Summary

### Setup Steps (Given)
| Step | Description |
|------|-------------|
| `ARepoWithTasks(tasks...)` | Create temp repo with seeded tasks |
| `ARepoWithNoTasks()` | Create temp repo with empty `.kanban/tasks/` |
| `ARepoWithRemote()` | Create temp repo + bare remote for push tests |
| `User(name)` | Unauthenticated user — no session stored |
| `User(name, "auth: valid")` | User with an existing valid auth session |
| `User(name, "auth: invalid")` | User with a stored token rejected by GitHub |
| `Users(names...)` | Register multiple users for multi-user tests |

### Action Steps (When)
| Step | Description |
|------|-------------|
| `IVisitTheBoard()` | Visit the board (no user identity required) |
| `IViewCard(title)` | View a card's details (no user identity required) |
| `UserVisitsBoard(params...)` | Named user visits the board |
| `UserViewsCard(params...)` | Named user views a specific card |
| `UserTriesToAddTask(params...)` | Named user attempts to add a task (may trigger auth prompt) |
| `UserAddsTask(params...)` | Named user adds a task with given fields |
| `UserAuthenticates(params...)` | Named user submits valid credentials |
| `UserAttemptsToAuthenticate(params...)` | Named user submits credentials (may be rejected) |
| `UserPushesNewTask(params...)` | Named user pushes a new task via git |

### Assertion Steps (Then)
| Step | Description |
|------|-------------|
| `BoardIsVisible()` | Board page loaded successfully |
| `BoardLoadsWithin(duration)` | Board loads within time limit |
| `BoardHasColumns(names...)` | Board has named columns |
| `ColumnContainsCards(params...)` | Column contains specified cards |
| `ColumnIsEmpty(column)` | Column has no cards |
| `CardAppearsBeforeInColumn(params...)` | Card ordering in column |
| `CardShows(params...)` | Card displays specified fields |
| `CardDoesNotShow(fields...)` | Card hides empty fields |
| `CardShowsTaskID()` | Task ID visible on card |
| `PromptedToAuthenticate()` | User is asked to authenticate |
| `AuthenticationIsRejected()` | Authentication attempt failed |
| `IAmOnTheBoard()` | User is on the board page |
| `ICanAddTasks()` | User has write access |
| `ICannotAddTasks()` | User does not have write access |
| `AddTaskFormIsShown()` | Add task form is displayed |
| `AddTaskOptionIsVisible()` | Option to add tasks is visible |
| `TaskCreationFails(reason)` | Task creation rejected with reason |
| `TaskCreationIsRejected()` | Task creation denied (auth) |
| `TaskExistsInRepo(params...)` | Task file exists with specified fields |
| `TaskFileIsValidFormat()` | Task file has correct Markdown + YAML format |
| `TaskHasSequentialID()` | Task ID follows TASK-NNN pattern |
| `NoNewTaskInRepo()` | No new task files created |
| `RemoteRepoContainsTask(title)` | Task pushed to remote repo |

---

## DSL Parameter Convention

Following the SimpleDSL pattern from `atdd/`, all parameters are optional with good defaults:

```go
// Task: title is the only positional param. Everything else is optional named.
dsl.Task("Fix login bug")
// Default: status=todo, no assignee, no priority, no description
dsl.Task("Fix login bug", "status: in-progress", "assignee: alice@example.com")
// Override only what the test cares about

// User: name is the positional param. auth is optional named.
dsl.User("visitor")                      // unauthenticated, no session
dsl.User("Alice", "auth: valid")         // pre-authenticated
dsl.User("bad-actor", "auth: invalid")   // invalid token

// User actions: all params named, name is always required.
dsl.UserAddsTask("name: Alice", "title: Update docs", "priority: medium")
dsl.UserVisitsBoard("name: stakeholder")
dsl.UserViewsCard("name: visitor", "title: Fix login bug")
dsl.UserPushesNewTask("name: developer", "title: Late addition")

// Assertions: multi-value params where applicable.
dsl.ColumnContainsCards("column: Done", "title: Deploy v1", "title: Ship release")
dsl.CardShows("title: Fix bug", "assignee: alice@example.com", "status: in-progress")
```
### Defaults

| Parameter | Default |
|-----------|---------|
| status | `todo` |
| priority | (empty — not shown) |
| assignee | (empty — not shown) |
| description | (empty — not shown) |
| created_by | (context user or empty) |
| created_at | (current time) |
