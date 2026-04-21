# Spec 008 — AI CLI Tool Structure Parity (Research)

> **Phase:** R of RPI. This document ends at the Human Gate — no planning or implementation until explicitly approved.
>
> **Scope:** Ensure `ai-setup` scaffolds each supported AI CLI tool's on-disk layout correctly for each `SetupScope` (project, global, workspace), and evaluate whether to drive scaffolding via each tool's own CLI commands (e.g. `gemini mcp add`, `claude mcp add`) or continue direct file writes.
>
> **Date:** 2026-04-17
> **Branch:** `feature/go-migration`

---

## 1. Supported Tools

Per `internal/adapter/registry.go:22-26` and `internal/types/types.go:29-34`, five tools are registered:

| ToolId | Adapter file | Memory file |
|---|---|---|
| `opencode` | `internal/adapter/opencode.go` | `AGENTS.md` |
| `claude-code` | `internal/adapter/claudecode.go` | `CLAUDE.md` |
| `gemini` | `internal/adapter/gemini.go` | `GEMINI.md` |
| `copilot` | `internal/adapter/copilot.go` | `.github/copilot-instructions.md` |
| `codex` | `internal/adapter/codex.go` | `AGENTS.md` |

`SetupScope` values: `project`, `global`, `workspace` (`internal/types/types.go:19-23`).

---

## 2. Canonical On-Disk Layout per Tool

Sources: upstream docs for each CLI (URLs embedded in §5). The table below is the *target* structure we want `ai-setup` to produce. "Workspace" throughout means the user-selected workspace directory treated as a project-shaped layout — there is no tool-native "workspace" concept; it is our own scope.

### 2.1 Claude Code

| Artifact | Project | Global | Workspace |
|---|---|---|---|
| Agents | `.claude/agents/<name>.md` | `~/.claude/agents/<name>.md` | `<ws>/.claude/agents/<name>.md` |
| Skills | `.claude/skills/<name>/SKILL.md` | `~/.claude/skills/<name>/SKILL.md` | `<ws>/.claude/skills/<name>/SKILL.md` |
| Slash commands | `.claude/commands/<name>.md` | `~/.claude/commands/<name>.md` | `<ws>/.claude/commands/<name>.md` |
| Hooks | `.claude/settings.json` `hooks` key | `~/.claude/settings.json` `hooks` key | `<ws>/.claude/settings.json` |
| MCP servers | `.claude/mcp.json` or `.claude/mcp/<name>.json` | `~/.claude/mcp.json` | `<ws>/.claude/mcp.json` |
| Settings (checked in) | `.claude/settings.json` | `~/.claude/settings.json` | `<ws>/.claude/settings.json` |
| Settings (local, gitignored) | `.claude/settings.local.json` | — | `<ws>/.claude/settings.local.json` |
| Memory / instructions | `CLAUDE.md` at project root | `~/.claude/CLAUDE.md` | `<ws>/CLAUDE.md` |

Frontmatter: agents and skills use YAML frontmatter; `name` + `description` required. Hooks/settings/MCP use JSON.

### 2.2 OpenCode

| Artifact | Project | Global | Workspace |
|---|---|---|---|
| Agents | `.opencode/agents/<name>.md` | `~/.config/opencode/agents/<name>.md` | `<ws>/.opencode/agents/<name>.md` |
| Skills | `.opencode/skills/<name>/SKILL.md` | `~/.config/opencode/skills/<name>/SKILL.md` | `<ws>/.opencode/skills/<name>/SKILL.md` |
| Slash commands | `.opencode/commands/<name>.md` | `~/.config/opencode/commands/<name>.md` | `<ws>/.opencode/commands/<name>.md` |
| Plugins (≈ hooks) | `.opencode/plugins/*.{js,ts}` | `~/.config/opencode/plugins/` | `<ws>/.opencode/plugins/` |
| MCP servers | `mcp` key in `opencode.json` at project root | `mcp` key in `~/.config/opencode/opencode.json` | `mcp` key in `<ws>/opencode.json` |
| Settings | `opencode.json[c]` at project root | `~/.config/opencode/opencode.json` | `<ws>/opencode.json` |
| Memory / instructions | `AGENTS.md` at project root | `~/.config/opencode/AGENTS.md` | `<ws>/AGENTS.md` |

Note: OpenCode has *no native hooks surface*; extensibility is via plugin JS/TS modules. Skills `name` must match directory and regex `^[a-z0-9]+(-[a-z0-9]+)*$`. Prefer plural dir names (`agents/`, `commands/`, `plugins/`, `skills/`).

