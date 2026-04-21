# Task 009 — MCP install via `claude mcp add-json` (with silent fallback)

**Phase:** 3 (CLI orchestration for MCP)
**Estimated LOC:** ~120

## Goal

Route MCP server registration through the `claude` CLI when available. For each enabled server, call `claude mcp add-json <name> <json> -s <scope>`. Pre-check with `claude mcp get <name>` and skip if already present. On any failure (binary missing, non-zero exit) fall back to the existing direct-write path silently, with a single warning line per install.

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/claudecode.go` | Replace the current `installClaudeMCPViaCLI` path with a new implementation: probe → for each server: `mcp get` → if absent, `mcp add-json <name> <json> -s <scope>`; capture per-server errors. |
| `internal/adapter/mcp_compiler.go` | Minor — export or expose a transformer that converts canonical MCP entry to the JSON payload `claude mcp add-json` expects. |
| `internal/adapter/claudecode.go` | Fallback: if the CLI is missing or any server fails, fall through to the existing `configmerge.MergeJSONFile` path for the whole batch (don't mix modes). Emit one warning line: `warn: claude CLI unavailable — writing MCP servers via direct-write (reason: <...>)`. |

## Scope → flag mapping

| ai-setup scope | `claude mcp add-json` scope flag | `workingDir` |
|---|---|---|
| `SetupScopeGlobal` | `-s user` | (empty / user home) |
| `SetupScopeProject` | `-s project` | `ctx.TargetDir` |
| `SetupScopeWorkspace` | `-s project` | `ctx.TargetDir` (workspace dir) |

## JSON payload shape

`claude mcp add-json` expects the Claude MCP server schema:

```json
{
  "command": "npx",
  "args": ["-y", "my-mcp-server"],
  "env": { "API_KEY": "${API_KEY}" }
}
```

For HTTP/SSE servers use the appropriate shape (`type: "http"`, `url`, `headers`). Task should ship unit test fixtures pinning the transformation for stdio, http, and sse servers.

## Acceptance criteria

- [ ] With `claude` present: every enabled MCP server is added via `mcp add-json`; `claude mcp list` afterwards shows them all
- [ ] With `claude` absent: behavior matches end-of-Phase-1 (direct merge into `~/.claude/settings.json` or `.mcp.json`), plus one warning line
- [ ] Pre-existing servers (per `mcp get <name>`) are skipped, not duplicated
- [ ] A CLI failure on one server falls back to direct-write for the whole batch (don't leave a half-migrated state)

## Test plan

- Injected fake runner: assert exact invocation for each scope (one test per scope).
- Fake runner with `mcp get` returning "found" → assert `mcp add-json` is NOT called for that server.
- Fake runner with `mcp add-json` returning error → assert fallback path runs and warning is emitted.
- `LookupClaudeBinary` returning `false` → assert direct-write path runs and warning is emitted.

## Dependencies

- Task 008 (probe + runner).
