---
title: Kiro CLI — Tool System
summary: Built-in tools, custom tools, and MCP configuration for Kiro CLI.
status: verified
verified_on: 2026-06-29
provenance: official-docs (kiro.dev) via research subagent
---

# Kiro CLI — Tool System

## Built-in tools

Canonical names with aliases:

| Canonical | Alias | Purpose |
|---|---|---|
| `fs_read` | `read` | File read |
| `fs_write` | `write` | File write |
| `execute_bash` | `shell` | Shell/bash execution |
| `use_aws` | `aws` | AWS CLI integration |

No exhaustive built-in catalog is published; tools are documented implicitly via hook matchers and agent examples. Default behavior: all tools available, **confirmation required for most operations**.

### Permission / allow-deny model

Five overlapping mechanisms:

1. Agent `tools` — whitelist of tool names the agent may use.
2. Agent `allowedTools` — pre-approved tools that run **without** confirmation.
3. MCP `autoApprove` array — tool names that bypass prompting; `["*"]` = all.
4. MCP `disabledTools` array — tool names to omit entirely from that server.
5. MCP `disabled: true` — disable the whole server without removing config.

Hook matcher selectors: `@builtin` (all built-ins), `@git` (all tools from a named MCP server), `@git/status` (a specific MCP tool), `*` (all tools), or no matcher (all tools).

## Custom tools

Kiro has **no user function-call tool API** and no slash-command authoring API. Extensibility comes through:

### Custom agents (primary mechanism)

JSON files.

| Scope | Path |
|---|---|
| Global | `~/.kiro/agents/<name>.json` |
| Workspace | `.kiro/agents/<name>.json` |

Creation: `/agent create my-agent` (AI-assisted) or `--manual`; `kiro-cli agent create my-agent`; `/agent create my-agent --from existing-agent`.

Schema:

```json
{
  "name": "my-agent",
  "description": "A custom agent",
  "tools": ["read", "write"],
  "allowedTools": ["read"],
  "resources": [
    "file://README.md",
    "file://.kiro/steering/**/*.md",
    "skill://.kiro/skills/**/SKILL.md"
  ],
  "prompt": "You are a helpful coding assistant",
  "model": "claude-sonnet-4",
  "mcpServers": { "fetch": { "command": "fetch3.1", "args": [] } },
  "includeMcpJson": false,
  "hooks": {}
}
```

| Field | Type | Description |
|---|---|---|
| `name` | string | Agent identifier |
| `description` | string | Human description |
| `tools` | array | Whitelist of tool names |
| `allowedTools` | array | Tools auto-approved without prompting |
| `resources` | array | Files/globs/skills injected into context (`file://`, `skill://`) |
| `prompt` | string | System prompt |
| `model` | string | Model override (e.g. `claude-sonnet-4`) |
| `mcpServers` | object | MCP servers scoped to this agent |
| `includeMcpJson` | boolean | Whether to also include workspace/user `mcp.json` servers |
| `hooks` | object | Hook configuration |

Reserved built-in agents (cannot be edited): `kiro_default`, `kiro_help`, `kiro_planner`.

### Hooks (lifecycle scripting)

Defined inside the agent config. Events: `agentSpawn`, `userPromptSubmit`, `preToolUse`, `postToolUse`, `stop`.

Stdin payload:

```json
{
  "hook_event_name": "preToolUse",
  "cwd": "/current/dir",
  "session_id": "abc123",
  "tool_name": "read",
  "tool_input": { "operations": [{ "mode": "Line", "path": "..." }] }
}
```

Exit-code semantics:

- `preToolUse` exit `2` → block tool execution; STDERR returned to the LLM.
- `stop` stdout JSON `{"decision":"block","reason":"..."}` → agent continues instead of stopping.
- exit `0` → success.

### Steering files

Persistent project context. `.kiro/steering/*.md` (workspace) or `~/.kiro/steering/*.md` (global). Foundational files: `product.md`, `tech.md`, `structure.md`. `AGENTS.md` (workspace root or `~/.kiro/steering/`) is always included.

!!! warning "Custom agents do NOT auto-load steering"
    Only the default `kiro_default` agent auto-loads steering. Custom agents must declare it explicitly: `"resources": ["file://.kiro/steering/**/*.md"]`.

## MCP tools

