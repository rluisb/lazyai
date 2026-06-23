# LazyAI + vibe-lab — Official Tool Compliance Matrix

Verification date: 2026-06-21.

This matrix turns official tool docs into adapter requirements. Each adapter must remain native to the host tool and must fail validation when LazyAI would emit unsupported or deprecated configuration.

---

## 1. Summary matrix

| Tool | Official concepts to support | Required LazyAI canonical assets | Native outputs | Current risk |
|---|---|---|---|---|
| OpenCode | `AGENTS.md`, agents, subagents, permissions, MCP, skills, commands, plugins | agents, skills, rules, hooks, MCP, commands | `AGENTS.md`, `opencode.json`, `.opencode/*` | Low if `steps` and permissions are current |
| Claude Code | `CLAUDE.md`, `.claude`, skills, subagents, hooks, MCP, permissions, plugins, managed settings | agents, skills, rules, hooks, MCP, commands | `CLAUDE.md`, `.claude/*`, `.mcp.json` | Medium: current LazyAI state must add `CLAUDE.md` |
| GitHub Copilot | repo instructions, path instructions, custom agents, skills, hooks, MCP, plugins | rules, agents, skills, hooks, MCP, prompts | `.github/*`, `.vscode/mcp.json`, optional plugin | Medium: many surfaces differ between IDE, CLI, cloud |
| Pi | settings, project trust, skills, prompts, TypeScript extensions, compaction, packages | skills, prompts, hooks, rules, MCP where supported | `.pi/*`, `.agents/skills`, `AGENTS.md` | Medium: project trust and no sandbox must be explicit |
| OMP | AGENTS/context, plugins, skills, commands, hooks, MCP, compaction, handoff | agents, skills, hooks, MCP, commands, handoff | `.omp/*`, `AGENTS.md` | Medium/high: docs content partially JS-rendered |
| Antigravity/Gemini | `.agents/skills`, `.agents/rules`, hooks, MCP, plugins, permissions, sandbox/settings | skills, rules, hooks, MCP, root instructions | `.agents/*`, `AGENTS.md`, `GEMINI.md`, settings guidance | Medium/high: docs content partially JS-rendered |
| Kiro | specs, steering, hooks, MCP, supervised/autopilot, trusted commands, protected paths | agents, skills, prompts, MCP | `.kiro/agents/*`, `.kiro/skills/*`, `.kiro/prompts/*`, `.kiro/settings/mcp.json` | Low/medium: steering, specs, runtime hooks, and `.kiroignore` not yet emitted (requires external docs refresh) |

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
- plugins in `.opencode/plugins/` or global plugin config.

LazyAI requirements:

- [ ] Generate `AGENTS.md`.
- [ ] Generate `.opencode/agents/*.md` or `opencode.json` entries.
- [ ] Use `permission`, not deprecated `tools`, by default.
- [ ] Use `steps`, not deprecated `maxSteps`.
- [ ] Generate `.opencode/skills/<name>/SKILL.md`.
- [ ] Validate skill name regex and description length.
- [ ] Generate `.opencode/commands/*.md`.
- [ ] Generate MCP config.
- [ ] Generate plugin hooks when hook policies require event interception.

Conformance tests:

