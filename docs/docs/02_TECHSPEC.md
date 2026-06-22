# LazyAI + vibe-lab — Full Product Technical Specification

Status: Draft v0.1  
Verification date: 2026-06-21  
Scope: complete product architecture and adapter implementation plan

---

## 1. System architecture

LazyAI is a Go CLI with an embedded asset library, canonical schema parser, validation engine, adapter compiler system, safe file writer, migration/eject engine, optional runtime-adjacent services, and conformance test suite.

```text
                 ┌─────────────────────────────────────────┐
                 │                 .ai/                    │
                 │  canonical manifests/assets/mcp/evals   │
                 └───────────────────┬─────────────────────┘
                                     │
                                     ▼
┌─────────────────────────────────────────────────────────────────────┐
│                             lazyai-cli                              │
│                                                                     │
│  discover → parse → normalize → resolve → validate → plan → write   │
│                                                                     │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────────────┐  │
│  │ asset loader │  │ validators  │  │ adapter registry/compiler  │  │
│  └──────────────┘  └─────────────┘  └────────────────────────────┘  │
│  ┌──────────────┐  ┌─────────────┐  ┌────────────────────────────┐  │
│  │ diff planner │  │ safe writer │  │ docs conformance fixtures  │  │
│  └──────────────┘  └─────────────┘  └────────────────────────────┘  │
└──────────────┬──────────────┬─────────────┬──────────────┬──────────┘
               │              │             │              │
               ▼              ▼             ▼              ▼
        OpenCode files   Claude files   Copilot files   Pi/OMP/Kiro/Gemini
```

Execution is always delegated to host tools.

---

## 2. Repository/package layout

Recommended Go package layout:

```text
packages/cli/
  cmd/lazyai/
    main.go
  internal/
    app/
    cli/
    config/
    detect/
    assets/
    schema/
    validate/
    compile/
    adapters/
      opencode/
      claudecode/
      copilot/
      pi/
      omp/
      antigravity/
      kiro/
    mcp/
    hooks/
    skills/
    agents/
    rules/
    prompts/
    diff/
    writer/
    migrate/
    eject/
    doctor/
    security/
    docsconformance/
    runtime/
      session/
      ledger/
      memory/
      metrics/
      cost/
  library/
    canonical/agents/
    skills/
    rules/
    fragments/
    prompts/
    hooks/
    templates/
    standards/
    specs-agents/
    constitution/
    infra/
    cupcake/
  testdata/
    golden/
    fixtures/
    docs-snapshots/
```

---

## 3. Core domain model

### 3.1 Project manifest

File: `.ai/lazyai.json`

```json
{
  "$schema": "https://lazyai.dev/schemas/lazyai.schema.json",
  "version": "0.1",
  "profile": "team",
  "targets": ["opencode", "claude", "copilot", "pi", "omp", "antigravity", "kiro"],
  "source": {
    "assets": ["agents", "skills", "rules", "hooks", "prompts", "templates", "standards"],
    "packs": ["vibe-lab/starter"]
  },
  "adapters": {
    "opencode": { "enabled": true, "emitSharedAgentsSkills": true },
    "claude": { "enabled": true, "emitClaudeMd": true, "emitPlugin": false },
    "copilot": { "enabled": true, "emitPlugin": false, "emitAgentsSkillsCompat": true },
    "pi": { "enabled": true, "emitPiSystem": false },
    "omp": { "enabled": true },
    "antigravity": { "enabled": true, "emitGeminiMd": true },
    "kiro": { "enabled": true, "emitSpecs": true }
  },
  "safety": {
    "requireDiffBeforeWrite": true,
    "allowGlobalWrites": false,
    "denyInlineSecrets": true,
    "warnIfNoSandbox": true
  }
}
```

### 3.2 Lockfile

File: `.ai/lock.json`

Purpose:

- record LazyAI version;
- record adapter versions;
- record source asset hashes;
- record generated file paths/hashes;
- record last compile targets;
- detect drift.

