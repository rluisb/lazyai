---
name: codebase-exploration
description: Locate the real entry points, relevant files, and existing patterns before making a change.
trigger: /codebase-exploration
tier: balanced
thinking: low
risk: 2
---

# Codebase Exploration

## Use when

You know the user goal, but not yet the exact files, symbols, or seams to change.

## Workflow

1. Search for the owning commands, packages, or symbols.
2. Read the smallest complete sections that explain the behavior.
3. Note existing patterns, invariants, and validation paths.
4. Hand the implementation task off only after the touch map is concrete.

## Guardrails

- Prefer search + targeted reads over whole-file scans.
- Do not speculate about code behind collapsed summaries.
- Stop when the next edit target is grounded.
