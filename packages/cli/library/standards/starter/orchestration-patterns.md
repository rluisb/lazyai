# Standard: Orchestration Patterns

**Category:** Process
**Scope:** project
**Date:** 2026-05-01
**Owner:** AI Setup Maintainers
**Status:** Active
**Constitution article(s):** I, III, IV, V

> **Purpose.** Keep agent workflows explicit, bounded, and reviewable so multi-step work can be traced from plan to implementation without inventing a new orchestration model for each task.

---

## Scope Cascade

This starter standard is written at project scope and applies to generated workflows in this repository until a more specific workspace or project standard replaces it.

```
global standards
        ↓
workspace standards
        ↓
project standards
        ↓
code and workflow definitions
```

**Where this standard lives:** project

**Overrides:** none — this is starter guidance for new projects.

---

## Rule

Orchestrated work MUST use explicit sequential steps, named gates, and documented handoff artifacts unless an approved ADR introduces a different execution model.

**Trigger:** Creating or changing an agent chain, team workflow, approval gate, or handoff step.

---

## Rationale

Explicit orchestration makes Quality Gate 4 Pattern Consistency reviewable: reviewers can cite the workflow shape and confirm that agents follow the approved path.

- **Article support:** Article III keeps workflow docs as source of truth; Article IV prevents speculative workflow primitives; Article V prefers simple sequential paths.
- **Origin:** Wave 1 starter standards for concrete pattern review.

---

## Examples

**Compliant:**
```
research -> plan -> plan-quality -> approval-gate -> implement -> review

Each step has a named owner, expected artifact, and success/failure transition.
```

**Non-compliant:**
```
Run research, planning, validation, and implementation in an implicit background loop.
```

**Why the non-compliant case fails:** Reviewers cannot identify the gate, owner, artifact, or rollback point.

---

## Enforcement

| Mechanism | Where | When |
|---|---|---|
| Workflow shape review | `library/orchestration/` or project workflow files | Gate 4 Pattern Consistency |
| Task checklist | Feature task file or handoff | Before implementation starts |
| Human approval gate | Chain or PR review | Before high-risk mutation |

---

## Exceptions

Exceptions require an explicit note in the plan or PR explaining why a sequential workflow is insufficient and how the alternate path remains observable.

- Emergency operational fix — requires maintainer approval and a follow-up handoff entry.

---

## Workspace Awareness

| Repo | Override | Reason |
|---|---|---|
| All repos | none | Starter guidance applies until repo-specific orchestration standards exist. |

---

## Related

- **Supersedes:** none
- **Related ADR(s):** none
- **Related standards:** context-loading, agent-security

---

## Memory

- [ ] Standard published at the declared scope.
- [ ] Workflow deviations reference the task, PR, or handoff that approved the exception.
- [ ] Repeated orchestration exceptions are candidates for a dedicated ADR.
