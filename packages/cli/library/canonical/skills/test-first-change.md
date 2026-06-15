---
name: test-first-change
description: Drive behavior changes through a failing test, then make the smallest code change that turns it green.
trigger: /test-first-change
tier: balanced
thinking: low
risk: 3
---

# Test First Change

## Workflow

1. Write or update the narrowest test that proves the requested behavior.
2. Confirm it fails for the right reason.
3. Implement the smallest change that makes it pass.
4. Re-run the targeted seam, then the affected package checks.
5. Refactor only if behavior stays unchanged.

## Guardrails

- No greenfield behavior without a failing test first.
- Do not weaken assertions to get green.
- Keep each loop focused on one observable behavior.
