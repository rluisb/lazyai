---
title: Antigravity (Google) — Tool System
summary: Built-in tools, permissions engine, custom tools (skills/plugins/sidecars/hooks), and the two-file MCP config for Antigravity.
status: verified
verified_on: 2026-06-29
provenance: official-docs (antigravity.google) via research subagent; two-file MCP strategy cross-checked against LazyAI adapter source
---

# Antigravity (Google) — Tool System

## Built-in tools

Fixed set across all surfaces (2.0, CLI, SDK, IDE). Not individually toggleable; governed by the permissions engine. Names matter for hook matchers:

- **Files/dirs**: `view_file`, `write_to_file`, `replace_file_content`, `multi_replace_file_content`, `list_dir`, `find_by_name`
- **Search/research**: `grep_search`, `search_web`, `read_url_content`
- **System/exec**: `run_command`, `manage_task`, `schedule`, `list_permissions`, `ask_permission`
- **Agent collaboration**: `invoke_subagent`, `define_subagent`, `send_message`, `manage_subagents`
- **Interaction/media**: `ask_question`, `generate_image`

### Permissions engine

Every sensitive op is `action(target)`, evaluated **Deny > Ask > Allow** (strict).

| Action | Target format | Default |
|---|---|---|
| `read_file` | `read_file(/path)`, `(dir)`, `(*)` | Ask (workspace auto-allowed) |
| `write_file` | `(/path)`, `(*)` — implicitly grants read | Ask (workspace auto-allowed) |
| `read_url` | `(domain)`, `(*)` | Ask |
| `execute_url` | `(domain)`, `(*)` — browser actuation | Ask |
| `command` | `(prefix)`, `(regex)`, `(*)` | Ask |
| `unsandboxed` | `(prefix/regex/*)` | Ask |
| `mcp` | `mcp(server/tool)`, `mcp(server/*)`, `mcp(*)` | Ask |

Implicit rules: `write_file(X)` grants `read_file(X)`; `deny read_file(X)` blocks `write_file(X)`; workspace files auto-allowed. Terminal sandbox (preview, macOS/Linux) compiles grants into FS/network allowlists. Interactive prompts allow scope editing for files/URLs/MCP but **not** commands.

## Custom tools

No CLI function-tool format. Surface-dependent mechanisms:

### 1. Agent Skills

`SKILL.md` directory + YAML frontmatter (`name` optional, `description` required).

| Scope | Path |
|---|---|
| Workspace | `<root>/.agents/skills/<name>/SKILL.md` |
| Global (CLI/2.0) | `~/.gemini/config/skills/<name>/SKILL.md` |
| Global (IDE) | `~/.gemini/antigravity/skills/<name>/SKILL.md` |
| Backward-compat | `.agent/skills/` (deprecated) |

### 2. Plugins

`plugin.json` manifest bundling skills/rules/MCP/hooks/sidecars.

```
<plugin>/
  plugin.json        # { "name": "..." } — name optional, defaults to dir
  skills/  rules/  mcp_config.json  hooks.json  sidecars/
```

Discovery: `.agents/plugins/<dir>/` or `_agents/plugins/<dir>/` (workspace); `~/.gemini/config/plugins/<dir>/` (global).

### 3. Rules / context

`.agents/rules/*.md` (workspace — the IDE reads this, **not** a bare root `AGENTS.md`); `~/.gemini/GEMINI.md` (global, user-managed).

### 4. Sidecars (background processes)

`sidecar.json` at `~/.gemini/config/sidecars/<name>/` or `<plugin>/sidecars/<name>/`. Fields: `command` (xor `builtin`, currently only `schedule`), `args`, `restart_policy` (`always`/`on-failure`/`never`), `description`, `env`, `display_name`. **Disabled by default** — enable in `~/.gemini/config/config.json`: `{ "sidecar-name": { "enabled": true } }`.

### 5. SDK custom subagents (Python)

Arbitrary Python callables registered as tools; declarative safety policies via `deny()`/`allow()`/`ask_user()` from `google.antigravity.hooks.policy`.

### 6. Hooks

`hooks.json` at `.agents/hooks.json` (workspace) or `~/.gemini/config/hooks.json` (global).

```json
{
  "<hook-name>": {
    "enabled": true,
    "PreToolUse": [
      { "matcher": "run_command",
        "hooks": [ { "type": "command", "command": "./script.sh", "timeout": 10 } ] }
    ],
    "PostToolUse": [],
    "PreInvocation": [],
    "PostInvocation": [],
    "Stop": [ { "type": "command", "command": "./gate.sh" } ]
  }
}
```

| Event | Has matcher? | Purpose |
|---|---|---|
| `PreToolUse` | Yes | Before tool; can `allow`/`deny`/`ask`/`force_ask` |
| `PostToolUse` | Yes | After tool; returns `{}` |
| `PreInvocation` | No | Before model call; inject steps |
| `PostInvocation` | No | After tool calls; inject steps + control loop |
| `Stop` | No | On termination; `"continue"` re-enters loop |

Matchers: `""`/`"*"` all, `"run_command"` exact, `"a|b"` either, `"browser_.*"` regex. Handler fields: `type` (only `"command"`), `command`, `timeout` (seconds in `hooks.json`, default 30). camelCase JSON via stdin/stdout.

