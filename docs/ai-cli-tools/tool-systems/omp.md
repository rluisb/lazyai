---
title: OMP (Oh My Pi) — Tool System
summary: Built-in tools, TypeScript custom tools, MCP configuration, and the broader authoring surface for OMP.
status: verified
verified_on: 2026-06-29
provenance: official-docs (omp.sh) via research subagent
---

# OMP (Oh My Pi) — Tool System

## Built-in tools

Full set (24): `ast_edit`, `ast_grep`, `bash`, `browser`, `debug`, `edit`, `eval`, `find`, `generate_image`, `github`, `inspect_image`, `irc`, `job`, `lsp`, `read`, `recipe`, `report_tool_issue`, `resolve`, `search`, `task`, `todo`, `web_search`, `write`.

### Enable/disable model

- `debug` is **off by default**; enable in `~/.omp/agent/config.yml`:
  ```yaml
  debug: { enabled: true }
  ```
- CLI: `--tools read,find,search` (restrict), `--no-lsp`.
- `tools.discoveryMode` controls MCP tool loading (see below).
- `tools.approvalMode` / `--approval-mode always-ask|write|yolo` / `--yolo` / `--auto-approve`.
- SDK: `createAgentSession({ toolNames: [...] })`.
- In-session: `/tools`.
- No per-tool allow/deny YAML except MCP `disabledServers`.

## Custom tools

**TypeScript factory** (default export). The host injects `pi.zod`.

| Scope | Path |
|---|---|
| User | `~/.omp/agent/tools/<name>/index.ts` |
| Project | `.omp/tools/<name>/index.ts` |
| Also read | `.claude/tools/<name>/index.ts`, `.codex/tools/<name>/index.ts` |

```typescript
import type { CustomToolFactory } from "@oh-my-pi/pi-coding-agent";

const factory: CustomToolFactory = (pi) => ({
  name: "repo_stats",
  label: "Repo Stats",
  description: "Count tracked files matching a glob",
  parameters: pi.zod.object({ glob: pi.zod.string().optional().default("**/*.ts") }),
  async execute(_toolCallId, params, onUpdate, _ctx, signal) {
    onUpdate?.({ content: [{ type: "text", text: `Listing ${params.glob}` }] });
    const result = await pi.exec("git", ["ls-files", params.glob], { signal, cwd: pi.cwd });
    if (result.code !== 0) throw new Error(result.stderr || "git ls-files failed");
    const files = result.stdout.split("\n").filter(Boolean);
    return { content: [{ type: "text", text: `Found ${files.length} files` }], details: { count: files.length } };
  },
});
export default factory;
```

Fields: `name` (req, no collision with built-ins/other customs), `label`, `description` (req), `parameters` (Zod via `pi.zod`), `execute(toolCallId, params, onUpdate, ctx, signal)`, optional `renderCall`/`renderResult`. Return `{ content, details, isError }` — `content` is model-visible, `details` is not. Name collisions are **rejected at load**; built-ins always win. Inspect with `omp -p '/extensions'`.

### Other authoring surfaces

- **Skills**: `<scope>/skills/<name>/SKILL.md`. Frontmatter `name` (opt), `description` (req), `hide` (opt). Non-recursive, one skill per dir. Disable via `--no-skills`, `skills.enabled:false`, `skills.ignoredSkills`/`includeSkills`.
- **Hooks**: `<scope>/hooks/pre/*.ts` (gate) + `hooks/post/*.ts` (rewrite). Events: `tool_call` (`{block,reason}`), `tool_result`, `context`, `session_before_compact`/`branch`/`switch`/`tree`, lifecycle events, `ttsr_triggered`. `HookAPI` vs full `ExtensionAPI`.
- **Prompt templates (slash commands)**: `<scope>/commands/<name>.md` or `<name>/index.ts`. Also reads `~/.claude/commands/`, `.claude/commands/`, `.codex/commands/`. Placeholders `$1`/`$2`/`$@`/`$ARGUMENTS`.
- **Subagents**: `<scope>/agents/<name>.md`. Frontmatter `name`/`description` (req), `tools`, `model`, `spawns`, `thinkingLevel`, `output`, `blocking`, `autoloadSkills`, `read-summarize`. Resolution: project → user → plugin → bundled.
- **TTSR rules**: `.omp/rules/<rule>.md`, `~/.omp/agent/rules/<rule>.md`. Frontmatter `condition` (regex) or `astCondition`, `scope`. Disable via `ttsr.disabledRules`.
- **Extension packages / plugins**: `package.json` with `omp.extensions[]` (legacy `pi.extensions`), or full plugin layout; `omp install ./my-extension`.

## MCP tools

Config priority (highest → lowest):

| Priority | Scope | Path |
|---|---|---|
| 1 | Project (omp-managed) | `.omp/mcp.json` |
| 2 | User (omp-managed) | `~/.omp/agent/mcp.json` |
| 3 | Other harnesses (auto-discovered) | `.claude/mcp.json`, `.cursor/mcp.json`, `.vscode/mcp.json`, `.gemini/mcp.json`, `.windsurf/mcp.json`, `opencode.json` |
| 4 | Repo-root standalone | `mcp.json` or `.mcp.json` |

