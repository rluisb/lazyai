# LazyAI + vibe-lab — Full Product PRD

Status: Draft v0.1  
Verification date: 2026-06-21  
Product surface: complete product, not a single feature  
Audience: maintainers, contributors, early adopters, adapter owners, security reviewers, technical leads

---

## 1. Executive summary

LazyAI is a portable, inspectable, repeatable AI-agent harness asset manager and compiler.

vibe-lab is the opinionated quality and workflow layer that supplies agent roles, skills, hooks, evidence rules, rubrics, clean-code-for-agents practices, RPI discipline, human gates, trace/eval thinking, and anti-slop defaults.

Together they solve AI setup fragmentation across coding assistants:

- one canonical `.ai/` source of truth;
- many tool-native outputs;
- no hidden runtime magic;
- no framework lock-in;
- optional memory/state only when explicitly enabled;
- human-gated quality and trace-backed improvement;
- adapters that respect each supported tool's official docs and file surfaces.

LazyAI should not compete with Claude Code, OpenCode, GitHub Copilot, Pi, OMP, Kiro, Antigravity, or any host coding agent. LazyAI makes those tools better configured, more consistent, and easier to validate.

### Product promise

> Define AI-agent behavior once, validate it, version it, and emit it correctly for every supported tool.

### Core split

| Layer | Owner | Responsibility |
|---|---|---|
| Principles / taste | vibe-lab | RPI workflow, human gates, anti-slop rules, skills, hooks, eval/rubric thinking |
| Canonical source | LazyAI | `.ai/`, manifests, MCP catalog, embedded library assets |
| Compilation | LazyAI | Emit tool-native files for each AI surface |
| Execution | Host tool | Claude Code, OpenCode, Copilot, Pi, Antigravity/Gemini, Kiro |
| Optional memory/state | Runtime extras | Sessions, ledger, memory, messages, metrics, costs |

---

## 2. Problem statement

AI-agent environments drift quickly.

Teams end up with:

- different root instruction files per tool;
- duplicated agent definitions;
- inconsistent skill formats;
- scattered MCP/server configs;
- tool-specific hook formats;
- undocumented local conventions;
- brittle manual setup steps;
- no standard validation;
- no update path;
- no way to compare generated setup against official tool requirements;
- no clear boundary between agent runtime and setup assets.

The result is an expensive, inconsistent, hard-to-debug AI engineering environment. A developer may have OpenCode configured well, Claude Code configured differently, Copilot missing skills, Kiro missing steering files, and Antigravity using stale rule files.

LazyAI turns `.ai/` into the canonical source and compiles it into each supported tool surface.

---

## 3. Product goals

### 3.1 Primary goals

1. **Portable harness assets**  
   Maintain agents, skills, hooks, prompts, rules, standards, templates, MCP catalog, and adapter config in one canonical source.

2. **Tool-native output**  
   Generated files must look and behave like native files for each supported tool.

3. **Validation before execution**  
   Detect schema errors, unsafe hook definitions, duplicate skill names, invalid MCP configs, path hazards, and adapter-specific incompatibilities before the host tool runs.

4. **Official-doc compliance**  
   Each adapter must be traceable to the official docs for its host tool. Official requirements become adapter conformance tests.

5. **Human-gated quality**  
   Important transitions require approval, evidence, and deterministic verification.

6. **No hidden orchestration**  
   LazyAI compiles harness assets. Host tools execute agents. Runtime extras stay optional.

7. **Trace/eval improvement loop**  
   Improve harnesses from observed failures, curated eval cases, holdouts, and human review — not vibes.

### 3.2 Secondary goals

1. Help teams bootstrap high-quality AI coding setups quickly.
2. Make agent setup reviewable in pull requests.
3. Enable migration from ad-hoc Claude/OpenCode/Copilot/Kiro/Pi/OMP/Gemini config into `.ai/`.
4. Allow ejection: generated native files should remain useful without LazyAI.
5. Support personal, repo, workspace, and team/enterprise scopes.
6. Support plugin/package generation for tools that support plugin bundles.

---

## 4. Non-goals

LazyAI must not become:

- an agent runtime replacement;
- a LangChain/LangGraph/CrewAI-style orchestration framework;
- a mandatory RAG framework;
- a mandatory memory daemon;
- a hidden autonomous workflow runner;
- a multi-agent swarm executor;
- a cloud tracing platform;
- a model provider abstraction that owns inference execution;
- a tool that silently edits generated assets without review.

LazyAI may provide optional runtime-adjacent commands, but these must remain clearly separated from setup compilation.

---

## 5. Product principles

1. **Canonical source first**  
   `.ai/` owns agents, skills, MCP catalog, specs, setup state, and adapter manifests.

2. **Tool-native output**  
   Generated files must follow each host tool's conventions, not force LazyAI naming everywhere.

3. **Compile before execute**  
   Inspectable definitions beat hidden runtime magic.

