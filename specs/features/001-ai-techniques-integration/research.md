# Research: AI Techniques Integration into ai-setup Installation

**Feature:** 001 — Evaluate and rate AI techniques for integration into the ai-setup installation process.  
**Source:** `ai-techniques-patterns-review.md` — 13 NICE TO HAVE + 18 DO NOT HAVE techniques  
**Phase:** Research (RPI Phase 1)  
**Date:** 2026-05-01  
**Mode:** Analysis and rating only — no implementation planning yet.

---

## 1. Evaluation Framework

Each technique is rated on five axes, producing a final priority/effort recommendation.

### Axes

| Axis | Scale | Meaning |
|------|-------|---------|
| **Integration Complexity** | Trivial / Low / Medium / High / Very-High | How hard is it to add to ai-setup's install scaffold? |
| **Installation Impact** | None / Files-only / Feature-flag / Runtime / Deep | What layers of ai-setup must change? |
| **User Value** | Polish / Useful / High-Value / Transformative | How much does this improve the experience for ai-setup adopters? |
| **Alignment** | Misaligned / Peripheral / Good-Fit / Core | How naturally does this belong in ai-setup vs. in a downstream project? |
| **Dependency** | Self-contained / Needs-other-technique / Blocked | Does it depend on other techniques being implemented first? |

### Priority Scale

| Priority | Meaning |
|----------|---------|
| **P0** | Do now — foundational gap, blocks other work, or fixes broken contract |
| **P1** | Do soon — high value, well-aligned, moderate effort |
| **P2** | Do later — useful but not urgent; dependent on other work or lower value |
| **P3** | Nice someday — polish, niche, or belongs in downstream projects |

### Effort Scale

| Effort | What it means |
|--------|--------------|
| **XS** | A few lines in a config file or template edit (< 1 hour) |
| **S** | One new file or small fragment (1–3 hours) |
| **M** | Multiple files across library/agents/skills/templates + feature flag (1–3 days) |
| **L** | New orchestrator runtime capability + library content + CLI changes (1–2 weeks) |
| **XL** | Major architectural addition (multiple weeks) |

---

## 2. NICE TO HAVE — Analysis (13 techniques)

These techniques exist in documentation or concepts but enforcement, automation, or completeness is limited.

---

### N1: Few-Shot Prompting

**What it is:** A systematic catalog of few-shot examples (successful task completions, reviewed code patterns, effective reasoning traces) that agents can retrieve to guide new tasks. Goes beyond the current ad-hoc `library/prompts/` directory with a retrieval mechanism and structured format.

**Current state in ai-setup:**
- `library/prompts/` has 6 prompt files with few-shot examples for plan, research, implement, and compact.
- `specs/prompts/local-examples/` has 3 reference examples (commit-message, preflight-framing, react-trace).
- `library/fragments/` and `library/tool-templates/shared/root.template.md` provide the compiled instructions.
- No retrieval mechanism — agents must manually load prompts.

**Integration complexity:** Low-Medium. The content already exists. The gap is systematization, not creation.

**Installation impact:** Files-only + potential skill. Could be done as:
1. Restructure `library/prompts/` into a proper few-shot catalog with tags, categories, and difficulty levels.
2. Add a `few-shot` skill that agents load to retrieve relevant examples.
3. Add a feature flag to include few-shot guidance in compiled AGENTS.md.

**User value:** Useful. Improves consistency across sessions by giving agents concrete examples. Especially valuable for new ai-setup adopters who don't have their own project-specific examples yet.

**Alignment:** Good-Fit. ai-setup is the natural place to define what good agent behavior looks like. A few-shot catalog is essentially "standards by example."

**Dependency:** Self-contained, but becomes more powerful when combined with RAG retrieval (D1).

**Rating:** **P2 / S-M**
- **Priority P2** — Useful but not urgent. Current ad-hoc prompts work adequately for most workflows.
- **Effort S-M** — S if just reorganizing existing content + adding a skill; M if adding retrieval mechanism.
- **Recommendation:** Restructure `library/prompts/` with frontmatter tags. Add a `few-shot` skill that loads relevant examples based on task type. Ship as a new feature flag `fewShotPrompting`.

---

### N2: Generated Knowledge

**What it is:** Before acting on a task, the agent generates structured knowledge statements (facts, constraints, assumptions, unknowns) and uses those as context for the actual implementation. A reasoning technique that front-loads understanding. Think: "first, write down everything you know about this problem."

**Current state in ai-setup:**
- The reasoning protocol (`library/fragments/reasoning-protocol.md`) has a "Reasoning Protocol" with "Think before acting" step. This is close but not identical.
- The preflight task framing (`specs/prompts/local-examples/preflight-task-framing.md`) requires listing files, assumptions, uncertainty level.
- No dedicated "generate knowledge" step before acting.

**Integration complexity:** Low. This is primarily a prompt engineering change — adding a step to the reasoning protocol fragment.

**Installation impact:** Fragment edit only. Modify `reasoning-protocol.md` to add a "Generate Knowledge" step before "Think before acting." Could optionally be gated by a feature flag.

**User value:** Useful. Improves decision quality by forcing explicit knowledge surfacing. Especially helpful for junior-level models or unfamiliar codebases.

**Alignment:** Good-Fit. This is a reasoning technique that belongs in the root instruction set.

**Dependency:** Self-contained.

**Rating:** **P2 / XS**
- **Priority P2** — Simple to add, but the existing reasoning protocol already covers much of this territory. Marginal improvement.
- **Effort XS** — A paragraph in the reasoning protocol fragment.
- **Recommendation:** Add a "Knowledge Surfacing" step to `reasoning-protocol.md`. Not worth a separate feature flag — fold it into the existing `chainOfThought` flag. This is essentially an enhancement of the reasoning protocol, not a standalone feature.

---

### N3: Observability Readiness

**What it is:** Gate 5 in the quality ladder requires observability readiness (logging, metrics, tracing). Currently this is documented in `quality-gates.xml` but has no automated enforcement or CI integration.

**Current state in ai-setup:**
- `library/fragments/quality-gates.xml` defines Gate 5: "Observability Readiness — logging for failures, metrics for key operations, tracing for cross-service flows."
- No CI integration, no pre-commit hook for this gate, no automated check.
- The `feature` chain's review step can flag missing observability, but it's human-driven.

**Integration complexity:** Medium-High. Automating observability checks is inherently project-specific — what constitutes "good observability" depends on the stack, framework, and team conventions.

**Installation impact:** Could be done as:
1. A new `specs/standards/observability/` standard with stack-specific patterns.
2. CI template additions (lint rules for logging, test coverage for instrumentation).
3. A `compliance.md` or rule addition for observability requirements.

However, the core challenge is that observability is fundamentally a **runtime concern** — the installation scaffold can only provide guidance, patterns, and lint rules. It can't automatically verify that your error handling includes proper logging.

**User value:** High-Value for production teams, but limited for what installation can provide. The real value would come from project-specific standards and lint rules.

**Alignment:** Peripheral. ai-setup can provide patterns and rules, but the enforcement is deeply project-specific. This belongs more in `specs/standards/` content than in core installation logic.

**Dependency:** Depends on N11 (Standards-as-Code) for the standards to be meaningful.

**Rating:** **P2 / M**
- **Priority P2** — Valuable but not urgent. The gate is already defined; the gap is enforcement, which is inherently project-specific.
- **Effort M** — Requires creating observability standards, CI templates, and potentially lint rule templates.
- **Recommendation:** Create `specs/standards/observability/observability-patterns.md` as a standard. Add CI template additions. Do NOT attempt automated enforcement — that's a downstream concern. This is primarily a content addition (N11 — Standards-as-Code), not a new feature.

---

### N4: Coverage Thresholds

**What it is:** The quality gates document specifies "85% client / 90% server" coverage thresholds, but root AGENTS.md still contains `[YOUR_COVERAGE_THRESHOLD]%` placeholder. The gap is that the installation process doesn't capture and propagate specific coverage thresholds.

**Current state in ai-setup:**
- `quality-gates.xml` mentions 85%/90% thresholds.
- `stackDetection` already detects test frameworks.
- The placeholder in AGENTS.md is unresolved during installation.
- The constitution template includes coverage thresholds but as placeholders.

**Integration complexity:** Trivial. This is a placeholder-resolution problem. The stack detection already knows the test framework; it just needs to surface a configurable threshold.

**Installation impact:**
1. Add a wizard question in Phase 2: "What coverage threshold do you enforce?" with sensible defaults per stack.
2. Resolve `[YOUR_COVERAGE_THRESHOLD]%` during compilation.
3. Add to `specs/rules/testing.md` if selected.

**User value:** Useful. Reduces manual placeholder editing. Every project needs this configured; making it part of installation eliminates a friction point.

**Alignment:** Core. This is a project configuration parameter that installation should capture.

**Dependency:** Self-contained, but related to C8 (Constitution Population).