```json
{
  "version": "0.1",
  "lazyaiVersion": "0.1.0",
  "compiledAt": "2026-06-21T00:00:00Z",
  "adapters": {
    "opencode": { "version": "2026.06.21", "docsSnapshot": "2026-06-21" }
  },
  "generated": [
    {
      "path": "AGENTS.md",
      "target": "opencode",
      "sourceHash": "sha256:...",
      "outputHash": "sha256:...",
      "managed": true
    }
  ]
}
```

### 3.3 Canonical asset metadata

All assets can use YAML frontmatter plus Markdown body.

Required common fields:

```yaml
---
name: pr-review
type: skill
version: 0.1.0
description: Review a pull request or diff with evidence, risks, and verification steps.
owners:
  - vibe-lab
compatibility:
  targets:
    - opencode
    - claude
    - copilot
    - pi
    - omp
    - antigravity
  minimumLazyAI: "0.1.0"
safety:
  canRunCommands: false
  canModifyFiles: false
  requiresHumanGate: true
---
```

### 3.4 Agent schema

Canonical file:

```text
.ai/agents/reviewer.md
```

Fields:

```yaml
---
name: reviewer
type: agent
description: Reviews code for correctness, maintainability, security, and evidence completeness.
mode: subagent
modelClass: medium-reasoning
tools:
  read: allow
  search: allow
  edit: deny
  bash: ask
permissions:
  destructive: deny
  network: ask
handoff:
  required: true
evidence:
  required: true
---
```

Body sections:

```text
## Role
## When to use
## When not to use
## Required process
## Evidence required
## Output format
## Handoff format
```

### 3.5 Skill schema

Canonical file:

```text
.ai/skills/pr-review/SKILL.md
```

Fields:

```yaml
---
name: pr-review
description: Review a pull request, staged diff, or patch for correctness, risk, and verification evidence.
license: MIT
compatibility: lazyai,opencode,claude,copilot,pi,omp,antigravity
metadata:
  workflow: review
  owner: vibe-lab
---
```

Required body sections:

```text
## What this skill does
## Trigger examples
## Non-trigger examples
## Inputs
## Procedure
## Evidence required
## Output format
## Failure modes
```

Validation rules:

- name must be lowercase kebab-case;
- directory name must match `name` for strict mode;
- description must be specific and non-empty;
- body must include trigger and non-trigger guidance;
- body must not contain inline secrets;
- scripts must be referenced relatively.

### 3.6 MCP schema

File: `.ai/mcp.json`

```json
{
  "$schema": "https://lazyai.dev/schemas/mcp-catalog.schema.json",
  "version": "0.1",
  "servers": {
    "ai-memory": {
      "type": "local",
      "command": "ai-memory",
      "args": ["server"],
      "env": {
        "AI_MEMORY_DB": "${AI_MEMORY_DB:-.ai/runtime/memory.db}"
      },
      "tools": [
        "memory_query",
        "memory_recent",
        "memory_handoff_accept",
        "memory_write_page",
        "memory_auto_improve",
        "memory_consolidate"
      ],
      "preferredUse": "hybrid"
    },
    "ripgrep": {
      "type": "local",
      "command": "rg-mcp",
      "args": [],
      "tools": ["search", "list-files"],
      "preferredUse": "cli-first"
    }
  }
}
```

Canonical server types:

- `local`;
- `remote`;
- `stdio` alias for local stdio tools;
- `http` alias for remote HTTP/SSE tools;
- `disabled`.

---

## 4. Compilation pipeline

### 4.1 Stages

```text
1. Detect workspace
2. Load manifest
3. Load embedded and project assets
4. Normalize asset metadata
5. Resolve fragments/includes
6. Build canonical graph
7. Validate canonical graph
8. Select target adapters
9. Build adapter-specific intermediate representation
10. Validate adapter IR
11. Build file write plan
12. Render diffs
13. Write atomically
14. Update lockfile
15. Run post-compile smoke checks
```

### 4.2 Compile plan

A compile plan contains:

```go
type CompilePlan struct {
    Targets       []TargetPlan
    Writes        []FileWrite
    Deletes       []FileDelete
    Warnings      []Diagnostic
    Errors        []Diagnostic
    LockfileDelta LockfileDelta
}
```