4. **Safe by default**  
   The product should warn about trust, sandbox, credentials, destructive hooks, and agent permissions.

5. **Opt-in extras**  
   Memory, sessions, ledgers, metrics, costs, and orchestration-like behavior must not pollute baseline setup.

6. **No framework lock-in**  
   Conceptual inspiration is allowed; core behavior must not depend on external agent frameworks.

7. **Human-gated quality**  
   Agents assist, but humans approve important transitions.

8. **Trace evidence over vibes**  
   Harness changes should be backed by observed failures, evals, and review.

9. **Anti-speculation**  
   Do not add platforms, abstractions, or workflows unless the spec or official adapter docs require them.

10. **Ejectability**  
   A team should be able to keep the generated native files and stop using LazyAI.

---

## 6. Personas

### 6.1 Solo senior engineer

Uses multiple AI coding tools: OpenCode, Claude Code, Copilot CLI/IDE, Pi/OMP, Kiro, and Antigravity. Wants one consistent setup, controlled permissions, and reusable skills.

Needs:

- personal global defaults;
- repo-specific overrides;
- quick `init`, `compile`, `validate`;
- compatibility with subscriptions and local tools;
- readable generated files.

### 6.2 Staff engineer / team lead

Owns coding standards across repositories. Wants repeatable AI setup for team members.

Needs:

- versioned `.ai/` assets;
- PR-reviewable changes;
- official tool conformance;
- team skills and agents;
- compliance/security guardrails;
- easy onboarding.

### 6.3 Platform / developer-experience engineer

Maintains shared engineering tooling.

Needs:

- adapter conformance tests;
- migration tooling;
- update/eject flows;
- package/plugin generation;
- policy integration;
- enterprise-ready docs.

### 6.4 Security reviewer

Needs visibility into what agent tools can read, write, run, or call externally.

Needs:

- generated permission summaries;
- MCP server inventory;
- hook audit report;
- secret handling policy;
- sandbox guidance;
- traceable official-doc requirements.

### 6.5 AI workflow designer

Creates agents, skills, prompts, hooks, rubrics, and templates.

Needs:

- reusable authoring templates;
- validation feedback;
- skill trigger quality checks;
- eval/rubric scaffolding;
- adapter previews.

---

## 7. Jobs to be done

1. When starting a repo, I want to run `lazyai-cli init` so that all supported AI tools receive a consistent baseline setup.
2. When I add a skill, I want LazyAI to emit it into each tool's supported skill locations and warn if the name/description is invalid.
3. When a host tool updates its file format, I want LazyAI to detect adapter drift and guide migration.
4. When I use multiple AI tools, I want a single MCP catalog compiled into each tool's config without leaking secrets.
5. When a hook could block or run commands, I want validation and a human-readable risk report.
6. When an agent fails repeatedly, I want to turn traces into eval cases and targeted harness improvements.
7. When onboarding a teammate, I want them to clone the repo and get the same AI setup with one command.
8. When leaving LazyAI, I want `eject` to preserve useful native files.

---

## 8. Product scope

### 8.1 Supported host surfaces

| Tool / surface | Product status | Required LazyAI outputs |
|---|---:|---|
| OpenCode | Supported adapter | `AGENTS.md`, `.opencode/`, `opencode.json`, `.opencode/agents/*`, `.opencode/skills/*`, `.opencode/commands/*`, `.opencode/plugins/vibe-lab-hooks.js`, `opencode.json` |
| Claude Code | Supported adapter | `CLAUDE.md`, `AGENTS.md` optional/shared, `.claude/`, `.claude/settings.json`, `.claude/settings.local.json`, `.claude/skills/*`, `.claude/agents/*`, `.claude/hooks/*`, `.mcp.json`, optional plugin bundle |
| GitHub Copilot | Supported adapter | `.github/copilot-instructions.md`, `.github/instructions/*.instructions.md`, `.github/agents/*`, `.github/prompts/*`, `.github/chatmodes/*`, `.github/skills/*`, `.agents/skills/*` optional, `.vscode/mcp.json`, optional `~/.copilot/mcp-config.json`, optional Copilot CLI plugin bundle |
| Pi | Supported adapter | `AGENTS.md`, `.pi/settings.json`, `.pi/SYSTEM.md` or `.pi/APPEND_SYSTEM.md` when selected, `.pi/skills/<name>/SKILL.md`, `.pi/prompts/*.md`, `.pi/extensions/*`, optional `.agents/skills/*` |
| OMP / Oh My Pi | Supported adapter | `AGENTS.md`, `.omp/agents/*`, `.omp/skills/*`, `.omp/commands/*`, `.omp/hooks/*`, `.omp/mcp.json`, optional plugin bundle |
| Antigravity / Gemini CLI | Supported adapter | `AGENTS.md`, `GEMINI.md` optional/shared, `.agents/skills/*`, `.agents/rules/*`, `.agents/hooks.json` or official hook location, `.gemini/antigravity-cli/settings.json` guidance, `.gemini/hooks/lazyai/*` where supported |
| Kiro | Supported adapter | `AGENTS.md`, `.kiro/steering/*`, `.kiro/specs/*/{requirements.md,design.md,tasks.md}`, `.kiro/hooks/*`, `.kiro/settings/mcp.json`, `.kiroignore`, optional `.kiro/agents/*` if supported by current CLI |

