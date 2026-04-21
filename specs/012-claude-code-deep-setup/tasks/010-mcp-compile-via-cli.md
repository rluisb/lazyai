# Task 010 — MCP compile via CLI (project/workspace `.mcp.json` reconciliation)

**Phase:** 3 (CLI orchestration for MCP)
**Estimated LOC:** ~70

## Goal

Make `compileClaudeCodeMCP` CLI-aware for project and workspace scopes. Compile is invoked on re-runs to reconcile the canonical MCP catalog with on-disk state. When `claude` is present, delegate to the CLI (`mcp list` → diff → `mcp add-json` for new / `mcp remove` for removed); when absent, fall back to the current direct `.mcp.json` write.

Global scope stays compile-skip (already correct per `mcp_compiler.go:220`).

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/mcp_compiler.go` | Add CLI path to `compileClaudeCodeMCP` (L215-237). Probe via `LookupClaudeBinary`; when present, list servers with `claude mcp list`, parse output, diff against enabled set, add missing / remove stale. On error fall back to direct-write. |
| `internal/adapter/mcp_compiler.go` | Possibly factor the direct-write branch into a helper so the CLI and fallback paths are both readable. |

## List-parse concern

`claude mcp list` output is human-readable (name, URL/command, status). There's no `--json` flag. Options:

1. Parse the human-readable output with a small regex (fragile).
2. Call `claude mcp get <name>` per enabled server and check exit code (cheap, robust).

**Pick option 2.** Iterate enabled servers, per-server `mcp get`; if exit 0 = present (skip add), if non-zero = add. Iterate current-state servers (how?) — if we can't enumerate easily, **just re-add all enabled servers** (`mcp add-json` is idempotent-ish via the pre-check) and accept that stale servers hang around until the user `mcp remove`s them.

## Acceptance criteria

- [ ] With `claude` present, compile reconciles `.mcp.json` through the CLI; file exists and matches `mcp list` afterward
- [ ] With `claude` absent, compile writes `.mcp.json` directly (unchanged behavior)
- [ ] Global scope continues to skip compile
- [ ] Stale-server removal is a known limitation and is documented in a comment + spec Pending list

## Test plan

- Injected fake runner for both present and absent cases.
- Existing `TestCompileMCPForTool_ClaudeGlobalSkips` stays green.
- New test: project scope, CLI present, fake runner records `mcp add-json` calls for each enabled server.

## Dependencies

- Task 008 (probe + runner).
- Task 009 (JSON-payload transformer is shared).
