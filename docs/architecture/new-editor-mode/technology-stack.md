# Technology Stack — new-editor-mode

**Wave**: DESIGN
**Date**: 2026-03-22

---

## Stack Overview

This feature is a brownfield addition to an existing Go CLI. No new technologies are introduced. All stack decisions were made in prior ADRs and remain unchanged.

---

## Existing Stack (Unchanged)

| Component | Technology | Version | License | ADR |
|-----------|-----------|---------|---------|-----|
| Language | Go | 1.22+ | BSD 3-Clause | ADR-003 |
| CLI framework | cobra | v1.x | Apache 2.0 | ADR-003 |
| YAML parsing | gopkg.in/yaml.v3 | v3.x | MIT | ADR-002 |
| Architecture style | Hexagonal (ports and adapters) | — | — | ADR-001 |
| Task file format | Markdown + YAML front matter | — | — | ADR-002 |
| Architecture enforcement | go-arch-lint | current | MIT | CLAUDE.md |
| CI/CD | CircleCI + goreleaser v2 | current | OSS | ADR-005 |

---

## New Introductions

None. The editor-mode feature uses:

- `os.CreateTemp` — Go stdlib
- `os/exec.Command` — Go stdlib
- `os.Remove` — Go stdlib
- `yaml.Marshal` — already a project dependency (gopkg.in/yaml.v3)

No new third-party dependencies are added. No new build tooling. No new runtime dependencies.

---

## Editor Subprocess

The `$EDITOR` environment variable is resolved at runtime. The feature makes no assumption about which editor is used. The fallback is `vi` (matches existing `kanban edit` behaviour). This is a runtime dependency on the developer's environment, not a build-time or package dependency.

Supported editors: any process that accepts a file path as its first argument, blocks until the user saves and quits, and exits 0 on success. This covers vim, nano, emacs, `code --wait`, `subl --wait`, etc.

---

## Architecture Enforcement Tooling

Tool: `go-arch-lint` (already active)
Configuration: existing YAML config in project root
New rules required: none — the new file `internal/usecases/editor.go` falls within the existing `usecases` package rules.