**Rating:** **P1 / XS**
- **Priority P1** — Low effort, fixes a real placeholder gap, improves first-run experience.
- **Effort XS** — A wizard question + variable resolution in the template compiler.
- **Recommendation:** Add a wizard question for coverage threshold with sensible defaults. Resolve the placeholder during compilation. This is a quick win that closes a real gap.

---

### N5: API Cost Monitoring

**What it is:** The `specs/rules/cost.md` rule documents token budget tracking and alerts, but there is no actual integration with API/model spend data. No real-time cost tracking against budgets.

**Current state in ai-setup:**
- `library/rules/cost.md` defines budget gates (70%/85%/95%) and tracking rules.
- The orchestrator's `get_budget` tool tracks estimated spend per chain/team/workflow run.
- No integration with actual provider billing APIs (OpenAI, Anthropic, etc.).
- Budget tracking uses estimated token counts, not real costs.

**Integration complexity:** Very-High. Real API cost monitoring requires:
1. Per-provider API integrations (OpenAI usage API, Anthropic billing API, etc.).
2. API key management for billing endpoints.
3. Cost aggregation and alerting infrastructure.
4. Mapping between orchestrator runs and provider billing records.

This is fundamentally an infrastructure/operations concern, not an installation concern.

**Installation impact:** Minimal — the installation can scaffold cost rules and budget gates, but cannot provide a running cost monitoring service.

**User value:** High-Value for cost-conscious teams, but impossible to deliver through an installation scaffold alone.

**Alignment:** Misaligned. ai-setup is a project-scaffolding tool. Real-time cost monitoring requires a running service with provider integrations. The orchestrator's `get_budget` is the right level of abstraction for ai-setup; deeper integration belongs in a separate tool or service.

**Dependency:** Heavily dependent on external services and provider APIs.

**Rating:** **P3 / XL**
- **Priority P3** — Not appropriate for ai-setup. The scaffolding provides budget rules; actual cost monitoring is a separate operational concern.
- **Effort XL** — Would require a new service, provider integrations, and billing infrastructure.
- **Recommendation:** Do NOT integrate into ai-setup installation. The orchestrator's budget estimation is sufficient for installation purposes. If cost monitoring is needed, build it as a separate observability service that integrates with the orchestrator's run events.

---

### N6: Cross-Repo Workflows

**What it is:** The workspace protocol defines cross-repo workflow patterns (`workspace-protocol.md`), but there is no strong evidence of active cross-repo workflow runs in practice.

**Current state in ai-setup:**
- `library/fragments/workspace-protocol.md` defines the pattern.
- Workspace scope (`--scope workspace`) creates per-repo ledgers and state files.
- The RPI skill declares `cross_repo: true`.
- No actual multi-repo chain/team execution evidence in the catalog.

**Integration complexity:** Medium. The infrastructure exists (workspace scope, per-repo ledgers). The gap is making cross-repo execution a first-class orchestration concept with real chain/team definitions that span repos.

**Installation impact:**
1. Add cross-repo chain/team definitions to `library/orchestration/`.
2. Enhance workspace wizard to capture cross-repo dependency relationships.
3. Add cross-repo task coordination to the feature chain.

**User value:** High-Value for teams with microservice architectures or monorepos. Transformative for the specific use case, but only relevant for multi-repo projects.

**Alignment:** Good-Fit. The workspace protocol already exists; this is about making it operational through orchestration definitions and installation guidance.

**Dependency:** Self-contained — the workspace scope is already implemented.

**Rating:** **P2 / M**
- **Priority P2** — Important for multi-repo teams but not universally needed. The foundation is there; this is about shipping orchestration content.
- **Effort M** — New orchestration definitions + wizard enhancements.
- **Recommendation:** Create a `cross-repo-feature` chain and `cross-repo-review` team. Add a workspace wizard question about cross-repo dependency types. This makes the existing workspace protocol operational without changing the installation architecture.

---

### N7: Compaction Automation

**What it is:** The compaction protocol is defined (70%/85%/95% gates, 15-20 exchange trigger, what to preserve/drop), but it is advisory only — no automated enforcement. Agents can ignore compaction guidance and continue with bloated context.

**Current state in ai-setup:**
- `specs/rules/workflow.md` and `AGENTS.md` define the compaction protocol.
- `library/prompts/compact.md` provides the prompt format for compaction.
- No mechanism to detect when an agent has exceeded budget and force compaction.
- No tool-enforced context window monitoring.

**Integration complexity:** High. Automated compaction enforcement requires either:
1. Tool-level integration (CLI tools must expose context window usage) — not something ai-setup controls.
2. Agent-level self-monitoring (agents track their own token usage) — unreliable.
3. MCP-level monitoring (orchestrator tracks tool calls and estimates token usage) — possible but complex.

**Installation impact:** The installation can provide stronger guidance and tools (e.g., a compaction checklist skill, token budget alerts), but cannot force tool-level enforcement without deeper CLI integration.

**User value:** Useful but limited. The protocol is already well-documented; the gap is enforcement, which requires tool-level changes ai-setup cannot control.

**Alignment:** Good-Fit (Claude Code only, tool-asymmetric). Compaction enforcement is a CLI tool concern, but Claude Code stop hooks provide a concrete install path.

**Dependency:** Claude Code only — OpenCode and Copilot have no equivalents.

**Rating:** **P2 / S (reconciled from P3)**
- **Priority P2** — OPT-IN. Claude Code stop hooks provide a concrete integration path (`.claude/settings.json` compaction hook). Ship as a feature flag; documented as tool-asymmetric.
- **Effort S** — A `library/hooks/claude-code/compaction.json` snippet emitted into `.claude/settings.json` when `compactionHook` feature flag is enabled.
- **Recommendation:** Ship as OPT-IN feature flag. Claude Code only. Document the gap for OpenCode/Copilot. This was reconciled from P3 after Claude's analysis found a concrete install path I missed.

---

### N8: Constitution Population

**What it is:** The root AGENTS.md still contains placeholders like `[YOUR_PROJECT_OVERVIEW]`, `[YOUR_LANGUAGE]`, `[YOUR_TEST_FRAMEWORK]`. The constitution template exists but actual population during installation is incomplete.

**Current state in ai-setup:**
- Stack detection (`detectProjectStack`) auto-detects language, framework, package manager, test framework.
- `scaffoldCompiledRoot()` uses detected stack info in template variables.
- Some placeholders ARE resolved (e.g., `{{PROJECT_NAME}}`, `{{BUILD_COMMAND}}`).
- Many placeholders are NOT resolved: project overview, codebase map, naming conventions, error handling conventions, API conventions, import conventions, test command, lint command, coverage threshold, protected branch.

**Integration complexity:** Medium. Some placeholders can be auto-detected (test command, lint command, coverage threshold). Others require human input (project overview, conventions). The installation wizard already captures some of this but doesn't feed it all into the template.

**Installation impact:**
1. Expand Phase 2 wizard to capture: project overview, naming conventions, error handling conventions, protected branch.
2. Auto-detect and resolve: test command, lint command, build command, coverage threshold, stack details.
3. Auto-generate a codebase map from the project structure.
4. Feed all of this into the template compiler's FragmentContext.

**User value:** High-Value. This is the most impactful quick-win in the entire review. Every ai-setup adopter must manually fill these placeholders today, which weakens the harness from day one.

**Alignment:** Core. The root contract is the foundation of the entire harness. Placeholders create ambiguity that cascades into every agent interaction.

**Dependency:** Self-contained. N4 (Coverage Thresholds) is a subset of this.

**Rating:** **P0 / M**
- **Priority P0** — Foundational gap. An incomplete root contract weakens every downstream agent behavior. This should be done before shipping any other feature.
- **Effort M** — Requires expanding the wizard, adding auto-detection for more fields, and wiring everything into the template compiler. Multiple files but straightforward changes.
- **Recommendation:** This is the #1 priority. Expand the Phase 2 wizard to capture all remaining placeholders. Add auto-detection for commands. Generate codebase map from project structure. The installation should produce a fully populated AGENTS.md with zero remaining `[YOUR_*]` placeholders.

---

### N9: Knowledge Graph Integration

**What it is:** CodeGraph, QMD, Memoria, graphify, and knowledge-stack tools/skills are available, but knowledge graphs are not automatically updated or injected into every workflow. There's no "pre-flight context retrieval."

**Current state in ai-setup:**
- 5 knowledge tools are enabled by default in the "recommended" MCP preset: codegraph, qmd, graphify, obsidian, memoria.
- The `knowledge-stack` skill teaches agents which tool to use for which question.
- The orchestrator's `runBootstrap()` can detect qmd/codegraph drift but doesn't auto-sync.
- No automatic injection of knowledge graph results into agent prompts at chain start.

**Integration complexity:** Medium-High. Making this automatic requires:
1. A pre-flight hook in the orchestrator that runs before every chain start.
2. Querying CodeGraph for relevant symbols for the task.
3. Querying QMD for relevant documentation.
4. Summarizing results into a context packet.
5. Injecting that packet into the root layer of agent composition.

