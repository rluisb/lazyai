# Context: Bug Fixes

**Category:** Process
**Status:** Active

---

## Rule

Follow the standard development workflow for all bug fixes.

## Rationale

Ensures consistency, quality, and traceability across the codebase.

## Guidelines

- **Research:** Scout agent explores the codebase and existing patterns.
- **Plan:** Planner agent creates a phased task list with dependencies.
- **Implement:** Builder agent implements one task at a time, writing tests first.
- **Review:** Reviewer agent verifies the implementation against the spec.
- **Merge:** Human review and merge.

## Enforcement

- PR review checklist
- CI gates for tests and linting
