# Pi Adapter

**Adapter ID:** `pi`  
**Source:** `packages/cli/internal/adapter/pi.go`  
**Status:** stable  
**Config directory:** `.pi`

## Overview

The Pi adapter generates native configuration for [Pi](https://pi.ai) (the `pi` CLI). It emits agents, skills, prompts, and TypeScript extensions into `.pi/`. Pi supports project, workspace, and global scopes.

## Generated Files

| Path | Description |
|---|---|
| `.pi/agents/<name>.md` | Agent definitions (canonical agents) |
| `.pi/skills/<name>/SKILL.md` | Skill directories |
| `.pi/prompts/<name>.md` | Prompt templates |
| `.pi/extensions/<name>.ts` | TypeScript extensions (safety hooks) |
| `.pi/settings.json` | Global/project settings |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.pi/agents/<name>.md` |
| Skills | dir-per-item | `.pi/skills/<name>/SKILL.md` |
| Prompts | flat | `.pi/prompts/<name>.md` |
| Templates | none | — |
| Commands | none | — |
| Chat modes | none | — |
| Output styles | none | — |

## MCP Behavior

**Pi MCP is a no-op.** The `CompileMCP` method returns `ctx.FileRecords` unchanged (`pi.go:81-83`). The adapter declares MCP capability in its `Capabilities()` for future compatibility, but no MCP configuration is emitted for Pi. Pi has no native MCP surface.

## Hook Behavior

Pi has no `.pi/hooks` path. Safety hooks ship as TypeScript extensions at `.pi/extensions/`. The adapter copies extension files from `pi/extensions/` in the library.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.pi/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.pi/agents/`. Pi's subagent extension reads agent definitions from this directory.

## Prompt Behavior

Prompt templates are copied as flat markdown files to `.pi/prompts/`. The adapter supports selection filtering.

## Scope Support

| Scope | Supported |
|---|---|
| Global | yes |

## Headless Support

No (`CanRunHeadless() = false`).

## Known Limitations

- **MCP is a no-op** — no MCP configuration is emitted despite the capability declaration
- No templates, commands, or chat modes
- No `.pi/hooks` path; hooks ship as TypeScript extensions
- Pi project trust is not a sandbox (warned at install time)

## Test Coverage

| Test file | What it verifies |
|---|---|
| `pi_adapter_test.go` | Agents, skills, prompts, extensions; confirms no `.pi/hooks` path |
| `adapter_adapters_test.go` | Full install from FS |
