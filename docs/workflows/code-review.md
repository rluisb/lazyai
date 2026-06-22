---
name: code-review
description: Use when reviewing code changes with explicit stage, target, and focus so own-code checks and others’ PR reviews stay distinct but share one contract.
status: draft
---

# Code Review Workflow

## Trigger

Start when code changes need review before commit, before PR, during PR review, or while responding to review feedback.

## Inputs

- `review_target` — `own | others`
- `review_stage` — `pre-commit | pre-pr | pr | post-feedback`
- `input_shape` — `diff | branch | pr | comments`
- `focus` — `correctness | regression | security | scope | docs | mixed`
- `validate` — expected review output or follow-up path

## Purpose

Catch correctness, regression, scope, and evidence problems before the code ships or before feedback is answered.

## Four Points

- WHAT: Produce an actionable review of the code changes.
- HOW: Read the relevant diff/context, compare against specs/docs/tests, classify findings by severity and evidence.
- DON'T WANT: Generic praise, style-only noise, or comments without proof.
- VALIDATE: Review output names concrete findings or explicitly states none were found in the checked scope.

## Behavior Scenario

- Given the proposed code change and its claimed behavior
- When the relevant diff, tests, and supporting docs are reviewed
- Then the review identifies real issues or explicitly confirms the checked scope is clear

## TDD Mode

n/a for review-only runs. If review feedback reveals missing behavior coverage, point back to the required TDD mode for the implementation workflow.

## Steps

1. Confirm review target, stage, focus, and source material.
2. Read the diff and the smallest surrounding context needed to judge it.
3. Check behavior/test evidence against the claimed outcome.
4. Classify findings by severity, evidence, and whether they are questions or change requests.
5. Return the review summary and the smallest concrete next action.

## Adapters

- Claude Code: markdown workflow only; reviewer/feedback-review style work stays in existing tools and docs.
- OpenCode: markdown workflow only; no workflow runtime claims.
- Pi: markdown-only guidance.

## Exit Gate

Stop only when the review scope is explicit, every non-trivial finding is evidence-backed, and the next action is clear.

## Failure

Stop and report when the diff/context is incomplete, the claimed behavior has no observable evidence, or the requested verdict would require guessing.
