---
name: pr-review
description: Review a change set against requirements, tests, regressions, and repository conventions.
trigger: /pr-review
tier: frontier
thinking: high
risk: 4
---

# PR Review

## Workflow

1. Read the requested scope and changed files.
2. Check tests and verification claims before style or taste.
3. Compare behavior against the approved contract.
4. Report blockers, risks, and follow-ups with exact evidence.

## Guardrails

- Findings must cite the changed seam.
- Separate must-fix issues from nice-to-have feedback.
- Do not silently fix code while acting as reviewer.
