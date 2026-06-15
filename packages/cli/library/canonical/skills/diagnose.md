---
name: diagnose
description: Reproduce the failure, rank hypotheses, instrument the right seam, and add a regression test for the fix.
trigger: /diagnose
tier: frontier
thinking: high
risk: 4
---

# Diagnose

## Workflow

1. Build a deterministic feedback loop.
2. Reproduce the reported failure.
3. Rank plausible root-cause hypotheses.
4. Instrument one hypothesis at a time.
5. Add the regression test at the real seam.
6. Fix the cause, remove debug traces, and rerun verification.

## Guardrails

- No guess-and-check without a loop.
- No workaround shipped as the final fix.
- Capture the failure cause and the regression evidence in the report.
