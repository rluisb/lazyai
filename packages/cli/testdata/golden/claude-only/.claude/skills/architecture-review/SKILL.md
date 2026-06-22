---
name: architecture-review
description: Use before structural changes to make a lightweight ADR-style decision with constraints, trade-offs, and consequences.
---

# Architecture Review

## When to Use

Use this skill before changes that affect module boundaries, public contracts, dependencies, data flow, storage, or cross-CLI behavior.

Skip it for local edits that do not change structure or caller expectations.

## Review Checkpoint

Before implementation, write a short decision note:

```md
## Decision
<one sentence>

## Context
- <constraint or fact>
- <constraint or fact>

## Options
1. <option> — trade-off
2. <option> — trade-off

## Outcome
Chosen: <option>. Reason: <why this wins now>.
Consequences: <costs, risks, follow-up watchouts>.
```

## Decision Rules

- Prefer the existing pattern when it satisfies the task.
- Prefer a markdown file or short script over a framework.
- Reject speculative abstractions that are not required by current callers.
- Treat reversibility as a first-class constraint.
- Consider what each option makes easier or harder to delete later.

## Anti-Overengineering Check

Before choosing the more complex option, answer:

1. Which current caller or constraint requires this extra structure now?
2. Would a simpler concrete solution satisfy the current behavior?
3. Is there a real second consumer, boundary, or test seam today?
4. If we duplicated this once, would that be cheaper and clearer right now?
5. What becomes harder to delete if we add this?

If the argument depends mostly on possible future use, simplify the design.


## Recording

If the decision will matter after the task, record it in `docs/adr/` or the nearest durable project doc.

If it is only a local implementation choice, keep the note in the task response or PR description.

## Constraints

- Do not create ADR ceremony for trivial edits.
- Do not introduce dependencies without naming the operational cost.
- Do not change architecture to solve a one-off symptom.