### 4.3 Safe writer

Requirements:

- write to temp file;
- fsync where possible;
- atomic rename;
- create backup before overwriting non-generated file;
- preserve file mode where appropriate;
- prevent path traversal;
- reject symlink writes unless `--follow-symlinks` explicitly set;
- support generated regions inside existing files;
- generate reversible write report.

### 4.4 Generated file headers

Recommended generated header:

```text
<!--
Generated by LazyAI.
Source: .ai/
Target: opencode
Do not edit this generated region directly unless you intend to own the native file.
Run: lazyai compile --target opencode
-->
```

For files where comments break syntax, store metadata in lockfile instead.

---

## 5. Adapter architecture

### 5.1 Adapter interface

```go
type Adapter interface {
    ID() string
    DisplayName() string
    Detect(ctx Context) DetectionResult
    Capabilities() Capabilities
    Compile(ctx CompileContext, graph CanonicalGraph) (AdapterPlan, error)
    Validate(ctx ValidateContext, files FileSet) []Diagnostic
    Migrate(ctx MigrateContext, files FileSet) (CanonicalPatch, []Diagnostic)
    Eject(ctx EjectContext, files FileSet) (EjectPlan, []Diagnostic)
}
```

### 5.2 Capability model

```go
type Capabilities struct {
    RootInstructions bool
    Agents           bool
    Subagents        bool
    Skills           bool
    Hooks            bool
    Commands         bool
    PromptTemplates  bool
    ChatModes        bool
    MCP              bool
    Permissions      bool
    Plugins          bool
    Specs            bool
    Steering         bool
    Compaction       bool
    Sessions         bool
    GlobalConfig     bool
}
```

### 5.3 Adapter support levels

```text
stable       official docs verified + golden tests + smoke tests
beta         official docs verified + golden tests, limited runtime smoke
experimental docs partially verified or host tool still moving quickly
deprecated   adapter kept for migration only
```

---

## 6. OpenCode adapter spec

### 6.1 Inputs

- canonical agents;
- canonical skills;
- root instructions;
- MCP catalog;
- commands;
- hook policies;
- permissions.

### 6.2 Outputs

```text
AGENTS.md
opencode.json
.opencode/agents/<name>.md
.opencode/skills/<name>/SKILL.md
.opencode/commands/<name>.md
.opencode/plugins/vibe-lab-hooks.js
.opencode/lazyai.mcp.jsonc
```

### 6.3 Agent mapping

Canonical agent:

```yaml
name: reviewer
mode: subagent
tools:
  edit: deny
  bash: ask
```

OpenCode Markdown output:

```yaml
---
description: Reviews code for correctness, risk, and verification evidence
mode: subagent
permission:
  edit: deny
  bash: ask
  read: allow
  grep: allow
---
```

Rules:

- `modelClass` maps to configured model only if user has target model mapping.
- `steps` maps to OpenCode `steps`; never output deprecated `maxSteps`.
- Tool permissions map to `permission`, not deprecated `tools`, unless compatibility mode.

### 6.4 Skill mapping

Output skill paths:

```text
.opencode/skills/<name>/SKILL.md
.agents/skills/<name>/SKILL.md    # optional compatibility
.claude/skills/<name>/SKILL.md    # optional Claude compatibility if multi-target compile
```

Validation:

- name regex: `^[a-z0-9]+(-[a-z0-9]+)*$`;
- name length 1–64;
- `description` length 1–1024;
- directory matches name.

### 6.5 MCP mapping

Option A: merge into `opencode.json`:

```json
{
  "mcp": {
    "ai-memory": {
      "type": "local",
      "command": ["ai-memory", "server"],
      "enabled": true,
      "environment": {
        "AI_MEMORY_DB": "${AI_MEMORY_DB}"
      }
    }
  }
}
```

Option B: generate `.opencode/lazyai.mcp.jsonc` and include/reference if OpenCode supports includes in current config.

### 6.6 Hook mapping

