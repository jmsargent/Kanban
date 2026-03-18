# ADR-010: goreleaser Installation via go install with CircleCI Cache

**Status**: Accepted
**Date**: 2026-03-18
**Feature**: testable-pipeline
**Resolves**: goreleaser CI installation pattern (US-TP-01)

---

## Context

The current CI `release` job installs goreleaser using `curl -sfL https://goreleaser.com/static/run | VERSION=... bash`. This is the goreleaser project's recommended "install and run" script. It has two problems in this project:

1. **No caching**: goreleaser is downloaded on every CI run, adding latency and making the release job dependent on goreleaser's CDN being available
2. **No local equivalent**: the developer cannot `make release-snapshot` locally without a separate installation step; the `curl | bash` pattern is a CI-only invocation, not reproducible locally in the same form

The other tools in the pipeline (golangci-lint, go-arch-lint, go-semver-release) are installed via `go install` with a `restore_cache / if-guard / save_cache` pattern, making them:
- Cached between CI runs (fast)
- Installable locally with the same command
- Version-pinned to the pipeline parameter

---

## Decision

Install goreleaser using `go install github.com/goreleaser/goreleaser/v2@<version>` and cache the binary using the existing `restore_cache / if [ ! -f ] / save_cache` pattern.

Add an `install-goreleaser` reusable command to `cicd/config.yml` following the exact same structure as `install-golangci-lint` and `install-go-arch-lint`:

```
restore_cache: goreleaser-<< pipeline.parameters.goreleaser-version >>
run: if [ ! -f "$(go env GOPATH)/bin/goreleaser" ]; then
       go install github.com/goreleaser/goreleaser/v2@<version>
     fi
save_cache: goreleaser-<< pipeline.parameters.goreleaser-version >>
paths: /home/circleci/go/bin/goreleaser
```

The cache key is version-pinned. A goreleaser version bump in `cicd/config.yml` invalidates the cache and triggers a fresh installation on the next CI run.

The `make release-snapshot` Makefile target runs goreleaser locally using the same binary installed via `go install`.

---

## Alternatives Considered

### Alternative 1: Retain curl | bash, add separate local install instructions

Keep the CI installation unchanged. Document `go install github.com/goreleaser/goreleaser/v2@<version>` as the local install step in the README.

Rejection rationale:
- Does not solve the caching problem — goreleaser is still downloaded on every CI run
- Two installation methods: `curl | bash` in CI, `go install` locally. Two places to update when goreleaser version changes. This creates the same synchronisation problem that ADR-008 (Makefile) addresses for command sequences.
- `check-versions.sh` reads the goreleaser version from `cicd/config.yml` and checks the locally installed binary. If the local binary was installed a different way, version mismatches are harder to diagnose.

### Alternative 2: Docker image with goreleaser pre-installed

Use a custom CircleCI Docker image that includes goreleaser at the pinned version.

Rejection rationale:
- Requires maintaining a custom Docker image — operational overhead for a single-developer project
- Image rebuild required on every goreleaser version bump
- No benefit over the `go install` + cache pattern, which achieves the same result (one-time install, subsequent cache hits) with zero image maintenance overhead

### Alternative 3: Keep curl | bash with CircleCI cache on the goreleaser binary

Cache the goreleaser binary installed by `curl | bash` using the same restore/save pattern.

Rejection rationale:
- The `curl | bash` script installs to a path that may differ between runs; the `go install` path is deterministic (`$(go env GOPATH)/bin/goreleaser`)
- The `curl | bash` script is not reproducible locally in the same way as `go install`
- `go install` is already the established pattern for this project's tools; consistency is a maintainability benefit

---

## Consequences

**Positive**:
- goreleaser cached after first run — subsequent CI runs restore from cache in ~1 second vs downloading (~5–10 seconds)
- Identical installation command for local and CI (`go install github.com/goreleaser/goreleaser/v2@<version>`)
- `check-versions.sh` already checks the goreleaser version locally; CI and local installations are now verifiable by the same mechanism
- Version bump workflow: edit `goreleaser-version` in `cicd/config.yml` → update local install → `make release-snapshot` validates locally → push (CI installs fresh from new cache key)
- Consistent with existing installation pattern for all other tools in the pipeline

**Negative**:
- goreleaser v2 module path is `github.com/goreleaser/goreleaser/v2` — the `/v2` suffix is required. This is documented in the `install-goreleaser` command and in the Makefile's `release-snapshot` target comment.
- First CI run after cache invalidation (i.e., after a version bump) incurs the full `go install` download time. This is the same as the current behavior on every run — it is not a regression.

**Known gap (accepted)**:
- `make release-snapshot` cannot validate GitHub Release creation, Homebrew formula update, or git tag push — these require live secrets. The `--snapshot` flag intentionally skips these. Documented in the Makefile inline comment for the `release-snapshot` target.
