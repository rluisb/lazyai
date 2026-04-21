# Task 002 — `library/copilot/instructions/*.instructions.md` starter set

**Phase:** 1 (library content)
**Estimated LOC:** ~90 (mostly content)

## Goal

Author the language-specific instructions trio (locked decision Q3). Each file uses the VS Code instructions frontmatter schema (`applyTo` glob) and will be copied to `.github/instructions/` at project/workspace scope.

## Files to create

| File | `applyTo` value | Content focus |
|---|---|---|
| `library/copilot/instructions/typescript.instructions.md` | `**/*.{ts,tsx}` | TypeScript conventions: strict null checks, no `any`, prefer `type` over `interface` for unions, import order, testing with Vitest/Jest norms |
| `library/copilot/instructions/go.instructions.md` | `**/*.go` | Go conventions: `go fmt`, error wrapping with `fmt.Errorf("...: %w", err)`, `t.TempDir()` + `t.Setenv()` in tests, no `panic()` outside main |
| `library/copilot/instructions/tests.instructions.md` | `**/*_test.*,**/*.test.*,**/*.spec.*` | Testing guidance: table-driven tests, no sleep-based waits, deterministic fixtures, arrange-act-assert |

## Frontmatter schema (from research §2.1)

```yaml
---
applyTo: "**/*.ts,**/*.tsx"    # required (single glob or comma-separated)
description: <one-line>        # optional
---

<markdown body — instructions the model should follow when editing matching files>
```

## Acceptance criteria

- [ ] Three `*.instructions.md` files exist under `library/copilot/instructions/`
- [ ] Each has non-empty `applyTo` frontmatter
- [ ] Each has a `description` one-liner
- [ ] Body is ≤200 lines, focused, actionable
- [ ] Content aligns with any existing `specs/standards/coding/` conventions we've already committed

## Test plan

Task 003 covers parse + schema. No new tests in this task.

## Notes

- Pull convention content from `specs/standards/coding/` if it exists; otherwise draft from common practice matching our Go project style (see root `CLAUDE.md` "Conventions").
- Globs use comma-separated form for VS Code compatibility.
