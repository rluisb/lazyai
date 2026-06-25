# Antigravity / Gemini Adapter

**Adapter ID:** `antigravity`  
**Source:** `packages/cli/internal/adapter/antigravity.go`  
**Status:** **stable**
**Config directory:** `.gemini`

## Overview

The Antigravity adapter generates native configuration for the Antigravity/Gemini CLI surface. It emits skills, hooks, settings, and a user-level MCP configuration into `.gemini/`, `.agents/`, and `~/.gemini/config/`.

**Stable (verified 2026-06-23, #486):** the Antigravity IDE and Gemini CLI docs are JS-rendered and were snapshot-verified by rendering — see [Beta adapter verification 2026-06](snapshots/beta-adapter-verification-2026-06.md). All emitted surfaces (workspace + global skills, IDE + CLI hooks, MCP, root instructions) are verified against official docs. The two former beta gaps were closed and pinned by conformance tests: global-scope skills now write `~/.gemini/config/skills/`, and root instructions are discovered via a generated `GEMINI.md` (imports `@./AGENTS.md`, for Gemini CLI) and `.agents/rules/lazyai.md` (imports `@/AGENTS.md`, for Antigravity IDE workspaces). Global rules (`~/.gemini/GEMINI.md`) remain user-managed.

## Generated Files

| Path | Description |
|---|---|
| `.gemini/settings.json` | Gemini CLI settings; the adapter merges a `hooks` block (event-keyed `BeforeTool`/`AfterAgent` entries pointing at `.gemini/hooks/lazyai/*.sh`) via deep-merge, preserving any user-owned keys |
| `.gemini/hooks/lazyai/<name>.sh` | Hook scripts |
| `.agents/skills/<name>/SKILL.md` | Agent Skills (workspace; global installs use `~/.gemini/config/skills/<name>/SKILL.md`) |
| `.agents/hooks.json` | Antigravity IDE hook event configuration (workspace/project; global installs use `~/.gemini/config/hooks.json`) |
| `~/.gemini/config/mcp_config.json` | Antigravity desktop-IDE MCP config (`serverUrl` key); merged additively for IDE compatibility |
| `.gemini/settings.json` `mcpServers` block | Gemini CLI MCP entries (`httpUrl` key); merged into the scope-appropriate settings file alongside hook config |
| `AGENTS.md` | Root instructions |
| `.agents/rules/lazyai.md` | Antigravity IDE workspace rule importing `@/AGENTS.md` |
| `GEMINI.md` | Gemini CLI context file importing `@./AGENTS.md` |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | **none** | — |
| Skills | dir-per-item | `.agents/skills/<name>/SKILL.md` |
| Templates | none | — |
| Commands | none | — |
| Chat modes | none | — |
| Output styles | none | — |
| Prompts | none | — |

**Note:** Antigravity does not emit agent files (`output_mapping.go:346-350`: `ShapeNone` for agents). Skills are written to `.agents/skills/` (not `.gemini/skills/`).

## MCP Behavior

Antigravity MCP is compiled via `CompileMCPForTool` and writes to **two targets** so both Antigravity desktop-IDE and open-source Gemini CLI users get working MCP:

- **`~/.gemini/config/mcp_config.json`** — Antigravity desktop-IDE format. Remote servers use the `serverUrl` key. Merged additively; existing entries are preserved.
- **`.gemini/settings.json` (`mcpServers` block)** — Gemini CLI format (`settings.schema.json`). Remote servers use the `httpUrl` key (Streamable HTTP). Merged alongside hook config via deep-merge.

Local (stdio) servers use `command`/`args`/`env` in both outputs.

## Hook Behavior

Hook scripts are copied to `.gemini/hooks/lazyai/`. Hook event configuration is written to `.agents/hooks.json` (workspace/project) or `~/.gemini/config/hooks.json` (global) with references to the `.gemini/hooks/lazyai/` script paths. The adapter copies hooks from `antigravity/hooks/` in the library.

## Skill Behavior

Skills are written as Agent Skills-compatible directories at `.agents/skills/<name>/SKILL.md` (not under `.gemini/`). The adapter copies selected skills from the canonical library.

## Agent Behavior

Antigravity does not emit agent files. The adapter has no agents surface.

## Scope Support

| Scope | Supported |
|---|---|
| Global | yes |

## Headless Support

No (`CanRunHeadless() = false`).

## Upgrading from Beta

If you compiled Antigravity integration while it was a beta adapter (before the stable promotion on 2026-06-23), your global-scope skills were written to `~/.agents/skills/`. 

As a stable adapter, global skills are now correctly written to `~/.gemini/config/skills/`. 
Re-running `lazyai compile` will write the skills to the new path. You can safely delete the old files from `~/.agents/skills/` if you do not use other AI coding tools like Kiro or OpenCode.

## Known Limitations

- **No agents** — the adapter does not emit agent files (`output_mapping.go`: `ShapeNone`)
- No templates, commands, chat modes, output styles, or prompts
- Workspace skills written to `.agents/skills/`; global skills to `~/.gemini/config/skills/` — never `.gemini/skills/`
- Hook event config written to `.agents/hooks.json` (workspace/project) or `~/.gemini/config/hooks.json` (global) — never `~/.agents/hooks.json`
- Global rules (`~/.gemini/GEMINI.md`) are user-managed and never written by the adapter (same policy as Claude's global `CLAUDE.md`)

## Test Coverage

| Test file | What it verifies |
|---|---|
| `antigravity_install_test.go` | Skills at `.agents/skills/`, hooks at `.agents/hooks.json`, settings at `.gemini/settings.json`, workspace root resolution, minimal surface |
| `antigravity_install_test.go` | #486 gap fixes: global skills at `~/.gemini/config/skills/`, `.agents/rules/lazyai.md` importing `@/AGENTS.md`; #497 global hooks at `~/.gemini/config/hooks.json` |
| `capabilities_test.go` | Antigravity is stable; no beta adapters remain |
| `scaffold/root_test.go` | Generated `GEMINI.md` imports `@./AGENTS.md` and is tracked |
| `antigravity_install_test.go` | #554: `RemoteServerGeminiCLIDiscoverable` — remote MCP lands in `settings.json` with `httpUrl`; `HTTPUsesServerUrl` — desktop `mcp_config.json` retains `serverUrl` |
