---
name: feature
description: Use when adding or changing product behavior that must ship with explicit purpose, scenarios, and verification.
status: draft
---

# Feature Workflow

## Trigger

Start when the task adds or changes externally observable behavior, a user-visible capability, or a stable internal contract.

## Inputs

- `goal` — plain-language outcome
- `constraints` — non-goals, boundaries, rejected approaches
- `validate` — observable proof path
- `risk` — `low | normal | high`
- `spec_mode` — `plain | speckit | auto`
- `tdd_mode` — `lightweight | medium | heavy-aggressive | required`

## Purpose

Deliver the requested behavior with the smallest change that satisfies the goal, preserves existing patterns, and keeps vibe-lab out of runtime ownership.

## Four Points

- WHAT: Ship the requested feature behavior.
- HOW: Clarify missing facts, research existing patterns when needed, plan with TDD mode, implement the smallest viable change, then verify.
- DON'T WANT: Workflow-engine creep, speculative abstractions, or unsupported runtime claims.
- VALIDATE: Focused red/green evidence plus the smallest repo gate for touched surfaces.

## Behavior Scenario

- Given a caller or user in the relevant starting state
- When they exercise the new or changed behavior
- Then the observable result matches the request and validation path

## TDD Mode

Choose and record a TDD mode before code changes. If no red test is practical, record an explicit exemption and validation substitute.

## Steps

1. Confirm the Four Points; clarify only material gaps.
2. Research current code/docs/tests to reuse existing patterns.
3. Write the plan, including behavior scenario, boundaries, and TDD mode.
4. Add the red test or approved exemption.
5. Implement the smallest behavior change.
6. Run focused verification, then the next-smallest affected repo gate.
7. Update directly affected docs/adapters only after the behavior works.

## Adapters

- Claude Code: workflow is consumed as markdown context; implementation uses existing skills/agents/hooks.
- OpenCode: workflow is consumed as markdown context; no workflow runtime is claimed.
- OMP/Pi: markdown-only guidance; no workflow or hook runtime support is claimed.

## Exit Gate

Stop only when the requested behavior is observed through validation, the TDD evidence or exemption is recorded, and any out-of-scope discoveries are called out instead of silently absorbed.

## Failure

Stop and report when the goal conflicts with current constraints, required validation cannot be run, or the implementation would require a new runtime/state surface not forced by the request.
