# Platform Architecture — kanban-tasks

**Wave**: DEVOPS
**Date**: 2026-03-16
**Feature**: kanban-tasks (Git-Native Kanban Task Management CLI)

---

## Distribution Model

kanban-tasks is a developer CLI tool distributed as a static binary. There is no server infrastructure, no container orchestration, and no service mesh. The distribution architecture is entirely build-and-release.

```
Developer workstation
        |
        | git push to main
        v
   GitHub (source host)
        |
        | webhook trigger
        v
   CircleCI (build / test / release)
        |
        |--- CI workflow: build + test + acceptance
        |
        +--- CD workflow (on CI pass): goreleaser
                |
                |--- GitHub Releases (cross-compiled binaries)
                |--- go install (module proxy via GitHub)
                +--- Homebrew tap (homebrew-kanban repo)
```

## Component Roles

| Component | Role | Responsibility |
|-----------|------|----------------|
| GitHub | Source host | Code, Issues, Releases, Homebrew tap repo |
| CircleCI | Build / Test / Release | CI pipeline, CD trigger, artifact promotion |
| goreleaser | Release automation | Cross-compilation, archive packaging, GitHub Release creation, Homebrew formula update |
| GitHub Releases | Binary distribution | Download endpoint for all 6 platform targets |
| `go install` | Go-native distribution | Module proxy access; no binary hosting needed |
| homebrew-kanban repo | Homebrew tap | Formula file updated by goreleaser on every release |

## Artifact Lifecycle

The central principle is **promote, not rebuild**: the binary produced in CI is the exact artifact shipped to users. goreleaser is invoked in CD with the git tag created by CircleCI, and it rebuilds for all targets from the same commit that CI validated — but the Go toolchain is deterministic, so the same source + same toolchain version = identical binary. No separate artifact storage (e.g., S3) is required.

```
git push to main
    |
    v
CircleCI CI workflow
    1. go-arch-lint check        (architecture gate)
    2. go vet ./...              (static analysis)
    3. golangci-lint run         (composite linter)
    4. go test ./...             (unit + integration)
    5. go build -o kanban        (compile binary)
    6. acceptance tests          (E2E, binary subprocess)
    7. persist workspace         (binary + git metadata)
    |
    | all steps pass
    v
CircleCI CD workflow
    1. attach workspace
    2. goreleaser release        (reads cicd/goreleaser.yml)
           |
           |-- cross-compile 6 targets
           |-- create GitHub Release with auto-changelog
           |-- upload archives + checksums
           +-- push Homebrew formula to homebrew-kanban repo
```

## Security Model

- All secrets (GITHUB_TOKEN, HOMEBREW_TAP_GITHUB_TOKEN) are stored in CircleCI project contexts, never in source code.
- goreleaser authenticates to GitHub via GITHUB_TOKEN (scoped to releases and repo contents).
- goreleaser authenticates to the Homebrew tap repo via a separate HOMEBREW_TAP_GITHUB_TOKEN (scoped to the homebrew-kanban repo only).
- No production secrets exist: this is a CLI tool with no runtime service.

## No Server Infrastructure

This section exists to document explicit deferrals:

- **Container orchestration**: none — binary distribution requires no containers.
- **Cloud infrastructure (IaC)**: none — no server to provision.
- **Service observability (metrics/tracing/alerting)**: deferred — see `observability-design.md`.
- **Database**: none — task state lives in `.kanban/` within each user's git repository.

## Homebrew Tap Setup

A new GitHub repository `homebrew-kanban` must be created under the same GitHub organization/user as the main repo. goreleaser is authorized to push formula updates to this repo via HOMEBREW_TAP_GITHUB_TOKEN.

Installation command for end users once tap is published:
```sh
brew tap <org>/kanban
brew install kanban
```

## go install Support

Because the module is a standard Go module with a `cmd/kanban` entry point, `go install` works automatically:
```sh
go install github.com/<org>/kanban-tasks/cmd/kanban@latest
```

No additional infrastructure is needed to support this distribution channel.
