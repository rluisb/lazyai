# OMP Adapter

**Adapter ID:** `omp`  
**Source:** `packages/cli/internal/adapter/omp.go`  
**Status:** **beta**  
**Config directory:** `.omp`

## Overview

The OMP adapter generates native configuration for [OMP](https://ohmyposh.dev) (Oh My Posh / OMP CLI). It emits agents, skills, commands, prompts, hooks, and MCP configuration into `.omp/`.

**Beta status:** OMP is marked beta because its partially JS-rendered official documentation has not been fully snapshot-verified (matrix §1, EC-006). The adapter is functional and tested, but the compliance surface may shift as official docs are fully captured.

## Generated Files

| Path | Description |
|---|---|
| `.omp/agents/<name>.md` | Task agent definitions (canonical agents) |
| `.omp/skills/<name>/SKILL.md` | Skill directories |
| `.omp/commands/<name>.md` | Slash commands |
| `.omp/prompts/<name>.md` | Prompt templates |
| `.omp/hooks/pre/<name>.ts` | TypeScript hook factories |
| `.omp/mcp.json` | MCP server configuration |
| `AGENTS.md` | Root instructions |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.omp/agents/<name>.md` |
| Skills | dir-per-item | `.omp/skills/<name>/SKILL.md` |
| Commands | flat | `.omp/commands/<name>.md` |
| Prompts | flat | `.omp/prompts/<name>.md` |
| Templates | none | — |
| Chat modes | none | — |
| Output styles | none | — |

## MCP Behavior

OMP MCP is compiled via `CompileMCPForTool`. The adapter writes to `.omp/mcp.json` with standard compile path.

## Hook Behavior

OMP discovers TypeScript hook factories from `.omp/hooks/pre/*.ts`. The adapter copies hook scripts from the canonical library to this directory.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.omp/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.omp/agents/`. OMP reads task agents from this directory.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes |

## Headless Support

No (`CanRunHeadless() = false`).

## Known Limitations

- **Beta status** — compliance surface may shift as official docs are fully snapshot-verified
- No templates, chat modes, or output styles
- No headless support

## Test Coverage

| Test file | What it verifies |
|---|---|
| `omp_adapter_test.go` | Agents + skills, commands + prompts, hooks, global scope, MCP compile |
