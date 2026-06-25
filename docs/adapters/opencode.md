# OpenCode Adapter

**Adapter ID:** `opencode`  
**Source:** `packages/cli/internal/adapter/opencode.go`  
**Status:** stable  
**Config directory:** `.opencode`

## Overview

The OpenCode adapter generates native configuration for the [OpenCode](https://github.com/sst/opencode) CLI. It emits agents, skills, commands, OpenCode-specific mode files, MCP configuration, hooks, and plugins into `.opencode/`.

## Generated Files

| Path | Description |
|---|---|
| `.opencode/agents/<name>.md` | Agent definitions (canonical agents with frontmatter rewrite) |
| `.opencode/skills/<name>/SKILL.md` | Skill directories |
| `.opencode/commands/<name>.md` | Slash commands |
| `.opencode/modes/<name>.md` | OpenCode-specific mode files |
| `.opencode/lazyai.mcp.jsonc` | Legacy MCP config (migration input only; not written by current compiler) |
| `.opencode/plugins/` | Plugin hooks |
| `opencode.json` | Root config with instructions, permissions, and MCP entries |
| `AGENTS.md` | Root instructions |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.opencode/agents/<name>.md` |
| Skills | dir-per-item | `.opencode/skills/<name>/SKILL.md` |
| Templates | none | — |
| Commands | flat | `.opencode/commands/<name>.md` |
| OpenCode modes | flat | `.opencode/modes/<name>.md` |
| Output styles | none | — |
| Prompts | none | — |

## MCP Behavior

OpenCode MCP is compiled via `CompileMCPForTool`. The adapter merges MCP servers into the root `opencode.json` config file. Legacy `.opencode/lazyai.mcp.jsonc` entries are read as migration input only; the current compiler does not write that file. User-authored servers are preserved across re-runs.

## Hook Behavior

Hooks are installed as OpenCode plugins under `.opencode/plugins/`. The adapter copies managed plugin files from the LazyAI library into `.opencode/plugins/` and does NOT shell out to install external plugins.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.opencode/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Mode Behavior

Mode files are written under `.opencode/modes/<name>.md`. Each file uses current OpenCode frontmatter: `mode: primary` declares the mode type and `permission:` (with `edit`/`bash`/`webfetch` set to `deny`) replaces the deprecated `tools:` boolean map. LazyAI currently treats this as an OpenCode-specific surface rather than a cross-tool abstraction.

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
- No documented template surface

## Test Coverage

| Test file | What it verifies |
|---|---|
| `opencode_adapter_test.go` | Install from FS, global scope, preserves root json, commands+modes, selection filters, instructions key resolution, default agent, skill surface, package.json |
| `opencode_frontmatter_test.go` | Frontmatter rewrite for OpenCode agents |
| `opencode_validate_test.go` | Validation of generated opencode.json |
| `opencode_plugin_test.go` | Plugin installation |
| `internal/library/integration_test.go` | #561: mode files contain `permission:` key; deprecated `tools:` map is absent |
