# Supported Tools

`lazyai-cli` compiles embedded library content into native formats for five setup surfaces.

## OpenCode

- **Description:** root instructions plus agents, skills, commands, modes, and a managed hook plugin
- **Root file:** `AGENTS.md`
- **Config directory:** `.opencode/`
- **Project/workspace config:** `.opencode/opencode.jsonc`
- **Global scope support:** Yes — `~/.config/opencode/`
- **MCP config:** merged into `.opencode/opencode.jsonc`
- **Special behavior:** canonical agent frontmatter is rewritten to OpenCode format; managed hook runtime lands at `.opencode/plugins/vibe-lab-hooks.js`

## Claude Code

- **Description:** root instructions plus agents, skills, rules scaffold, commands, output styles, and managed hook scripts
- **Root file:** `AGENTS.md` (existing root `CLAUDE.md` is preserved and receives an `AGENTS.md` reference)
- **Config directory:** `.claude/`
- **Global scope support:** Yes — `~/.claude/`
- **MCP config:** `.mcp.json` at project/workspace scope; global MCP lives in Claude settings
- **Special behavior:** generates `.claude/settings.json`, `.claude/hooks/*.sh`, and a sample `.claude/rules/typescript.md`

## GitHub Copilot

- **Description:** repo or user instructions, agent YAML files, prompts, chatmodes, MCP config, and project/workspace hook assets
- **Root files:** `.github/copilot-instructions.md` and `AGENTS.md`
- **Config directory:** `.github/`
- **Global scope support:** Yes — probe-gated on `copilot` CLI or `~/.copilot/`
- **MCP config:** `.vscode/mcp.json` at project/workspace scope; `~/.copilot/mcp-config.json` at global scope when probe passes
- **Special behavior:** skills are converted into `.github/agents/<name>.agent.yaml`; prompts remain `.prompt.md`; project/workspace hook assets land under `.github/hooks/`

## OMP/Pi

- **Description:** shared root instructions plus skills-only surface
- **Root file:** `AGENTS.md`
- **Config directory:** `.pi/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No
- **Special behavior:** emits `.pi/skills/<name>/SKILL.md` only; no Pi agents, prompt surface, or runtime hooks are generated

## Antigravity

- **Description:** shared root instructions plus minimal `.gemini` settings and hook surface
- **Root file:** `AGENTS.md`
- **Config directory:** `.gemini/`
- **Project/workspace scope support:** Yes
- **Global scope support:** No
- **Special behavior:** emits `.gemini/settings.json` and `.gemini/hooks/vibe-lab/*.sh`; no Antigravity agent or skills directory is generated

## Comparison

| Capability | OpenCode | Claude Code | Copilot | OMP/Pi | Antigravity |
|---|---|---|---|---|---|
| Project scope | Yes | Yes | Yes | Yes | Yes |
| Workspace scope | Yes | Yes | Yes | Yes | Yes |
| Global scope | Yes | Yes | Yes (probe-gated) | No | No |
| Primary agent entry | `.opencode/agents/primary-agent.md` | `.claude/agents/primary-agent.md` | `.github/agents/primary-agent.agent.yaml` | — | — |
| Skills surface | `.opencode/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | `.github/agents/<skill>.agent.yaml` | `.pi/skills/<name>/SKILL.md` | — |
| Hook runtime | `.opencode/plugins/vibe-lab-hooks.js` | `.claude/hooks/*.sh` + settings hooks | `.github/hooks/*.{json,sh}` | — | `.gemini/hooks/vibe-lab/*.sh` + settings hooks |
| MCP output | `.opencode/opencode.jsonc` | `.mcp.json` / Claude settings | `.vscode/mcp.json` / `~/.copilot/mcp-config.json` | — | — |

## Tool selection during init

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
