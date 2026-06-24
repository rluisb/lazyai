# Kiro Adapter

**Adapter ID:** `kiro`  
**Source:** `packages/cli/internal/adapter/kiro.go`  
**Status:** stable  
**Config directory:** `.kiro`

## Overview

The Kiro adapter generates native configuration for [Kiro](https://kiro.dev) (Kiro IDE/CLI). It emits agents, skills, prompts, native Kiro v3 hook JSON, MCP configuration, and permissions/global-configuration metadata into `.kiro/`.

## Generated Files

| Path | Description |
|---|---|
| `.kiro/agents/<name>.md` | Custom agent profiles (canonical agents with frontmatter) |
| `.kiro/skills/<name>/SKILL.md` | Skill directories |
| `.kiro/prompts/<name>.md` | Prompt templates |
| `.kiro/hooks/<name>.json` | Native Kiro v3 hook descriptors |
| `.kiro/hooks/*.sh` | Referenced shell hook scripts when a hook action runs a command |
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

Kiro emits native hook JSON files at `.kiro/hooks/<name>.json` using the Kiro CLI v3 hook schema. Only source-verified trigger mappings are emitted; currently the adapter emits `block-destructive-shell` with `PreToolUse` and `matcher: "shell"`. Referenced shell scripts are installed alongside the JSON file.

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

- **No specs or steering** — Kiro does not emit native specs or steering files; specs are user-authored workflow artifacts and steering remains unimplemented.
- **No repo-local permissions file** — Kiro docs forbid cloned repos from injecting permission rules; `Permissions: true` is host-support metadata, not an emitted `.kiro/permissions.yaml`.
- **No direct `.kiro/powers/` output** — Powers remain a future importable-package direction.
- No templates, commands, chat modes, or output styles
- No headless support
- No plugin surface

## Test Coverage

| Test file | What it verifies |
|---|---|
| `kiro_adapter_test.go` | Agent profiles, skills, prompts; confirms no `.kiro/workflows` |
| `adapter_adapters_test.go` | Full install from FS |
