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
| `.omp/AGENTS.md` | Root instructions (project scope; native provider, priority 100) |

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

## Root Instructions Behavior

Project-scope root instructions land at `.omp/AGENTS.md`, which is read by OMP's native context provider at priority 100. This is higher precedence than the generic `agents-md` provider (which reads bare `AGENTS.md` at priority 10). Workspace and global scopes are unaffected.

## Agent Behavior

Canonical agents are **transformed** (not copied verbatim) into OMP-native subagent files under `.omp/agents/<name>.md`. The transform (`RewriteAgentForOMP` in `agent_transform.go`) produces frontmatter that OMP can natively consume.

**Emitted frontmatter fields:**

| Field | Source | Notes |
|---|---|---|
| `name` | canonical `name:` | required |
| `description` | canonical `description:` | double-quoted |
| `tools` | derived from canonical `tools:` grants | OMP allowlist; see mapping below |
| `thinkingLevel` | derived from grants + agent name | `"low"` for read-only, `"high"` for planner, `"auto"` otherwise |
| `autoloadSkills` | canonical `skills:` list | omitted when the source has no `skills:` list |

**Dropped fields (LazyAI-only, not OMP-native):** `role`, `mode`, `temperature`, `steps`.

**Canonical grant → OMP tool name mapping:**

| Canonical grant | OMP tool name(s) |
|---|---|
| `read` | `read` |
| `edit` | `edit`, `write` |
| `shell` | `bash` |
| `search` | `search` |
| `web` | `web_search` |
| `mcp` | (omitted — no generic OMP MCP token; server-specific names are configured separately) |
| `spawn` | `task` |

**Read-only restriction:** agents whose canonical `tools:` list contains only `read` and/or `search` (`researcher`, `reviewer`, `evidence-verifier`) receive `tools: ["read", "search"]` only. Mutation tools (`bash`, `edit`, `write`, `task`) are absent.

**No `tools:` field in source:** agents without a canonical `tools:` declaration receive the full default OMP set (`read`, `search`, `bash`, `edit`, `write`, `web_search`, `task`), preserving unrestricted legacy behaviour.

**Body:** the agent system-prompt body is preserved verbatim after the vibe-lab managed marker.

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
- `.omp/tools/` and `.omp/extensions/` are documented OMP-native discovery surfaces (see `omp://custom-tools.md` and `omp://extension-loading.md`), but the adapter does not emit them; executable-module generation is out of current product scope. User-authored tools and extension modules remain user-managed.

## Test Coverage

| Test file | What it verifies |
|---|---|
| `omp_adapter_test.go` | Agents + skills, commands + prompts, hooks, global scope, MCP compile |
| `omp_frontmatter_test.go` | `RewriteAgentForOMP` unit tests (read-only restriction, shell grant, planner thinking level, managed marker, body preservation, nil-grants fallback, multi-skill mapping); `TestOmpAdapter_Install_AgentFrontmatterContent` integration test (installed researcher.md has OMP-native fields, no LazyAI fields) |
| `scaffold/root_test.go` | #560: `TestScaffoldCompiledRootOmpProjectLandsInOmpDir` — project-scope root at `.omp/AGENTS.md`; `TestScaffoldCompiledRootOmpDoesNotAffectOtherTargets` |