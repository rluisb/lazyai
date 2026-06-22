---
name: doc-backed-clarify
description: Use at task intake when requirements or repository context are unclear. Supports lightweight, grill-me, and grill-me-with-docs clarification levels while always preserving the four-point pattern.
---

# Doc-Backed Clarify

## When to Use

Use this skill at the start of a task when any of the four points are missing, ambiguous, or likely answered by local documentation.

## Rule

Every clarification level MUST resolve or explicitly mark the four points: WHAT, HOW, DON'T WANT, VALIDATE.

## Clarification Levels

Use `canonical/clarification-levels.md` as the source of truth.

### lightweight

Use when the user supplied at least three points and risk is low.

- Resolve obvious gaps from repo docs.
- Ask at most one focused question.
- Proceed once the missing point is answered or documented as a safe assumption.

### grill-me

Use when the request is vague, risky, or internally inconsistent.

- Ask targeted questions grouped by WHAT, HOW, DON'T WANT, VALIDATE.
- Do not proceed until all four points have explicit answers or approved assumptions.
- Keep questions focused; do not turn clarification into scope expansion.

### grill-me-with-docs

Use when local docs likely contain constraints or the change crosses subsystem boundaries.

- Read relevant docs first.
- Cite the doc-backed fact or gap for each unresolved point.
- Ask only questions that docs cannot answer.

## Intake Check

Confirm the task has:

1. **WHAT** — the outcome in plain language.
2. **HOW** — broad direction, not implementation lock-in.
3. **DON'T WANT** — constraints, rejected paths, or prior failures.
4. **VALIDATE** — the command, test, or signal that proves completion.

If a missing point changes implementation materially and cannot be inferred from repo docs, ask before coding.

## Documentation Grounding

Before proposing code:

1. Search `docs/`, `canonical/`, and relevant `.agents/` artifacts for the task area.
2. Read only sections that affect the decision.
3. State the concrete doc-backed constraint that affects the next action.

## Refusal Pattern

```text
I need <missing point> before coding because <specific risk>.
```

Ask one focused question. Do not turn clarification into a questionnaire unless `grill-me` or `grill-me-with-docs` is selected.

## Verification Checklist

- [ ] WHAT is explicit.
- [ ] HOW is broad enough to allow a better implementation.
- [ ] DON'T WANT captures constraints and rejected paths.
- [ ] VALIDATE names a concrete proof.
- [ ] Documentation-derived constraints are cited when using grill-me-with-docs.
