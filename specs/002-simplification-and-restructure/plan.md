# 002 — Simplification & Restructure: Plan

## Date
2026-04-06

## Context
ai-setup is a CLI tool that scaffolds AI development environments for multiple tools (Claude Code, OpenCode, Copilot, Gemini, Pi, Codex). A multi-perspective review revealed significant overengineering: too many files generated (~73), too many concepts (24), too many uncollected placeholders (29 per file), and content duplication across agents, skills, prompts, and templates.

The architecture is sound (canonical source → compile model, adapter pattern, migration engine), but the product needs aggressive simplification to be adoptable by small teams (5-6 engineers).

## Goals
1. Reduce generated file count from ~73 to ~45 (Standard preset, 2 tools)
2. Eliminate `[YOUR_*]` placeholder burden through auto-detection and wizard questions
3. Make prompt engineering fragments actually effective for mid-tier/budget models
4. Support meaningful workspace scope (multi-repo coordination with knowledge persistence)
5. Maintain backward compatibility — never delete user content, safe re-runs

## Approach

### Wave 1: Foundation

**Goal**: Unify compilation paths, rewrite fragments, implement presets, rename to specs/.

#### 1.1 — Unify templates into single shared + overrides

Current: 6 tool templates (318 lines, nearly identical) + 4 non-compiled templates (1,346 lines).
After: 1 shared template (~53 lines) + 6 override files (~3-5 lines each). Non-compiled path removed.

Files changed:
- Delete: `library/tool-templates/{opencode,claude-code,copilot,gemini,pi,codex}/root.template.md`
- Delete: `library/root/AGENTS.template.md`, `CLAUDE.template.md`, `GEMINI.template.md`, `copilot-instructions.template.md`
- Create: `library/tool-templates/shared/root.template.md`
- Create: `library/tool-templates/{tool}/overrides.md` (6 files, 3-5 lines each)
- Modify: `src/compiler/template-compiler.ts` — load shared + merge overrides
- Modify: `src/scaffold/compiled-root.ts` — remove non-compiled path reference
- Modify: `src/scaffold/root-files.ts` — remove or repurpose
- Modify: `src/commands/init.ts` — remove `--no-compiled-root` flag
- Modify: `src/wizard/index.ts` — remove `useCompiledRoot` branching

#### 1.2 — Rewrite fragments as actionable protocols

Current: 11 XML fragments (~265 lines) describing techniques abstractly.
After: ~8 markdown fragments (~200 lines) with actionable protocols, output templates, and when-to-skip rules.

| Current fragment | Action |
|-----------------|--------|
| chain-of-thought.xml | Rewrite → `reasoning-protocol.md` with `<cot>` output container |
| tree-of-thoughts.xml | Rewrite → `decision-protocol.md` with fill-in comparison format |
| context-engineering.xml | Rewrite → `context-discipline.md` with concrete budgets |
| rpi-workflow.xml | Keep + improve (add session compaction trigger) |
| quality-gates.xml | Keep as-is |
| bug-resolution.xml | Keep as-is |
| git-conventions.xml | Keep as-is |
| agent-harness.xml | Rewrite → coordination protocol (when to use which agent) |
| pivot-handling.xml | Merge into rpi-workflow.md (3 lines of trigger + response) |
| adr-enforcement.xml | Merge into decision-protocol.md |
| system-context.xml | Keep as-is |

Files changed:
- Delete: chain-of-thought.xml, tree-of-thoughts.xml, context-engineering.xml, pivot-handling.xml, adr-enforcement.xml
- Create: reasoning-protocol.md, decision-protocol.md, context-discipline.md
- Modify: rpi-workflow.xml → rpi-workflow.md (add pivot handling + compaction)
- Modify: agent-harness.xml → agent-harness.md (rewrite as coordination protocol)
- Modify: shared/root.template.md — update fragment includes

#### 1.3 — Implement preset system

Current: 9-toggle multiselect in wizard, all ON by default.
After: 4-option select (Minimal / Standard / Full / Custom). Scope-aware defaults.