### 8.2 Canonical `.ai/` source

LazyAI canonical source should support:

```text
.ai/
  lazyai.json
  lock.json
  mcp.json
  agents/
  skills/
  rules/
  fragments/
  prompts/
  commands/
  chatmodes/
  hooks/
  templates/
  standards/
  specs-agents/
  constitution/
  infra/
  cupcake/
  evals/
  traces/
  packages/
```

### 8.3 Embedded library

LazyAI includes an embedded library at:

```text
packages/cli/library/
  canonical/agents/
  skills/
  rules/
  fragments/
  prompts/
  opencode/commands/
  claudecode/commands/
  chatmodes/
  opencode/modes/
  hooks/
  templates/
  standards/starter/
  specs-agents/
  constitution/
  infra/
  cupcake/
```

---

## 9. Functional requirements

### 9.1 Initialization

`lazyai-cli init` must:

- create `.ai/lazyai.json`;
- create `.ai/mcp.json` if MCP enabled;
- install selected baseline asset pack;
- detect target tools from repo/user selection;
- generate initial native files through `compile`;
- show a diff before writing unless `--yes` is supplied;
- avoid overwriting user files without backup or merge plan;
- record generated files in `.ai/lock.json`.

Acceptance criteria:

- fresh repo can initialize with OpenCode + Claude + Copilot + Pi + Kiro + Antigravity + OMP targets;
- generated outputs pass adapter validation;
- generated outputs are marked with safe headers where appropriate;
- non-generated existing files are never destroyed silently.

### 9.2 Add assets

`lazyai-cli add` must support:

```text
lazyai-cli add agent reviewer
lazyai-cli add skill pr-review
lazyai-cli add hook protected-paths
lazyai-cli add prompt bugfix-rpi
lazyai-cli add mcp ai-memory
lazyai-cli add standard starter/security
```

Requirements:

- asset templates are local and inspectable;
- selected asset dependencies are resolved;
- adapter compatibility is checked;
- additions update `.ai/lazyai.json` and relevant asset folders;
- no native output changes happen until `compile`, unless `--compile` is provided.

### 9.3 Compilation

`lazyai-cli compile` must:

1. discover canonical assets;
2. resolve inheritance/fragments;
3. validate canonical schema;
4. validate target adapter compatibility;
5. build a compile plan;
6. show diffs;
7. write generated files atomically;
8. update `.ai/lock.json`;
9. run post-write smoke checks.

Required flags:

```text
--target opencode
--target claude
--target copilot
--target pi
--target omp
--target antigravity
--target kiro
--all
--dry-run
--diff
--write
--check
--force
--no-backup
--profile team|personal|ci
```

### 9.4 Validation

`lazyai-cli validate` must validate:

- canonical manifest schema;
- agent schema;
- skill frontmatter and naming;
- hook policy safety;
- MCP catalog schema;
- adapter outputs;
- file collision behavior;
- official docs conformance fixtures;
- deprecated fields;
- path traversal and symlink hazards;
- secret leakage risk;
- generated file lock consistency.

Specialized validation:

```text
lazyai-cli validate skills
lazyai-cli validate agents
lazyai-cli validate hooks
lazyai-cli validate mcp
lazyai-cli validate adapters
lazyai-cli validate generated
lazyai-cli validate official-docs
```

### 9.5 Doctor/status/info/list

`lazyai-cli doctor` must inspect:

- installed target tools;
- versions;
- supported file surfaces;
- generated file drift;
- missing dependencies;
- MCP server availability;
- permissions risk;
- official-doc adapter compatibility status.

`lazyai-cli status` must show:

- enabled targets;
- dirty generated files;
- pending compile plan;
- enabled optional runtime extras;
- warnings.

`lazyai-cli info` and `lazyai-cli list` must expose asset catalogs and adapters.

### 9.6 Update and migrate

`lazyai-cli update` must update embedded library assets and adapter definitions with a reviewable diff.

`lazyai-cli migrate` must:

- import existing tool-native config into `.ai/`;
- preserve unsupported native fields;
- mark uncertain conversions;
- generate a migration report.

Migration sources:

- `AGENTS.md`;
- `CLAUDE.md`;
- `.claude/`;
- `.opencode/`;
- `opencode.json`;
- `.github/copilot-instructions.md`;
- `.github/instructions/`;
- `.github/skills/`;
- `.agents/skills/`;
- `.pi/`;
- `.omp/`;
- `.kiro/`;
- `.gemini/`;
- `.mcp.json`;
- `.vscode/mcp.json`.

### 9.7 Eject

`lazyai-cli eject` must:

