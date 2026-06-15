---
name: test-first-change
description: Use when changing behavior so the agent drives the edit through a failing test, preserves existing tests, follows the selected TDD mode, and verifies red-green-refactor evidence.
---

# Test-First Change

## When to Use

Use this skill for behavior changes, bug fixes, parser changes, validation logic, public API changes, and non-trivial refactors where behavior can be expressed as a test.

Skip only for documentation-only changes, generated-output-only changes, mechanical renames, or an explicitly approved exemption.

## Rule

Behavior changes need a failing test first unless explicitly exempt; existing tests must not be deleted, skipped, weakened, or rewritten to pass without approval.

## TDD Mode

Use `canonical/tdd-planning.md` to select the mode before editing:

- `lightweight` — low-risk narrow change; record test intent in the task plan.
- `medium` — normal implementation; add `## TDD Plan` to the plan.
- `heavy-aggressive` — critical or high-regression work; create standalone `.vibe-lab/tdd/<slug>.md` or equivalent spec section.
- `required` — implementation is blocked until red test exists or exemption is approved.

## Workflow

1. Identify the behavior contract in one sentence.
2. Confirm or choose the TDD mode.
3. Find the smallest existing public test layer.
4. Add or adjust one test that fails for the right reason.
5. Run only that test and observe the failure.
6. Implement the smallest change that makes the test pass.
7. Run the focused test again.
8. Refactor only after green.
9. Run the next-smallest relevant check.

## Exemption Format

```md
Test-first exemption: <why a failing test is not practical>
Approval source: <user | plan | spec | existing repo constraint>
Validation instead: <command or manual check that proves the behavior>
Risk: <what this does not cover>
```

## Test Choice

Prefer tests that observe behavior through the public seam already used by callers.

- Bug fix: reproduce the bug with the narrowest regression test.
- API change: test public contract and error behavior.
- Refactor: pin externally visible behavior before moving internals.
- CLI/script: run the command or script-level test with representative input.

## Constraints

- Do not write tests that assert private implementation details.
- Do not add mocks when a real local seam exists.
- Do not broaden the suite until the focused test is green.
- Do not keep a test that passes before the behavior change unless it proves a documented invariant.
- Do not remove existing tests to make verification pass.
