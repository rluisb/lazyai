# Task 010 — MCP user-scope compile: deep-merge `~/.copilot/mcp-config.json`

**Phase:** 4 (global MCP)
**Estimated LOC:** ~130

## Goal

Extend `compileCopilotMCP` so that at `SetupScopeGlobal` it writes `~/.copilot/mcp-config.json` via deep-merge (managed servers win on key collision; user-authored servers preserved). At project/workspace scope, keep existing `.vscode/mcp.json` behavior (from task 006).

## Files to touch

| File | Change |
|---|---|
| `internal/adapter/mcp_compiler.go` | Split `compileCopilotMCP` into `compileCopilotMCPProject(ctx, servers)` (current behavior) and `compileCopilotMCPGlobal(ctx, servers)` (new). Dispatch on `ctx.SetupScope`. |
| `internal/adapter/mcp_compiler.go` | New helper `toCopilotCliMcp(servers)` emitting the `mcpServers` top-level shape from research §2.4 (stdio: `type/command/args/env`; http/sse: `type/url/headers`). |
| `internal/adapter/mcp_compiler.go` | Use `configmerge.MergeJSONFile` with `.bak` backup-on-first-touch to write `~/.copilot/mcp-config.json`. Per-server deep merge keyed on server name; new/updated managed servers win; user-authored servers untouched. |
| `internal/adapter/mcp_compiler_test.go` | Add `TestCopilotMCP_UserScope_DeepMerge` covering: empty pre-existing file; pre-existing user server preserved; pre-existing managed server updated; ai-setup-added server emitted fresh. |
| `internal/adapter/mcp_compiler_scope_test.go` | Add (copilot, global) entry to scope-parity matrix expecting `~/.copilot/mcp-config.json` under home dir. |

## Schema to emit (from research §2.4)

```jsonc
{
  "mcpServers": {
    "<name>": {
      // stdio
      "type": "stdio",
      "command": "<bin>",
      "args": ["..."],
      "env": { "KEY": "VAL" }
    }
    // or
    "<name>": {
      "type": "http",
      "url": "https://...",
      "headers": { "Authorization": "Bearer ..." }
    }
  }
}
```

## Deep-merge semantics

Model on OpenCode's spec 011 task 004 per-server merge. Pseudocode:
```
existing = readJSONWithBackupOnce(path)
existing.mcpServers ??= {}
for name, server in managedServers:
    existing.mcpServers[name] = toCopilotCliMcp(server)    // managed wins
writeJSON(path, existing)
```

User-authored keys at the top level (beyond `mcpServers`) are preserved untouched via `configmerge.MergeJSONFile`.

## Acceptance criteria

- [ ] Global scope: `~/.copilot/mcp-config.json` written with managed servers under `mcpServers`
- [ ] User-authored servers present in pre-existing file survive the merge
- [ ] User-authored top-level keys (e.g. custom config) survive the merge
- [ ] `.bak` sidecar created on first touch; not re-created on subsequent runs
- [ ] Project/workspace scope `.vscode/mcp.json` behavior unchanged (task 006's inputs scaffolding still applies)

## Test plan

- `TestCopilotMCP_UserScope_FreshInstall` — no pre-existing file, servers written
- `TestCopilotMCP_UserScope_PreservesUserServer` — pre-seed `{mcpServers:{user-foo:{...}}}`, add managed `bar`, assert both present
- `TestCopilotMCP_UserScope_UpdatesManagedServer` — pre-seed with a managed name that we re-emit, assert updated
- `TestCopilotMCP_UserScope_PreservesTopLevelKeys` — pre-seed a `customKey: "x"` alongside `mcpServers`, assert untouched
- `TestCopilotMCP_UserScope_BackupOnce` — assert `.bak` only created once

## Notes

- Canonical output uses `"type": "stdio"` (never the undocumented `"local"` alias).
- No CLI orchestration — Copilot has no `mcp add-json` non-interactive subcommand (research §3).
