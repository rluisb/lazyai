# 002 — Simplification & Restructure: Decisions & Clarifications

## Date
2026-04-06

This document captures all decisions made during the research and planning phase, including the questions asked, clarifications provided, and rationale for each decision.

---

## D1: Rewrite Fragments as Actionable Protocols

**Decision**: Keep CoT, ToT, CE, and Agent Harness fragments but rewrite them from abstract technique descriptions to actionable protocols with output templates and when-to-skip rules.

**Original assessment**: Remove fragments — "prompt engineering theater" that doesn't change model behavior.

**Challenge raised**: These techniques help users use cheaper models (Sonnet, GPT-4.1, Haiku, Gemini Flash, GPT-4o-mini) to perform above their weight class, saving money and tokens. Not every user can afford frontier models.

**Evidence (from second_brain)**:
- PE-Real-Case-RAG-Agent: Explicit `<cot>` tags took GPT-4.1 accuracy from 89% → 93% → 100%
- PE-GPT5-Reasoning-Parameter: GPT-5 has native reasoning, but GPT-4.1 and below benefit from explicit CoT
- PE-Chain-of-Thoughts: Effective CoT uses concrete output containers + numbered steps + verification

**Corrected take**: The fragments are valuable for non-frontier models. The problem is they describe techniques instead of implementing them. Rewrite with:
- `<cot>` output container (proven to improve accuracy)
- Concrete numbered steps (not abstract "analyze" / "synthesize")
- When-to-skip rules (don't waste tokens on trivial tasks)
- Fill-in-the-blank output templates

**Key finding**: The non-compiled AGENTS.template.md already implements these techniques better as markdown protocols (Reasoning Protocol, Architecture Decision Protocol, Token Discipline, Trace Protocol, Confidence Gate). The rewrite should match that quality level.

---

## D2: Keep 3-Concept Structure, Deduplicate Content

**Decision**: Keep Agent, Skill, and Prompt as separate file types but make each carry only its unique contribution.

**Rationale**: Different AI tools consume these from different directories:
- OpenCode: `agents/builder.md` ≠ `skills/implement.md` (separate dirs)
- Copilot: `agents/builder.md` ≠ `prompts/implement.prompt.md` (different formats)
- Merging into 1 file would break tool conventions

**Content split**:
- Agent = WHO + constraints (identity, model recommendation, behavioral boundaries)
- Skill = WORKFLOW + integration (steps, trace protocol, what feeds what)
- Prompt = EXAMPLES + anti-patterns (few-shot input→output, mistakes to avoid)

**Impact**: 1,646 lines → ~830 lines (-50%). Zero unique value lost.

---

## D3: Scope-Aware Templates

**Decision**: Create different template variants (or conditional sections) for each scope.

**Clarification asked**: "What happens if I select global or workspace? Shouldn't we auto-detect only those projects I marked to work in workspace?"

**Answer**:
- **Global**: Machine-wide defaults. No project to detect. Template should have personal AI preferences, reasoning protocols — NOT `[YOUR_LANGUAGE]` placeholders.
- **Workspace**: Planning repo is a coordination hub, not a code project. Template should have repo inventory table, sync rules, coordination instructions — NOT single-project stack info.
- **Project**: Auto-detection applies. Pre-fill language, framework, commands from repo.

**Principle**: Auto-detection scope matches setup scope. Global = nothing to detect. Workspace = detect each referenced repo individually. Project = detect current directory.

---

## D4: Use Repo Detection in Wizard

**Decision**: Connect existing `detectRepoType()` to the wizard. Pre-fill detectable values, ask user to confirm.

**Current state**: `repo-detection.ts` detects ruby-rails, nextjs-typescript, react-typescript, go, rust, python. Reads descriptions from package.json/Cargo.toml/pyproject.toml. Already imported in wizard but only used for workspace sibling scanning.

**What becomes auto-detected**: language, framework, package manager, test framework, test/lint/build/dev/install commands, project description, protected branch (git default).

**What user still fills manually**: database, ORM (if not detectable from deps), naming convention, error pattern, API convention, import order, coverage threshold, codebase paths.

**Impact**: 29 placeholders → ~8 (non-detectable only).

---

## D5: Workspace Scope — Planning Repo + Referenced Repos

**Decision**: Planning repo gets full setup. Referenced repos get lightweight root + permissions. Single `init` command does everything. Any directory can be marked to workspace (not just siblings). Absolute paths with read/write permissions.

**Clarification asked**: "Explain the planning repo and path structure."

**How it works**:
- Planning repo: no code, stores documentation/planning/standards. Gets full AI setup (agents, skills, prompts, specs/ structure).
- Referenced repos: code repositories. Get lightweight root file (auto-filled with detected stack) + pointer to planning repo + permission config.
- Plans in planning repo reference repos by absolute path. A feature plan might say "implement in api-service (/home/company/api-service) and web-app (/home/company/web-app)."

**Clarification asked**: "Permissions — validate at setup or runtime?"

**Answer**: Ask during init with multiselect per repo. Options are clear for the user (Read / Write / Run commands / Run destructive / Git operations). Compile to each tool's native format.

---

## D6: Ledger Pattern for Knowledge Persistence

**Decision**: Planning repo maintains append-only ledger per referenced repo, tracking who/what/when.

**Problem**: In a team, people change things. Someone might work without AI, delete context files, or move information. How to keep the planning repo accurate?

**Solution**: Three layers of protection:
1. `ledger.md` — append-only, AI adds row after every task. Never edit past entries.
2. `last-known-state.md` — snapshot, AI rewrites after each session. If deleted, next session recreates from repo scan + git log.
3. Git history — always there, even after deletions. AI instructed to check git log if docs seem incomplete.

**Multi-user safety**: AI checks git status + git log before working. Flags drift between plan and reality. Never force-push or rebase shared branches.

---

## D7: Per-Repo Permission Multiselect

**Decision**: During workspace init, ask permission level per repo via multiselect. Compile to each AI tool's native permission format.

**User-facing options** (per repo):
- Read files (always on)
- Write/edit code files
- Run commands (test, lint, build)
- Run destructive commands (migrations, deploy)
- Git operations (commit, push)

**Compiled to**:
- Claude Code → `.claude/settings.json` (allow/deny lists)
- OpenCode → config.toml (workspaces + permissions)
- Copilot → workspace trust settings
- Others → tool-specific format via adapter pattern

---

## D8: Never Remove Existing Files

**Decision**: Track "ours" vs "theirs" in file manifest. Never overwrite user content by default.

**Clarification asked**: "What happens if I create a skill or there are existing agents/skills/commands/hooks in projects?"

**Current state**: Two systems handle this:
1. Migration/Absorb (runs first): detects existing tool setups, copies to .ai/ canonical dir. If canonical file already exists, silently skips.
2. Conflict Resolution (Phase 3): for each planned file that exists, user chooses skip/backup-and-replace/align.

**Gaps identified**:
- No distinction between "ours" (library files) vs "theirs" (user/migrated files) during re-init
- Custom files created after init aren't tracked in manifest
- No merge option for modified library files
- Workspace referenced repos: writing into repos we don't own

**Solution**:
- Add `owner` field to manifest: `library | user | migrated`
- Re-init: library files → show diff, offer update. User files → skip automatically.
- Referenced repos: "add alongside" as default (keep everything, only add what's missing). User can choose "review each" or "skip repo."
- **Principle**: In referenced repos, we are a guest.

---

## D9: Replace 9 Feature Flags With 4 Presets

**Decision**: Replace 9-toggle multiselect with Minimal / Standard / Full / Custom presets.

**Rationale**: 9 abstract toggles (Context Engineering, Tree of Thoughts, Pivot Handling) cause decision fatigue. All 9 pre-selected means the UI doesn't function as a choice. Users leave everything on by default.

**Presets**:
- Minimal: Quality Gates + Git Conventions (2 features)
- Standard: + Reasoning Protocol + RPI Workflow + Bug Resolution (5 features, recommended)
- Full: + Context Discipline + Decision Protocol + ADR Enforcement + Agent Harness (all 9)
- Custom: pick individually

**Scope-aware defaults**: Global → Minimal, Project → Standard, Workspace → Standard.

---

## D10: Rename Features to Action Names

**Decision**: Use names that describe what features DO, not the technique name.

| Current name | New name |
|-------------|----------|
| Chain of Thought | Reasoning Protocol |
| Tree of Thoughts | Decision Protocol |
| Context Engineering | Context Discipline |

**Rationale**: "Reasoning Protocol" is self-explanatory. "Chain of Thought" requires knowing the research. Users who've never heard of ToT can still understand "Decision Protocol."

---

## D11: Unify Compiled + Non-Compiled Paths

**Decision**: Single compilation path. One shared template + tool-specific overrides. Markdown fragments. Remove `--no-compiled-root` flag.

**Problem**: Two paths produce the same thing differently. The non-compiled path has better content but no feature flag control. The compiled path has feature flags but weaker content. 6 compiled tool templates are nearly identical.

**Solution**:
- 1 shared template (`library/tool-templates/shared/root.template.md`)
- 6 override files (3-5 lines each: tool name + file location notes)
- Fragments in markdown format (not XML), with XML output tags where useful (e.g., `<cot>`)
- Same feature flag system controls both content and scaffolding depth

---

## D12: Rename docs/ → specs/

**Decision**: Planning artifacts and project reference documentation live under `specs/` instead of `docs/`.

**Structure**:
```
specs/
├── features/     ← Planning: feature specs
├── bugfixes/     ← Planning: bug investigations
├── refactors/    ← Planning: refactor specs (Full preset)
├── tech-debt/    ← Planning: tech debt assessments (Full preset)
├── adrs/         ← Reference: architecture decisions
├── standards/    ← Reference: codebase patterns
├── rules/        ← Reference: team conventions
├── memory/       ← Reference: AI knowledge persistence
└── templates/    ← Reference: document templates
```

---

## D13: plan.md Replaces PRD + TechSpec

**Decision**: Use `plan.md` (WHAT + WHY + HOW combined) instead of separate `prd.md` + `techspec.md`.

**Clarification asked**: "Should we use PRD, TechSpec, tasks breakdown or something like Speckit?"

**Answer**: For a 5-6 person team, the PRD/TechSpec split adds overhead without proportional value. The same person writes both. Adopted Speckit-inspired naming:
- `plan.md` — combines what + why + how
- `spec.md` — optional, detailed specification for complex features/refactors
- `checklists/requirements.md` — centralized verification criteria
- `quickstart.md` — optional, team onboarding to a feature
- `data-model.md` — optional, significant data changes
- `progress.md` — removed (use handoffs + task status)
- `tasks/tasks.md` overview — removed (task list = directory listing)

**Important clarification**: ai-setup is its own tool, not a Speckit integration. It takes inspiration from Speckit's structure but provides a more complete solution (multi-tool, workspace, prompt engineering, standards extraction).

---

## D14: Unified Workflow AGENTS.md

**Decision**: One workflow file covering all types instead of 4 separate files.

**Current**: 4 files (features.md, bugfixes.md, refactors.md, tech-debt.md) totaling 253 lines with ~80% overlap.

**Proposed**: 1 file (~80 lines) with type comparison table + shared rules written once. Per-type AGENTS.md files available as optional add-on in Full preset.

---

## D15: /extract-standards Skill

**Decision**: Create a skill for automated standards extraction from code.

**Clarification asked**: "We should have something like an instruction or command to read each project or workspace repos and create standards. Maybe asking user to confirm what was identified? Too overwhelming?"

**Answer**: Not overwhelming if done per-category, on-demand.

**How it works**:
- Takes optional category (testing, coding, security, architecture, data)
- If no category: scans lightly, suggests extraction order based on code volume
- Per pattern found: shows pattern + real file reference, user confirms (accept/skip/edit)
- `--refresh` flag: compare existing standards against codebase, flag drift
- Workspace variant: per-repo + cross-repo standards, user chooses scope during extraction

**Clarification asked**: "How, where and why will we store that information after extraction?"

**Answer**:
- **Where**: `specs/standards/{category}/` for project scope. `specs/standards/{repo-name}/{category}/` for workspace. Plus `specs/standards/cross-repo/` for shared patterns.
- **Why there**: AI loads standards based on task type (from AGENTS.md progressive loading table). Standards must be where the AI looks.
- **How maintained**: AI checks references on every task + `--refresh` for explicit re-scan + ledger drift detection for workspace.

---

## D16: Templates 11 → 7

**Decision**: Reduce document templates for Standard preset.

| Template | Status |
|----------|--------|
| plan-template.md | New (merges prd + techspec) |
| spec-template.md | New (detailed specification) |
| task-template.md | Keep |
| adr-template.md | Keep |
| bugfix-rca-template.md | Keep |
| standard-template.md | Keep |
| checklist-template.md | New (from Speckit) |
| code-review-template.md | Deferred to Full preset |
| postmortem-template.md | Deferred to Full preset |
| prd-template.md | Removed (merged into plan) |
| techspec-template.md | Removed (merged into plan) |
| tasks-template.md | Removed (task list = directory) |
| progress-template.md | Removed (use handoffs) |

---

## Open Questions (for future consideration)

### Q1: Model-adaptive fragments
Should fragments behave differently based on model tier? E.g., skip `<cot>` tags for frontier models (Opus, GPT-5), use full protocol for mid-tier (Sonnet, GPT-4.1), add few-shot examples for budget (Haiku, Flash).

**Status**: Acknowledged as valuable, deferred. Current approach: write protocols that work across all tiers. Frontier models ignore unnecessary instructions gracefully.

### Q2: CI/CD integration
Should ai-setup generate GitHub Actions workflows, Jira integration, or similar?

**Status**: Out of scope for this restructure. May be a future feature.

### Q3: Team presets
Should teams be able to save and share preset configurations? E.g., "Our team uses Standard + these 3 custom rules" saved as a sharable config.

**Status**: Not planned for this restructure. The config.yml already stores selections, which could be shared via git.

### Q4: Speckit migration path
If a team currently uses Speckit and wants to switch to ai-setup, should there be a migration command?

**Status**: Not planned. ai-setup's migration engine could be extended to detect `.specify/` directories, but this is a separate feature.
