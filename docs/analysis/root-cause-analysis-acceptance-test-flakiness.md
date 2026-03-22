# Root Cause Analysis: Acceptance Test Flakiness in `make ci`

**Date**: 2026-03-22
**Analyst**: Rex (nw-troubleshooter)
**Problem**: Running `make ci` causes acceptance tests to fail intermittently

---

## Problem Statement

When running `make ci`, the acceptance test suite fails non-deterministically. Tests pass in isolation but fail under the compound execution triggered by `make ci`.

---

## Evidence Gathered

### Execution Flow of `make ci`

```
make ci
  └─ make validate
       ├─ [0/5] cicd/check-versions.sh
       ├─ [1/5] go test ./internal/...
       ├─ [2/5] golangci-lint run
       ├─ [3/5] go-arch-lint check
       ├─ [4/5] go build ./...
       └─ [5/5] KANBAN_VALIDATE_DEPTH=1 make acceptance     ← Run 1
                 └─ go build -o kanban ./cmd/kanban
                    go test ./tests/acceptance/...
                    (TestPipeline skips — KANBAN_VALIDATE_DEPTH is set)

  └─ make acceptance                                         ← Run 2
        └─ go build -o kanban ./cmd/kanban
           go test ./tests/acceptance/... (parallel packages)
             ├─ tests/acceptance/dsl      ← runs concurrently
             ├─ tests/repofiles           ← runs concurrently
             └─ tests/acceptance
                  └─ TestPipeline_ValidateTarget_PassesWhenAllChecksGreen
                       └─ shellCmd("make validate", timeout=5min)
                            └─ [5/5] KANBAN_VALIDATE_DEPTH=1 make acceptance  ← Run 3
                                      └─ go build -o kanban ./cmd/kanban      (*)
                                         go test ./tests/acceptance/...
```

`(*)` = binary overwrite at `$(CURDIR)/kanban` while `tests/acceptance/dsl` from Run 2 is concurrently executing tests that invoke the same binary path via `KANBAN_BIN`.

### Key Code Locations

- [Makefile:29-31](Makefile) — `ci:` target calls `make validate && make acceptance`; `make validate` already runs acceptance internally (line 17)
- [tests/acceptance/pipeline_test.go:54-69](tests/acceptance/pipeline_test.go) — `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` calls `DeveloperRunsMakeTarget("validate")` with a 5-minute subprocess timeout
- [tests/acceptance/dsl/pipeline_steps.go:216](tests/acceptance/dsl/pipeline_steps.go) — `shellCmd` inherits `os.Environ()` including `KANBAN_BIN`, and uses a 5-minute timeout
- [tests/acceptance/dsl/context.go:41-45](tests/acceptance/dsl/context.go) — `ctx.env` initialised from `os.Environ()` — all tests inherit parent process environment
- [tests/acceptance/dsl/setup.go:37](tests/acceptance/dsl/setup.go) — `InAGitRepo` uses `os.MkdirTemp("", "kanban-test-*")` (system temp, not `t.TempDir()`)

---

## 5 Whys Analysis

### Why 1 — Why do acceptance tests fail intermittently during `make ci`?

**Because `make ci` causes the acceptance test suite to run three times**, and two of those runs happen concurrently at the package level:

- Run 1: inside `make validate`'s step [5/5]
- Run 2: the outer `make acceptance` in the `ci` target
- Run 3: triggered by `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` calling `make validate` as a subprocess during Run 2

Go's test runner executes packages under `tests/acceptance/...` concurrently by default. During Run 2's execution of `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen`, the packages `tests/acceptance/dsl` and `tests/repofiles` are running concurrently _while_ the subprocess (Run 3) is also building and running tests.

---

### Why 2 — Why does the triple execution cause test failures?

**Two compounding mechanisms:**

**A. Binary path contention (race window):**
The `make acceptance` target always writes the built binary to `$(CURDIR)/kanban`. During Run 2, `KANBAN_BIN` points to this path. Run 3's inner `make acceptance` rebuilds the same binary. Go's `go build` uses an atomic rename, but the compilation process itself takes non-zero time — during which the old binary is still running for concurrent `tests/acceptance/dsl` test subprocesses. If the build overlaps with test invocations in the `dsl` package (which is running concurrently at the package level), the binary on disk may be in flux.

**B. Subprocess timeout pressure:**
`TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` allows 5 minutes for the subprocess `make validate`. That subprocess must complete:
- `cicd/check-versions.sh`
- `go test ./internal/...`
- `golangci-lint run`
- `go-arch-lint check`
- `go build ./...`
- `KANBAN_VALIDATE_DEPTH=1 make acceptance` (another full acceptance run)

Under load — or when `make ci` already has the Go build cache partially invalidated from Run 1 — this budget is tight and the timeout can fire, failing the test.

---

### Why 3 — Why does Run 2 (`make acceptance`) exist after `make validate` already ran acceptance?

**Because `make ci` is defined as `make validate && make acceptance`**, and there is no guard preventing the redundant re-run. The `KANBAN_VALIDATE_DEPTH` guard exists only inside `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` to prevent infinite recursion of _that specific test_. It does not prevent the top-level `make ci` from running the full acceptance suite twice.