| Feature | Minimal | Standard | Full |
|---------|:-------:|:--------:|:----:|
| Quality Gates | ✅ | ✅ | ✅ |
| Git Conventions | ✅ | ✅ | ✅ |
| Reasoning Protocol (CoT) | — | ✅ | ✅ |
| RPI Workflow | — | ✅ | ✅ |
| Bug Resolution | — | ✅ | ✅ |
| Context Discipline (CE) | — | — | ✅ |
| Decision Protocol (ToT) | — | — | ✅ |
| ADR Enforcement | — | — | ✅ |
| Agent Harness | — | — | ✅ |

Scope defaults: global → Minimal, project → Standard, workspace → Standard.

Files changed:
- Modify: `src/wizard/phase2-features.ts` — replace multiselect with preset select + custom fallback
- Modify: `src/store/schema.ts` — add preset field, keep individual flags for Custom/storage
- Modify: `src/types.ts` — add PresetLevel type

#### 1.4 — Rename docs/ → specs/

Current: `docs/` with 10 subdirectories.
After: `specs/` with same structure.

Files changed:
- Modify: `src/types.ts` — rename ALL_DOCS_DIRS
- Modify: `src/scaffold/docs.ts` — update paths
- Modify: all library/docs-agents/*.md — update glob paths in `<rule>` tags
- Modify: all references in library content (agents, skills, prompts, fragments, templates)

---

### Wave 2: Content Deduplication

**Goal**: Cut library content by ~50% by giving each concept type its unique role.

#### 2.1 — Deduplicate agents (identity + constraints only)

Rewrite 6 agent files. Remove workflow steps and output formats (those belong in skills).
Each agent keeps: name, model recommendation, identity statement, behavioral constraints, after-task checklist.

468 lines → ~250 lines.

#### 2.2 — Deduplicate skills (workflow + integration only)

Rewrite 7 skill files (merge lessons-learned into memory-write, move parallel-execution to Full-only).
Each skill keeps: workflow steps, trace protocol, integration chain (which agent, what feeds what).
Remove: identity, few-shot examples, reasoning instructions.

817 lines → ~400 lines.

#### 2.3 — Deduplicate prompts (examples + anti-patterns only)

Rewrite 5 prompt files. Each prompt keeps: few-shot examples (input → output shape), common mistakes.
Remove: workflow steps, CoT instructions, output format definitions.

361 lines → ~180 lines.

#### 2.4 — Update document templates

- Merge prd-template.md + techspec-template.md → `plan-template.md`
- Add `spec-template.md` (detailed specification, optional)
- Add `checklist-template.md` (centralized verification)
- Remove `tasks-template.md` (task list = directory listing)
- Remove `progress-template.md` (use handoffs + task status)
- Keep: task-template, adr, bugfix-rca, standard
- Defer to Full preset: code-review-template, postmortem-template

11 templates → 7 (Standard), 9 (Full).

#### 2.5 — Unify workflow AGENTS.md

Replace 4 separate workflow files (features, bugfixes, refactors, tech-debt) with 1 unified file.
Contains: type comparison table, shared rules written once, per-type specifics as subsections.

253 lines → ~80 lines.

---

### Wave 3: Wizard Intelligence

**Goal**: Make the wizard pre-fill values through auto-detection and ask the right questions.

#### 3.1 — Connect repo detection to wizard

Call `detectRepoType()` at wizard start for project scope.
Read package.json scripts (test, lint, build, dev) and lock files (package manager).
Pre-fill detectable values, ask user to confirm.

New wizard prompts (project scope):
- "Detected: TypeScript + Next.js + Vitest. Correct? (Y/n)"
- "Test command: pnpm test (from package.json) — correct? (Y/n)"
- "Database (if any): [PostgreSQL / MySQL / MongoDB / SQLite / None / Other]"

#### 3.2 — Scope-aware templates

Create 3 template variants (or conditional sections in shared template):
- **Project**: full stack info, commands, codebase map (auto-detected)
- **Workspace**: repo inventory table, sync rules, cross-repo coordination
- **Global**: personal AI defaults, reasoning preferences, no project specifics

#### 3.3 — Preset controls scaffolding depth

| Preset | specs/ directories | Templates | Rules |
|--------|:-----------------:|:---------:|:-----:|
| Minimal | standards, memory | 0 | 0 |
| Standard | + features, bugfixes, rules, adrs, templates | 7 | 5 |
| Full | + refactors, tech-debt, all | 9 | 9 |

#### 3.4 — "Ours" vs "theirs" tracking

Add `owner: 'library' | 'user' | 'migrated'` to file manifest.
Re-init behavior: library files → show diff, offer update. User/migrated files → skip automatically, never overwrite.

---

### Wave 4: Workspace Scope

**Goal**: Full multi-repo coordination with knowledge persistence.

#### 4.1 — Lightweight setup for referenced repos

During workspace init, generate per referenced repo:
- Auto-filled root file (AGENTS.md/CLAUDE.md) with detected stack info
- Pointer to planning repo ("Plans and standards live in: /absolute/path/to/documentation/")
- Permission config (per tool's native format)

#### 4.2 — Per-repo permission multiselect

For each referenced repo during init:
```
? What can AI do in api-service?
  ◉ Read files
  ◉ Write/edit code files
  ◉ Run commands (test, lint, build)
  ○ Run destructive commands (migrations, deploy)
  ○ Git operations (commit, push)
```

Compile to Claude Code `.claude/settings.json`, OpenCode config, Copilot settings, etc.

#### 4.3 — Ledger structure in planning repo

Generate per referenced repo:
- `specs/memory/repos/{name}/ledger.md` — append-only activity log
- `specs/memory/repos/{name}/last-known-state.md` — snapshot of repo state

AI appends to ledger after every task. Drift detection on next session start.

#### 4.4 — Absolute paths + permission validation

Store absolute paths for all workspace repos. Validate read/write access at init time.
Any directory can be marked to workspace (not just siblings). Warn on permission issues.

#### 4.5 — /extract-standards skill

New skill for automated standards extraction:
- Takes optional category (testing, coding, security, architecture, data)
- If no category, suggests order based on code volume
- Per pattern: shows pattern + real file reference, user confirms
- `--refresh` flag: compare existing standards against codebase, flag drift
- Workspace variant: per-repo + cross-repo standards, user chooses scope

---

## Scope

### In scope
- All 4 waves as described above
- All 3 setup scopes (global, workspace, project)
- All 6 supported tools (Claude Code, OpenCode, Copilot, Gemini, Pi, Codex)
- Migration/absorb compatibility with existing setups
- Backward compatibility (existing installs continue to work)

### Explicitly out of scope
- New AI tool support (e.g., Cursor, Windsurf)
- CI/CD integration (GitHub Actions, Jira sync)
- GUI / web interface
- Model-specific prompt optimization (adaptive fragment behavior per model tier — future feature)
- Plugin/extension marketplace

---

## Risks

| Risk | Likelihood | Mitigation |
|------|:----------:|-----------|
| Breaking existing installs when renaming docs/ → specs/ | High | Migration path: detect old structure, offer rename. Keep backward compat in detection. |
| Fragment rewrite reduces quality for frontier models | Low | Frontier models ignore unnecessary instructions gracefully. Protocols are also clearer for frontier models. |
| Preset "Standard" missing a feature someone needs | Medium | Custom option always available. Users can switch presets on re-run. |
| Workspace ledger becomes stale if team doesn't use AI consistently | Medium | Ledger is advisory, not blocking. AI checks git log as fallback. |
| Template unification breaks tool-specific edge cases | Low | Override files per tool handle edge cases. Test with all 6 tools. |

---

## Dependencies

- Wave 1 must complete before Wave 3 (presets needed for scaffolding depth)
- Wave 3 must complete before Wave 4 (repo detection needed for workspace auto-fill)
- Waves 1 and 2 can be parallelized (different files touched)
