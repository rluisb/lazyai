# Task 014 — MCP round-trip + inputs-scaffold integration tests

**Phase:** 6 (tests)
**Estimated LOC:** ~110

## Goal

Integration-level coverage for the two MCP emission paths. Exercises real file I/O with `configmerge` on a tempdir; validates inputs-scaffold regex behavior across realistic catalog shapes.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/copilot_mcp_integration_test.go` (new) | Seed a user-authored `~/.copilot/mcp-config.json` with one custom server + one ai-setup-owned server. Seed `.ai/mcp.json` with a new managed server + the same managed name as the pre-existing. Run compile at global scope. Assert: custom preserved; both managed written; `.bak` sidecar created. |
| Same file | Seed `.ai/mcp.json` with a server whose `env` contains `${GITHUB_TOKEN}` and `${FOO}`; another server whose `env` also contains `${GITHUB_TOKEN}`. Run compile at project scope. Assert `.vscode/mcp.json` has exactly two `inputs` entries (GITHUB_TOKEN, FOO), deduped. |
| Same file | Seed `.ai/mcp.json` with only literal env values. Assert `.vscode/mcp.json` omits `inputs` key entirely. |

## Acceptance criteria

- [ ] User-authored `mcpServers` keys preserved across compile
- [ ] ai-setup-managed keys updated on re-compile
- [ ] Inputs scaffolded on placeholder detection, omitted otherwise
- [ ] Deduped across multiple servers sharing the same placeholder
- [ ] `.bak` sidecar present after first compile; not recreated on second

## Test plan

Two parallel integration tests — one for global-scope deep-merge, one for project-scope inputs scaffolding. Both use `t.TempDir()` and direct file assertions (not stubs).

## Notes

- If the test hits rate-limits or flakes on concurrent file access, wrap setup in `t.Parallel()` carefully — probably just keep these serial.
