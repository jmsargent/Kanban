# DISCUSS Decisions: kanban-web-view

## Wave Summary
**Duration**: 2026-03-29
**Rounds**: 2 rounds of interactive Q&A with founder
**Outcome**: 5 stories defined, DoR passed, ready for DESIGN wave handoff

---

## Key Decisions

### D1: MVP includes both viewing AND task creation
**Decision**: The web view is not read-only. Non-technical users can add tasks via the web.
**Rationale**: Founder stated "They want to add cards to the board" as a core need alongside viewing. Read-only was the DISCOVER assumption; DISCUSS revealed that adding tasks is equally important.
**Impact**: Server must support write operations (commit + push). Adds US-WV-04 and US-WV-05 to scope.

### D2: Auth model — public view + token for writes
**Decision**: Public repos can be viewed without authentication. Adding tasks requires a GitHub Personal Access Token.
**Rationale**: Simplest incremental approach. Avoids OAuth complexity for MVP. Token grants both identity (via display name) and authorization (repo write access).
**Deferred**: GitHub OAuth (US-WV-06) replaces token entry later.

### D3: Token storage decision postponed
**Decision**: Whether the server stores the token server-side (session) or the client stores it (localStorage) is explicitly deferred.
**Rationale**: Founder chose to postpone. Both approaches work; the choice has security and UX tradeoffs that can be resolved during DESIGN wave.

### D4: Card fields match CLI output
**Decision**: Web cards display: title, description, status, priority, assignee, created_by. Due date shown if present.
**Rationale**: Founder requested "full information of the cards" and "same fields as the CLI."
**Note**: Started time and finished time were requested but are NOT in the current domain model (`internal/domain/task.go`). Decision: show only what the task format already has. Timestamps can be added later.

### D5: Board layout — three columns, sorted by date
**Decision**: Todo | Doing | Done columns, left to right. Cards sorted by creation date within columns (oldest first).
**Rationale**: Classic kanban layout. Matches CLI `kanban board` output semantics.
**Deferred**: Pagination/scroll for 50+ tasks.

### D6: Description rendered as plain text
**Decision**: Task descriptions shown as plain text for MVP.
**Rationale**: Founder said "in the beginning regular text, later on markdown sounds interesting."
**Deferred**: Markdown rendering.

### D7: Five stories across two releases
**Decision**: Walking skeleton (US-WV-01) → Release 1: view board + card details (US-WV-02, 03) → Release 2: token entry + add task (US-WV-04, 05).
**Rationale**: Each release is independently valuable. Release 1 validates the core value prop (view without CLI). Release 2 adds write capability.

---

## Constraints Carried from DISCOVER

| Constraint | Source |
|-----------|--------|
| Zero budget (GCP free tier only) | Founder (student) |
| Go language | Existing codebase |
| Hexagonal architecture | Existing codebase |
| Public repos first | Auth model decision |
| Task file format (Markdown + YAML front matter) | Existing CLI |

---

## Handoff Readiness

### Ready for DESIGN Wave
- [x] User stories defined with BDD acceptance criteria
- [x] Journey maps with emotional arcs
- [x] Effort estimates (5-7 days total)
- [x] Dependencies mapped
- [x] Auth model defined
- [x] Card fields confirmed against domain model

### Conditions for DESIGN
1. Resolve token storage decision (D3) — server-side session vs client-side localStorage
2. Define the hexagonal architecture integration — web server as new primary adapter, new secondary adapter for git push operations
3. Choose HTML templating approach (Go templates, or serve a static frontend)

### Kill Criteria (from DISCOVER)
- Zero non-developer views within 2 weeks of deployment → reassess feature value
