# Lean Canvas: kanban-web-view

## Feature ID
kanban-web-view

---

## 1. Problem
1. **Non-technical team members cannot see the kanban board** -- CLI-only access creates a developer dependency bottleneck.
2. **Mermaid diagram in README is insufficient** -- lacks detail, interactivity, and requires GitHub repo navigation.
3. **Private repos block casual access** -- non-technical viewers would need GitHub accounts and repo access.

Existing alternatives: Mermaid diagram in README | screen-sharing by developers | switching to Trello/Jira (abandoning git-native approach).

## 2. Customer Segments
**Primary**: Developers using kanban CLI who need to share board visibility with their team.
**Secondary**: Non-technical team members (POs, stakeholders) who consume board status.

Job-to-be-done: "When I want to understand project status, help me see what tasks exist, their current state, and who is working on them, so I can make decisions without interrupting a developer."

Early adopter: Solo developer or small team already using kanban CLI, wanting to share progress with a non-technical collaborator.

## 3. Unique Value Proposition
**See your git-native kanban board in a browser -- no CLI, no GitHub account, just a URL.**

The kanban board stays in git (single source of truth) while becoming accessible to the entire team through a simple web view.

## 4. Solution
1. **Go HTTP server on GCP Compute Engine** -- maintains local git clone, reads `.kanban/tasks/` from filesystem, renders board as HTML.
2. **Full card details** -- title, status, assignee, dates, description visible on every card.
3. **Works with private repos** -- server holds deploy key, viewers only need a URL.
4. **Always-on with low latency** -- e2-micro VM, no cold starts, periodic git pull for near-real-time sync.

## 5. Channels
- **Primary**: Developer sets up the web view, shares URL with team (developer-first adoption).
- **Secondary**: GitHub repo README includes link to live board.
- **Future**: kanban CLI `init` command could offer web view setup as optional step.

## 6. Revenue Streams
**N/A for this feature.** kanban is an open-source CLI tool. The web view is a free feature that increases adoption by making the tool useful for entire teams, not just developers.

If monetized in future: hosted version (SaaS) where teams don't need to deploy their own server.

## 7. Cost Structure
| Cost | Amount | Notes |
|------|--------|-------|
| GCP Compute Engine | $0 | e2-micro free tier (1 instance, us-central1 or similar) |
| Network egress | $0 | 1 GB/month free (sufficient for small team) |
| GitHub API / git operations | $0 | Local clone, no API costs |
| Development effort | ~3 sprints | Solo developer time |
| Domain (optional) | $0-12/yr | Custom domain optional, can use IP or free subdomain |

Total recurring cost: $0 (free tier).

## 8. Key Metrics
| Metric | What it measures | Target |
|--------|-----------------|--------|
| Board page views | Are non-technical members actually looking? | >1 view/week per team member |
| Unique viewers | How many people beyond the developer use it? | >1 non-developer viewer |
| Return visits | Do people come back without being asked? | >50% return within 2 weeks |
| Time to deploy | Is setup friction low enough? | <30 minutes first-time setup |
| Sync latency | How fresh is the board data? | <60 seconds after push |

## 9. Unfair Advantage
- **Direct reuse of kanban domain code**: the Go backend imports the same packages as the CLI -- zero translation layer, guaranteed consistency.
- **Git as single source of truth**: no sync, no database. The repo IS the database. Local clone means filesystem-speed reads.
- **Zero cost**: GCP free tier e2-micro + local git clone = $0 hosting for small teams.
- **Always-on**: Unlike serverless, no cold start penalty. Board loads fast every time.

---

## 4 Big Risks Assessment

| Risk | Status | Evidence | Mitigation |
|------|--------|----------|------------|
| **Value** | YELLOW | Founder confirms need. No end-user validation yet. Market evidence strong (CLI tools routinely need web UIs). | Incremental delivery: hello world first. Kill if zero adoption after sharing URL. |
| **Usability** | YELLOW | Full card info requested. Layout/UX untested. | Usability test with 2-3 people after working prototype (H4). |
| **Feasibility** | GREEN | Go backend is well-understood. GCP e2-micro free tier exists. Local git clone is trivial. Domain code is portable (hexagonal arch). | Spike deployment (H1) and git clone sync (H2) before full build. |
| **Viability** | GREEN | $0 cost. Open-source project. No revenue dependency. Feature increases adoption surface. | No financial risk. Only risk is wasted development time if nobody uses it. |

---

## Go / No-Go Recommendation

**GO -- with conditions.**

Proceed to DISCUSS wave handoff with the following conditions:
1. **Spike first**: H1 (GCP deployment) and H2 (local git clone + pull) must be validated before any stories beyond hello-world are committed to.
2. **Usability test required**: After working board prototype, test with at least 1 non-technical person before declaring feature complete.
3. **Kill criteria**: If after deploying and sharing the URL, zero non-developer views occur within 2 weeks, reassess the feature's value.

## Phase 4 Status
- **Lean Canvas**: Complete
- **4 Big Risks**: 2 GREEN (feasibility, viability), 2 YELLOW (value, usability)
- **Go/No-Go**: GO with conditions
- **Unresolved**: Value and usability risks remain yellow until real users interact with the product. This is accepted given pre-market status.
