# Testing Rule

## Rule

All production code must be accompanied by tests. Tests must be written before or alongside production code, never deferred.

## Rationale

Test-first development catches defects earlier, drives better design, and provides living documentation of intended behavior. Untested code is a liability.

## Guidelines

### Test-First Mandate
- Every behavior in a task must have a corresponding test
- Tests must exist before or in the same commit as production code
- No production code commit should exist without a corresponding test

### Coverage Expectations
- New code: 100% of specified behaviors covered
- Bug fixes: regression test required for every fix
- Refactors: existing tests must continue to pass

### Test Quality Standards
- Tests should verify behavior, not implementation details
- Each test should have a clear, descriptive name
- Tests must be deterministic — no flaky tests allowed
- Prefer integration tests at boundaries, unit tests for pure logic

### Git-Diff Verification
After implementation, verify test ordering:
- Test files appear in git history at or before the production code commit
- Every new public method/endpoint has a corresponding test
- No test file is added in a later, separate "add tests" commit

### Task File Requirements
Every task must declare:
- **Acceptance Test**: description of what the test verifies
- **Test File**: expected path to test file
- **Order**: TEST → IMPLEMENT → VERIFY

## Enforcement

- Code review checklist includes test verification
- CI pipeline runs full test suite on every PR
- Reviewer agent checks test-first ordering via git history
- TDD-loop skill provides the process; this rule provides the mandate
