# Scopes

`lazyai-cli` supports three setup scopes. Choose the one that matches how you work.

## Overview

| Scope | Best for | Canonical target | Notes |
|---|---|---|---|
| `project` | One repo, self-contained setup | `./.ai/` | Default day-to-day setup |
| `global` | Personal defaults across projects | `~/.ai/` | All 7 LazyAI-supported targets are available globally |
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
  --enable-servers filesystem,ai-memory \
  --no-interactive
```

**Key directories created:** `.ai/`, `specs/`, `.opencode/`, `.claude/`, `.github/`, `.vscode/`, `.ai-setup.db`

## Global scope

**What it is:** personal defaults shared across all projects on your machine.

**When to use it:**
- You want your own baseline AI operating system everywhere
- You use any of the 7 supported tools across many repos
- You want project repos to layer on top of a personal default config
**Canonical target:** `~/.ai/`

**Tool-native targets:**

- `~/.config/opencode/` — OpenCode
- `~/.claude/` — Claude Code
- `~/.copilot/` — GitHub Copilot
- `~/.pi/` — Pi
- `~/.omp/agent/` — OMP
- `~/.kiro/` — Kiro
- `~/.gemini/` — Antigravity
**Example command:**

```bash
lazyai-cli init \
  --scope global \
  --tools opencode,claude-code,copilot,pi,omp,kiro,antigravity \
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
  --workspace-root ./planning-repo \
  --tools opencode,claude-code,copilot \
  --name acme-workspace \
  --preset standard \
  --no-interactive
```

---

## Sidecar and scope behavior

LazyAI can optionally store docs, specs, and plans in a **sidecar** directory instead of the scope root. Sidecar configuration is scope-aware and follows a deterministic resolution chain.

### Resolution priority

When resolving docs/specs/plans directories, LazyAI checks sidecar configuration in this order:

1. **Workspace sidecar** — highest priority, configured in the active workspace entry of `~/.lazyai/workspaces.yaml`
2. **Project sidecar** — configured in `.lazyai-sidecar.yaml` at the project root
3. **Global sidecar** — configured in `~/.lazyai/sidecar.yaml`
4. **Scope default** — if no sidecar is configured at any level:
   - `project` scope → project root
   - `workspace` scope → active workspace path
   - `global` scope → `~/.lazyai/`

### Workspace as the primary use case

The **workspace scope** is the recommended way to use sidecars. In a multi-repo team, you typically have one planning repo that acts as the workspace root. By attaching a sidecar to that workspace, all projects in the workspace share the same docs/specs/plans directory by default.

Example workspace sidecar block in `~/.lazyai/workspaces.yaml`:

```yaml
workspaces:
  - name: acme-workspace
    path: /Users/me/projects/acme-planning
    sidecar:
      path: /Users/me/kb/acme-docs
      specs_dir: specs
      docs_dir: docs
      plans_dir: plans
active: acme-workspace
```

### Optional fallback behavior

Sidecar is **always optional**. If no sidecar is configured at any level, LazyAI falls back to the scope default with no behavior change from today. No sidecar = no errors, no warnings, no migration needed.

### Explicit exclusions

- **No Skeeper or external provider integration.** The sidecar is purely local. There is no `skeeper` field, no provider abstraction, and no remote sync.
- **No content migration.** `sidecar init` does not move existing docs/specs/plans.
- **No multi-sidecar.** One sidecar per scope level.
- **No auto-discovery.** Sidecars are explicitly configured, not detected from parent directories or environment variables.