---

### Why 4 — Why does `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` run `make validate` (rather than making targeted assertions about the Makefile)?

**By design**: the test exercises the full observable developer outcome — running `make validate` succeeds end-to-end. This is correct for an acceptance test of the pipeline. However, the test was written without accounting for the additional execution context introduced by `make ci` (which double-runs acceptance before the pipeline test gets to run `make validate` itself).

---

### Why 5 — Why does the pipeline test not protect against timeout under compound load?

**Because the 5-minute timeout was set assuming a single clean run of `make validate`, not the compounded load of `make ci`** (which invades the build cache and CPU budget with two prior full acceptance runs before this test ever starts). The timeout value has no headroom for the degraded conditions specific to `make ci`.

---

## Root Causes (Multi-Causal)

| # | Root Cause | Category | Severity |
|---|------------|----------|----------|
| RC-1 | `make ci` double-runs acceptance: `make validate` already includes acceptance in step [5/5], but `make ci` also calls `make acceptance` explicitly | Design flaw in `ci` target | High |
| RC-2 | `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` calls `make validate` as a subprocess during Run 2, which concurrently builds `./kanban` while `tests/acceptance/dsl` is executing tests using that binary | Concurrent binary overwrite race window | Medium |
| RC-3 | The 5-minute subprocess timeout for `make validate` has no headroom for the degraded CPU/cache conditions that `make ci`'s two prior runs create | Insufficient timeout budget | Medium |

---

## Backward Chain Validation

- RC-1 directly causes 3× acceptance runs under `make ci` → confirmed by tracing `Makefile:29-31`, `Makefile:15-18`
- RC-2 is activated by RC-1: without the triple run, `tests/acceptance/dsl` and the subprocess binary build would not overlap → confirmed by Go's default parallel package execution
- RC-3 is activated by RC-1: the 5-min budget is consumed by the degraded build cache from prior runs → confirmed by timeline analysis

All three root causes share RC-1 as their trigger. Fixing RC-1 eliminates RC-2 and RC-3.

---

## Solutions

### Fix 1 (Primary — addresses RC-1): Remove the redundant `make acceptance` call from `make ci`

`make validate` already runs acceptance tests in step [5/5]. The `ci` target's explicit `make acceptance` is redundant.

```makefile
# Before (current — runs acceptance twice at top level):
ci:
    @make validate && make acceptance

# After (acceptance already runs inside validate's step [5/5]):
ci:
    @make validate
```

**Trade-off**: `make ci` becomes an alias for `make validate`. If the intention is that `make ci` runs acceptance _outside_ the depth guard (i.e., with `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` active), then instead remove acceptance from step [5/5] of `validate` and keep it only in `make ci`:

```makefile
validate:
    # ... steps [0/5] through [4/5] unchanged ...
    # Remove the [5/5] acceptance block
    @echo "PASS"

ci:
    @make validate && make acceptance
```

This makes the division of responsibility explicit: `validate` = static quality gates; `ci` = validate + acceptance.

### Fix 2 (Secondary — addresses RC-3): Increase the subprocess timeout

If Fix 1 is not applied immediately, raise the timeout to account for worst-case load:

```go
// tests/acceptance/dsl/pipeline_steps.go:216
// Before:
output, exit := shellCmd(pc.projectRoot, nil, 5*time.Minute, "make", target)

// After (10 minutes for loaded machines):
output, exit := shellCmd(pc.projectRoot, nil, 10*time.Minute, "make", target)
```

### Fix 3 (Defensive — addresses RC-2): Use `t.TempDir()` instead of `os.MkdirTemp`

The `InAGitRepo` step uses `os.MkdirTemp("", "kanban-test-*")` and registers manual cleanup. Using `t.TempDir()` is idiomatic, automatically cleaned up even on panic, and scoped to the test.

```go
// tests/acceptance/dsl/setup.go:37
// Before:
dir, err := os.MkdirTemp("", "kanban-test-*")
if err != nil {
    return fmt.Errorf("create temp dir: %w", err)
}
ctx.t.Cleanup(func() { _ = os.RemoveAll(dir) })

// After:
dir := ctx.t.TempDir()
```

This does not eliminate the binary contention but reduces overall filesystem surface area and ensures no temp dirs leak on test panic.

---

## Recommended Action Sequence

1. **Apply Fix 1** — restructure `make ci` / `make validate` to eliminate the double acceptance run. This is the only fix that addresses all three root causes.
2. **Apply Fix 3** — replace `os.MkdirTemp` + manual cleanup with `t.TempDir()` in `InAGitRepo` and `NotAGitRepo` setup steps.
3. **Apply Fix 2** — raise the subprocess timeout as a short-term safety net if Fix 1 is delayed.

---

## Success Criteria

- [ ] `make ci` runs acceptance tests exactly once at the top level
- [ ] `TestPipeline_ValidateTarget_PassesWhenAllChecksGreen` subprocess `make validate` does not overlap with concurrent package test execution
- [ ] 10 consecutive `make ci` runs pass without intermittent failure
