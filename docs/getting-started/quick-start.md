# Quick Start

> **Important:** LazyAI is Go-only. Install `lazyai-cli` with `go install`; npm/npx distribution has been removed.

## Install

```bash
go install github.com/rluisb/lazyai/packages/cli/cmd/lazyai-cli@latest
```

## Interactive setup

```bash
lazyai-cli init
```

This launches a wizard where you choose:

- **Scope**: `project`, `global`, or `workspace`
- **Tools**: OpenCode, Claude Code, GitHub Copilot
- **Preset**: `minimal`, `standard`, or `full`
- **Optional MCP servers**: filesystem, ripgrep, memory, and other catalog entries

## Non-interactive setup

### Project scope (default)

```bash
lazyai-cli init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --enable-servers filesystem,memory \
  --name my-app \
  --preset standard \
  --no-interactive
```

### Global scope

```bash
lazyai-cli init \
  --scope global \
  --tools opencode,claude-code \
  --name global \
  --preset minimal \
  --no-interactive
```

### Workspace scope

```bash
lazyai-cli init \
  --scope workspace \
  --planning-repo ./planning-repo \
  --repos ../app-one,../app-two \
  --tools opencode,claude-code \
  --name team-workspace \
  --preset standard \
  --no-interactive
```

## What happens after `init`

1. Creates canonical, tool-agnostic files under `.ai/`
2. Scaffolds specs, templates, rules, and infrastructure based on the selected preset
3. Compiles root instructions for each selected tool
4. Generates tool-native directories such as `.opencode/`, `.claude/`, `.github/`, and `.vscode/`
5. Writes `.ai-setup.json` to track managed files, hashes, selections, and operations
6. Generates `.env.example` when enabled MCP servers require environment variables

## OpenCode default behavior

When OpenCode is selected during `init`, LazyAI installs the neutral canonical adapter path. The default install includes:

- Canonical agents, including `primary-agent`
- `.opencode/opencode.jsonc` with `default_agent: primary-agent`

Fortnite agents, `.opencode/STARTUP.md`, and `loop-driver` are not installed by default.

Existing safety constraints still apply: no push/deploy by default, and project files are created with no-overwrite behavior where applicable.

## Verify the setup

```bash
lazyai-cli status
lazyai-cli doctor
```

## Next steps

- Read [How It Works](../concepts/how-it-works.md) to understand the canonical source model.
- Browse [Scopes](../concepts/scopes.md) to choose the right setup mode.
- See [MCP Integration](../integration/mcp.md) to configure optional servers.
