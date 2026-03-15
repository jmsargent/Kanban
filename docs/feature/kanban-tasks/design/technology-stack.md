# Technology Stack: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DESIGN
**Date**: 2026-03-15

---

## Language and Runtime

| Decision | Choice | Rationale |
|---------|--------|-----------|
| Language | Go 1.22+ | Team preference; single static binary; sub-ms startup; strong CLI tooling |
| Runtime | None (compiled binary) | No runtime dependency on target machines or CI agents |
| Build target | Single binary (`kanban`) | `go build -o kanban ./cmd/kanban` |

See ADR-003 for full rationale and rejected alternatives.

---

## Core Dependencies

| Component | Library | License | Version | Purpose |
|-----------|---------|---------|---------|---------|
| CLI framework | `github.com/spf13/cobra` | Apache 2.0 | ^1.8 | Command/subcommand structure, flag parsing, shell completion |
| YAML parsing | `gopkg.in/yaml.v3` | MIT | v3 | `.kanban/config` parsing; YAML front matter serialisation |
| Front matter parsing | `github.com/adrg/frontmatter` | MIT | ^0.2 | Extracts YAML front matter from Markdown task files |
| Terminal color | `github.com/fatih/color` | MIT | ^1.16 | TTY color output; auto-disables on `NO_COLOR` and non-TTY |
| Testing assertions | `github.com/stretchr/testify` | MIT | ^1.9 | `assert` and `require` for unit and integration tests |

**No proprietary dependencies.** All libraries are MIT or Apache 2.0.

---

## Architecture Enforcement

| Tool | License | Version | Purpose |
|------|---------|---------|---------|
| `github.com/fe3dback/go-arch-lint` | MIT | ^1.0 | Enforces hexagonal dependency rules in CI |

Configured via `.go-arch-lint.yml` at repo root. Runs in CI before `go test`. See architecture-design.md §5 for rule specification.

---

## Development Tooling

| Tool | Purpose |
|------|---------|
| `go test ./...` | Unit and integration tests (stdlib test runner) |
| `go vet ./...` | Static analysis (stdlib) |
| `golangci-lint` | Linting (MIT; composite linter wrapping staticcheck, errcheck, etc.) |
| `go build` | Compilation; cross-compile via `GOOS`/`GOARCH` env vars |

---

## Distribution

| Channel | Mechanism |
|---------|---------|
| GitHub Releases | Pre-built binaries: macOS (amd64/arm64), Linux (amd64/arm64), Windows (amd64) |
| Homebrew (macOS) | Homebrew formula pointing to GitHub Release binary |
| `go install` | For teams with Go toolchain: `go install github.com/[org]/kanban@latest` |

---

## Dependency Rationale Notes

**cobra over `flag` stdlib**: cobra provides subcommand routing, automatic `--help` generation, and shell completion (bash/zsh/fish) out of the box. These directly satisfy NFR-5. The stdlib `flag` package does not support subcommands or completion.

**`adrg/frontmatter` for task files**: lighter than a full Markdown parser. Parses only the `---`-delimited front matter block and returns the remaining body as a string. The hook and CI step can also perform a targeted `status:` line replacement without invoking the parser at all (single-line regex), keeping the hot path fast.

**`fatih/color` over manual ANSI**: handles `NO_COLOR` environment variable, non-TTY detection, and Windows terminal compatibility in a single dependency. Directly satisfies NFR-4.

**No ORM / database library**: task files are plain files. No database abstraction is warranted. The `FileSystemAdapter` uses `os` stdlib for file operations and `gopkg.in/yaml.v3` for front matter serialisation.
