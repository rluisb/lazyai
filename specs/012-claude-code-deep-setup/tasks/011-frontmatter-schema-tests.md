# Task 011 — Frontmatter schema tests for every emitted Claude artifact

**Phase:** 4 (verification)
**Estimated LOC:** ~120

## Goal

Add a test suite that parses every agent, skill, command, and output-style file written during install and asserts required frontmatter fields per Claude's documented schemas. Catches delimiter drift, missing `description`, schema regressions.

## Files to create / touch

| File | Change |
|---|---|
| `internal/adapter/claudecode_frontmatter_test.go` (new) | Install to a temp dir at each scope; walk `agents/`, `skills/`, `commands/`, `output-styles/`; parse YAML frontmatter per file; assert per-schema required fields. |

## Schemas under test

- **agents/*.md** — require `description`; if `tools` present, assert whitespace-separated.
- **skills/*/SKILL.md** — require `description`; allowed keys from `skills` docs only.
- **commands/*.md** — require `description`.
- **output-styles/*.md** — require `name` and `description`.

## Acceptance criteria

- [ ] Test runs at all three scopes (project / workspace / global)
- [ ] Parse failure on any file = test failure with file path in the error
- [ ] Delimiter check: agent `tools` is whitespace-separated for Claude (catches Task 004 regressions)
- [ ] Assertion catalog is readable — one table-driven test per artifact type

## Test plan

- Self-verifying: this IS the test.
- Add a "negative" golden: intentionally emit a bad agent file (via a test fixture) and assert the parser rejects it.

## Dependencies

- Tasks 001, 004, 007 (emitter changes must be in before the test is authoritative).
