# AI Techniques — Additions Analysis (Research + Plan, no implementation)

**Companion to:** `ai-techniques-patterns-review.md`
**Question:** Which "Nice to Have" and "Do Not Have" techniques should `ai-setup` ship as part of its installation process (`init`, presets, library content, MCP catalog)?
**Date:** 2026-05-01
**Mode:** Research and rate. No code changes proposed in this pass.

---

## How "installing" works in this repo

Anything we add must land in one of these surfaces, otherwise it is not installable:

| Surface | What it is | Examples already shipped |
|---|---|---|
| **Library content** | Bundled `.md` artifacts copied into target repos | rules (`code-style`, `security`), skills (`plan`, `tdd-loop`), agents (`builder`, `red-team`), templates (`adr`, `spec`) |
| **Feature preset flag** | Toggle in `--preset minimal/standard/full` | `rpiWorkflow`, `qualityGates`, `chainOfThought` |
| **MCP server** | Pre-registered in `.ai/mcp.json` catalog | `memory`, `codegraph`, `qmd`, `orchestrator` |
| **Orchestration definition** | Chain / team / workflow / domain / mode under `.ai/orchestration/` | `feature` chain, `review-team`, `rpi` workflow |
| **Wizard prompt** | Question asked during `init` that drives `.ai/` content | scope, tools, preset, MCP server selection |
| **Compile-time emitter** | Code in `packages/ai-setup-ts` that writes tool-native files | per-tool MCP config, `AGENTS.md`, `.opencode/`, `.claude/` |

Anything that needs an active runtime not exposed by an MCP server, or that depends on telemetry/observability infra, is **out of scope** for the scaffolder — even if it is a valuable AI engineering technique.

---

## Rating system

| Symbol | Meaning |
|---|---|
| ⭐⭐⭐⭐⭐ | Foundational. Must ship — already underpins something the harness claims to do. |
| ⭐⭐⭐⭐ | High value, fits the scaffold model cleanly, ~1–3 days of work. |
| ⭐⭐⭐ | Worth shipping, packageable, but not load-bearing. Default to opt-in. |
| ⭐⭐ | Niche or tool-specific. Document, optional. |
| ⭐ | Out of scope for scaffolder; skip or punt to docs/ADR. |

**Verdict legend:** `ADD-NOW` (fix gap immediately) · `ADD-NEXT` (next preset bump) · `OPT-IN` (feature flag, off by default) · `DEFER` (revisit after current gaps closed) · `SKIP` (does not fit scaffold model)

**Rubric I applied to each item:**

1. **Scaffoldable?** Can it be expressed as a static artifact, MCP entry, or compile-time output?
2. **Portable?** Does it work across OpenCode + Claude Code + Copilot, or is it tool-specific?
3. **Useful in target projects?** Does the average team adopting `ai-setup` actually need this?
4. **Effort to ship?** Hours, days, weeks, or a research project?
5. **Risk of shipping it on by default?** Could it bloat installs or confuse users?

---

# A. Nice-to-Have (13 items) — currently lightly implemented

## A1. Few-Shot Prompting ⭐⭐⭐ — `OPT-IN`

**What it is.** Embedding 2–5 worked examples of the desired output style directly into a prompt or skill so the model imitates the format.

**Why it matters.** Few-shot is one of the highest-leverage techniques for narrow output formatting (handoffs, commit messages, ADR sections). Far more reliable than describing the format in prose.

**Install shape.** A new bundled directory `library/examples/` with categorised snippets (e.g., `examples/handoffs/`, `examples/commits/`, `examples/adr/`), and a small convention that skills can reference them with a short `<examples>…</examples>` block. The existing `specs/prompts/local-examples/` already hints at this pattern — formalize it.

**Effort.** ~1 day to define structure + curate 8–12 starter examples. Maintenance is the real cost (examples rot fast).

**Why not higher.** Hard to avoid example drift across language stacks. Better as an extension pattern teams populate themselves.

**Verdict: `OPT-IN`.** Ship the directory + a `library/examples/AGENTS.md` documenting the convention. Skills opt in by reference.

---

## A2. Generated Knowledge ⭐⭐ — `DEFER`