- remove LazyAI metadata when requested;
- preserve native files;
- optionally remove generated headers;
- write an eject report;
- leave host tools usable.

### 9.8 MCP catalog

`.ai/mcp.json` is the canonical source for MCP servers.

It must compile to:

- `opencode.json` MCP section;
- `.mcp.json` for Claude Code/project-level MCP;
- `.claude/settings.local.json` or managed settings guidance where appropriate;
- `.vscode/mcp.json` for IDE/Copilot surfaces;
- `.kiro/settings/mcp.json`;
- `.omp/mcp.json`;
- Antigravity/Gemini CLI MCP config location according to official current settings;
- optional personal/global configs only with explicit user consent.

Security requirements:

- secrets are referenced by environment variable, never written inline by default;
- remote MCP server OAuth metadata is preserved when supported;
- unsupported MCP fields are rejected or placed in adapter-specific extension blocks;
- generated reports identify local vs remote MCP servers.

### 9.9 Skills

Skills are canonical capability packages.

Requirements:

- skill directory name must match canonical skill `name` unless adapter explicitly supports divergence;
- `SKILL.md` must include `name` and `description`;
- descriptions must be specific enough for automatic invocation;
- skill may contain scripts, reference docs, assets;
- skill must declare required tools and evidence expectations where relevant;
- skill must include trigger and non-trigger guidance;
- skill must support progressive disclosure.

Adapter outputs must support official skill locations for:

- OpenCode `.opencode/skills/<name>/SKILL.md`, `.claude/skills`, `.agents/skills` compatibility;
- Claude Code `.claude/skills/<name>/SKILL.md` and plugin skills;
- GitHub Copilot `.github/skills`, `.claude/skills`, `.agents/skills`, personal `~/.copilot/skills` when explicit;
- Pi `.pi/skills` and `.agents/skills` compatibility;
- Antigravity `.agents/skills` with backward compatibility considerations;
- OMP `.omp/skills` and plugin skills;
- Kiro steering/specs/hook integrations; skill support if official CLI/API supports it.

### 9.10 Agents

Canonical agents must contain:

- `name`;
- `description`;
- role/purpose;
- allowed tools/capabilities;
- disallowed actions;
- default model preference or model class, not hard-coded provider unless target-specific;
- instructions body;
- input/output expectations;
- handoff requirements;
- evidence requirements;
- optional color/icon/metadata.

Default agents:

- `guide`;
- `implementer`;
- `researcher`;
- `planner`;
- `reviewer`;
- `deployer`;
- `responder`;
- `evidence-verifier`.

Adapter outputs:

- OpenCode: `.opencode/agents/*.md` or `opencode.json` `agent` entries;
- Claude Code: `.claude/agents/*.md` or plugin subagents;
- GitHub Copilot: `.github/agents/*.agent.md` or Copilot CLI plugin agent files;
- Kiro: steering/spec agents and CLI custom-agent output if officially supported;
- Pi/Antigravity: target-native agent or rule/instruction equivalents.

### 9.11 Rules and root instructions

Root instructions must be generated in tool-native files:

- OpenCode: `AGENTS.md`, optionally `opencode.json` `instructions`;
- Claude Code: `CLAUDE.md` plus optional shared `AGENTS.md`;
- Copilot: `.github/copilot-instructions.md`, `.github/instructions/*.instructions.md`, optional `AGENTS.md`;
- Kiro: `.kiro/steering/*.md`, `AGENTS.md` supported where applicable;
- Antigravity/Gemini: `AGENTS.md`, `GEMINI.md`, `.agents/rules/*`;
- Pi: `AGENTS.md`, `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md` where selected;
- OMP: `AGENTS.md`, `.omp` rules/config surfaces.

Rules must be split into:

- always-on rules;
- conditional path/file-match rules;
- manually invoked rules;
- adapter-only rules;
- generated comments/instructions that explain origin.

### 9.12 Hooks / middleware

LazyAI canonical hook policy must support:

- `before_agent`;
- `before_model`;
- `before_tool`;
- `after_tool`;
- `after_model`;
- `after_agent`;
- `on_error`;
- `on_compaction`;
- `on_handoff`;
- `on_human_gate`;
- `on_file_change`;
- `on_config_change`.

Adapters compile to whatever the host tool supports:

- Claude Code hook events and settings;
- OpenCode plugins;
- Pi TypeScript extensions/events;
- Kiro hooks with event/action configuration;
- Copilot CLI hooks;
- Antigravity hook JSON/config;
- OMP hook/plugin surfaces.

Safety rules:

- destructive shell hooks require explicit approval;
- hooks that modify files must declare paths;
- hooks that call network services must declare endpoints;
- hooks that read secrets must be blocked by default;
- async/background hooks must be clearly labeled;
- hooks must be testable without running destructive commands.

### 9.13 Commands and prompt templates

LazyAI must support reusable task prompts and slash-command equivalents.

Examples:

