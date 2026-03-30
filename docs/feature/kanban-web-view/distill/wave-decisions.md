# DISTILL Decisions: kanban-web-view

## Wave Summary
**Date**: 2026-03-29
**Outcome**: Acceptance test specifications defined for all 5 stories

---

## Key Decisions

### DD-DISTILL-01: Backend subfolder for web acceptance tests
**Decision**: Web acceptance tests live in `tests/acceptance/backend/` with their own `dsl/` and `driver/` packages.
**Rationale**: User requirement. Separates web backend tests from existing CLI tests. Each test suite has its own context, drivers, and DSL tailored to its system under test.

### DD-DISTILL-02: Three-layer architecture (test / DSL / driver)
**Decision**: Follow the LMAX SimpleDSL pattern from `atdd/` wiki — tests call DSL, DSL calls drivers.
**Rationale**: User requirement. Mirrors the existing CLI acceptance test structure. DSL provides domain-language readability, drivers encapsulate protocol details (HTTP, git, HTML parsing).

### DD-DISTILL-03: Server as subprocess
**Decision**: Tests compile and start `kanban-web` as a subprocess, same as CLI tests use the `kanban` binary.
**Rationale**: True end-to-end testing — tests exercise the real binary, not in-process handlers. Health check `/healthz` gates test execution.

### DD-DISTILL-04: GitHub API stub via httptest
**Decision**: A local `httptest.Server` stubs GitHub API (`GET /user`) for token validation. Server URL injected via env var.
**Rationale**: Tests must not hit real GitHub. The stub is per-test configurable (accept/reject specific tokens). Follows the existing pattern of testing against real infrastructure (git repos) with stubs only for external APIs.

### DD-DISTILL-05: SimpleDSL string parameter convention
**Decision**: DSL methods use positional required params + named optional `"key: value"` string params.
**Rationale**: Matches the SimpleDSL pattern from `atdd/` wiki and the existing CLI test style (e.g., `dsl.Task("title", "status: in-progress", "assignee: alice")`).

### DD-DISTILL-07: Flat user action functions, no method chains
**Decision**: User actions are standalone DSL functions (`UserVisitsBoard`, `UserAddsTask`, etc.) that accept `"name: <value>"` as a named param. There are no chained methods (`User(name).Action()`).
**Rationale**: Consistent with the `"key: value"` convention used everywhere else in the DSL. Flat functions are simpler to implement, easier to read in a Given/When/Then sequence, and avoid introducing a builder/fluent pattern that differs from all other DSL methods.

User state is set up in `Given` via `User(name)` / `User(name, "auth: valid")` / `User(name, "auth: invalid")`. Actions in `When` reference that user by `"name: <value>"`. This replaces the former `AFirstTimeVisitor()`, `AnAuthenticatedUser(name)`, and `AnUnauthorizedUser()` helpers.

### DD-DISTILL-06: Real git repos for all tests
**Decision**: Every test creates a real temp git repo via `t.TempDir()` + `git init`. Push tests use a bare remote.
**Rationale**: Follows existing CLI test strategy. No mocking of git — tests verify real file I/O and git operations.

---

## Test Count by Story

| Story | Tests | Coverage Focus |
|-------|-------|----------------|
| US-WV-01 | 2 | Server starts, responds, acceptable latency |
| US-WV-02 | 6 | Columns, cards, sorting, sync, empty board, no auth needed |
| US-WV-03 | 4 | Full details, missing fields, task ID, htmx fragment |
| US-WV-04 | 5 | Token prompt, valid/invalid token, persistence, view-only |
| US-WV-05 | 6 | All fields, title only, validation, format, push, unauth |
| **Total** | **23** | |
