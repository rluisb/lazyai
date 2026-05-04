# Supported Tools

`ai-setup` compiles canonical content into native formats for three AI coding assistants.

## OpenCode

- **Description:** project instructions for OpenCode plus agent, skill, and command directories
- **Root file:** `AGENTS.md`
- **Config directory:** `.opencode/`
- **Project config:** `.opencode/opencode.json`
- **Global scope support:** Yes — `~/.config/opencode/`
- **MCP config:** `.opencode/opencode.jsonc`
- **Special behavior:** agent YAML frontmatter is stripped and a `<!-- Recommended model: ... -->` comment is injected when a `model:` frontmatter key exists

## Claude Code

- **Description:** Claude Code agents, skills, rules scaffold, with root instructions in `AGENTS.md`
- **Root file:** `AGENTS.md` (existing root `CLAUDE.md` is preserved and receives an `AGENTS.md` reference)
- **Config directory:** `.claude/`
- **Global scope support:** Yes — `~/.claude/`
- **MCP config:** `.mcp.json`
- **Special behavior:** generates `.claude/settings.json` and a sample `.claude/rules/typescript.md` rule with `paths:` frontmatter

## GitHub Copilot

- **Description:** repo instructions and prompt files for GitHub Copilot workflows
- **Root files:** `.github/copilot-instructions.md` and `AGENTS.md`
- **Config directory:** `.github/`
- **Global scope support:** No
- **MCP config:** `.vscode/mcp.json`
- **Special behavior:** skills are transformed into `.prompt.md` files with `mode: agent` frontmatter; prompt templates also compile to `.prompt.md`

## Comparison

| Capability | OpenCode | Claude Code | Copilot |
|---|---|---|---|
| Project scope | Yes | Yes | Yes |
| Global scope | Yes | Yes | No |
| MCP config | `.opencode.jsonc` | `.mcp.json` | `.vscode/mcp.json` |
| Agent directories | Yes | Yes | Prompts only |
| Skill directories | Yes | Yes | Prompts only |
| Orchestrator guidance | `.opencode/agents/orchestrator.md` | `.claude/agents/orchestrator.md` | `.github/prompts/orchestrator.prompt.md` |

## Tool selection during init

```bash
ai-setup init --tools opencode,claude-code,copilot
```

You can add a tool later:

```bash
ai-setup add copilot
```

Then recompile:

```bash
ai-setup compile
```