Canonical hooks that require event interception compile to `.opencode/plugins/vibe-lab-hooks.js`.

Plugin responsibilities:

- protected path guard;
- post-tool evidence capture;
- optional command denylist;
- summary/handoff prompt injection if supported.

---

## 7. Claude Code adapter spec

### 7.1 Outputs

```text
CLAUDE.md
AGENTS.md
.mcp.json
.claude/settings.json
.claude/settings.local.json
.claude/skills/<name>/SKILL.md
.claude/agents/<name>.md
.claude/hooks/<hook>.sh
.claude/plugins/lazyai-vibelab/
```

### 7.2 Root instructions

`CLAUDE.md` is mandatory for Claude Code adapter compliance.

Recommended generated pattern:

```markdown
# Project instructions for Claude Code

This file is generated by LazyAI from `.ai/`.

## Shared agent harness instructions

@include ./AGENTS.md

## Claude-specific notes

- Use configured skills for repeatable procedures.
- Respect permission prompts.
- Do not bypass human gates.
```

If `@include` is not supported in the current Claude Code version, LazyAI must inline the rendered shared content.

### 7.3 Skills

Output:

```text
.claude/skills/<name>/SKILL.md
```

Claude-specific frontmatter extensions should be opt-in and isolated in adapter metadata.

### 7.4 Subagents

Output:

```text
.claude/agents/reviewer.md
```

Minimum structure:

```yaml
---
name: reviewer
description: Reviews code without editing files.
tools: Read,Grep,Glob
model: inherit
---
```

Tool restrictions must mirror canonical permissions.

### 7.5 Hooks

Claude hook config must be generated in settings using official event names and hook schema.

Hook command script example:

```bash
#!/usr/bin/env bash
set -euo pipefail
payload="$(cat)"
lazyai hook run protected-paths --payload-json "$payload"
```

Hook scripts must:

- be executable;
- read JSON from stdin when appropriate;
- return supported exit codes;
- avoid destructive actions by default;
- log to safe local directory with secret redaction.

### 7.6 MCP

Project MCP output:

```text
.mcp.json
```

Settings local output may reference managed or local MCP configuration, but must not write secrets.

### 7.7 Managed settings awareness

`doctor` must detect and warn when managed settings may block:

- project/user skills;
- project/user agents;
- project/user hooks;
- project/user MCP servers;
- unmanaged plugin sources.

---

## 8. GitHub Copilot adapter spec

### 8.1 Outputs

```text
.github/copilot-instructions.md
.github/instructions/<scope>.instructions.md
.github/agents/<name>.agent.md
.github/prompts/<name>.prompt.md
.github/chatmodes/<name>.chatmode.md
.github/skills/<name>/SKILL.md
.agents/skills/<name>/SKILL.md
.vscode/mcp.json
```

Optional plugin:

```text
.copilot-plugin/lazyai-vibelab/
  plugin.json
  agents/<name>.agent.md
  skills/<name>/SKILL.md
  hooks.json
  .mcp.json
```

### 8.2 Instructions

Canonical root rules compile to `.github/copilot-instructions.md`.

Path-specific rules compile to:

```text
.github/instructions/rust.instructions.md
.github/instructions/frontend.instructions.md
.github/instructions/security.instructions.md
```

Each path-specific instruction file must include matching frontmatter according to current GitHub docs.

### 8.3 Skills

Canonical skills compile to at least one official project skill location.

Default:

```text
.github/skills/<name>/SKILL.md
```

Compatibility:

```text
.agents/skills/<name>/SKILL.md
```

### 8.4 Hooks

Canonical hooks compile to Copilot CLI hook files only when target profile includes Copilot CLI.

Hook limitations:

- not all IDE/cloud Copilot surfaces support hooks equally;
- generated hooks must be labeled CLI-specific;
- unsupported hooks must degrade to instructions or warnings.

### 8.5 MCP

Default IDE output:

```text
.vscode/mcp.json
```

Plugin output:

```text
.copilot-plugin/lazyai-vibelab/.mcp.json
```

Global user output only with explicit consent.

### 8.6 Plugin precedence