**Installation impact:**
1. Enhance `runBootstrap()` to actually run queries instead of just detecting drift.
2. Add a "retrieval preflight" step to the orchestrator's chain startup.
3. Potentially add a feature flag `knowledgeRetrieval` to control this behavior.
4. Could be an MCP server that wraps codegraph+qmd into a single "retrieve context" tool.

**User value:** High-Value. This is the review's #3 recommendation. Automatic context retrieval would dramatically reduce manual context loading and improve task quality.

**Alignment:** Core. This directly improves the quality of agent work by providing relevant context automatically. It's one of the most impactful enhancements possible.

**Dependency:** Self-contained — the tools exist, just need wiring.

**Rating:** **P1 / L**
- **Priority P1** — High value, well-aligned, transformative for agent quality. The tools and infrastructure exist; the gap is orchestration.
- **Effort L** — Requires significant orchestrator runtime changes: new bootstrap logic, context packet generation, agent composition layer modifications. This is a feature addition to the orchestrator MCP server, not just content.
- **Recommendation:** This is the #1 runtime enhancement. Build a "retrieval preflight" into the orchestrator that automatically queries codegraph/qmd for relevant context before each chain starts. The installation CLI doesn't change much — this is orchestrator runtime work.

---

### N10: Agent Progression Levels

**What it is:** The L1-L4 supervision model (L1=full autonomy, L2=confirm before action, L3=propose only, L4=human-driven) is documented but has no enforcement or tracking mechanism. Agents don't know their level and can't self-regulate.

**Current state in ai-setup:**
- The orchestrator's mode skills implement a similar concept: `autonomous` (minimal approval), `senior` (normal approval), `junior` (strict approval).
- The composer's `approvalPolicy` cascading (strictest wins) provides dynamic control.
- No explicit L1-L4 progression tracking across sessions.
- No mechanism to advance an agent from L3 to L2 based on performance.

**Integration complexity:** Medium. The mode skill system already provides the infrastructure. Adding explicit progression levels would mean:
1. Mapping L1-L4 to approval policies and model hints.
2. Adding a tracking mechanism (which agents have demonstrated L2 capability?).
3. Adding a promotion/demotion mechanism based on review outcomes.

**Installation impact:**
1. Add a `progression` configuration to `.ai-setup.json` or the TOML config.
2. Enhance mode skills with explicit L1-L4 definitions.
3. Add tracking of agent performance in the ledger.

**User value:** Useful but niche. Most teams won't use formal progression tracking. The mode skill system (autonomous/senior/junior) covers 90% of the use case.

**Alignment:** Peripheral. The mode skill system is the right abstraction for ai-setup. Explicit L1-L4 progression is an operational concern that few teams will implement.

**Dependency:** Depends on E11 (Structured Evaluation Benchmark) for meaningful progression decisions.

**Rating:** **P3 / M**
- **Priority P3** — The mode skill system already provides the essential capability. Formal progression is a niche need.
- **Effort M** — Would require new tracking infrastructure and mode skill refinement.
- **Recommendation:** Do NOT add explicit L1-L4 progression. Enhance the mode skill descriptions to include more granular behavior guidance within each level. The autonomous/senior/junior trinary is sufficient for installation purposes.

---

### N11: Standards-as-Code

**What it is:** The `specs/standards/` directory exists and rules exist, but the standards directory is mostly empty. The review recommends creating concrete standards: orchestration patterns, test patterns, error handling, agent security, context loading. Gate 4 (pattern consistency) depends on these standards existing — without them, pattern review is subjective.

**Current state in ai-setup:**
- `specs/standards/` directory is scaffolded during installation but empty.
- `specs/rules/` has 9-11 rule files depending on preset.
- The `extract-standards` skill can extract standards from code, but it's reactive.
- Quality Gate 4 references "pattern consistency" but has nothing to check against.

**Integration complexity:** Low-Medium. This is content creation, not system changes. The hardest part is writing good initial standards that are generic enough to be useful defaults but specific enough to be actionable.

**Installation impact:**
1. Create 5+ standard files in `library/standards/` (new directory).
2. Scaffold them into `specs/standards/` during installation.
3. Wire them into quality gate checks and review workflows.
4. Potentially add a feature flag `standardsEnforcement` or fold into existing `qualityGates`.

**User value:** High-Value. Empty standards make Gate 4 meaningless. Populated standards give agents concrete patterns to enforce, making reviews more consistent and objective.

**Alignment:** Core. Gate 4 depends on this. Without standards, the quality ladder has a broken rung.

**Dependency:** Self-contained.

**Rating:** **P1 / S**
- **Priority P1** — Critical for making the quality gates functional. A broken gate undermines the entire quality ladder.
- **Effort S** — Content creation. No system changes needed; the scaffolding already supports standards.
- **Recommendation:** Create an initial set of 5-8 standard files in `library/standards/` covering the most important patterns. Ship as part of the standard/full presets. The review's recommendations (orchestration patterns, test patterns, error handling, agent security, context loading) are a good starting set.

---

### N12: TillDone Protocol

**What it is:** The workflow rules define a "TillDone" protocol: agents should not stop early; they should continue working until all acceptance criteria are met or they're blocked. Currently there is no automated detection if agents stop early.

**Current state in ai-setup:**
- The implementor agent's instructions say "execute exactly ONE task per session" but don't enforce continuation until done.
- The tdd-loop skill drives RED → GREEN → REFACTOR cycles but doesn't have a "continue until done" enforcement.
- No mechanism to detect that an agent stopped with incomplete work and prompt it to continue.

**Integration complexity:** Medium. Detecting "done vs. stopped early" is challenging because it requires understanding of acceptance criteria completeness. Basic detection could check:
1. All declared tasks have output files.
2. All tests pass.
3. No remaining TODO/FIXME markers in changed files.

**Installation impact:**
1. Add "TillDone" rules to `specs/rules/workflow.md`.
2. Add a till-done verification step to the review chain.
3. Enhance the implementor agent's instructions with explicit continuation criteria.

**User value:** Useful. Early-stopping is a common failure mode for AI agents. Automated detection would reduce incomplete work slipping through.

**Alignment:** Good-Fit. This is a quality enforcement mechanism that belongs in the agent instructions and chain definitions.

**Dependency:** Self-contained.

**Rating:** **P2 / S**
- **Priority P2** — Useful quality improvement but not critical. Review catches most early-stop issues.
- **Effort S** — Primarily instruction changes + a verification step addition.
- **Recommendation:** Add a "TillDone" check to the implementor agent's instructions and the feature chain's review step. Add a rule in `specs/rules/workflow.md`. This is primarily a prompt engineering change.

---

### N13: Multi-CLI Support

**What it is:** Plans and research exist for Claude Code, Copilot, Gemini, Codex, and OpenCode. OpenCode appears primary; others are partial or work-in-progress. The review flags that multi-CLI support is incomplete.

**Current state in ai-setup:**
- Three tool adapters are implemented: OpenCode, Claude Code, Copilot.
- Each has full agent/skill/prompt installation.
- MCP compilation is per-tool.
- The `compile` and `update` commands work across all configured tools.
- No Gemini or Codex adapters exist.
- Copilot support is described as "partial/WIP" in some areas.

**Integration complexity:** High for new adapters. Each new CLI tool requires:
1. Understanding its agent/prompt/skill format.
2. Understanding its MCP configuration format.
3. Writing an adapter with install/compile/setup methods.
4. Testing across the full installation pipeline.

**Installation impact:** New adapters only — the tool abstraction layer already exists.

**User value:** Varies by tool. Adding more adapter support broadens ai-setup's reach but each tool has a different user base.

**Alignment:** Core. Multi-CLI support is a core value proposition of ai-setup — "configure once, work everywhere."

**Dependency:** Self-contained per adapter.

**Rating:** **P3 / SKIP (reconciled from P2 — strategic scope reduction)**
- **Priority P3/SKIP** — The team deliberately removed Codex/Gemini adapters in commit `8c6864b refactor: reduce supported AI CLI tools`. Adding new adapters is a strategic choice, not a technique-coverage decision.
- **Effort L** — Each new adapter is significant work, but the strategic decision has already been made.
- **Recommendation:** SKIP. Out of scope unless strategy changes. The existing 3 adapters (OpenCode, Claude Code, Copilot) are the supported set.

---

## 3. DO NOT HAVE — Analysis (18 techniques)

These are established AI/agent techniques that are not meaningfully implemented in the current state.

---

### D1: Retrieval-Augmented Generation (RAG)

**What it is:** An automatic retrieval pipeline that queries codebases, documentation, and standards for relevant context, then injects it into agent prompts at startup. Goes beyond manual tool use — this is automatic, contextual, and pre-flight.

**Current state in ai-setup:**
- Tools exist: codegraph, qmd, memoria, knowledge-stack.
- No automatic retrieval pipeline.
- The review's #3 recommendation: "Wire retrieval into agent startup."
- The orchestrator's `runBootstrap()` could be enhanced for this.

