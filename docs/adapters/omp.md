# OMP Adapter

**Adapter ID:** `omp`  
**Source:** `packages/cli/internal/adapter/omp.go`  
**Status:** **stable**  
**Config directory:** `.omp`

## Overview

The OMP adapter generates native configuration for [OMP (Oh My Pi)](https://github.com/can1357/oh-my-pi), the AI coding-agent harness. It emits agents, skills, commands, prompts, hooks, and MCP configuration into `.omp/`.

**Verification:** every emitted surface is source-verified against the authoritative OMP documentation (the in-harness `omp://` docs set). See [Beta adapter verification 2026-06](snapshots/beta-adapter-verification-2026-06.md). OMP was promoted from beta to stable on 2026-06-23 (#486).

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

- No templates, chat modes, or output styles
- No headless support

## Test Coverage

| Test file | What it verifies |
|---|---|
| `omp_adapter_test.go` | Agents + skills, commands + prompts, hooks, global scope, MCP compile |