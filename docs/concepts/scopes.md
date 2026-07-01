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

LazyAI can optionally store docs, specs, and plans in a **sidecar** directory instead of the scope root. Sidecar configuration is discovered positionally — by directory position, not by name or registration — and resolved with a field-level merge across up to three layers.

### Layer discovery (walk-up)

LazyAI looks for `.lazyai/sidecar.yaml` files at up to three positions relative to the current working directory (`cwd`):

1. **Project** — `cwd` itself. If `cwd/.lazyai/sidecar.yaml` exists, that's the project layer.
2. **Workspace** — the nearest ancestor directory *above* `cwd` (never `cwd` itself) that has a `.lazyai/sidecar.yaml`. LazyAI walks upward one directory at a time, stopping at the first hit. The walk never enters `$HOME` (it stops just before) and never goes past the filesystem root.
3. **Global** — `~/.lazyai/sidecar.yaml`. Always present as a layer (built-in defaults apply if the file itself is absent).

There is no name-based registration and no "active workspace" pointer — whichever ancestor directory happens to contain a `.lazyai/sidecar.yaml` *is* the workspace layer for anything run from below it.

### Field-level merge

Resolution is a per-field merge, not a single winning layer. For each of `docs_dir`, `specs_dir`, and `plans_dir` independently, LazyAI takes the first non-empty value found in this precedence order:

**project > workspace > global > built-in default**

This means a project-level config that sets only `docs_dir` still inherits `specs_dir`/`plans_dir` from the workspace or global layer (whichever set them) rather than losing them. `sidecar status` reports which layer supplied each resolved field.

### Workspace as an ancestor directory

A "workspace" is no longer something you attach or activate by name — it's simply whichever ancestor directory above your project has a `.lazyai/sidecar.yaml`. In a multi-repo team, you typically run `sidecar init --scope workspace` once from the shared planning/parent directory; every project underneath it then automatically picks up its settings on the next resolution, with no separate registration step. Moving or renaming that ancestor directory has no special effect beyond the normal filesystem walk finding (or not finding) it again.

### Optional fallback behavior

Sidecar is **always optional**. If no `.lazyai/sidecar.yaml` is found at any layer, LazyAI falls back to the built-in defaults with no behavior change. No sidecar = no errors, no warnings, no migration needed.

### Explicit exclusions

- **No Skeeper or external provider integration.** The sidecar is purely local. There is no `skeeper` field, no provider abstraction, and no remote sync.
- **No content migration.** `sidecar init` does not move existing docs/specs/plans.
- **No multi-sidecar.** One `.lazyai/sidecar.yaml` per discovered layer.
- **Bounded discovery only.** The walk-up looks for `.lazyai/sidecar.yaml` files only, stopping before `$HOME` or at the filesystem root — it does not read environment variables, and it does not keep climbing past the first workspace hit to look for a more distant layer.
