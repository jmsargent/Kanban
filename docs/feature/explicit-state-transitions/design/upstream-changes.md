# Upstream Changes — explicit-state-transitions DESIGN wave

**Date**: 2026-03-22
**For review by**: Product Owner (Luna)

---

## Change 1: `kanban init` no longer auto-commits

**Origin**: Design-time code audit of `internal/usecases/init_repo.go`

**Original DISCUSS assumption** (implicit — not explicitly scoped):
DISCUSS scoped the "no commit" change to `kanban ci-done`. It did not address `kanban init`.

**Reality found in code**:
`InitRepo.Execute()` calls:
- `git.InstallHook(repoRoot)` — installs hook to `.git/hooks/commit-msg`
- `git.CommitFiles(repoRoot, "kanban: initialise repository", [...])` — commits `.kanban/config` and `.gitignore`

Both are incompatible with constraint C-03 ("binary never calls git commit or git add").

**New behaviour**:
`kanban init` creates `.kanban/tasks/`, writes `.kanban/config`, and exits. It does NOT install the hook, does NOT call `git add`, does NOT call `git commit`. The developer runs `git add .kanban/ && git commit -m "kanban: init"` after the command.

**User-facing impact**:
Developers who previously relied on `kanban init` to auto-commit the setup must add a `git add && git commit` step to their onboarding doc / README. The init command output should be updated to print this reminder.

**Suggested new acceptance criterion (for DISCUSS backfill)**:
> AC-init-1: Given `kanban init` runs successfully, then no git commit or git add subprocess is invoked, and stdout includes a reminder to commit `.kanban/` manually.

---

## Change 2: `kanban log` output format changes

**Origin**: Design decision D6 — `GetTaskHistory` simplified to use only `GitPort.LogFile()`

**Original DISCUSS assumption** (implicit):
`kanban log` was not in scope of this feature. Its behaviour was assumed stable.

**Reality**:
`GetTaskHistory` previously merged transitions.log entries (structured `from→to` data) with git log entries. After removing transitions.log, `kanban log` shows only the git commit history for the task file. The structured `from→to` format is no longer shown — commits show the raw commit message, author, and timestamp.

**Impact**:
Low — `kanban log` is a read-only display command. The output is less structured (no `from→to` annotations) but the git history is still useful and complete.

**No acceptance criteria change needed** — `kanban log` was not covered in the feature's ACs.