**What it is.** Asking the model to first emit relevant background facts/context, then answer using that scratchpad.

**Why it matters.** Useful for ambiguous tasks where the model would otherwise jump to code. In practice, the existing `research` skill and `rpi`'s SPECIFY phase already do this work, just framed differently.

**Install shape.** A `generate-knowledge` skill, or a single section added to `research`'s template.

**Effort.** Half a day. Low intrinsic complexity.

**Why low rating.** Largely redundant with `research`. The marginal benefit of a second skill is low and creates "which one do I run?" confusion.

**Verdict: `DEFER`.** Reconsider only if telemetry shows `research` is being skipped on ambiguous tasks.

---

## A3. Observability Readiness ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Quality Gate 5 — code must emit logs/metrics/traces appropriate to its layer. Currently advisory text in `quality-gates.md`.

**Why it matters.** Gate 5 is named in the harness but cannot be enforced without (a) a concrete checklist and (b) project-specific observability standards.

**Install shape.**
- A `library/standards/observability.md` template populated during `init` (with placeholders for log lib, metrics backend, trace lib).
- A `gate-5-observability` checklist sub-skill.
- Wizard prompt: "Logging library? Metrics backend? Trace exporter?" → fills the standards template.

**Effort.** 1–2 days including wizard hookups.

**Verdict: `ADD-NEXT`.** Pairs naturally with the constitution-population work.

---

## A4. Coverage Thresholds ⭐⭐⭐⭐⭐ — `ADD-NOW`

**What it is.** Concrete coverage numbers (e.g., 85% client / 90% server). Quality gates reference them; root template still has `[YOUR_COVERAGE_THRESHOLD]%`.

**Why it matters.** Unfilled placeholders propagate into compiled `AGENTS.md` files in target projects. Every Gate 4/5 check that reads "thresholds met" silently passes because there is no number to compare against.

**Install shape.**
- Wizard prompts during `init`: "Minimum coverage threshold?" (default 80%).
- TOML config key `coverage_threshold` for non-interactive runs.
- Replace placeholder in compiled root files.

**Effort.** ~2 hours. Trivial fix, biggest single quality-gate win.

**Verdict: `ADD-NOW`.** Subset of "fill the constitution" recommendation #1 in the review.

---

## A5. API Cost Monitoring ⭐⭐ — `SKIP`

**What it is.** Tracking model/API spend per task or session.

**Why it matters.** Useful for orgs paying for Claude/OpenAI tokens.

**Install shape.** Would need either a runtime hook or an MCP server that wraps requests. Neither tool the scaffolder targets exposes a clean injection point for cost telemetry today.

**Effort.** Weeks; effectively a new product.

**Verdict: `SKIP`.** Out of scope for scaffolder. Document in a rule (`cost.md` already exists) and leave actual tracking to platform tools.

---

## A6. Cross-Repo Workflows ⭐⭐⭐ — `OPT-IN` (already scaffolded, needs depth)

**What it is.** Coordinating work across multiple repos via a planning hub.

**Why it matters.** `workspace` scope already exists. The bones are there. What is missing is concrete **ledger sync, handoff templates between repos, and standards propagation**.

**Install shape.**
- Add `library/templates/cross-repo-handoff.md`.
- Bundle a `workspace-protocol` rule with explicit cross-repo ledger update steps.
- Add a `workspace` chain (e.g., propagate-standards) in `.ai/orchestration/chains/`.

**Effort.** 2–3 days.

**Verdict: `OPT-IN`.** Already the right shape, just under-populated. Improve as workspace adoption grows.

---

## A7. Compaction Automation ⭐⭐⭐ — `OPT-IN` (Claude Code only)

**What it is.** Automatic context compaction at 85%/95% thresholds.

**Why it matters.** Documented but advisory. Claude Code supports stop hooks; OpenCode and Copilot do not have equivalents.

**Install shape.**
- A `library/hooks/claude-code/compaction.json` snippet emitted into `.claude/settings.json` when a `compactionHook` feature flag is on.
- Documentation that the hook is Claude-specific.

**Effort.** ~1 day.

**Verdict: `OPT-IN`.** Tool-asymmetric — that argues against making it default. Ship as a feature flag.