```text
opencode_starter_generates_agents
opencode_skill_name_validation
opencode_no_maxSteps
opencode_permissions_not_tools
opencode_mcp_shape
opencode_commands_frontmatter
opencode_plugin_exists_when_hooks_enabled
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
- plugins that bundle skills, agents, hooks, MCP, LSP, monitors.

LazyAI requirements:

- [ ] Generate `CLAUDE.md` for Claude target.
- [ ] Do not treat `AGENTS.md` as enough for Claude Code.
- [ ] Generate `.claude/skills/<name>/SKILL.md`.
- [ ] Generate `.claude/agents/<name>.md` when subagents enabled.
- [ ] Generate `.claude/hooks/*` and settings hook config when hooks enabled.
- [ ] Generate `.mcp.json` for project MCP.
- [ ] Generate permission settings where safe.
- [ ] Warn about managed settings that block project/user customizations.
- [ ] Support plugin bundle generation.

Conformance tests:

```text
claude_generates_claude_md
claude_agents_written_to_claude_agents
claude_skills_written_to_claude_skills
claude_hooks_schema_valid
claude_mcp_json_valid
claude_managed_settings_warning
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
- MCP servers across Copilot surfaces;
- Copilot CLI plugins including agents, skills, hooks, MCP configs;
- plugin precedence and dedup behavior.

LazyAI requirements:

- [ ] Generate `.github/copilot-instructions.md`.
- [ ] Generate path-scoped `.github/instructions/*.instructions.md`.
- [ ] Generate `.github/agents/*.agent.md` where enabled.
- [ ] Generate `.github/skills/<name>/SKILL.md` or `.agents/skills` compatibility output.
- [ ] Generate `.vscode/mcp.json`.
- [ ] Generate CLI hooks only for Copilot CLI target profile.
- [ ] Generate plugin bundle optionally.
- [ ] Validate duplicate agent IDs and skill names.

Conformance tests:

```text
copilot_generates_repo_instructions
copilot_path_instructions_scope
copilot_skills_project_location
copilot_mcp_vscode_shape
copilot_plugin_manifest_valid
copilot_precedence_warning
```

---

## 5. Pi compliance checklist

Officially required/available surfaces:

- minimal core extended by TypeScript extensions, skills, prompt templates, themes, packages;
- `.pi/settings.json` global/project settings;
- `.pi/skills`, `.agents/skills`;
- `.pi/prompts`;
- `.pi/extensions`;
- `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md`;
- project trust for project-local resources;
- no built-in sandbox;
- compaction and JSONL session format.

LazyAI requirements:

- [ ] Generate `.pi/skills/<name>/SKILL.md`.
- [ ] Generate `.pi/prompts/*.md` for prompt templates.
- [ ] Generate `.pi/extensions/lazyai/*` only when hooks need it.
- [ ] Generate `.pi/settings.json` without secrets.
- [ ] Warn that Pi project trust is not a sandbox.
- [ ] Warn that Pi runs with local user permissions.
- [ ] Support `.agents/skills` compatibility output.

Conformance tests:

```text
pi_skills_frontmatter
pi_prompt_template_frontmatter
pi_project_trust_warning
pi_no_inline_secret_settings
pi_extension_only_when_needed
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
- marketplace/plugin install concepts;
- compatibility with Claude Code plugin registry concepts.

LazyAI requirements:

- [ ] Generate `AGENTS.md`.
- [ ] Generate `.omp/mcp.json`.
- [ ] Generate `.omp/skills/*/SKILL.md`.
- [ ] Generate `.omp/commands/*`.
- [ ] Generate `.omp/hooks/*` when supported.
- [ ] Generate plugin bundle optionally.
- [ ] Compile handoff/compaction assets.
- [ ] Mark adapter as beta until official docs snapshots are fully captured.

Conformance tests:

```text
omp_generates_mcp_json
omp_generates_skills
omp_generates_commands
omp_plugin_bundle_shape
omp_handoff_skill_exists
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

- [ ] Generate `.agents/skills/<name>/SKILL.md`.
- [ ] Generate `.agents/rules/*.md`.
- [ ] Generate `AGENTS.md` and optionally `GEMINI.md`.
- [ ] Generate hooks config only when current schema is verified.
- [ ] Do not write user-global settings without explicit consent.
- [ ] Warn about permissions/sandbox assumptions.
- [ ] Mark adapter as beta/experimental until docs snapshots are fully captured.

Conformance tests:

```text
antigravity_generates_agents_skills
antigravity_generates_agents_rules
antigravity_no_global_settings_by_default
antigravity_hook_schema_guarded
antigravity_sandbox_warning
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

- [ ] Generate `.kiro/steering/*.md`.
- [ ] Generate steering frontmatter for inclusion modes.
- [ ] Generate `.kiro/specs/<name>/requirements.md`, `design.md`, `tasks.md` when specs enabled.
- [ ] Generate `.kiro/hooks/*` from canonical hooks only after a source-verified native hook output contract exists. Current LazyAI Kiro hooks are instruction-only.
- [ ] Generate `.kiro/settings/mcp.json`.
- [ ] Generate `.kiroignore` for sensitive exclusions.
- [ ] Warn that Supervised mode is not a sandbox.
- [ ] Warn about broad trusted command wildcards.

Conformance tests:

```text
kiro_steering_foundational_files
kiro_steering_inclusion_modes
kiro_specs_three_phase
kiro_mcp_settings_shape
kiro_security_warnings
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
| root rules | Claude `CLAUDE.md` plus optional `AGENTS.md` |
| root rules | Copilot `.github/copilot-instructions.md` |
| scoped rules | Copilot `.github/instructions/*.instructions.md` |
| root/scoped rules | Kiro `.kiro/steering/*.md` |
| root/scoped rules | Antigravity `AGENTS.md`, `GEMINI.md`, `.agents/rules/*` |
| root/rules | Pi `AGENTS.md`, `.pi/SYSTEM.md`/`APPEND_SYSTEM.md` optional |
| root/rules | OMP `AGENTS.md`, `.omp` surfaces |

### 9.3 MCP compatibility

MCP config must be target-specific and must preserve:

- command/args;
- local vs remote;
- headers/env references;
- OAuth config where supported;
- timeouts where supported;
- enabled/disabled flags;
- tool allow/deny mapping where supported.

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
