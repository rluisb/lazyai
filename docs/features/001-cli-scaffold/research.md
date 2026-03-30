# Research: CLI Scaffold Tool

**Date:** 2026-03-28
**Agent:** Scout (manual)

---

## What Exists

### Our Setup Assets (Source of Truth)
- **Location:** `~/Documents/AI-Agentic-Setup-Templates/` (29 files)
- **Content:** Agent definitions, skills, templates, rules, AGENTS.md context files, infrastructure
- **Format:** All markdown. Tool-agnostic content in `docs/`. Pi-specific in `.pi/`.

### Distribution Models Researched

| Model | Source | Mechanism |
|-------|--------|-----------|
| Skills CLI | agentskills.io | `npx skills add org/repo` — clones SKILL.md into project |
| Spec-Kit Presets | github/spec-kit | `specify preset add name` — stackable overlays with priority |
| Cross-Agent | disler/pi-vs-claude-code | Pi extension scans .claude/.gemini/.codex/ and imports |
| tophawks-code-agent | TOP-HAWKS | Git clone entire agent repo |
| the-library | disler | Shared repo of reusable agent assets |

### Tool Directory Structures

| Tool | Agents | Commands/Skills | Reads at Start |
|------|--------|----------------|---------------|
| **Pi** | `.pi/agents/*.md` | `.pi/skills/*.md` | `CLAUDE.md` |
| **OpenCode** | `.opencode/agents/*.md` | `.opencode/commands/*.md` | `AGENTS.md` |
| Claude Code | `.claude/commands/*.md` | `.claude/rules/*.md` | `CLAUDE.md` |
| Codex | (via AGENTS.md) | (via AGENTS.md) | `AGENTS.md` |
| Gemini CLI | `.gemini/agents/` | `.gemini/commands/*.toml` | `AGENTS.md` |
| Copilot | `.github/agents/*.agent.md` | `.github/agents/*.prompt.md` | `.github/copilot-instructions.md` |

### Key Findings

1. **Two-layer architecture:** `docs/` is 100% shared across tools. Tool-specific dirs differ only in location + frontmatter.
2. **Pi and OpenCode both read markdown** — same content body, different frontmatter.
3. **Workspace vs Project** is a setup-time decision affecting where shared config lives.
4. **gum works in bash, but for npx we need Node.js prompts** — `@clack/prompts` or `inquirer`.
5. **Adapter pattern:** one adapter per tool. Library content stays unchanged. Adding a tool = adding one adapter file.

## Gotchas

- Pi reads `CLAUDE.md`, OpenCode reads `AGENTS.md` — both must exist with identical content
- OpenCode agent format from the workflow report: agents in `.opencode/agents/`, commands in `.opencode/commands/`
- Pi skills have `name`, `description`, `usage` in frontmatter. OpenCode commands have only `description`.
- The `update` command must be smart: skip customized files, update untouched, add new.

## Questions for Planner

1. Should we use TypeScript or plain JavaScript for the CLI?
2. Which prompt library: `@clack/prompts` (prettier) or `inquirer` (more features)?
3. Should the library content live inside the npm package or be fetched from a git repo?
4. How do we detect "customized" files during update? (hash comparison vs git status)
