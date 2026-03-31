<template>
  <name>TechSpec — Technical Specification</name>
  <output>docs/features/NNN-feature-name/techspec.md</output>
  <input>PRD + Research + Codebase + Project Rules + Standards</input>
  <phase>Plan — Step 2</phase>
</template>

# TechSpec: [Feature Name]

**Feature:** NNN-feature-name
**Date:** YYYY-MM-DD
**Status:** Draft | Approved | Implemented
**Author:** [name]
**PRD:** [link to prd.md]
**Research:** [link to research.md]

---

## Simplicity Gate

<!-- MANDATORY. Answer ALL before writing anything else.
     If any answer is YES, use the simpler path. Do one thing well. -->

| Question | Answer |
|----------|--------|
| Can this be done without a new service/package? | [YES → do it in-place / NO → justify] |
| Can this be done without a new database table? | [YES → use existing / NO → justify] |
| Can this be done without a new dependency? | [YES → use stdlib or existing / NO → justify] |
| Is there an existing pattern in the codebase? | [YES → follow it, cite the file / NO → document new pattern] |
| Does this fit inside an existing module? | [YES → extend, don't create / NO → justify new module] |
| What is the simplest version that satisfies P1? | [describe it] |

## Summary

<!-- 2-3 sentences. What this techspec delivers and the chosen approach. -->

[SUMMARY]

## Approach Options

<!-- MANDATORY. Explore at least 2 approaches before committing.
     This is Tree of Thoughts — branch, evaluate, prune.
     The rejected option feeds the ADR's "Alternatives Considered." -->

### Option A: [name]
- **How:** [brief description]
- **Pros:** [list]
- **Cons:** [list]
- **Complexity:** Low | Medium | High

### Option B: [name]
- **How:** [brief description]
- **Pros:** [list]
- **Cons:** [list]
- **Complexity:** Low | Medium | High

### Decision: Option [X]
**Why:** [reason — reference simplicity, YAGNI, existing patterns, project standards]

## Architecture

<!-- How the pieces fit together. ASCII diagrams.
     Box-and-arrow level. No code. -->

```
[ASCII diagram of components and data flow]
```

### Components

| Component | Responsibility | Location |
|-----------|---------------|----------|
| [name] | [one sentence] | [file/module path] |
| [name] | [one sentence] | [path] |

## Patterns to Follow

<!-- MANDATORY. Reference actual files in the codebase that this feature
     should mirror. The Builder reads these as the "how to write it" guide.
     Reference standards docs where they exist.
     If no pattern exists: "New pattern — document in standards after implementation." -->

| Pattern | Standard | Reference File | Apply To |
|---------|----------|---------------|----------|
| [e.g. API endpoint] | [docs/standards/api-patterns.md] | [src/modules/auth/auth.controller.ts] | [new endpoints] |
| [e.g. Service layer] | [docs/standards/service-patterns.md] | [src/modules/auth/auth.service.ts] | [new services] |
| [e.g. Error handling] | [docs/standards/error-handling.md] | [src/shared/errors/app-error.ts] | [all error cases] |
| [e.g. Test structure] | [docs/standards/test-patterns.md] | [src/modules/auth/__tests__/] | [new tests] |

## Data Model

<!-- Only if the feature involves data. Delete this section entirely if not applicable.
     YAGNI: only fields needed for P1. No speculative fields. -->

### Entities

| Entity | Key Fields | Notes |
|--------|-----------|-------|
| [name] | [fields] | [relationships, constraints] |

### Migrations Required

- [ ] [migration description — or "None"]

## API Contracts

<!-- Only if the feature exposes or consumes APIs. Delete if not applicable. -->

| Method | Endpoint | Input | Output | Notes |
|--------|----------|-------|--------|-------|
| [verb] | [path] | [shape] | [shape] | [auth, rate limit] |

## Dependencies

<!-- New dependencies this feature introduces. Justify each one.
     YAGNI: if you can do it without the dependency, do it without. -->

| Dependency | Why Needed | Alternative Rejected |
|------------|-----------|---------------------|
| [name] | [reason] | [what we considered] |
| None | — | — |

## Conformance Check

<!-- Map decisions against project rules and standards. Flag deviations. -->

- [ ] Follows code style (docs/rules/code-style.md)
- [ ] Follows testing rules (docs/rules/testing.md)
- [ ] Respects path access (docs/rules/access.md)
- [ ] No security violations (docs/rules/security.md)
- [ ] Follows existing patterns (docs/standards/)
- **Deviations:** [list with justification — or "None"]

## Test Strategy

| Level | What | How |
|-------|------|-----|
| Unit | [what gets unit tested] | [approach] |
| Integration | [what gets integration tested] | [approach] |
| E2E | [if applicable] | [approach] |

## Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| [risk] | Low/Med/High | Low/Med/High | [how we handle it] |

## ADRs

<!-- Architectural decisions worth recording permanently.
     If Approach Options above had a non-obvious choice, it needs an ADR. -->

- [ ] ADR needed: [decision] → docs/adrs/NNN-title.md
- [ ] No ADR needed for this feature

## Unknowns

<!-- Max 3 [NEEDS CLARIFICATION]. Everything else: simplest assumption, documented. -->

- [NEEDS CLARIFICATION: specific technical question]
- **Assumption:** [what we assumed and why]

---

<!-- PRINCIPLES CHECK — verify before approving:
- [ ] Simplicity gate passed — no unnecessary complexity
- [ ] At least 2 approaches explored (Tree of Thoughts)
- [ ] Decision justified with simplicity/YAGNI/patterns rationale
- [ ] Architecture uses existing patterns where possible
- [ ] Patterns to Follow section references real files
- [ ] No new dependencies without justification
- [ ] Data model is minimal — only P1 fields (YAGNI)
- [ ] Conformance check passed against project rules
- [ ] Test strategy covers P1 acceptance criteria
- [ ] Risks are honest, not optimistic
-->
