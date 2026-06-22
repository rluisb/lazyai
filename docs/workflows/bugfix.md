---
name: bugfix
description: Use when fixing broken behavior and the priority is reproduce, root cause, regression proof, and the smallest safe repair.
status: draft
---

# Bugfix Workflow

## Trigger

Start when behavior is wrong, a test/build/runtime failure must be repaired, or a regression needs a direct fix.

## Inputs

- `symptom` — failing behavior or error
- `expected_behavior` — what should happen instead
- `validate` — failing and passing proof path
- `risk` — `low | normal | high`
- `tdd_mode` — `lightweight | medium | heavy-aggressive | required`

## Purpose

Repair the real root cause with the smallest safe change and leave behind regression evidence.

## Four Points

- WHAT: Restore correct behavior.
- HOW: Reproduce, isolate root cause, add regression proof, patch minimally, verify.
- DON'T WANT: Symptom suppression, broad cleanup, or speculative hardening unrelated to the bug.
- VALIDATE: Red repro first, green regression proof after the fix.

## Behavior Scenario

- Given the failing pre-fix state
- When the repro path runs
- Then the pre-fix check fails and the post-fix check passes with the intended behavior

## TDD Mode

Choose and record the TDD mode before editing. Regression coverage is required unless an explicit exemption is approved.

## Steps

1. Reproduce the bug or capture the exact observed failure.
2. State the root-cause hypothesis and verify it against code/tests.
3. Add or tighten the regression check.
4. Implement the smallest source fix.
5. Re-run the focused repro/regression path.
6. Run the next-smallest affected repo gate.
7. Update docs only if the user-facing contract or workflow changed.

## Adapters

- Claude Code: markdown workflow only; existing diagnose/test-first skills and agents perform the work.
- OpenCode: markdown workflow only; no runtime orchestration is claimed.
- Pi: markdown-only guidance.

## Exit Gate

Stop only when the bug is reproduced, the root-cause fix is landed, and the regression proof passes.

## Failure

Stop and report when the failure cannot be reproduced, the suspected fix only masks the symptom, or required regression evidence cannot be observed.