**Integration complexity:** High. RAG requires:
1. A query generation step (convert the task into search queries).
2. Multi-source retrieval (codegraph for symbols, qmd for docs, memoria for history).
3. Result ranking and deduplication.
4. Context packet assembly (summarize into a tight injection).
5. Injection into the agent composition pipeline.

**Installation impact:**
1. Orchestrator runtime enhancement (add RAG preflight to chain startup).
2. New MCP tool: `retrieve_context` that wraps codegraph+qmd+memoria.
3. Feature flag: `ragPreflight` to control this behavior.
4. Installation might auto-enable the needed MCP servers when RAG is selected.

**User value:** Transformative. Automatic context retrieval is one of the most impactful enhancements possible. It addresses the biggest source of agent quality degradation: missing context.

**Alignment:** Core. This is the natural evolution of the knowledge stack tools into an automatic system.

**Dependency:** Builds on N9 (Knowledge Graph Integration). The tools exist; this is the automation layer.

**Rating:** **P1 / L**
- **Priority P1** — High value, transformative for agent quality. The review identified this as a top recommendation.
- **Effort L** — Major orchestrator runtime addition. Requires query generation, multi-source retrieval, ranking, and composition integration.
- **Recommendation:** This is closely related to N9 (Knowledge Graph Integration) — essentially N9 is "make knowledge graphs auto-inject" and D1 is "build the retrieval pipeline." Implement together as a single RAG feature. Start with a simple codegraph+qmd query and grow from there.

---

### D2: Automated Model Selection/Routing

**What it is:** Automatically select which model to use for a task based on complexity, domain, cost sensitivity, and performance requirements. Simple tasks use cheap/fast models; complex tasks use expensive/smart models.

**Current state in ai-setup:**
- The orchestrator's composer has a model resolution cascade (step → mode → domain → base → fallback).
- Mode skills hint at models (senior→opus, autonomous→sonnet).
- No automatic task complexity analysis that routes to the right model.
- Cost rules are advisory, not automated routing rules.

**Integration complexity:** High. Automated routing requires:
1. Task complexity assessment (token count? number of files? domain?).
2. Model capability/cost matrix.
3. Routing rules that can be configured per project.
4. Integration with the composer's model resolution to override based on routing.

**Installation impact:**
1. Add a `modelRouter` configuration in `.ai-setup.json` or TOML config.
2. Enhance the orchestrator composer to accept routing decisions.
3. Add a task complexity heuristic (simple heuristic to start, ML later).

**User value:** High-Value for cost-conscious teams. Reduces spend by 30-60% by using cheap models for simple tasks while preserving quality for complex work.

**Alignment:** Good-Fit. Model routing is a natural orchestrator responsibility — the orchestrator already controls which agent+model handles each step.

**Dependency:** Self-contained, though made more powerful with cost data (N5) and evaluation benchmarks (E11).

**Rating:** **P2 / DEFER (tool constraint noted)**
- **Priority P2** — Valuable but constrained by tool capabilities. The installer can ship a `library/rules/model-routing.md` that recommends mappings, but actual routing happens in the host CLI tool (OpenCode supports per-agent overrides; Claude Code uses session-level models).
- **Effort L** — Complexity assessment logic + configuration. But the scaffolding portion is much simpler (rule doc only).
- **Recommendation:** Document recommended model-to-task mappings in a `model-routing` rule. Real routing belongs in the host CLI tool, not in the installer. This was independently confirmed by Claude's analysis.

---

### D3: Multi-Agent Debate/Consensus

**What it is:** When agents disagree (e.g., reviewer says "blocking" but implementor disagrees), a structured debate protocol resolves the disagreement through adversarial argumentation and consensus building. Goes beyond simple review/pass to full debate.

**Current state in ai-setup:**
- Reviewer and red-team exist as separate agents in review chains.
- The review team has parallel perspectives (correctness, security, quality) with orchestrated synthesis.
- No formal debate protocol — disagreements are synthesized by the orchestrator without debate.
- No mechanism for agents to challenge each other's findings and converge.

**Integration complexity:** High. Debate requires:
1. A debate protocol (structured turns, argument format, evidence requirements).
2. A debate moderator (new orchestrator capability or new agent role).
3. Convergence criteria (when is debate resolved?).
4. Escalation path when consensus cannot be reached.

**Installation impact:**
1. New chain: `debate` — structured debate workflow.
2. New agent: `moderator` or enhancement to orchestrator.
3. Integration into review chains as an optional escalation path.

**User value:** Useful but limited. Multi-agent review already provides diverse perspectives. Formal debate adds structure but most disagreements are resolvable through the existing review → fix cycle.

**Alignment:** Good-Fit for high-stakes decisions but niche. Most projects won't need formal debate.

**Dependency:** Self-contained but benefits from E11 (Evaluation Benchmark) for measuring debate quality.

**Rating:** **P2 / S (reconciled from P3)**
- **Priority P2** — OPT-IN. Low-effort workflow to ship (2 days). The existing review team covers 90% of cases; debate is for high-stakes architecture disagreements. Off by default; enabled via feature flag.
- **Effort S** — A `debate` workflow under `.ai/orchestration/workflows/` with three rounds: argue, counter, synthesize. Reuses existing reviewer, red-team, and planner agents.
- **Recommendation:** Ship as OPT-IN (off by default, even in full preset). The existing review team with orchestrator synthesis is sufficient for most cases. Debate provides structure for high-stakes disagreements when a team specifically wants it.

---

### D4: Chain-of-Verification

**What it is:** A verifier that runs across the entire workflow artifact chain (spec → plan → tasks → code → tests) checking for consistency, traceability, and completeness. Not per-artifact checks (which exist), but chain-wide cross-artifact verification.

**Current state in ai-setup:**
- Per-artifact checks exist: speckit-analyze checks spec→tasks traceability.
- Quality gates check per-artifact completeness.
- No chain-wide verifier that checks plan↔spec consistency, tasks↔plan completeness, code↔tasks alignment.
- The RPI workflow's analyze phase is the closest, but it's spec→tasks only.

**Integration complexity:** Medium-High. Chain-of-verification requires:
1. Defined consistency rules between artifact types.
2. A verifier agent that reads all artifacts and checks cross-references.
3. Automated detection of: spec requirement not in plan, plan task not implemented, implemented feature not in spec.
4. A verification report with traceability matrix.

**Installation impact:**
1. New skill: `chain-verify` — cross-artifact consistency checking.
2. Integration into the RPI workflow as a gate after implementation.
3. Could be a new agent type: `verifier`.

**User value:** High-Value for formal workflows. Ensures nothing gets lost between phases. Critical for regulated industries or complex features.

**Alignment:** Good-Fit. This strengthens the RPI workflow's quality guarantees.

**Dependency:** Depends on having populated artifacts to verify — works with existing speckit outputs.

**Rating:** **P2 / M**
- **Priority P2** — Valuable for formal workflows but not universally needed. The speckit-analyze phase covers the most critical consistency check (spec→tasks).
- **Effort M** — New skill + integration into RPI workflow. Primarily content and chain definition changes.
- **Recommendation:** Create a `chain-verify` skill that checks cross-artifact consistency. Integrate into the RPI workflow as an optional phase after implementation. Focus on the highest-value checks: spec FRs → plan tasks → implementation → tests.

---

### D5: Automated Agentic Error Recovery

**What it is:** Agents autonomously recover from known failure patterns without waiting for human confirmation. The recovery protocol currently recommends waiting for human confirmation for most recovery actions. Automated recovery would handle common cases automatically.

**Current state in ai-setup:**
- Recovery patterns are defined: retry, fix-resume, escalate, handoff.
- The orchestrator reports all failures to the user before acting.
- "Recovery usually waits for human confirmation" — this is the identified gap.
- The chain-machine has retry logic (N attempts, then fallback).

**Integration complexity:** Medium. Automated recovery requires:
1. A classification of failures into "safe to auto-recover" vs. "needs human."
2. Recovery strategies per failure type.
3. Confidence thresholds for automated decisions.
4. Audit trail of automated recovery actions.

**Installation impact:**
1. Enhance the chain-machine with automated recovery classifications.
2. Add a feature flag: `autoRecovery` that controls automation level.
3. Add recovery policies to `.ai-setup.json` or TOML config.

**User value:** Useful. Reduces interruption for routine failures (transient tool errors, known edge cases) while preserving human gates for ambiguous situations.

**Alignment:** Good-Fit. The orchestration runtime is the right place for recovery logic.

**Dependency:** Self-contained — the recovery patterns and retry infrastructure exist.

**Rating:** **P2 / M**
- **Priority P2** — Valuable time-saver but needs careful safety boundaries. Getting recovery wrong can compound errors.
- **Effort M** — Classification logic + chain-machine enhancement.
- **Recommendation:** Start conservatively. Auto-recover only: transient tool errors (timeout, rate limit), known compilation/lint fix patterns, and test failures where the fix is a clear pattern. Everything else stays human-gated. The orchestrator should log all auto-recovery actions for audit.

