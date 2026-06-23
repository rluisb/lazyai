# Supported Tools

`lazyai-cli` compiles embedded library content into native formats for seven setup surfaces.

LazyAI's workflow catalog is canonical source material, not a universal runtime surface. Adapters only emit workflow-like behavior through a host tool's verified native format. Unsupported `workflows/` directories are intentionally not created.

## OpenCode

- **Description:** root instructions plus agents, skills, commands, chat modes, MCP, permissions, and a managed hook/plugin surface
- **Root file:** `AGENTS.md`
- **Config directory:** `.opencode/`
- **Project/workspace config:** `opencode.json`
- **MCP output:** baseline root `opencode.json` includes managed MCP entries under top-level `mcp`.
- **Workflow delivery:** no verified native `.opencode/workflows` surface; workflow helpers must use commands, skills, modes, instructions, or a verified OpenCode plugin.
- **Special behavior:** canonical agent frontmatter is rewritten to OpenCode format; legacy MCP entries are migrated into root `opencode.json`.

## Claude Code

- **Description:** root instructions plus agents, skills, rules scaffold, commands, output styles, MCP, permissions, and managed hook scripts
- **Root file:** `AGENTS.md` (existing root `CLAUDE.md` is preserved and receives an `AGENTS.md` reference)
- **Config directory:** `.claude/`
- **Global scope support:** Yes — `~/.claude/`
- **MCP config:** `.mcp.json` at project/workspace scope; global MCP lives in Claude settings
- **Workflow delivery:** candidate native `.claude/workflows/` support is tracked separately; current stable output projects workflow discipline through skills, commands, hooks, and instructions.
- **Special behavior:** generates `.claude/settings.json`, `.claude/hooks/*.sh`, and a sample `.claude/rules/typescript.md`

## GitHub Copilot

- **Description:** repo or user instructions, canonical agent markdown files, prompts, chatmodes, MCP config, and project/workspace hook assets
- **Root files:** `.github/copilot-instructions.md` and `AGENTS.md`
- **Config directory:** `.github/`
- **Global scope support:** Yes — probe-gated on `copilot` CLI or `~/.copilot/`
- **MCP config:** `.vscode/mcp.json` at project/workspace scope; `~/.copilot/mcp-config.json` at global scope when probe passes
- **Workflow delivery:** no native workflow directory; workflow guidance is delivered through instructions, prompts, agents, chat modes, and hooks.
- **Special behavior:** skills are emitted to `.github/skills/<name>/SKILL.md` (global `~/.copilot/skills/<name>/SKILL.md`); prompts remain `.prompt.md`; project/workspace hook assets land under `.github/hooks/`

## Pi

