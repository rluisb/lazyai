# LazyAI + vibe-lab — Official Tool Compliance Matrix

Verification date: 2026-06-22.

This matrix turns official tool docs into adapter requirements. Each adapter must remain native to the host tool and must fail validation when LazyAI would emit unsupported or deprecated configuration.

---

## 1. Summary matrix

| Tool | Support level | Official concepts to support | Required LazyAI canonical assets | Native outputs | Current risk |
|---|---|---|---|---|---|
| OpenCode | stable | `AGENTS.md`, agents, subagents, permissions, MCP, skills, commands, plugins, chat modes | agents, skills, rules, hooks, MCP, commands, chat modes | `AGENTS.md`, `opencode.json`, `.opencode/*` | Low: permissions and steps are current; post-install validation via `opencode debug` |
| Claude Code | stable | `CLAUDE.md`, `.claude`, skills, subagents, hooks, MCP, permissions, plugins, managed settings, commands, output styles | agents, skills, rules, hooks, MCP, commands, output styles | `CLAUDE.md`, `.claude/*`, `.mcp.json` | Low: `CLAUDE.md` generated; DriveCLI and LocalSecrets paths tested |
| GitHub Copilot | stable | repo instructions, path instructions, custom agents, skills, hooks, MCP, plugins, prompt templates, chat modes | rules, agents, skills, hooks, MCP, prompts, chat modes | `.github/*`, `.vscode/mcp.json`, `~/.copilot/mcp-config.json` | Medium: many surfaces differ between IDE, CLI, cloud; global scope probe-gated |
| Pi | stable | settings, project trust, skills, prompts, TypeScript extensions, compaction, packages, subagents | skills, prompts, hooks, rules, agents | `.pi/*`, `.agents/skills`, `AGENTS.md` | Medium: project trust and no sandbox must be explicit; MCP is a no-op |
| OMP | stable | AGENTS/context, plugins, skills, commands, hooks, MCP, compaction, handoff, prompts | agents, skills, hooks, MCP, commands, handoff, prompts | `.omp/*`, `AGENTS.md` | Low: all emitted surfaces verified against authoritative OMP docs (see `docs/adapters/snapshots/beta-adapter-verification-2026-06.md`); plugin bundle and handoff/compaction assets not yet emitted |
| Antigravity/Gemini | stable | `.agents/skills`, hooks, MCP, plugins, permissions, sandbox/settings | skills, hooks, MCP, root instructions | `.gemini/*`, `.agents/skills/*`, `.agents/hooks.json` | Low: all emitted surfaces verified against official Antigravity/Gemini docs; both 2026-06 gaps closed + pinned |
| Kiro | stable | agents, skills, prompts, MCP, permissions | agents, skills, prompts, MCP | `.kiro/agents/*`, `.kiro/skills/*`, `.kiro/prompts/*`, `.kiro/settings/mcp.json` | Low/medium: steering, specs, runtime hooks, and `.kiroignore` not yet emitted (requires external docs refresh) |

---

## 2. OpenCode compliance checklist

Officially required/available surfaces:

- root instructions: `AGENTS.md`;
- agent configuration via `opencode.json` or Markdown files;
- primary agents and subagents;
- Plan agent / restricted planning mode;
- permissions field for edit/bash/read/grep/etc.;
- MCP server config;
- skills in `.opencode/skills/<name>/SKILL.md`, plus Claude/Agent-compatible locations;
- commands in `.opencode/commands/*.md`;
- plugins in `.opencode/plugins/` or global plugin config;
- chat modes in `.opencode/modes/*.md`.

LazyAI requirements:

- [x] Generate `AGENTS.md` (via scaffold/root.go).
- [x] Generate `.opencode/agents/*.md` with frontmatter rewrite (description only baseline).
- [x] Use `permission`, not deprecated `tools`, by default in `opencode.json`.
- [x] Use `steps`, not deprecated `maxSteps`, in agent frontmatter.
- [x] Generate `.opencode/skills/<name>/SKILL.md`.
- [x] Validate skill name regex and description length (via `validate` package).
- [x] Generate `.opencode/commands/*.md` from `library/opencode/commands/`.
- [x] Generate MCP config (merged into `opencode.json` via `compileOpenCodeMCP`).
- [x] Generate plugin hooks when hook policies require event interception (`.opencode/plugins/vibe-lab-hooks.js`).
- [x] Generate `.opencode/modes/*.md` from `library/opencode/modes/`.
- [x] Post-install validation via `opencode debug config` and `opencode debug agent`.
- [x] Support headless init (`opencode init`).
- [x] Support global scope.

