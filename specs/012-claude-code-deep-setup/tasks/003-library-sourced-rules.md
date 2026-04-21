# Task 003 — Source rules from library/rules/ instead of hardcoded string

**Phase:** 1 (structural fixes, no CLI)
**Estimated LOC:** ~30

## Goal

Replace the hardcoded TypeScript rule content in `claudecode.go:63` with a copy from `library/rules/typescript.md`. Matches the sourcing pattern used by every other asset (agents, skills, templates) and removes a fragile hardcoded duplicate.

## Files to touch

| File | Change |
|---|---|
| `library/rules/typescript.md` | Create if missing. Content = whatever the current hardcoded string writes (lift verbatim). |
| `internal/adapter/claudecode.go` | Delete the hardcoded literal at ~L63. Replace with a read-from-library + write pattern (reuse `CopyLibraryFile` / existing helpers). |
| `internal/library/embed.go` (or wherever the embed tree is declared) | Ensure `library/rules/*.md` is included in the embedded FS if not already. |

## Acceptance criteria

- [ ] `library/rules/typescript.md` exists and is embedded
- [ ] `.claude/rules/typescript.md` at every scope is byte-identical to the library source
- [ ] Adding a second rule (e.g. `library/rules/go.md`) requires only a library drop, no adapter change (verify by manually dropping a test rule and confirming it appears at `~/.claude/rules/go.md`)

## Test plan

- New test: compare bytes of installed `.claude/rules/typescript.md` vs the library source.
- Update the existing scope-parity test if it asserts anything about the rule content.

## Notes

- Keep the fallback safe — if the library file is missing at build time, fail the build (embed assertion), don't silently skip at install time.