Validator must warn if plugin components conflict with project/personal components because Copilot plugin docs define precedence and dedup behavior.

---

## 9. Pi adapter spec

### 9.1 Outputs

```text
AGENTS.md
.pi/settings.json
.pi/SYSTEM.md
.pi/APPEND_SYSTEM.md
.pi/skills/<name>/SKILL.md
.pi/prompts/<name>.md
.pi/extensions/lazyai/<name>.ts
.agents/skills/<name>/SKILL.md
```

### 9.2 Trust-sensitive files

Pi treats project-local settings/resources/extensions as trust-sensitive. LazyAI must:

- warn when generating `.pi/settings.json`, `.pi/extensions`, `.pi/skills`, `.pi/prompts`, `.pi/SYSTEM.md`, `.pi/APPEND_SYSTEM.md`;
- explain that project trust does not sandbox tool behavior;
- provide `ai-jail` or container guidance.

### 9.3 Skills

Pi skill output must include:

```yaml
---
name: pr-review
description: Review a pull request, staged diff, or patch with evidence and risk analysis.
---
```

Pi allows leniency, but LazyAI strict mode should comply with shared Agent Skills expectations.

### 9.4 Prompt templates

Canonical prompt templates compile to:

```text
.pi/prompts/<name>.md
```

With frontmatter:

```yaml
---
description: Produce an RPI bugfix plan.
argument-hint: "<ticket-or-bug-description>"
---
```

### 9.5 Extensions

Canonical hook policies compile to TypeScript extensions only when required.

Example extension responsibilities:

- block unsafe shell commands;
- capture evidence after tool calls;
- path-protect generated files;
- custom compaction instructions.

---

## 10. OMP adapter spec

### 10.1 Outputs

```text
AGENTS.md
.omp/agents/<name>.md
.omp/skills/<name>/SKILL.md
.omp/commands/<name>.md
.omp/hooks/<name>.*
.omp/mcp.json
.omp/plugins/lazyai-vibelab/
```

### 10.2 MCP

Project output:

```text
.omp/mcp.json
```

User output only with explicit consent:

```text
~/.omp/agent/mcp.json
```

### 10.3 Plugins

Plugin bundle should be generated as a directory containing:

- skills;
- commands;
- hooks;
- custom tools;
- MCP server config;
- themes if selected.

### 10.4 Handoff/compaction

vibe-lab handoff and compaction assets should compile into OMP-native commands or skills:

```text
.omp/commands/handoff.md
.omp/skills/context-compaction/SKILL.md
```

### 10.5 Conformance caveat

Because some OMP docs pages are rendered with limited machine-readable text, adapter release must include manual official-doc snapshot review before stable status.

---

## 11. Antigravity / Gemini adapter spec

### 11.1 Outputs

```text
AGENTS.md
GEMINI.md
.agents/skills/<name>/SKILL.md
.agents/rules/<name>.md
.agents/hooks.json
.gemini/hooks/lazyai/<name>.sh
```

Global settings guidance only by default:

```text
~/.gemini/antigravity-cli/settings.json
```

### 11.2 Skills

Default skill location:

```text
.agents/skills/<name>/SKILL.md
```

Backward compatibility mode:

```text
.agent/skills/<name>/SKILL.md
```

### 11.3 Rules

Canonical rules compile to:

```text
.agents/rules/<rule>.md
```

Root context may compile to:

```text
AGENTS.md
GEMINI.md
```

### 11.4 Hooks

Canonical hooks compile to official hook JSON schema when available:

```json
{
  "hooks": {
    "protected-paths": {
      "events": ["preToolUse"],
      "command": ".gemini/hooks/lazyai/protected-paths.sh"
    }
  }
}
```

Actual event names must be validated against current Antigravity docs/runtime during adapter conformance.

### 11.5 Permissions and sandbox

- Permission changes in user-global settings require explicit consent.
- Sandbox settings require explicit consent and platform detection.
- `doctor` must report current sandbox/permissions assumptions.

---

## 12. Kiro adapter spec

### 12.1 Outputs