Conformance tests:

```text
opencode_starter_generates_agents
opencode_skill_name_validation
opencode_no_maxSteps
opencode_permissions_not_tools
opencode_mcp_shape
opencode_commands_frontmatter
opencode_plugin_exists_when_hooks_enabled
opencode_modes_exist
opencode_post_install_validation
```

---

## 3. Claude Code compliance checklist

Officially required/available surfaces:

- `CLAUDE.md` root instructions;
- `.claude` project directory;
- user and project skills;
- subagents;
- hooks;
- MCP servers;
- settings and managed settings;
- permissions;
- plugins that bundle skills, agents, hooks, MCP, LSP, monitors;
- commands (slash commands);
- output styles.

LazyAI requirements:

- [x] Generate `CLAUDE.md` for Claude target (via `ensureClaudeContextDoc`; imports `@AGENTS.md`).
- [x] Do not treat `AGENTS.md` as enough for Claude Code (FR-012: `CLAUDE.md` is the native root).
- [x] Generate `.claude/skills/<name>/SKILL.md`.
- [x] Generate `.claude/agents/<name>.md` with frontmatter rewrite (name + description only).
- [x] Generate `.claude/hooks/*` (block-destructive-shell.sh, objective-workflow-gate.sh) and settings hook config.
- [x] Generate `.mcp.json` for project MCP (or `.claude/settings.local.json` when `LocalSecrets` is set).
- [x] Generate permission settings in `.claude/settings.json` (allow/deny lists).
- [x] Warn about managed settings that block project/user customizations (via `displayInstallSummary`).
- [x] Support plugin bundle generation.
- [x] Generate `.claude/commands/*.md` (starter set).
- [x] Generate `.claude/output-styles/*.md` (starter set).
- [x] Support DriveCLI mode (`claude mcp add-json` with fallback to direct-write).
- [x] Support LocalSecrets mode (routes MCP to gitignored `.claude/settings.local.json`).
- [x] Support global scope (writes to `~/.claude/`; skips `CLAUDE.md` at global).
- [x] Legacy agent migration: flat `.claude/<agent>.md` → `.claude/agents/<agent>.md`.
- [x] Post-install summary of registered tools.

Conformance tests:

```text
claude_generates_claude_md
claude_agents_written_to_claude_agents
claude_skills_written_to_claude_skills
claude_hooks_schema_valid
claude_mcp_json_valid
claude_managed_settings_warning
claude_commands_exist
claude_output_styles_exist
claude_drive_cli_fallback
claude_local_secrets_routing
claude_legacy_agent_migration
```

---

## 4. GitHub Copilot compliance checklist

Officially required/available surfaces:

- `.github/copilot-instructions.md` repository instructions;
- `.github/instructions/*.instructions.md` path-specific instructions;
- AGENTS.md and other agent instruction files where supported;
- custom agents;
- agent skills in project/personal locations;
- Copilot CLI hooks;
- MCP servers across Copilot surfaces (VS Code `.vscode/mcp.json`, CLI `~/.copilot/mcp-config.json`);
- Copilot CLI plugins including agents, skills, hooks, MCP configs;
- plugin precedence and dedup behavior;
- prompt templates (`.github/prompts/*.prompt.md`);
- chat modes (`.github/chatmodes/*.chatmode.md`);
- skill tier support (Frontier, Speed, Balanced).

LazyAI requirements:

