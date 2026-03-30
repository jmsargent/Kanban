# Interview Log: kanban-web-view

## Feature ID
kanban-web-view

## Context
All interviews conducted with a single founder/developer (pre-market product, no external users exist). This is an adapted discovery process -- standard Mom Test requires 5+ distinct interviewees.

---

## Round 1: Initial Problem Exploration
**Date**: 2026-03-29
**Participant**: Founder (sole developer)
**Focus**: Problem validation -- is the non-technical access problem real?

### Questions and Responses

| # | Question | Response |
|---|----------|----------|
| 1 | When was the last time a PO or stakeholder wanted to add a task to your board? | No non-technical users yet. "The app MUST appeal to them." |
| 2 | How often do non-technical people need to see project status on your teams? | "Very common for non-technical people to be part of a development team." |
| 3 | What workarounds do non-technical people use today to interact with your board? | "They would have to go through a developer somehow." |
| 4 | Who specifically would use this? Can you name people? | POs and stakeholders (no specific individuals named). |
| 5 | What can non-technical people see right now? | "They can see a Mermaid diagram." |

### Evidence Quality
- Mostly future intent and market observation, not past behavior
- Strong founder conviction but no external validation
- One concrete artifact: the Mermaid diagram exists as current state

### Key Insights
- Problem is assumed, not validated with end users
- Mermaid diagram is the existing (insufficient) solution
- Founder sees this as a critical adoption enabler

---

## Round 2: Constraints and Adoption Model
**Date**: 2026-03-29
**Participant**: Founder (sole developer)
**Focus**: Hard constraints, current behavior, adoption model

### Questions and Responses

| # | Question | Response |
|---|----------|----------|
| 1 | Where does the Mermaid diagram live and how do people access it? | "Auto-generated in README. Non-technical people have to interact with the repo, which is okay but not optimal." |
| 2 | What is your hosting budget? | "Free. Student with access to Azure and GCP free tiers." |
| 3 | Do you work with public repos, private repos, or both? | "Both public and private." |
| 4 | Who would you show this tool to first? | "Developers first, but for them to present to a team, the entire team has to be able to use it." |
| 5 | When you say incremental delivery, what does that mean to you? | "Either path works." |

### Evidence Quality
- Mermaid-in-README is real current behavior (strongest evidence so far)
- "Not optimal" is a real dissatisfaction signal from someone actually using the current state
- Hard constraints (budget, repo types) are factual, not speculative

### Key Insights
- Zero budget is a hard constraint (student)
- Developer-first adoption then team expansion is the go-to-market model
- Private repo support is non-negotiable
- Current Mermaid solution exists but is explicitly described as suboptimal

---

## Round 3: Architecture and Platform Decisions
**Date**: 2026-03-29
**Participant**: Founder (sole developer)
**Focus**: Technical preferences, card content, access control, architecture choice

### Questions and Responses

| # | Question | Response |
|---|----------|----------|
| 1 | What GitHub plan are you on? (Affects Actions minutes) | "I'm on GitHub free plan." (2,000 min/month) |
| 2 | What is your experience with cloud platforms? | Initially: "I have Azure experience since before, and this is the path I think I would be most comfortable with going." Then reconsidered: GCP Compute Engine preferred for always-on VM with lowest sync latency. |
| 3 | What information must be visible on each card? | "Full information of the cards." (title, status, assignee, dates, description) |
| 4 | Do you need access control for the web view? | "Authentication can come later." (public URL acceptable initially) |
| 5 | Static site generation or backend server? | "I think a backend would be the preferable approach." |

### Evidence Quality
- Platform preference shifted from Azure to GCP with clear technical rationale (latency concern = real reasoning, not just comfort)
- "Full information" is a clear requirement -- not a vague wish
- Auth deferral is an explicit scope decision
- Backend preference overrides our static-first recommendation -- user has their own reasoning

### Key Insights
- GCP Compute Engine e2-micro chosen: always-on VM avoids cold starts, enables local git clone for fastest sync
- Full card info is mandatory -- rules out summary/minimal views
- Auth deferred = reduced scope for MVP
- User has clear opinions on architecture -- not just accepting recommendations

---

## Cross-Round Pattern Analysis

### Consistent Signals (across all 3 rounds)
1. **Non-technical access is a real concern**: mentioned in every round, tied to adoption model
2. **Zero budget**: confirmed and reconfirmed -- student, free tier only
3. **Private repo support**: mentioned in Round 2, architecture addresses it in Round 3
4. **Mermaid is insufficient**: "not optimal" (Round 2), "full information" requirement (Round 3)

### Evolving Signals
1. **Platform**: Azure (Round 2) -> GCP (Round 3) -- shifted based on latency reasoning
2. **Architecture**: Open (Round 1) -> Backend preference (Round 3) -- crystallized through discussion
3. **Scope**: Vague "web view" (Round 1) -> Full card detail, auth deferred, Go backend (Round 3)

### Absent Signals (gaps)
1. **No end-user voice**: All evidence is founder perspective. Zero non-technical person interviews.
2. **No commitment signals**: Nobody has said "I would use this" or "show me when it's ready" -- because no external users exist.
3. **No competing product comparison**: Founder has not tried sharing a Trello/Linear board to compare the experience.
4. **No frequency data**: How often would non-technical people actually check the board? Unknown.

### Evidence Strength Summary
| Evidence Type | Count | Quality |
|--------------|-------|---------|
| Past behavior (real) | 2 | Mermaid diagram exists, user has Azure/GCP experience |
| Hard constraints | 3 | Zero budget, private repos, GitHub free plan |
| Stated preferences | 4 | Backend, GCP, full cards, auth deferred |
| Future intent (unvalidated) | 3 | Non-technical adoption, team tool, write access later |
| Market observation | 1 | Non-technical people commonly on dev teams |
