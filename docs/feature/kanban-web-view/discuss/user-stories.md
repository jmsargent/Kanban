# User Stories: kanban-web-view

## Feature Summary

Web-based kanban board that allows non-technical team members to view and interact with the git-native kanban board through a browser. Go HTTP server on GCP Compute Engine e2-micro (free tier).

## Auth Model

- **Public repos**: View board without authentication
- **Adding tasks**: Requires a GitHub Personal Access Token
- **Token entry**: User enters token + display name on first write attempt
- **Token storage decision**: Postponed (server-side vs client-side TBD)
- **Later**: GitHub OAuth replaces token entry

## Dependencies

```
US-WV-01 (Hello World)
    └── US-WV-02 (View Board)
            └── US-WV-03 (Card Details)
            └── US-WV-04 (Token Entry)
                    └── US-WV-05 (Add Task)
```

---

## US-WV-01: Hello World Go Server on GCP

**As a** developer setting up kanban-web-view,
**I want** a Go HTTP server deployed on GCP Compute Engine e2-micro,
**so that** I can validate the deployment pipeline and confirm the free tier is sufficient.

**Effort**: 1 day
**Dependencies**: None (walking skeleton)

### Acceptance Criteria

```gherkin
Scenario: Server responds at public URL
  Given a Go HTTP server is deployed to GCP Compute Engine e2-micro
  When I send an HTTP GET request to the server's public IP
  Then I receive an HTTP 200 response
  And the response body contains "kanban-web-view"

Scenario: Server stays running
  Given the server has been deployed for 24 hours
  When I send an HTTP GET request to the server's public IP
  Then I receive an HTTP 200 response

Scenario: Response time is acceptable
  Given the server is running on GCP e2-micro
  When I send an HTTP GET request
  Then the response time is less than 500ms
```

### DoR Checklist
- [x] Acceptance criteria defined
- [x] Dependencies identified (none)
- [x] Effort estimated
- [ ] GCP project and billing set up (student free tier)

---

## US-WV-02: View Board for Public Repository

**As a** non-technical team member,
**I want** to see the kanban board in my browser by visiting a URL,
**so that** I can understand the current state of the project without using the CLI or navigating GitHub.

**Effort**: 1-2 days
**Dependencies**: US-WV-01

### Acceptance Criteria

```gherkin
Scenario: Board renders with three columns
  Given a public GitHub repo contains tasks in ".kanban/tasks/"
  And the server is configured to read from that repo
  When I visit the board URL in a browser
  Then I see three columns labeled "Todo", "Doing", and "Done"
  And each task appears as a card in its corresponding column

Scenario: Cards display summary information
  Given a task file exists with title "Fix login bug", status "in-progress", and assignee "alice@example.com"
  When I view the board
  Then the card in the "Doing" column shows the title "Fix login bug"
  And the card shows the assignee "alice@example.com"

Scenario: Cards are sorted by date within columns
  Given two tasks exist in the "todo" status
  And TASK-001 was created before TASK-002
  When I view the board
  Then TASK-001 appears above TASK-002 in the "Todo" column

Scenario: Board reflects recent changes
  Given a new task is pushed to the repo
  When I refresh the board after the server's sync interval
  Then the new task appears on the board

Scenario: Empty board
  Given a public repo with ".kanban/tasks/" containing no task files
  When I visit the board URL
  Then I see three empty columns labeled "Todo", "Doing", and "Done"

Scenario: No authentication required for public repos
  Given the configured repo is public on GitHub
  When I visit the board URL without any credentials
  Then I see the board without being prompted for authentication
```

### DoR Checklist
- [x] Acceptance criteria defined
- [x] Dependencies identified (US-WV-01)
- [x] Effort estimated
- [x] Task file format understood (Markdown + YAML front matter)
- [x] Domain model reviewed (Task struct in internal/domain/task.go)

---

## US-WV-03: View Card Details

**As a** non-technical team member,
**I want** to click on a card and see its full details,
**so that** I can understand what a task is about, who is working on it, and its priority.

**Effort**: 1 day
**Dependencies**: US-WV-02

### Acceptance Criteria

```gherkin
Scenario: Click card to see full details
  Given the board is displayed with cards
  When I click on a card with title "Fix login bug"
  Then I see a detail view showing:
    | Field       | Value                |
    | Title       | Fix login bug        |
    | Description | (task body content)  |
    | Status      | in-progress          |
    | Priority    | high                 |
    | Assignee    | alice@example.com    |
    | Created by  | Jonathan Sargent     |

Scenario: Description displayed as plain text
  Given a task has a description containing markdown-like formatting
  When I view the card details
  Then the description is displayed as plain text (not rendered as markdown)

Scenario: Missing optional fields
  Given a task has no assignee and no priority set
  When I view the card details
  Then the assignee and priority fields are either hidden or show a placeholder
  And the detail view does not display empty values confusingly

Scenario: Close detail view returns to board
  Given I am viewing a card's details
  When I close the detail view
  Then I see the full board again

Scenario: Task ID is visible but not prominent
  Given a task with ID "TASK-042"
  When I view the card details
  Then the task ID is visible but less prominent than the title
```

