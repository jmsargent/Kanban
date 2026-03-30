# Solution Testing: kanban-web-view

## Feature ID
kanban-web-view

## Solution Concept
Lightweight Go HTTP server deployed on a GCP Compute Engine e2-micro instance (free tier), serving a kanban board web page with full card details, reading task data from a local git clone of the GitHub repo.

## Hypotheses

### H1: GCP Compute Engine Free Tier Deployment
**We believe** deploying a Go HTTP server to a GCP e2-micro instance **will achieve** a reliable, always-on web endpoint with no cold start delays.
**We will know this is TRUE when** the server responds to HTTP requests consistently at a public IP/domain with <500ms response times.
**We will know this is FALSE when** the e2-micro instance lacks sufficient resources (CPU, RAM) or the free tier constraints make it impractical.

| Risk category | Feasibility |
|---------------|-------------|
| Test method | Spike: provision e2-micro, deploy hello-world Go binary, verify uptime over 24 hours |
| Effort | 1-2 hours |
| Success criteria | Server responds at public URL, <500ms latency, no crashes over 24 hours |

### H2: Local Git Clone + Periodic Pull
**We believe** maintaining a local git clone on the VM with periodic `git pull` **will achieve** near-real-time board data with minimal latency.
**We will know this is TRUE when** task file changes pushed to GitHub appear on the web board within 60 seconds.
**We will know this is FALSE when** git pull operations consume excessive CPU/RAM on the e2-micro, or sync delays exceed 2 minutes.

| Risk category | Feasibility |
|---------------|-------------|
| Test method | Spike: clone a test repo on e2-micro, run git pull every 30 seconds, measure resource usage |
| Effort | 1-2 hours |
| Success criteria | Sync latency <60 seconds, git pull CPU usage negligible, RAM stays within 1 GB |

### H3: Private Repo Access via Deploy Key
**We believe** using an SSH deploy key (read-only) on the VM **will achieve** access to private repo task files without exposing credentials to viewers.
**We will know this is TRUE when** the VM clones and pulls from a private repo, and no credentials are visible in the served HTML.
**We will know this is FALSE when** deploy key setup is fragile, rotation is painful, or GitHub changes access patterns.

| Risk category | Feasibility |
|---------------|-------------|
| Test method | Spike: configure deploy key on e2-micro, clone private repo, verify pull works |
| Effort | 30 minutes |
| Success criteria | Private repo cloned and updated, key not exposed in any served content |

### H4: Full Card Display Usability
**We believe** showing full card information (title, status, assignee, dates, description) on a web board **will achieve** sufficient visibility for non-technical team members.
**We will know this is TRUE when** a non-technical person can identify: which tasks exist, their status, who is working on them, and what each task is about -- without asking a developer.
**We will know this is FALSE when** the information density is overwhelming, or the layout is confusing, or key context is missing.

| Risk category | Usability |
|---------------|-----------|
| Test method | Show rendered board to 2-3 people, ask them to answer questions about project status |
| Effort | 1-2 hours per test session |
| Success criteria | >80% task completion (can answer status questions correctly), <10 sec to find a specific task |

### H5: Domain Code Reuse
**We believe** the kanban hexagonal architecture allows the web server to reuse `internal/domain` and `internal/usecases` packages **without modification**.
**We will know this is TRUE when** the web server imports and uses existing task parsing and board logic directly.
**We will know this is FALSE when** domain code has dependencies on filesystem or CLI that prevent reuse.

| Risk category | Feasibility |
|---------------|-------------|
| Test method | Code review of domain/usecases packages for portability |
| Effort | 1 hour |
| Success criteria | Domain types and board logic importable by a new HTTP adapter with zero changes to existing packages |

## Test Priority

| Hypothesis | Risk Score | Test Order | Rationale |
|-----------|------------|------------|-----------|
| H1 | High | First | If GCP free tier doesn't work, entire approach fails |
| H2 | High | Second | If local git clone + pull is unreliable, no board data |
| H5 | Medium | Third | Validates architecture benefit -- if domain can't be reused, effort increases significantly |
| H3 | Medium | With H2 | Trivial extension of H2, critical for private repos |
| H4 | Medium | After working prototype | Requires working board to test usability |

## Experiment Plan

### Sprint 1: Feasibility (H1 + H2 + H3 + H5)
- Provision GCP e2-micro, deploy hello-world Go server (H1)
- Clone repo locally on VM, set up periodic git pull (H2 + H3)
- Verify domain code portability (H5)
- **Gate**: All 4 hypotheses validated before proceeding to full board

### Sprint 2: Working Board (H4 prep)
- Server-rendered HTML board with columns and full card details
- Connected to a real repo via local git clone
- Share URL with 2-3 people for usability testing (H4)

### Sprint 3: Refinement
- Address usability findings from H4
- Add process supervision (systemd service) for reliability
- Document deployment process for other developers

## Phase 3 Status
- **Hypotheses defined**: 5
- **High-risk (test first)**: H1, H2
- **Experiment plan**: 3 sprints, feasibility first
- **Users to test with**: 2-3 people (founder + peers) for usability validation

## Gate G3 Evaluation (Pre-Testing)

| Criteria | Status | Notes |
|----------|--------|-------|
| 5+ users tested | PENDING | Adapted to 2-3 for pre-market solo developer |
| >80% task completion | PENDING | Requires working prototype |
| Core flow usable | PENDING | Requires prototype |
| Value + feasibility confirmed | PENDING | Requires spikes |

**Gate G3 Status: PENDING -- hypotheses defined, testing not yet started**

Note: G3 will be evaluated after Sprint 2 usability testing. For a pre-market solo developer project, the "5+ users" threshold is adapted to "2-3 testers including at least 1 non-technical person."
