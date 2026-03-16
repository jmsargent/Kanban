# ADR-006: Replace godog/Gherkin with Internal Go BDD DSL

**Status**: Accepted
**Date**: 2026-03-16
**Deciders**: Morgan (Solution Architect)

---

## Context

The acceptance test suite uses godog v0.15.1 with Gherkin `.feature` files. The test model is correct — it exercises the compiled kanban binary as a subprocess — but the tooling introduces unnecessary friction:

1. **Indirection layer**: each scenario step is a natural-language string matched to a Go function via regex. The binding exists at runtime, not at compile time. A whitespace or punctuation change in a `.feature` file silently breaks the step at test run time.

2. **Split artefacts**: every test requires two files in sync — a Gherkin `.feature` file and a corresponding Go step definition file. Refactoring a step name requires changes in both, and Go tooling (rename, find-usages) does not traverse Gherkin strings.

3. **Dependency weight**: godog requires six `indirect` dependencies (`cucumber/gherkin`, `cucumber/messages`, `gofrs/uuid`, `hashicorp/go-immutable-radix`, `hashicorp/go-memdb`, `hashicorp/golang-lru`). None of these add value to a single-binary CLI tool.

4. **BDD readability is achievable without a BDD framework**: the `Given / When / Then` structure that makes Gherkin readable can be reproduced in plain Go using named functions as arguments to orchestrator functions.

---

## Decision

Replace the godog/Gherkin acceptance test suite with an internal Go DSL that:

- Provides `Given`, `When`, `Then`, `And` orchestrator functions.
- Uses a `Step` struct (`Description string` + `Run func(*Context) error`) as the unit of test behaviour.
- Provides step factory functions grouped as Setup, Action, and Assertion categories.
- Wraps step errors with `t.Fatalf("<Phase>: <description>: <err>")` to stop at the first failing step.
- Lives in `tests/acceptance/dsl/` — a shared package imported by all feature test files.
- Imports only the Go standard library (and optionally testify/assert for richer assertion messages).
- Imports nothing from `internal/`.

Migration is incremental. godog coexists with the new DSL until all scenarios are ported, then is removed.

---

## Alternatives Considered

### Keep godog

**Pros**: no migration work; team familiarity; established BDD ecosystem.

**Cons**: all problems described in Context remain. Regex step binding is never compile-time safe. Six transitive dependencies are retained for no additional value beyond what idiomatic Go provides natively.

**Rejected**: the ongoing maintenance cost (two-file sync, non-refactorable step strings) outweighs the migration cost.

---

### Switch to testscript (`golang.org/x/tools/cmd/testscript`)

testscript is a Go-native test runner that uses a mini scripting language (`.txt` files) to describe command-line tool scenarios.

**Pros**: purpose-built for CLI tool testing; Go standard library adjacent; good subprocess model.

**Cons**: introduces a new custom scripting language — trades Gherkin indirection for testscript indirection. The `.txt` files are not Go; they require learning a separate syntax and are not refactorable with Go tooling. Assertion expressiveness is limited compared to Go functions.

**Rejected**: the project already has a working subprocess model in Go. Adopting testscript replaces one non-Go layer with another. The internal DSL achieves the same result in plain Go.

---

### Plain table-driven tests with no BDD structure

Use standard `t.Run` subtests with table-driven inputs, invoking the binary directly with no Given/When/Then scaffolding.

**Pros**: maximally idiomatic Go; zero abstraction overhead.

**Cons**: loses the scenario narrative structure that makes acceptance tests readable as specifications. For a product with explicit user stories (US-01 through US-08), the BDD vocabulary (`Given` precondition, `When` action, `Then` outcome) communicates intent to non-engineer stakeholders reading the test suite.

**Rejected**: the readability benefit of the BDD structure is retained at negligible cost (four trivial orchestrator functions) by the internal DSL approach.

---

## Consequences

### Positive

- All acceptance test behaviour is expressed in compiled, type-checked Go. Broken step factories are caught at `go build` time.
- Go tooling (rename, find-usages, extract function) works across the entire test codebase without exception.
- `go test ./tests/acceptance/...` is the sole command needed after migration — no separate godog runner invocation.
- Seven dependencies removed from `go.mod` after migration completes.
- Test files read as BDD narrative: `Given(ctx, InAGitRepo()); When(ctx, IRunKanbanStart("TASK-001")); Then(ctx, TaskHasStatus("TASK-001", "in-progress"))`.

### Negative

- **Migration cost**: each Gherkin scenario must be ported to a `TestXxx` function. This is bounded work (three feature files, approximately 50 scenarios, most `@skip`).
- **Temporary coexistence period**: both test suites run in CI until migration is complete. Minor CI overhead.
- **Loss of `@tags` filter**: godog's `--tags` flag for running subsets is replaced by `go test -run <regex>`. This is equally capable but requires familiarity with Go's test filter syntax.

### Neutral

- The subprocess model (compiled binary invoked via `os/exec`) is unchanged. The hexagonal boundary — acceptance tests exercise the CLI driving port only — is preserved and enforced by `go-arch-lint`.