### 2.3 Gemini CLI

| Artifact | Project | Global | Workspace |
|---|---|---|---|
| Agents | Distributed via extensions (`<ext>/agents/*.md`); no first-class `.gemini/agents/` path | Same, installed under `~/.gemini/extensions/<name>/` | Same |
| Skills | `.gemini/skills/<name>/SKILL.md` (alias `.agents/skills/...` — *pick one*) | `~/.gemini/skills/<name>/SKILL.md` | `<ws>/.gemini/skills/<name>/SKILL.md` |
| Slash commands | `.gemini/commands/<name>.toml` (subdirs → `:` namespace) | `~/.gemini/commands/<name>.toml` | `<ws>/.gemini/commands/<name>.toml` |
| Hooks | `hooks` key in `.gemini/settings.json` | `hooks` key in `~/.gemini/settings.json` | `hooks` key in `<ws>/.gemini/settings.json` |
| MCP servers | `mcpServers` key in `.gemini/settings.json` | `mcpServers` key in `~/.gemini/settings.json` | `mcpServers` key in `<ws>/.gemini/settings.json` |
| Settings | `.gemini/settings.json` | `~/.gemini/settings.json` | `<ws>/.gemini/settings.json` |
| Memory / instructions | `GEMINI.md` at project root (configurable) | `~/.gemini/GEMINI.md` *if configured* | `<ws>/GEMINI.md` |

Gemini has system-wide `/etc/gemini-cli/settings.json` — out of scope for this feature.

### 2.4 OpenAI Codex CLI

| Artifact | Project | Global | Workspace |
|---|---|---|---|
| Agents | `AGENTS.md` + optional `AGENTS.override.md` anywhere from repo root down | `~/.codex/AGENTS.md` + `~/.codex/AGENTS.override.md` | `<ws>/AGENTS.md` |
| Skills | `.agents/skills/<name>/SKILL.md` at repo root | `~/.agents/skills/<name>/SKILL.md` (note: under `$HOME/.agents/`, not `$HOME/.codex/`) | `<ws>/.agents/skills/<name>/SKILL.md` |
| Slash commands | TBD — not documented publicly | TBD | TBD |
| Hooks | `.codex/hooks.json` (feature-flagged; schema undocumented) | `~/.codex/hooks.json` | `<ws>/.codex/hooks.json` |
| MCP servers | `[mcp_servers.*]` in `.codex/config.toml` (trusted projects only) | `[mcp_servers.*]` in `~/.codex/config.toml` | `[mcp_servers.*]` in `<ws>/.codex/config.toml` |
| Settings | `.codex/config.toml` | `~/.codex/config.toml` (`$CODEX_HOME` overrides) | `<ws>/.codex/config.toml` |
| Memory / instructions | `AGENTS.md` | `~/.codex/AGENTS.md` | `<ws>/AGENTS.md` |

Anomaly worth flagging: Codex skills live under `$HOME/.agents/skills/`, *not* `$HOME/.codex/skills/`. This is the only tool where the global skills dir is outside the tool's own config dir.

### 2.5 GitHub Copilot

| Artifact | Project | Global | Workspace |
|---|---|---|---|
| Repo-wide instructions | `.github/copilot-instructions.md` | *VS Code client-only; not scaffold-target* | `<ws>/.github/copilot-instructions.md` |
| Path-specific instructions | `.github/instructions/<name>.instructions.md` | — | `<ws>/.github/instructions/...` |
| Prompt files (slash-command equivalent) | `.github/prompts/<name>.prompt.md` | — | `<ws>/.github/prompts/...` |
| Custom chat modes / agents | `.github/chatmodes/<name>.chatmode.md` *(legacy)* → `.github/agents/<name>.agent.md` *(current)* | — | `<ws>/.github/agents/...` |
| Hooks | — (none) | — | — |
| MCP servers | — (not in-repo) | — | — |
| Settings | — (client-side only) | — | — |
| Memory / instructions fallback | `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` at repo root (read by Copilot VS Code) | — | `<ws>/AGENTS.md` |

Copilot has **no CLI and no first-class global config file** for scaffolding purposes. Only project and workspace are meaningful.

---

## 3. What `ai-setup` Emits Today

Summary of the explorer pass over `internal/adapter/*.go`, `internal/scaffold/{root,filemap,scaffold}.go`, and `internal/globalpaths/globalpaths.go`.

### 3.1 Scope coverage matrix

