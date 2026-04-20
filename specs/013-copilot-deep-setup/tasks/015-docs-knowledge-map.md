# Task 015 — Docs + KNOWLEDGE_MAP updates

**Phase:** 6 (docs)
**Estimated LOC:** ~40

## Goal

Close the self-improvement loop: record spec 013 decisions, new packages, pending follow-ups.

## Files to touch

| File | Change |
|---|---|
| `specs/KNOWLEDGE_MAP.md` | Add spec 013 row to Feature Specs table (status + branch). Add decision rows: "Copilot × global probe-gated" (reason: detect standalone CLI); "Skills emit as `.agent.yaml` for Copilot" (reason: standalone CLI ignores `.prompt.md`); "Copilot MCP user-scope via deep-merge" (reason: no CLI `mcp add-json` exists); "Copilot chatmodes remain VS-Code-only" (reason: CLI has no chatmode concept). |
| `specs/KNOWLEDGE_MAP.md` | Add to Packages Reference: `internal/adapter/copilot_cli.go`, `internal/adapter/copilot_validate.go`, `library/copilot/agents/`, `library/copilot/instructions/`. |
| `specs/KNOWLEDGE_MAP.md` | Add to Pending/Follow-up: "[ ] Copilot cloud/marketplace publishing (Option C from spec 013 research §5) — future spec"; "[ ] Copilot `--skip-validation` flag to bypass per-agent smoke (spec 013 task 011 trade-off)". |
| Root `CLAUDE.md` | Append to Codebase Map any new package paths introduced. |

## Acceptance criteria

- [ ] Spec 013 row present with current branch
- [ ] Four new decision rows accurate to research §10 locks
- [ ] Pending/Follow-up entries for parked items
- [ ] Packages reference mentions new files

## Test plan

N/A — doc-only. Verify via `git diff` review.

## Notes

- This task closes the spec. Mark 013 ✅ Complete in the table once all previous tasks are merged.
