# Adapter Capability Matrix

> **Source of truth:** `packages/cli/internal/adapter/capabilities.go` (declarative `Capability` struct per adapter) and `packages/cli/internal/adapter/output_mapping.go` (per-tool asset-kind output targets).  
> **Verification date:** 2026-06-23.

## 1. Support Levels

| Level | Meaning | Adapters |
|---|---|---|
| **stable** | Official docs verified + golden tests + smoke tests | OpenCode, Claude Code, Copilot, Pi, Kiro, OMP, Antigravity |
| **beta** | Official docs verified + golden tests, limited runtime smoke | — |
| **experimental** | Docs partially verified or host tool still moving quickly | — |
| **deprecated** | Kept for migration only | — |

**OMP promotion (2026-06-23, #486):** every emitted OMP surface was verified against the authoritative OMP (Oh My Pi) docs (`omp://`), clearing the docs-snapshot blocker. See [snapshots/beta-adapter-verification-2026-06.md](snapshots/beta-adapter-verification-2026-06.md).

**Antigravity promotion (2026-06-23, #486):** the Antigravity IDE / Gemini CLI docs are JS-rendered and were snapshot-verified by rendering. All emitted surfaces are verified and the two former beta gaps are closed and pinned by conformance tests — global-scope skills now write `~/.gemini/config/skills/`, and root instructions are discovered (`GEMINI.md` for Gemini CLI, `.agents/rules/lazyai.md` for Antigravity IDE). No adapter remains below stable (EC-006 cleared).

## 2. Capability Matrix

| Capability | OpenCode | Claude Code | Copilot | Pi | OMP | Antigravity | Kiro |
|---|---|---|---|---|---|---|---|
| **Support level** | stable | stable | stable | stable | stable | stable | stable |
| Root instructions | yes | yes | yes | yes | yes | yes | yes |
| Agents | yes | yes | yes | ✓ | yes | — | yes |
| Subagents | yes | yes | — | — | — | — | — |
| Skills | yes | yes | yes | yes | yes | yes | yes |
| Hooks | yes | yes | yes | yes | yes | yes | instruction-only |
| Commands | yes | yes | — | — | yes | — | — |
| Prompt templates | — | — | yes | yes | — | — | yes |
| Chat modes | yes | — | yes | — | — | — | — |
| MCP | yes | yes | yes | yes | yes | yes | yes |
| Permissions | yes | yes | — | — | — | yes | yes |
| Plugins | yes | yes | yes | yes | yes | yes | — |
| Specs | — | — | — | — | — | — | — |
| Steering | — | — | — | — | — | — | — |
| Compaction | — | — | — | yes | yes | — | — |
| Sessions | — | — | — | — | yes | — | — |
| Global config | yes | yes | — | yes | yes | yes | yes |

## 3. Asset Output Mapping

Each cell shows the output shape for the (tool, asset-kind) pair. See `output_mapping.go` for the full `OutputTarget` struct.

| Asset kind | OpenCode | Claude Code | Copilot | Pi | OMP | Antigravity | Kiro |
|---|---|---|---|---|---|---|---|
| **Agents** | flat → `.opencode/agents/` | flat → `.claude/agents/` | flat → `.github/agents/` | flat → `.pi/agents/` | flat → `.omp/agents/` | none | flat → `.kiro/agents/` |
| **Skills** | dir-per-item → `.opencode/skills/` | dir-per-item → `.claude/skills/` | dir-per-item → `.github/skills/` | dir-per-item → `.pi/skills/` | dir-per-item → `.omp/skills/` | dir-per-item → `.agents/skills/` | dir-per-item → `.kiro/skills/` |
| **Templates** | none | flat → `.claude/templates/` | flat → `.github/instructions/` | none | none | none | none |
| **Commands** | flat → `.opencode/commands/` | flat → `.claude/commands/` | none | none | flat → `.omp/commands/` | none | none |
| **Chat modes** | flat → `.opencode/modes/` | none | flat → `.github/chatmodes/` | none | none | none | none |
| **Output styles** | none | flat → `.claude/output-styles/` | none | none | none | none | none |
| **Prompts** | none | none | rewrite-ext → `.github/prompts/` | flat → `.pi/prompts/` | flat → `.omp/prompts/` | none | flat → `.kiro/prompts/` |

**Shape legend:**
- **flat** — file copied with original basename to destination directory
- **dir-per-item** — per-item subdirectory created, file written as `SKILL.md` inside it
- **rewrite-ext** — extension rewritten (e.g. `.md` → `.prompt.md`)
- **none** — asset kind intentionally not installed for this tool

## 4. Scope Support

| Tool | Project | Workspace | Global |
|---|---|---|---|
| OpenCode | yes | yes | yes |
| Claude Code | yes | yes | yes |
| Copilot | yes | yes | yes (probes for CLI) |
| Pi | yes | yes | yes |
| OMP | yes | yes | yes |
| Antigravity | yes | yes | yes |
| Kiro | yes | yes | yes |

All 7 LazyAI-supported targets support project, workspace, and global scopes. Antigravity global-scope skills write the documented `~/.gemini/config/skills/` root (#486).

## 5. Headless Support

| Tool | CanRunHeadless |
|---|---|
| OpenCode | yes |
| Claude Code | yes |
| Copilot | yes |
| Pi | no |
| OMP | no |
| Antigravity | no |
| Kiro | no |

## 6. MCP Behavior

| Tool | MCP compile | Output location | Notes |
|---|---|---|---|
| OpenCode | `CompileMCPForTool` | `.opencode/lazyai.mcp.jsonc` (legacy), `.opencode/mcp.json` | Preserves user-authored servers across re-runs |
| Claude Code | `CompileMCPForTool` | `.mcp.json` (project), `settings.json` merge (global) | CLI-driven `claude mcp add-json` with fallback to direct-write |
| Copilot | `CompileMCPForTool` | `.vscode/mcp.json` (project), `~/.copilot/mcp-config.json` (CLI) | Dual output: VS Code + CLI; CLI probe-gated |
| Pi | **no-op** (`CompileMCP` returns `ctx.FileRecords`) | — | Pi has no native MCP surface; MCP capability is declared for future use |
| OMP | `CompileMCPForTool` | `.omp/mcp.json` | Standard compile path |
| Antigravity | `CompileMCPForTool` | `~/.gemini/config/mcp_config.json` | User-level Gemini MCP config |
| Kiro | `CompileMCPForTool` | `.kiro/settings/mcp.json` | Standard compile path |

## 7. Test Coverage

| Adapter | Key test files | What they verify |
|---|---|---|
| **OpenCode** | `opencode_adapter_test.go`, `opencode_frontmatter_test.go`, `opencode_validate_test.go`, `opencode_plugin_test.go` | Install from FS, global scope, preserves root json, commands+modes, selection filters, instructions key resolution, default agent, skill surface, package.json |
| **Claude Code** | `claudecode_frontmatter_test.go`, `claudecode_global_layout_test.go`, `claudecode_drivecli_test.go`, `claude_cli_test.go` | Hook scripts + settings, agent rewrite, global layout, CLI-driven MCP |
| **Copilot** | `copilot_chatmodes_test.go`, `copilot_skill_tier_test.go` | Chat mode install, skill tier resolution |
| **Pi** | `pi_adapter_test.go` | Agents, skills, prompts, extensions; confirms no `.pi/hooks` path |
| **OMP** | `omp_adapter_test.go` | Agents + skills, commands + prompts, hooks, global scope, MCP compile |
| **Antigravity** | `antigravity_install_test.go` | Skills at `.agents/skills/`, hooks at `.agents/hooks.json`, settings at `.gemini/settings.json`, workspace root resolution |
| **Kiro** | `kiro_adapter_test.go` | Agent profiles, skills, prompts; confirms no `.kiro/workflows` |
| **All** | `capabilities_test.go`, `output_mapping_test.go`, `adapter_contract_test.go`, `adapter_adapters_test.go` | Every adapter reports capabilities, output coverage is exhaustive, neutral contract, per-adapter install smoke tests |

## 8. Key Limitations

### Pi — MCP is a no-op
Pi's `CompileMCP` method returns `ctx.FileRecords` unchanged (`pi.go:81-83`). The adapter declares MCP capability for future compatibility, but no MCP configuration is emitted for Pi. Pi has no native MCP surface.

### Kiro — No specs or steering; hooks are instruction-only
Kiro does not emit native specs or steering files (`capabilities_test.go:68-69`). Hooks are instruction-only — no runtime `.kiro/hooks` files are emitted (`capabilities.go:163-164`). The adapter installs agents, skills, prompts, MCP, permissions, and global config, but specs, steering, and runtime hook files are intentionally absent.

### OMP — Stable (verified 2026-06-23, #486)
OMP was promoted from beta to stable after every emitted surface (root `AGENTS.md`, `.omp/agents`, `.omp/skills`, `.omp/hooks/pre/*.ts`, `.omp/commands`, `.omp/mcp.json`) was verified against the authoritative OMP (Oh My Pi) docs (`omp://`). `.omp/prompts/` is still emitted best-effort (discovery not docs-confirmed). `CanRunHeadless()=false` (as with Pi).

### Antigravity — Stable (verified 2026-06-23, #486), no agents
Antigravity is dual-target (Antigravity IDE `.agents/*` + Gemini CLI `.gemini/settings.json`). Workspace skills (`.agents/skills/`), global skills (`~/.gemini/config/skills/`), IDE hooks (`.agents/hooks.json`), CLI hooks (`.gemini/settings.json`), MCP (`~/.gemini/config/mcp_config.json`, HTTP via `serverUrl`), and root instructions (`GEMINI.md` importing `@./AGENTS.md` for Gemini CLI + `.agents/rules/lazyai.md` importing `@/AGENTS.md` for Antigravity IDE) are docs-verified. It does not emit agent files (`output_mapping.go:346-350`: `ShapeNone`). Promoted from beta to stable after the two #486 gaps (global-skills path, root-instructions discovery) were closed and pinned by conformance tests. Global rules (`~/.gemini/GEMINI.md`) remain user-managed.

### Copilot — No subagents, no commands
Copilot does not support subagents or slash commands. It compensates with prompt templates, chat modes, and dual MCP output (VS Code + CLI).

### OpenCode — No prompts, no output styles
OpenCode does not have a prompts directory or output-styles concept. Prompts ship as commands.

### Claude Code — No chat modes, no prompts
Claude Code has no chat modes concept. Prompts ship as commands.
