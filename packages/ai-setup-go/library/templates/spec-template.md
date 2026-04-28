# Spec: [###-feature-slug]

**Feature ID:** ###
**Feature name:** [feature-slug]
**Date:** YYYY-MM-DD
**Status:** Draft | Clarified | Approved | Implemented
**Owner:** [name]
**Constitution:** [link to active `constitution.md`]

> **Purpose.** A specification describes *what* the system should do and *why*, not *how*. Tech stack and architecture belong in `plan.md`. Tasks belong in `tasks.md`. This document is the contract every downstream artifact is judged against.

---

## User Scenarios

User scenarios are written as prioritized stories. P1 is the smallest end-to-end slice that delivers user value; P2 and P3 add scope only after P1 is shippable.

> Each story MUST be **independently testable** — there is at least one acceptance test that exercises only that story.

### P1 — [Primary user story title]
**As a** [user role]
**I want** [capability]
**So that** [outcome]

**Acceptance criteria**
- [ ] Given [precondition], when [action], then [observable outcome].
- [ ] Given [precondition], when [action], then [observable outcome].

### P2 — [Secondary story title]
**As a** [user role]
**I want** [capability]
**So that** [outcome]

**Acceptance criteria**
- [ ] [Criterion]
- [ ] [Criterion]

### P3 — [Tertiary story title — only if needed]
*(Optional. Delete if not applicable.)*

---

## Functional Requirements

Each requirement has a stable ID (`FR-NNN`) so plans, tasks, and tests can reference it.

| ID | Requirement | Priority | Source story |
|---|---|---|---|
| FR-001 | The system MUST [observable behavior]. | P1 | P1 |
| FR-002 | The system MUST [observable behavior]. | P1 | P1 |
| FR-003 | The system SHOULD [observable behavior]. | P2 | P2 |

> Use **MUST** for non-negotiable behavior, **SHOULD** for strongly preferred, **MAY** for optional. Avoid "should be able to" — write the observable outcome.

---

## Key Entities

Domain objects this feature introduces or changes. Names only — schema lives in `plan.md`.

| Entity | Description | Lifecycle |
|---|---|---|
| [Entity] | [one sentence] | created on … / retired on … |

*(Delete this section if no new entities.)*

---

## Success Criteria

How we know the feature works in production, not just in tests. Each criterion is **measurable** and **observable**.

- **SC-001 — [name]:** [precise, measurable statement]. Measured by: [metric / log / dashboard].
- **SC-002 — [name]:** [statement]. Measured by: [source].

> Success criteria differ from acceptance criteria: ACs are local checks; SCs are post-deploy reality checks tied to Gate 5 (Observability Readiness).

---

## Edge Cases

Boundary conditions, error scenarios, concurrency, and adversarial inputs. Each edge case names the expected behavior — not just the trigger.

- **EC-001 — [name]:** When [trigger], the system [expected behavior].
- **EC-002 — [name]:** When [trigger], the system [expected behavior].

> Common categories to consider: empty input, concurrent writers, network failure, auth boundary, large payloads, time skew, retries.

---

## Assumptions

Things believed to be true that, if false, invalidate the spec. List them so they can be challenged in `speckit-clarify` and `speckit-analyze`.

- **A-001:** [statement] — [confidence: HIGH / MEDIUM / LOW].
- **A-002:** [statement] — [confidence: HIGH / MEDIUM / LOW].

> Low-confidence assumptions MUST be resolved (verified or re-stated as risks) before this spec is approved.

---

## Out of Scope

What this spec deliberately does **not** cover. Anti-Speculation (Article IV) lives here.

- [Item explicitly out of scope]
- [Item deferred to a later spec]

---

## Clarifications

Filled in by `speckit-clarify`. Each entry records a question, the answer, and who decided.

| Date | Question | Answer | Decided by |
|---|---|---|---|
| YYYY-MM-DD | [question] | [answer] | [human / agent + human approval] |

---

## Constitutional Notes

How this spec aligns with the constitution. Cite article numbers explicitly.

- **Article I — Library-First:** [which existing libraries this spec relies on].
- **Article IV — YAGNI:** [what was deliberately not added].
- **Article V — Simplicity:** [the simpler alternative considered and why it was rejected, if any].

---

## Downstream Contract

| Produced for | Filename |
|---|---|
| `speckit-clarify` | this file (in-place updates) |
| `speckit-plan` | this file |
| `speckit-tasks` | indirectly via plan |
