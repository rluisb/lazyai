# How It Works

`lazyai-cli` uses a canonical-source-to-compile model. You edit one tool-agnostic layer; `lazyai-cli` generates the tool-native files.

## Canonical source

The canonical layer lives under `.ai/`:

```text
.ai/
├── constitution/
│   ├── constitution.md
│   ├── constraints.md
│   ├── quality-gates.md
│   └── uncertainty.md
└── mcp.json
```

This layer is human-editable and version-controllable.

## Compiled output

From `.ai/`, `lazyai-cli compile` generates tool-native files such as:

- `AGENTS.md` — root instructions for all tools
- `.opencode/` — OpenCode agents, skills, commands, and MCP config
- `.claude/` — Claude Code rules, agents, skills, and `.mcp.json`
- `.github/` — Copilot instructions and prompt files
- `.vscode/` — VS Code MCP config

## Workflow

1. **Initialize once**: `lazyai-cli init`
2. **Edit canonical files**: change rules, agents, or templates in `.ai/`
3. **Recompile**: `lazyai-cli compile` or `lazyai-cli update`
4. **Verify**: `lazyai-cli doctor` checks drift and missing files

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

This powers `lazyai-cli status`, `lazyai-cli doctor`, and `lazyai-cli update`.

## Conflict and update behavior

| Situation | Behavior |
|---|---|
| Tracked + unchanged | Safely overwritten with latest managed content |
| Tracked + customized | Prompts/backs up before overwrite; `--force` auto-overwrites with backup |
| Existing untracked collision | Prompts before replacement; replacement creates backup |
| Newly expected file | Created and added to `.ai-setup.json` |
| Tracked but missing | Reported as missing, not silently recreated |

Backups are written under `.ai-setup-backup/` with relative paths preserved.

## TOML defaults

You can provide CLI defaults via TOML:

- Project: `.ai-setup.toml`
- Global: `~/.config/lazyai/config.toml`

Precedence: `CLI flags > project TOML > global TOML > built-in defaults`

## Execution model

- Local native agents are the intended execution path: Claude Code, OpenCode, and Copilot.
- A2A is a config seam only and is not remote/network execution by default.
