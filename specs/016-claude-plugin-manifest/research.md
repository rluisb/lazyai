# Spec 016: Claude Plugin Manifest — Research

**Date:** 2026-04-20
**Status:** Research — awaiting human gate
**Sources:** [Plugins reference](https://code.claude.com/docs/en/plugins-reference), [Plugin marketplaces](https://code.claude.com/docs/en/plugin-marketplaces)

---

## §1 — The deferred item

From `specs/KNOWLEDGE_MAP.md` Pending/Follow-up:
> Ship ai-setup as a Claude plugin manifest (deferred from spec 012; plugin schema version + capabilities)

## §2 — What is a Claude plugin?

A plugin is a **self-contained directory** loaded by Claude Code at session start. It bundles five kinds of components into a single installable unit:

| Component | Default location | Purpose |
|---|---|---|
| Skills | `skills/<name>/SKILL.md` | Model-invoked capabilities + `/name` shortcuts |
| Agents | `agents/*.md` | Subagents with frontmatter + system prompt |
| Commands | `commands/*.md` | Flat-file skills (simpler variant) |
| Output styles | `output-styles/*.md` | Response formatting presets |
| Hooks | `hooks/hooks.json` | Event handlers (PostToolUse, SessionStart, …) |
| MCP servers | `.mcp.json` | Stdio/SSE/HTTP MCP configs |
| LSP servers | `.lsp.json` | Language server configs |
| Monitors | `monitors/monitors.json` | Background processes |

Metadata lives in `.claude-plugin/plugin.json`. The manifest is **optional** — only `name` is required when present; if omitted entirely, Claude Code auto-discovers components by default directory names.

### `plugin.json` minimal example

```json
{
  "name": "ai-setup",
  "version": "1.0.0",
  "description": "ai-setup agents, skills, commands, and output styles",
  "author": { "name": "Ricardo Borges" },
  "homepage": "https://github.com/rluisb/lazyai",
  "license": "MIT"
}
```

### Install scopes (mirror the settings-scope hierarchy)

| Scope | Settings file | Who benefits |
|---|---|---|
| `user` | `~/.claude/settings.json` | Just you, all projects (default) |
| `project` | `.claude/settings.json` | Team (committed) |
| `local` | `.claude/settings.local.json` | You on this machine, gitignored |
| `managed` | MDM-deployed | Read-only org policy |

### Distribution

Plugins are installed in three ways:
1. **Ephemeral** — `claude --plugin-dir <path>` for one session.
2. **Marketplace** — `claude plugin install <name>@<marketplace>` (persistent, auto-update).
3. **Local path** — development / testing via `/plugin install <path>`.

Marketplaces are themselves directories (or git repos) containing `.claude-plugin/marketplace.json` that lists plugins and their source locations. Anthropic's own `anthropics/claude-code` repo contains a marketplace.

---

## §3 — How this maps to ai-setup today

ai-setup's `library/` directory already contains most of what a Claude plugin needs:

| ai-setup library path | Plugin default location | Mapping |
|---|---|---|
| `library/agents/*.md` | `agents/*.md` | **Direct** — move/link |
| `library/skills/*.md` | `skills/<name>/SKILL.md` | **Restructure** — one subdir per skill |
| `library/claudecode/commands/*.md` | `commands/*.md` | **Direct** |
| `library/claudecode/output-styles/*.md` | `output-styles/*.md` | **Direct** |
| `library/claudecode/rules/*.md` | N/A (plugin has no `rules/`) | **Gap** — rules are settings-adjacent, not plugin-native |
| `library/mcp/catalog.json` entries | `.mcp.json` | **Transform** — only static servers; skip `${VAR}` placeholders |

**What a plugin can't do** (and what stays in ai-setup's `init` flow):
- Cross-tool scaffolding (OpenCode, Gemini, Copilot, Codex)
- `CLAUDE.md` root placement with org/team placeholder fill
- Deep-merge of `settings.json` / `.mcp.json` with user keys
- Global scope installs (plugins are installed by users, not written by tools)
- Orchestrator wiring (MCP + chain start)

**Conclusion:** a plugin is a **different surface**, not a replacement. It would give Claude-Code-only users a one-command install for the Claude-facing parts; ai-setup stays the cross-tool scaffolder.

---

## §4 — Options

### Option A — Static plugin directory bundled inside the repo

Ship `library/claudecode-plugin/` containing:
- `.claude-plugin/plugin.json`
- `agents/` (copies of `library/agents/*.md`)
- `skills/<name>/SKILL.md` (restructured from flat `.md`)
- `commands/` (copies of `library/claudecode/commands/*.md`)
- `output-styles/` (copies of `library/claudecode/output-styles/*.md`)

