# Task 012 — Scope-parity regression test for (copilot, global)

**Phase:** 6 (tests)
**Estimated LOC:** ~90

## Goal

Add explicit scope-parity coverage for Copilot at global scope. Assert the full expected file set lands under `~/.copilot/` when probe passes; assert nothing is written when probe fails.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/scope_test.go` (or existing `adapter_scope_test.go`) | Add (copilot, global) matrix entry. Expected paths: `~/.copilot/agents/<name>.agent.yaml` for each library agent + migrated skill; `~/.copilot/copilot-instructions.md`; `~/.copilot/mcp-config.json` when MCP catalog present. |
| `internal/adapter/scope_test.go` | Add paired `TestCopilot_GlobalScope_ProbeFails_NoWrites` — stub probe to return false, install, assert zero files written under `t.TempDir()`-rooted HOME. |
| `internal/adapter/scope_test.go` | Add `TestCopilot_GlobalScope_WithCatalog` — seed a `.ai/mcp.json`, run install + compile, assert `mcp-config.json` emitted with server entries. |

## Acceptance criteria

- [ ] (copilot, global) explicitly covered in scope-parity matrix
- [ ] Probe-fail path asserted (no side effects on HOME)
- [ ] MCP catalog → `mcp-config.json` path asserted
- [ ] Test isolates HOME via `t.TempDir()` + `t.Setenv("HOME", ...)`

## Test plan

Extend the existing table-driven scope-parity test rather than introducing a parallel structure.

## Notes

- If the existing scope-parity tests drive probe results via a global var, introduce an injection seam as part of this task rather than mutating live `exec.LookPath`.
