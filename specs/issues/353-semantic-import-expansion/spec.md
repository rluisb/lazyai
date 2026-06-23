# Spec: 353-semantic-import-expansion

**Issue:** #353
**Date:** 2026-06-23
**Status:** Planning
**Owner:** maintainers

> **Purpose.** Research and plan semantic import expansion beyond OpenCode. Inventory native input formats for all 6 non-OpenCode targets, classify lossless mappings, define confidence-label rules, choose one first target, and produce follow-up implementation issues. Do not implement parsers.

---

## 1. Current State

### 1.1 Import pipeline (ground truth)

The import pipeline lives in `packages/cli/internal/migration/` and `packages/cli/cmd/import_v2.go`:

1. **Detection** (`detector.go`): scans for known file/directory patterns per adapter. All 7 targets detected.
2. **Canonical parsing** (`parser.go`): only OpenCode has a parser (`parseOpenCodeSetup`). All other targets return `"unsupported detected setup"`.
3. **Canonical write** (`canonicalwriter.go`): writes parsed agents/skills/prompts/rules into `.ai/`.
4. **Raw preservation** (`import_v2.go:planRawPreservation`): every detected file is copied to `.ai/adapters/<target>/raw/`.
5. **Confidence labels** (`import_v2.go:migrationConfidenceLabel`): OpenCode = `high`, Claude Code = `medium` (CLAUDE.md only) or `low`, all others = `low`.

### 1.2 Detection patterns (detector.go)

| Adapter | Detection patterns |
|---|---|
| `opencode` | `.opencode`, `opencode.json`, `AGENTS.md` |
| `claude-code` | `.claude`, `CLAUDE.md` |
| `copilot` | `.github/copilot-instructions.md`, `.github/instructions`, `.github/skills` |
| `pi` | `.pi` |
| `omp` | `.omp` |
| `kiro` | `.kiro` |
| `antigravity` | `.agents` |

### 1.3 Current confidence labels

| Adapter | Label | Rationale |
|---|---|---|
| `opencode` | `high` | Full semantic parser exists |
| `claude-code` | `medium` (CLAUDE.md only) / `low` | Only root doc parsed; no `.claude/` structure |
| `copilot` | `low` | Raw preservation only |
| `pi` | `low` | Raw preservation only |
| `omp` | `low` | Raw preservation only |
| `kiro` | `low` | Raw preservation only |
| `antigravity` | `low` | Raw preservation only |

---

## 2. Native Format Inventory

For each target, we inventory what files the detector finds, what canonical `.ai/` surfaces they could map to, and what loss is expected.

### 2.1 Claude Code

**Detection:** `.claude/` directory, `CLAUDE.md` root file.

