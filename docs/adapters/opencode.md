# OpenCode Adapter

**Adapter ID:** `opencode`  
**Source:** `packages/cli/internal/adapter/opencode.go`  
**Status:** stable  
**Config directory:** `.opencode`

## Overview

The OpenCode adapter generates native configuration for the [OpenCode](https://github.com/sst/opencode) CLI. It emits agents, skills, commands, chat modes, MCP configuration, hooks, and plugins into `.opencode/`.

## Generated Files

| Path | Description |
|---|---|
| `.opencode/agents/<name>.md` | Agent definitions (canonical agents with frontmatter rewrite) |
| `.opencode/skills/<name>/SKILL.md` | Skill directories |
| `.opencode/commands/<name>.md` | Slash commands |
| `.opencode/modes/<name>.md` | Chat modes |
| `.opencode/templates/<name>.md` | Speckit templates |
| `.opencode/lazyai.mcp.jsonc` | Legacy MCP config (migration) |
| `.opencode/mcp.json` | MCP server configuration |
| `.opencode/plugins/` | Plugin hooks |
| `opencode.json` | Root config with instructions, agents, permissions |
| `AGENTS.md` | Root instructions |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.opencode/agents/<name>.md` |
| Skills | dir-per-item | `.opencode/skills/<name>/SKILL.md` |
| Templates | flat | `.opencode/templates/<name>.md` |
| Commands | flat | `.opencode/commands/<name>.md` |
| Chat modes | flat | `.opencode/modes/<name>.md` |
| Output styles | none | — |
| Prompts | none | — |

## MCP Behavior

OpenCode MCP is compiled via `CompileMCPForTool`. The adapter writes to `.opencode/mcp.json` (new) and maintains a legacy `.opencode/lazyai.mcp.jsonc` for backward compatibility. User-authored servers are preserved across re-runs.

## Hook Behavior

Hooks are installed as OpenCode plugins under `.opencode/plugins/`. The adapter shells out to `opencode plugin <module>` for each selected plugin when the binary is on PATH.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.opencode/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.opencode/agents/`. The adapter applies frontmatter rewrite for OpenCode compatibility. A default "guide" agent is always installed.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes |

## Headless Support

Yes (`CanRunHeadless() = true`). The adapter supports headless init and validation.

## Known Limitations

- No prompts directory; prompts ship as commands
- No output-styles concept
- Legacy `lazyai.mcp.jsonc` maintained during migration

## Test Coverage

| Test file | What it verifies |
|---|---|
| `opencode_adapter_test.go` | Install from FS, global scope, preserves root json, commands+modes, selection filters, instructions key resolution, default agent, skill surface, package.json |
| `opencode_frontmatter_test.go` | Frontmatter rewrite for OpenCode agents |
| `opencode_validate_test.go` | Validation of generated opencode.json |
| `opencode_plugin_test.go` | Plugin installation |
