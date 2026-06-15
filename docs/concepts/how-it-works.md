# How It Works

`lazyai-cli` uses a canonical-source-to-compile model. You edit one tool-agnostic layer; `lazyai-cli` generates the tool-native files. LazyAI owns the runtime/product surface that performs this setup, compilation, and local runtime-adjacent state management.

vibe-lab is an input to the product boundary: it supplies principles, assets, and adapter expectations that LazyAI may embed or adapt. It is not a runtime dependency, and LazyAI keeps execution ownership inside the LazyAI product.

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


The active default adapter contract uses `primary-agent` and current canonical library content. Retired Fortnite defaults, old orchestrator runtime files, obsolete eval surfaces, and archived research/rollback material are not part of generated default runtime output.

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
- A2A remains a config seam only and is not remote/network execution by default.
- Runtime-adjacent state in LazyAI is local: sessions, ledger, memory, messages, metrics, costs, and backups are optional CLI-managed state around the setup.
- Product scope is defined in [Product Boundaries](product-boundaries.md), including the shipped CLI command inventory, active embedded library, repository harness scripts, and retired/archived material.
