---
name: anti-speculation
description: Prevent scope creep and speculative implementation.
trigger: /anti-speculation
phase: implement
---
## Quick Reference

| | |
|---|---|
| **Use when** | [When to use this skill] |
| **Do not use when** | [When NOT to use this skill] |
| **Primary agent** | [Which agent uses this] |
| **Runtime risk** | [Low/Medium/High] |
| **Outputs** | [What this skill produces] |
| **Validation** | [How to validate output] |
| **Deep mode trigger** | [How to trigger full mode] |



# Anti-Speculation

Detect and prevent speculative implementation — code that goes beyond what the spec or task requires. This is the enforcement arm of the YAGNI principle.

## The 5 Speculation Patterns

### Pattern 1 — Feature Creep
Adding behaviors not listed in the spec or task description.

**Detection**: Implementation includes endpoints, UI elements, or business logic with no corresponding entry in the spec's scope or behaviors section.

### Pattern 2 — Premature Abstraction
Creating generic frameworks, base classes, or plugin systems for a single use case.

**Detection**: Abstract classes, strategy patterns, factory methods, or configuration-driven behavior where the spec describes exactly one concrete behavior.

### Pattern 3 — Future-Proofing
Adding parameters, flags, or extension points "for later."

**Detection**: Optional parameters not in the data contract. Feature flags for features not in scope. Database columns not referenced in any behavior.

### Pattern 4 — Gold Plating
Over-engineering error handling, logging, or monitoring beyond spec requirements.

**Detection**: Retry logic, circuit breakers, or caching where the spec doesn't mention performance or reliability concerns for that path.

### Pattern 5 — Scope Leakage
Touching files or services not listed in the task scope.

**Detection**: Diff includes files in services not assigned to the current task. Changes to shared libraries not mentioned in the plan.

## YAGNI Checkpoint

Before writing any production code, answer these four questions:

1. **Is this behavior in the spec?** → If NO, don't build it.
2. **Is this the simplest way to satisfy the spec?** → If NO, simplify.
3. **Would removing this code cause a spec'd test to fail?** → If NO, remove it.
4. **Does this belong in the current task's scope?** → If NO, flag it.

## Detection Heuristics

When reviewing implementation against spec:

- [ ] Every new public function/method maps to a spec behavior
- [ ] Every new database column maps to a data contract field
- [ ] Every new endpoint maps to a spec behavior
- [ ] No abstract base classes for single implementations
- [ ] No optional parameters beyond the data contract
- [ ] Diff stays within the scope of the task

## Halt Protocol

When speculation is detected:

1. **IDENTIFY** the speculation pattern (1–5)
2. **QUOTE** the spec section that should govern this code
3. **SHOW** the speculative code or design
4. **SUGGEST** the minimal spec-compliant alternative
5. **ASK** the user:
   - (A) Remove this and implement minimally
   - (B) Amend the spec to include this behavior
   - (C) Keep with explicit approval (noted in plan)

### Halt Output Format

```
## ⚠️ Speculation Detected — Pattern {N}: {Name}

**Spec says**: "{quote from spec}"
**Implementation adds**: "{description of speculative code}"

**Minimal alternative**: {what spec-compliant code would look like}

Options:
- (A) Remove and implement minimally
- (B) Amend the spec to include this behavior
- (C) Keep with explicit approval
```

## Integration

- **Builder agent**: Runs YAGNI checkpoint before writing production code
- **Reviewer agent**: Runs detection heuristics during code review
- **Planner agent**: References this skill when scoping tasks
- **Workflow rule**: Cross-references anti-speculation for scope enforcement
