# Technology Stack — task-creator-attribution

## No new dependencies introduced.

This feature is implemented entirely within the existing technology stack.
No new libraries, frameworks, or external services are required.

---

## Existing Stack (unchanged)

| Component | Technology | Version | Rationale (from ADR-003) |
|-----------|-----------|---------|--------------------------|
| Language | Go | 1.22+ | Static typing, single-binary compilation, strong stdlib |
| CLI framework | cobra | v1.x | Apache 2.0, idiomatic Go CLI, cobra.Command pattern |
| YAML serialization | gopkg.in/yaml.v3 | v3 | Already used for task front matter |
| Git operations | os/exec subprocess | stdlib | No go-git dependency; shell git is always available |
| Atomic file writes | os.Rename | stdlib | Cross-platform atomicity |

---

## Identity Resolution: Implementation Approach

`GetIdentity()` in `GitAdapter` runs:

```
git config user.name
git config user.email
```

as two separate `exec.Command` calls using the existing `runGitIn` pattern.

**Alternative evaluated**: Parse `~/.gitconfig` directly with a YAML/INI parser.
**Rejected**: `git config` handles all config sources (system, global, local, env overrides)
correctly. Parsing the file directly would miss local repo config, env var overrides
(`GIT_AUTHOR_NAME`), and conditional includes. The subprocess is the canonical approach.

**Alternative evaluated**: Use a Go git library (go-git) for config reading.
**Rejected**: go-git is a heavyweight dependency for a one-line operation. The existing
`os/exec` pattern is established and consistent. ADR-003 explicitly prefers stdlib + shell git.

---

## Mock Strategy for Tests

The existing port mock pattern is used:

- Unit/use case tests: implement `GitPort` as an in-memory struct; `GetIdentity()` returns a
  hardcoded `Identity{Name: "Test User"}` or a configured error
- Integration (git adapter) tests: use `t.TempDir()` + `git init` + `git config user.name "..."`
  in the test setup, then call `GetIdentity()` on a real `GitAdapter`
- Acceptance (end-to-end) tests: the compiled binary subprocess reads from the real git config
  of the test repo created with `git config user.name` in the test setup

No new mocking libraries required.
