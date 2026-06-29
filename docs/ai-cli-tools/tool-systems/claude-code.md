---
title: Claude Code — Tool System
summary: Built-in tools, custom tools (skills/subagents/hooks), and MCP configuration for Claude Code.
status: verified
verified_on: 2026-06-29
provenance: official-docs (code.claude.com) read directly in-session
---

# Claude Code — Tool System

## Built-in tools

The tool **name is the exact string** used in permission rules, subagent `tools`, and hook matchers. Disable a tool by adding its bare name to `permissions.deny`.

| Permission required (`Yes`) | No permission (`No`) |
|---|---|
| `Bash`, `Edit`, `Write`, `NotebookEdit`, `WebFetch`, `WebSearch`, `Monitor`, `PowerShell`, `Skill`, `Artifact`, `Workflow`, `ExitPlanMode`, `ShareOnboardingGuide` | `Read`, `Grep`, `Glob`, `LSP`, `Agent`, `AskUserQuestion`, `EnterPlanMode`, `EnterWorktree`, `ExitWorktree`, `Cron*`, `Task*`, `TodoWrite`, `ToolSearch`, `WaitForMcpServers`, `ListMcpResourcesTool`, `ReadMcpResourceTool`, `SendMessage`, `PushNotification`, `RemoteTrigger`, `ScheduleWakeup` |

> To add custom tools, the docs say: **connect an MCP server.** Skills run through the existing `Skill` tool and do **not** add a new tool entry.

### Permission / allow-deny model

`allow` / `ask` / `deny` arrays in `settings.json`, evaluated **deny → ask → allow** (first match wins; specificity does not reorder).

- Bare tool name (`Bash`) in `deny` removes the tool from context entirely.
- Scoped rule (`Bash(rm *)`) leaves the tool available and blocks matching calls.

Rule formats by tool:

| Rule format | Applies to |
|---|---|
| `Bash(npm run *)` | Bash, Monitor (glob; wrappers `timeout`/`time`/`nice`/`nohup`/`stdbuf` stripped) |
| `PowerShell(Get-ChildItem *)` | PowerShell (alias-canonicalized) |
| `Read(~/secrets/**)` | Read, Grep, Glob, LSP (gitignore-style paths) |
| `Edit(/src/**)` | Edit, Write, NotebookEdit (also grants Read on same path) |
| `Skill(deploy *)` | Skill |
| `Agent(Explore)` | Agent (subagent type) |
| `WebFetch(domain:example.com)` | WebFetch |
| `WebSearch` | WebSearch (no specifier) |
| `mcp__puppeteer` / `mcp__github__get_*` | MCP (server, or server+tool glob) |

Path anchors: `//abs`, `~/home`, `/project-root`, `path`/`./path` (cwd). Tool-name globs: `mcp__*` matches all MCP tools (deny/ask); allow globs require a literal `mcp__<server>__` prefix.

Permission modes (`defaultMode`): `default`, `acceptEdits`, `plan`, `auto`, `dontAsk`, `bypassPermissions`.

## Custom tools

The supported surfaces that behave like custom commands/tools:

### Skills (merged with slash commands)

`SKILL.md` directories. `.claude/commands/*.md` still work and are equivalent.

| Location | Path | Applies to |
|---|---|---|
| Personal | `~/.claude/skills/<name>/SKILL.md` | All your projects |
| Project | `.claude/skills/<name>/SKILL.md` | This project |
| Plugin | `<plugin>/skills/<name>/SKILL.md` | Where plugin enabled |
| Enterprise | managed settings dir | Org-wide |

Frontmatter (all optional; `description` recommended): `name`, `description`, `when_to_use`, `argument-hint`, `arguments`, `disable-model-invocation`, `user-invocable`, `allowed-tools`, `disallowed-tools`, `model`, `effort`, `context: fork`, `agent`, `hooks`, `paths`, `shell`.

String substitution: `$ARGUMENTS`, `$ARGUMENTS[N]`, `$N`, `$name`, `${CLAUDE_SESSION_ID}`, `${CLAUDE_EFFORT}`, `${CLAUDE_SKILL_DIR}`. Dynamic context injection via `` !`command` ``. Command name comes from the **directory name** (except a plugin-root `SKILL.md`, which uses frontmatter `name`).

### Subagents

`.md` + YAML frontmatter.

| Location | Scope | Priority |
|---|---|---|
| Managed settings | Org-wide | 1 (highest) |
| `--agents` CLI flag (JSON) | Session | 2 |
| `.claude/agents/` | Project | 3 |
| `~/.claude/agents/` | User | 4 |
| Plugin `agents/` | Plugin | 5 |

Frontmatter (`name`, `description` required): `tools`, `disallowedTools`, `model` (`sonnet`/`opus`/`haiku`/`fable`/full ID/`inherit`), `permissionMode`, `maxTurns`, `skills`, `mcpServers`, `hooks`, `memory`, `background`, `effort`, `isolation: worktree`, `color`, `initialPrompt`. The markdown body is the system prompt.

