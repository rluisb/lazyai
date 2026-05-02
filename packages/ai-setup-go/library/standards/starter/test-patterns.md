# Standard: Test Patterns

**Category:** Testing
**Scope:** project
**Date:** 2026-05-01
**Owner:** AI Setup Maintainers
**Status:** Active
**Constitution article(s):** II, III, V

> **Purpose.** Make expected testing behavior concrete so new work begins with observable failure, ends with targeted verification, and gives reviewers a standard to cite for Quality Gate 4 Pattern Consistency.

---

## Scope Cascade

This starter standard is written at project scope and applies to feature, bugfix, refactor, and standards-content tasks until a more specific testing standard replaces it.

```
global standards
        ↓
workspace standards
        ↓
project standards
        ↓
tests and quality gates
```

**Where this standard lives:** project

**Overrides:** none — this is starter guidance for new projects.

---

## Rule

Behavior-changing work MUST identify or add a targeted failing test before implementation and then run the nearest focused test plus the relevant package quality gate.

**Trigger:** Adding behavior, fixing a bug, changing generated content, or altering a public contract.

---

## Rationale

Tests are the fastest way to prove that a standard is not just prose. A targeted red test anchors scope; the package gate catches integration drift.

- **Article support:** Article II requires test-first behavior; Article III links tests to accepted contracts; Article V keeps verification close to the changed surface.
- **Origin:** Wave 1 starter standards for concrete quality-gate review.

---

## Examples

**Compliant:**
```
1. Add a test that expects exactly five starter standards.
2. Run the test and capture the failure.
3. Add the five files.
4. Re-run the focused test and package gate.
```

**Non-compliant:**
```
Add content first, skip a focused validation test, and rely on manual inspection only.
```

**Why the non-compliant case fails:** It leaves no repeatable evidence that future changes preserve the expected contract.

---

## Enforcement

| Mechanism | Where | When |
|---|---|---|
| Focused regression test | Package test file nearest the changed contract | Before and after implementation |
| Package quality gate | Repository or package test command | Before handoff or PR |
| Review checklist | Task summary and PR description | Gate 4 Pattern Consistency |

---

## Exceptions

Exceptions require a task summary note explaining why a failing test cannot be added and what repeatable validation replaced it.

- Documentation-only typo fix — targeted test may be omitted if no behavior or generated contract changes.

---

## Workspace Awareness

| Repo | Override | Reason |
|---|---|---|
| All repos | none | Starter guidance applies until repo-specific testing standards exist. |

---

## Related

- **Supersedes:** none
- **Related ADR(s):** none
- **Related standards:** orchestration-patterns, error-handling

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] New test patterns discovered during implementation are captured in project standards.
- [ ] Repeated manual-only validation is reviewed for automation opportunities.
