# LazyAI

Scaffold a canonical, multi-tool AI development environment from one CLI, with optional orchestration scaffolding and MCP runtime integration.

`lazyai-cli` initializes and maintains a **tool-agnostic canonical layer** under `.ai/`, then compiles it into native formats for OpenCode, Claude Code, and GitHub Copilot. It ships with bundled agents, skills, templates, rules, and optional MCP servers so teams can adopt a consistent AI operating system with minimal configuration.

## What it does

- **One-time setup**: `lazyai-cli init` scaffolds canonical files, tool-native directories, and an MCP catalog.
- **Compile**: `lazyai-cli compile` regenerates per-tool configs from the canonical source of truth.
- **Update**: `lazyai-cli update` refreshes library content while preserving customizations.
- **Doctor**: `lazyai-cli doctor` checks drift, missing files, and skill state.
- **Orchestration** (opt-in): `lazyai-cli init --enable-servers orchestrator` scaffolds chain/team/workflow definitions and registers the local orchestrator MCP runtime.

## Where to start

- [Quick Start](getting-started/quick-start.md) — run your first setup in minutes
- [Installation](getting-started/installation.md) — install options and prerequisites
- [How It Works](concepts/how-it-works.md) — canonical source, compile model, and manifest tracking
- [CLI Reference](cli/reference.md) — full command and flag documentation
- [GitHub Wiki](https://github.com/rluisb/lazyai/wiki) — short-form operational notes and release/install references

## Architecture at a glance

```text
lazyai-cli (Go binary)
   │ init ──▶ .ai/ (canonical)
   │ compile ──▶ .opencode/ + .claude/ + .github/ + .vscode/
   │ doctor ──▶ .ai-setup.json / .ai-setup.db (manifest + SQLite)
   │
   └── optional orchestrator MCP server
         └── lazyai-orchestrator (Go runtime)
             └── catalog, state, handoffs, prompt composition
```

The execution path uses **local native agents** (Claude Code, OpenCode, Copilot) directly. A2A is an optional config seam only; remote/network execution is not the default.

## Status

- CLI: Go module at `github.com/rluisb/lazyai/packages/cli`
- Platforms: macOS, Linux
- License: MIT
