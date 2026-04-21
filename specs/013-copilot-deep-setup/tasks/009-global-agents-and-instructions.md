# Task 009 — Global-scope emitters: `~/.copilot/agents/` + `~/.copilot/copilot-instructions.md`

**Phase:** 4 (global emitters)
**Estimated LOC:** ~110

## Goal

When global-scope install passes the probe, emit:
1. `~/.copilot/agents/*.agent.yaml` — same content as the `.github/agents/` set (library + migrated skills).
2. `~/.copilot/copilot-instructions.md` — template-filled **first-install only** (re-run leaves it alone).

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot.go` | In the global-scope branch after probe passes, resolve root via `ResolveToolRoot(ToolIdCopilot, SetupScopeGlobal, ctx)` → `<home>/.copilot`. Reuse the helpers from tasks 004 + 005 pointed at that root for `agents/`. Emit `copilot-instructions.md` at the root with first-install-only semantics. |
| `library/copilot/copilot-instructions.template.md` (new) | Template with placeholders mirroring `library/root/CLAUDE.template.md` — project/org/team fill-ins, mechanical-infer markers for anything we can derive. |
| `internal/scaffold/root.go` | If template fill logic is already factored (spec 010 `fillClaudeMdPlaceholders`), add a parallel `fillCopilotInstructionsPlaceholders` or extend the existing helper to take a template path. |

## First-install-only semantics

- Check `files.FileExists(destPath)` before writing.
- If exists, skip with a debug log; do not overwrite.
- If absent, template-fill then write.
- Track in `FileRecords` only when we actually wrote the file.

## Acceptance criteria

- [ ] `~/.copilot/agents/<name>.agent.yaml` emitted for every selected agent + migrated skill when probe passes
- [ ] `~/.copilot/copilot-instructions.md` emitted on first install; untouched on re-run with existing file
- [ ] Template placeholders filled consistently with CLAUDE.md hybrid fill behavior (spec 010)
- [ ] File records correctly tagged `Owner: FileOwnerLibrary` (agents) / `FileOwnerUser` if template is user-editable (match spec 010 choice)

## Test plan

- `TestCopilot_GlobalScope_EmitsAgents` — `t.Setenv("HOME", t.TempDir())`, install, assert files
- `TestCopilot_GlobalScope_InstructionsFirstInstallOnly` — pre-seed, install, assert content unchanged
- `TestCopilot_GlobalScope_InstructionsAbsent_FirstInstallFills` — no pre-seed, install, assert template filled
- Content equality — global `.agent.yaml` byte-identical to project `.agent.yaml` for the same library source

## Notes

- MCP compile is task 010; this task is files + instructions only.
- If `fillClaudeMdPlaceholders` refactor grows the LOC beyond the estimate, split into a follow-up task before merging.
