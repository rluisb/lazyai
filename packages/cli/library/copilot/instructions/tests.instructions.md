---
applyTo: "**/*_test.*,**/*.test.*,**/*.spec.*"
description: Testing best practices and conventions for this codebase
---

# Testing Guidelines

## Test structure (Arrange-Act-Assert)

- **Arrange:** Set up test fixtures, mocks, and state.
- **Act:** Invoke the code under test.
- **Assert:** Check that the result matches expectations. One logical assertion per test or a small group of related assertions.

## Test naming and organization

- Test function names describe the scenario and expected outcome: `TestUserUpdate_ValidInput_UpdatesUserAndReturnsSuccess`.
- Use descriptive subtests: `t.Run("user not found", func(t *testing.T) { ... })`.
- Group related tests in the same file with clear section comments.
- One test file per source file as a baseline; split if the file gets >500 lines of tests.

## Fixtures and data

- Use factories or builders for complex test objects (`NewTestUser()`, `builder.NewUser().WithEmail(...)`).
- Avoid test data singletons that are shared across tests — they hide dependencies.
- Seed known, minimal data; don't create a "realistic" dataset if a single row suffices.
- Document non-obvious test setup in a comment.

## Determinism and isolation

- Tests must pass consistently — no flakes. Avoid `time.Sleep()`, randomness, or real I/O.
- Use `t.TempDir()` for file operations; clean-up is automatic.
- Use `t.Setenv()` for environment variables; restore is automatic.
- Mock external calls (databases, APIs, filesystems). Make mocks deterministic and injectable.

## Assertions and error checking

- Prefer clear assertion libraries (e.g. `assert` from testify) for readability.
- Check both positive (success) and negative (error) paths.
- Use subtests to organize multiple scenarios.
- Never silently ignore errors — `t.Errorf` or `t.Fatalf` on unexpected failures.

## Coverage

- Aim for >80% coverage on public APIs and critical paths.
- Don't obsess over 100% coverage — some code is untestable (rarely).
- Test edge cases: empty inputs, nil pointers, boundaries, concurrency.
- Integration tests can be slower; keep unit tests fast.
