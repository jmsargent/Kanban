# ADR-001: Hexagonal Architecture (Ports and Adapters)

**Status**: Accepted
**Date**: 2026-03-15
**Feature**: kanban-tasks

---

## Context

The kanban CLI must support three distinct execution environments:

1. Interactive terminal (developer running `kanban add`, `kanban board`, etc.)
2. Git commit-msg hook (automated, fast, must exit 0, no TTY)
3. CI/CD pipeline step (non-TTY, no interactive prompts, must commit file changes back)

Each environment has different I/O characteristics but shares identical domain logic: task status transitions, ID generation, business rule validation (past due dates, directional transitions, unique IDs). Without a clear boundary, domain logic risks being entangled with filesystem and git operations, making it untestable in isolation and brittle when execution contexts change.

Additionally, the task file format (OD-01) and the hook delivery mechanism (OD-05) were open decisions at DISCUSS time. An architecture that hardwires these choices into the domain core would require invasive changes if those choices change.

Quality attributes driving this decision:
- **Testability** (highest priority): domain logic must be testable without git, filesystem, or TTY
- **Maintainability**: adapters for hook, CI step, and file format must be swappable independently
- **Portability**: the same domain core runs in all three execution contexts

---

## Decision

Apply Hexagonal Architecture (Ports and Adapters) as the structural style for the kanban CLI codebase.

The domain core contains pure business logic with no I/O dependencies. All external interactions (filesystem reads/writes, git operations, terminal I/O, CI environment reads) cross the hexagon boundary exclusively through typed port interfaces. Concrete adapters implement those ports and are wired at the entry point (CLI bootstrap / hook entry point / CI step entry point).

---

## Alternatives Considered

### Alternative 1: Layered Architecture (N-Tier)

Three horizontal layers: CLI -> Service -> Repository. Standard pattern, fast to understand.

Rejection rationale: In a layered architecture, the service layer typically imports from the repository layer directly. This means domain logic (task transition rules, validation) is coupled to the file-system representation at test time. Mocking file I/O for unit tests is fragile and platform-dependent. The three execution contexts (interactive, hook, CI) would share a service layer but diverge in the I/O layer with no enforced boundary -- high risk of business logic leaking into CLI handlers over time.

### Alternative 2: Vertical Slice per Command

Each command (`add`, `board`, `edit`, `delete`) is a self-contained slice owning its own domain logic, validation, and file access.

Rejection rationale: Business rules (BR-1 through BR-7) and the task entity model are shared across all commands. Status transitions, ID uniqueness, and the task file schema are cross-cutting. Vertical slices would duplicate this logic or force a shared "common" module that effectively becomes the domain core anyway, without the discipline of enforced dependency direction. Vertical slices are better suited to feature-heavy apps where features are genuinely independent.

---

## Consequences

**Positive**:
- Domain core is testable in isolation: no filesystem, no git, no process spawning required
- Execution contexts (interactive, hook, CI) share identical domain logic via different primary adapter wiring
- Adapter boundaries are explicit: swapping the file format or adding a new storage backend requires only a new secondary adapter implementation
- Architecture enforcement via dependency-cruiser is mechanical: domain module must have zero imports from adapter modules

**Negative**:
- More initial structure than a simple layered approach: requires port interfaces to be defined before adapters
- For a small greenfield project, the adapter boundary adds boilerplate that only pays back when the adapters need to be swapped or when test isolation becomes important (which it will, from day one, for the hook and CI contexts)

**Accepted trade-off**: The boilerplate cost is justified by the three distinct execution contexts and the testability requirement on a tool where hook correctness is safety-critical (a broken hook must never block git).
