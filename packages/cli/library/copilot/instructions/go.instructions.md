---
applyTo: "**/*.go"
description: Go conventions and best practices for this codebase
---

# Go Conventions

## Code style and formatting

- Use `go fmt` — non-negotiable. Code review will reject unformatted contributions.
- Use `gofmt` width of 100 chars (standard tab width).
- Variable names: prefer clarity over brevity. Use `count` not `c`, `user` not `u`.
- Package-level functions and types start with exported uppercase (`UserID`, `NewServer`).

## Error handling

- Always check errors. Never ignore them with `_`.
- Wrap errors with context: `fmt.Errorf("operation X on Y: %w", err)` — use `%w` to preserve stack.
- Define sentinel errors at package level: `var ErrNotFound = errors.New("not found")`.
- Use `errors.Is` to check sentinel errors; `errors.As` to extract wrapped values.

## Testing

- Test files live alongside source: `foo.go` → `foo_test.go`.
- Use table-driven tests for multiple cases: `for _, tt := range tests { ... }`.
- Use `t.TempDir()` for file I/O tests; `t.Setenv()` for environment variables.
- Avoid `time.Sleep()` in tests — use channels, mocks, or deterministic fixtures.
- Concurrent tests use `t.Parallel()` and isolated state per test.

## Concurrency

- Goroutines are cheap but still have a cost — don't spawn carelessly.
- Use channels to coordinate goroutines; avoid naked mutexes where possible.
- Always close channels from the sender side; receivers should not close.
- Use `context.Context` for cancellation and timeouts — pass it as the first argument.

## Dependencies and imports

- Avoid circular imports.
- Standard library imports come first, then third-party, then local.
- Use internal/ directories for packages not meant for external use.
- Prefer small, focused packages over god packages.