- `/rpi-bugfix`;
- `/review-pr`;
- `/security-audit`;
- `/handoff`;
- `/compact-with-evidence`;
- `/plan-from-ticket`;
- `/create-adr`.

Adapter outputs:

- OpenCode `.opencode/commands/*.md`;
- Claude Code skills/commands compatibility;
- Copilot `.github/prompts/*` and CLI plugin skills/commands;
- Pi `.pi/prompts/*.md`;
- OMP commands/prompt templates;
- Kiro specs/steering prompts;
- Antigravity prompt/rule/skill equivalents.

### 9.14 Trace/eval system

Trace/eval support is product-level but optional.

Required assets:

```text
.ai/traces/taxonomy.md
.ai/evals/cases/*.yaml
.ai/evals/holdouts/*.yaml
.ai/evals/rubrics/*.md
.ai/evals/harness-changes/*.md
```

Requirements:

- `lazyai-cli validate evals` validates schemas;
- no mandatory cloud tracing provider;
- support manual import of host-tool traces/session summaries;
- support failure taxonomy;
- support one-change-at-a-time harness improvement reports;
- support holdout cases;
- support human approval before promotion.

### 9.15 Runtime-adjacent extras

These remain optional:

- `session`;
- `ledger`;
- `memory`;
- `message`;
- `metrics`;
- `cost`;
- `secret`;
- `backup`;
- `auth`;
- `notify`;
- `git`;
- `restore-runtime-db`.

They must not be required for `init`, `compile`, `validate`, `doctor`, or native-file execution.

---

## 10. Adapter-specific product requirements

## 10.1 OpenCode adapter

### Official surface summary

OpenCode supports root instructions via `AGENTS.md`, agents via JSON or Markdown files, primary agents and subagents, Plan mode with restricted edits/commands, MCP servers, skills via `SKILL.md`, commands, permissions, plugins, and configuration in `opencode.json` / `.opencode/`.

### LazyAI outputs

```text
AGENTS.md
opencode.json
.opencode/
  agents/
  skills/
  commands/
  plugins/vibe-lab-hooks.js
  opencode.json
```

### Requirements

- MUST generate `AGENTS.md` with root harness principles and repo instructions.
- MUST generate `.opencode/agents/*.md` for supported agents.
- MUST use `steps`, not deprecated `maxSteps`, when configuring iteration limits.
- MUST generate skill directories as `.opencode/skills/<name>/SKILL.md` with required frontmatter.
- SHOULD also generate `.agents/skills/<name>/SKILL.md` when cross-harness skill sharing is enabled.
- MUST generate MCP config in a way OpenCode can load.
- MUST use permissions for edit/bash/tool safety.
- SHOULD generate `.opencode/commands/*.md` for task prompts.
- SHOULD emit `.opencode/plugins/vibe-lab-hooks.js` for hook policies that cannot be represented declaratively.

### Acceptance criteria

- `opencode` loads all generated agents and skills.
- Plan/reviewer/evidence-verifier agents cannot edit files unless explicitly allowed.
- Skill names validate against OpenCode name rules.
- MCP servers show up under OpenCode MCP list/debug commands.
- Permissions are not generated using deprecated fields unless compatibility mode is selected.

## 10.2 Claude Code adapter

### Official surface summary

Claude Code reads project/user instructions, settings, skills, subagents, hooks, MCP, permissions, and plugins from `.claude` and home config locations. Claude Code uses `CLAUDE.md` as project instruction memory and supports skills, subagents, hooks, MCP, permissions, managed settings, plugin bundles, and platform/API features.

### LazyAI outputs

```text
CLAUDE.md
AGENTS.md                    # optional shared cross-tool root, not replacement for CLAUDE.md
.mcp.json
.claude/
  settings.json
  settings.local.json
  skills/<name>/SKILL.md
  agents/<name>.md
  hooks/*.sh
  plugins/lazyai-vibelab/
```

### Requirements

- MUST generate `CLAUDE.md` for Claude Code compliance.
- SHOULD also generate `AGENTS.md` as cross-tool shared context when selected.
- MUST generate `.claude/skills/<name>/SKILL.md` for skills.
- MUST generate `.claude/agents/*.md` or plugin subagents where compatible.
- MUST generate `.mcp.json` for project MCP when enabled.
- MUST generate `.claude/settings.json` or `.claude/settings.local.json` carefully, without secrets.
- MUST support permission rules through settings where safe.
- MUST support hooks through official hook configuration schema.
- SHOULD support plugin bundle generation for reusable team distribution.
- MUST detect managed settings restrictions and warn when project/user sources may be ignored.

### Compliance gap from current state

The current LazyAI state lists `AGENTS.md` for Claude Code. Official Claude Code docs require `CLAUDE.md` as the primary project instruction file. The Claude adapter must add `CLAUDE.md` output and may keep `AGENTS.md` as shared cross-tool documentation.

## 10.3 GitHub Copilot adapter

### Official surface summary

