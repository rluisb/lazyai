# 002 — Simplification & Restructure: Research

## Date
2026-04-06

## Methodology
Multi-perspective review of the entire codebase, cross-referenced with Obsidian second_brain research on prompt engineering (PE-Chain-of-Thoughts, PE-Real-Case-RAG-Agent, PE-GPT5-Reasoning-Parameter, PE-XML-Tags, PE-System-Message-Blueprint, PE-Few-Shot-Prompting, PE-Multi-Step-Prompting, ACE-FCA, cost-per-context) and comparison with GitHub Speckit framework.

Reviewed from 4 engineering perspectives: Junior Developer, Senior Engineer, DevEx/Platform Engineer, and Engineering Manager.

---

## Current State Metrics

| Metric | Value |
|--------|:-----:|
| Files generated (2 tools, all features) | ~73 |
| Concepts user must learn | 24 |
| Feature flags | 9 (all ON by default) |
| `[YOUR_*]` placeholders after wizard | ~29 per root file |
| Library content files | 69 |
| Agent definitions | 6 (468 lines) |
| Skill definitions | 9 (817 lines) |
| Prompt definitions | 5 (361 lines) |
| XML fragments | 11 |
| Document templates | 11 |
| Rule files | 9 (in library, copied to docs/rules/) |
| Compiled tool templates | 6 (nearly identical) |
| Non-compiled root templates | 4 (1,346 lines, 85% overlap) |
| CLI flags on `init` | 18 |
| Wizard prompts (project scope) | ~10 |

---

## Finding 1: Fragments Describe Techniques Instead of Implementing Them

### Evidence

The `chain-of-thought.xml` fragment tells the model ABOUT CoT:
```xml
<thinking-framework>
  <step name="understand"><action>Restate the problem in your own words</action></step>
</thinking-framework>
<output-format><show-reasoning>true</show-reasoning></output-format>
```

But from PE-Real-Case-RAG-Agent (second_brain): explicit `<cot>` tags as output containers took accuracy from 89% → 93% → 100% on GPT-4.1. The difference: **concrete output container + numbered steps + verification**, not abstract framework description.

The non-compiled AGENTS.template.md already implements the same techniques more effectively as actionable protocols (Reasoning Protocol, Architecture Decision Protocol, Token Discipline, Trace Protocol, Confidence Gate) — with when-to-skip rules and concrete examples.

### Model Tier Insight

From PE-GPT5-Reasoning-Parameter: frontier models (GPT-5, Opus) have native reasoning and don't need explicit CoT. But mid-tier models (Sonnet, GPT-4.1, Gemini Pro) and budget models (Haiku, Flash, GPT-4o-mini) benefit significantly from explicit structured prompting. The tool's target audience likely uses mid-tier models to save cost — exactly where these protocols have the highest ROI.

### Impact
Fragments need rewriting, not removal. Current form: ~155 lines of technique descriptions. Proposed form: ~110 lines of actionable protocols with output templates.

---

## Finding 2: Agent/Skill/Prompt Triple Redundancy

### Evidence

The "implement" workflow exists in three places:
- `agents/builder.md` (58 lines): WHO + rules + workflow + output
- `skills/implement.md` (71 lines): workflow + trace + output
- `prompts/implement.md` (57 lines): workflow + examples + anti-patterns

Overlapping content: ~60% of each file repeats what another file says. Same pattern for "research" (scout/research/research) and "plan" (planner/plan/plan).

### Why Keep 3 Concepts

Different AI tools consume them differently:
- OpenCode: agents/ ≠ skills/ (separate directories)
- Copilot: agents/ ≠ prompts/ (different formats)
- Claude Code: only commands/ (skills map to commands)

Merging into 1 file would break tool conventions. The structure is correct, the content overlap is the problem.

### Impact
Agent = identity + constraints only. Skill = workflow + integration only. Prompt = examples + anti-patterns only. Total: 1,646 → ~830 lines (-50%).

---

## Finding 3: Wizard Asks Wrong Questions

### Evidence

The wizard collects: scope, tools, feature flags (9 abstract toggles), git patterns.
The wizard does NOT collect: language, framework, test command, build command, project description.

After wizard finishes: 29 `[YOUR_*]` placeholders per root file that user must fill manually.

Meanwhile, `repo-detection.ts` already detects ruby-rails, nextjs-typescript, react-typescript, go, rust, python — and reads descriptions from package.json/Cargo.toml/pyproject.toml. It also reads package.json scripts for test/lint/build/dev commands. **This detection code exists but is only used for workspace sibling scanning, never for pre-filling templates.**

