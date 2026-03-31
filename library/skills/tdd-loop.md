# TDD Loop

## Purpose

Guide implementation through disciplined Test-Driven Development red-green-refactor cycles. Production code must be driven by failing tests first.

## The TDD Cycle

```
RED → GREEN → REFACTOR → repeat
```

1. **RED**: Write one failing test for one behavior
2. **GREEN**: Write minimal code to pass that test
3. **REFACTOR**: Improve design with tests still green

## RED Phase — Write Failing Test

1. Identify the next smallest behavior from the task
2. Write exactly one test for that behavior
3. Run only that test
4. Verify it fails for the **expected reason** (not a syntax error or import issue)

### RED Phase Commands

| Stack | Command Pattern |
|-------|----------------|
| *Your stack* | *Fill: single-test run command* |

> Customize this table per your project's test runner and conventions.

## GREEN Phase — Make It Pass

1. Write the **minimal** code to satisfy the failing test
2. No speculative refactors in this phase
3. Re-run the same targeted test
4. Confirm it passes

## REFACTOR Phase — Improve Design Safely

1. Remove duplication, improve naming, extract helpers
2. Keep behavior unchanged — no new features
3. Re-run targeted tests frequently during refactor
4. Run full quality gates before ending the cycle

## Quality Gates

Run at the end of each TDD cycle (after REFACTOR):

| Gate | Command |
|------|---------|
| Linter | *Fill: your lint command* |
| Type checker | *Fill: your typecheck command* |
| Test suite | *Fill: your full test command* |
| Build | *Fill: your build command* |

> Customize per your project. All gates must pass before starting the next cycle.

## Test Strategy

### Unit Tests
Pure functions, serializers, mappers, small components. Fast, isolated, no external dependencies.

### Integration Tests
API/database/service boundaries. Multi-component interactions. May use test databases or fixtures.

### Contract Tests
External boundaries: payments, webhooks, public API response shapes. Validate both success and failure paths.

## Mocks and Real Services

- Prefer realistic integration boundaries when cost is acceptable
- Mocks are allowed for isolated unit tests
- Never use mocks to hide broken production behavior
- For external integrations, validate success AND failure paths explicitly

## Iteration Limits

- Maximum **10 TDD iterations** per task before escalating
- If stuck after 10 iterations, halt and report:
  - What was attempted
  - Current failing test output
  - Likely root cause
  - Recommended next path

## Anti-Patterns

- Writing production code before a failing test
- Writing multiple behaviors in one test iteration
- Commenting out failing tests to pass CI
- Skipping the REFACTOR phase entirely
- Running only happy-path tests
- Ignoring project quality gates

## Integration

- **Builder agent**: Primary executor — follows RED→GREEN→REFACTOR per behavior
- **Reviewer agent**: Verifies TDD compliance during review
- **Testing rule**: Cross-references for test-first enforcement
- **Iterate skill**: Complements TDD — iterate is fix-only (existing tests), TDD is create-new (write tests)