- [x] Generate `.github/copilot-instructions.md` (via scaffold/root.go).
- [x] Generate path-scoped `.github/instructions/*.instructions.md` from library.
- [x] Generate `.github/agents/*.agent.md` from canonical agents (with model tier resolution).
- [x] Generate `.github/skills/<name>/SKILL.md` (Agent Skills directories; global uses `.copilot/skills`).
- [x] Generate `.vscode/mcp.json` (project/workspace scope).
- [x] Generate `~/.copilot/mcp-config.json` (CLI surface, all scopes with probe).
- [x] Generate CLI hooks only for project/workspace scope (`.github/hooks/*.json` + `*.sh`).
- [x] Generate plugin bundle optionally.
- [x] Validate duplicate agent IDs and skill names.
- [x] Generate `.github/prompts/*.prompt.md` (prompt templates with `.prompt.md` extension).
- [x] Generate `.github/chatmodes/*.chatmode.md` (chat modes).
- [x] Support skill tier annotations (Frontier/Speed/Balanced → model resolution via `CopilotCatalog`).
- [x] Support global scope (probe-gated: requires `copilot` binary or `~/.copilot/` directory).
- [x] Legacy skill-agent cleanup (removes old skill-as-agent `.agent.md` files on re-run).

Conformance tests:

```text
copilot_generates_repo_instructions
copilot_path_instructions_scope
copilot_skills_project_location
copilot_mcp_vscode_shape
copilot_plugin_manifest_valid
copilot_precedence_warning
copilot_prompts_exist
copilot_chatmodes_exist
copilot_skill_tier_resolution
copilot_global_scope_probe
copilot_legacy_cleanup
```

---

## 5. Pi compliance checklist

Officially required/available surfaces:

- minimal core extended by TypeScript extensions, skills, prompt templates, themes, packages;
- `.pi/settings.json` global/project settings;
- `.pi/skills`, `.agents/skills`;
- `.pi/prompts`;
- `.pi/extensions`;
- `.pi/agents` (subagent definitions);
- `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md`;
- project trust for project-local resources;
- no built-in sandbox;
- compaction and JSONL session format.

LazyAI requirements:

- [x] Generate `.pi/skills/<name>/SKILL.md`.
- [x] Generate `.pi/prompts/*.md` for prompt templates.
- [x] Generate `.pi/extensions/*.ts` (safety hooks as TypeScript extensions; Pi has no `.pi/hooks` path).
- [x] Generate `.pi/agents/*.md` (subagent definitions from canonical agents).
- [ ] Generate `.pi/settings.json` without secrets. (Not currently emitted; Pi has no settings.json merge in the adapter.)
- [x] Warn that Pi project trust is not a sandbox.
- [x] Warn that Pi runs with local user permissions.
- [x] Support `.agents/skills` compatibility output.
- [ ] MCP compile is a no-op (Pi has no native MCP surface; `CompileMCP` returns empty).
- [ ] No global scope support (project/workspace only).

Conformance tests:

```text
pi_skills_frontmatter
pi_prompt_template_frontmatter
pi_project_trust_warning
pi_no_inline_secret_settings
pi_extension_only_when_needed
pi_agents_exist
pi_mcp_noop
```

---

## 6. OMP compliance checklist

Officially required/available surfaces from docs/search extracts:

- terminal-first coding agent;
- AGENTS/context files;
- subagents, plan mode, LSP/DAP;
- compaction and handoff;
- hindsight memory;
- plugins bundling skills, commands, hooks, custom tools, MCP servers, themes;
- `.omp/mcp.json` project MCP and user MCP config;
- `.omp/prompts/*.md` prompt templates;
- marketplace/plugin install concepts;
- compatibility with Claude Code plugin registry concepts.

LazyAI requirements:

- [x] Generate `AGENTS.md` (via scaffold/root.go).
- [x] Generate `.omp/mcp.json` (via `compileOmpMCP`; reuses Claude Code MCP format).
- [x] Generate `.omp/skills/*/SKILL.md`.
- [x] Generate `.omp/commands/*` (canonical slash commands).
- [x] Generate `.omp/prompts/*.md` (prompt templates; discovery docs-confirmed via `omp://config-usage.md` §6).
- [x] Generate `.omp/hooks/pre/*.ts` (TypeScript hook factories for OMP).
- [ ] Generate plugin bundle optionally. (Not yet implemented.)
- [ ] Compile handoff/compaction assets. (Not yet implemented.)
- [x] Mark adapter as stable (all emitted surfaces verified against authoritative OMP docs; see `docs/adapters/snapshots/beta-adapter-verification-2026-06.md`).
- [ ] No global scope support (project/workspace only).
- [ ] No headless support.