!!! warning "Hook array shapes differ by event"
    `PreToolUse`/`PostToolUse` are arrays of `{ matcher, hooks[] }`. `Stop`/`PreInvocation`/`PostInvocation` are **flat arrays of handlers** (no `matcher` wrapper). Mixing shapes silently fails to fire.

## MCP tools — two files, two schemas

Antigravity writes MCP config to **two files with different field names**.

### File 1 — `mcp_config.json` (Antigravity native / CLI)

Remote servers use **`serverUrl`**. `url`/`httpUrl` are explicitly **not supported** here.

```json
{
  "mcpServers": {
    "sqlite-explorer": {
      "command": "node",
      "args": ["/usr/local/bin/sqlite-mcp-server.js"],
      "env": { "SQLITE_DB_PATH": "/var/data/app.db" }
    },
    "my-remote-server": {
      "serverUrl": "https://api.example.com/mcp/",
      "headers": { "Authorization": "Bearer YOUR_API_TOKEN" }
    }
  }
}
```

Fields: `command`, `serverUrl`, `args`, `env`, `cwd`, `headers`, `authProviderType` (`"google_credentials"`), `oauth` (`{clientId, clientSecret}`), `disabled`, `disabledTools`.

### File 2 — `settings.json` (open-source Gemini CLI)

Remote servers use **`httpUrl`** (or `url`+`type`). Per the official Gemini CLI JSON schema `MCPServerConfig`: `command`, `args`, `env`, `cwd`, `url`, `httpUrl`, `headers`, `tcp` (websocket), `type` (`stdio`/`sse`/`http`), `timeout`, `trust`.

### Locations by surface

| Surface | Scope | Path | Format |
|---|---|---|---|
| 2.0 / IDE / CLI | Global | `~/.gemini/config/mcp_config.json` | `serverUrl` |
| IDE / CLI | Workspace | `.agents/mcp_config.json` | `serverUrl` |
| Gemini CLI (OSS) | Global | `~/.gemini/settings.json` | `httpUrl` |
| Gemini CLI (OSS) | Workspace | `.gemini/settings.json` | `httpUrl` |
| Plugin | any | `<plugin>/mcp_config.json` | `serverUrl` |

### Transports & auth

- **stdio**: `command`+`args`+`env`.
- **Streamable HTTP / SSE**: `serverUrl` (`mcp_config.json`) / `httpUrl` or `url`+`type` (`settings.json`).
- **WebSocket**: `tcp` (Gemini CLI schema only).
- Auth: Google ADC (`authProviderType:"google_credentials"`); OAuth auto-DCR; manual `oauth:{clientId,clientSecret}` (redirect `https://antigravity.google/oauth-callback`, tokens at `~/.gemini/antigravity/mcp_oauth_tokens.json`); custom `headers`.

### Permissions & disabling

MCP governed by `mcp` action (default Ask). Disable specific tools: `"disabledTools": ["dangerous_tool"]`. Disable a server: `"disabled": true`. CLI `/mcp` overlay; MCP Store (40+ servers) in 2.0/IDE.

## Gotchas

1. `serverUrl` (native) vs `httpUrl` (Gemini CLI) — wrong field is silently ignored. Write both files so both surfaces discover MCP.
2. Global skills/hooks live under `~/.gemini/config/`, **not** `~/.agents/`.
3. IDE reads `.agents/rules/*.md`, not a bare root `AGENTS.md` (use a bridge file importing `@/AGENTS.md`).
4. Permissions are strict `Deny > Ask > Allow`: `command(*)` in Ask + `command(git)` in Allow → git still prompts.
5. Scope editing on prompts unavailable for commands.
6. Sidecars disabled by default; missing enable entry in `config.json` = never runs.
7. Plugin-scoped skills live under `<plugin>/skills/`; standalone under `<scope>/skills/<name>/` — not interchangeable.
8. `hooks.json` `Stop`/`*Invocation` are flat handler arrays; `*ToolUse` use `{matcher,hooks[]}`.
9. `hooks.json` `timeout` is seconds; hooks embedded in `settings.json` may be milliseconds per the schema — don't mix.
10. Antigravity is GUI-oriented; headless validation/init is a no-op.

## Sources

- https://antigravity.google/docs/home
- https://antigravity.google/docs/permissions
- https://antigravity.google/docs/subagents
- https://antigravity.google/docs/mcp
- https://antigravity.google/docs/hooks
- https://antigravity.google/docs/plugins
- https://antigravity.google/docs/sidecars
- https://antigravity.google/docs/skills
- https://antigravity.google/docs/ide/skills
- https://antigravity.google/docs/settings
- https://antigravity.google/docs/sdk/overview
- https://antigravity.google/docs/cli/overview
- https://antigravity.google/docs/overview
- https://antigravity.google/docs/getting-started
- https://raw.githubusercontent.com/google-gemini/gemini-cli/main/schemas/settings.schema.json

!!! note "Non-doc corroboration"
    The two-file write strategy and pinned paths were additionally cross-checked against LazyAI's own `packages/cli/internal/adapter/antigravity.go`, `mcp_compiler.go`, and install tests (repo source, not official docs).
