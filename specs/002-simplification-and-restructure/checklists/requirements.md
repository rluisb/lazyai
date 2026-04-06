# 002 — Simplification & Restructure: Requirements Checklist

## Wave 1: Foundation

### 1.1 — Unified templates
- [ ] Single shared template exists at `library/tool-templates/shared/root.template.md`
- [ ] 6 override files exist (one per tool, 3-5 lines each)
- [ ] Old 6 per-tool templates deleted
- [ ] Old 4 non-compiled root templates deleted
- [ ] `--no-compiled-root` flag removed from CLI
- [ ] `useCompiledRoot` branching removed from wizard
- [ ] All 6 tools produce correct output from shared template + override
- [ ] Existing tests updated and passing

### 1.2 — Actionable fragments
- [ ] `reasoning-protocol.md` replaces `chain-of-thought.xml` — has `<cot>` output container, 4 numbered steps, when-to-skip
- [ ] `decision-protocol.md` replaces `tree-of-thoughts.xml` — has fill-in comparison format, when-to-skip
- [ ] `context-discipline.md` replaces `context-engineering.xml` — has concrete file budget, session hygiene, priority order
- [ ] `rpi-workflow.md` includes pivot handling (merged from pivot-handling.xml)
- [ ] `agent-harness.md` rewritten as coordination protocol
- [ ] Old XML fragments deleted (chain-of-thought.xml, tree-of-thoughts.xml, context-engineering.xml, pivot-handling.xml, adr-enforcement.xml)
- [ ] Fragment includes updated in shared template
- [ ] Each fragment has when-to-skip guidance
- [ ] Each fragment has at least one mini example or output template

### 1.3 — Preset system
- [ ] Wizard shows 4-option select: Minimal / Standard / Full / Custom
- [ ] Custom falls back to individual toggles (current behavior)
- [ ] Scope-aware defaults: global → Minimal, project → Standard, workspace → Standard
- [ ] Feature names renamed in wizard UI (Reasoning Protocol, Decision Protocol, Context Discipline)
- [ ] Presets correctly map to feature flag combinations
- [ ] Stored in config.yml alongside individual flags

### 1.4 — Rename docs/ → specs/
- [ ] All generated directories use `specs/` instead of `docs/`
- [ ] ALL_DOCS_DIRS constant renamed and paths updated
- [ ] All library content references updated (docs-agents, fragments, templates, agent files)
- [ ] Backward compat: detection/migration recognizes old `docs/` structure

---

## Wave 2: Content Deduplication

### 2.1 — Agents deduplicated
- [ ] Each agent file contains ONLY: identity, model recommendation, constraints, after-task checklist
- [ ] No workflow steps in agent files (moved to skills)
- [ ] No output format in agent files (moved to prompts)
- [ ] All 6 agents under 45 lines each
- [ ] Total agent content ≤ 270 lines

### 2.2 — Skills deduplicated
- [ ] Each skill file contains ONLY: workflow steps, trace protocol, integration chain
- [ ] No identity/persona in skill files
- [ ] No few-shot examples in skill files
- [ ] lessons-learned merged into memory-write
- [ ] parallel-execution moved to Full-only (not in Standard)
- [ ] Total: 7 skills ≤ 420 lines

### 2.3 — Prompts deduplicated
- [ ] Each prompt file contains ONLY: few-shot examples, common mistakes
- [ ] No workflow steps in prompt files
- [ ] No CoT instructions in prompt files
- [ ] Total: 5 prompts ≤ 200 lines

### 2.4 — Templates updated
- [ ] `plan-template.md` exists (merges PRD + TechSpec)
- [ ] `spec-template.md` exists (detailed specification)
- [ ] `checklist-template.md` exists (verification criteria)
- [ ] `prd-template.md` removed
- [ ] `techspec-template.md` removed
- [ ] `tasks-template.md` removed
- [ ] `progress-template.md` removed
- [ ] Standard preset installs 7 templates
- [ ] Full preset installs 9 templates