| Scope | Path | Format |
|---|---|---|
| Workspace | `<project>/.kiro/settings/mcp.json` | JSON |
| User/global | `~/.kiro/settings/mcp.json` | JSON |

Both merge when present; workspace takes precedence at the server-name key level (no field-level merge).

### Schema

```json
{
  "mcpServers": {
    "local-server": {
      "command": "uvx",
      "args": ["awslabs.aws-documentation-mcp-server@latest"],
      "env": { "API_KEY": "hard-coded", "OTHER": "${SHELL_ENV_VAR}" },
      "disabled": false,
      "autoApprove": ["tool1", "tool2"],
      "disabledTools": ["tool3"]
    },
    "remote-server": {
      "url": "https://api.example.com/mcp",
      "headers": { "Authorization": "Bearer token" },
      "oauth": { "clientId": "your-app-client-id", "redirectUri": "127.0.0.1:8080" },
      "oauthScopes": ["read", "write"],
      "disabled": false,
      "autoApprove": ["*"],
      "disabledTools": ["danger-tool"]
    },
    "typed-http": {
      "type": "http",
      "url": "https://api.example.com/mcp",
      "oauth": { "redirectUri": "127.0.0.1:8080", "oauthScopes": ["read", "write"] }
    }
  }
}
```

Local server: `command` (req), `args` (req), `env` (`${VAR}` expansion), `disabled`, `autoApprove`, `disabledTools`.

Remote server: `url` (req), `type` (`"http"`, default inferred from `url`), `headers`, `env`, `oauth.clientId`, `oauth.redirectUri` (`"host:port"`), `oauthScopes` (**top-level, NOT under `oauth`**), `disabled`, `autoApprove`, `disabledTools`.

### Transports

- **stdio** — default when `command`+`args` present.
- **HTTP** — when `url` present (or `type:"http"`). Supports OAuth. No separate SSE type; HTTP covers SSE endpoints.

### CLI setup

```bash
kiro-cli mcp add \
  --name "awslabs.aws-documentation-mcp-server" \
  --scope global \
  --command "uvx" \
  --args "awslabs.aws-documentation-mcp-server@latest" \
  --env "FASTMCP_LOG_LEVEL=ERROR"
```

`--scope global` → `~/.kiro/settings/mcp.json`. In-session inspection: `/mcp`.

### Agent-level MCP override

`includeMcpJson: false` → agent uses only its own `mcpServers`, ignoring all `mcp.json` files. `true` → merges user+workspace `mcp.json` with the agent's own.

### Tool name validation

- Must match `^[a-zA-Z][a-zA-Z0-9_]*$`; ≤64 chars incl. server prefix.
- Description must be non-empty.
- Violations → tool silently excluded (warning shown).
- Descriptions >10,000 chars → performance warning, still works.

## Gotchas

1. `disabledTools` removes specific tools; `disabled:true` kills the whole server.
2. Two separate allow-lists: MCP `autoApprove` (server-scoped MCP tools) vs agent `allowedTools` (agent-scoped built-ins).
3. `includeMcpJson:false` silently drops all `mcp.json` servers.
4. Steering not auto-loaded in custom agents — declare in `resources`.
5. `oauthScopes` is top-level; placing it under `oauth` is silently ignored.
6. OAuth: only public/PKCE clients (no `client_secret`) are supported.
7. Empty-scopes workaround: `"oauthScopes": []` if scope errors occur.
8. Hyphenated MCP tool names (`get-file`) fail validation and are excluded.
9. `mcp.json` merge is server-key-level; no within-entry field merge.
10. `--description`/`--mcp-server` flags exist only on the `/agent create` slash command, not `kiro-cli agent create`.
11. `type:"http"` is explicit in agent-level examples but inferred from `url` in top-level `mcp.json`. [INFERENCE] `url` alone is sufficient there, not explicitly confirmed.

## Sources

- https://kiro.dev/docs/cli/
- https://kiro.dev/docs/cli/mcp/
- https://kiro.dev/docs/cli/custom-agents/
- https://kiro.dev/docs/cli/custom-agents/creating/
- https://kiro.dev/docs/cli/hooks/
- https://kiro.dev/docs/cli/steering/
- https://kiro.dev/docs/cli/chat/
- https://kiro.dev/docs/cli/reference/slash-commands/
- https://kiro.dev/docs/mcp/
- https://kiro.dev/docs/mcp/configuration/
