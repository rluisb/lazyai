# Adapter Capability Matrix

> **Source of truth:** `packages/cli/internal/adapter/capabilities.go` (declarative `Capability` struct per adapter) and `packages/cli/internal/adapter/output_mapping.go` (per-tool asset-kind output targets).  
> **Verification date:** 2026-06-25 (compliance re-audit complete; see `docs/adapters/official-compliance-audit-2026-06-25.md`).

## 1. Support Levels

| Level | Meaning | Adapters |
|---|---|---|
| **stable** | Official docs verified + golden tests + smoke tests | OpenCode, Claude Code, Copilot, Pi, Kiro, OMP, Antigravity |
| **beta** | Official docs verified + golden tests, limited runtime smoke | — |
| **experimental** | Docs partially verified or host tool still moving quickly | — |
| **deprecated** | Kept for migration only | — |

**OMP promotion (2026-06-23, #486):** every emitted OMP surface was verified against the authoritative OMP (Oh My Pi) docs (`omp://`), clearing the docs-snapshot blocker. See [snapshots/beta-adapter-verification-2026-06.md](snapshots/beta-adapter-verification-2026-06.md).

**Antigravity promotion (2026-06-23, #486):** the Antigravity IDE / Gemini CLI docs are JS-rendered and were snapshot-verified by rendering. All emitted surfaces are verified and the two former beta gaps are closed and pinned by conformance tests — global-scope skills now write `~/.gemini/config/skills/`, and root instructions are discovered (`GEMINI.md` for Gemini CLI, `.agents/rules/lazyai.md` for Antigravity IDE). No adapter remains below stable (EC-006 cleared).

**Compliance re-audit (2026-06-25):** an independent per-adapter audit identified eight medium-severity divergences (M1–M8). All eight were resolved and merged to `main` the same day (#554–#561). Every adapter is fully aligned with its official docs as of this date.

## 2. Capability Matrix

| Capability | OpenCode | Claude Code | Copilot | Pi | OMP | Antigravity | Kiro |
|---|---|---|---|---|---|---|---|
| **Support level** | stable | stable | stable | stable | stable | stable | stable |
| Root instructions | yes | yes | yes | yes | yes | yes | yes |
| Agents | yes | yes | yes | yes | yes | — | yes |
| Subagents | yes | yes | — | — | — | — | — |
| Skills | yes | yes | yes | yes | yes | yes | yes |
| Hooks | yes | yes | yes | yes | yes | yes | yes |
| Commands | yes | yes | — | — | yes | — | — |
| Prompt templates | — | — | yes | yes | yes | — | yes |
| Chat modes | yes | — | yes | — | — | — | — |
| MCP | yes | yes | yes | no | yes | yes | yes |
| Permissions | yes | yes | — | — | — | yes | yes |
| Plugins | yes | yes | yes | yes | yes | yes | — |
| Specs | — | — | — | — | — | — | — |
| Steering | — | — | — | — | — | — | — |
| Compaction | — | — | — | yes | yes | — | — |
| Sessions | — | — | — | — | yes | — | — |
| Global config | yes | yes | — | yes | no | yes | yes |

> **Pi notes (#531/#532):** `Agents` is `yes` — the adapter installs `.pi/agents/<name>.md`. `MCP` is `no` — `CompileMCP` is a no-op (Pi has no native MCP surface). `GlobalConfig` is now `yes` — the adapter emits `.pi/settings.json` for project/workspace scope and `~/.pi/agent/settings.json` for global scope.
> **Global config note (OMP):** `GlobalConfig` is intentionally `false` for OMP. OMP supports global agent configuration (`omp://settings.md`), but the adapter does not emit it — it is user-managed. This conservative claim was set in #523 and must not be reverted to `yes`.

## 3. Asset Output Mapping

Each cell shows the output shape for the (tool, asset-kind) pair. See `output_mapping.go` for the full `OutputTarget` struct.

| Asset kind | OpenCode | Claude Code | Copilot | Pi | OMP | Antigravity | Kiro |
|---|---|---|---|---|---|---|---|
| **Agents** | flat → `.opencode/agents/` | flat → `.claude/agents/` | flat → `.github/agents/` | flat → `.pi/agents/` | flat → `.omp/agents/` | none | flat → `.kiro/agents/` |
| **Skills** | dir-per-item → `.opencode/skills/` | dir-per-item → `.claude/skills/` | dir-per-item → `.github/skills/` | dir-per-item → `.pi/skills/` | dir-per-item → `.omp/skills/` | dir-per-item → `.agents/skills/` | dir-per-item → `.kiro/skills/` |
| **Templates** | none | flat → `.claude/templates/` | flat → `.github/instructions/` | none | none | none | none |
| **Commands** | flat → `.opencode/commands/` | flat → `.claude/commands/` | none | none | flat → `.omp/commands/` | none | none |
| **Chat modes** | flat → `.opencode/modes/` | none | flat → `.github/agents/` (`.agent.md` ext; migrated from `.github/chatmodes/` in #555) | none | none | none | none |
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

> **Scope vs. GlobalConfig (OMP):** global *scope* (writing files to `~/.omp/agent/`) is supported; `GlobalConfig` (emitting global agent configuration) is not — see the note in §2 and issue #523.

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
| Claude Code | `CompileMCPForTool` | `.mcp.json` (project), `settings.json` merge (global) | CLI-driven `claude mcp add-json` with fallback to direct-write; global hooks path and MCP scope gaps fixed in #558/#559 |
| Copilot | `CompileMCPForTool` | `.vscode/mcp.json` (project), `~/.copilot/mcp-config.json` (CLI) | Dual output: VS Code + CLI; CLI probe-gated; CLI remote transport fixed to `http` in #557 |
| Pi | **no-op** (`CompileMCP` returns `ctx.FileRecords`) | — | Pi has no native MCP surface; capability is intentionally `no` |
| OMP | `CompileMCPForTool` | `.omp/mcp.json` | Standard compile path |
| Antigravity | `CompileMCPForTool` | `~/.gemini/config/mcp_config.json` (IDE) + `.gemini/settings.json` → `mcpServers` (Gemini CLI) | Dual output since #554: `serverUrl` for Antigravity IDE; `httpUrl` for Gemini CLI |
| Kiro | `CompileMCPForTool` | `.kiro/settings/mcp.json` | Standard compile path; extraneous `type` field in remote entries removed in #556 |

## 7. Test Coverage

| Adapter | Key test files | What they verify |
|---|---|---|
| **OpenCode** | `opencode_adapter_test.go`, `opencode_frontmatter_test.go`, `opencode_validate_test.go`, `opencode_plugin_test.go` | Install from FS, global scope, preserves root json, commands+modes, selection filters, instructions key resolution, default agent, skill surface, package.json |
| **Claude Code** | `claudecode_frontmatter_test.go`, `claudecode_global_layout_test.go`, `claudecode_drivecli_test.go`, `claude_cli_test.go` | Hook scripts + settings, agent rewrite, global layout, CLI-driven MCP |
| **Copilot** | `copilot_chatmodes_test.go`, `copilot_skill_tier_test.go` | Custom agent install (`.agent.md` at `.github/agents/`), skill tier resolution |
| **Pi** | `pi_adapter_test.go` | Agents, skills, prompts, extensions; confirms no `.pi/hooks` path |
| **OMP** | `omp_adapter_test.go` | Agents + skills, commands + prompts, hooks, global scope, MCP compile |
| **Antigravity** | `antigravity_install_test.go` | Skills at `.agents/skills/`, hooks at `.agents/hooks.json`, settings at `.gemini/settings.json`, workspace root resolution |
| **Kiro** | `kiro_adapter_test.go` | Agent profiles, skills, prompts; confirms no `.kiro/workflows` |
| **All** | `capabilities_test.go`, `output_mapping_test.go`, `adapter_contract_test.go`, `adapter_adapters_test.go` | Every adapter reports capabilities, output coverage is exhaustive, neutral contract, per-adapter install smoke tests |

## 8. Key Limitations

### Pi — MCP is a no-op
Pi's `CompileMCP` method returns `ctx.FileRecords` unchanged (`pi.go:81-83`). The adapter declares MCP capability for future compatibility, but no MCP configuration is emitted for Pi. Pi has no native MCP surface.

### Kiro — No specs or steering; permissions and Powers are non-goals
Kiro emits native `.kiro/hooks/<name>.json` files using the Kiro CLI v3 hook schema. Specs are not emitted because they are user-authored workflow artifacts. Repo-local permissions are forbidden; `Permissions: true` is host-support metadata, not an emitted repo file. Direct `.kiro/powers/` output is not emitted; Powers remain a future importable-package direction.

**MCP (#556):** Remote MCP entries previously included an extraneous `type:"http"` key not in Kiro's documented remote shape (`{url, headers}`). Fixed in #556 by adding a dedicated `toKiroMcp()` serializer that omits `type`.

### OMP — Stable (verified 2026-06-23, #486)
OMP was promoted from beta to stable after every emitted surface (`.omp/AGENTS.md`, `.omp/agents`, `.omp/skills`, `.omp/hooks/pre/*.ts`, `.omp/commands`, `.omp/prompts`, `.omp/mcp.json`) was verified against the authoritative OMP (Oh My Pi) docs (`omp://`). `.omp/prompts/*.md` discovery is docs-confirmed via `omp://config-usage.md` §6 (native `.omp` provider loads `prompts/*.md`). `GlobalConfig` is intentionally `false`: OMP's global agent configuration (`omp://settings.md`) is user-managed and the adapter does not emit it (conservative claim, #523). `CanRunHeadless()=false` (as with Pi). `.omp/tools/` and `.omp/extensions/` are documented OMP-native discovery surfaces (see `omp://custom-tools.md` and `omp://extension-loading.md`) but are not emitted by LazyAI; executable-module generation is out of current product scope and user-authored tools/extension modules remain user-managed.

**Root path (#560):** Project-scope root instructions now land at `.omp/AGENTS.md` (native OMP provider, priority 100). Previously they landed at the bare `AGENTS.md` root (OMP `agents-md` provider, priority 10). Fixed in #560 by redirecting `memoryDocDestPath` for OMP at project scope; global and workspace scopes unchanged.

### Antigravity — Stable (verified 2026-06-23, #486), no agents
Antigravity is dual-target (Antigravity IDE `.agents/*` + Gemini CLI `.gemini/settings.json`). Workspace skills (`.agents/skills/`), global skills (`~/.gemini/config/skills/`), IDE hooks (`.agents/hooks.json`), CLI hooks (`.gemini/settings.json`), MCP (dual output — see below), and root instructions (`GEMINI.md` importing `@./AGENTS.md` for Gemini CLI + `.agents/rules/lazyai.md` importing `@/AGENTS.md` for Antigravity IDE) are docs-verified. It does not emit agent files (`output_mapping.go:346-350`: `ShapeNone`). Promoted from beta to stable after the two #486 gaps (global-skills path, root-instructions discovery) were closed and pinned by conformance tests. Global rules (`~/.gemini/GEMINI.md`) remain user-managed.

**MCP (#554):** `compileAntigravityMCP` now writes to both targets: `~/.gemini/config/mcp_config.json` with `serverUrl` (Antigravity IDE) and the scope-appropriate `.gemini/settings.json` → `mcpServers` with `httpUrl` (Gemini CLI, per `settings.schema.json`). Previously only the IDE path was written; Gemini CLI users saw no MCP servers. Fixed in #554.

### Copilot — No subagents, no commands
Copilot does not support subagents or slash commands. It compensates with prompt templates, custom agents (`.github/agents/`), and dual MCP output (VS Code + CLI).

**Custom agents (#555):** VS Code 2026 deprecated chat modes (`.github/chatmodes/<name>.chatmode.md`) in favor of custom agents at `.github/agents/<name>.agent.md`. LazyAI now emits chat-mode library assets to `.github/agents/` with the `.agent.md` extension. The `chatmodes` library source subdir is unchanged; only the destination and extension changed. Fixed in #555.

**CLI MCP transport (#557):** Copilot CLI remote MCP entries previously used the deprecated `sse` transport. Now emits `type: "http"` per the Copilot CLI MCP docs. Fixed in #557.

### OpenCode — No prompts, no output styles
OpenCode does not have a prompts directory or output-styles concept. Prompts ship as commands.

**Mode frontmatter (#561):** Bundled mode files previously used the deprecated `tools:` boolean-map key and omitted `mode:`. Now emit `permission:` with a deny-list and `mode: primary` per the current OpenCode docs schema. Fixed in #561.

### Claude Code — No chat modes, no prompts
Claude Code has no chat modes concept. Prompts ship as commands.

**Global hooks path (#558):** At global scope, `settings.json` hook commands previously referenced `${CLAUDE_PROJECT_DIR:-$PWD}/.claude/hooks/<x>.sh` — a project-relative path absent in global-only installs. Now emits `~/.claude/hooks/<x>.sh` at global scope. Project-scope installs unaffected. Fixed in #558.

**Global MCP scope (#559):** An internal comment incorrectly stated "mcpServers live in settings.json". Global/user MCP is written via DriveCLI (`claude mcp add-json -s user` → `~/.claude.json`); `settings.json` does not hold MCP config. Comment and internal docs corrected in #559.