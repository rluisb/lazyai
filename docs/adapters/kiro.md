# Kiro Adapter

**Adapter ID:** `kiro`  
**Source:** `packages/cli/internal/adapter/kiro.go`  
**Status:** stable  
**Config directory:** `.kiro`

## Overview

The Kiro adapter generates native configuration for [Kiro](https://kiro.dev) (Kiro IDE/CLI). It emits agents, skills, prompts, hooks, MCP configuration, and permissions into `.kiro/`.

## Generated Files

| Path | Description |
|---|---|
| `.kiro/agents/<name>.md` | Custom agent profiles (canonical agents with frontmatter) |
| `.kiro/skills/<name>/SKILL.md` | Skill directories |
| `.kiro/prompts/<name>.md` | Prompt templates |
| `.kiro/settings/mcp.json` | MCP server configuration |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.kiro/agents/<name>.md` |
| Skills | dir-per-item | `.kiro/skills/<name>/SKILL.md` |
| Prompts | flat | `.kiro/prompts/<name>.md` |
| Templates | none | — |
| Commands | none | — |
| Chat modes | none | — |
| Output styles | none | — |

## MCP Behavior

Kiro MCP is compiled via `CompileMCPForTool`. The adapter writes to `.kiro/settings/mcp.json` with standard compile path.

## Hook Behavior

Kiro supports hooks via the capability model (`capabilities.go:170`: `Hooks: true`), but the adapter does not install hook scripts directly. Hook configuration is handled through the standard compile path.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.kiro/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.kiro/agents/`. A default "guide" agent is always installed. Kiro CLI v3 discovers custom agent profiles from this directory. Agent files include frontmatter with `description` and markdown body.

## Prompt Behavior

Prompt templates are copied as flat markdown files to `.kiro/prompts/`. The adapter supports selection filtering.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes |

## Headless Support

No (`CanRunHeadless() = false`).

## Known Limitations

- **No specs or steering** — Kiro does not emit native specs or steering files (`capabilities_test.go:68-69`). The adapter installs agents, skills, prompts, hooks, MCP, permissions, and global config, but specs and steering are intentionally absent.
- No templates, commands, chat modes, or output styles
- No headless support
- No plugin surface

## Test Coverage

| Test file | What it verifies |
|---|---|
| `kiro_adapter_test.go` | Agent profiles, skills, prompts; confirms no `.kiro/workflows` |
| `adapter_adapters_test.go` | Full install from FS |