### DoR Checklist
- [x] Acceptance criteria defined
- [x] Dependencies identified (US-WV-02)
- [x] Effort estimated
- [x] Card fields confirmed: title, description, status, priority, assignee, created_by, due date

---

## US-WV-04: Enter GitHub Token for Write Access

**As a** non-technical team member who wants to add tasks,
**I want** to enter my GitHub token and display name,
**so that** I can authenticate and create tasks on the board.

**Effort**: 1 day
**Dependencies**: US-WV-02

### Acceptance Criteria

```gherkin
Scenario: Prompted for token when attempting write action
  Given I am viewing a public board without having entered a token
  When I click "Add Task"
  Then I am prompted to enter a GitHub Personal Access Token
  And I am prompted to enter my display name

Scenario: Token accepted and stored
  Given I am on the token entry form
  When I enter a valid GitHub token and my display name
  And I submit the form
  Then I am returned to the board
  And I can now perform write actions

Scenario: Invalid token rejected
  Given I am on the token entry form
  When I enter an invalid or expired GitHub token
  And I submit the form
  Then I see an error message indicating the token is invalid
  And I am not granted write access

Scenario: Token persists across page refreshes
  Given I have previously entered a valid token
  When I refresh the page
  Then I still have write access without re-entering the token

Scenario: View-only access without token
  Given I have not entered a token
  When I browse the board
  Then I can view the board and card details
  But I cannot add or modify tasks
```

### DoR Checklist
- [x] Acceptance criteria defined
- [x] Dependencies identified (US-WV-02)
- [x] Effort estimated
- [ ] Token storage decision postponed (server-side session vs client-side localStorage)

---

## US-WV-05: Add Task via Web Interface

**As a** non-technical team member,
**I want** to add a new task through the web interface,
**so that** I can contribute to the project board without using the CLI or git.

**Effort**: 1-2 days
**Dependencies**: US-WV-04

### Acceptance Criteria

```gherkin
Scenario: Add a task with all fields
  Given I have entered a valid GitHub token
  When I click "Add Task"
  And I fill in:
    | Field       | Value                        |
    | Title       | Update onboarding docs       |
    | Description | The current docs are outdated |
    | Priority    | medium                       |
    | Assignee    | bob@example.com              |
  And I submit the form
  Then a new task file is created in ".kanban/tasks/" in the repo
  And the file is committed and pushed to the repo
  And the new card appears in the "Todo" column on the board

Scenario: Add a task with only a title
  Given I have entered a valid GitHub token
  When I click "Add Task"
  And I fill in only the title "Quick bug fix"
  And I submit the form
  Then a new task is created with status "todo"
  And optional fields are left empty
  And the "created_by" field is set to my display name

Scenario: Title is required
  Given I am on the add task form
  When I submit without entering a title
  Then I see a validation error indicating the title is required
  And no task is created

Scenario: Task file follows existing format
  Given I add a task via the web interface
  When I inspect the committed file in the repo
  Then it follows the Markdown + YAML front matter format
  And it has a sequential task ID (e.g., TASK-015 if TASK-014 was the last)

Scenario: Created by is set from display name
  Given I entered "Alice Johnson" as my display name during token setup
  When I add a new task
  Then the "created_by" field in the task file is "Alice Johnson"

Scenario: Token lacks write permission
  Given I have entered a GitHub token without repo write permissions
  When I try to add a task
  Then I see an error message indicating insufficient permissions
  And no task is created
```

### DoR Checklist
- [x] Acceptance criteria defined
- [x] Dependencies identified (US-WV-04)
- [x] Effort estimated
- [x] Task file format confirmed (matches CLI output)
- [x] Task ID generation understood (sequential)

---

## Deferred Stories

### US-WV-06: GitHub OAuth Login (Deferred)
Replace manual token entry with "Login with GitHub" OAuth flow. Provides smoother UX and eliminates token copy-pasting. Requires registering a GitHub OAuth App.

### US-WV-07: Private Repo Support (Deferred)
Support viewing boards from private repositories. Requires authentication (token or OAuth) for all operations including viewing.

---

## Effort Summary

| Story | Effort | Release |
|-------|--------|---------|
| US-WV-01 | 1 day | Walking Skeleton |
| US-WV-02 | 1-2 days | Release 1 |
| US-WV-03 | 1 day | Release 1 |
| US-WV-04 | 1 day | Release 2 |
| US-WV-05 | 1-2 days | Release 2 |
| **Total** | **5-7 days** | |