!!! note "Plugin subagents are restricted"
    Plugin subagents ignore `hooks`, `mcpServers`, and `permissionMode`.

Built-in subagents: **Explore** (read-only, Haiku), **Plan** (read-only), **general-purpose** (all tools), plus `statusline-setup`, `claude-code-guide`.

### Hooks

Lifecycle automation configured in `settings.json` (and `ConfigChange`, `SubagentStart`, `Stop`, `PreToolUse`, etc.). Hook `matcher` fields use **bare tool names**, not the parenthesized rule format.

## MCP tools

Three scopes:

| Scope | Loads in | Shared | Stored in |
|---|---|---|---|
| Local (default) | Current project | No | `~/.claude.json` (per-project) |
| Project | Current project | Yes (committed) | `.mcp.json` in project root |
| User | All your projects | No | `~/.claude.json` |

> Scope rename: older `project` → now `local`; older `global` → now `user`. The shared committed file is the **`project`** scope.

### Transports

- **http** (recommended): `claude mcp add --transport http <name> <url>`. In JSON, `type` accepts `streamable-http` as an alias for `http`.
- **sse** (deprecated): `--transport sse`.
- **stdio**: `claude mcp add [opts] <name> -- <command> [args...]`. `${CLAUDE_PROJECT_DIR}` is set in the spawned server's env; in non-plugin `.mcp.json` reference it as `${CLAUDE_PROJECT_DIR:-.}`.
- **ws** (WebSocket): `type:"ws"` via `claude mcp add-json` only.

### JSON entry (stdio + remote)

```json
{
  "mcpServers": {
    "db": {
      "command": "${CLAUDE_PLUGIN_ROOT}/servers/db-server",
      "args": ["--config", "${CLAUDE_PLUGIN_ROOT}/config.json"],
      "env": { "DB_URL": "${DB_URL}" }
    },
    "events": {
      "type": "ws",
      "url": "wss://mcp.example.com/socket",
      "headers": { "Authorization": "Bearer YOUR_TOKEN" }
    }
  }
}
```

Per-server fields: `command`/`args`/`env`, `url`, `type`, `headers`, `headersHelper`, `timeout` (ms; <1000 ignored), `alwaysLoad`.

### Tool names

`mcp__<server>__<tool>`. Plugin-bundled: `mcp__plugin_<plugin-name>_<server-name>__<tool-name>`. Server name `workspace` is reserved (skipped with a warning).

### Management

`claude mcp list` / `get <name>` / `remove <name>`; in-session `/mcp` (auth, status, tool counts). Project `.mcp.json` servers appear as `⏸ Pending approval` until approved interactively. HTTP/SSE auto-reconnect with backoff; stdio is not reconnected. Managed allow/deny: `allowedMcpServers`, `deniedMcpServers`, `allowManagedMcpServersOnly`.

## Config files & scopes summary

| Feature | User | Project | Local |
|---|---|---|---|
| Settings | `~/.claude/settings.json` | `.claude/settings.json` | `.claude/settings.local.json` |
| Subagents | `~/.claude/agents/` | `.claude/agents/` | — |
| MCP servers | `~/.claude.json` | `.mcp.json` | `~/.claude.json` (per-project) |
| CLAUDE.md | `~/.claude/CLAUDE.md` | `CLAUDE.md` or `.claude/CLAUDE.md` | `CLAUDE.local.md` |

Settings precedence: Managed > CLI args > Local > Project > User. Permission rules **merge** across scopes (not override).

## Gotchas

1. "Custom tools" = MCP. Skills/subagents/commands are not new tool entries.
2. Scope names changed; the shared `.mcp.json` is `project`, not `local`.
3. Project `.mcp.json` servers require interactive approval (`⏸ Pending approval`).
4. `${CLAUDE_PROJECT_DIR}` in non-plugin `.mcp.json` needs `:-.` default; plugin configs substitute it directly.
5. `type` accepts `streamable-http` as alias for `http`; `sse` is deprecated.
6. Permission deny of a bare tool name removes it from context; scoped deny only blocks matches.
7. Plugin subagents silently ignore `hooks`/`mcpServers`/`permissionMode`.
8. `mcp__*` tool-name glob denies all MCP tools; allow globs must anchor to a literal server prefix.

## Sources

- https://code.claude.com/docs/en/overview
- https://code.claude.com/docs/en/tools-reference.md
- https://code.claude.com/docs/en/permissions.md
- https://code.claude.com/docs/en/settings.md
- https://code.claude.com/docs/en/mcp.md
- https://code.claude.com/docs/en/sub-agents.md
- https://code.claude.com/docs/en/slash-commands.md (skills)
