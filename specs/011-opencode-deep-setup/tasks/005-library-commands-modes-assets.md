# Task 005 — Add `library/opencode/commands/` and `library/opencode/modes/` assets

**Phase:** 3
**Status:** ✅ complete (2026-04-19)
**Depends on:** none

## Implementation Notes

- Added `library/opencode/commands/` with three starter commands — `review.md` (diff review), `test.md` (run + summarize tests), `commit.md` (draft Conventional Commit for staged diff). Each uses opencode command frontmatter (`description` only; agent/model left for opencode defaults).
- Added `library/opencode/modes/` with two starter modes — `plan.md` (read-only planning: write/edit/bash/patch all `false`) and `audit.md` (security/quality audit: also read-only, with `webfetch: false` for tighter isolation).
- No embed-directive change needed — `//go:embed all:library` in `main.go:13` already captures the new subtree.
- Extended `internal/library/integration_test.go`:
  - Added `opencode/commands` and `opencode/modes` to the expected-dirs list.
  - New `TestGetLibraryFS_OpenCodeAssetsAreEmbedded` walks each starter asset, asserts non-empty bytes, leading `---` frontmatter fence, and the required frontmatter key (`description:` for commands, `tools:` for modes).

## Verification

- `go test ./internal/library/ -count=1` — PASS (new test + existing)
- `go vet ./...` — clean

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