---

### D6: Execution Plan Validation

**What it is:** An automated validator that checks plan completeness, feasibility, and consistency before implementation begins. The review notes that "plan approval is mostly human/manual" — there is no automated plan quality check.

**Current state in ai-setup:**
- The planner produces plans, and human approval is the gate.
- No automated checks: are all acceptance criteria covered? are all tasks scoped? are risks documented? are rollback paths defined?
- The speckit-analyze phase checks spec→tasks traceability but doesn't validate the plan itself.

**Integration complexity:** Low-Medium. Plan validation is rule-based checking:
1. Structural checks: all required sections present, all ACs have tasks.
2. Scope checks: no task touches >N files without justification.
3. Risk checks: every risk has a mitigation, every task has verification.
4. Completeness: every dependency is declared, every output is defined.

**Installation impact:**
1. New skill: `plan-validate` — automated plan quality scoring.
2. Integration into the feature chain before the plan approval gate.
3. The validator runs automatically and surfaces issues before asking for human approval.

**User value:** High-Value. Catches incomplete plans before they reach implementation, saving time and reducing rework. The review identifies this as a top recommendation (#4).

**Alignment:** Core. Plan quality is a critical gate that should be automated where possible. Human approval can focus on strategy and trade-offs, not completeness.

**Dependency:** Self-contained but relates to N8 (Constitution Population) — plans can be checked against a fully-populated constitution.

**Rating:** **P1 / S**
- **Priority P1** — High value, low complexity, directly improves a critical gate. The review's #4 recommendation.
- **Effort S** — Primarily a new skill with rule-based checks. No orchestrator runtime changes needed.
- **Recommendation:** Create a `plan-validate` skill that runs as an automated pre-gate before human plan approval. The skill checks: completeness, scope containment, risk coverage, test coverage, dependency declaration, and anti-speculation compliance. Results are surfaced alongside the plan for the human gate.

---

### D7: Guardrails/Railings System

**What it is:** A runtime system that enforces output/input safety constraints on generated code. Security rules are documented (`specs/rules/agent-security.md`, `specs/rules/security.md`) but there is no guardrails runtime that actively prevents violations.

**Current state in ai-setup:**
- Security rules exist as documentation.
- Agent instructions include security constraints.
- No runtime enforcement — agents can still generate code that violates security rules.
- The reviewer and red-team catch violations, but only after code is written.

**Integration complexity:** Very-High for runtime guardrails. A guardrails runtime requires:
1. Real-time code analysis during generation.
2. Pattern matching against known dangerous patterns.
3. Blocking or warning on violations before output is committed.
4. Integration with each CLI tool's generation pipeline.

This is fundamentally a tool-level concern, not an installation concern.

**Installation impact:** The installation can provide:
1. Stronger security rules with concrete violation patterns.
2. Pre-commit hooks that scan for security violations.
3. CI templates with security scanning.
But cannot provide real-time guardrails without CLI tool integration.

**User value:** High-Value for security-critical applications. But the value is in runtime enforcement, which ai-setup cannot provide through installation alone.

**Alignment:** Misaligned for real-time enforcement. Well-aligned for providing security scanning in CI/pre-commit hooks.

**Dependency:** Depends on CLI tool capabilities and external security scanning tools.

**Rating:** **P3 / XL (runtime), P2 / S (static)**
- **Priority P3** — Real-time guardrails require CLI tool integration beyond ai-setup's scope.
- **Priority P2** — Static guardrails (pre-commit hooks, CI templates) are achievable and useful.
- **Effort: XL (runtime) / S (static)**
- **Recommendation:** Do NOT pursue real-time guardrails. Instead, enhance the security infrastructure: add security-focused pre-commit hooks, CI security scanning templates, and stronger violation patterns in security rules. This provides guardrail value without requiring CLI tool integration.

---

### D8: Explicit Agent State Machine

**What it is:** Each agent has an explicit lifecycle state machine (idle → loading → ready → working → waiting → done → error) that is observable and recoverable. Currently, chain state exists but individual agent lifecycle is implicit.

**Current state in ai-setup:**
- Chain state machine: `created → running → gated → completed/abandoned/handoff`.
- Step states: `pending → running → completed → gated → failed`.
- No per-agent state tracking — the chain knows a step is "running" but not what state the agent is in.
- The orchestrator doesn't track whether an agent is loading context, thinking, generating code, or waiting for input.

**Integration complexity:** Medium-High. An agent state machine requires:
1. Agent lifecycle states defined per agent type.
2. State transitions triggered by MCP tool calls or agent output patterns.
3. Observable state exposed through `get_status`.
4. Recovery actions keyed to agent state (not just step state).

**Installation impact:**
1. Enhance the orchestrator's chain-machine to track per-step agent state.
2. Add agent state to `get_status` output.
3. Define standard agent lifecycle for all agent types.

**User value:** Useful for debugging and observability. Knowing "the agent is stuck in thinking state" vs. "the agent is generating code" changes how you intervene.

**Alignment:** Good-Fit. Agent state is a natural extension of the orchestrator's chain state tracking.

**Dependency:** Can be done independently; synergizes with streaming (E16) for real-time state visibility.

**Rating:** **P2 / M**
- **Priority P2** — Useful for observability but not critical. Chain state provides sufficient tracking for most use cases.
- **Effort M** — Chain-machine enhancement + status output changes.
- **Recommendation:** Add a lightweight agent state model (loading/thinking/generating/waiting/done/error) to the chain-machine. Expose through `get_status`. Keep it simple — don't model every micro-state, just the observable phases.

---

### D9: Structured Human-in-the-Loop Feedback

**What it is:** Human feedback during gates captures more than just approve/reject — it captures structured feedback (what specifically to change, why, priority) that agents can act on. Current gates are binary: approve or reject.

**Current state in ai-setup:**
- Chain gates: `user_approval` pauses the chain and waits for `advance_chain` with `outcome: "approved"` or `"rejected"`.
- The RPI workflow defines APPROVE / REQUEST_CHANGES outcomes.
- No structured feedback format — when a human rejects, there's no standard way to communicate what needs to change.
- The `advance_chain` tool accepts an `output` object but doesn't have a schema for feedback.

**Integration complexity:** Low. Structured feedback is a schema definition + UI enhancement, not an architectural change.

**Installation impact:**
1. Define a feedback schema: `{ verdict, required_changes: [...], suggestions: [...], priority, deadline }`.
2. Enhance the gate resolution in chain-machine to parse and propagate structured feedback.
3. Feed structured feedback into the next step's context (e.g., planner receives reviewer's specific change requests).

**User value:** Useful. Reduces ambiguity in human feedback and speeds up iteration cycles. "Rejected" without context is wasted time.

**Alignment:** Core. Human gates are a central part of the harness; structured feedback makes them more effective.

**Dependency:** Self-contained.

**Rating:** **P2 / S**
- **Priority P2** — Improves an existing feature without changing architecture. Low risk, clear benefit.
- **Effort S** — Schema definition + chain-machine enhancement.
- **Recommendation:** Define a structured feedback format for gate rejections. Enhance `advance_chain` to accept structured feedback in the output. Pass feedback into the next step's context. This is a prompt/contract enhancement, not new infrastructure.

---

### D10: Dynamic Prompt Optimization

**What it is:** Prompts and instructions automatically improve based on failures and success patterns. When an agent consistently fails on a specific type of task, the system adjusts the instructions. Prompts are currently static files; they don't learn from execution history.

**Current state in ai-setup:**
- `self-improve` skill exists for analyzing human interventions and improving library content.
- Memory-write captures lessons but doesn't auto-modify prompts.
- Prompts are static files in `library/prompts/` and `library/agents/`.
- No feedback loop from execution outcomes back to prompt content.

**Integration complexity:** High-Very High. Dynamic optimization requires:
1. Execution outcome tracking (success/failure per prompt version).
2. Pattern detection (what prompt patterns correlate with failures?).
3. Safe prompt modification (A/B testing, gradual rollout).
4. Rollback capability for bad prompt changes.

**Installation impact:** This is a meta-system that would require significant infrastructure. The installation could:
1. Track prompt versions and outcomes.
2. Provide a "suggest improvement" mechanism.
But true dynamic optimization requires ML-style infrastructure that is far beyond installation scope.

**User value:** High-Value in theory, but the risks of auto-modifying prompts (degrading quality, introducing subtle errors) are significant. Manual improvement cycles are safer for most use cases.

**Alignment:** Peripheral. The `self-improve` skill provides the right level of automation — human-in-the-loop improvement, not automatic optimization.

**Dependency:** Depends on E11 (Evaluation Benchmark) for measuring prompt quality. Depends on E18 (Continuous Learning) for the feedback data.

