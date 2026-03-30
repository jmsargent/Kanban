# Journey Map: kanban-web-view

## Journey 1: View the Board (Public Repo)

**Persona**: Non-technical team member (PO, stakeholder)
**Goal**: Understand current project status without asking a developer
**Trigger**: Developer shares the board URL with the team

### Steps

| Step | Action | Thinking | Feeling | Touchpoint |
|------|--------|---------|---------|------------|
| 1 | Receives board URL from developer | "Let me check this out" | Curious | Slack/email |
| 2 | Opens URL in browser | "What will this look like?" | Expectant | Browser |
| 3 | Sees three-column board immediately | "Oh, I can see what's going on" | Relieved, oriented | Board page |
| 4 | Scans columns: Todo, Doing, Done | "There are 3 things in progress" | Informed | Board columns |
| 5 | Clicks a card to see details | "What exactly is this task about?" | Curious | Card detail view |
| 6 | Reads description, sees assignee | "Alice is working on the login fix" | Satisfied | Card detail view |
| 7 | Closes detail, continues scanning | "I have a good picture now" | Confident | Board page |

### Emotional Arc

```
Curious → Expectant → Relieved → Informed → Satisfied → Confident
   1          2           3          4          5-6          7
```

**Peak moment**: Step 3 — seeing the board immediately with no setup. This must feel instant and clear.

**Risk moment**: Step 3 — if the board is slow to load, confusing, or looks like a developer tool, the user bounces.

### Error Paths

| Error | What happens | Recovery |
|-------|-------------|----------|
| Board URL is wrong | 404 page | User asks developer for correct URL |
| Repo has no tasks | Empty board with three columns | User understands there are no tasks yet |
| Server is down | Connection error | User tries again later or contacts developer |

---

## Journey 2: Add a Task (Requires Token)

**Persona**: Non-technical team member who wants to contribute
**Goal**: Add a bug report, feature request, or task to the board
**Trigger**: User spots something missing from the board or has a new request

### Steps

| Step | Action | Thinking | Feeling | Touchpoint |
|------|--------|---------|---------|------------|
| 1 | Viewing the board | "I need to add a task for this" | Motivated | Board page |
| 2 | Clicks "Add Task" button | "How do I add something?" | Hopeful | Board page |
| 3 | Sees token entry prompt (first time only) | "What's a GitHub token?" | Confused/uncertain | Token form |
| 4 | Enters GitHub token + display name | "I hope this works" | Cautious | Token form |
| 5 | Token accepted, form appears | "Okay, I'm in" | Relieved | Add task form |
| 6 | Fills in title and description | "This is straightforward" | Comfortable | Add task form |
| 7 | Optionally fills priority, assignee | "I'll set priority to high" | Engaged | Add task form |
| 8 | Submits the form | "Did it work?" | Anticipating | Add task form |
| 9 | Card appears on the board in Todo column | "It's there! I did it." | Accomplished | Board page |

### Emotional Arc

```
Motivated → Hopeful → Confused → Cautious → Relieved → Comfortable → Accomplished
    1           2         3          4           5          6-7            8-9
```

**Peak moment**: Step 9 — seeing the card appear on the board. Immediate feedback is critical.

**Pain point**: Step 3 — the token entry is the biggest friction point. The user may not know what a GitHub token is. This is where GitHub OAuth (US-WV-06) will improve the experience later.

**Mitigation for Step 3**: Include a brief explanation and a link to GitHub's token creation page. Keep the instructions simple and non-technical.

### Error Paths

| Error | What happens | Recovery |
|-------|-------------|----------|
| Invalid token | Error message: "This token is invalid" | User re-enters or asks developer for help |
| Token lacks write permission | Error message: "Token needs write access" | User creates a new token with correct scope |
| Title left empty | Validation error on form | User fills in required title |
| Push fails (network) | Error message: "Could not save task" | User retries |

### Returning User Flow (Token Already Saved)

Steps 3-4 are skipped on subsequent visits. The user goes directly from "Add Task" (step 2) to the form (step 5). This makes the experience much smoother after the first time.

---

## Journey 3: Developer Sets Up the Web View

**Persona**: Developer who uses the kanban CLI
**Goal**: Deploy the web view so the team can see the board
**Trigger**: Team needs visibility into the kanban board

This journey is not a user-facing feature but defines the setup experience.

| Step | Action |
|------|--------|
| 1 | Provision GCP Compute Engine e2-micro |
| 2 | Deploy the Go binary to the VM |
| 3 | Configure the server with the public repo URL |
| 4 | Verify the board loads at the VM's public IP |
| 5 | Share the URL with the team |

This journey is covered by US-WV-01 (deployment) and the deployment documentation that will be produced during the DEVOPS wave.