```text
AGENTS.md
.kiro/steering/product.md
.kiro/steering/tech.md
.kiro/steering/structure.md
.kiro/steering/vibe-lab.md
.kiro/steering/security.md
.kiro/specs/<name>/requirements.md
.kiro/specs/<name>/design.md
.kiro/specs/<name>/tasks.md
.kiro/hooks/<name>.json
.kiro/settings/mcp.json
.kiroignore
```

### 12.2 Steering mapping

Always-on canonical rules compile to:

```yaml
---
inclusion: always
---
```

Path-scoped rules compile to:

```yaml
---
inclusion: fileMatch
fileMatchPattern: "src/**/*.rs"
---
```

Manual rules compile to:

```yaml
---
inclusion: manual
---
```

Auto/semantic rules compile to:

```yaml
---
inclusion: auto
name: security-review
description: Security review guidance for auth, permissions, secrets, and network code.
---
```

### 12.3 Specs mapping

Canonical templates compile to:

```text
.kiro/specs/<feature>/requirements.md
.kiro/specs/<feature>/design.md
.kiro/specs/<feature>/tasks.md
```

Bugfix specs may use:

```text
.kiro/specs/<bug>/bugfix.md
.kiro/specs/<bug>/design.md
.kiro/specs/<bug>/tasks.md
```

### 12.4 Hooks

Canonical hooks compile to Kiro hook definitions with:

- title;
- description;
- event;
- optional tool name;
- optional file pattern;
- action type: ask Kiro or run command;
- instructions/command.

### 12.5 Security mapping

`doctor` must warn:

- Supervised mode is a review workflow, not a sandbox;
- Autopilot and supervised modes have the same underlying access scope;
- protected paths should be configured for sensitive files;
- trusted command wildcards should be narrow;
- credentials should be scoped to workspace needs.

---

## 13. Validation engine

### 13.1 Diagnostic model

```go
type Diagnostic struct {
    Severity    Severity // error, warning, info
    Code        string
    Message     string
    Path        string
    Target      string
    DocsRef     string
    Suggestion  string
}
```

Examples:

```text
ERROR SKILL_NAME_INVALID .ai/skills/PR Review/SKILL.md
Skill name must be lowercase kebab-case for OpenCode, Copilot, Pi, and shared Agent Skills compatibility.

WARNING CLAUDE_MISSING_CLAUDE_MD
Claude adapter requires CLAUDE.md. LazyAI will generate it from AGENTS.md unless disabled.

ERROR SECRET_INLINE_MCP .ai/mcp.json
MCP server github contains inline token. Use ${GITHUB_TOKEN} instead.

WARNING PI_TRUST_NOT_SANDBOX .pi/settings.json
Pi project trust controls loading local resources but does not sandbox tool execution.
```

### 13.2 Validators

Required validators:

- manifest validator;
- asset schema validator;
- skill validator;
- agent validator;
- hook validator;
- MCP validator;
- adapter validator;
- generated-file validator;
- official-doc conformance validator;
- security validator;
- migration validator.

### 13.3 Official-doc conformance fixtures

Each adapter has:

```text
testdata/docs-snapshots/<target>/<date>.md
testdata/golden/<target>/<case>/expected-tree.txt
testdata/golden/<target>/<case>/expected-files/*
```

Conformance tests should assert:

- required paths exist;
- required frontmatter fields exist;
- deprecated fields are not emitted;
- permissions use current schema;
- MCP config shape is valid;
- hooks use supported events;
- root instruction file matches official requirement.

---

## 14. Hook policy compiler

### 14.1 Canonical hook policy

```yaml
---
name: protected-paths
type: hook
description: Block agent edits to sensitive files unless user approves.
events:
  - before_tool
match:
  tools:
    - edit
    - write
    - apply_patch
paths:
  deny:
    - ".git/**"
    - ".env"
    - "**/*secret*"
action:
  type: require_approval
  message: "This path is protected. Confirm before modifying."
---
```

### 14.2 Adapter compilation