- **Description:** shared root instructions plus skills, prompt templates, extension-based safety hooks, and declared MCP capability with no emitted Pi MCP config
- **Root file:** `AGENTS.md`
- **Config directory:** `.pi/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No
- **Workflow delivery:** no `.pi/workflows` surface; workflow helpers must use Pi extensions, skills, and prompts.
- **Special behavior:** emits `.pi/skills/<name>/SKILL.md`, `.pi/prompts/*.md`, and Pi safety hooks as `.pi/extensions/*.ts`

## OMP

- **Description:** shared root instructions plus agents, skills, commands, hooks, MCP, plugins/extensions, compaction, sessions/handoff, and global config
- **Root file:** `AGENTS.md`
- **Config directory:** `.omp/`
- **Project/workspace scope support:** Yes
- **Global scope support:** Yes — `~/.omp/agent/`
- **Workflow delivery:** no verified native `.omp/workflows` surface; workflow helpers must use OMP skills, commands, hooks, agents, or verified extensions.
- **Support level:** beta until OMP's official extension docs are fully snapshot-verified.

## Kiro

- **Description:** shared root instructions plus agents, skills, prompts, and MCP config
- **Root file:** `AGENTS.md`
- **Config directory:** `.kiro/`
- **Project/workspace scope support:** Yes
- **Global scope support:** Yes
- **Workflow delivery:** no `.kiro/workflows` emission; workflow intent must map only to verified Kiro-native surfaces.
- **Special behavior:** emits `.kiro/agents/<name>.md`, `.kiro/skills/<name>/SKILL.md`, `.kiro/prompts/*.md`, and `.kiro/settings/mcp.json`; no `.kiro/workflows` directory is emitted

## Antigravity

- **Description:** shared root instructions plus selected emitted skills, minimal `.gemini` settings, hook surface, MCP config, plugin capabilities, and permissions metadata
- **Root file:** `AGENTS.md`
- **Config directory:** `.gemini/` plus `.agents/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No for setup files; MCP compilation writes user-level Gemini config
- **Workflow delivery:** no verified native workflow directory; workflow helpers must use a verified Antigravity plugin layout, Agent Skills, hooks, or rules.
- **Support level:** beta until Antigravity plugin docs and install locations are fully snapshot-verified.
- **Special behavior:** emits `.gemini/settings.json`, `.gemini/hooks/lazyai/*.sh`, selected Agent Skills at `.agents/skills/<name>/SKILL.md`, and MCP config at `~/.gemini/config/mcp_config.json`; no custom agent files are emitted for Antigravity

## Workflow delivery matrix

| Target | Native workflow directory emitted today | Correct LazyAI delivery |
|---|---|---|
| OpenCode | No | Commands, skills, modes, instructions, or verified plugin helper |
| Claude Code | No | Candidate native `.claude/workflows/` support is evaluated separately; current output uses skills, commands, hooks, and instructions |
| Copilot | No | Instructions, prompts, agents, chat modes, and hooks |
| Pi | No | Pi extensions, skills, and prompts |
| OMP | No | OMP skills, commands, hooks, agents, or verified extension helper |
| Kiro | No | Verified Kiro-native surfaces; no `.kiro/workflows` |
| Antigravity | No | Verified plugin layout, Agent Skills, hooks, or rules |

## Comparison

| Capability | OpenCode | Claude Code | Copilot | Pi | OMP | Kiro | Antigravity |
|---|---|---|---|---|---|---|---|
| Project scope | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Workspace scope | Yes | Yes | Yes | Yes | Yes | Yes | Yes |
| Global scope | Yes | Yes | Yes (probe-gated) | No | Yes | Yes | No |
| Default agent entry | `.opencode/agents/guide.md` | `.claude/agents/guide.md` | `.github/agents/guide.agent.md` | — | `.omp/agents/guide.md` | `.kiro/agents/guide.md` | — |
| Skills surface | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | `.github/skills/<name>/SKILL.md` | `.pi/skills/<name>/SKILL.md` | `.omp/skills/<name>/SKILL.md` | `.kiro/skills/<name>/SKILL.md` | `.agents/skills/<name>/SKILL.md` |
| Hook runtime | `.opencode/plugins/vibe-lab-hooks.js` | `.claude/hooks/*.sh` + settings hooks | `.github/hooks/*.{json,sh}` | `.pi/extensions/*.ts` | `.omp/hooks/*` | — | `.gemini/hooks/lazyai/*.sh` + settings hooks |
| MCP output | `opencode.json` (managed MCP under top-level `mcp`) | `.mcp.json` / Claude settings | `.vscode/mcp.json` / `~/.copilot/mcp-config.json` | Capability only; no config currently written | `.omp/mcp.json` / OMP config | `.kiro/settings/mcp.json` | `~/.gemini/config/mcp_config.json` |
| Workflow directory | — | — | — | — | — | — | — |

```bash
lazyai-cli init --tools opencode,claude-code,copilot,pi,omp,kiro,antigravity
```

You can add a tool later:

```bash
lazyai-cli add copilot
```

Then recompile managed MCP output:

```bash
lazyai-cli compile
```
