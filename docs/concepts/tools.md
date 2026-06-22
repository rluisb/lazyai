# Supported Tools

`lazyai-cli` compiles embedded library content into native formats for five setup surfaces.

## OpenCode

- **Description:** root instructions plus agents, skills, commands, chat modes, and a managed hook plugin
- **Root file:** `AGENTS.md`
- **Config directory:** `.opencode/`
- **Project/workspace config:** `opencode.json`
- **MCP output:** baseline root `opencode.json` includes managed MCP entries under top-level `mcp`.
- **Special behavior:** canonical agent frontmatter is rewritten to OpenCode format; legacy `.opencode/lazyai.mcp.jsonc` entries are migrated into root `opencode.json`.

- **Description:** root instructions plus agents, skills, rules scaffold, commands, output styles, and managed hook scripts
- **Root file:** `AGENTS.md` (existing root `CLAUDE.md` is preserved and receives an `AGENTS.md` reference)
- **Config directory:** `.claude/`
- **Global scope support:** Yes — `~/.claude/`
- **MCP config:** `.mcp.json` at project/workspace scope; global MCP lives in Claude settings
- **Special behavior:** generates `.claude/settings.json`, `.claude/hooks/*.sh`, and a sample `.claude/rules/typescript.md`

## GitHub Copilot

- **Description:** repo or user instructions, canonical agent markdown files, prompts, chatmodes, MCP config, and project/workspace hook assets
- **Root files:** `.github/copilot-instructions.md` and `AGENTS.md`
- **Config directory:** `.github/`
- **Global scope support:** Yes — probe-gated on `copilot` CLI or `~/.copilot/`
- **MCP config:** `.vscode/mcp.json` at project/workspace scope; `~/.copilot/mcp-config.json` at global scope when probe passes
- **Special behavior:** skills are emitted to `.github/skills/<name>/SKILL.md` (global `~/.copilot/skills/<name>/SKILL.md`); prompts remain `.prompt.md`; project/workspace hook assets land under `.github/hooks/`

## OMP/Pi

- **Description:** shared root instructions plus skills-only surface
- **Root file:** `AGENTS.md`
- **Config directory:** `.pi/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No
- **Special behavior:** emits `.pi/skills/<name>/SKILL.md` only; no Pi agents, prompt surface, or runtime hooks are generated

## Antigravity

- **Description:** shared root instructions plus selected emitted skills, minimal `.gemini` settings, and hook surface
- **Root file:** `AGENTS.md`
- **Config directory:** `.gemini/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No
- **Special behavior:** emits `.gemini/settings.json`, `.gemini/hooks/lazyai/*.sh`, and selected Agent Skills at `.agents/skills/<name>/SKILL.md`; no custom agent files are emitted for Antigravity

## Comparison

| Capability | OpenCode | Claude Code | Copilot | OMP/Pi | Antigravity |
|---|---|---|---|---|---|
| Project scope | Yes | Yes | Yes | Yes | Yes |
| Workspace scope | Yes | Yes | Yes | Yes | Yes |
| Global scope | Yes | Yes | Yes (probe-gated) | No | No |
| Default agent entry | `.opencode/agents/guide.md` | `.claude/agents/guide.md` | `.github/agents/guide.agent.md` | — | — |
| Skills surface | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | `.github/skills/<name>/SKILL.md` | `.pi/skills/<name>/SKILL.md` | `.agents/skills/<name>/SKILL.md` |
| Hook runtime | `.opencode/plugins/vibe-lab-hooks.js` | `.claude/hooks/*.sh` + settings hooks | `.github/hooks/*.{json,sh}` | — | `.gemini/hooks/lazyai/*.sh` + settings hooks |
| MCP output | `opencode.json` (managed MCP in top-level `mcp`; legacy source migrated from `lazyai.mcp.jsonc`) | `.mcp.json` / Claude settings | `.vscode/mcp.json` / `~/.copilot/mcp-config.json` | — | — |
```bash
lazyai-cli init --tools opencode,claude-code,copilot,pi,antigravity
```

You can add a tool later:

```bash
lazyai-cli add copilot
```

Then recompile managed MCP output:

```bash
lazyai-cli compile
```
