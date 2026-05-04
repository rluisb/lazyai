# How It Works

`ai-setup` uses a **canonical source → compile** model. You edit one tool-agnostic layer; `ai-setup` generates the rest.

## Canonical source

The canonical layer lives under `.ai/`:

```text
.ai/
├── constitution/
│   ├── constitution.md
│   ├── constraints.md
│   ├── quality-gates.md
│   └── uncertainty.md
├── mcp.json
└── orchestration/         # when orchestrator is enabled
    ├── chains/
    ├── teams/
    ├── workflows/
    └── skills/
        ├── domains/
        └── modes/
```

This is where you make changes. Everything in `.ai/` is human-editable and version-controllable.

## Compiled output

From `.ai/`, `ai-setup compile` generates tool-native files:

- `AGENTS.md` — root instructions for all tools
- `.opencode/` — OpenCode agents, skills, commands, and MCP config
- `.claude/` — Claude Code rules, agents, skills, and `.mcp.json`
- `.github/` — Copilot instructions and prompt files
- `.vscode/` — VS Code MCP config
- per-tool orchestrator guidance (when enabled)

## Workflow

1. **Initialize once**: `ai-setup init`
2. **Edit canonical files**: change rules, agents, or templates in `.ai/`
3. **Recompile**: `ai-setup compile` (or `ai-setup update` to refresh library content)
4. **Verify**: `ai-setup doctor` checks drift and missing files

## Manifest tracking

Every managed setup gets a manifest:

```text
.ai-setup.json
```

It tracks:

- setup scope and selected tools
- project/workspace metadata
- selected agents, skills, prompts, templates, and rules
- feature flags and git conventions
- managed file paths and content hashes
- operation history and sync metadata

This powers `ai-setup status`, `ai-setup doctor`, and `ai-setup update`.

## Conflict and update behavior

| Situation | Behavior |
|---|---|
| Tracked + unchanged | Safely overwritten with latest managed content |
| Tracked + customized | Prompts/backs up before overwrite; `--force` auto-overwrites with backup |
| Existing untracked collision | Prompts before replacement; replacement creates backup |
| Newly expected file | Created and added to `.ai-setup.json` |
| Tracked but missing | Reported as missing, not silently recreated |

Backups are written under `.ai-setup-backup/` with relative paths preserved.

## TOML defaults (optional)

You can provide CLI defaults via TOML:

- Project: `.ai-setup.toml`
- Global: `~/.config/ai-setup/config.toml`

Precedence: `CLI flags > project TOML > global TOML > built-in defaults`

## Execution model

- **Local native agents** are the intended execution path (Claude Code, OpenCode, Copilot).
- **A2A** is a config/seam only and is not remote/network execution by default.
- The optional `ai-setup-orchestrator` is a Go runtime invoked via `connect` so multiple MCP clients share a single daemon process.
