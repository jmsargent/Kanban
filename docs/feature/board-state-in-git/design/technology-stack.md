# Technology Stack — board-state-in-git

**Feature**: board-state-in-git
**Date**: 2026-03-18
**Status**: Approved

---

## 1. Summary

This feature introduces no new external dependencies. All implementation uses Go 1.22+ stdlib and the existing OSS libraries already present in the module. The transitions.log adapter uses `syscall.Flock`, which is part of the Go standard library on POSIX systems (Linux, macOS). No third-party locking library is required.

---

## 2. Technology Choices

### 2.1 Language and Runtime

| Component | Choice | License | Rationale |
|-----------|--------|---------|-----------|
| Language | Go 1.22+ | BSD-3-Clause | Established by ADR-003. No change. |
| CLI framework | cobra v1.x | Apache 2.0 | Established by ADR-003. New `kanban log` command follows existing pattern. |

### 2.2 File I/O — Append-Only Log

| Concern | Approach | Stdlib Component |
|---------|----------|-----------------|
| Append write | `os.OpenFile` with `O_APPEND\|O_CREATE\|O_WRONLY` | `os` package |
| Advisory locking | `syscall.Flock(fd, syscall.LOCK_EX)` | `syscall` package |
| Fsync on close | `f.Sync()` before `f.Close()` | `os.File` |
| Reverse scan for LatestStatus | Read full file, scan lines in reverse | `bufio.Scanner` |

No third-party file locking library (e.g., `github.com/gofrs/flock`) is required. The `syscall` package provides POSIX advisory locking on all supported platforms (Linux, macOS). Windows is not a supported platform for this tool (git-native, POSIX shell hook — out of scope).

### 2.3 Git Log Parsing

| Concern | Approach |
|---------|----------|
| Git history for a file | Invoke `git log --follow --format="%H|%aI|%ae|%s" -- <path>` via `os/exec` |
| Output parsing | Split on `|` (pipe), parse ISO 8601 timestamp with `time.Parse` |
| Existing pattern | Follows the existing `GitAdapter` pattern of subprocess invocation |

No new git library dependency. The existing adapter pattern uses `os/exec` to invoke the git binary.

### 2.4 Architecture Enforcement

| Tool | Version | License | Purpose |
|------|---------|---------|---------|
| go-arch-lint | existing | MIT | Enforce hexagonal dependency rules. Already active in CI. Extended config only. |

---

## 3. Rejected Alternatives

### Alternative: Third-party file locking library (`github.com/gofrs/flock`)

`gofrs/flock` provides a cross-platform file locking abstraction including Windows support. It is well-maintained (Apache 2.0, active releases).

**Rejection rationale**: Windows is not a supported platform. Adding a dependency for Windows compatibility that is not needed adds module graph surface area with no benefit. `syscall.Flock` is sufficient, already available, and has zero import cost.

### Alternative: SQLite for transitions log (`modernc.org/sqlite`)

SQLite would provide indexed queries, removing the need for linear scans in `LatestStatus`.

**Rejection rationale**: SQLite introduces a CGO or pure-Go dependency that significantly increases binary size and build complexity. A plain text log is inspectable with standard Unix tools (`grep`, `tail`, `awk`), aligns with the git-native philosophy, and performs adequately at expected scale (see performance analysis in architecture-design.md §10). SQLite would be the correct choice if the query patterns required joins or aggregations — they do not.

### Alternative: Structured log format (JSON lines)

Each entry as a JSON object: `{"ts":"2026-03-18T14:22:01Z","task":"TASK-001",...}`.

**Rejection rationale**: JSON lines are more verbose (larger file), require a JSON parser for every read, and are harder to inspect with `grep`. The space-separated format is inspectable with `grep TASK-001 .kanban/transitions.log` directly, which is consistent with the developer-experience goals of the tool. JSON offers no advantage for this use case over the fixed-field text format.

---

## 4. External Integrations

This feature has no external API integrations. No contract testing annotation required. All I/O is local filesystem and local git process.

---

## 5. Platform Compatibility

| Platform | Support | Notes |
|----------|---------|-------|
| macOS (arm64, amd64) | Full | Primary development platform |
| Linux (amd64, arm64) | Full | CI platform; goreleaser cross-compile targets |
| Windows | Not supported | `syscall.Flock` not available; POSIX hook not applicable |

The `syscall.Flock` call should be wrapped in a build-tagged file (`flock_unix.go`) to allow the binary to compile on Windows without the locking behaviour, producing a compile error or no-op with a clear message. This is a software-crafter implementation decision; the architecture requires the behaviour to be present on POSIX.