| Tool | Project | Global | Workspace |
|---|---|---|---|
| Claude Code | ✅ `<target>/.claude/` | ✅ `~/.claude/` (via `HomeDir`) | ❌ silently falls back to project layout |
| OpenCode | ✅ `<target>/.opencode/` | ✅ `~/.config/opencode/` (via `ResolveGlobalToolTargetDir`) | ❌ silently falls back to project layout |
| Codex | ✅ `<target>/.agents/` (note: `.agents/`, not `.codex/`) | ⚠ `filepath.Join(filepath.Dir(ctx.TargetDir), ".agents")` — writes one level *above* TargetDir | ❌ falls back to the parent-dir logic |
| Copilot | ✅ `<target>/.github/` | ⚠ no scope guard — writes project layout even when scope=global | ❌ falls back to project layout |
| Gemini | ✅ `<target>/.gemini/` | ⚠ no scope guard — writes project layout even when scope=global | ❌ falls back to project layout |

`internal/globalpaths/globalpaths.go:49-58` only resolves global paths for `opencode` and `claude-code`; `gemini`, `copilot`, `codex` return empty. `LogUnsupportedGlobalTool` (lines 66-73) warns for Copilot/Gemini but not Codex, and does not prevent the adapters from writing.

### 3.2 Root memory-file scaffolding

`internal/scaffold/root.go` writes the memory doc (AGENTS.md / CLAUDE.md / GEMINI.md / `.github/copilot-instructions.md`) by calling `filepath.Join(opts.TargetDir, outputFile)` at `root.go:168`. This is **scope-blind**: global-scope installs still write the doc under `<target>` instead of under the tool's global directory. ClaudeCode's adapter partially compensates at `claudecode.go:134-137` by overriding the path, creating duplication risk.

### 3.3 Concrete gaps (to be addressed by the forthcoming plan)

1. **Workspace scope is unimplemented across all five adapters.** User's intent: workspace = "same layout as project, rooted at the user-chosen workspace directory." Today, `scaffold.go:96-110` only uses `SetupScopeWorkspace` to trigger repo-ledger scaffolding, not per-tool config.
2. **Gemini and Copilot have no global-scope guard** at the adapter layer — they write project-shaped output at `<target>/` when scope=global, which is wrong because Copilot has no global concept and Gemini's global dir is `~/.gemini/`, not `<target>/.gemini/`.
3. **Codex global-scope path is a parent-dir fallback** (`codex.go:37`), not `~/.codex/` or `~/.agents/`. Matches neither the tool's convention nor the pattern used by Claude/OpenCode.
4. **Gemini global path not registered** in `globalpaths.ResolveGlobalToolTargetDir` even though Gemini *does* have a canonical `~/.gemini/` structure upstream.
5. **Codex global path not registered** in `globalpaths.ResolveGlobalToolTargetDir`; there is a real `~/.codex/` and `~/.agents/skills/` to target.
6. **Root memory doc is scope-blind.** `root.go:168` always writes to `<target>`. For global scope Claude/OpenCode this should emit under the tool's global dir; for Gemini global scope same.
7. **`.agents/` vs `.codex/` confusion for Codex.** Skills live under `.agents/skills/`; settings live under `.codex/config.toml`. Current adapter uses `.agents/` as the only directory and never emits `config.toml`.
8. **No adapter currently emits Codex `config.toml` or the `[mcp_servers.*]` tables** that Codex actually reads for MCP registration.
9. **Settings/hooks/MCP merge semantics are inconsistent.** `settings.json` and `opencode.json` must be *merged* (not overwritten) when ai-setup runs against an existing install. Today the adapters do not have a shared merge helper; each does its own thing.

---

## 4. Can We Drive Each Tool's Own CLI to Bootstrap?

User's idea: rather than writing files directly, invoke `claude mcp add …`, `gemini skills install …`, etc., feeding them our library content. Findings:

| Tool | Non-interactive CLI for agents/skills/commands/MCP? | Verdict |
|---|---|---|
| Claude Code | `claude mcp add\|remove\|list` exist but are **interactive prompts**. `/agents` is in-session UI only. `--agents` accepts inline JSON but is session-scoped, not persisted. | **Stay direct-write.** No non-interactive, file-input path. |
| OpenCode | `opencode agent create` and `opencode mcp add` are **interactive only**. No `opencode init`, no `command add`. | **Stay direct-write.** |
| Gemini CLI | `gemini mcp add <name> <cmd> [args...]` **argv-driven, non-interactive**. `gemini skills install <git-url\|path\|.skill> --scope workspace\|user` **non-interactive**. `gemini extensions install <src>` **non-interactive**. | **Opportunity to drive CLI for MCP + skills + extensions.** Still need direct writes for `.gemini/commands/*.toml` and settings hooks. |
| Codex | No documented shell subcommands for artifact management. `/init` is an in-session slash command, not shell-invokable. | **Stay direct-write.** |
| Copilot | No CLI at all. | **Must be direct-write.** |

