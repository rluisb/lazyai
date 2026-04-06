<template>
  <name>PRD — Product Requirements Document</name>
  <output>specs/features/NNN-feature-name/prd.md</output>
  <input>Research artifact + human feature request</input>
  <phase>Plan — Step 1</phase>
</template>

# PRD: [Feature Name]

**Feature:** NNN-feature-name
**Date:** YYYY-MM-DD
**Status:** Draft | Approved | Implemented
**Author:** [name]
**Research:** [link to research.md]

---

## Problem Statement

<!-- One paragraph. What pain exists today. Who feels it. Why it matters now. -->

[PROBLEM]

## Goals

<!-- Measurable outcomes. If you can't measure it, rephrase it. -->

- G1: [measurable goal]
- G2: [measurable goal]
- G3: [measurable goal]

## Non-Goals (Out of Scope)

<!-- MANDATORY. What we are explicitly NOT doing. This section prevents scope creep.
     If you can't list at least 2 non-goals, you haven't scoped tightly enough.
     YAGNI: if it's not needed for P1, it's a non-goal. -->

| What | Why excluded |
|------|-------------|
| [thing we won't do] | [reason] |
| [thing we won't do] | [reason] |

## Existing Landscape

<!-- MANDATORY. What already exists that relates to this feature.
     Pulled from research.md. Prevents building what you already have.
     Respect existing patterns — don't reinvent. -->

- Similar feature/module: [what exists, where, how it partially solves this]
- Reusable components: [what can be reused from the codebase]
- Nothing exists: [state this explicitly if true]

## User Stories

<!-- Ordered by priority. P1 = must ship. P2 = should ship. P3 = nice to have.
     Each story must be independently testable and deliverable.
     A shipped P1 with no P2/P3 is still a valid release. -->

### US-1: [Title] (P1) ⭐ MVP

**As a** [user type], **I want** [action] **so that** [outcome].

**Acceptance Criteria:**
1. GIVEN [context] WHEN [action] THEN [result]
2. GIVEN [context] WHEN [action] THEN [result]

### US-2: [Title] (P2)

**As a** [user type], **I want** [action] **so that** [outcome].

**Acceptance Criteria:**
1. GIVEN [context] WHEN [action] THEN [result]

### US-3: [Title] (P3)

**As a** [user type], **I want** [action] **so that** [outcome].

**Acceptance Criteria:**
1. GIVEN [context] WHEN [action] THEN [result]

## Functional Requirements

<!-- Numbered, testable, unambiguous. Use MUST/SHOULD/MAY (RFC 2119).
     Flag uncertainty: [NEEDS CLARIFICATION: specific question] — max 3 allowed. -->

- **FR-001:** System MUST [capability]
- **FR-002:** System MUST [capability]
- **FR-003:** System SHOULD [capability]
- **FR-004:** [NEEDS CLARIFICATION: question about scope or behavior]

## Constraints

<!-- Technical, business, or regulatory limits that shape the solution.
     Keep it real — only constraints that actually exist, not imagined ones. -->

- [constraint]
- [constraint]

## Success Criteria

<!-- How we know this worked. Measurable. Not "users are happy." -->

- SC-1: [metric]
- SC-2: [metric]

## Open Questions

<!-- Questions that need human answers before or during implementation. -->

1. [question for PM/stakeholder]
2. [question for team]

---

<!-- PRINCIPLES CHECK — verify before approving:
- [ ] Problem is real and clearly stated
- [ ] Goals are measurable
- [ ] Non-goals section has ≥2 items (YAGNI enforced)
- [ ] Existing landscape checked — not reinventing
- [ ] P1 stories are independently shippable as MVP
- [ ] No HOW or technical decisions (that's techspec's job)
- [ ] [NEEDS CLARIFICATION] count ≤ 3
- [ ] Scope is the smallest thing that delivers value
-->