| Canonical event | Claude | OpenCode | Pi | Kiro | Copilot CLI | Antigravity | OMP |
|---|---|---|---|---|---|---|---|
| `before_tool` | `PreToolUse` | plugin event | extension `tool_call` intercept | pre tool hook | CLI hook trigger | hook JSON | hook/plugin |
| `after_tool` | `PostToolUse` | plugin event | extension event | post tool hook | CLI hook trigger | hook JSON | hook/plugin |
| `on_compaction` | `PreCompact`/`PostCompact` | native/agent | compaction extension | no-op/instruction | no-op/instruction | no-op/instruction | compaction/handoff |
| `on_file_change` | async/file event | plugin if supported | extension | file save/create/delete | hook if supported | hook JSON | hook/plugin |

If target cannot support a hook, compiler must:

1. emit warning;
2. degrade to instruction/rule if useful;
3. never claim enforcement.

---

## 15. MCP compiler

### 15.1 Secret handling

Supported value types:

```json
{
  "env": {
    "TOKEN": "${TOKEN}",
    "OPTIONAL": "${OPTIONAL:-default}"
  }
}
```

Rejected by default:

```json
{
  "env": {
    "TOKEN": "ghp_actualtoken"
  }
}
```

### 15.2 Target-specific mapping

| Target | Output |
|---|---|
| OpenCode | `opencode.json` `mcp` section or `.opencode/lazyai.mcp.jsonc` include strategy |
| Claude Code | `.mcp.json`, `.claude/settings.local.json` where needed |
| Copilot | `.vscode/mcp.json`, plugin `.mcp.json`, optional global |
| Pi | settings/extension strategy if MCP supported directly; otherwise guidance or package output |
| OMP | `.omp/mcp.json` |
| Kiro | `.kiro/settings/mcp.json` |
| Antigravity | Antigravity/Gemini CLI MCP settings path according to current docs |

### 15.3 MCP server inventory report

`lazyai doctor mcp` output:

```text
MCP servers:
  ai-memory   local   tools=6   targets=claude,opencode,copilot,kiro,omp
  filesystem  local   tools=4   targets=claude,opencode,copilot,kiro,omp
  codegraph   local   tools=3   targets=claude,opencode,copilot
Warnings:
  - filesystem has write_file enabled; reviewer agent denies write tools.
  - obsidian requires OBSIDIAN_VAULT env var.
```

---

## 16. Migration engine

### 16.1 Migration flow

```text
scan native files
classify ownership
parse supported formats
extract canonical candidates
detect conflicts
write .ai/migration-report.md
write canonical assets
compile dry-run
```

### 16.2 Confidence levels

```text
exact       direct canonical mapping
high        mapping likely complete, minor target metadata lost
medium      instructions imported but semantics uncertain
low         copied as raw target-specific asset
unsupported left in native file only
```

### 16.3 Preservation rules

- Never delete native fields that LazyAI cannot represent.
- Store adapter-specific overrides under `.ai/adapters/<target>/raw/` if needed.
- Keep original files unless user confirms replacement.
- Mark generated regions only after successful compile.

---

## 17. Runtime-adjacent extras

Runtime extras are optional local features.

### 17.1 Session

Purpose:

- record local work sessions;
- capture handoff summaries;
- optionally import host session summaries.

Storage:

```text
.ai/runtime/sessions/*.jsonl
```

### 17.2 Ledger

Purpose:

- record decisions, commands, evidence, and approvals.

Storage:

```text
.ai/runtime/ledger.sqlite
```

### 17.3 Memory

Purpose:

- local project memory and durable annotations;
- optional MCP server integration.

Must remain opt-in.

### 17.4 Metrics/cost

Purpose:

- local estimates and operational awareness;
- no provider scraping unless user configures it;
- no cloud telemetry by default.

---

## 18. Testing strategy

### 18.1 Unit tests

- schema parsing;
- frontmatter parsing;
- permission mapping;
- MCP mapping;
- hook translation;
- safe writer;
- secret scanner;
- path resolver.

### 18.2 Golden tests

For each adapter:

```text
testdata/golden/<adapter>/starter/
  input/.ai/...
  expected/...
```

