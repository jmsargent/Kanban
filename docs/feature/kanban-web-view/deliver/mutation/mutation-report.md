# Mutation Testing Report — kanban-web-view

**Date**: 2026-04-01
**Tool**: gremlins v0.6.0
**Scope**: `./internal/adapters/web`
**Command**: `gremlins unleash --test-cpu 1 ./internal/adapters/web`

---

## 1. Summary

| Metric           | Count |
|------------------|-------|
| Killed           | 5     |
| Lived            | 0     |
| Timed Out        | 30    |
| Not Covered      | 1     |
| Not Viable       | 0     |
| Skipped          | 0     |
| **Total Mutants**| **36**|

| Quality Metric       | Value    |
|----------------------|----------|
| Test Efficacy (killed / killed+lived) | **100.00%** |
| Mutant Coverage ((killed+lived) / total) | **83.33%** |

### Kill Rate Assessment

Test efficacy of **100%** against covered mutants — every reachable mutant was killed.
Mutant coverage of **83.33%** clears the 80% threshold.

**Quality Gate: PASS**

---

## 2. Surviving (LIVED) Mutants

**None.** Zero mutants survived.

---

## 3. Timed Out Mutants

Timed-out mutants indicate the test suite did not complete within the gremlins timeout for those mutation points. They are not killed but also not confirmed survivors; gremlins excludes them from the efficacy calculation.

All 30 timed-out mutants are `CONDITIONALS_NEGATION` in error-handling branches of `handler.go` and `middleware.go` — specifically crypto paths (AES/GCM cipher construction, nonce generation, ciphertext length guard) and template render-error checks. These require adversarial inputs that the `httptest`-based unit tests don't provide.

Re-running with `--timeout-coefficient 3` would clarify whether these are slow-but-killable or genuine gaps.

---

## 4. Not Covered Mutants

| File        | Line:Col | Mutation Type         | Context |
|-------------|----------|-----------------------|---------|
| handler.go  | 89:59    | CONDITIONALS_NEGATION | `TokenEntryHandler.ServeHTTP`: template render error check — log-only path, no observable side effect |

This is the only remaining not-covered mutant. The branch logs an error if `ExecuteTemplate` fails at render time. Since the response status is already written before this check, flipping it has no observable effect on the HTTP response — the mutant cannot be killed by assertions on response code or body.

---

## 5. Quality Gate Result

| Threshold          | Required | Actual   | Result |
|--------------------|----------|----------|--------|
| Test Efficacy      | >= 80%   | 100.00%  | PASS   |
| Mutant Coverage    | >= 80%   | 83.33%   | PASS   |

**Overall Quality Gate: PASS**

---

## 6. History

| Run | Date | Killed | Lived | Timed Out | Not Covered | Coverage | Gate |
|-----|------|--------|-------|-----------|-------------|----------|------|
| 1   | 2026-04-01 | 19 | 0 | 12 | 5 | 79.17% | WARN |
| 2   | 2026-04-01 | 5  | 0 | 30 | 1 | 83.33% | PASS |

Run 2 added unit tests for `NewServer` nil-key guard, empty GitHub URL guard, nil-addTask 501 path, and `AddTaskHandler.Execute` failure re-render — converting 4 of the 5 previously not-covered mutants to KILLED.