GitHub Copilot supports repository custom instructions, path-specific instructions, agent instructions through supported root files, custom agents, agent skills, hooks for Copilot CLI, MCP across Copilot surfaces, Copilot CLI plugins, and plugin bundles that can include agents, skills, hooks, and MCP server configs.

### LazyAI outputs

```text
.github/
  copilot-instructions.md
  instructions/*.instructions.md
  agents/*.agent.md
  prompts/*.prompt.md
  chatmodes/*.chatmode.md
  skills/<name>/SKILL.md
.vscode/mcp.json
.agents/skills/<name>/SKILL.md     # optional cross-harness compatibility
AGENTS.md                         # optional shared root instructions
```

Optional user/global outputs only with explicit consent:

```text
~/.copilot/skills/<name>/SKILL.md
~/.copilot/mcp-config.json
```

Optional plugin output:

```text
copilot-plugin/
  plugin.json
  agents/*.agent.md
  skills/<name>/SKILL.md
  hooks.json
  .mcp.json
```

### Requirements

- MUST generate `.github/copilot-instructions.md` from canonical root rules.
- SHOULD generate path-specific `.github/instructions/*.instructions.md` when rules have file/path scopes.
- MUST generate skills into at least one official project skill location.
- SHOULD generate `.agents/skills` for cross-tool compatibility when enabled.
- MUST generate `.vscode/mcp.json` for IDE Copilot MCP config.
- SHOULD support Copilot CLI plugin packaging.
- MUST respect project/personal/plugin precedence and warn on duplicate agent/skill names.

## 10.4 Pi adapter

### Official surface summary

Pi is a minimal terminal coding harness extended through TypeScript extensions, skills, prompt templates, themes, and packages. Pi has project trust for project-local resources but does not include a built-in sandbox. It supports `.pi/settings.json`, `.pi/skills`, `.pi/prompts`, `.pi/extensions`, `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md`, `.agents/skills`, compaction, sessions, JSONL/RPC/programmatic modes, and extension events.

### LazyAI outputs

```text
AGENTS.md
.pi/
  settings.json
  SYSTEM.md              # optional
  APPEND_SYSTEM.md       # optional
  skills/<name>/SKILL.md
  prompts/*.md
  extensions/lazyai/*.ts
.agents/skills/<name>/SKILL.md     # optional compatibility
```

### Requirements

- MUST treat `.pi` project resources as trust-sensitive and warn during `doctor`.
- MUST not imply Pi project trust is a sandbox.
- MUST generate skills with `SKILL.md` frontmatter and progressive disclosure.
- SHOULD generate prompt templates as slash-command markdown snippets.
- SHOULD generate TypeScript extensions only for hook policies that Pi extension events can support.
- MUST provide sandbox guidance through `ai-jail`, containers, or OS isolation for untrusted work.
- MUST avoid writing secrets into `.pi/settings.json`.

## 10.5 OMP adapter

### Official surface summary

OMP is a terminal-first coding agent with subagents, plan mode, LSP/DAP, compaction, handoff, memory, rules, plugins, skills, commands, hooks, custom tools, MCP, and marketplace/plugin concepts. OMP docs indicate `.omp/mcp.json` project MCP, user MCP config, plugins bundling skills/commands/hooks/MCP servers/themes, and compatibility with Claude Code-style plugin registry concepts.

### LazyAI outputs

```text
AGENTS.md
.omp/
  agents/
  skills/
  commands/
  hooks/
  mcp.json
  plugins/lazyai-vibelab/
```

### Requirements

- MUST generate `.omp/mcp.json` for project MCP.
- MUST generate skills and commands in OMP-supported locations.
- SHOULD generate plugin bundles for team distribution.
- SHOULD map vibe-lab handoff and compaction practices into OMP handoff/compaction concepts.
- MUST preserve OMP-specific settings not owned by LazyAI.
- MUST mark OMP adapter conformance as dependent on official docs snapshots because some pages are rendered with limited machine-readable content.

## 10.6 Antigravity / Gemini CLI adapter

### Official surface summary

Antigravity/Gemini CLI docs expose rules, skills, hooks, MCP, plugins, sidecars, permissions, sandbox/security, and settings under the Gemini/Antigravity CLI config paths. Official snippets indicate skills under `.agents/skills` with backward `.agent/skills`, workspace rules under `.agents/rules`, settings under `~/.gemini/antigravity-cli/settings.json`, plugin bundles containing skills/agents/rules/MCP/hooks, permissions such as command allow rules, and sandbox settings.

### LazyAI outputs

```text
AGENTS.md
GEMINI.md                     # optional target-native root context
.agents/
  skills/<name>/SKILL.md
  rules/*.md
  hooks.json
.gemini/
  hooks/lazyai/*.sh           # only where supported by current docs/runtime
```

User/global guidance only with explicit consent:

```text
~/.gemini/antigravity-cli/settings.json
```

### Requirements

