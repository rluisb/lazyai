# Task 005 — Starter commands for Claude

**Phase:** 2 (starter library assets)
**Estimated LOC:** ~60 (mostly markdown content)

## Goal

Ship three starter slash commands under `library/claudecode/commands/` with Claude-conformant frontmatter. Mirrors the OpenCode set from spec 011 (`review`, `test`, `commit`) so users get immediate value at all scopes.

## Files to create

| File | Content |
|---|---|
| `library/claudecode/commands/review.md` | `description`, `argument-hint: "[pr-number-or-path]"`, `allowed-tools`, body: review current changes or a named PR/path. |
| `library/claudecode/commands/test.md` | `description`, `argument-hint: "[package-or-file]"`, `allowed-tools: Bash(go test *)`, body: run tests for the target, report failures. |
| `library/claudecode/commands/commit.md` | `description`, `argument-hint`, `allowed-tools: Bash(git *)`, body: stage, draft Conventional commit message from diff, confirm, commit. |

## Frontmatter contract (Claude commands)

```yaml
---
description: Brief one-liner — what this command does
argument-hint: "[optional-hint]"
allowed-tools: Bash(...) Read Edit   # whitespace-separated
model: sonnet                         # optional
---
```

Body uses `$ARGUMENTS`, `$0`, etc. per Claude's skill/command substitution rules.

## Acceptance criteria

- [ ] Three markdown files exist with valid frontmatter
- [ ] Each file parses cleanly as YAML frontmatter + markdown body
- [ ] Content is genuinely useful (not stub placeholder text)
- [ ] Files are embedded in the library FS

## Test plan

- Parser test: load each file, verify frontmatter has `description` and `allowed-tools`.
- Smoke: manually run `claude /review` etc. in a test project after install (documented in spec `## Verification`).

## Dependencies

- Library FS embed must pick up `library/claudecode/**` — check `internal/library/integration_test.go` or equivalent.
