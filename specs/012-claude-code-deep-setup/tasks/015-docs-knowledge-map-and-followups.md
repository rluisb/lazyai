# Task 015 — Docs: KNOWLEDGE_MAP + root codebase map + follow-ups

**Phase:** 4 (verification)
**Estimated LOC:** ~30

## Goal

Update the project's meta-documentation to reflect spec 012's completion and record the deferred items from the research decisions.

## Files to touch

| File | Change |
|---|---|
| `specs/KNOWLEDGE_MAP.md` | Add row: `012 | Claude Code deep setup (global/project/workspace) | <status> | <branch>`. Move Pending items for Claude Code to `~~struck~~` once complete. Add new Pending entry: `settings.local.json coverage (deferred from spec 012)`. Add new Pending entry: `ship ai-setup as a Claude plugin manifest (deferred from spec 012)`. |
| `CLAUDE.md` (root) | If Task 008 added `internal/adapter/claude_cli.go`, append to the Codebase Map table. |
| `specs/KNOWLEDGE_MAP.md` — Packages Reference | Add `internal/adapter/claude_cli.go` with its purpose if created. |

## Acceptance criteria

- [ ] KNOWLEDGE_MAP row exists for spec 012 with correct branch ref
- [ ] Two deferred items listed in Pending with (deferred from spec 012) tags
- [ ] Codebase Map in root `CLAUDE.md` is consistent with actual package list
- [ ] `grep -r "012-claude-code-deep-setup"` shows references from KNOWLEDGE_MAP and plan.md only (no orphan references)

## Test plan

- Manual review pass.
- `git diff specs/KNOWLEDGE_MAP.md CLAUDE.md` is coherent and minimal.

## Dependencies

- All prior tasks complete (this is the final doc-update task).