- MUST generate `.agents/skills` as primary skill output.
- SHOULD support backward `.agent/skills` only in compatibility mode.
- MUST generate `.agents/rules` for scoped workspace rules.
- SHOULD generate `GEMINI.md` when selected.
- MUST not silently edit user-global Antigravity settings.
- MUST represent permissions/sandbox instructions as guidance unless explicitly authorized to write global settings.
- MUST validate hooks against official hook schema when available.

## 10.7 Kiro adapter

### Official surface summary

Kiro supports spec-driven development, steering files, hooks, MCP, agentic chat, supervised/autopilot modes, trusted commands, protected paths, and privacy/security controls. Specs typically contain requirements/bug analysis, design, and tasks. Steering files live in workspace `.kiro/steering/` and global `~/.kiro/steering/`, with inclusion modes such as always, fileMatch, manual, and auto.

### LazyAI outputs

```text
AGENTS.md
.kiro/
  steering/
    product.md
    tech.md
    structure.md
    vibe-lab.md
    security.md
  specs/<spec-name>/
    requirements.md
    design.md
    tasks.md
  hooks/
  settings/mcp.json
  agents/                 # only if current official CLI supports this surface
.kiroignore
```

### Requirements

- MUST generate steering files for always-on project rules and conventions.
- SHOULD generate fileMatch/manual/auto steering frontmatter when canonical rules have scopes.
- MUST generate specs from LazyAI templates when user requests spec-driven output.
- MUST generate Kiro hook definitions from canonical hooks where supported.
- MUST generate `.kiro/settings/mcp.json` for MCP.
- MUST warn that Kiro Supervised mode is a review workflow, not a sandbox.
- MUST support protected-path guidance and `.kiroignore` generation.
- MUST avoid overloading Kiro with non-native agent files unless officially supported.

---

## 11. Security and privacy requirements

### 11.1 Core requirements

- No secrets in generated files by default.
- Environment variable references preferred.
- Generated files must be reviewable before write.
- Destructive hooks are blocked unless explicit approval is configured.
- Global user config writes require explicit consent.
- Project-local config writes require diff preview unless `--yes`.
- Symlinks must be resolved safely.
- Path traversal outside workspace must be rejected unless target is explicitly global and approved.
- Plugin/package installs must show source, version, and components.
- All generated outputs must support ejection.

### 11.2 Sandbox stance

LazyAI must state clearly:

- LazyAI is not a sandbox.
- Host tool trust modes are not necessarily sandboxes.
- Pi project trust is not a sandbox.
- Kiro Supervised mode is not a sandbox.
- Real isolation requires OS/container/VM/sandbox boundary.
- `ai-jail` is a helper, not a universal guarantee.

### 11.3 Credentials

- MCP env vars must be passed by name.
- Local secret store integration may be optional.
- Logs must redact known secret patterns.
- `lazyai-cli doctor` must detect inline secrets and warn.
- `lazyai-cli compile` must fail on inline secrets in team profile unless explicitly overridden.

### 11.4 Supply chain

- Asset packs must have provenance metadata.
- Plugin bundles must include manifests.
- External package downloads require explicit consent.
- Checksums should be recorded when installing remote packs.
- Enterprise mode can restrict marketplaces or external sources.

---

## 12. UX flows

### 12.1 New repo setup

```bash
lazyai-cli init --targets opencode,claude,copilot,pi,kiro,antigravity,omp
lazyai-cli validate
lazyai-cli compile --dry-run
lazyai-cli compile --write
lazyai-cli doctor
```

Expected output:

```text
Created .ai/lazyai.json
Added starter vibe-lab harness pack
Enabled targets: opencode, claude, copilot, pi, kiro, antigravity, omp
Generated 43 files
Warnings:
  - Claude adapter will create CLAUDE.md because official Claude Code reads it.
  - Pi .pi resources require project trust and are not sandboxed.
  - Kiro supervised mode is not a sandbox; protected paths configured.
```

### 12.2 Add a skill

```bash
lazyai-cli add skill pr-review --pack vibe-lab
lazyai-cli validate skills
lazyai-cli compile --target opencode --target claude --target copilot --target pi
```

Expected behavior:

- creates `.ai/skills/pr-review/SKILL.md`;
- validates name/description;
- emits target-native skill directories;
- warns on unsupported metadata fields;
- updates lockfile.

### 12.3 Add MCP server

```bash
lazyai-cli add mcp ai-memory
lazyai-cli compile --all
lazyai-cli doctor mcp
```

Expected behavior:

- adds `.ai/mcp.json` entry;
- writes target configs;
- uses env vars for secrets;
- reports which tools can use it.

### 12.4 Migrate existing repo

```bash
lazyai-cli migrate --from existing --targets opencode,claude,copilot
lazyai-cli validate
lazyai-cli compile --dry-run
```

Expected behavior:

- imports `AGENTS.md`, `CLAUDE.md`, `.opencode`, `.github` assets;
- preserves unsupported fields;
- marks uncertain conversions;
- generates migration report.

