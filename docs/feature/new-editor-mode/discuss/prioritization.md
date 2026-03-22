# Prioritization: new-editor-mode

## Release Priority

| Priority | Release | Target Outcome | KPI | Rationale |
|----------|---------|---------------|-----|-----------|
| 1 | Walking Skeleton = US-01 | Developers capture tasks interactively without inline title composition | Task creation rate via editor mode | Only one story; skeleton and full feature are the same slice |

---

## Backlog

| Story | Release | MoSCoW | Priority Score | Outcome Link | Dependencies |
|-------|---------|--------|----------------|-------------|--------------|
| US-01: kanban new editor mode | WS / R1 | Must Have | 5×4/2 = 10 | KPI-1 (task capture without friction) | Existing EditFilePort, AddTask use case |

Priority score formula: Value(5) × Urgency(4) / Effort(2) = 10

- **Value = 5**: directly removes friction from the most common task creation path; the feature request is explicit
- **Urgency = 4**: user-requested, no hard deadline but clear pain felt in daily use
- **Effort = 2**: brownfield; reuses `openEditor()`, `EditFilePort`, and `AddTask.Execute` unchanged; only the CLI adapter needs a new branch

---

## Alternatives Considered

| Alternative | Why Rejected |
|---|---|
| Interactive prompts (ask for title, then priority etc. one field at a time) | Breaks flow more than the editor; does not leverage $EDITOR preference; inconsistent with `kanban edit` which uses the editor |
| New subcommand `kanban create` | Unnecessary new surface area; user expectation is that `kanban new` is the creation command |
| Template selection (multiple blank templates) | Out of scope; single blank template matches `kanban edit` pattern |

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|------------|--------|------------|
| `openEditor()` is private to usecases package and cannot be shared | Low | High | Function is already in usecases/edit_task.go; extract to package-level or move to a shared helper — design decision for solution-architect |
| WriteTemp expects a populated `domain.Task` and panics on zero value | Low | High | Acceptance test with blank task covers this; solution-architect to verify WriteTemp implementation handles zero-value task |
| EDITOR env var contains arguments (e.g. `code --wait`) and exec.Command fails | Medium | Medium | Existing `kanban edit` has this same risk; solution-architect may address by wrapping in shell invocation — carry forward |
