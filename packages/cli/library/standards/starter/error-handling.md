# Standard: Error Handling

**Category:** Code
**Scope:** project
**Date:** 2026-05-01
**Owner:** AI Setup Maintainers
**Status:** Active
**Constitution article(s):** III, V, VI

> **Purpose.** Keep failures actionable by requiring errors to include context, preserve recoverability decisions, and avoid silent fallback behavior that hides contract violations.

---

## Scope Cascade

This starter standard is written at project scope and applies to application code, scripts, validators, and scaffolding logic until a more specific error-handling standard replaces it.

```
global standards
        ↓
workspace standards
        ↓
project standards
        ↓
runtime and test failures
```

**Where this standard lives:** project

**Overrides:** none — this is starter guidance for new projects.

---

## Rule

Errors MUST state the failing operation and relevant file, command, or contract, and fallbacks MUST be explicit in code or task evidence.

**Trigger:** Reading files, parsing configuration, running external commands, validating generated content, or handling optional integrations.

---

## Rationale

Clear error context shortens recovery and keeps rollback decisions evidence-based. Silent fallback behavior can make invalid generated artifacts appear successful.

- **Article support:** Article III requires errors to reference the contract being checked; Article V favors understandable recovery paths; Article VI rejects hidden complexity.
- **Origin:** Wave 1 starter standards for concrete pattern review.

---

## Examples

**Compliant:**
```
return fmt.Errorf("read standards/starter/%s: %w", name, err)
```

**Non-compliant:**
```
return err
```

**Why the non-compliant case fails:** The caller cannot tell which operation or artifact failed without re-running with extra instrumentation.

---

## Enforcement

| Mechanism | Where | When |
|---|---|---|
| Unit tests for failure paths | Package tests near parsing or scaffolding code | Gate 3 Behavioral Validation |
| Code review | Error-returning branches | Gate 4 Pattern Consistency |
| Handoff evidence | Task summary for known fallback paths | Before release or rollback |

---

## Exceptions

Exceptions require a reviewer note showing that the surrounding caller already adds equivalent context.

- Sentinel error comparison — allowed when wrapping would break documented equality checks.

---

## Workspace Awareness

| Repo | Override | Reason |
|---|---|---|
| All repos | none | Starter guidance applies until repo-specific error standards exist. |

---

## Related

- **Supersedes:** none
- **Related ADR(s):** none
- **Related standards:** test-patterns, context-loading

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] Incidents caused by missing error context link back to this standard.
- [ ] New sentinel-error exceptions are documented with their owning package.
