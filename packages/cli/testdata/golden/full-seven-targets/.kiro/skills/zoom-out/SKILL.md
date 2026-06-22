---
name: zoom-out
description: Use when stuck in implementation details and losing sight of the bigger picture, or when a bug suggests an architectural problem rather than a local fix. Forces a structured step back to re-evaluate assumptions and design.
---

# Zoom Out

## When to Use

Use this skill when:
- You have tried 3+ hypotheses and none explain the bug (suggests architectural issue)
- The code you are touching keeps requiring more and more hacks to work
- You are about to add a workaround that violates the project constraints
- A fix in one place breaks something in an unrelated area
- The user says "step back" or "look at the bigger picture"
- You catch yourself optimizing or refactoring before understanding the actual requirement

Do not use for trivial bugs with clear local causes. Do not use when the solution is already evident.

## Rule

When implementation-level debugging fails, the problem is often in the assumptions, not the code. Zoom out to the architecture, the requirements, or the constraints.

## Workflow

1. **Acknowledge the stall**
   - State clearly: "Local investigation is not resolving this. I am zooming out."
   - List the failed hypotheses and why they were rejected
   - Define what "zoomed out" means for this specific case (architecture, requirements, constraints, or data flow)

2. **Re-read the spec or requirement**
   - What is the actual goal of this feature/component?
   - What constraints must hold (performance, security, compatibility)?
   - What has changed since the original design was laid out?

3. **Map the system boundary**
   - Draw or list the components this code interacts with
   - Identify the data flow in and out
   - Note which contracts (APIs, schemas, protocols) are involved
   - Look for mismatches between what the code assumes and what the system provides

4. **Challenge assumptions**
   - What did you assume about the input data? Is it actually true?
   - What did you assume about the execution environment?
   - What did you assume about other components behavior?
   - What did you assume about the user intent?

5. **Form a system-level hypothesis**
   - The new hypothesis must explain all the local failures
   - It must be testable without touching the problematic code (e.g., check logs, verify contracts, inspect data)

6. **Verify at the boundary**
   - Test the system-level hypothesis by checking inputs, contracts, or configurations
   - Do not modify the implementation yet — verify the diagnosis first

7. **Decide: refactor, redesign, or patch**
   - If the architecture is wrong: propose a redesign (use `architecture-review`)
   - If the requirement is wrong: clarify with the user (use `doc-backed-clarify`)
   - If the assumption was wrong: fix the local code with the correct understanding
   - If the fix is a workaround that violates constraints: reject it (use `no-workarounds`)

## Constraints

- Zooming out is not an excuse to avoid reading code. You must have attempted local diagnosis first.
- Do not propose sweeping rewrites without a specific architectural diagnosis.
- If the bug is intermittent, check system-level causes (timing, resource limits, external state) before blaming the implementation.
- Always return to the original problem statement after zooming out. Do not solve a different problem.

## Verification Checklist

- [ ] 3+ local hypotheses were attempted and documented
- [ ] Assumptions about inputs, environment, and contracts are explicitly listed
- [ ] System-level hypothesis explains all observed local failures
- [ ] Hypothesis was verified before any code change
- [ ] Decision (refactor/redesign/patch) is justified against project constraints

## Related Skills

- `diagnose` — For local hypothesis-driven debugging before zooming out
- `architecture-review` — For redesign proposals after zoom-out diagnosis
- `no-workarounds` — For rejecting patch-level fixes that need architectural solutions
- `doc-backed-clarify` — For requirement clarification when the spec is ambiguous
