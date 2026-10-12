# Which harnesses load Agent Plugins natively?

Research ticket: [#633](https://github.com/rluisb/lazyai/issues/633) (map [#632](https://github.com/rluisb/lazyai/issues/632)).
Access date for all sources: 2026-10-12. Facts only; no design recommendations.

Legend: **Yes** = documented by the vendor (and/or verified locally); **No** = vendor docs describe a different format and do not mention support; **Unverified** = not confirmed from a primary source.

## The two standards

| Standard | Version | What it defines | Source |
|---|---|---|---|
| Agent Plugins | **1.0.0** (published 2026-08-06) | Root `plugin.json` with `$schema: https://agent-plugins.org/schemas/1.0.0/plugin.schema.json` and `name` (required); components are only `skills/<name>/SKILL.md` and root `mcp.json`. Commands, hooks, agents, rules, LSP are explicitly outside v1; client-specific content goes under reverse-domain namespaces (`extensions.<ns>` / `<ns>/` dir). No install/discovery paths defined. | https://agent-plugins.org/specification ; https://vercel.com/blog/introducing-agent-plugins |
| Agent Skills | Unversioned (the spec page carries no version number) | Folder with `SKILL.md` (frontmatter `name`, `description`; optional `license`, `compatibility`, `metadata`, `allowed-tools`). No discovery paths mandated. | https://agentskills.io/specification ; https://agentskills.io (client showcase) |

Vercel's launch post lists the day-one Agent Plugins clients as: ChatGPT and Codex, Cursor, GitHub Copilot, Kiro, VS Code. It does not mention Claude Code, OpenCode, Pi or Antigravity.

## Summary table

| Harness (version checked) | Agent Plugins 1.0.0 | Agent Skills | AGENTS.md | Skills (project / user) | Subagents (project / user) | Hooks (project / user) | Commands (project / user) | MCP (project / user) | Context files |
|---|---|---|---|---|---|---|---|---|---|
| **Claude Code** 2.1.296 | **No** (own format, `.claude-plugin/plugin.json`) | Yes | Yes, fallback only (since 2.1.277) | `.claude/skills/` (+ nested `<subdir>/.claude/skills/`) / `~/.claude/skills/` | `.claude/agents/` / `~/.claude/agents/` | `.claude/settings.json`, `.claude/settings.local.json` / `~/.claude/settings.json` | `.claude/commands/` (legacy, merged into skills) / `~/.claude/skills/` | `.mcp.json` / `~/.claude.json` | `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md`, `.claude/rules/` / `~/.claude/CLAUDE.md`, `~/.claude/rules/`; `AGENTS.md` only when no CLAUDE.md |
| **Codex** CLI 0.162.1 | **Yes** (root `plugin.json`; `.codex-plugin/plugin.json` = compat fallback) | Yes | Yes (native) | `.agents/skills/` (cwd up to repo root) / `~/.agents/skills/` (+ `/etc/codex/skills`) | `.codex/agents/*.toml` / `~/.codex/agents/*.toml` | `.codex/hooks.json` or `[hooks]` in `.codex/config.toml` / `~/.codex/hooks.json` or `~/.codex/config.toml` | none (project) / `~/.codex/prompts/` (deprecated) | `.codex/config.toml` (trusted) / `~/.codex/config.toml` | `AGENTS.override.md`, `AGENTS.md` (+ fallback names) per dir / `~/.codex/AGENTS(.override).md` |
| **Pi** 1.1.0 | **No** (Pi packages: `package.json` `pi` key) | Yes | Yes (native, also CLAUDE.md) | `.pi/skills/`, `.agents/skills/` / `~/.pi/agent/skills/`, `~/.agents/skills/` | Not built in | Extension events (`pi.on`) in `.pi/extensions/` / `~/.pi/agent/extensions/` | `.pi/prompts/` / `~/.pi/agent/prompts/` | `.pi/mcp.json` / `~/.pi/agent/mcp.json` | `AGENTS.override.md`, `AGENTS.md`, `CLAUDE.md` in cwd + parents / same names in `~/.pi/agent/` |
| **OpenCode** 1.17.9 | **No** (plugins are JS/TS modules) | Yes | Yes (native, CLAUDE.md fallback) | `.opencode/skills/`, `.claude/skills/`, `.agents/skills/` / `~/.config/opencode/skills/`, `~/.claude/skills/`, `~/.agents/skills/` | `.opencode/agents/` / `~/.config/opencode/agents/` (or `agent` key in config) | Plugin hooks in `.opencode/plugins/` / `~/.config/opencode/plugins/` | `.opencode/commands/` / `~/.config/opencode/commands/` | `mcp` key in `opencode.json` / `~/.config/opencode/opencode.json` | `AGENTS.md` (else `CLAUDE.md`) walking up / `~/.config/opencode/AGENTS.md` (else `~/.claude/CLAUDE.md`) |
| **agy** (Antigravity CLI) 1.3.3 | **Partial / unverified** (root `plugin.json` accepted; `skills/` loads; root `mcp.json` not detected locally; no Agent Plugins mention in docs or changelog) | Yes (docs link agentskills.io) | Yes (native, also GEMINI.md) | `.agents/skills/` (legacy `.agent/skills/`) / `~/.gemini/antigravity-cli/skills/` per docs (see notes) | `.agents/agents/` / `~/.gemini/config/agents/` | `.agents/hooks.json` / `~/.gemini/config/hooks.json` or `~/.gemini/antigravity-cli/settings.json` | Not documented (workflows being migrated to skills) | `.agents/mcp_config.json` / `~/.gemini/config/mcp_config.json` | `AGENTS.md`, `GEMINI.md`, `.agents/AGENTS.md`, `.agents/rules/*.md` / `~/.gemini/{AGENTS,GEMINI}.md`, `~/.gemini/config/{AGENTS,GEMINI}.md`, `~/.gemini/config/rules/`, `~/.gemini/antigravity-cli/rules/` |
| **Kiro** (IDE 1.0.288+, CLI V3) | **Yes, IDE Powers** (IDE 1.0.288, 2026-08-07; rolling out); CLI detects IDE-installed powers | Yes | Yes (native) | `.kiro/skills/` / `~/.kiro/skills/` (IDE only) | `.kiro/agents/*.{json,md}` / `~/.kiro/agents/` | `.kiro/hooks/*.json` / `~/.kiro/hooks/` (CLI V3 only) | `.kiro/prompts/` (CLI; unverified detail) / unverified | `.kiro/settings/mcp.json` / `~/.kiro/settings/mcp.json` | `.kiro/steering/`, `AGENTS.md` (root + subdirs) / `~/.kiro/steering/` (IDE only, incl. AGENTS.md) |
| **Cursor** (docs; cursor-agent 2026.07.09) | **Yes** (plus own `.cursor-plugin/plugin.json` format) | Yes | Yes (native, nested) | `.cursor/skills/`, `.agents/skills/` (+ `.claude/`, `.codex/` compat) / `~/.cursor/skills/`, `~/.agents/skills/` (+ compat) | `.cursor/agents/` (+ `.claude/agents/`, `.codex/agents/`) / `~/.cursor/agents/` (+ compat) | `.cursor/hooks.json` / `~/.cursor/hooks.json` (+ Claude Code hooks compat) | `.cursor/commands/` / `~/.cursor/commands/` (**unverified**, not found in fetched docs) | `.cursor/mcp.json` / `~/.cursor/mcp.json` | `.cursor/rules/`, `AGENTS.md` (root + subdirs) / User Rules in app settings (no file) |

## Per-harness notes

### Claude Code (local 2.1.296)

- Agent Plugins: plugin manifest is "the `plugin.json` file in a plugin's `.claude-plugin/` directory"; manifest optional, standard layout uses `skills/`, `commands/`, `agents/`, `hooks/hooks.json`, `.mcp.json` (dot-prefixed), etc. No mention of agent-plugins.org or root `plugin.json`/`mcp.json`. `$schema` "Claude Code ignores it at load time". https://code.claude.com/docs/en/plugins-reference
- Local probe (`claude plugin validate` on a minimal Agent Plugins 1.0.0 dir with root `plugin.json`, `skills/`, `mcp.json`): "Validation passed" but in "Validating components" mode (no `.claude-plugin/` manifest found). Implication from docs: `skills/` would load via the default layout; root `mcp.json` is not a documented default (`.mcp.json` is). Not a statement of spec support.
- Third-party issue noting lack of support: https://github.com/Azure/git-ape/issues/294 (secondary).
- Agent Skills: skills "follow the Agent Skills open standard". Locations: managed, `~/.claude/skills/`, `.claude/skills/`, nested `<subdir>/.claude/skills/`, plugin `skills/`. `.agents/skills/` is not listed. https://code.claude.com/docs/en/skills
- AGENTS.md: read only when no `CLAUDE.md` / `.claude/CLAUDE.md` / `CLAUDE.local.md` exists in cwd or above; does not read `AGENTS.override.md` or anything under `.agents/`. Changelog 2.1.277: "Added AGENTS.md support: in a project with no CLAUDE.md, Claude Code reads AGENTS.md instead". https://code.claude.com/docs/en/memory ; local `~/.claude/cache/changelog.md`.
- Subagents: managed, `--agents`, `.claude/agents/`, `~/.claude/agents/`, plugin `agents/`. https://code.claude.com/docs/en/sub-agents
- Hooks: `~/.claude/settings.json`, `.claude/settings.json`, `.claude/settings.local.json`, managed, plugin `hooks/hooks.json`, skill/subagent frontmatter. https://code.claude.com/docs/en/hooks
- Commands: "Custom commands have been merged into skills"; `.claude/commands/` still works. https://code.claude.com/docs/en/slash-commands
- MCP: project `.mcp.json`; local and user scope both in `~/.claude.json`. https://code.claude.com/docs/en/mcp

### Codex (local codex-cli 0.162.1)

- Agent Plugins: "A portable plugin has a `plugin.json` manifest at its root" with the agent-plugins.org 1.0.0 schema; `skills/` and root `mcp.json`; Codex-specific settings under `extensions.com.openai`; `.codex-plugin/plugin.json` is a "compatibility fallback"; legacy and Claude-compatible manifests still accepted; hooks discovered at `hooks/hooks.json`; hook env gets `PLUGIN_ROOT`/`PLUGIN_DATA` plus `CLAUDE_PLUGIN_ROOT`/`CLAUDE_PLUGIN_DATA`. https://developers.openai.com/plugins/build/plugins
- Plugin marketplaces: repo `$REPO_ROOT/.agents/plugins/marketplace.json`, personal `~/.agents/plugins/marketplace.json`, legacy `$REPO_ROOT/.claude-plugin/marketplace.json`; install cache `~/.codex/plugins/cache/$MARKETPLACE/$PLUGIN/$VERSION/`. Same source. Local CLI has `codex plugin add|list|marketplace|remove`.
- Launch-day client per https://vercel.com/blog/introducing-agent-plugins.
- Skills: "Skills build on the open agent skills standard"; `$CWD/.agents/skills` up to `$REPO_ROOT/.agents/skills`, `$HOME/.agents/skills`, `/etc/codex/skills`, bundled. https://learn.chatgpt.com/docs/build-skills (redirect from developers.openai.com/codex/skills)
- Subagents: TOML files in `.codex/agents/` and `~/.codex/agents/`. https://learn.chatgpt.com/docs/agent-configuration/subagents
- Hooks: `~/.codex/hooks.json`, `~/.codex/config.toml`, `<repo>/.codex/hooks.json`, `<repo>/.codex/config.toml` (project only when trusted), plus enabled plugins. https://learn.chatgpt.com/docs/hooks
- Commands: custom prompts in `~/.codex/prompts/` are deprecated in favour of skills. https://learn.chatgpt.com/docs/custom-prompts
- MCP: `~/.codex/config.toml`; project `.codex/config.toml` for trusted projects. https://learn.chatgpt.com/docs/extend/mcp?surface=cli
- AGENTS.md: `~/.codex/AGENTS.override.md` / `AGENTS.md`; project root down to cwd, one file per dir (`AGENTS.override.md`, `AGENTS.md`, `project_doc_fallback_filenames`), 32 KiB cap. https://learn.chatgpt.com/docs/agent-configuration/agents-md

### Pi (local 1.1.0; latest release v1.1.0, 2026-10-07; repo now `earendil-works/pi`)

- Agent Plugins: no mention. Distribution unit is a Pi package (npm/git/local) with conventional `extensions/`, `skills/`, `prompts/`, `themes/` dirs or a `pi` key in `package.json`; declarations in `~/.pi/agent/settings.json` or `.pi/settings.json` (`pi install -l`). https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/packages.md
- Agent Skills: "Pi implements the Agent Skills specification"; also supports `~/.agents/skills/` and `.agents/skills/` (cwd ancestors up to repo root). https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/skills.md
- Config layout (agent dir `~/.pi/agent`, override `PI_CODING_AGENT_DIR`; project `.pi/`, loaded after project trust): `settings.json`, `mcp.json`, `extensions/`, `skills/`, `prompts/` (slash commands), `themes/`, `SYSTEM.md`, `APPEND_SYSTEM.md`; user context `AGENTS.override.md`, `AGENTS.md`, `AGENTS.MD`, `CLAUDE.md`, `CLAUDE.MD`. Context files load from agent dir, cwd and parents, without trust. https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/configuration.md
- MCP: built in (`builtin:mcp` extension, `pi mcp` CLI), config `mcp.json`. Same doc + settings.md.
- Hooks: no hooks file; lifecycle interception is via extension `pi.on()` events (e.g. `tool_call` can block). https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/extensions.md
- Subagents: README says Pi "skips features like sub-agents and plan mode" (available via packages). https://github.com/earendil-works/pi/blob/main/README.md

### OpenCode (local 1.17.9)

- Agent Plugins: plugins page describes JS/TS modules in `.opencode/plugins/`, `~/.config/opencode/plugins/` or npm via `plugin` in `opencode.json`; no `plugin.json`/agent-plugins.org mention. No upstream issue found. https://opencode.ai/docs/plugins/
- Agent Skills: listed on agentskills.io showcase. Paths: `.opencode/skills/`, `.claude/skills/`, `.agents/skills/` (cwd up to git worktree) and `~/.config/opencode/skills/`, `~/.claude/skills/`, `~/.agents/skills/`. https://opencode.ai/docs/skills/
- Context: `AGENTS.md` (fallback `CLAUDE.md`) walking up; global `~/.config/opencode/AGENTS.md` (fallback `~/.claude/CLAUDE.md`); `instructions` config; `OPENCODE_DISABLE_CLAUDE_CODE*` toggles. https://opencode.ai/docs/rules/
- Agents: `.opencode/agents/`, `~/.config/opencode/agents/`, or `agent` in config. https://opencode.ai/docs/agents/
- Commands: `.opencode/commands/`, `~/.config/opencode/commands/`. https://opencode.ai/docs/commands/
- MCP: `mcp` key in config; config precedence remote `.well-known/opencode` → `~/.config/opencode/opencode.json` → `OPENCODE_CONFIG` → project `opencode.json` → `.opencode` dirs → `OPENCODE_CONFIG_CONTENT` → managed. https://opencode.ai/docs/config/
- Hooks: hook functions returned by plugins (e.g. `tool.execute.before`). https://opencode.ai/docs/plugins/

### agy / Antigravity CLI (local 1.3.3)

- Plugins: own system. "Every plugin requires a `plugin.json` file at its root"; components `skills/`, `agents/`, `rules/`, `mcp_config.json`, `hooks.json`; install to workspace `.agents/plugins/`, global `~/.gemini/config/plugins/`, CLI staging `~/.gemini/antigravity-cli/plugins/<name>/`; marketplace `antigravity-plugins-official`. Docs do not mention Agent Plugins / agent-plugins.org. https://antigravity.google/docs/plugins/
- Changelog (`agy changelog`, 1.3.2/1.3.3): manifest may live in `.antigravity-plugin/` or `.gemini-plugin/`; skills may be listed in `plugin.json`; `PLUGIN_DATA`; `bin/`; root `agents.json`, `skills.json`, `rules.json` (1.3.3); `agy plugin import` from "gemini or claude". No "Agent Plugins" string in the changelog.
- Local probe: `agy plugin validate` on a minimal Agent Plugins 1.0.0 dir → `[ok]`, `skills: 1 processed`, `mcpServers: skipped (not found)` — i.e. root `plugin.json` + `skills/` are accepted, the spec's root `mcp.json` was not detected (agy uses `mcp_config.json`). Hence "partial / unverified".
- Agent Skills: "Skills are an open standard" (links agentskills.io). Workspace `.agents/skills/` (legacy `.agent/skills/`); global per surface: Antigravity 2.0 and IDE `~/.gemini/config/skills/`, **CLI `~/.gemini/antigravity-cli/skills/`**, IDE legacy `~/.gemini/antigravity/skills/`; CLI plugin skills `~/.gemini/antigravity-cli/plugins/<name>/skills/`. https://antigravity.google/docs/skills
- Rules/context: dir-scoped `AGENTS.md`, `GEMINI.md`, `.agents/AGENTS.md`, `.agents/GEMINI.md`, `.agents/rules/*.md` (legacy `.agent/rules/`); global `~/.gemini/AGENTS.md`, `~/.gemini/GEMINI.md`, `~/.gemini/config/{AGENTS,GEMINI}.md`, `~/.gemini/config/rules/*.md`, CLI also `~/.gemini/antigravity-cli/rules/*.md`. https://antigravity.google/docs/rules
- Subagents: `.agents/agents/<name>.md` or `.agents/agents/<name>/agent.md`; global `~/.gemini/config/agents/`. https://antigravity.google/docs/subagents
- Hooks: `.agents/hooks.json`; global `~/.gemini/config/hooks.json` or inside `~/.gemini/antigravity-cli/settings.json`; plugin `hooks.json`. https://antigravity.google/docs/hooks
- MCP: CLI global `~/.gemini/config/mcp_config.json`, workspace `.agents/mcp_config.json`. https://antigravity.google/docs/mcp
- Commands: no file location documented for custom slash commands; docs point to a "Workflows to skills" migration. https://antigravity.google/docs/slash-commands
- Also detects Gemini CLI extensions in `~/.gemini/extensions` (local log line `gemini_extensions.go: Detecting Gemini extensions in ~/.gemini/extensions`).

#### Which `~/.gemini/antigravity*` directory is canonical for agy 1.3.3

Evidence from this machine (read-only, 2026-10-12):

| Directory | Contents | Last modified | Status |
|---|---|---|---|
| `~/.gemini/antigravity-cli/` | `settings.json`, `mcp_config.json`, `conversations/`, `brain/`, `builtin/skills/`, `log/`, `history.jsonl`, `updater/` | 2026-10-12 | **Canonical CLI app-data dir.** The agy log states `CLI app data directory: ~/.gemini/antigravity-cli` and `appDataDir=~/.gemini/antigravity-cli`. |
| `~/.gemini/config/` (not `antigravity*` but relevant) | `config.json`, `hooks.json`, `mcp_config.json`, `projects/` | 2026-10-12 | **Canonical shared user customization dir** (hooks, MCP, rules, agents, plugins, `skills.json`/`rules.json`/`plugins.json`/`agents.json`). Log: `applyUserSettings ... ~/.gemini/config/config.json`; changelog 1.2.15 names `~/.gemini/config` for "global rules and customization files". |
| `~/.gemini/antigravity/` | `antigravity_state.pbtxt`, `agyhub_summaries_proto.pb`, `installation_id`, `knowledge/` | 2026-06-04 | Not used by CLI 1.3.3; docs name `~/.gemini/antigravity/skills/` only as the IDE's legacy global skills path. |
| `~/.gemini/antigravity-ide/` | same shape as above | 2026-06-04 | Not used by the CLI (IDE-era state; purpose unverified). |
| `~/.gemini/antigravity-backup/` | same shape as above | 2026-06-04 | Not used by the CLI (appears to be a backup copy; unverified). |

Note a docs inconsistency: the Skills page gives the CLI global skills path as `~/.gemini/antigravity-cli/skills/`, while Rules/Subagents/Hooks/MCP/Plugins pages and the changelog put global customization under `~/.gemini/config/`. Which global skills path(s) the 1.3.3 CLI scans is **unverified** (no skills directory exists in either location on this machine).

### Kiro (IDE 1.0.288+, CLI V3; not installed locally)

- Agent Plugins: IDE 1.0.288 changelog (2026-08-07): "Install powers aligned with the open Agent Plugin format from a local folder or GitHub URL." Blog: "Kiro powers is rolling out support for Agent Plugins 1.0.0" (skills under `skills/`, tools in `mcp.json`; legacy `POWER.md` powers keep working). Kiro CLI V3: "Powers installed via the IDE are automatically detected by CLI sessions." Whether the CLI installs Agent Plugins on its own is **unverified**. https://kiro.dev/changelog/ide/1-0-288/ ; https://kiro.dev/blog/powers-supports-plugins/ ; https://kiro.dev/docs/cli/v3/new-features/
- Agent Skills: listed on agentskills.io. `.kiro/skills/` (workspace), `~/.kiro/skills/` (global, IDE only per table); `.agents/skills/` not mentioned. https://kiro.dev/docs/skills/
- Context: `.kiro/steering/` (IDE+CLI), `~/.kiro/steering/` (IDE only); `AGENTS.md` from workspace root, subdirectories, and `~/.kiro/steering/`, always included. https://kiro.dev/docs/steering/
- Custom agents: `.kiro/agents/[name].json|.md` (trusted workspaces) and `~/.kiro/agents/`; usable as subagents. https://kiro.dev/docs/custom-agents/
- Hooks: `.kiro/hooks/*.json`; global `~/.kiro/hooks/` in CLI V3 only. https://kiro.dev/docs/hooks/
- MCP: `.kiro/settings/mcp.json`, `~/.kiro/settings/mcp.json` (merged, workspace wins). https://kiro.dev/docs/mcp/configuration/
- Commands: CLI has `/prompts`; help-agent doc mentions "Prompts in .kiro/prompts/"; skills can be invoked as slash commands (CLI 2.1, secondary source). Global prompt path **unverified**. https://kiro.dev/docs/cli/chat/help-agent/

### Cursor (docs; local cursor-agent 2026.07.09)

- Agent Plugins: "A plugin that follows the Agent Plugins specification loads in Cursor" (skills + MCP). Cursor Plugins (`.cursor-plugin/plugin.json`) add rules, agents, commands, hooks and "continue to develop in parallel with the standard". Local dev plugins in `~/.cursor/plugins/local`; marketplace installs at project or user scope. Local `cursor-agent --help` shows `--plugin-dir <path>`. https://cursor.com/docs/plugins
- Agent Skills: `.agents/skills/`, `.cursor/skills/`; `~/.agents/skills/`, `~/.cursor/skills/`; compat `.claude/skills/`, `.codex/skills/`, `~/.claude/skills/`, `~/.codex/skills/`. https://cursor.com/docs/context/skills
- Rules/context: `.cursor/rules`; `AGENTS.md` in root and any subdirectory; User Rules live in app settings (Customize → Rules), not a file. CLAUDE.md not mentioned. https://cursor.com/docs/context/rules
- Subagents: `.cursor/agents/`, `.claude/agents/`, `.codex/agents/`; `~/.cursor/agents/`, `~/.claude/agents/`, `~/.codex/agents/`. https://cursor.com/docs/subagents
- Hooks: enterprise `/Library/Application Support/Cursor/hooks.json` (macOS), team (dashboard), project `.cursor/hooks.json`, user `~/.cursor/hooks.json`; loads Claude Code hooks for compatibility. https://cursor.com/docs/hooks
- MCP: `.cursor/mcp.json`, `~/.cursor/mcp.json`. https://cursor.com/docs/context/mcp
- Commands: file locations not found in fetched docs (`/migrate-to-skills` converts user- and workspace-level commands to skills). `.cursor/commands/` / `~/.cursor/commands/` is **unverified**.

## Sources not used / caveats

- WebFetch summaries were used for long vendor pages; exact quotes above are as returned. Pages with redirects (developers.openai.com → learn.chatgpt.com) were followed.
- Kiro and Cursor IDE were not installed locally; their facts are docs-only.
- "Version checked" for Claude Code, Codex, Pi, OpenCode, agy, cursor-agent is from local `--version` on 2026-10-12.
