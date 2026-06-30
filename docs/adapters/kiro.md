# Kiro Adapter

**Adapter ID:** `kiro`  
**Source:** `packages/cli/internal/adapter/kiro.go`  
**Status:** stable  
**Config directory:** `.kiro`

## Overview

The Kiro adapter generates native configuration for [Kiro CLI v3](https://kiro.dev/docs/cli/v3/agent-config). It emits agents, skills, prompts, native Kiro v3 hook JSON, MCP configuration, and permissions/global-configuration metadata into `.kiro/`.

**Target: Kiro CLI.** All paths and schemas are verified against the `kiro.dev/docs/cli/` tree (not the Kiro IDE docs). The CLI and IDE are distinct product surfaces.

## Generated Files

| Path | Description |
|---|---|
| `.kiro/agents/<name>.json` | Custom agent profiles (JSON; required format per official Kiro CLI v3 docs) |
| `.kiro/skills/<name>/SKILL.md` | Skill directories |
| `.kiro/prompts/<name>.md` | Prompt templates |
| `.kiro/hooks/<name>.json` | Native Kiro v3 hook descriptors |
| `.kiro/hooks/*.sh` | Referenced shell hook scripts when a hook action runs a command |
| `.kiro/settings/mcp.json` | MCP server configuration |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.kiro/agents/<name>.json` |
| Skills | dir-per-item | `.kiro/skills/<name>/SKILL.md` |
| Prompts | flat | `.kiro/prompts/<name>.md` |
| Templates | none | — |
| Commands | none | — |
| Chat modes | none | — |
| Output styles | none | — |

## MCP Behavior

Kiro MCP is compiled via `CompileMCPForTool`. The adapter writes to `.kiro/settings/mcp.json` using a dedicated `toKiroMcp()` serializer. Local (stdio) servers emit `{command, args, env}`; remote servers emit `{url, headers}` — matching Kiro's documented schema exactly (`kiro.dev/docs/mcp/configuration`). No `type` field is emitted for remote entries.

## Hook Behavior

Kiro emits native hook JSON files at `.kiro/hooks/<name>.json` using the Kiro CLI v3 hook schema. Only source-verified trigger mappings are emitted; currently the adapter emits `block-destructive-shell` with `PreToolUse` and `matcher: "shell"`. Referenced shell scripts are installed alongside the JSON file.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.kiro/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are transformed to `.kiro/agents/<name>.json` JSON files. The JSON format is required by Kiro CLI v3 (`.md` is not recognized). A default "guide" agent is always installed. Kiro CLI v3 discovers agent profiles from `.kiro/agents/` (workspace) and `~/.kiro/agents/` (global) per `kiro.dev/docs/cli/custom-agents/`. Each emitted file contains `name`, `description`, `tools`, `allowedTools`, and `prompt` fields; `tools` and `allowedTools` are populated from the canonical agent `tools:` frontmatter via `frontmatter.ParseAgentToolGrants`.

## Prompt Behavior

Prompt templates are copied as flat markdown files to `.kiro/prompts/`. Kiro CLI stores reusable prompts here and references them via `@name` (`kiro.dev/docs/cli/chat/manage-prompts`). The adapter supports selection filtering.

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
| `mcp_compiler_test.go` | #556: `KiroRemoteOmitsTypeField` — remote MCP entry has `url`+`headers`, no `type`; `KiroLocalUsesCommandArgs` |