**Rating:** **P3 / XL**
- **Priority P3** — Risks outweigh benefits for an installation tool. Dynamic prompt optimization is a research problem, not an installation feature.
- **Effort XL** — Would require ML infrastructure, A/B testing, and safety mechanisms.
- **Recommendation:** Do NOT pursue dynamic prompt optimization. The `self-improve` skill + manual improvement cycles are the right approach for ai-setup. If optimization is needed, it should be a separate research project, not an installation feature.

---

### D11: Structured Evaluation Benchmark

**What it is:** A benchmark/evaluation suite that measures agent quality across sessions. Tracks metrics like: task completion rate, review pass rate, rework cycles, test coverage achieved, human intervention frequency. Currently there is no benchmark for agent performance — quality is assessed anecdotally.

**Current state in ai-setup:**
- The ledger tracks per-task outcomes but doesn't aggregate into quality metrics.
- No benchmark tasks with known correct outputs.
- No quality scoring per agent type, model, or prompt version.
- The orchestrator tracks chain completion but doesn't compute quality scores.

**Integration complexity:** Medium. A benchmark requires:
1. Standardized evaluation tasks with acceptance criteria.
2. Automated scoring based on code quality, test coverage, review findings.
3. Aggregation and trending over time.
4. Comparison across agents, models, and prompt versions.

**Installation impact:**
1. Add benchmark task definitions to the library.
2. Add an `eval` command to the CLI.
3. Add a `benchmark` skill that runs standard evaluations.
4. Track scores in the ledger.

**User value:** High-Value for teams iterating on their AI workflow. Without benchmarks, you can't tell if prompt changes improved quality or just changed output style.

**Alignment:** Good-Fit. Benchmark-driven improvement is a natural extension of the quality gates. ai-setup should be able to measure the quality of the workflows it creates.

**Dependency:** Depends on having concrete quality criteria defined. Synergizes with all quality-related techniques.

**Rating:** **P2 / L**
- **Priority P2** — Important for serious workflow optimization but not needed for initial adoption. Most teams won't use benchmarks until they've been using ai-setup for months.
- **Effort L** — Requires benchmark task design, scoring infrastructure, and CLI tooling.
- **Recommendation:** Design the benchmark architecture now but implement later. Start with a simple "quality score" derived from existing metrics (test pass rate, review findings, human interventions). Grow into formal benchmarks over time. The ledger already tracks the raw data.

---

### D12: Tool Creation by Agents

**What it is:** Agents can create bounded, safe tools for repetitive tasks they encounter. Instead of writing the same bash command 10 times, the agent creates a script/tool and reuses it. Currently, agents only use existing tools; they cannot create new ones.

**Current state in ai-setup:**
- The MCP server catalog is static — defined at installation time.
- Agents can use bash, write files, and run commands, which could theoretically create tools, but there's no structured mechanism.
- No "create tool" MCP capability.
- No sandboxing or safety review for agent-created tools.

**Integration complexity:** Very-High. Safe tool creation requires:
1. A tool creation protocol (define interface, implement, test, register).
2. Sandboxing (agent-created tools must have limited scope and permissions).
3. Safety review (human approval before a created tool is used).
4. Lifecycle management (created tools expire, are versioned, can be removed).

**Installation impact:** This would be a major new capability. The installation could:
1. Provide the tool creation framework.
2. Add safety gates around created tools.
But the core challenge (safe execution of agent-created code) is a deep security problem.

**User value:** Useful in principle but the safety risks are significant. Most repetitive tasks in software development don't benefit from ad-hoc tool creation — they benefit from better existing tools or scripts in the project.

**Alignment:** Misaligned. Tool creation by agents is a capability for advanced AI coding assistants, not for an installation scaffold. The risks (infinite loops, destructive operations, security bypass) exceed the benefits.

**Dependency:** Would require fundamental changes to the MCP architecture and security model.

**Rating:** **P3 / XL**
- **Priority P3** — Not appropriate for ai-setup. This is an AI coding assistant capability, not an installation feature.
- **Effort XL** — Requires sandboxing, security review, lifecycle management.
- **Recommendation:** Do NOT pursue agent-created tools. If specific repetitive patterns are identified, add them as first-class skills or MCP servers in the catalog. Agents should use well-tested, reviewed tools, not create ad-hoc ones.

---

### D13: Formal Causal Reasoning Framework

**What it is:** A rigorous diagnosis protocol for bugs — 5 Whys, fault-tree analysis, causal chain mapping. The bugfix chain has RCA-like phases but no formal causal reasoning protocol. Improves rigor of root cause analysis.

**Current state in ai-setup:**
- The bugfix chain has `reproduce → diagnose → fix → verify → document`.
- The `bugfix-rca-template.md` provides an RCA output format.
- No formal causal reasoning method (5 Whys, fault tree, Ishikawa diagram).
- Diagnosis is free-form rather than structured.

**Integration complexity:** Low. This is a prompt/methodology change — adding causal reasoning structure to the existing bugfix workflow.

**Installation impact:**
1. Add a `causal-reasoning` fragment or skill.
2. Enhance the bugfix chain's `diagnose` step with causal reasoning instructions.
3. Update the bugfix RCA template with causal reasoning sections.

**User value:** Useful. Improves bug diagnosis quality, especially for complex or intermittent bugs where root cause isn't obvious.

**Alignment:** Good-Fit. The bugfix workflow is a core ai-setup feature; better diagnosis methodology directly improves it.

**Dependency:** Self-contained.

**Rating:** **P2 / S**
- **Priority P2** — Valuable methodology enhancement but not critical. The existing bugfix workflow works for most bugs.
- **Effort S** — A new fragment + chain step enhancement.
- **Recommendation:** Create a `causal-reasoning.md` fragment with 5 Whys method, fault-tree template, and diagnostic checklist. Inject it into the bugfix chain's diagnose step. Update the RCA template. This is a content/prompt addition, not a system change.

---

### D14: Environment-Aware Planning

**What it is:** Plans incorporate operational constraints: token budget remaining, CI pipeline latency, model cost per task, context window limits. Currently, budget tracking is separate from planning — the planner doesn't know how much budget is left or how expensive each task will be.

**Current state in ai-setup:**
- Budget tracking exists in the orchestrator (`get_budget`).
- The planner produces plans without considering cost or environment constraints.
- No mechanism for the planner to say "this task sequence optimizes for budget" vs. "this optimizes for speed."
- Environment factors (CI time, context limits) are not surfaced to planning.

**Integration complexity:** Medium. Environment-aware planning requires:
1. Exposing budget and environment state to the planner agent.
2. Adding cost/speed trade-off dimensions to the plan template.
3. Letting the user choose optimization priorities (cost vs. speed vs. quality).

**Installation impact:**
1. Enhance the plan template with budget/environment sections.
2. Add environment context to the planner's agent prompt.
3. Add an optimization preference to the plan approval gate.

**User value:** Useful. Helps teams make informed trade-offs — "we have 100K tokens left, should we batch these tasks or run them sequentially?"

**Alignment:** Good-Fit. Planning + environment awareness is a natural quality improvement.

**Dependency:** Self-contained — the budget tracking already exists.

**Rating:** **P2 / S**
- **Priority P2** — Useful enhancement but not critical. Most teams don't need fine-grained cost optimization.
- **Effort S** — Template enhancement + prompt changes.
- **Recommendation:** Add a "Resource Constraints" section to the plan template. Enhance the planner agent with environment context. Include estimated token cost per task in task breakdowns. This makes cost visible in planning without requiring automated routing (D2).

---

### D15: Multi-Model Ensemble

**What it is:** For high-stakes decisions, run the same task through multiple different models and combine their outputs (voting, consensus, best-of-N). Goes beyond role-specialized agents — this is the same role, different models.

**Current state in ai-setup:**
- Role-specialized agents provide diverse perspectives (scout vs. reviewer vs. red-team).
- No mechanism to run the same agent+task through multiple models.
- The review team has parallel reviewers but they're the same agent type (reviewer) with different focuses, not different models.

**Integration complexity:** Medium. Multi-model ensemble requires:
1. A "run with N models" execution mode.
2. Output combination strategy (voting, confidence-weighted, best-of-N).
3. Cost management (running N models is N× cost).

**Installation impact:**
1. New team: `ensemble-team` — runs same task through N agents with different models.
2. Synthesizer that combines or selects the best output.
3. Cost controls — ensemble is expensive, needs explicit budget approval.

**User value:** Useful for specific high-stakes decisions (architecture design, security review, data migration). Not needed for routine tasks.

**Alignment:** Peripheral. The existing role-specialized approach provides diversity. Ensemble adds cost for marginal quality improvement.

**Dependency:** Depends on D2 (Model Selection) for automatic model diversity.

**Rating:** **P3 / M**
- **Priority P3** — Niche use case with significant cost. Role-specialization already provides diversity.
- **Effort M** — New team definition + synthesis logic.
- **Recommendation:** Do NOT implement as a general feature. If ensemble is needed, it can be implemented as a custom team where multiple instances of the same agent run in parallel. This is an advanced user pattern, not an installation feature.

---

### D16: Streaming/Progressive Output

