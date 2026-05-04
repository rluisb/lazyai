# Scopes

`lazyai-cli` supports three setup scopes. Choose the one that matches how you work.

## Overview

| Scope | Best for | Canonical target | Notes |
|---|---|---|---|
| `project` | One repo, self-contained setup | `./.ai/` | Default day-to-day setup |
| `global` | Personal defaults across projects | `~/.ai/` | Only OpenCode + Claude Code are supported globally |
| `workspace` | Multi-repo team coordination | `planning-repo/.ai/` | Planning repo becomes the hub |

## Project scope

**What it is:** the default, self-contained setup in the current working directory.

**When to use it:**

- You want everything versioned with the repo
- You want generated instructions checked into source control
- You need all supported tools available side-by-side

**Example command:**

```bash
lazyai-cli init \
  --scope project \
  --tools opencode,claude-code,copilot \
  --name my-app \
  --preset standard \
  --enable-servers orchestrator \
  --no-interactive
```

**Key directories created:** `.ai/`, `specs/`, `.opencode/`, `.claude/`, `.github/`, `.vscode/`, `.ai-setup.json`

## Global scope

**What it is:** personal defaults shared across all projects on your machine.

**When to use it:**

- You want your own baseline AI operating system everywhere
- You use OpenCode or Claude Code across many repos
- You want project repos to layer on top of a personal default config

**Canonical target:** `~/.ai/`

**Tool-native targets:**

- `~/.config/opencode/`
- `~/.claude/`

**Example command:**

```bash
lazyai-cli init \
  --scope global \
  --tools opencode,claude-code \
  --name global \
  --preset minimal \
  --no-interactive
```

## Workspace scope

**What it is:** a multi-repo team setup with a dedicated planning repo as the central hub.

**When to use it:**

- Your team coordinates work across multiple repositories
- You want one planning repo for specs, ADRs, memory, and ledgers
- You want a single place for all AI tool configuration

**How it works:**

- Everything lives in the **workspace root** (the planning repo) — referenced repos are never touched
- The workspace root gets the full canonical setup: `.ai/`, specs, tool directories, MCP configs
- Referenced repos are scanned for stack detection (language, framework, commands)
- Detected repo info is included in compiled root files so AI agents know what is in the workspace
- Per-repo ledgers and state snapshots are written under `specs/memory/repos/`
- Launch your AI tool from the workspace root

**Example command:**

```bash
lazyai-cli init \
  --scope workspace \
  --planning-repo ./planning-repo \
  --repos ../api,../web,../worker \
  --tools opencode,claude-code,copilot \
  --name acme-workspace \
  --preset standard \
  --no-interactive
```
