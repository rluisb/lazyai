# Pi Adapter

**Adapter ID:** `pi`  
**Source:** `packages/cli/internal/adapter/pi.go`  
**Status:** stable  
**Config directory:** `.pi`

## Overview

The Pi adapter generates native configuration for [Pi](https://pi.ai) (the `pi` CLI). It emits agents, skills, prompts, and TypeScript extensions into `.pi/`. Pi supports project, workspace, and global scopes.

## Generated Files

| Path | Description |
|---|---|
| `.pi/agents/<name>.md` | Agent definitions (canonical agents) |
| `.pi/skills/<name>/SKILL.md` | Skill directories |
| `.pi/prompts/<name>.md` | Prompt templates |
| `.pi/extensions/<name>.ts` | Flat TypeScript extensions (safety hooks) |
| `.pi/extensions/<name>/index.ts` | Directory-layout extensions (multi-file) |
| `.pi/SYSTEM.md` | Project system prompt (replaces Pi default) |
| `.pi/APPEND_SYSTEM.md` | Appended system prompt (extends Pi default) |
| `.pi/settings.json` | Project settings (resource references for extensions/skills/prompts) |
| `~/.pi/agent/settings.json` | Global settings (same resource references, resolved relative to `~/.pi/agent`) |

## Supported Asset Types

| Asset kind | Shape | Destination |
|---|---|---|
| Agents | flat | `.pi/agents/<name>.md` |
| Skills | dir-per-item | `.pi/skills/<name>/SKILL.md` |
| Prompts | flat | `.pi/prompts/<name>.md` |
| Templates | none | — |
| Commands | none | — |
| Chat modes | none | — |
| Output styles | none | — |
| System prompts | flat | `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md` |

## MCP Behavior

**Pi MCP is a no-op.** The `CompileMCP` method returns `ctx.FileRecords` unchanged. The adapter does **not** declare MCP capability, and no MCP configuration is emitted for Pi because Pi has no native MCP surface.

## Extension Layouts

Pi has no `.pi/hooks` path. Safety hooks ship as TypeScript extensions under `.pi/extensions/`. The adapter copies the `pi/extensions/` library directory recursively, preserving both supported Pi layouts:

| Layout | Library shape | Destination |
|---|---|---|
| Flat | `pi/extensions/<name>.ts` | `.pi/extensions/<name>.ts` |
| Directory | `pi/extensions/<name>/index.ts` (+ co-located modules, `package.json`) | `.pi/extensions/<name>/index.ts` (+ co-located files) |

Pi auto-discovers both layouts from `.pi/extensions/` (project-local) and `~/.pi/agent/extensions/` (global). The `settings.json` `"extensions": ["extensions"]` resource reference (see #532) points Pi at this directory; both layouts load without additional settings keys.

### Extension-local dependencies

Co-located `package.json` files are copied verbatim so directory extensions can declare dependencies. However, LazyAI does **not** run `npm install` or resolve `node_modules/` — dependency installation is the operator's responsibility after compile. Pi resolves imports from `node_modules/` at runtime once installed.

The following Pi extension reference mechanisms are **out of scope** for this adapter and explicitly unsupported until the settings/package contract lands (#532/#537):

- `settings.json` `"extensions"` array entries pointing at arbitrary local paths outside `.pi/extensions/`
- `settings.json` `"packages"` array (npm/git package-managed extensions)

Library extensions must live under `pi/extensions/` and ship as source (`.ts`); the adapter does not transpile or bundle.

## Skill Behavior

Skills are written as Agent Skills-compatible directories: `.pi/skills/<name>/SKILL.md`. The adapter copies selected skills from the canonical library.

## Agent Behavior

Canonical agents are written as flat markdown files under `.pi/agents/`. Pi's subagent extension reads agent definitions from this directory.

## Prompt Behavior

Prompt templates are copied as flat markdown files to `.pi/prompts/`. The adapter supports selection filtering.

## Settings Behavior

Pi uses JSON settings files with project settings overriding global settings on a per-key basis. LazyAI emits settings at both scopes:

- **Project scope:** `.pi/settings.json` — paths resolve relative to `.pi`
- **Global scope:** `~/.pi/agent/settings.json` — paths resolve relative to `~/.pi/agent`

The adapter deep-merges the LazyAI-managed resource references (`extensions`, `skills`, `prompts`) into any existing settings via `configmerge.MergeJSONFile`, preserving user-authored keys and producing idempotent output on re-runs. A `.bak` sidecar is created on first touch.

The default settings patch declares only resource directory references pointing at the on-disk subdirectories Install already populated. Model, theme, compaction, and other personal preferences stay user-owned. Sibling issues extend this patch map: #537 adds package configuration, #533 may adjust extension references.

## System Prompt Behavior

Pi reads two project system-prompt files from the `.pi/` root:

- `.pi/SYSTEM.md` — **replaces** Pi's built-in default system prompt entirely. When present, Pi stops auto-loading `AGENTS.md`/`CLAUDE.md` context files, skills, and extensions discovered from the project. Use only when you want full control over the prompt.
- `.pi/APPEND_SYSTEM.md` — **appends** to the default system prompt without replacing it. Pi still auto-discovers context files, skills, and extensions. This is the recommended way to add project-specific instructions.

Both are distinct from generic root instructions (`AGENTS.md`/`CLAUDE.md`), which Pi loads as context files regardless of project trust. The adapter sources them from `pi/SYSTEM.md` and `pi/APPEND_SYSTEM.md` in the library; when either is absent, no file is emitted (omission is silent).

These files require project trust before Pi loads them interactively.

## Scope Support

| Scope | Supported |
|---|---|
| Project | yes |
| Workspace | yes |
| Global | yes |

## Headless Support

No (`CanRunHeadless() = false`).

## Known Limitations

- **MCP is a no-op** — no MCP configuration is emitted for Pi, and the adapter intentionally does not declare MCP capability
- No templates, commands, or chat modes
- No `.pi/hooks` path; hooks ship as TypeScript extensions
- Pi project trust is not a sandbox (Pi loads trusted project resources, but execution still runs with user permissions)
- Extension-local `node_modules/` are not installed by the adapter (operator runs `npm install`)
- `settings.json` `"packages"` and arbitrary-path `"extensions"` entries are unsupported until #537

## Test Coverage

| Test file | What it verifies |
|---|---|
| `pi_adapter_test.go` | Agents, skills, prompts, flat + directory-layout extensions, settings emission at project/global scope, system prompts, and non-leakage between surfaces |
| `adapter_adapters_test.go` | Full install from FS |