Project entries shadow user entries by server key. Disable a server via `disabledServers` in the **user file only**.

### stdio

```json
{
  "$schema": "https://raw.githubusercontent.com/can1357/oh-my-pi/main/packages/coding-agent/src/config/mcp-schema.json",
  "mcpServers": {
    "fs": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "${HOME}/projects"],
      "env": { "LOG_LEVEL": "info" },
      "cwd": "${HOME}"
    }
  }
}
```

`type` defaults to `stdio` when `command` set. `${VAR}` / `${VAR:-default}` expand across `command`, `args`, `env`, `cwd`, `url`, `headers`, `auth`, `oauth`. A leading `!` in a header/env value runs a shell command and uses trimmed stdout.

### Streamable HTTP

```json
{
  "mcpServers": {
    "linear": { "type": "http", "url": "https://mcp.linear.app/sse",
      "headers": { "Authorization": "Bearer ${LINEAR_TOKEN}" } },
    "github": { "type": "http", "url": "https://api.githubcopilot.com/mcp/",
      "oauth": { "clientId": "${GH_CLIENT_ID}", "clientSecret": "${GH_CLIENT_SECRET}" } }
  }
}
```

OAuth tokens land in `~/.omp/agent/agent.db` (SQLite), not the JSON; complete the flow with `/mcp reauth <name>`. No separate `sse` type — `type:"http"` handles SSE endpoints.

### Tool names & discovery

Tool names: `mcp__<server>_<tool>` (lowercased, non-`[a-z_]`→`_`, collapsed, redundant `<server>_` prefix stripped once).

`tools.discoveryMode` in `config.yml`: `auto` (default — BM25 discovery past 40 tools), `mcp-only` (all MCP behind discovery), `off`, `all`.

### Management

Slash commands: `/mcp add|remove|enable|disable|test|reauth|unauth|reconnect|reload|list|resources|prompts|notifications|smithery-*`. Connections are parallel with a 250 ms fast-start gate; failures isolated per server; transports auto-reconnect with backoff.

## Config files & locations

| Scope | Path | Format |
|---|---|---|
| User settings | `~/.omp/agent/config.yml` | YAML |
| User MCP | `~/.omp/agent/mcp.json` | JSON (`disabledServers` here) |
| Credentials | `~/.omp/agent/agent.db` | SQLite |
| User custom tools | `~/.omp/agent/tools/<name>/index.ts` | TypeScript |
| User context | `~/.omp/agent/AGENTS.md` / `RULES.md` / `SYSTEM.md` / `APPEND_SYSTEM.md` | Markdown |
| Project MCP | `.omp/mcp.json` | JSON (highest priority) |
| Project custom tools | `.omp/tools/<name>/index.ts` | TypeScript |
| Project context | `<repo>/AGENTS.md`, `.omp/RULES.md`/`SYSTEM.md`/`APPEND_SYSTEM.md` | Markdown |

Overrides: `PI_CODING_AGENT_DIR` (moves `~/.omp/agent`), `PI_CONFIG_DIR` (renames `.omp`). Note: no project-level `config.yml`; project uses the global one.

## Gotchas

1. `type:"stdio"` is default when `command` set; if you set `type:"http"` but include `command`, http wins and command is ignored — be explicit.
2. Sanitized MCP tool-name collisions are last-write-wins.
3. `disabledServers` is user-file-only; adding it to `.omp/mcp.json` is [INFERENCE] silently ignored.
4. Built-in tools always win name collisions; custom `bash`/`read` rejected at load.
5. Malformed `config.yml` silently falls back to defaults — run `omp config list` after edits.
6. Hook + skill discovery is non-recursive.
7. `SYSTEM.md` removes built-in tool-usage guidance; prefer `APPEND_SYSTEM.md`.
8. OAuth creds in `agent.db` beat env vars; `/logout <provider>` to force env precedence.
9. `discoveryMode: mcp-only` hides MCP tools until searched.
10. Redundant `<server>_` prefix stripped once only — plan tool names accordingly.
11. Plugins run arbitrary TypeScript every turn — install only from trusted sources.
12. `!`-prefixed env/header that fails resolves to empty string, not an error.
13. omp-managed project file is `.omp/mcp.json`; repo-root `mcp.json`/`.mcp.json` is lowest priority and other-harness files can outrank it.

## Sources

- https://omp.sh/docs
- https://omp.sh/docs/tools
- https://omp.sh/docs/custom-tools
- https://omp.sh/docs/mcp
- https://omp.sh/docs/mcp-authoring
- https://omp.sh/docs/skills
- https://omp.sh/docs/hooks
- https://omp.sh/docs/settings
- https://omp.sh/docs/plugins
- https://omp.sh/docs/extension-authoring
- https://omp.sh/docs/context-files
- https://omp.sh/docs/prompt-templates
- https://omp.sh/docs/subagent-authoring
- https://omp.sh/docs/ttsr
- https://omp.sh/docs/env
- https://omp.sh/docs/secrets
- https://omp.sh/docs/providers
- https://omp.sh/docs/custom-models
- https://omp.sh/docs/cli
- https://omp.sh/docs/slash
- https://omp.sh/docs/sdk
