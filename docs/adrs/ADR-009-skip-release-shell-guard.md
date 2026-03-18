# ADR-009: [skip release] Implementation via Shell Guard

**Status**: Accepted
**Date**: 2026-03-18
**Feature**: testable-pipeline
**Resolves**: OD-01 ([skip release] implementation mechanism)

---

## Context

The developer needs to push commits to `main` for CI validation purposes without triggering the `tag` and `release` jobs (which produce a real GitHub Release and update the Homebrew formula). This is needed when validating CI configuration changes that affect the tag or release jobs and cannot be fully reproduced locally.

The DISCUSS wave decided the convention: `[skip release]` in the commit message. Three implementation mechanisms were identified (US-TP-04 technical notes):

1. **CircleCI pipeline parameter** set via API trigger (requires a trigger script)
2. **CircleCI workflow `when` condition** using `<< pipeline.git.commit_message >>` parameter
3. **Shell `if` guard** inside the job script, reading `git log -1 --pretty=%B`

---

## Decision

Implement as a shell `if` guard (Option 3) inside the `tag` and `release` job scripts.

The guard reads the most recent commit message at job runtime and exits 0 early if `[skip release]` is present (case-insensitive):

```sh
COMMIT_MSG=$(git log -1 --pretty=%B)
if echo "$COMMIT_MSG" | grep -qi '\[skip release\]'; then
  echo "[skip release] detected — skipping tag/release"
  exit 0
fi
```

This guard is inserted in the `tag` job after the existing "already tagged" guard, and identically in the `release` job.

The matching behavior: case-insensitive, position-independent in the commit message (grep -i matches anywhere in the string).

---

## Alternatives Considered

### Alternative 1: CircleCI pipeline parameter via API

Set a `skip_release` pipeline parameter to `true` when triggering via the CircleCI API. Add a `when` condition on the `tag` and `release` jobs in the workflow.

Rejection rationale:
- Requires a separate API trigger script or manual API call — the developer would need to invoke a non-standard push command for "skip release" pushes. The convention must be embeddable in the commit message for it to be low-friction.
- Does not support the commit message convention (developer intention is expressed in the commit message, not in an API parameter)
- Adds operational complexity: the developer must know to use `curl` or a trigger script instead of `git push`

### Alternative 2: CircleCI workflow `when` condition with `<< pipeline.git.commit_message >>`

Add a `when` condition to the workflow's `tag` and `release` job entries using CircleCI's pipeline value `<< pipeline.git.commit_message >>`.

Rejection rationale:
- `<< pipeline.git.commit_message >>` is populated only for pipelines triggered via the CircleCI API with an explicit `pipeline.git.commit_message` parameter; it is NOT populated for pipelines triggered by a VCS push (the normal case for this project)
- This has been a recurring source of confusion in CircleCI's documentation; the parameter is effectively unavailable for push-triggered pipelines
- Using it would produce a condition that appears syntactically correct but never matches, silently failing to implement the skip behavior

---

## Consequences

**Positive**:
- Zero additional tooling or API calls — the convention is entirely in the commit message
- Jobs still appear in the CircleCI UI (green, exited early) — full audit trail of what ran and why it skipped
- The guard is POSIX sh — no CircleCI-specific features required
- Case-insensitive matching (`grep -qi`) and position-independence match the AC in US-TP-04
- Recovery from accidental `[skip release]` is a single empty commit: `git commit --allow-empty -m "ci: trigger release"`

**Negative**:
- The `tag` and `release` jobs still run (job execution time consumed) even when skipping — they exit 0 after the guard, not before job start. Mitigation: the guard runs at the start of the relevant run step, so job overhead is ~seconds (checkout + guard check), not the full job runtime.
- The convention relies on developer discipline. Mitigation: single-developer project; discipline cost is near zero. The recovery pattern is documented in `cicd/config.yml` inline and in the project README/CONTRIBUTING.