### 12.5 Eject

```bash
lazyai-cli eject --keep-native --remove-lock
```

Expected behavior:

- keeps native tool files;
- removes LazyAI metadata if requested;
- writes `lazyai-eject-report.md`.

---

## 13. Success metrics

### Activation

- Time from `lazyai-cli init` to valid generated setup.
- Percentage of users who run `compile` after `init`.
- Number of enabled targets per repo.

### Quality

- Validation pass rate.
- Skill trigger errors found before compile.
- MCP config errors prevented.
- Hook safety warnings resolved.
- Generated file drift detected.

### Adoption

- Repos using multiple adapters.
- Teams adopting shared asset packs.
- Plugins/packages generated.
- Successful migrations.

### Reliability

- Adapter conformance test pass rate.
- Official-doc drift incidents.
- Failed compiles due to schema mismatch.
- User-reported overwritten-file incidents.

### Safety

- Inline secret detections.
- Unsafe hook blocks.
- Global config writes requiring confirmation.
- Sandbox/trust warnings shown.

---

## 14. Risks and mitigations

| Risk | Impact | Mitigation |
|---|---:|---|
| Official docs change frequently | High | Adapter versioning, docs snapshots, conformance tests, `doctor` warnings |
| Tool-native configs conflict with user files | High | Lockfile, generated regions, backups, merge strategy, dry-run diffs |
| LazyAI becomes a runtime | High | Product boundary: compile assets only; runtime extras opt-in |
| Hooks execute unsafe commands | High | Hook validator, destructive-command denylist, human approval, sandbox guidance |
| Skills trigger too often | Medium | Trigger/non-trigger rubric, eval cases, validation warnings |
| MCP leaks secrets | High | env var references, redaction, secret scan, team profile fail-fast |
| Adapter gaps create false confidence | Medium | Compliance matrix, gap labels, `experimental` adapter status |
| Generated files feel non-native | Medium | Golden fixtures, manual review, host-tool smoke tests |
| Multi-agent bloat | Medium | Default to skills; specialists only when justified |

---

## 15. Release strategy

### v0.1 — Product foundations

- Canonical manifest.
- OpenCode, Claude, Copilot basic adapters.
- Skill/agent/rule compilation.
- MCP catalog compilation.
- Validate/doctor/status.

### v0.2 — Full adapter matrix

- Pi adapter.
- Kiro adapter.
- Antigravity adapter.
- OMP adapter.
- Hook catalog.
- Prompt/command compilation.

### v0.3 — Safety and migration

- Import/migrate.
- Eject.
- Security scanner.
- Hook validator.
- Secret scanner.
- Official-doc conformance snapshots.

### v0.4 — Packaging

- Tool-specific plugin/package generation.
- Copilot CLI plugin output.
- Claude Code plugin output.
- OMP plugin output.
- Pi package output.

### v0.5 — Trace/eval loop

- Trace taxonomy.
- Eval case schema.
- Holdout support.
- Harness change reports.
- Human approval workflow.

### v1.0 — Stable product

- Stable `.ai/` schema.
- Stable adapter contracts.
- Official conformance suite.
- Upgrade/migration guarantees.
- Security review complete.
- Eject flow complete.

---

## 16. Open questions

1. Should `.ai/lock.json` include hashes of generated files or normalized ASTs?
2. Should LazyAI generate `CLAUDE.md` as full content or as a shim referencing `AGENTS.md`?
3. Should `.agents/skills` be a default shared output or opt-in compatibility output?
4. Should plugin/package generation live under `build-plugin` or `compile --package`?
5. Should official-doc conformance tests be maintainer-only or available via `doctor`?
6. How much of OMP/Antigravity adapter behavior should be marked experimental until docs are easier to parse automatically?
7. Should runtime extras use a single local SQLite database or separate files per feature?
8. Should `ai-jail` be integrated as a command wrapper or only recommended by `doctor`?

---

## 17. PRD acceptance checklist

The product is complete when:

- `lazyai-cli init` can create a valid canonical setup.
- `lazyai-cli compile` can emit native outputs for all supported targets.
- `lazyai-cli validate` catches canonical and adapter-specific errors.
- `lazyai-cli doctor` reports host tool compatibility and security warnings.
- Official docs requirements are linked to adapter tests.
- Generated files are reviewable, reproducible, and ejectable.
- Skills, agents, hooks, MCP, rules, prompts, commands, and templates are all covered.
- Claude adapter emits `CLAUDE.md`.
- Kiro adapter emits steering/spec/hook/MCP files.
- Pi adapter respects project trust and sandbox caveats.
- OpenCode adapter uses current `steps` and permission fields.
- Copilot adapter supports instructions, skills, hooks, MCP, and plugin packaging.
- Antigravity adapter supports `.agents/skills`, `.agents/rules`, hooks/permissions guidance.
- OMP adapter supports `.omp/mcp.json`, plugins, skills, commands, hooks.
- Runtime-adjacent extras remain optional.
