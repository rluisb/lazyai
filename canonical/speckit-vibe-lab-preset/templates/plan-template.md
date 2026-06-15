# Plan: {{plan_name}}

## WHAT

One-sentence goal. Link to the spec this plan serves.

## HOW

### Simplest viable approach

Describe the minimal change that satisfies the spec.

### Existing pattern review

List existing code, conventions, or tests this plan must reuse or preserve.

### Rejected alternatives

| Alternative | Why rejected |
|-------------|------------|
| (name) | (concrete reason) |

### Boundaries

- In scope:
- Out of scope:
- Deferred V2:

### Risk

- What can break:
- How this plan avoids unnecessary dependency, runtime state, hook, or generated adapter changes:
  - If any are required, name the spec requirement that forces them.

## DON'T WANT

- Complex abstractions for simple changes.
- Unverified integration claims.
- Changes outside the stated boundaries.

## VALIDATE

- [ ] Plan has an observable verification path (test, command, or diff).
- [ ] Existing tests are preserved or updated.
- [ ] Risk is listed and mitigated.
- [ ] Cleanup/docs are gated on smoke verification, not done first.
