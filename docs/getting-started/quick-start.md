# Quick Start

> **Important:** `ai-setup` is **not published to npm**. Install and run it directly from GitHub with `npx`.

## Interactive setup

```bash
npx github:ricardoborges-teachable/ai-setup init
```

This launches a wizard where you choose:

- **Scope**: `project`, `global`, or `workspace`
- **Tools**: OpenCode, Claude Code, GitHub Copilot
- **Preset**: `minimal`, `standard`, or `full`
- **Optional MCP servers**: filesystem, ripgrep, memory, orchestrator, etc.

## Non-interactive setup

### Project scope (default)

```bash
npx github:ricardoborges-teachable/ai-setup init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --enable-servers orchestrator \
  --name my-app \
  --preset standard \
  --no-interactive
```

### Global scope

```bash
npx github:ricardoborges-teachable/ai-setup init \
  --scope global \
  --tools opencode,claude-code \
  --name global \
  --preset minimal \
  --no-interactive
```

### Workspace scope

```bash
npx github:ricardoborges-teachable/ai-setup init \
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
7. Scaffolds `.ai/orchestration/` when the optional `orchestrator` MCP server is enabled

## Verify the setup

```bash
ai-setup status
ai-setup doctor
```

## Next steps

- Read [How It Works](../concepts/how-it-works.md) to understand the canonical source model.
- Browse [Scopes](../concepts/scopes.md) to choose the right setup mode.
- See [MCP Integration](../integration/mcp.md) to configure optional servers.
