# Wave Decisions: kanban-web-view

## Feature ID
kanban-web-view

## Discovery Wave Summary
**Duration**: 2026-03-29 (single session, 3 interview rounds)
**Approach**: Pre-market adapted discovery (founder interviews, no external users)
**Outcome**: GO with conditions

---

## Key Decisions

### D1: Go Backend Architecture (Not Static Site)
**Decision**: Build a Go HTTP server that reads task files and serves a web board.
**Rationale**: User explicitly preferred backend over static site generation. Enables future write operations, reuses kanban domain code directly, and provides server-side GitHub token handling for private repos.
**Alternatives rejected**: Static site via GitHub Actions (user preference), CLI-generated HTML (insufficient for ongoing use).

### D2: GCP Compute Engine e2-micro (Not Azure, Not Serverless)
**Decision**: Deploy on GCP Compute Engine e2-micro free tier instance.
**Rationale**: User initially preferred Azure but switched to GCP after considering latency. An always-on VM avoids cold starts that serverless/PaaS functions would impose. Local git clone enables filesystem-speed reads instead of API-per-request.
**Alternatives rejected**: Azure App Service/Functions (cold starts), GCP Cloud Run (cold starts), GitHub Pages (static only).

### D3: Local Git Clone (Not GitHub API)
**Decision**: Server maintains a local git clone with periodic `git pull`, reads `.kanban/tasks/` from the filesystem.
**Rationale**: Follows naturally from D2. Always-on VM can maintain persistent state. Local reads are faster than GitHub API calls. No API rate limit concerns. Works identically for public and private repos (deploy key).
**Alternative rejected**: GitHub Contents API per request (latency, rate limits, auth complexity).

### D4: Full Card Information
**Decision**: Web board displays all task fields -- title, status, assignee, dates, description.
**Rationale**: User explicitly requested "full information of the cards." No minimal/summary view.
**Implication**: Requires card detail view (expand/click), not just column headers.

### D5: Authentication Deferred
**Decision**: No authentication for MVP. Public URL, anyone with the link can view.
**Rationale**: User explicitly said "authentication can come later." Reduces scope and complexity.
**Risk accepted**: Anyone with the URL can see the board. Acceptable for now.
**Future**: Add auth when needed (OAuth, basic auth, or similar).

### D6: Incremental Delivery (3 Steps)
**Decision**: Ship in 3 increments: hello world -> board reading -> full card display.
**Rationale**: Each step validates a key hypothesis before investing in the next. Hello world validates deployment. Board reading validates data pipeline. Full cards validates usability.

---

## Hard Constraints

| Constraint | Source | Impact |
|-----------|--------|--------|
| Zero budget | Founder (student) | Must use GCP free tier only |
| Private repo support | Founder requirement | Need deploy key or token on server |
| GitHub Free plan | Founder's current plan | 2,000 Actions min/month (less relevant now with backend approach) |
| Full card info | Founder requirement | Cannot ship a minimal summary view |
| Go language | Existing codebase | Backend must be Go to reuse domain code |
| Hexagonal architecture | Existing codebase | Web server is a new primary adapter + secondary adapter for GitHub/git |

---

## Validated Assumptions

| # | Assumption | Evidence | Confidence |
|---|-----------|----------|------------|
| A4 | Hosting can be free | GCP e2-micro free tier exists and is documented | High |
| A8 | Private repo support required | Explicitly stated in Round 2 | Confirmed |
| A6 | Mermaid is insufficient | "Not optimal" -- real dissatisfaction | Medium (founder only) |

## Invalidated Assumptions

| # | Assumption | What happened |
|---|-----------|---------------|
| -- | Azure is the preferred platform | Initially stated, then reconsidered in favor of GCP for latency reasons |
| -- | Static site is sufficient for MVP | User explicitly preferred backend approach |
| -- | Minimal card info is acceptable | User demanded "full information" |

## Unvalidated Assumptions (Carry Forward)

| # | Assumption | Why unvalidated | When to test |
|---|-----------|----------------|-------------|
| A1 | Non-technical people will want to view the board | No end users exist | After first deployment, share URL |
| A3 | Non-technical users will find the board understandable | No usability testing yet | Sprint 2 prototype testing |
| A9 | Developer-first adoption model works | No adoption data | After public release |
| A10 | GCP e2-micro has sufficient resources | No deployment spike yet | Sprint 1 (H1 hypothesis) |

---

## Gate Summary

| Gate | Status | Key Evidence |
|------|--------|-------------|
| G1: Problem Validation | CONDITIONALLY PASSED | Pre-market adapted. Mermaid friction is real. Market validates the problem class. |
| G2: Opportunity Mapping | PASSED | 5 opportunities, top 3 score >8, direction aligned. |
| G3: Solution Testing | PENDING | Hypotheses defined, spikes not yet executed. |
| G4: Market Viability | PASSED (adapted) | Lean Canvas complete, 2 GREEN + 2 YELLOW risks, GO with conditions. |

---

## Handoff Readiness

### Ready for DISCUSS Wave
The discovery package is ready for handoff to product-owner with the following artifacts:
- `problem-validation.md` -- 3 rounds of founder interviews, adapted G1
- `opportunity-tree.md` -- 5 opportunities, Go backend on GCP selected
- `solution-testing.md` -- 5 hypotheses, 3-sprint experiment plan
- `lean-canvas.md` -- Complete canvas, GO with conditions
- `interview-log.md` -- Full log of all 3 rounds with pattern analysis
- `wave-decisions.md` -- This file

### Conditions on Handoff
1. Sprint 1 stories should include deployment spike (H1) and git clone spike (H2) as validation steps.
2. Product-owner should define kill criteria: if URL gets zero non-developer views in 2 weeks post-deployment, reassess.
3. Value and usability risks (YELLOW) must be addressed during DISTILL wave through real user testing.

### Pre-Market Acknowledgment
This discovery was conducted without external users. All "validation" is founder perspective + market observation. The incremental delivery approach (hello world first) is the primary risk mitigation -- each step is cheap to build and provides real usage data to validate or invalidate assumptions.