Golden tests must compare:

- file tree;
- generated content;
- lockfile;
- diagnostics.

### 18.3 Integration tests

Run in temporary repos:

```bash
lazyai init --targets opencode,claude,copilot,pi,omp,antigravity,kiro
lazyai validate
lazyai compile --all
lazyai doctor
```

Smoke tests may attempt host command detection but must not require paid accounts.

### 18.4 Security tests

- inline secret detection;
- path traversal;
- symlink attacks;
- unsafe hook commands;
- global config write consent;
- file overwrite behavior;
- generated file drift.

### 18.5 Official docs conformance tests

Each adapter must have:

- docs source registry entry;
- conformance checklist;
- last verification date;
- stable/beta/experimental status;
- failing test when an adapter emits deprecated or unsupported fields.

---

## 19. CLI command technical contracts

### 19.1 `init`

```bash
lazyai init [--targets ...] [--profile team|personal|ci] [--pack vibe-lab/starter] [--yes]
```

Contract:

- idempotent;
- safe by default;
- creates canonical source;
- optionally compiles.

### 19.2 `compile`

```bash
lazyai compile [--target t] [--all] [--dry-run] [--diff] [--write]
```

Contract:

- does not execute host tools;
- writes only planned files;
- updates lockfile.

### 19.3 `validate`

```bash
lazyai validate [skills|agents|hooks|mcp|adapters|generated|official-docs]
```

Contract:

- exits non-zero on errors;
- CI-friendly JSON output via `--json`.

### 19.4 `doctor`

```bash
lazyai doctor [--target t] [--security] [--mcp] [--json]
```

Contract:

- detects host tools;
- checks generated files;
- warns about sandbox/trust/permissions.

### 19.5 `migrate`

```bash
lazyai migrate --from existing [--targets ...] [--dry-run]
```

Contract:

- reads native files;
- creates canonical candidates;
- never deletes originals without confirmation.

### 19.6 `eject`

```bash
lazyai eject [--keep-native] [--remove-generated-headers] [--remove-ai-dir]
```

Contract:

- preserves host tool usability;
- writes report.

---

## 20. Backward compatibility

### 20.1 Schema versioning

- `.ai/lazyai.json` contains schema version.
- Migrations are explicit and reviewable.
- Breaking changes require migration command.

### 20.2 Adapter versioning

- Adapter versions correspond to official docs verification dates.
- Tool-specific breaking changes are isolated to adapter package.
- `doctor` warns when local host tool version is unknown or incompatible.

### 20.3 Generated file ownership

A file can be:

```text
managed       fully generated by LazyAI
region        generated region inside user file
adopted       existing file adopted after migration
native        user-owned, not managed
ignored       intentionally unmanaged
```

---

## 21. Performance requirements

- `lazyai validate` on a typical repo should complete in under 2 seconds excluding host tool detection.
- `lazyai compile --dry-run` should complete in under 3 seconds for 100 assets.
- File writes should be batched and atomic.
- Large asset libraries should be lazily loaded.
- No network calls during compile unless explicitly requested.

---

## 22. Error handling

Errors must include:

- severity;
- code;
- file path;
- target adapter;
- explanation;
- fix suggestion;
- official docs reference when relevant.

Example:

```text
ERROR CLAUDE001: Missing CLAUDE.md output for Claude Code target.
Path: .ai/lazyai.json
Fix: enable adapters.claude.emitClaudeMd or remove claude target.
Docs: Claude Code overview / .claude directory docs.
```

---

## 23. Stable v1.0 technical acceptance criteria

- Canonical schema finalized.
- All adapters have golden tests.
- OpenCode, Claude, Copilot, Pi, Kiro, Antigravity, and OMP adapters compile valid starter outputs.
- Skill/agent/hook/MCP validators are complete.
- Safe writer prevents destructive overwrites.
- Secret scanner blocks inline secrets by default.
- Official docs conformance suite exists.
- Migration and eject flows are tested.
- Optional runtime extras are isolated.
- Product documentation clearly separates LazyAI, vibe-lab, and host tool responsibilities.