### Impact
Connect detection → wizard → templates. ~15 placeholders become auto-detected confirms. Remaining ~8 are truly manual (naming conventions, error patterns, coverage threshold).

---

## Finding 4: Compilation Paradox

### Evidence

Two parallel paths exist:
- **Path A (compiled, default)**: tool-templates/{tool}/root.template.md → compiler resolves XML fragments → AGENTS.md. Content: abstract technique descriptions. Controlled by feature flags.
- **Path B (--no-compiled-root)**: library/root/AGENTS.template.md → copied as-is → AGENTS.md. Content: actionable protocols with examples. NOT controlled by feature flags.

Path B produces better content than Path A. But Path A has the better architecture (feature flags, fragment compilation, tool-specific output).

Additionally, the 6 compiled tool templates are byte-for-byte identical except for one sentence naming the tool and 2-3 lines of tool-specific notes.

### Impact
Unify into one path: single shared template + tool-specific overrides + markdown fragments with actionable content. Remove `--no-compiled-root` flag.

---

## Finding 5: Scope-Template Mismatch

### Evidence

One root template (AGENTS.template.md) is used for all three scopes. It contains project-specific placeholders (`[YOUR_LANGUAGE]`, `[YOUR_TEST_COMMAND]`) that don't apply to global scope (machine-wide defaults) or workspace scope (multi-repo, no single language).

For workspace scope: the wizard detects referenced repos and stores their type/description in config.yml, but this information never flows into any template.

### Impact
Three template variants needed: project (auto-detected stack), workspace (repo inventory + coordination), global (personal AI defaults).

---

## Finding 6: docs/ Structure Is Process Scaffolding

### Evidence

The tool generates 10 directories under docs/ with AGENTS.md files: features, bugfixes, refactors, tech-debt, adrs, memory, standards, rules, templates, prompts. This is a project management methodology, not an AI tool setup.

4 workflow variants (features/bugfixes/refactors/tech-debt AGENTS.md) are 80% identical — differing only in which RPI steps to skip and whether ADR is mandatory.

11 document templates are generated regardless of team size or needs (including postmortem-template.md for solo developers).

### Impact
Rename to `specs/`. Preset controls depth (Minimal=2 dirs, Standard=6, Full=all). Unify 4 workflow files into 1. Templates from 11 to 7.

---

## Finding 7: Speckit Alignment

### Evidence

The team uses a Speckit-inspired structure. Speckit uses `plan.md` (combines PRD + TechSpec), `spec.md` (detailed specification), `checklists/`, `quickstart.md`. ai-setup uses separate `prd.md` + `techspec.md` + `tasks/tasks.md` + `progress.md`.

For a 5-6 person team, the separate PRD + TechSpec split adds overhead without proportional value — the same person writes both documents.

### Impact
Adopt Speckit document naming: plan.md (replaces prd+techspec), spec.md (optional, complex only), checklists/requirements.md (centralized verification). Remove progress.md (use handoffs + task status). Add quickstart.md (optional, team onboarding).

---

## Sources

### Obsidian second_brain (primary)
- `resources/rhawk_prompt_engineering/PE-Chain-of-Thoughts.md`
- `resources/rhawk_prompt_engineering/PE-Real-Case-RAG-Agent.md`
- `resources/rhawk_prompt_engineering/PE-GPT5-Reasoning-Parameter.md`
- `resources/rhawk_prompt_engineering/PE-XML-Tags.md`
- `resources/rhawk_prompt_engineering/PE-System-Message-Blueprint.md`
- `resources/rhawk_prompt_engineering/PE-Few-Shot-Prompting.md`
- `resources/rhawk_prompt_engineering/PE-Multi-Step-Prompting.md`
- `resources/rhawk_prompt_engineering/PE-Conditional-Prompts.md`
- `prompt-engineering/techniques/Chain-of-Thought Prompting.md`
- `prompt-engineering/techniques/Tree of Thoughts.md`
- `prompt-engineering/techniques/Self-Consistency.md`
- `prompt-engineering/techniques/ReAct Prompting.md`
- `ai-agents/ACE-FCA.md`
- `resources/developer-toolkit-ai/shared-workflows/context-management/cost-per-context.md`

### External references
- GitHub Speckit framework (structure and SDD methodology)
- Anthropic prompt engineering documentation (XML tag effectiveness)
- OpenAI prompt engineering guide (structured prompting patterns)