Users install via `claude --plugin-dir /path/to/library/claudecode-plugin`. Publishing as a marketplace is a follow-up.

| Pro | Con |
|---|---|
| Simple, no new ai-setup command | Content duplication — same files in two places |
| No build step | Drift risk — flat skills and SKILL.md dirs get out of sync |
| Users skip `ai-setup init` if they only want Claude content | Plugin won't auto-update; users have to pull the repo |

### Option B — `ai-setup build-plugin` command (generate on demand)

Add a new subcommand that **generates** a plugin directory from `library/`:
- Restructures skills into `<name>/SKILL.md` dirs
- Copies agents, commands, output-styles verbatim
- Synthesises a `plugin.json` from ai-setup's version + metadata
- Optionally synthesises `.mcp.json` from `library/mcp/catalog.json` (static servers only)

Output: a ready-to-`claude plugin install <path>` directory. Could target `dist/plugin/` or a user-specified path.

| Pro | Con |
|---|---|
| Single source of truth — no duplicated content | New command to design, test, document |
| Users (and CI) can regenerate the plugin on every ai-setup release | Adds complexity if nobody actually uses the plugin yet |
| Opens door to a marketplace.json generator later | Generator logic is its own maintenance burden |

### Option C — Marketplace-first (publish to anthropics/claude-code or a dedicated repo)

Ship a separate `ai-setup-plugin` repo with `.claude-plugin/marketplace.json` pointing to the plugin directory. Requires either:
- CI that regenerates the plugin on every ai-setup release, OR
- Manual copy-paste at release time.

| Pro | Con |
|---|---|
| Discoverable via `claude plugin search` / marketplace browsers | Two repos to maintain |
| Users get auto-update semantics for free | Overkill if there's no evidence the plugin has demand yet |

### Option D — Skip / document only

Close out the deferred item with a short design note explaining why ai-setup doesn't ship a plugin (or ships a thin stub). Users who want Claude-only installs continue to use `ai-setup init --tools claude-code`.

| Pro | Con |
|---|---|
| Zero code | Leaves a known gap in the "Pending" list |
| Honest about lack of clear user demand | No new distribution channel |

**Recommendation: Option B (build-plugin command).** It avoids duplication, gives ai-setup users a path to publish their own plugin bundles, and sets up Option C (marketplace) as a follow-up. The generator is small (~150 LOC) since the sources already exist.

---

## §5 — Decision interview (need human answers)

1. **Q1** — Pick Option A (static bundle), B (generator command), C (marketplace repo), or D (doc-only)?
2. **Q2** — Which components should the plugin include?
   - **a.** Agents only
   - **b.** Agents + skills + commands + output styles (full Claude-Code surface)
   - **c.** Full + MCP servers from canonical catalog (static servers only; placeholder-bearing servers skipped)
3. **Q3** — Output path for the generated plugin (Option B): `dist/plugin/` default, or user-specified via `--out`?
4. **Q4** — `plugin.json` metadata source: hardcoded in code, or read from `pkg.json`-like metadata file in the repo?
5. **Q5** — Ship a `marketplace.json` alongside (to make the directory publishable to a marketplace repo), or leave that for a future spec?
6. **Q6** — Plugin name: `"ai-setup"` (repo match) or `"rluisb-lazyai"` (more specific)?
7. **Q7** — Should the plugin carry a version number synchronized with ai-setup's binary version, or start at `1.0.0` independently?

---

## §6 — Risks & constraints

| # | Risk | Mitigation |
|---|---|---|
| R1 | Skill-restructure bug: flat `library/skills/<name>.md` → `skills/<name>/SKILL.md` loses content | Unit test: read every source skill, assert the generated SKILL.md has the same body bytes and valid frontmatter `name` field |
| R2 | Plugin and ai-setup install the same agents at different scopes — conflict in `.claude/agents/<name>.md` vs plugin-shipped namespaced `ai-setup:<name>` | Plugin namespacing is automatic (`<plugin>:<agent>`), so no path collision. Document that running both is safe. |
| R3 | MCP servers with `${VAR}` placeholders can't be shipped in a plugin (no prompt-for-value equivalent) | Generator skips placeholder-bearing entries; logs a warning listing skipped servers |
| R4 | Plugin update cadence coupled to ai-setup release, not user's wish | Version matches binary; users can pin to specific version via marketplace entry |

---

## §7 — Out of scope (for this spec)

- Automated marketplace publishing (push to GitHub on release) — follow-up only
- Plugin hooks that dynamically re-scaffold via `ai-setup` CLI — no user value, complex
- LSP server bundles — ai-setup doesn't own any language servers
- Plugin userConfig for org/team (Claude plugin feature) — ai-setup handles this via `init --org --team`