**What it is:** Long-running agent tasks provide incremental visibility — progress updates, partial results, intermediate state. Currently, outputs are batch-style: the agent works silently until done, then returns the full result.

**Current state in ai-setup:**
- Chain/step states are updated when `advance_chain` is called.
- No intermediate progress events during a step's execution.
- The `subscribe_run` tool could support streaming but currently returns events at step boundaries.
- Agents output batch results, not progressive updates.

**Integration complexity:** Medium-High. Streaming requires:
1. Agent output streaming (depends on CLI tool capability).
2. Progressive state events from the orchestrator.
3. A notification/event protocol for intermediate progress.
4. Client support (CLI tools must display progressive output).

**Installation impact:**
1. Enhance the orchestrator's event bus with progressive events.
2. Define a streaming event protocol.
3. Integration depends on CLI tool support for streaming MCP responses.

**User value:** Useful for long-running tasks (large refactors, multi-file implementations). Reduces the "black box" feeling of agent execution.

**Alignment:** Peripheral. Streaming is primarily a CLI tool capability, not an installation feature. The orchestrator can emit events, but the tool must consume them.

**Dependency:** Depends heavily on CLI tool streaming support.

**Rating:** **P3 / L**
- **Priority P3** — Important user experience improvement but depends on tool capabilities ai-setup doesn't control.
- **Effort L** — Event protocol design + orchestrator changes + per-tool integration.
- **Recommendation:** Design the streaming event protocol now but implement when CLI tools support it. The orchestrator's event bus already exists; the gap is tool consumption. Add progressive events (e.g., `step.progress` with percentage, current file, message) to the event bus for future use.

---

### D17: Adversarial Self-Play During Design

**What it is:** During the design/specification phase, run adversarial analysis (red-team thinking) to find design flaws before implementation. Currently, red-team is post-implementation — they find bugs in code, not flaws in design.

**Current state in ai-setup:**
- Red-team is used in review chains (post-implementation) and assessment teams (architecture review).
- The speckit workflow has `specify → clarify → plan → tasks → analyze → implement` — no adversarial design step.
- The analyze phase checks consistency but doesn't attack the design.

**Integration complexity:** Low-Medium. This is a workflow addition — add an adversarial design review phase before implementation.

**Installation impact:**
1. Add an adversarial design analysis step to the RPI workflow.
2. Create an `adversarial-design` skill that red-teams the specification and plan.
3. Insert it between Plan and Tasks phases.

**User value:** High-Value. Finding design flaws before implementation is 10× cheaper than fixing them in code. The review identifies this as a gap.

**Alignment:** Core. This is a missing phase in the RPI workflow that directly improves quality.

**Dependency:** Self-contained but synergizes with D3 (Debate) if adversarial findings are disputed.

**Rating:** **P1 / S**
- **Priority P1** — High value, low complexity. Directly addresses a gap in the RPI workflow. Finding flaws before code is written is one of the highest-leverage improvements possible.
- **Effort S** — New skill + workflow phase addition.
- **Recommendation:** Create an `adversarial-design` skill that runs after the Plan phase. The red-team agent reviews the spec and plan for: missing edge cases, security flaws, scalability issues, abuse scenarios, assumption failures. Insert as an optional-but-recommended phase in the RPI workflow. Findings feed back into plan revision before tasks are created.

---

### D18: Continuous Learning from Human Corrections

**What it is:** When humans correct agent behavior, the corrections are automatically analyzed and turned into prompt/rule improvements. The `self-improve` skill exists but "extraction across sessions is not automated" — corrections from one session don't improve the next.

**Current state in ai-setup:**
- `self-improve` skill analyzes human interventions and suggests library improvements.
- `memory-write` captures lessons from individual sessions.
- No automated aggregation — each session's lessons are isolated.
- No mechanism to detect that the same correction has been made 5 times and automatically update the relevant rule.

**Integration complexity:** Medium. Continuous learning requires:
1. Correction tracking (what was corrected, by whom, when, in what context).
2. Pattern detection (same correction across multiple sessions).
3. Automatic suggestion or application of rule/prompt updates.
4. Safety review before changes go live (human approval gate).

**Installation impact:**
1. Enhance the ledger with correction entries.
2. Add a `learning` analysis that runs periodically.
3. Integrate with `self-improve` to automate suggestions.

**User value:** High-Value for teams using ai-setup long-term. The system gets better with use instead of stagnating at installation time.

**Alignment:** Core. This is the natural evolution of the self-improvement protocol — from manual to semi-automated.

**Dependency:** Depends on D9 (Structured Human Feedback) for structured correction data. Synergizes with D10 (Dynamic Prompt Optimization) but takes a safer approach.

**Rating:** **P2 / M**
- **Priority P2** — Important for long-term value but not needed for initial adoption. Teams need weeks/months of usage before the learning data is meaningful.
- **Effort M** — Ledger enhancement + pattern detection + integration with self-improve.
- **Recommendation:** Start with correction tracking in the ledger. Add a `learning-summary` command that surfaces repeated corrections. Keep the "apply improvement" step human-gated — the system suggests, the human approves. This is a safer approach than D10 (Dynamic Prompt Optimization) while still capturing the value of learning from corrections.

---

## 4. Summary Rankings

### Priority Ranking

| Priority | Techniques | Count |
|----------|-----------|-------|
| **P0** | N8 (Constitution Population) | 1 |
| **P1** | N4 (Coverage Thresholds), N9 (Knowledge Graph Integration), N11 (Standards-as-Code), D1 (RAG), D6 (Plan Validation), D17 (Adversarial Self-Play During Design) | 6 |
| **P2** | N1 (Few-Shot), N2 (Generated Knowledge), N3 (Observability Readiness), N6 (Cross-Repo Workflows), N7 (Compaction — OPT-IN), N12 (TillDone), D3 (Debate — OPT-IN), D4 (Chain-of-Verification), D5 (Auto Recovery), D7 (Guardrails-static), D8 (Agent State Machine), D9 (Structured Feedback), D13 (Causal Reasoning), D14 (Environment-Aware Planning), D18 (Continuous Learning) | 15 |
| **P3** | N5 (API Cost Monitoring), N10 (Progression Levels), D7 (Guardrails-runtime), D10 (Dynamic Optimization), D11 (Evaluation Benchmark), D12 (Tool Creation), D15 (Ensemble), D16 (Streaming) | 8 |

### Effort Distribution

| Effort | Techniques | Count |
|--------|-----------|-------|
| **XS** | N2 (Generated Knowledge), N4 (Coverage Thresholds) | 2 |
| **S** | N1 (Few-Shot-catalog), N7 (Compaction-hook), N11 (Standards-as-Code), N12 (TillDone), D3 (Debate-workflow), D6 (Plan Validation), D7 (Guardrails-static), D9 (Structured Feedback), D13 (Causal Reasoning), D14 (Environment-Aware Planning), D17 (Adversarial Design) | 11 |
| **M** | N3 (Observability), N6 (Cross-Repo), N8 (Constitution), N10 (Progression), D4 (Chain-Verify), D5 (Auto Recovery), D8 (Agent State), D15 (Ensemble), D18 (Continuous Learning) | 9 |
| **L** | N9 (Knowledge Graph), D1 (RAG), D2 (Model Selection), D11 (Benchmark), D16 (Streaming) | 5 |
| **XL** | N5 (Cost Monitoring), D7 (Guardrails-runtime), D10 (Dynamic Optimization), D12 (Tool Creation) | 4 |

### Quick Wins (P0-P1 + XS/S effort)

These are the techniques to implement first:
1. **N8 — Constitution Population** (P0/M): Fill placeholders in AGENTS.md during installation.
2. **N4 — Coverage Thresholds** (P1/XS): Wizard question + placeholder resolution.
3. **N11 — Standards-as-Code** (P1/S): Create 5-8 standard files.
4. **D6 — Plan Validation** (P1/S): Automated plan quality check before human approval.
5. **D17 — Adversarial Self-Play During Design** (P1/S): Red-team the spec and plan.

### Higher Effort but High Value

These are the biggest-impact but require significant work:
1. **N9 + D1 — Knowledge Graph Integration + RAG** (P1/L): Automatic context retrieval preflight. This is the review's #3 recommendation and the single most transformative runtime enhancement.
2. **D11 — Evaluation Benchmark** (P2/L): Measures agent quality. Critical infrastructure for all quality improvements.

### Misaligned / Out of Scope

These techniques don't belong in ai-setup installation:
1. **N5 — API Cost Monitoring** (P3/XL): Separate service, not installation.
2. **D7 — Guardrails runtime** (P3/XL): CLI tool concern.
3. **D10 — Dynamic Prompt Optimization** (P3/XL): Research problem, not installation.
4. **D12 — Tool Creation by Agents** (P3/XL): Advanced assistant capability, not installation.

---

## 5. Integration Map — Where Each Technique Fits

