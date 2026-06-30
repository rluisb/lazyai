---
name: refactor
description: Use when improving structure without changing behavior, with explicit invariants and verification that the contract stays the same. Uses write tools, shell execution, and may spawn subagents. Invoke with enable_write_tools=true and enable_subagent_tools=true.
status: draft
---

# Refactor Workflow

## Trigger

Start when code structure, naming, or boundaries need improvement but externally observable behavior must remain the same.

## Inputs

- `target` — files/modules/paths to reshape
- `behavior_invariant` — what must stay unchanged
- `validate` — tests or checks proving no behavior change
- `risk` — `low | normal | high`
- `tdd_mode` — `lightweight | medium | heavy-aggressive | required`

## Purpose

Improve maintainability, readability, or boundaries while preserving the current contract exactly.

## Four Points

- WHAT: Reshape code without changing behavior.
- HOW: Name the invariant, plan the smallest structural change, preserve tests, verify no drift.
- DON'T WANT: Drive-by feature work, silent contract changes, or parallel conventions.
- VALIDATE: Existing or added invariant checks still pass before and after the refactor.

## Behavior Scenario

- Given the current supported behavior
- When the same public path is exercised after refactor
- Then the observable result is unchanged

## TDD Mode

Choose and record a TDD mode when coverage needs tightening around the invariant; otherwise record how existing tests prove the invariant.

## Steps

1. Name the invariant and touch points.
2. Read the current pattern fully; do not guess structure.
3. Add or identify invariant coverage.
4. Apply the smallest structural change.
5. Re-run the invariant checks and the next-smallest affected repo gate.
6. Remove obsolete scaffolding or comments only after the invariant is proven.

## Adapters

- Claude Code: markdown workflow only; existing planning/reviewer skills enforce the invariant mindset.
- OpenCode: markdown workflow only; no runtime claims.
- OMP/Pi: markdown-only guidance.

## Exit Gate

Stop only when the structural change is complete, behavior invariants are observed as unchanged, and no hidden feature work was absorbed.

## Failure

Stop and report when the refactor requires contract changes, coverage cannot prove the invariant, or the work expands into a new feature/bugfix.
