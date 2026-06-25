# GitHub Copilot Adapter

**Adapter ID:** `copilot`  
**Source:** `packages/cli/internal/adapter/copilot.go`  
**Status:** stable  
**Config directory:** `.github`

## Overview

The GitHub Copilot adapter generates native configuration for [GitHub Copilot](https://github.com/features/copilot) across IDE, CLI, and cloud surfaces. It emits agents, skills, instructions, prompts, hooks, and MCP configuration into `.github/` and `.vscode/`. All agent output — both canonical agents and library chatmode sources — lands at `.github/agents/<name>.agent.md`. The legacy `.github/chatmodes/` path is not used.

## Generated Files

| Path | Description |
|---|---|
| `.github/agents/<name>.agent.md` | Custom agent definitions (canonical agents rendered to Copilot `.agent.md` format) |
| `.github/skills/<name>/SKILL.md` | Skill directories |
| `.github/instructions/<name>.instructions.md` | Path-specific instructions (from templates) |
| `.github/prompts/<name>.prompt.md` | Prompt templates with `.prompt.md` extension |
| `.github/hooks/<name>.json` | Hook configuration (project scope only) |
| `.github/hooks/<name>.sh` | Hook scripts (project scope only) |
| `.vscode/mcp.json` | VS Code MCP server configuration |
| `~/.copilot/mcp-config.json` | Copilot CLI MCP configuration (probe-gated); remote servers use `http` transport |
| `.github/copilot-instructions.md` | Repository instructions |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.github/agents/<name>.agent.md` (canonical agents + library chatmode sources combined) |
| Skills | dir-per-item | `.github/skills/<name>/SKILL.md` |
| Templates | flat | `.github/instructions/<name>.instructions.md` |
| Prompts | rewrite-ext | `.github/prompts/<name>.prompt.md` |
| Commands | none | — |
| Output styles | none | — |

## MCP Behavior

Copilot MCP is compiled via `CompileMCPForTool` with dual output:
- **VS Code:** `.vscode/mcp.json` — written at project/workspace scope unconditionally
- **CLI:** `~/.copilot/mcp-config.json` — written when the Copilot CLI probe passes (binary found or `~/.copilot/` exists); deep-merge preserves user-authored servers; remote servers serialized with `type: "http"` (SSE deprecated)

At global scope, only the CLI output is written (no project directory for `.vscode/mcp.json`).

## Hook Behavior

Hooks are installed at project/workspace scope only. JSON and shell hook files are copied from `copilot/hooks/` in the library to `.github/hooks/`. Global Copilot support is agent/instructions/MCP oriented; no verified user-scope hook surface is emitted there.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.github/skills/<name>/SKILL.md`. The adapter also cleans up legacy skill-as-agent outputs from previous installs. Skill tier metadata (Frontier, Speed, Balanced) is resolved against `CopilotCatalog` and embedded in the agent markdown.

## Agent Behavior

Canonical agents are rendered into Copilot's `.agent.md` format via `copilotAgentMarkdownContent`. The adapter transforms canonical frontmatter into Copilot-compatible agent definitions with model resolution based on skill tier.

## Prompt Behavior

Prompts are copied with a `.prompt.md` extension rewrite (e.g. `plan.md` → `plan.prompt.md`). The adapter supports selection filtering.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes (probes for Copilot CLI or `~/.copilot/`) |

## Headless Support

Yes (`CanRunHeadless() = true`). The adapter supports headless init.

## Known Limitations

- No subagents concept
- No slash commands surface
- Global scope hooks are not emitted (no verified user-scope hook surface)
- Dual MCP output (VS Code + CLI) adds complexity

## Test Coverage

| Test file | What it verifies |
|---|---|
| `copilot_chatmodes_test.go` | Custom agent installation at `.github/agents/`; verifies deprecated `.github/chatmodes/` path is never written |
| `copilot_skill_tier_test.go` | Skill tier resolution (Frontier, Speed, Balanced) |
| `adapter_adapters_test.go` | Full install from FS |