Conformance tests:

```text
omp_generates_mcp_json
omp_generates_skills
omp_generates_commands
omp_plugin_bundle_shape
omp_handoff_skill_exists
omp_prompts_exist
omp_hooks_exist
```

---

## 7. Antigravity / Gemini compliance checklist

Officially required/available surfaces from docs/search extracts:

- CLI features and plugins;
- plugins as namespaced bundles containing skills, agents, rules, MCP servers, hooks;
- skills under `.agents/skills`, backward `.agent/skills` support;
- rules under `.agents/rules`;
- settings under `~/.gemini/antigravity-cli/settings.json`;
- permissions in settings;
- hooks JSON mapping hook names to event configs;
- sandbox/security settings.

LazyAI requirements:

- [x] Generate `.agents/skills/<name>/SKILL.md`.
- [ ] Generate `.agents/rules/*.md`. (Not currently emitted; no rules directory created.)
- [ ] Generate `AGENTS.md` and optionally `GEMINI.md`. (AGENTS.md is handled by scaffold/root.go, not the adapter; GEMINI.md is not emitted.)
- [x] Generate hooks config (`.agents/hooks.json` with hook event mappings).
- [x] Generate `.gemini/hooks/lazyai/*.sh` (hook scripts).
- [x] Generate `.gemini/settings.json` (merged with defaults).
- [x] Generate MCP config (`~/.gemini/config/mcp_config.json` via `compileAntigravityMCP`).
- [x] Do not write user-global settings without explicit consent (merge preserves user keys).
- [x] Warn about permissions/sandbox assumptions.
- [x] Mark adapter as stable (all emitted surfaces verified against official Antigravity/Gemini docs; both 2026-06 gaps closed + pinned).
- [ ] No global scope support (project/workspace only).
- [ ] No headless support.

Conformance tests:

```text
antigravity_generates_agents_skills
antigravity_generates_agents_rules
antigravity_no_global_settings_by_default
antigravity_hook_schema_guarded
antigravity_sandbox_warning
antigravity_mcp_config_shape
```

---

## 8. Kiro compliance checklist

Officially required/available surfaces:

- specs: requirements/bug analysis, design, tasks;
- steering files in `.kiro/steering` and global `~/.kiro/steering`;
- steering inclusion modes: always, fileMatch, manual, auto;
- hooks for file events, prompt submission, agent turn completion, tool invocation, spec task execution, manual triggers;
- MCP config;
- Autopilot vs Supervised mode;
- trusted commands;
- protected paths and `.kiroignore`.

LazyAI requirements:

- [x] Generate `.kiro/agents/*.md` (custom agent profiles; Kiro CLI v3 discovers from `.kiro/agents/`).
- [x] Generate `.kiro/skills/<name>/SKILL.md`.
- [x] Generate `.kiro/prompts/*.md`.
- [x] Generate `.kiro/settings/mcp.json` (via `compileKiroMCP`; reuses Claude Code MCP format).
- [ ] Generate `.kiro/steering/*.md`. (Not yet implemented — requires external docs refresh.)
- [ ] Generate steering frontmatter for inclusion modes. (Not yet implemented — requires external docs refresh.)
- [ ] Generate `.kiro/specs/<name>/requirements.md`, `design.md`, `tasks.md` when specs enabled. (Not yet implemented — requires external docs refresh.)
- [x] Generate `.kiro/hooks/*.json` from canonical hooks with source-verified Kiro v3 triggers only. Repo-local permissions are forbidden; Powers is a future importable-package direction.
- [ ] Generate `.kiroignore` for sensitive exclusions. (Not yet implemented — requires external docs refresh.)
- [x] Warn that Supervised mode is not a sandbox.
- [x] Warn about broad trusted command wildcards.
- [ ] No global scope support (project/workspace only).
- [ ] No headless support.

Conformance tests:

```text
kiro_steering_foundational_files
kiro_steering_inclusion_modes
kiro_specs_three_phase
kiro_mcp_settings_shape
kiro_security_warnings
kiro_agents_exist
kiro_skills_exist
kiro_prompts_exist
```

---

## 9. Cross-tool compliance requirements

### 9.1 Skill compatibility