| Technique | Library Fragment | Feature Flag | New Skill | New Standard | New Rule | Orchestrator Runtime | Wizard Change | MCP Config |
|-----------|-----------------|--------------|-----------|-------------|----------|---------------------|---------------|------------|
| N1 Few-Shot | — | `fewShotPrompting` | `few-shot` | — | — | — | — | — |
| N2 Generated Knowledge | `reasoning-protocol.md` edit | — | — | — | — | — | — | — |
| N3 Observability | — | — | — | `observability-patterns.md` | — | — | — | — |
| N4 Coverage Thresholds | `root.template.md` edit | — | — | — | — | — | Phase 2 question | — |
| N5 Cost Monitoring | — | — | — | — | — | — | — | — |
| N6 Cross-Repo | — | — | — | — | — | New chains/teams | Workspace question | — |
| N7 Compaction | — | — | — | — | — | — | — | — |
| N8 Constitution | `root.template.md` edit | — | — | — | — | — | Phase 2 expansion | — |
| N9 Knowledge Graph | — | `knowledgeRetrieval` | — | — | — | `runBootstrap()` enhance | — | — |
| N10 Progression | — | — | — | — | — | — | — | — |
| N11 Standards | — | — | — | 5-8 files | — | — | — | — |
| N12 TillDone | `implementor.md` edit | — | — | — | `workflow.md` edit | — | — | — |
| N13 Multi-CLI | — | — | — | — | — | — | — | — |
| D1 RAG | — | `ragPreflight` | `retrieve-context` | — | — | Preflight hook | — | Auto-enable servers |
| D2 Model Selection | — | `modelRouting` | — | — | — | Composer enhance | — | — |
| D3 Debate | — | — | — | — | — | New chain | — | — |
| D4 Chain-Verify | — | `chainVerification` | `chain-verify` | — | — | — | — | — |
| D5 Auto Recovery | — | `autoRecovery` | — | — | — | Chain-machine enhance | — | — |
| D6 Plan Validation | — | — | `plan-validate` | — | — | — | — | — |
| D7 Guardrails-static | — | — | — | — | `security.md` edit | — | — | — |
| D8 Agent State Machine | — | — | — | — | — | Chain-machine enhance | — | — |
| D9 Structured Feedback | — | — | — | — | — | `advance_chain` enhance | — | — |
| D10 Dynamic Optimization | — | — | — | — | — | — | — | — |
| D11 Benchmark | — | — | `benchmark` | — | — | New CLI command | — | — |
| D12 Tool Creation | — | — | — | — | — | — | — | — |
| D13 Causal Reasoning | `causal-reasoning.md` | — | — | — | — | — | — | — |
| D14 Environment-Aware | `plan-template.md` edit | — | — | — | — | — | — | — |
| D15 Ensemble | — | — | — | — | — | New team | — | — |
| D16 Streaming | — | — | — | — | — | Event bus enhance | — | — |
| D17 Adversarial Design | — | `adversarialDesign` | `adversarial-design` | — | — | — | — | — |
| D18 Continuous Learning | — | — | — | — | — | Ledger enhance | — | — |

---

## 6. Recommended Implementation Sequence

### Wave 1: Fix the Foundation (P0-P1, XS-M effort)
1. **N8 — Constitution Population** (M): Complete the root contract.
2. **N4 — Coverage Thresholds** (XS): Wizard question + resolution.
3. **N11 — Standards-as-Code** (S): Populate specs/standards/.
4. **D6 — Plan Validation** (S): Automated plan quality checks.
5. **D17 — Adversarial Self-Play During Design** (S): Red-team the design.

### Wave 2: Enhance Quality and Context (P1-P2, M-L effort)
6. **D4 — Chain-of-Verification** (M): Cross-artifact consistency.
7. **N12 — TillDone Protocol** (S): Early-stop detection.
8. **N2 — Generated Knowledge** (XS): Enhance reasoning protocol.
9. **D9 — Structured Feedback** (S): Better human gates.
10. **D13 — Causal Reasoning** (S): Better bug diagnosis.
11. **D14 — Environment-Aware Planning** (S): Budget-aware plans.
12. **D5 — Auto Recovery** (M): Safe automated recovery.
13. **D8 — Agent State Machine** (M): Better observability.

### Wave 3: Retrieval and Intelligence (P1-P2, L effort)
14. **N9 + D1 — Knowledge Graph + RAG** (L): Automatic context retrieval.
15. **D2 — Model Selection** (L): Smart model routing.
16. **N1 — Few-Shot Prompting** (S): Systematic example catalog.
17. **N6 — Cross-Repo Workflows** (M): Operational multi-repo support.

### Wave 4: Learning and Measurement (P2, M-L effort)
18. **D18 — Continuous Learning** (M): Correction-driven improvement.
19. **D11 — Evaluation Benchmark** (L): Quality measurement.
20. **N13 — Multi-CLI Support** (L): New adapters as needed.

### Deferred / Not Recommended
- N3 (Observability — better as standard, not feature)
- N5 (Cost Monitoring — separate service)
- N10 (Progression — mode skills sufficient)
- D7 runtime (Guardrails — tool concern)
- D10 (Dynamic Optimization — research)
- D12 (Tool Creation — safety risk)
- D15 (Ensemble — cost/benefit poor)
- D16 (Streaming — tool dependent)

**Updated verdicts (cross-analysis reconciliation):**
- **N7 (Compaction)**: Promoted from P3/DEFER to **P2/OPT-IN** — Claude Code stop hooks provide a concrete install path (`.claude/settings.json` compaction hook). Documented as tool-asymmetric (Claude Code only).
- **D3 (Debate)**: Promoted from P3/DEFER to **P2/OPT-IN** — low-effort workflow to ship (2 days), off by default. The existing review team covers 90% of cases; debate is for high-stakes architecture disagreements.

---

## 7. Cross-Analysis Reconciliation (Claude Code Companion Analysis)

An independent analysis (`ai-techniques-additions-analysis.md`) was run against the same 31 techniques. Key differences were framed as 4 open questions. Answers are now locked into this document.

### Q1: A8 (Constitution) vs A11 (Standards) — which lands first?

**Answer: Constitution-first, standards-immediately-after.** Both analyses agree. Constitution is P0/ADD-NOW; standards are P1/ADD-NEXT. Neither blocks the other — they are separate workstreams that can overlap in the same sprint. But the constitution must be compiled into output AGENTS.md first: agents need to know stack, naming conventions, and project profile before they can meaningfully check standards.

### Q2: B1 RAG — default-on or opt-in?

**Answer: Opt-in via `retrievalPreflight` feature flag, enabled by standard and full presets, disabled in minimal.** Both analyses independently reached this. The flag provides discoverability (standard/full users get it automatically) while respecting cost-sensitive users (minimal users opt out).

### Q3: OPT-IN preset placement — `--preset full` only, `--features` only, or hybrid?

**Answer: Hybrid.** Every OPT-IN item is a named feature flag AND the standard/full presets turn on curated subsets. Casual users get sensible bundles via presets; power users get granular control via `--features`. Full preset does NOT enable everything — debate and auto-recovery stay off even in full.

### Q4: B7 Guardrails — portable MCP server or accept tool asymmetry?

**Answer: Accept asymmetry.** Ship Claude Code hooks as OPT-IN; document the gap for OpenCode/Copilot. Both analyses independently reached DEFER/SKIP. Do not build infrastructure without demonstrated demand. Revisit when a real customer asks.

### Cross-Analysis Verdict Map

| Verdict | My Count | Claude Count | Notes |
|---------|----------|-------------|-------|
| ADD-NOW (P0) | 2 | 2 | Constitution + Coverage — identical |
| ADD-NEXT (P1) | 8 | 7 | Claude rates Causal Reasoning as ADD-NEXT (I had P2) — adopted |
| OPT-IN (P2) | 9 | 9 | After reconciliation (Debate, Compaction moved in) |
| DEFER (P3) | 8 | 7 | Minor differences on specific items |
| SKIP | 6 | 7 | My Multi-CLI = P2/SKIP; Claude = SKIP (strategic) — adopted |

**Of 28 techniques compared, 26 had identical verdicts across both analyses.**

---

## 8. Open Questions for Plan Phase

> Flag granularity, preset assignment, and RAG architecture were resolved in Section 7 (Cross-Analysis Reconciliation). The remaining open questions are:

1. **Standards scope:** How many initial standards are enough? The review recommends 5; is that the right number, or should we start with 3-4 and grow?

2. **Backward compatibility:** How do techniques that modify compiled templates (N8, N4, D14) affect existing installations that `ai-setup update`?

3. **Extension interaction:** How do new feature flags interact with extensions? If an extension provides an agent but RAG is enabled, does the extension agent get RAG context injection?

---

## 9. Next Steps

1. **Resolved:** All 4 cross-analysis questions answered and locked in.
2. **Human gate:** Review the reconciled priority ratings and confirm the implementation sequence.
3. **Plan:** Proceed to plan phase with detailed implementation plan and task breakdowns.
4. **Implement:** Begin Wave 1 — fix the foundation.