**Implication for design.** Direct writes are the only universal approach. Driving each tool's CLI is only a viable *opt-in optimisation* for Gemini (and partially Claude if we accepted interactivity, which we don't want in CI). The recommended shape of the forthcoming plan is:

- **Primary path:** direct file writes, parameterised by a scope-resolver that returns the correct base dir per `(tool, scope)`.
- **Optional enhancement (out of Wave 1):** a `--drive-cli` flag that, when the target CLI is installed, calls `gemini mcp add` / `gemini skills install` / `gemini extensions install` to register the artifacts. No parity behaviour attempted for Claude/OpenCode/Codex/Copilot because their CLIs are interactive or absent.

This keeps the scaffolder deterministic, idempotent, diff-friendly, offline-capable, and CI-safe, while leaving a seam for users who prefer the tool's own registration semantics.

---

## 5. Sources

Authoritative docs consulted (all retrieved 2026-04-17):

**Claude Code** — https://code.claude.com/docs/en/claude-directory , https://code.claude.com/docs/en/settings , https://code.claude.com/docs/en/skills , https://code.claude.com/docs/en/sub-agents , https://code.claude.com/docs/en/hooks , https://code.claude.com/docs/en/mcp , https://code.claude.com/docs/en/cli-reference

**OpenCode** — https://opencode.ai/docs/ , https://github.com/sst/opencode/tree/dev/packages/web/src/content/docs (agents, skills, commands, plugins, rules, mcp-servers, cli)

**Gemini CLI** — https://github.com/google-gemini/gemini-cli/tree/main/docs (cli/settings, cli/skills, cli/custom-commands, cli/cli-reference, extensions/reference, hooks/index, tools/mcp-server)

**OpenAI Codex** — https://developers.openai.com/codex/config-reference , https://developers.openai.com/codex/guides/agents-md , https://developers.openai.com/codex/skills , https://github.com/openai/codex/blob/main/docs/config.md

**GitHub Copilot** — https://docs.github.com/en/copilot/how-tos/configure-custom-instructions/add-repository-instructions , https://code.visualstudio.com/docs/copilot/customization/prompt-files , https://code.visualstudio.com/docs/copilot/customization/custom-chat-modes

---

## 6. Open Questions for the User (blocking the Plan phase)

1. **Workspace semantics.** Confirming the user's brief: workspace = "same layout as project, but the writable root is the user-selected workspace directory (not the current repo)." Do we need the tool memory docs (AGENTS.md/CLAUDE.md/GEMINI.md) to live at the *workspace* root, or also mirrored into each repo listed in `Config.Repos`?
2. **Scope unsupported by a tool.** When the user picks `scope=global` + Copilot, do we (a) reject the combination in the wizard, (b) warn and skip that adapter, or (c) silently map it to project? Today the behaviour is silent and wrong.
3. **Codex global path.** Pick one canonical target: `~/.codex/` (settings + `AGENTS.md`) plus `~/.agents/skills/` (skills), or consolidate everything under `~/.codex/`. The upstream docs use the split; our current `filepath.Dir(TargetDir)` logic matches neither.
4. **Settings/config merge policy.** For existing `settings.json` / `opencode.json` / `.codex/config.toml` files, confirm the merge strategy: preserve user keys, add ours, error on conflict? Or overwrite with a backup?
5. **CLI-driven scaffolding (Gemini).** Is a `--drive-cli` flag worth building in Wave 1, or deferred? Recommendation: defer; ship direct-writes first.
6. **Spec directory convention.** Filed at `specs/008-cli-tool-structure-parity/` to match the existing top-level `specs/NNN-*` pattern (001–007), rather than `specs/features/NNN-*/` from the CLAUDE.md decision tree. Confirm which convention to keep going forward — the mismatch itself is worth a separate cleanup.

---

## 7. Human Gate

Per `specs/bugfixes/AGENTS.md` RPI flow: **stop here.** Do not proceed to `plan.md` until the user reviews §3 (gaps), §4 (CLI-driving viability), and §6 (open questions) and gives explicit approval or redirection.