Canonical skills must be emitted to each target's official or compatible skill locations. Validation must ensure:

- `SKILL.md` exact uppercase filename;
- YAML frontmatter;
- required `name` and `description`;
- lowercase kebab-case name;
- directory-name consistency in strict mode;
- trigger/non-trigger guidance.

### 9.2 Root instruction compatibility

LazyAI must not assume one root instruction file works everywhere.

Required mapping:

| Canonical | Target output |
|---|---|
| root rules | OpenCode `AGENTS.md` |
| root rules | Claude `CLAUDE.md` (generated with `@AGENTS.md` import) plus `AGENTS.md` |
| root rules | Copilot `.github/copilot-instructions.md` |
| scoped rules | Copilot `.github/instructions/*.instructions.md` |
| root/scoped rules | Kiro `.kiro/steering/*.md` (not yet emitted — requires external docs refresh) |
| root/scoped rules | Antigravity `AGENTS.md` (via scaffold), `.agents/rules/*` (not yet emitted) |
| root/rules | Pi `AGENTS.md`, `.pi/SYSTEM.md`/`APPEND_SYSTEM.md` optional |
| root/rules | OMP `AGENTS.md`, `.omp` surfaces |

### 9.3 MCP compatibility

MCP config must be target-specific and must preserve:

- command/args;
- local vs remote (URL-based);
- headers/env references;
- OAuth config where supported;
- timeouts where supported;
- enabled/disabled flags;
- tool allow/deny mapping where supported.

Per-target MCP output:

| Tool | MCP output path | Format | Notes |
|---|---|---|---|
| OpenCode | `opencode.json` (merged into `mcp` key) | OpenCode native | Legacy `.opencode/lazyai.mcp.jsonc` also read for migration |
| Claude Code | `.mcp.json` (project) or `.claude/settings.local.json` (LocalSecrets) | `mcpServers` | DriveCLI path uses `claude mcp add-json` |
| Copilot | `.vscode/mcp.json` (IDE) + `~/.copilot/mcp-config.json` (CLI) | `servers` / `mcpServers` | CLI probe-gated; VS Code includes `inputs` for env placeholders |
| Pi | (none) | — | MCP compile is a no-op |
| OMP | `.omp/mcp.json` | `mcpServers` (Claude Code format) | — |
| Kiro | `.kiro/settings/mcp.json` | `mcpServers` (Claude Code format) | — |
| Antigravity | `~/.gemini/config/mcp_config.json` | `mcpServers` (Antigravity native) | Merged via configmerge |

### 9.4 Security compatibility

For every target, `doctor` must report:

- whether generated project-local files are trust-sensitive;
- whether the host tool has sandboxing or only review/trust controls;
- generated hook risk;
- generated MCP risk;
- generated permission posture;
- global config writes.

---

## 10. Adapter release gate

An adapter can be marked stable only when:

- official docs registry entry exists;
- docs snapshot exists;
- golden tests exist;
- smoke tests exist or documented as impossible without account;
- all required outputs validate;
- security/trust warnings are implemented;
- migration/eject behavior is tested;
- unsupported features produce warnings, not silent omissions.

### 10.1 Beta graduation criteria

Adapters currently at `beta` must satisfy these additional criteria before promotion to `stable`:

- Official docs snapshots are fully captured (no partially JS-rendered gaps).
- All required outputs from the compliance checklist are implemented (no `[ ]` items).
- Golden tests exist for every emitted output surface.
- Smoke tests exist or are documented as impossible without an account.
- Security/trust warnings are implemented and tested.
- Migration/eject behavior is tested.
- Unsupported features produce warnings, not silent omissions.

### 10.2 Current adapter status

| Adapter | Status | Since | Blocking items |
|---|---|---|---|
| OpenCode | stable | — | None |
| Claude Code | stable | — | None |
| GitHub Copilot | stable | — | None |
| Pi | stable | — | None |
| OMP | stable | 2026-06-23 | Plugin bundle and handoff/compaction assets not yet emitted (non-blocking; see compliance checklist §6) |
| Antigravity | stable | 2026-06-23 | None (both 2026-06 gaps closed + pinned) |
| Kiro | stable | — | Steering, specs, hooks, `.kiroignore` not yet emitted (requires external docs refresh) |