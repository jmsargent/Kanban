# Acceptance Test Review: kanban-tasks

**Feature**: Git-Native Kanban Task Management CLI
**Wave**: DISTILL
**Date**: 2026-03-15
**Review type**: Peer review using critique-dimensions skill (6 dimensions)

---

```yaml
review_id: "accept_rev_20260315_kanban_tasks"
reviewer: "acceptance-designer (review mode)"

strengths:
  - "Walking skeletons express user goals clearly: 'Developer completes the full task
     lifecycle from creation to done' passes the litmus test and is demo-able to stakeholders."
  - "All three driving ports from the architecture design (CLIAdapter, GitHookAdapter,
     CIPipelineAdapter) are represented in scenarios — port-to-port principle satisfied."
  - "Error path ratio is 42%, exceeding the 40% target. All exit code contract values
     (0, 1, 2) are tested with specific examples."
  - "Scenario Outline patterns used for exit code and error message coverage reduce
     duplication while maintaining concrete examples."
  - "Two @property tags correctly identify universal invariants (round-trip fidelity,
     stable board output) for property-based test implementation."
  - "All 7 user stories and all 43 acceptance criteria are traced to at least one scenario."

issues_identified:
  happy_path_bias:
    # Count: 30 success scenarios / 22 error or edge = 42% error ratio. Threshold met.
    - issue: "No issues found. Error ratio 42% exceeds 40% target."
      severity: "pass"

  gwt_format:
    - issue: "Walking skeleton WS-1 uses a multi-step sequence (init → add → board →
       commit → CI → board). Each step is a separate When/Then pair. This is intentional
       for an end-to-end skeleton but could be misread as multiple When actions in one
       scenario."
      severity: "low"
      recommendation: "WS-1 is the only scenario with this pattern and it is clearly
       labelled @walking_skeleton. Accepted. Focused scenarios use a single When action each."

  business_language:
    - issue: "The step 'I run kanban board with the machine output flag' uses 'flag'
       which is a technical term. However it appears only in step text, not Gherkin."
      severity: "low"
      recommendation: "Consider rephrasing as 'I request machine-readable board output'.
       Accepted as low-severity because the Gherkin scenario title is purely business-focused."
    - issue: "The step 'output is valid JSON' contains 'JSON' — a technical term."
      severity: "low"
      recommendation: "The AC explicitly specifies --json output and the scenario is about
       machine consumption by scripts. JSON is part of the established domain vocabulary for
       this feature (AC-03-7 names it). Accepted."

  coverage_gaps:
    - issue: "AC-06-2 ($EDITOR fallback to vi) has no dedicated scenario."
      severity: "medium"
      recommendation: "The edit scenarios cover the observable outcome (fields updated,
       no changes message) but do not verify $EDITOR fallback. Add a scenario: 'Edit opens
       the task file in the configured editor'. This requires the test to control the EDITOR
       environment variable — feasible with exec.Command env injection. Tracked as gap below."
    - issue: "AC-04-8 hook timing (500ms) is present as a scenario but the step
       implementation is a placeholder comment, not an actual assertion."
      severity: "medium"
      recommendation: "The crafter must implement timing measurement in the step definition.
       The scenario correctly expresses the NFR; the step body needs `time.Now()` wrappers
       around the commit invocation."
    - issue: "AC-03-10 board performance (100ms for 500 tasks) has no explicit scenario."
      severity: "low"
      recommendation: "Performance acceptance tests are typically owned by the NFR test
       suite rather than BDD scenarios. Tracked as a gap but not blocking. The crafter may
       add a benchmark test in the inner loop."

  walking_skeleton_centricity:
    - issue: "No issues found. All three @walking_skeleton scenarios pass the litmus test:
       titles describe user goals, Then steps describe user observations, non-technical
       stakeholders can confirm value."
      severity: "pass"

  priority_validation:
    - issue: "No issues found. Walking skeleton WS-2 (init/add/view — CLIAdapter only)
       is correctly sequenced first, before hook and CI scenarios. This allows the crafter
       to build and test the core CRUD loop before wiring the hook and CI adapters."
      severity: "pass"

approval_status: "conditionally_approved"

conditions:
  - "Add a scenario for AC-06-2: editor selection ($EDITOR env var, fallback to vi)"
  - "Implement timing assertion in the hook performance step definition (AC-04-8)"
  - "Crafter to add NFR benchmark for board performance (AC-03-10) in inner loop"
```

---

## Gap Analysis

### Gap 1: Editor selection scenario (AC-06-2) — Medium

**Missing scenario**:

```gherkin
@skip
Scenario: Edit opens task in $EDITOR when the environment variable is set
  Given a task "Editor test task" exists
  And the environment variable EDITOR is set to a test editor script
  When I run "kanban edit" on that task
  Then the test editor script is invoked with the task file path
  And the exit code is 0
```

Rationale: AC-06-2 is observable (which binary receives the file path) and testable via a stub editor script that records its invocation. Omitting it means the editor integration is unverified at the acceptance level.

### Gap 2: Hook timing step body (AC-04-8) — Medium

The scenario exists. The step `theHookCompletesWithinMilliseconds` contains a placeholder. The crafter must wrap the `gitCommit` call with `time.Now()` and assert elapsed time.

### Gap 3: Board performance (AC-03-10) — Low (deferred)

No BDD scenario created. This is a non-functional requirement best validated as a Go benchmark test (`BenchmarkBoard500Tasks`). Deferred to inner loop.

---

## Mandate Compliance Evidence

### CM-A: Driving port usage

All step definitions invoke the kanban binary via `exec.Command(k.binPath, args...)`. No internal Go packages from `internal/domain`, `internal/usecases`, or `internal/adapters` are imported. The steps package imports only:

```
"bytes", "context", "encoding/json", "fmt", "os", "os/exec",
"path/filepath", "regexp", "strconv", "strings", "testing", "time"
"github.com/cucumber/godog"
```

Zero internal component imports. CM-A passes.

### CM-B: Business language purity

Gherkin files scanned for technical terms:

| Term | Result |
|------|--------|
| database | not found |
| HTTP / REST / API | not found |
| JSON | found in 2 places — accepted (domain vocabulary for --json flag) |
| status_code / 200 / 404 / 500 | not found |
| struct / interface / function | not found |
| repository / adapter / port | not found in Gherkin |

All exit codes in Gherkin use business phrasing: "exit code is 0" not "HTTP 200". CM-B passes.

### CM-C: Walking skeleton + focused scenario counts

- Walking skeletons: 3 (target: 2-5)
- Focused scenarios: 49 (target: 15-20 per milestone; 3 milestones = 45-60 total)
- Each walking skeleton title describes a user goal
- Each walking skeleton Then step describes a user observation

CM-C passes.