### 2.5 — Unified workflow
- [ ] Single `specs/AGENTS.md` (or `specs/features/AGENTS.md`) covers all workflow types
- [ ] Type comparison table present (feature/bugfix/refactor/tech-debt)
- [ ] Shared rules written once (observability, human gates, self-improvement)
- [ ] Per-type specifics as subsections
- [ ] Old 4 separate workflow files removed (or gated behind Full preset)
- [ ] Total ≤ 100 lines

---

## Wave 3: Wizard Intelligence

### 3.1 — Repo detection connected
- [ ] `detectRepoType()` called at wizard start for project scope
- [ ] Package.json scripts read (test, lint, build, dev)
- [ ] Lock file detected (package manager)
- [ ] Detected values shown to user for confirmation
- [ ] User can override any detected value
- [ ] Undetectable values asked as questions (database, ORM if not in deps)
- [ ] Remaining non-detectable values have sensible defaults or `<!-- TODO -->` markers

### 3.2 — Scope-aware templates
- [ ] Project scope: root file has auto-detected stack info, commands, codebase map
- [ ] Workspace scope: root file has repo inventory table, sync rules, no `[YOUR_LANGUAGE]`
- [ ] Global scope: root file has personal defaults, reasoning prefs, no project placeholders
- [ ] Detected repo info (from workspace scan) flows into workspace template

### 3.3 — Preset controls scaffolding
- [ ] Minimal: generates 2 specs/ directories (standards, memory)
- [ ] Standard: generates 6 specs/ directories + 7 templates + 5 rules
- [ ] Full: generates all directories + all templates + all rules
- [ ] File count matches target: ~45 for Standard + 2 tools

### 3.4 — Ours vs theirs
- [ ] File manifest tracks `owner` field (library / user / migrated)
- [ ] Re-init: library files show diff, offer update
- [ ] Re-init: user/migrated files skipped automatically with note
- [ ] New files created after init are not overwritten on re-init

---

## Wave 4: Workspace Scope

### 4.1 — Lightweight repo setup
- [ ] Each referenced repo gets auto-filled root file (language, framework, commands)
- [ ] Each referenced repo gets pointer to planning repo (absolute path)
- [ ] Each referenced repo gets permission config (per tool's native format)
- [ ] Planning repo root file has repo inventory table (auto-generated from detection)

### 4.2 — Permission multiselect
- [ ] Per-repo permission prompt during workspace init
- [ ] Options: Read / Write / Run commands / Run destructive / Git operations
- [ ] Compiled to Claude Code settings.json format
- [ ] Compiled to OpenCode config format
- [ ] Compiled to Copilot settings format
- [ ] Stored in config.yml as tool-agnostic representation

### 4.3 — Ledger structure
- [ ] `specs/memory/repos/{name}/ledger.md` generated per referenced repo
- [ ] `specs/memory/repos/{name}/last-known-state.md` generated per referenced repo
- [ ] Ledger format documented (Date / Who / What / Plan ref / Verified)
- [ ] AI instructions include "append to ledger after every task"

### 4.4 — Paths and validation
- [ ] Absolute paths stored for all workspace repos
- [ ] Any directory can be marked (not just siblings)
- [ ] Read permission validated at init time
- [ ] Write permission validated if enabled
- [ ] Clear warning on permission issues

### 4.5 — Extract standards skill
- [ ] Skill file exists at `library/skills/extract-standards.md`
- [ ] Takes optional category parameter
- [ ] Without category: suggests extraction order
- [ ] Per pattern: shows pattern + file reference, user confirms
- [ ] Writes to `specs/standards/{category}/` using standard-template
- [ ] Workspace variant: per-repo and cross-repo options
- [ ] `--refresh` flag: detects drift in existing standards

---

## Cross-Cutting Requirements

- [ ] All existing tests pass after each wave
- [ ] New tests written for: preset system, repo detection in wizard, scope-aware templates, permission compilation
- [ ] No breaking changes for existing installs (migration path for docs/ → specs/)
- [ ] README.md updated to reflect new structure, presets, and commands
- [ ] Library file count reduced by ≥25% (from ~69 to ≤52)
- [ ] Generated file count (Standard, 2 tools) reduced by ≥30% (from ~73 to ≤52)
