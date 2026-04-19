# Task 005 — Add `library/opencode/commands/` and `library/opencode/modes/` assets

**Phase:** 3
**Status:** pending
**Depends on:** none

## Scope

Create the library source bundles for opencode commands and modes. Keep the initial set small and curated — YAGNI.

## Changes

- `library/opencode/commands/` (new):
  - `review.md` — review current branch changes.
  - `test.md` — run project tests and summarize failures.
  - `commit.md` — draft a Conventional-style commit message from staged diff.
  - Frontmatter per opencode docs:
    ```yaml
    ---
    description: <short>
    agent: <optional — primary by default>
    model: <optional>
    ---
    ```
- `library/opencode/modes/` (new):
  - `plan.md` — read-only planning mode with tool restrictions.
  - `audit.md` — code-audit focused mode.
  - Frontmatter per opencode docs:
    ```yaml
    ---
    description: <short>
    model: <optional>
    tools: { write: false, edit: false, bash: false, read: true, grep: true }
    ---
    ```
- `internal/library/embed.go` (or equivalent): extend the embed directive to include `library/opencode/**`.

## Tests

- `internal/library/integration_test.go`:
  - Walk embedded FS; assert `library/opencode/commands/review.md` and `library/opencode/modes/plan.md` are present.
  - Parse each file's frontmatter; assert required keys are present.

## Definition of Done

- All starter assets written with valid opencode frontmatter.
- Embedded FS tests green.
