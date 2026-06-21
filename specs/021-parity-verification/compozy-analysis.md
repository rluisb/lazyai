> **SUPERSEDED — Go/TS parity is moot; the TS CLI was removed. Retained for historical context.**
>
> This document is historical; `packages/ai-setup-ts` is no longer present.

# Compozy Adaptation Analysis for ai-setup

## Executive Summary

Compozy has a few setup-time patterns that map well to `ai-setup`'s existing mission: better installation topology, more portable content formats, stronger drift visibility, and broader environment detection. The highest-value adaptations are: a symlink-based install mode layered on top of the current tracked-file model, a normalized `SKILL.md` contract, improved reusable-agent packaging, and non-blocking drift detection integrated into `doctor` and `update`.

What is not worth adapting is anything that turns `ai-setup` from a scaffold/install tool into a runtime orchestrator. That includes ACP, daemon/workflow execution, multi-phase delivery pipelines, and hard execution blocking when skills are missing. `ai-setup` already has the right boundary: install, compile, detect, migrate, and verify local AI tool configuration. The roadmap below keeps that boundary intact.

## 1. Symlink Installation Mode

- **What Compozy does:** Installs one canonical copy of each skill/agent/extension, then symlinks tool-specific destinations to that canonical source. Copy mode remains available for users who want isolated files.
- **Current ai-setup behavior:** `ai-setup` writes concrete files into tool directories and tracks them in `.ai-setup.json`. The current model is file-copy based, with tracked file hashes in `src/store/schema.ts` and path-based inference in `src/utils/manifest.ts`. Existing adapter tests assume real files are written, for example `.opencode/skills/implement/SKILL.md` in `src/__tests__/adapters-files.test.ts`.
- **Adaptation proposal:**
  1. Add an install strategy setting, e.g. `installMode: "copy" | "symlink"`, to the store schema in `src/store/schema.ts` under `config`.
  2. Treat `library/` as the canonical source for generated assets and keep tool adapters responsible for destination path mapping only.
  3. Update adapter install paths so supported tool resources can be materialized either by copy or symlink. Start with skills and agents for `opencode` and `claude-code`; defer prompts/rules until the filesystem behavior is proven.
  4. Extend tracked file records with link metadata, e.g. `kind: "file" | "symlink"` and `linkTarget?: string`, so `doctor` can distinguish drift from a broken link.
  5. Add `setup --install-mode=copy|symlink` and surface the choice in the wizard summary before apply.
  6. Keep copy mode as default initially for backward compatibility; make symlink opt-in in the first release.
- **Effort:** medium
- **Risk:** medium
- **Priority:** P1

## 2. Unified SKILL.md Format

- **What Compozy does:** Uses a normalized `SKILL.md` contract with YAML frontmatter fields like `name`, `description`, and `argument-hint`, regardless of runtime.
- **Current ai-setup behavior:** `ai-setup` already partially converges on this shape for installed skills: the planner maps skills for Claude/OpenCode/Codex/Gemini to `<skill>/SKILL.md` in `src/wizard/planner.ts`. However, the generator still produces `library/skills/<slug>.md` with markdown headings and no required frontmatter in `src/generators/skill.ts`.
- **Adaptation proposal:**
  1. Make `library/skills/*` canonical source files include YAML frontmatter by default:
     ```md
     ---
     name: implement
     description: Implement requested changes safely
     argument-hint: [task-or-scope]
     ---
     ```
  2. Update `src/generators/skill.ts` to emit that frontmatter for all newly generated skills.
  3. Add a small shared parser/validator beside `src/utils/frontmatter.ts` so `info`, `setup`, and future diagnostics can read skill metadata uniformly.
  4. Preserve adapter-specific rendering differences only at install time. The library representation should stay runtime-neutral.
  5. Add tests covering backward compatibility for old skill markdown without frontmatter, then migrate bundled library skills incrementally.
- **Effort:** low
- **Risk:** low
- **Priority:** P1

## 3. Reusable Agent Pattern (AGENT.md + mcp.json)

