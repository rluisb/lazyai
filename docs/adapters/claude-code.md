# Claude Code Adapter

**Adapter ID:** `claude-code`  
**Source:** `packages/cli/internal/adapter/claudecode.go`  
**Status:** stable  
**Config directory:** `.claude`

## Overview

The Claude Code adapter generates native configuration for [Claude Code](https://docs.anthropic.com/en/docs/claude-code/overview) (the `claude` CLI). It emits agents, skills, commands, output styles, hooks, MCP configuration, and managed settings into `.claude/` and `.mcp.json`.

## Generated Files

| Path | Description |
|---|---|
| `.claude/agents/<name>.md` | Agent definitions (canonical agents with Claude Code frontmatter rewrite) |
| `.claude/skills/<name>/SKILL.md` | Skill directories |
| `.claude/commands/<name>.md` | Slash commands |
| `.claude/output-styles/<name>.md` | Output style definitions |
| `.claude/templates/<name>.md` | Speckit templates |
| `.claude/hooks/<name>.sh` | Hook scripts (block-destructive-shell, objective-workflow-gate) |
| `.claude/settings.json` | Managed settings with permissions, hooks config, MCP servers |
| `.claude/rules/typescript.md` | Sample rule |
| `.mcp.json` | Project MCP server configuration |
| `CLAUDE.md` | Root instructions |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.claude/agents/<name>.md` |
| Skills | dir-per-item | `.claude/skills/<name>/SKILL.md` |
| Templates | flat | `.claude/templates/<name>.md` |
| Commands | flat | `.claude/commands/<name>.md` |
| Output styles | flat | `.claude/output-styles/<name>.md` |
| Chat modes | none | — |
| Prompts | none | — |

## MCP Behavior

Claude Code MCP is compiled via `CompileMCPForTool`. At project scope, the adapter writes `.mcp.json`. At global scope, MCP servers are merged into `settings.json` (the `mcpServers` key). When `DriveCLI=true`, the adapter attempts CLI-driven registration via `claude mcp add-json` with silent fallback to direct-write.

## Hook Behavior

Hook scripts are copied from `claudecode/hooks/` in the library to `.claude/hooks/`. The adapter configures them in `settings.json` under `hooks.PreToolUse` and `hooks.Stop`. Two hooks are installed by default:
- `block-destructive-shell.sh` — PreToolUse matcher for Bash
- `objective-workflow-gate.sh` — Stop hook

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.claude/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.claude/agents/`. The adapter applies a Claude Code-specific frontmatter rewrite (`RewriteAgentForClaudeCode`) that strips LazyAI-specific metadata and emits only `name` and `description`. A default "guide" agent is always installed. At global scope, legacy flat-layout agents are migrated into the `agents/` subdirectory.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes |

## Headless Support

Yes (`CanRunHeadless() = true`). The adapter supports headless init and validation, including a post-install summary of registered tools.

## Known Limitations

- No chat modes concept
- No prompts directory; prompts ship as commands
- Global scope MCP is settings.json-only (no `.mcp.json`)

## Test Coverage

| Test file | What it verifies |
|---|---|
| `claudecode_frontmatter_test.go` | Agent frontmatter rewrite |
| `claudecode_global_layout_test.go` | Global scope layout |
| `claudecode_drivecli_test.go` | CLI-driven MCP registration |
| `claude_cli_test.go` | CLI interaction |