---

## A8. Constitution Population ⭐⭐⭐⭐⭐ — `ADD-NOW`

**What it is.** Wizard fills `AGENTS.md` and `.ai/constitution/` with real values (stack, codebase map, naming, error handling, test/lint/build commands).

**Why it matters.** This is the #1 review recommendation. Every downstream gate degrades when these are placeholders. **Currently the root `AGENTS.md` template has 14+ unfilled fields**.

**Install shape.**
- Extend wizard with a "Project profile" step (~10 fields).
- Optional `--from-existing` mode: run `extract-standards` skill against the target repo to pre-fill defaults.
- TOML config schema for non-interactive runs (already partially supported via `project_name`, `default_tools`).

**Effort.** 3–5 days. The extract-standards integration is the long pole.

**Verdict: `ADD-NOW`.** Foundation for nearly every other improvement.

---

## A9. Knowledge Graph Integration ⭐⭐⭐ — `ADD-NEXT`

**What it is.** Auto-injection of `codegraph`, `qmd`, `memoria` results into agent context at task start.

**Why it matters.** All three MCPs are bundled but disabled by default and no skill calls them automatically. Their value is wasted unless integrated.

**Install shape.**
- A `preflight` skill (or extend `research`) that calls `codegraph_search` / `qmd_query` / `memoria_search_memories` for the active task and emits a context packet.
- Enable `memoria` by default (already is); enable `codegraph` when the wizard detects a supported language.

**Effort.** 2–3 days.