- **What Compozy does:** Packages reusable agents as `.compozy/agents/<name>/AGENT.md` plus optional colocated `mcp.json`.
- **Current ai-setup behavior:** `ai-setup` already has most of this concept in scan/setup flows. `src/commands/setup.ts` detects reusable agents under `.ai/agents/<id>/`, requires `AGENT.md`, and optionally reads sibling `mcp.json`. The parser accepts YAML frontmatter for `title`, `name`, `description`, and `tools`.
- **Adaptation proposal:**
  1. Standardize the canonical reusable-agent source layout under `library/agents/<name>/` instead of only flat `library/agents/<name>.md`.
  2. Introduce support for dual-source compatibility during migration:
     - existing: `library/agents/builder.md`
     - new: `library/agents/builder/AGENT.md` and optional `library/agents/builder/mcp.json`
  3. Update `src/generators/agent.ts` to generate directory-based agents with `AGENT.md` frontmatter and optional starter `mcp.json` template.
  4. Refactor adapter install code to consume a normalized agent artifact object rather than assuming a single markdown file.
  5. Reuse the existing observer logic in `src/commands/setup.ts` for validation so install-time and scan-time rules stay aligned.
  6. Record this as an internal packaging change only; do not expose runtime orchestration semantics.
- **Effort:** medium
- **Risk:** medium
- **Priority:** P1

## 4. TOML Configuration

- **What Compozy does:** Supports both global and project `.toml` config with explicit precedence.
- **Current ai-setup behavior:** Persistent state lives in `.ai-setup.json`, with schema managed in `src/store/schema.ts`. Some tool configs are also emitted as JSON/JSONC. The CLI already supports global, workspace, and project scopes, but its own management state is JSON-based.
- **Adaptation proposal:**
  1. Do not replace `.ai-setup.json` as the operational manifest; it already tracks file hashes, ownership, sync state, and operations.
  2. Add an optional user-authored config layer, e.g. `.ai-setup.toml` and `~/.config/ai-setup/config.toml`, for preferences rather than inventory.
  3. Limit TOML to declarative defaults such as:
     - default tools
     - default setup scope
     - preferred install mode
     - default enabled servers
     - wizard presets
  4. Resolve precedence as: CLI flags > project TOML > global TOML > existing wizard defaults.
  5. Parse TOML into the existing internal config shape before command execution; keep store writes in JSON.
  6. Document the split clearly: TOML is desired configuration, `.ai-setup.json` is observed/applied state.
- **Effort:** medium
- **Risk:** low
- **Priority:** P2

## 5. Extended Agent Auto-Detection

- **What Compozy does:** Detects 40+ agent environments by checking known filesystem locations and config signatures.
- **Current ai-setup behavior:** Detection is intentionally narrower. `src/migration/detector.ts` currently defines patterns for `opencode`, `claude-code`, `gemini`, and `copilot`. User-facing setup scope also recognizes `codex` and `pi`, as seen in `src/commands/setup.ts` and `src/wizard/phase1-context.ts`.
- **Adaptation proposal:**
  1. First close internal parity gaps by making migration detection fully cover all first-class `ai-setup` tools: add `codex` and `pi` to `DETECTION_PATTERNS` in `src/migration/detector.ts`.
  2. Split detection into two tiers:
     - **supported**: tools `ai-setup` can install/manage today
     - **observed**: extra environments detected and reported, but not managed
  3. Introduce a registry-driven detector table instead of hard-coded constants, e.g. `src/migration/registry/discovery.ts` or a new `src/detection/catalog.ts`.
  4. Add a `setup --list-detectors` or `doctor --scan` view that shows detected environments and whether each is manageable, migratable, or informational only.
  5. Expand to adjacent tools only when there is a migration or reporting use case. Recommended next wave: Cursor, Aider, Continue, Zed. Do not chase Compozy's full 40+ matrix unless user demand emerges.
- **Effort:** medium
- **Risk:** low
- **Priority:** P2

## 6. Interactive Setup Improvements

- **What Compozy does:** Uses an interactive setup wizard with previews before installation.
- **Current ai-setup behavior:** `ai-setup` already has a wizard and a planner. `src/wizard/planner.ts` computes file plans, and `src/wizard/index.ts` already supports phase-based setup, prior defaults, workspace handling, and some confirmation flows.
- **Adaptation proposal:**
  1. Add an explicit final preview screen driven by `computePlan()` from `src/wizard/planner.ts`, grouped by category and tool.
  2. Show install semantics in the preview: new file, overwrite, symlink, adopt, import, or unchanged.
  3. Add a per-tool summary section so users can see exactly which roots are affected, especially for global/workspace setups.
  4. Include reusable-agent and MCP summaries when those assets are selected.
  5. Reuse existing dry-run structures from `src/commands/setup.ts` so wizard preview and CLI `setup --dry-run` stay consistent.
  6. Keep the current non-blocking wizard model; do not introduce Compozy-style runtime dependency gating.
- **Effort:** low
- **Risk:** low
- **Priority:** P1

## 7. Extension System

