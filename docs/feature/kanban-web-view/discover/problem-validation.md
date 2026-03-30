# Problem Validation: kanban-web-view

## Feature ID
kanban-web-view

## Problem Statement
Non-technical team members (POs, stakeholders) cannot interact with the kanban board without going through a developer. The CLI-only interface creates a dependency bottleneck that excludes key contributors from the workflow.

## Discovery Context
- **Product status**: Pre-release MVP. No external users exist yet.
- **Validation approach**: Assumption-driven founder interviews (3 rounds). No real end-users available.
- **Evidence type**: Founder domain knowledge + market observation + stated platform preferences.
- **Limitation**: All evidence comes from a single founder. Standard Mom Test thresholds (5+ interviews, >60% confirmation from distinct users) cannot be met pre-market.

---

## Interview Round 1: Initial Problem Exploration

### Key Findings

| Question | Response | Evidence Type |
|----------|----------|---------------|
| Last time a PO wanted to add a task? | No non-technical users yet. "The app MUST appeal to them." | Future intent (assumption) |
| How often does this happen? | "Very common for non-technical people to be part of a development team." | Market observation |
| What workarounds exist? | "They would have to go through a developer somehow." | Hypothetical workaround |
| Who specifically? | POs and stakeholders (no specific individuals named) | Generic segment |
| Current visibility? | "They can see a Mermaid diagram." | Actual current state |

### Signals
- Past behavior evidence: None (pre-market)
- Commitment signals: Strong founder conviction
- Emotional intensity: Present in founder, unvalidated with end users
- Workaround spending: $0

---

## Interview Round 2: Constraints and Adoption Model

### Key Findings

| Question | Response | Evidence Type |
|----------|----------|---------------|
| Where does the Mermaid diagram live? | "Auto-generated in README. Non-technical people have to interact with the repo, which is okay but not optimal." | Current behavior (real) |
| Hosting budget? | "Free. Student with access to Azure and GCP free tiers." | Hard constraint |
| Public or private repos? | "Both public and private." | Hard constraint |
| Who would you show this to first? | "Developers first, but for them to present to a team, the entire team has to be able to use it." | Adoption model insight |
| What does incremental mean? | "Either path works." | Low preference signal |

### Signals
- Mermaid-in-README is the real current state -- confirms friction exists even pre-market
- Adoption model: developer-first, then team expansion
- Hard constraints: zero budget, private repo support required
- Founder describes current approach as "not optimal" -- real dissatisfaction

---

## Interview Round 3: Architecture and Platform Preferences

### Key Findings

| Question | Response | Evidence Type |
|----------|----------|---------------|
| GitHub Actions availability? | "I'm on GitHub free plan." | Hard constraint (2,000 Actions min/month) |
| Cloud platform preference? | Initially preferred Azure (past experience), then reconsidered: GCP Compute Engine preferred for always-on VM with lowest sync latency. | Stated preference with technical rationale |
| Minimum info on cards? | "Full information of the cards." | Requirement (title, status, assignee, dates, description) |
| Access control needs? | "Authentication can come later." | Explicit deferral -- public URL acceptable for now |
| Static site vs backend? | "I think a backend would be the preferable approach." | Stated architecture preference |

### Signals
- **GCP Compute Engine preferred**: user initially leaned Azure (past experience) but switched to GCP after reasoning about latency -- an always-on VM avoids cold starts that serverless/PaaS would impose. This is a deliberate technical decision, not just comfort.
- **Full card info requirement**: rules out minimal/summary views. The web board must show everything the CLI shows.
- **Auth explicitly deferred**: security through obscurity acceptable initially. Reduces scope.
- **Backend preference overrides our static-first recommendation**: user wants a Go backend, not a static site generator. This shifts the architecture.

---

## Assumptions Tracker (Final)

| # | Assumption | Category | Risk Score | Priority | Status |
|---|-----------|----------|------------|----------|--------|
| A1 | Non-technical team members will want to view the board | Value | 16 | Test first | Unvalidated (pre-market) |
| A2 | A read-only web view is sufficient as MVP | Value | 11 | Test soon | Supported by auth deferral |
| A3 | Non-technical users will find the board understandable | Usability | 14 | Test first | Unvalidated |
| A4 | Hosting on GCP Compute Engine free tier is viable | Feasibility | 6 | Test later | Supported (e2-micro free tier exists) |
| A5 | GitHub API provides sufficient access to task data | Feasibility | 11 | Test soon | Needs spike |
| A6 | Mermaid diagram is insufficient for non-technical users | Value | 14 | Test first | Partially validated (founder: "not optimal") |
| A7 | Non-technical users will eventually need write access | Value | 8 | Test soon | Deferred |
| A8 | Private repo support is required | Feasibility | N/A | Confirmed | Hard constraint |
| A9 | Developer-first adoption model works | Value | 8 | Test soon | Aligns with CLI-first product |
| A10 | Go backend on GCP e2-micro can serve board with low latency (always-on, no cold start) | Feasibility | 11 | Test soon | Needs spike |
| A11 | Full card info can be displayed clearly on a web board | Usability | 8 | Test soon | Needs prototype |

---

## Gate G1 Evaluation

| Criteria | Status | Notes |
|----------|--------|-------|
| 5+ interviews | ADAPTED (3 rounds, 1 founder) | Pre-market: no end-users available |
| >60% confirm pain | PARTIAL | Founder confirms strongly. Mermaid-in-README friction is real. |
| Problem in customer words | PARTIAL | "Not optimal" -- founder words, not end-user words |
| 3+ concrete examples | PARTIAL | Mermaid diagram exists. No end-user friction stories. |

**Gate G1 Status: CONDITIONALLY PASSED (pre-market adapted)**

Rationale: Standard gate criteria require 5+ distinct interviews with end-users. This is impossible pre-market. However:
1. The underlying problem (CLI-only tools exclude non-technical team members) is well-established in the market (Trello, Linear, Jira all exist because of this).
2. Real current-state friction exists (Mermaid diagram in README is "not optimal").
3. The feature is positioned as the bridge from "developer tool" to "team tool" -- this is a validated go-to-market pattern.
4. Risk is mitigated by incremental delivery (hello world first, then board reading).

Proceeding with explicit acknowledgment that value assumptions (A1, A3, A6) remain unvalidated until real users interact with the product.
