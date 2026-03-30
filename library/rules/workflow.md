# Rule: Workflow

**Category:** Process
**Status:** Active

---

## Rule

Follow the standard development workflow for all changes.

## Rationale

Ensures consistency, quality, and traceability across the codebase.

## Workflow

1. **Research:** Scout agent explores the codebase and existing patterns.
2. **Plan:** Planner agent creates a phased task list with dependencies.
3. **Implement:** Builder agent implements one task at a time, writing tests first.
4. **Review:** Reviewer agent verifies the implementation against the spec.
5. **Merge:** Human review and merge.

## Enforcement

- PR review checklist
- CI gates for tests and linting
