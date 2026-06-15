# PRD → Plan → Todo (Non-Spec Kit)

> Lightweight requirements flow for projects that are not using Spec Kit.

## When to Use

Use this when the work is too large or ambiguous for direct implementation, but adding a full Spec Kit packet would be heavier than the task needs.

Do not use this for:

- small direct edits with obvious validation;
- speculative roadmap ideas with no implementation intent.

## Escalation Rule

Escalate to Spec Kit when the work crosses subsystem boundaries, changes public contracts, needs multi-session implementation, or requires formal traceability.
## The Four Points

- **WHAT** — The product or engineering outcome in plain language.
- **HOW** — Broad approach, key interfaces, TDD mode, and sequencing.
- **What I DON'T want** — Non-goals, constraints, rejected approaches, and known traps.
- **How we VALIDATE** — The command, test, manual scenario, or observable signal that proves the work.

## Artifacts

| Artifact | Purpose | Owner |
|----------|---------|-------|
| PRD | Decide what should exist and why. | Human + agent pair |
| Plan | Decide how to build it without violating constraints. | Agent proposes, human can correct |
| Todo | Execute the plan in dependency order. | Agent updates while working |

A single markdown file is enough unless the work crosses subsystem boundaries.

Recommended location:

```text
docs/prd/<slug>.md
```

For project-local private planning, use:

```text
.vibe-lab/tasks/<slug>.md
```

## 1. PRD

The PRD is a decision packet, not a sales document.

```markdown
# PRD: <Outcome>

## Problem

<What hurts today. Include concrete examples.>

## Purpose

<Why this run exists and what decision or shipped outcome it serves.>

## Goal

<What should be true after this ships.>

## Users / Callers

- <Human, agent, CLI, API caller, or subsystem affected.>

## Non-Goals

- <What this explicitly will not solve.>

## Constraints

- <Compatibility, security, performance, UX, dependency, or repo constraints.>

## Behavior Scenario

- Given <initial state>
- When <action>
- Then <observable outcome>

Use `n/a` only when the work is deliberately docs-only or read-only.

## Acceptance Criteria

- [ ] <Observable behavior.>
- [ ] <Error or edge case behavior.>
- [ ] <Regression prevention.>

## Validation

- Command: `<test / lint / build / manual scenario>`
- Expected signal: <what green looks like>
```

Rules:

1. Acceptance criteria must be observable.
2. Non-goals must be explicit when scope could expand.
3. Validation must name a command or manual scenario.
4. If any of the Four Points is missing, clarify before planning.

## 2. Plan

The plan translates the PRD into implementation boundaries.

```markdown
## Plan

### Current State

<Relevant files, APIs, invariants, and existing patterns.>

### Approach

<Smallest design that satisfies the PRD.>

### Alternative Considered

<One viable option rejected and why.>

### TDD Mode

- Mode: <lightweight | medium | heavy-aggressive | required>
- Red check: <test or explicit exemption>
- Green check: <focused verification command>

### Touch Points

- `<file or module>` — <why it changes>

### Risks

- <What can break and how the plan avoids it.>
```

Rules:

1. Reuse existing patterns; parallel conventions are not allowed.
2. Name touch points before editing.
3. Choose a TDD mode for behavior-affecting changes.
4. Keep the plan boring; do not add abstractions because they might be useful later.

## 3. Todo

The todo list is the execution queue, not a brainstorm.

```markdown
## Todo

- [ ] Research existing pattern and constraints.
- [ ] Add or update the red check.
- [ ] Implement the smallest behavior change.
- [ ] Run focused verification.
- [ ] Update directly affected docs/adapters after behavior works.

## Exit Gate

- [ ] Purpose is fulfilled or explicitly blocked.
- [ ] Validation evidence is observed.
- [ ] Out-of-scope discoveries are called out instead of silently absorbed.
```

Rules:

1. Order todos by dependency, not by convenience.
2. Every implementation todo must have an associated proof path.
3. Do not pre-allocate cleanup work before the behavior has been smoke-tested.
4. After smoke test passes, add final cleanup tasks for docs, generated adapters, and obsolete scaffolding if needed.
5. Marking a todo complete is a transition, not a stopping point.

## Failure Gates

Stop and revise the PRD/Plan when:

- acceptance criteria conflict;
- validation cannot be run or observed;
- the plan requires a new dependency not justified by the PRD;
- implementation reveals a different root cause;
- a test must be deleted, weakened, or skipped.

## Handoff Shape

If the work spans sessions:

```markdown
## Status

<done | in-progress | blocked>

## Decisions

- <Decision and reason.>

## Open Questions

- <Question or explicit assumption.>

## Next Actions

1. <Concrete next command/edit.>
2. <Concrete verification step.>
```

No placeholders. No fake progress. No "MVP" relabeling for unfinished work.
