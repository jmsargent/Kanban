# Opportunity Solution Tree: kanban-web-view

## Feature ID
kanban-web-view

## Desired Outcome
Enable non-technical team members to view the kanban board without developer assistance, driving adoption from "developer tool" to "team tool."

## Opportunity Map

```
Desired Outcome: Team-wide kanban visibility without CLI dependency
  |
  +-- O1: Non-technical people cannot see the board without navigating GitHub (Score: 14)
  |     +-- [SELECTED] S-B: Go backend on GCP Compute Engine reading task data from local git clone
  |     +-- S-A: Static site generated via GitHub Actions (rejected: user prefers backend)
  |     +-- S-C: CLI-generated static HTML (stepping stone only)
  |
  +-- O2: Private repos block unauthenticated access to board data (Score: 13)
  |     +-- [SELECTED] S-E: Go backend with git clone using deploy key or token (server-side)
  |     +-- S-D: GitHub Actions pre-render (viable but rejected with static approach)
  |
  +-- O3: Mermaid diagram lacks detail -- user wants full card information (Score: 11)
  |     +-- [SELECTED] S-G: Server-rendered HTML board with full card details
  |     +-- S-H: Enhanced Mermaid with links (insufficient -- user wants "full information")
  |
  +-- O4: Developers cannot demo the tool to non-technical teammates (Score: 10)
  |     +-- [SELECTED] S-I: Shareable URL from the Go backend
  |
  +-- O5: Board state is stale between commits (Score: 7)
        +-- [SELECTED] S-K: Server does periodic git pull (always-on VM = fast sync, no cold start)
```

## Opportunity Scores

Scoring: Score = Importance + Max(0, Importance - Satisfaction)
Source: Founder interviews (3 rounds). Pre-market -- no end-user validation.

| # | Opportunity | Importance | Satisfaction | Score | Action |
|---|------------|------------|--------------|-------|--------|
| O1 | Cannot see board without GitHub | 8 | 2 | 14 | Pursue |
| O2 | Private repos block access | 8 | 3 | 13 | Pursue |
| O3 | Mermaid lacks detail (full info needed) | 7 | 4 | 11 | Pursue |
| O4 | Cannot demo to non-technical teammates | 7 | 3 | 10 | Evaluate |
| O5 | Board state stale between commits | 5 | 3 | 7 | Deprioritize |

## Chosen Direction: Go Backend on GCP Compute Engine

### Why This Direction
The user explicitly preferred a backend approach, and after further consideration chose Google Cloud over Azure for lower latency:

1. **Always-on VM**: Compute Engine e2-micro is always running -- no cold starts, no function wake-up delays
2. **Local git clone**: Server maintains a local clone of the repo, does periodic `git pull` -- faster than API calls
3. **Lowest sync latency**: Always-on + local clone = board updates within seconds of a push (via webhook or short polling interval)
4. **Full card info**: requires server-side rendering, not minimal static HTML
5. **Future extensibility**: backend supports eventual write operations
6. **Private repo access**: server holds deploy key or token, no client-side auth complexity

### Architecture: Go Backend on GCP Compute Engine

```
GitHub Repo (.kanban/tasks/)
         |
         | git pull (periodic or webhook-triggered)
         v
Local git clone on Compute Engine e2-micro
         |
         | Go binary reads .kanban/tasks/ from filesystem
         v
Go HTTP Server (server-rendered HTML)
         |
         | HTTPS
         v
Browser (any team member with the URL)
```

### Key Architecture Properties

| Property | Value |
|----------|-------|
| Language | Go (reuses kanban domain code directly) |
| Hosting | GCP Compute Engine e2-micro (free tier eligible) |
| Data source | Local git clone of the repo (filesystem read) |
| Repo sync | Periodic `git pull` (e.g., every 30s) or GitHub webhook |
| Auth for private repos | SSH deploy key or GitHub token (server-side) |
| Auth for viewers | Deferred -- public URL initially |
| Frontend | Server-rendered HTML (Go templates) |
| Card detail level | Full: title, status, assignee, dates, description |
| Freshness | Seconds (always-on, no cold start) |

### Advantages of Compute Engine Over Serverless/PaaS

| Dimension | Compute Engine e2-micro | Serverless (Cloud Run / Azure Functions) |
|-----------|------------------------|------------------------------------------|
| Cold start | None (always on) | 1-10 seconds |
| Repo sync | Local git clone (instant reads) | GitHub API per request (network latency) |
| Latency | Low and consistent | Variable |
| Free tier | 1 e2-micro instance free (us-central1, etc.) | Limited free invocations |
| Complexity | Manage VM (updates, uptime) | Managed infrastructure |
| Cost risk | Predictable (free tier is monthly) | Pay-per-request can surprise |

### Incremental Delivery Plan

```
Step 1: Hello World
  Go HTTP server deployed to GCP Compute Engine e2-micro.
  Returns "Hello, kanban-web-view."
  Validates: GCP deployment works, free tier is sufficient, VM stays up.

Step 2: Board Reading
  Server clones repo, reads .kanban/tasks/ from local filesystem.
  Renders a basic HTML board with columns (todo, in-progress, done).
  Periodic git pull to keep data fresh.
  Validates: Task parsing, board rendering, sync latency.

Step 3: Full Card Display
  Cards show all fields: title, status, assignee, created/updated dates, description.
  Clickable cards expand to show full detail.
  Validates: Usability for non-technical users, information sufficiency.
```

### Constraints

| Constraint | Detail |
|-----------|--------|
| Budget | $0 -- GCP free tier only |
| GCP e2-micro free tier | 1 instance, 0.25 vCPU, 1 GB RAM, 30 GB disk, select regions only |
| Network egress | 1 GB/month free (sufficient for small team board views) |
| Private repos | Require SSH deploy key or token with repo access |
| VM management | User responsible for OS updates, process supervision |

## Phase 2 Status
- **Interviews**: 3 rounds (founder)
- **Opportunities identified**: 5
- **Top opportunities**: O1 (14), O2 (13), O3 (11) -- all >8
- **Solution direction**: Go backend on GCP Compute Engine (user-selected)
- **Delivery approach**: Incremental (hello world -> board reading -> full cards)

## Gate G2 Evaluation

| Criteria | Status | Notes |
|----------|--------|-------|
| OST complete | PASS | 5 opportunities mapped with solutions |
| Top 2-3 score >8 | PASS | O1=14, O2=13, O3=11 |
| Team aligned | PASS (adapted) | Single founder, clear stated preference for backend on GCP |

**Gate G2 Status: PASSED**