**Native structure:**
- `CLAUDE.md` — root instructions (markdown, may include `@AGENTS.md` import)
- `.claude/agents/<name>.md` — agent definitions with YAML frontmatter
- `.claude/skills/<name>/SKILL.md` — skill definitions with YAML frontmatter
- `.claude/commands/<name>.md` — slash commands
- `.claude/output-styles/<name>.md` — output style definitions
- `.claude/hooks/*` — hook scripts (`.sh`)
- `.claude/settings.json` — project settings (permissions, MCP)
- `.mcp.json` — MCP server config (project root)
- `.claude/settings.local.json` — local secrets MCP config

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `CLAUDE.md` | `.ai/agents/guide.md` or root rules | Low | Markdown content maps directly; `@AGENTS.md` import directive is Claude-specific |
| `.claude/agents/*.md` | `.ai/agents/<name>.md` | Low | Same frontmatter schema (name, description); Claude has `model` field not in canonical |
| `.claude/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format; canonical already uses this shape |
| `.claude/commands/*.md` | `.ai/skills/<name>.md` | Medium | Commands are a different concept from skills; canonical has no separate commands dir |
| `.claude/output-styles/*.md` | `.ai/prompts/` or custom | Medium | No canonical output-styles surface; could map to prompts |
| `.claude/hooks/*` | `.ai/hooks/` | Low | Hook scripts map directly to canonical hooks |
| `.claude/settings.json` | `.ai/housekeeping/` | Medium | Settings contain permissions, MCP refs; partial mapping |
| `.mcp.json` | `.ai/mcp.json` | Low | MCP server definitions map directly |

**Key format details:**
- Agent frontmatter: `name`, `description`, optional `model`, `permissions`, `groups`
- Skill frontmatter: `name`, `description`, optional `model`, `trigger`, `temperature`
- Commands: markdown with `name` and `description` frontmatter
- All files are markdown with YAML frontmatter (same as canonical)

### 2.2 GitHub Copilot

**Detection:** `.github/copilot-instructions.md`, `.github/instructions/`, `.github/skills/`

**Native structure:**
- `.github/copilot-instructions.md` — root instructions
- `.github/instructions/*.instructions.md` — path-scoped instructions
- `.github/agents/*.agent.md` — custom agent definitions
- `.github/skills/<name>/SKILL.md` — Agent Skills
- `.github/prompts/*.prompt.md` — prompt templates
- `.github/chatmodes/*.chatmode.md` — chat mode definitions
- `.github/hooks/*.json` + `*.sh` — hook configs and scripts
- `.vscode/mcp.json` — MCP server config (IDE)
- `~/.copilot/mcp-config.json` — MCP server config (CLI)

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `.github/copilot-instructions.md` | `.ai/agents/guide.md` or root rules | Low | Markdown instructions map directly |
| `.github/instructions/*.instructions.md` | `.ai/rules/` | Low | Path-scoped rules map to canonical rules |
| `.github/agents/*.agent.md` | `.ai/agents/<name>.md` | Medium | `.agent.md` has Copilot-specific frontmatter (`model`, `tier`, `temperature`) |
| `.github/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format |
| `.github/prompts/*.prompt.md` | `.ai/prompts/<name>.md` | Low | Prompt templates map directly |
| `.github/chatmodes/*.chatmode.md` | `.ai/prompts/` or custom | Medium | No canonical chatmodes surface |
| `.github/hooks/*` | `.ai/hooks/` | Low | Hook configs map to canonical hooks |
| `.vscode/mcp.json` | `.ai/mcp.json` | Low | MCP server definitions map directly |

**Key format details:**
- Agent `.agent.md` frontmatter: `name`, `description`, `model` (with tier resolution), `instructions`
- Instructions `.instructions.md` frontmatter: `appliesTo` (path globs), `description`
- Prompt `.prompt.md` frontmatter: `name`, `description`
- Chatmode `.chatmode.md` frontmatter: `name`, `description`, `instructions`
- All files are markdown with YAML frontmatter

### 2.3 Pi

**Detection:** `.pi/` directory

**Native structure:**
- `.pi/settings.json` — project settings
- `.pi/skills/<name>/SKILL.md` — skill definitions
- `.pi/prompts/<name>.md` — prompt templates
- `.pi/agents/<name>.md` — subagent definitions
- `.pi/extensions/*.ts` — TypeScript extensions (safety hooks)
- `.pi/SYSTEM.md` / `.pi/APPEND_SYSTEM.md` — system prompt overrides
- `AGENTS.md` — root instructions (shared with OpenCode)

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `AGENTS.md` | `.ai/agents/guide.md` | Low | Same format as OpenCode root |
| `.pi/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format |
| `.pi/prompts/*.md` | `.ai/prompts/<name>.md` | Low | Prompt templates map directly |
| `.pi/agents/*.md` | `.ai/agents/<name>.md` | Low | Subagent definitions are markdown with frontmatter |
| `.pi/extensions/*.ts` | `.ai/hooks/` | Medium | TypeScript extensions are Pi-specific; canonical hooks are shell scripts |
| `.pi/SYSTEM.md` | `.ai/constitution/` | Low | System prompt overrides map to constitution |
| `.pi/settings.json` | `.ai/housekeeping/` | Medium | Settings contain trust config, MCP refs |

**Key format details:**
- Skills: same Agent Skills format (SKILL.md with YAML frontmatter)
- Agents: markdown with optional frontmatter (name, description)
- Prompts: markdown with optional frontmatter
- Extensions: TypeScript files, not markdown

### 2.4 OMP

**Detection:** `.omp/` directory

**Native structure:**
- `AGENTS.md` — root instructions (shared with OpenCode)
- `.omp/agents/<name>.md` — task agent definitions
- `.omp/skills/<name>/SKILL.md` — skill definitions
- `.omp/commands/<name>.md` — slash commands
- `.omp/prompts/<name>.md` — prompt templates
- `.omp/hooks/pre/*.ts` — TypeScript hook factories
- `.omp/mcp.json` — MCP server config

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `AGENTS.md` | `.ai/agents/guide.md` | Low | Same format as OpenCode root |
| `.omp/agents/*.md` | `.ai/agents/<name>.md` | Low | Task agents are markdown with frontmatter |
| `.omp/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format |
| `.omp/commands/*.md` | `.ai/skills/<name>.md` | Medium | Commands are a different concept; canonical has no separate commands dir |
| `.omp/prompts/*.md` | `.ai/prompts/<name>.md` | Low | Prompt templates map directly |
| `.omp/hooks/pre/*.ts` | `.ai/hooks/` | Medium | TypeScript hooks vs canonical shell hooks |
| `.omp/mcp.json` | `.ai/mcp.json` | Low | MCP server definitions map directly |

**Key format details:**
- Agents: markdown with YAML frontmatter (name, description, model, tools)
- Skills: same Agent Skills format
- Commands: markdown with frontmatter (name, description)
- Hooks: TypeScript files, not shell scripts

### 2.5 Kiro

**Detection:** `.kiro/` directory

**Native structure:**
- `.kiro/agents/<name>.md` — custom agent profiles
- `.kiro/skills/<name>/SKILL.md` — skill definitions
- `.kiro/prompts/<name>.md` — prompt templates
- `.kiro/steering/*.md` — steering instructions (not yet emitted by LazyAI)
- `.kiro/settings/mcp.json` — MCP server config
- `.kiro/specs/` — spec files (requirements, design, tasks)
- `.kiro/hooks/*` — hook definitions
- `.kiroignore` — path exclusion file

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `.kiro/agents/*.md` | `.ai/agents/<name>.md` | Low | Agent profiles are markdown with frontmatter |
| `.kiro/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format |
| `.kiro/prompts/*.md` | `.ai/prompts/<name>.md` | Low | Prompt templates map directly |
| `.kiro/steering/*.md` | `.ai/rules/` | Low | Steering instructions are rules |
| `.kiro/settings/mcp.json` | `.ai/mcp.json` | Low | MCP server definitions map directly |

**Key format details:**
- Agents: markdown with YAML frontmatter (name, description, model)
- Skills: same Agent Skills format
- Prompts: markdown with optional frontmatter
- Steering: markdown with frontmatter (inclusion mode, fileMatch)

### 2.6 Antigravity

**Detection:** `.agents/` directory

**Native structure:**
- `.agents/skills/<name>/SKILL.md` — skill definitions
- `.agents/rules/*.md` — rule definitions (not yet emitted by LazyAI)
- `.agents/hooks.json` — hook event mappings
- `.gemini/settings.json` — settings
- `.gemini/hooks/lazyai/*.sh` — hook scripts
- `~/.gemini/config/mcp_config.json` — MCP server config
- `AGENTS.md` — root instructions (shared with OpenCode)

**Canonical mapping:**

| Native file | Canonical target | Lossiness | Notes |
|---|---|---|---|
| `AGENTS.md` | `.ai/agents/guide.md` | Low | Same format as OpenCode root |
| `.agents/skills/*/SKILL.md` | `.ai/skills/<name>.md` | Low | Same Agent Skills format |
| `.agents/rules/*.md` | `.ai/rules/` | Low | Rules map directly |
| `.agents/hooks.json` | `.ai/hooks/` | Low | Hook event mappings map to canonical hooks |
| `.gemini/settings.json` | `.ai/housekeeping/` | Medium | Settings contain permissions, sandbox config |
| `.gemini/hooks/lazyai/*.sh` | `.ai/hooks/` | Low | Hook scripts map directly |
| MCP config | `.ai/mcp.json` | Low | MCP server definitions map directly |

**Key format details:**
- Skills: same Agent Skills format
- Rules: markdown with optional frontmatter
- Hooks: JSON event mapping + shell scripts
- Settings: JSON

---

## 3. Lossless Mapping Classification

### 3.1 Mapping categories

| Category | Definition | Examples |
|---|---|---|
| **Exact** | Native format is identical to canonical format | Agent Skills SKILL.md, MCP server definitions |
| **Low-loss** | Content maps directly; minor metadata or target-specific fields dropped | Agent markdown files, root instructions, prompt templates |
| **Medium-loss** | Concept exists but canonical surface differs in structure or semantics | Commands vs skills, output styles, chat modes |
| **High-loss** | Concept has no canonical equivalent; only raw preservation possible | Pi TypeScript extensions, Kiro steering inclusion modes, Claude `model` field in agent frontmatter |
| **Unsupported** | File type or concept LazyAI cannot represent at all | Binary assets, IDE-specific configs |

### 3.2 Per-target lossless surface summary

| Target | Exact | Low-loss | Medium-loss | High-loss | Unsupported |
|---|---|---|---|---|---|
| **Claude Code** | Skills, MCP | Agents, Root instructions, Hooks, Prompts | Commands, Output styles, Settings | Model field, Groups | — |
| **Copilot** | Skills, MCP | Root instructions, Path instructions, Prompts | Agents (`.agent.md` format), Chat modes | Tier/risk annotations | — |
| **Pi** | Skills | Agents, Prompts, Root instructions | Extensions (TS vs shell), Settings | — | — |
| **OMP** | Skills, MCP | Agents, Prompts, Root instructions | Commands, Hooks (TS vs shell) | — | — |
| **Kiro** | Skills, MCP | Agents, Prompts, Steering | — | Specs (3-phase), Hooks | — |
| **Antigravity** | Skills, MCP | Rules, Hooks, Root instructions | Settings | — | — |

### 3.3 Key insight: all targets share the Agent Skills format

Every target uses the same `SKILL.md` with YAML frontmatter format for skills. This is the **highest-value, lowest-risk** surface to map first — the format is identical, the content is lossless, and the canonical `.ai/skills/` directory already exists.

---

## 4. Confidence-Label Rules

### 4.1 Proposed label schema

| Label | Meaning | When applied |
|---|---|---|
| `exact` | Direct canonical mapping, no data loss | Skills (SKILL.md), MCP configs |
| `high` | Mapping likely complete, minor target metadata lost | Agents, root instructions, prompts, rules, hooks |
| `medium` | Instructions imported but semantics uncertain | Commands, output styles, chat modes, settings |
| `low` | Copied as raw target-specific asset | Any file without a semantic parser |
| `unsupported` | Left in native file only | Binary files, IDE-specific configs |

### 4.2 Per-target label rules (proposed)

For each target, the label is the **minimum** of all detected file labels. If any file is `low`, the whole target is `low` unless all files have a parser.

**Claude Code:**
- If only `CLAUDE.md` detected → `medium` (current behavior)
- If `.claude/agents/` or `.claude/skills/` detected → `high` (with parser)
- If `.claude/commands/` or `.claude/output-styles/` detected → `medium` (partial parser)
- Fallback: `low`

**Copilot:**
- If `.github/skills/` or `.github/prompts/` detected → `high` (with parser for those surfaces)
- If `.github/agents/*.agent.md` detected → `high` (with agent parser)
- If `.github/chatmodes/` detected → `medium`
- Fallback: `low`

**Pi:**
- If `.pi/skills/` or `.pi/prompts/` or `.pi/agents/` detected → `high` (with parser)
- If `.pi/extensions/` detected → `medium` (TS vs shell mapping)
- Fallback: `low`

**OMP:**
- If `.omp/skills/` or `.omp/prompts/` or `.omp/agents/` detected → `high` (with parser)
- If `.omp/commands/` detected → `medium`
- If `.omp/hooks/` detected → `medium`
- Fallback: `low`

**Kiro:**
- If `.kiro/skills/` or `.kiro/prompts/` or `.kiro/agents/` detected → `high` (with parser)
- If `.kiro/steering/` detected → `high` (rules mapping)
- Fallback: `low`

**Antigravity:**
- If `.agents/skills/` or `.agents/rules/` detected → `high` (with parser)
- If `.agents/hooks.json` detected → `high`
- Fallback: `low`

### 4.3 Label override rules

1. **Raw preservation is non-negotiable.** Even with a semantic parser, every detected file is always copied to `.ai/adapters/<target>/raw/`.
2. **No label upgrade without a parser.** A target cannot receive `high` or `exact` unless a real `parse<target>Setup` function exists in `parser.go`.
3. **Partial parser = partial label.** If only skills are parsed but commands are not, the label is `medium` (not `high`), reflecting the gap.
4. **Parser must preserve unsupported fields.** Any parser must store unrecognized frontmatter fields in the canonical output's metadata or custom sections, never silently drop them.

---

## 5. First Target Recommendation

### 5.1 Evaluation criteria

| Criterion | Weight | Rationale |
|---|---|---|
| **Value** (user impact) | High | How many users have this setup? How much friction does raw-only import cause? |
| **Format stability** | High | Is the native format documented and stable? Does it change frequently? |
| **Preservation risk** | High | How much data is lost without a parser? Can we verify round-trip fidelity? |
| **Parser complexity** | Medium | How many surfaces need parsers? How different are they from OpenCode's parser? |
| **Surface overlap with OpenCode** | Medium | Can we reuse OpenCode's parser patterns? |

### 5.2 Target comparison

| Target | Value | Format stability | Preservation risk | Parser complexity | Overlap | Score |
|---|---|---|---|---|---|---|
| **Claude Code** | High | High (stable, documented) | High (CLAUDE.md + agents + skills + commands + hooks) | Low (same markdown+frontmatter as OpenCode) | High | **Strongest** |
| Copilot | High | Medium (multiple surfaces, IDE/CLI/cloud differences) | Medium (instructions + skills + prompts are low-loss) | Medium (`.agent.md` format differs) | Medium | Strong |
| Pi | Medium | High (stable) | Medium (skills + prompts + agents are low-loss) | Low (same formats) | High | Moderate |
| OMP | Medium | Medium (beta, docs not fully captured) | Medium | Low (same formats) | High | Moderate |
| Kiro | Low-Medium | Medium (steering/specs not yet emitted) | Low-Medium (skills + agents + prompts are low-loss) | Low (same formats) | High | Lower |
| Antigravity | Low-Medium | Medium (beta, docs not fully captured) | Low-Medium (skills + rules are low-loss) | Low (same formats) | High | Lower |

### 5.3 Recommendation: Claude Code first

**Claude Code** is the clear first target for semantic import expansion:

1. **Highest value.** Claude Code is the most widely used AI coding tool alongside OpenCode. Users migrating from Claude Code to LazyAI currently get only `medium`/`low` confidence and raw preservation — a poor experience for a tool with rich, well-structured native config.

2. **Format stability.** Claude Code's native format (`.claude/agents/*.md`, `.claude/skills/*/SKILL.md`, `CLAUDE.md`) is stable, well-documented, and has not changed significantly in months. The compliance matrix (§3) is fully verified.

3. **Lowest preservation risk.** The native format is markdown with YAML frontmatter — identical to OpenCode's format. The OpenCode parser (`parseOpenCodeSetup`) can be adapted directly. Skills are already the same Agent Skills format. Agents use the same `name`/`description` frontmatter.

4. **Lowest parser complexity.** The existing `parseOpenCodeSetup` function provides a template. Claude Code's directory structure mirrors OpenCode's: `agents/*.md`, `skills/*/SKILL.md`, `commands/*.md`. The parser can reuse the same `parseOpenCodeMarkdownDir` helper.

5. **Surface overlap.** Claude Code shares agents, skills, commands, and MCP with OpenCode. Only output styles and hooks are Claude-specific additions.

6. **Current gap is visible.** The migration report already shows `"Canonical extraction: no (raw preserved under .ai/adapters/claude-code/raw/)"` for Claude Code — users see the gap directly.

### 5.4 What Claude Code import would cover

**Phase 1 (core, high confidence):**
- `CLAUDE.md` → `.ai/agents/guide.md` (root instructions)
- `.claude/agents/*.md` → `.ai/agents/<name>.md` (agent definitions)
- `.claude/skills/*/SKILL.md` → `.ai/skills/<name>.md` (skill definitions)

**Phase 2 (medium confidence):**
- `.claude/commands/*.md` → `.ai/skills/<name>.md` (commands as skills)
- `.claude/hooks/*` → `.ai/hooks/` (hook scripts)
- `.mcp.json` → `.ai/mcp.json` (MCP server definitions)

**Phase 3 (lower confidence / deferred):**
- `.claude/output-styles/*.md` → `.ai/prompts/` or custom surface
- `.claude/settings.json` → `.ai/housekeeping/` (partial)

---

## 6. Raw Preservation Guarantee

**Non-negotiable:** Every detected native file is always copied to `.ai/adapters/<target>/raw/`, regardless of whether a semantic parser exists. This ensures:

1. **No data loss.** Even with a perfect parser, the original files are preserved for audit and recovery.
2. **Graceful degradation.** If a parser misses a field or a future Claude Code format change breaks the parser, the raw copy is the fallback.
3. **User confidence.** Users can inspect the raw directory to verify nothing was lost.

The raw preservation code in `planRawPreservation` (import_v2.go:28-56) already implements this correctly. No changes needed.

---

## 7. Follow-up Implementation Issues

### Issue 1: Claude Code semantic import parser

**Title:** Implement Claude Code semantic import parser
**Parent:** #353
**Priority:** P3-medium

**Acceptance criteria:**
- [ ] `parseClaudeCodeSetup` function in `parser.go` that reads:
  - `CLAUDE.md` → root instructions
  - `.claude/agents/*.md` → agent definitions
  - `.claude/skills/*/SKILL.md` → skill definitions
- [ ] Reuses `parseOpenCodeMarkdownDir` or equivalent helper for agents and skills
- [ ] Preserves unrecognized frontmatter fields in canonical output metadata
- [ ] Confidence label for Claude Code upgrades to `high` when agents or skills are detected
- [ ] Raw preservation remains active for all detected files
- [ ] Migration report shows `"Canonical extraction: yes"` for Claude Code
- [ ] Test fixtures: a minimal `.claude/` tree with agents, skills, and CLAUDE.md
- [ ] Round-trip test: import → compile → verify native output matches original structure

**Non-goals:**
- Do not implement output-styles or settings.json parsing
- Do not modify native files during import
- Do not upgrade Claude Code confidence to `exact` (reserved for skills-only imports)

### Issue 2: Claude Code import — commands and hooks (Phase 2)

**Title:** Extend Claude Code import to commands, hooks, and MCP
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `.claude/commands/*.md` parsed and mapped to `.ai/skills/` with a `command` metadata tag
- [ ] `.claude/hooks/*` parsed and mapped to `.ai/hooks/`
- [ ] `.mcp.json` parsed and mapped to `.ai/mcp.json`
- [ ] Confidence label for Claude Code remains `high` (commands are medium-loss)
- [ ] Raw preservation remains active

### Issue 3: Copilot semantic import parser

**Title:** Implement Copilot semantic import parser
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `parseCopilotSetup` function that reads:
  - `.github/copilot-instructions.md` → root instructions
  - `.github/instructions/*.instructions.md` → rules
  - `.github/skills/*/SKILL.md` → skills
  - `.github/prompts/*.prompt.md` → prompts
- [ ] `.github/agents/*.agent.md` parsed with Copilot-specific frontmatter preserved
- [ ] Confidence label upgrades to `high` when skills or prompts are detected
- [ ] Raw preservation remains active

### Issue 4: Pi semantic import parser

**Title:** Implement Pi semantic import parser
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `parsePiSetup` function that reads:
  - `.pi/skills/*/SKILL.md` → skills
  - `.pi/prompts/*.md` → prompts
  - `.pi/agents/*.md` → agents
  - `AGENTS.md` → root instructions
- [ ] `.pi/extensions/*.ts` preserved as raw (medium-loss, no canonical TS surface)
- [ ] Confidence label upgrades to `high` when skills or prompts are detected
- [ ] Raw preservation remains active

### Issue 5: OMP semantic import parser

**Title:** Implement OMP semantic import parser
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `parseOmpSetup` function that reads:
  - `.omp/skills/*/SKILL.md` → skills
  - `.omp/prompts/*.md` → prompts
  - `.omp/agents/*.md` → agents
  - `.omp/commands/*.md` → skills (with command metadata)
  - `AGENTS.md` → root instructions
- [ ] `.omp/hooks/pre/*.ts` preserved as raw (medium-loss)
- [ ] Confidence label upgrades to `high` when skills or prompts are detected
- [ ] Raw preservation remains active

### Issue 6: Kiro semantic import parser

**Title:** Implement Kiro semantic import parser
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `parseKiroSetup` function that reads:
  - `.kiro/skills/*/SKILL.md` → skills
  - `.kiro/prompts/*.md` → prompts
  - `.kiro/agents/*.md` → agents
  - `.kiro/steering/*.md` → rules
- [ ] Confidence label upgrades to `high` when skills or prompts are detected
- [ ] Raw preservation remains active

### Issue 7: Antigravity semantic import parser

**Title:** Implement Antigravity semantic import parser
**Parent:** #353
**Priority:** P3-low

**Acceptance criteria:**
- [ ] `parseAntigravitySetup` function that reads:
  - `.agents/skills/*/SKILL.md` → skills
  - `.agents/rules/*.md` → rules
  - `.agents/hooks.json` → hooks
  - `AGENTS.md` → root instructions
- [ ] Confidence label upgrades to `high` when skills or rules are detected
- [ ] Raw preservation remains active

---

## 8. Implementation Order

```
Phase 1: Claude Code (highest value, lowest risk)
  └─ Issue 1: Core parser (agents, skills, root instructions)
  └─ Issue 2: Extended parser (commands, hooks, MCP)

Phase 2: Copilot (high value, medium complexity)
  └─ Issue 3: Full parser

Phase 3: Pi, OMP, Kiro, Antigravity (lower value, low complexity)
  └─ Issues 4-7: Per-target parsers
```

Each phase is independent. Phase 1 is the recommended first implementation.

---

## 9. Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Claude Code format changes | Low | Medium | Raw preservation is the fallback; parser tests detect drift |
| Parser silently drops fields | Medium | Medium | Store unrecognized frontmatter in canonical metadata; test with fixture containing unknown fields |
| False confidence from partial parser | Medium | Medium | Label rules require all surfaces to be parsed for `high`; partial parser = `medium` |
| Round-trip fidelity gaps | Medium | Medium | Round-trip tests: import → compile → compare with original |
| Copilot format varies by surface (IDE vs CLI) | Medium | Low | Parse only project-local files; CLI/global configs remain raw |

---

## 10. Out of Scope

- Implementing any parser (this is a planning issue)
- Modifying native files during import
- Deleting or renaming detected files
- Adding new CLI flags or commands
- Changing the compile pipeline
- Adding runtime dependencies
- Supporting Codex or Gemini-extension targets

---

## 11. Downstream Contract

| Produces for | Filename |
|---|---|
| Implementation issues | This spec → follow-up issues #354–#360 (to be created) |
| Parser implementation | `packages/cli/internal/migration/parser.go` (new parse functions) |
| Confidence label update | `packages/cli/cmd/import_v2.go:migrationConfidenceLabel` |
| Migration report update | `packages/cli/cmd/import_v2.go:writeMigrationReport` |
| Test fixtures | `packages/cli/testdata/` (per-target fixture trees) |