**Verdict: `ADD-NEXT`.** Same shape as the RAG recommendation (#3 in review). See B1.

---

## A10. Agent Progression Levels ⭐⭐ — `DEFER`

**What it is.** L1–L4 supervision model (junior → autonomous).

**Why it matters.** The library already ships `junior`, `senior`, `autonomous` mode skills. The L1–L4 ladder is largely redundant.

**Install shape.** Documenting how the existing modes map to L1–L4 in `library/modes/AGENTS.md`. Maybe a wizard prompt to set default mode per scope.

**Effort.** Half a day if reusing existing modes.

**Verdict: `DEFER`.** Repackaging what exists, not new capability.

---

## A11. Standards-as-Code ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Concrete `specs/standards/` content that Gate 4 (pattern consistency) actually checks against.

**Why it matters.** The review explicitly calls out the empty `specs/standards/` as the second-biggest gap. Without standards, Gate 4 is decorative.

**Install shape.**
- Pre-populate `library/standards/starter/` with five concrete standards: orchestration patterns, test patterns, error handling, security patterns, context loading.
- Wizard prompt: "Pre-populate starter standards? [Y/n]".
- Hook `extract-standards` skill into `init` for retroactive extraction.

**Effort.** 2–3 days for the starter set; ongoing for refinement.

**Verdict: `ADD-NEXT`.** Pairs with A8 (constitution).

---

## A12. TillDone Protocol ⭐⭐⭐ — `OPT-IN`

**What it is.** A rule that prevents agents from declaring "done" before all acceptance criteria are met.

**Why it matters.** Cheap insurance against partial implementations.

**Install shape.**
- A `library/rules/till-done.md` with the protocol stated.
- A small completion checklist sub-skill invoked at the end of `implement`.

**Effort.** ~1 day.

**Verdict: `OPT-IN`.** Not all teams want this rigidity. Ship as part of `qualityGates` preset.

---

## A13. Multi-CLI Support ⭐ — `SKIP` (deliberate scope reduction)

**What it is.** Adapters for Codex, Gemini CLI in addition to OpenCode / Claude Code / Copilot.

**Why it matters.** The team **explicitly removed** Codex/Gemini in commit `8c6864b refactor: reduce supported AI CLI tools…`. Reversing that is a strategic choice, not a technique-coverage decision.

**Verdict: `SKIP`.** Out of scope unless strategy changes.

---

# B. Do-Not-Have (18 items) — missing or not evident

## B1. Retrieval-Augmented Generation (RAG) ⭐⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Automatic retrieval of relevant project context (code symbols, docs, prior decisions) and injection into the agent prompt.

**Why it matters.** The review's #3 recommendation. The infrastructure (`codegraph`, `qmd`, `memoria`, knowledge-stack) is bundled — only the binding is missing. Probably the highest-leverage missing technique.

**Install shape.**
- A `library/skills/retrieval-preflight/SKILL.md` that runs first in any chain: classify task → query graph + memory + standards → emit a 200–500 token context packet.
- Wire `feature`, `bugfix`, `refactor` chains to call it as their first step.
- Feature flag: `retrievalPreflight`.

**Effort.** 3–5 days including chain wiring.

**Risks.** Token cost if poorly tuned; bad retrieval is worse than no retrieval.

**Verdict: `ADD-NEXT`.** Single biggest unlock once A8/A9/A11 land.

---

## B2. Automated Model Selection / Routing ⭐⭐ — `SKIP`

**What it is.** Choosing Opus vs Sonnet vs Haiku per task type to balance cost and quality.

**Why it matters.** Real money for active teams.

**Install shape.** Tool-specific. OpenCode supports model overrides per agent; Claude Code does not (model is session-level). Could ship a `library/rules/model-routing.md` that **recommends** mappings, but actual routing happens in the host CLI.

**Verdict: `SKIP`.** Document recommended mappings in the cost rule. Real routing belongs in the host tool.

---

## B3. Multi-Agent Debate / Consensus ⭐⭐⭐ — `OPT-IN`

**What it is.** Two or more agents argue opposing positions; a third synthesizes.

**Why it matters.** Useful for high-stakes architecture calls. Reviewer + red-team + planner already exist as roles — the missing piece is a **debate workflow** that runs them in opposition.

**Install shape.** A `debate` workflow under `.ai/orchestration/workflows/` with three rounds: argue, counter, synthesize.

**Effort.** ~2 days.

**Verdict: `OPT-IN`.** High value but rarely triggered. Off by default.

---

## B4. Chain-of-Verification (CoV) ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Cross-artifact consistency check (every FR in spec → task → test → code).

**Why it matters.** `speckit-analyze` already does part of this. CoV would extend it to a **running gate** that re-runs at each phase boundary, not only at ANALYZE.

**Install shape.**
- Extend `speckit-analyze` with a per-phase invocation, or split out a thin `cov-gate` skill that runs between every chain step.
- Output schema: traceability matrix.

**Effort.** 2–3 days.

**Verdict: `ADD-NEXT`.** Builds on existing skills.

---

## B5. Automated Agentic Error Recovery ⭐⭐ — `OPT-IN`

**What it is.** Agents recover from common failures (test fail, lint fail, type error) without waiting for human approval.

**Why it matters.** Faster iteration, lower friction.

**Install shape.** A `library/rules/auto-recovery.md` with a clear safe-action allowlist (e.g., re-run failed test, fix obvious type errors). Hooks would help on Claude Code.

**Risks.** Default-on auto-recovery violates the "confirm risky actions" stance in CLAUDE.md. Risk of compounding errors.

**Verdict: `OPT-IN`.** Ship as a feature flag, default off. Pair with `autonomous` mode.

---

## B6. Execution Plan Validation ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Validate a plan before implementation. Checks every AC has a task, every task has a verification, no speculative scope, risks/rollback documented.

**Why it matters.** The review's #4 recommendation. Concrete, packageable, prevents the most common failure mode (under-specified plans).

**Install shape.**
- A `library/skills/plan-validate/SKILL.md` with the checklist.
- Wired into `feature` chain between PLAN and TASKS.
- Output: PASS/HOLD/REWORK with itemized reasoning.

**Effort.** 1–2 days.

**Verdict: `ADD-NEXT`.** Easy win.

---

## B7. Guardrails / Railings System ⭐⭐ — `DEFER`

**What it is.** Runtime enforcement of input/output constraints (PII filters, prompt-injection detectors, tool-call validators).

**Why it matters.** Real safety value, but enforcement requires runtime hooks the scaffolder cannot uniformly provide.

**Install shape.** Claude Code hooks can do some; OpenCode and Copilot have less leverage. Could ship per-tool examples, but not a portable solution.

**Verdict: `DEFER`.** Document in `agent-security` rule. Reconsider as an opt-in MCP server (e.g., `guardrails-mcp`).

---

## B8. Explicit Agent State Machine ⭐⭐ — `DEFER`

**What it is.** Per-agent lifecycle states (idle, planning, executing, blocked, completed) — observable.

**Why it matters.** Mostly observability/debugging. Chain state already exists in the orchestrator runtime.

**Install shape.** Orchestrator package work; not scaffold work.

**Verdict: `DEFER`.** Belongs in `packages/orchestrator`, not in `init`.

---

## B9. Structured Human-in-the-Loop Feedback ⭐⭐⭐ — `OPT-IN`

**What it is.** When a human rejects/edits an agent output, capture **why** in structured form (not just APPROVE/REJECT).

**Why it matters.** Required substrate for any future continuous-learning loop (B18).

**Install shape.**
- A `library/templates/human-feedback.md` template.
- A small `human-feedback` skill that prompts the reviewer for: kept what / changed what / why.
- Stored in `specs/memory/feedback/`.

**Effort.** ~1 day.

**Verdict: `OPT-IN`.** Cheap to ship, opens future doors.

---

## B10. Dynamic Prompt Optimization ⭐ — `SKIP`

**What it is.** Prompts auto-improve from outcome telemetry.

**Why it matters.** Frontier capability, but needs telemetry pipeline + offline tuning + eval.

**Install shape.** Years of work, requires ML infra outside the scaffolder's scope.

**Verdict: `SKIP`.**

---

## B11. Structured Evaluation Benchmark ⭐⭐ — `DEFER`

**What it is.** A repeatable eval suite that scores agent outputs across versions of prompts/skills.

**Why it matters.** Without evals, "did the prompt change help?" is unanswerable.

**Install shape.** Could ship a starter `library/evals/` directory with 3–5 canned tasks and an `evaluate.md` skill.

**Effort.** 1–2 weeks to do well; less for a token-effort version.

**Verdict: `DEFER`.** Foundation worth building once core gaps close. Track as a separate initiative.

---

## B12. Tool Creation by Agents ⭐ — `SKIP`

**What it is.** Agents define their own bounded tools at runtime for repetitive work.

**Why it matters.** Powerful but unproven; high abuse surface.

**Verdict: `SKIP`.** Not maturity-appropriate as a default.

---

## B13. Formal Causal Reasoning Framework (5-Whys / fault-tree) ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Structured root-cause analysis (5-Whys, fishbone, fault-tree) instead of "first plausible cause".

**Why it matters.** The bugfix chain has RCA-like phases but no formal method. Concrete and packageable.

**Install shape.**
- A `library/skills/five-whys/SKILL.md` invoked during bugfix root-cause phase.
- Extend `bugfix-rca-template` with explicit 5-Why structure.

**Effort.** ~1 day.

**Verdict: `ADD-NEXT`.** Pairs naturally with the existing bugfix chain.

---

## B14. Environment-Aware Planning ⭐⭐ — `DEFER`

**What it is.** Plans incorporate token budget, CI latency, model cost, infra constraints.

**Why it matters.** Realistic plans need cost-awareness.

**Install shape.** Hard without runtime cost data. Could add cost-estimation columns to plan templates as a heuristic.

**Verdict: `DEFER`.** Wait until B11 (eval data) gives real numbers.

---

## B15. Multi-Model Ensemble ⭐ — `SKIP`

**What it is.** Same prompt across multiple models; pick majority answer.

**Why it matters.** Diversity reduces single-model bias.

**Install shape.** Requires runtime model dispatch the scaffolder doesn't own.

**Verdict: `SKIP`.**

---

## B16. Streaming / Progressive Output ⭐ — `SKIP`

**What it is.** Visible incremental output for long tasks.

**Why it matters.** UX, not technique.

**Install shape.** Tool-level concern (each CLI handles streaming itself).

**Verdict: `SKIP`.**

---

## B17. Adversarial Self-Play During Design ⭐⭐⭐⭐ — `ADD-NEXT`

**What it is.** Red-team the **plan**, not just the code — find design flaws before implementation cost.

**Why it matters.** The `red-team` agent already exists; today it is review-oriented. Extending it earlier is cheap and high-value.

**Install shape.**
- Add a "Red-team plan" step to the `feature` chain after PLAN, before TASKS.
- New `library/skills/red-team-plan/SKILL.md` invoking the existing red-team agent in plan-critique mode.
- Wire to a feature flag (`adversarialPlanning`).

**Effort.** ~1 day.

**Verdict: `ADD-NEXT`.** Reuses existing assets.

---

## B18. Continuous Learning from Human Corrections ⭐⭐⭐ — `OPT-IN` (post-B9)

**What it is.** Detect repeated human corrections and auto-suggest rule/standard updates.

**Why it matters.** Closes the loop between ad-hoc corrections and durable improvements. The `self-improve` skill exists; what's missing is **cross-session aggregation**.

**Install shape.**
- A `process-audit` chain (already partially shipped per the orchestration commit) that reads `specs/memory/feedback/` (from B9), clusters corrections, and proposes rule edits.
- Manual approval still required.

**Effort.** 2–3 days, depends on B9 shipping first.

**Verdict: `OPT-IN`.** Sequence after B9.

---

# Summary table — sorted by verdict

| Verdict | Items | Why |
|---|---|---|
| **ADD-NOW** | A4 Coverage thresholds · A8 Constitution population | Block downstream gates; trivial fixes; in the review's top 3. |
| **ADD-NEXT** | A3 Observability gate · A9 Knowledge graph integration · A11 Standards-as-code · B1 RAG preflight · B4 Chain-of-Verification · B6 Plan validation · B13 5-Whys · B17 Adversarial planning | Concrete, packageable, build on existing assets. ~1–3 days each. |
| **OPT-IN** | A1 Few-shot examples · A6 Cross-repo workflows · A7 Compaction hooks · A12 TillDone · B3 Debate workflow · B5 Auto-recovery · B9 Human feedback capture · B18 Continuous learning | Useful but situational; off by default; gated by feature flag. |
| **DEFER** | A2 Generated knowledge · A10 Agent progression · B7 Guardrails runtime · B8 Agent state machine · B11 Eval benchmark · B14 Env-aware planning | Re-evaluate after foundations land or once tool support improves. |
| **SKIP** | A5 API cost monitoring · A13 Multi-CLI · B2 Model routing · B10 Dynamic prompt optimization · B12 Tool creation by agents · B15 Ensemble · B16 Streaming | Out of scaffolder scope, deliberate strategic choice, or research-level effort. |

---

# Recommended sequencing (if you green-light the verdicts above)

**Phase 1 — Foundations (1 sprint, ~5–7 days):**
1. A4 Coverage thresholds (2h)
2. A8 Constitution population — wizard + extract-standards integration (3–5 days)
3. A11 Standards-as-code — starter set (2–3 days)

**Phase 2 — Gates that actually gate (1 sprint, ~5–7 days):**
4. A3 Observability gate (1–2 days)
5. B6 Plan validation skill (1–2 days)
6. B4 Chain-of-Verification (2–3 days)

**Phase 3 — Retrieval & analysis loops (1 sprint, ~5–7 days):**
7. A9 Knowledge graph auto-injection (2–3 days)
8. B1 RAG preflight skill (3–5 days, partially overlaps with #7)
9. B13 5-Whys for bugfix (1 day)
10. B17 Adversarial plan review (1 day)

**Phase 4 — Opt-in feature flags (rolling):**
11. A1, A7, A12, B3, B5, B9 — ship as feature flags, on-demand.
12. B18 — once B9 has produced enough feedback data.

**Outside this plan:** the DEFER and SKIP rows. Revisit DEFER quarterly; SKIP only if strategy changes.

---

# Open questions for you

1. **A8 vs A11 sequencing.** Both are foundational. Constitution-first is safer (every project gets it) but standards-first creates more leverage if `extract-standards` can be hooked into `init`. Preference?
2. **B1 RAG scope.** Should preflight retrieval be a default on all chains, or opt-in via feature flag? Default-on costs tokens on every task.
3. **OPT-IN preset placement.** Do these belong in a new `--preset full` extension, or under `--features` only?
4. **B7 Guardrails.** Worth investing in a portable `guardrails-mcp` server, or accept tool asymmetry?

This document is research only. No code or config changed.
