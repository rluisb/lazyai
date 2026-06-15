---
name: diagnose
description: Use when debugging a failing test, broken build, runtime error, or unexpected system behavior. Drives hypothesis-based investigation, root-cause fixes, verification, and reusable learning capture.
---

# Diagnose

## When to Use

Use this skill when:
- A test fails and the reason is not obvious.
- A build breaks with a cryptic error.
- A runtime exception occurs.
- An integration returns unexpected results.
- The user says "this is broken" without a clear cause.

Do not use for trivial typos or one-line fixes where the error points directly at the solution.

## Rule

Debugging is hypothesis-driven investigation, not shotgun-fixing. Every proposed fix must trace back to a specific observation.

## Workflow

1. Reproduce the failure and capture exact error, environment, and steps.
2. State one hypothesis in one sentence.
3. Gather evidence from code, logs, tests, docs, or recent changes.
4. Test the hypothesis with the smallest experiment.
5. Fix the root cause, not the symptom.
6. Verify by rerunning the failing scenario and the smallest relevant broader check.
7. Capture reusable learning when the root cause is likely to recur.

## Diagnosis Template

```markdown
Failure: <exact symptom>
Environment: <version/config/context>
Hypothesis: <one sentence>
Evidence: <observed fact>
Root Cause: <specific cause>
Fix: <why this fix addresses that cause>
Verification: <command/scenario and result>
```

## Learning Capture

Use `canonical/learning-template.md` for reusable root causes, traps, environment gotchas, or diagnostic patterns.

Classify diagnosis learnings as:
- `trap`: false assumption or sharp edge.
- `pattern`: recurring debug workflow.
- `rule`: stable prevention rule.
- `template`: reusable diagnostic report shape.

Promote only through `memory-promotion` after approval.

## Constraints

- One hypothesis at a time.
- Read relevant code before guessing.
- No speculation in commit messages.
- Intermittent bugs need reproduction conditions.
- Environment details matter.

## Verification Checklist

- [ ] Failure reproduced or reproduction limit documented.
- [ ] Hypothesis stated before fixing.
- [ ] Evidence gathered from observed sources.
- [ ] Fix addresses root cause.
- [ ] Failing scenario rerun.
- [ ] Reusable learning captured or intentionally skipped.

## Related Skills

- `test-first-change` — drive fixes through failing tests.
- `issue-triage` — classify before debugging.
- `no-workarounds` — reject symptom patches.
- `memory-promotion` — promote durable diagnostic learning.
