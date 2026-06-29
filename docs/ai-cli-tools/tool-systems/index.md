---
title: AI CLI Tool Systems — Reference
summary: How built-in tools, custom tools, and MCP tools work inside five researched upstream AI CLIs (Kiro, Claude Code, Antigravity, Pi, OMP).
status: verified
verified_on: 2026-06-29
scope: upstream tool behavior (NOT LazyAI compile output)
applies_to: [kiro, claude-code, antigravity, pi, omp]
---

# AI CLI Tool Systems — Reference

This section documents how **tools** work *inside each of five researched upstream AI CLIs*, in three categories:

1. **Built-in tools** — native tools the agent ships with, and how they are enabled/disabled/permissioned.
2. **Custom tools** — how a user defines their own tools/commands/functions (file formats, schemas, discovery paths).
3. **MCP tools** — how Model Context Protocol servers are configured and consumed.

!!! warning "Coverage scope"
    This batch covers **five** of LazyAI's seven supported targets: **Kiro, Claude Code, Antigravity, Pi, OMP**. **OpenCode** and **Copilot** are supported compile targets but were **not** part of this research set — do not assume their tool/MCP semantics from these pages.

It exists so LazyAI install/compile work never has to make a web request to recall a config key, transport name, file path, or naming rule. It is **upstream behavior**, distinct from the LazyAI-compile-focused pages in this same directory (`claude-code.md`, `pi.md`, etc.).

!!! info "Provenance"
    Verified **2026-06-29** against official documentation. The **Claude Code** page was read directly from `code.claude.com/docs` in-session. The **Kiro, Antigravity, Pi, OMP** pages were gathered by parallel research subagents from each tool's official docs; every claim carries the source URL on its page. Anything not directly observed is marked `[INFERENCE]`.

## Per-tool pages

| Tool | Custom-tool model | Native MCP? | Page |
|---|---|---|---|
| Kiro CLI | Custom agents (JSON) + hooks — **no function-tool API** | ✅ JSON | [kiro.md](kiro.md) |
| Claude Code | Skills/commands + subagents + hooks — **"use MCP" for tools** | ✅ JSON | [claude-code.md](claude-code.md) |
| Antigravity | Skills + plugins + sidecars + hooks; SDK Python callables | ✅ **two files** | [antigravity.md](antigravity.md) |
| Pi | **TypeScript `registerTool()` extensions** | ❌ **none** | [pi.md](pi.md) |
| OMP | **TypeScript factory tools** | ✅ JSON | [omp.md](omp.md) |

**Cross-CLI agent-tools mapping:** see the [Agent Tools Matrix](agent-tools-matrix.md) — how a canonical agent's tool capability maps onto each target's native per-agent tool model, plus current LazyAI gap status.

## Cross-tool matrix — MCP configuration

| Tool | MCP file(s) | Remote-URL field | stdio | Disable mechanism | Tool-name form |
|---|---|---|---|---|---|
| Kiro | `.kiro/settings/mcp.json`, `~/.kiro/settings/mcp.json` | `url` (+ optional `type:"http"`) | ✅ | `disabled` / `disabledTools` | `^[a-zA-Z][a-zA-Z0-9_]*$`, ≤64 |
| Claude | `.mcp.json` (project), `~/.claude.json` (user/local) | `url` + `type`: `http`/`streamable-http`/`sse`/`ws` | ✅ | `deny` rules / managed allow-deny | `mcp__<server>__<tool>` |
| Antigravity | `mcp_config.json` **AND** `settings.json` | **`serverUrl`** (native) / **`httpUrl`** (Gemini CLI) | ✅ | `disabled` / `disabledTools` + `mcp()` perms | engine `server/tool` |
| Pi | **none** | — | extension code only | — | — |
| OMP | `.omp/mcp.json`, `~/.omp/agent/mcp.json` | `url` (`type:"http"`) | ✅ | `disabledServers` (user file only) | `mcp__<server>_<tool>` |

## Cross-tool matrix — config roots

| Tool | User root | Project root |
|---|---|---|
| Kiro | `~/.kiro/` | `.kiro/` |
| Claude | `~/.claude/` (+ `~/.claude.json`) | `.claude/` (+ `.mcp.json`) |
| Antigravity | `~/.gemini/config/` (+ `~/.gemini/`) | `.agents/` (+ `.gemini/`) |
| Pi | `~/.pi/agent/` | `.pi/` |
| OMP | `~/.omp/agent/` | `.omp/` |

## Install-critical findings (top misconfiguration risks)

1. **Pi has no MCP system.** Never emit any `mcp.json` for the Pi target — Pi ships no MCP client. MCP must be a hand-written TypeScript extension or omitted entirely.
2. **Antigravity needs two MCP files with different field names.** Remote servers use **`serverUrl`** in `mcp_config.json` and **`httpUrl`** in `settings.json`. Using `url`/`httpUrl` in `mcp_config.json` is silently ignored.
3. **Antigravity global paths diverge from `~/.agents/`.** Global skills/hooks live under `~/.gemini/config/`; the IDE reads workspace rules from `.agents/rules/*.md`, not a bare root `AGENTS.md`.
4. **Kiro drops hyphenated MCP tool names** (must match `^[a-zA-Z][a-zA-Z0-9_]*$`). Kiro `oauthScopes` is **top-level**, not nested in `oauth`. `includeMcpJson:false` silently drops all `mcp.json` servers.
5. **Claude scope names changed.** `local` (was `project`) vs `project` (the shared `.mcp.json`) vs `user` (was `global`). Project `.mcp.json` servers start as `⏸ Pending approval`. `${CLAUDE_PROJECT_DIR}` in non-plugin configs needs a default (`${CLAUDE_PROJECT_DIR:-.}`).
6. **OMP `disabledServers` is user-file-only**; built-in tools always win name collisions; `discoveryMode: mcp-only` hides MCP tools until searched.
7. **Pi project resources are trust-gated.** Non-interactive runs need `defaultProjectTrust: "always"`. Global settings require the `agent/` subdir: `~/.pi/agent/settings.json`.
8. **"Custom tools" means different things per tool.** Only Pi and OMP expose real function-call custom tools (TypeScript). Kiro / Claude / Antigravity route "custom" capability through MCP, skills, subagents, or SDK callables — emitting a tool-schema file for them is meaningless.

## How to use this offline

- Read any page with your file tools (e.g. `read docs/ai-cli-tools/tool-systems/kiro.md`). No network needed.
- Each page ends with a **Sources** list of the exact official doc URLs, for re-verification when a tool ships a breaking change.
- To refresh: re-run the per-tool research against the Sources URLs and bump `verified_on` in the frontmatter.