- **What Compozy does:** Offers an extension SDK with hooks, decorators, review providers, and packaged skill bundles.
- **Current ai-setup behavior:** `ai-setup` has extensible content generation and orchestration catalog concepts, but not a formal plugin API. Relevant extension-like seams already exist in generators, scaffolders, adapters, and orchestration catalog code under `src/generators/`, `src/scaffold/`, and `src/orchestration/`.
- **Adaptation proposal:**
  1. Do not implement a full plugin runtime in the near term.
  2. Extract a narrow, setup-time extension surface first:
     - custom generator packs
     - custom template/rule/skill libraries
     - custom detector entries
     - custom tool adapters loaded from a local path
  3. Represent extension metadata declaratively in project config rather than executing arbitrary lifecycle hooks initially.
  4. If demand appears, add a read-only hook model later for plan augmentation and preview decoration before allowing write-time hooks.
  5. Record a hard boundary: no background services, no runtime task execution hooks, no implicit network access.
- **Effort:** high
- **Risk:** high
- **Priority:** P3

## 8. Skill Verification / Drift Detection

- **What Compozy does:** Verifies whether required skills are current, missing, or drifted, and can block execution when prerequisites are missing.
- **Current ai-setup behavior:** `ai-setup` already has the primitives for drift detection. The manifest/store tracks hashes and ownership in `src/store/schema.ts`; `doctor` verifies integrity against `.ai-setup.json`; `update` and migration flows already reason about managed versus user-owned files.
- **Adaptation proposal:**
  1. Add a first-class content status model for managed assets: `current`, `missing`, `modified`, `drifted`, `user-owned`.
  2. Extend `doctor` output to summarize drift by category: root files, agents, skills, prompts, MCP configs.
  3. Add `update --check` or `status` enhancements that compare installed hashes to the current `library/` sources and report upgradable content.
  4. For symlink mode, treat link target mismatch and broken symlinks as drift.
  5. Make verification advisory, not blocking. The right UX is warning plus suggested repair commands, not execution prevention.
  6. Optionally add `setup --repair-skills` later as a convenience wrapper over existing update/repair mechanics.
- **Effort:** medium
- **Risk:** low
- **Priority:** P1

## Summary: Prioritized Roadmap

| Rank | Feature | Effort | Impact | Timeline |
|------|---------|--------|--------|----------|
| 1 | Unified `SKILL.md` format | low | high | next minor release |
| 2 | Skill verification / drift detection | medium | high | next minor release |
| 3 | Interactive setup preview improvements | low | medium | next minor release |
| 4 | Reusable agent packaging (`AGENT.md` + optional `mcp.json`) | medium | high | 1-2 releases |
| 5 | Symlink installation mode | medium | high | 1-2 releases |
| 6 | TOML preference layer | medium | medium | 2 releases |
| 7 | Extended agent auto-detection | medium | medium | 2 releases |
| 8 | Setup-time extension surface | high | medium | later / validate demand |

## Design Decisions

1. **Keep `ai-setup` as a setup tool, not a runtime orchestrator.**
   - ADR recommended: yes.
   - Reason: prevents scope creep into ACP, daemons, task pipelines, and review execution.

2. **Separate desired config from observed state.**
   - ADR recommended: yes.
   - TOML should hold user intent and defaults; `.ai-setup.json` should remain the authoritative applied-state manifest.

3. **Adopt unified content packaging without enforcing runtime lock-step.**
   - ADR recommended: yes.
   - Standardized `SKILL.md` and `AGENT.md` layouts improve portability, but verification must remain advisory.

4. **Introduce symlinks as an install strategy, not a storage rewrite.**
   - ADR recommended: maybe.
   - If implemented, the canonical source remains `library/`; adapters still own destination layout.

5. **Prefer registry-driven detection over one-off conditionals.**
   - ADR recommended: no, unless external detector packs are added.

## Appendix: Compozy Config Examples

### Example `~/.compozy/config.toml`

```toml
[defaults]
install_mode = "symlink"
tools = ["claude", "codex", "opencode"]

[paths]
skills_dir = "~/.compozy/skills"
agents_dir = "~/.compozy/agents"
```

### Example project `.compozy/config.toml`

```toml
[project]
name = "ai-setup"
skill_verification = true

[agents.builder]
enabled = true

[agents.reviewer]
enabled = true
```

### Equivalent direction for `ai-setup`

If `ai-setup` adopts TOML, the closest safe shape would be preference-oriented rather than state-oriented:

```toml
default_scope = "project"
default_tools = ["opencode", "claude-code", "codex"]
install_mode = "copy"
enabled_servers = ["filesystem", "ripgrep"]

[wizard]
preset = "standard"
show_preview = true
```

That file should influence setup defaults only. It should not replace `.ai-setup.json`, which must continue tracking installed files, hashes, ownership, operations, and sync status.
