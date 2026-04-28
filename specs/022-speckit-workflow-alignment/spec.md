# Spec 022 — Workflow Library Alignment with Speckit + Harness Engineering

**Status:** In Spec | **Date:** 2026-04-27

## Goal

Rebuild ai-setup's library (skills, agents, prompts, templates, fragments) to provide a comprehensive, tool-agnostic workflow engine that:

1. **Absorbs speckit methodology** as the core SDLC (Spec-Driven Development workflow)
2. **Adds complementary workflows** beyond spec-kit's scope (bugfix, spike, PoC, housekeeping, review, memory, self-improvement)
3. **Embeds Harness Engineering Protocol** concepts (Feed Forward, Dual-Agent Contract, Feedback & Sensors, Memory & State, Anti-Slope)
4. **Embeds prompt engineering techniques** (CoT, ToT, ReAct, Self-Consistency, Reflexion, Few-Shot, Prompt Chaining)
5. **Works across all supported tools** (Claude Code, OpenCode, Codex, Gemini, Copilot, Pi) — not just Claude Code
6. **Integrates with workspace engine** (multi-repo ledgers, standards propagation, planning repo awareness)

## Architecture Decision

**Option B: Adapt & Enhance** — speckit is the reference methodology, not a runtime dependency.

| Aspect | Decision |
|---|---|
| Speckit naming | Use `speckit-*` skill names for compatibility signaling |
| Speckit content | Absorb methodology, enhance with ai-setup concepts |
| Speckit CLI | NOT a dependency — ai-setup generates tool-native files |
| Template format | Match speckit template structure, add ai-setup extensions |
| Prompt techniques | Embedded directly in skill instructions (NOT external plugins) |
| Harness Engineering | Woven into workflow phases as gates and checkpoints |

---

## Workflow Catalog

### Tier 1 — Core SDLC

#### WF-01: Spec-Driven Development (SDD)
- **Purpose:** Full feature development from spec to implementation
- **Speckit alignment:** `/speckit.specify` → `/speckit.clarify` → `/speckit.plan` → `/speckit.tasks` → `/speckit.checklist` → `/speckit.analyze` → `/speckit.implement`
- **When to use:** New feature, greenfield, brownfield with significant scope
- **Prompt techniques:** Chain-of-Thought (specify phase), ReAct (plan + implement), Prompt Chaining (pipeline), Self-Consistency (analyze)
- **Harness Engineering:** Feed Forward (spec blueprint), Contract (analyze verification), Feedback (checklist validation), Memory (ledger update on completion)
- **Skills produced:** `speckit-specify`, `speckit-clarify`, `speckit-plan`, `speckit-tasks`, `speckit-analyze`, `speckit-checklist`, `speckit-implement`
- **Templates produced:** `spec-template.md`, `plan-template.md`, `tasks-template.md`, `checklist-template.md`

#### WF-02: RPI (Research → Plan → Implement)
- **Purpose:** Lighter-weight version of SDD for non-trivial but bounded tasks
- **Speckit alignment:** Subset of SDD — skip full spec.md, keep plan.md + tasks + implement
- **When to use:** Task that requires research and planning but not a full feature spec (e.g., "refactor auth middleware", "add rate limiting")
- **Prompt techniques:** ReAct (research + plan interleaved), Few-Shot (task examples), Decision Protocol (multiple approaches)
- **Harness Engineering:** Feed Forward (constrained scope), Contract (acceptance criteria), Feedback (quality gates per task)
- **Skills produced:** `rpi` (updated — orchestration skill that chains research + plan + implement)
- **Templates used:** Reuses plan-template.md, task.md

#### WF-03: Bugfix
- **Purpose:** Root cause analysis + minimal surgical fix
- **Speckit alignment:** None — this is a separate, lighter workflow
- **When to use:** Bug reports, issues that don't require re-specification
- **Prompt techniques:** Trace execution (follow the bug through the code), Hypothesis testing (propose → verify → fix), Few-Shot (similar bug patterns)
- **Harness Engineering:** Contract (expected vs actual behavior), Feedback (test the fix, verify no regression), Memory (update bugfix ledger)
- **Skills produced:** `bugfix`
- **Templates produced:** `bugfix-rca-template.md` (update existing)

#### WF-04: Spike / Deep Research
- **Purpose:** Multi-round research with different perspectives to reduce uncertainty before committing to a plan
- **Speckit alignment:** Enhanced version of `/speckit.clarify` + `/speckit.plan` research phase
- **When to use:** High uncertainty, new technology, architectural decisions
- **Prompt techniques:** Tree-of-Thoughts (explore multiple paths, evaluate, backtrack), Self-Consistency (multiple perspectives, consensus), Generated Knowledge (pre-generate context), Reflexion (self-evaluate findings)
- **Harness Engineering:** Feed Forward (research scope and questions upfront), Dual-Agent (two perspectives on same problem), Memory (capture findings)
- **Skills produced:** `spike`
- **Templates produced:** `spike-template.md`

#### WF-05: Proof-of-Concept (PoC)
- **Purpose:** Validate feasibility with minimal implementation — shallower than full RPI
- **Speckit alignment:** Mini-RPI — research → lightweight plan → spike implementation
- **When to use:** "Can we do X with Y?", technology evaluation, prototype
- **Prompt techniques:** Few-Shot (similar PoC patterns), Chain-of-Thought (feasibility reasoning), Generated Knowledge (pre-research)
- **Harness Engineering:** Feed Forward (success criteria for PoC), Contract (what "validated" means), Anti-Slope (throw away the PoC — don't ship it)
- **Skills produced:** `proof-of-concept`
- **Templates produced:** `poc-template.md`

#### WF-06: Housekeeping
- **Purpose:** Dependency updates, tech debt resolution, cleanup — structured maintenance
- **Speckit alignment:** None — maintenance workflow
- **When to use:** Scheduled maintenance, version bumps, cleanup runs
- **Prompt techniques:** Structured Output (checklist format), Few-Shot (common maintenance patterns)
- **Harness Engineering:** Feed Forward (maintenance scope), Feedback (verify after each change), Memory (update tech-debt ledger)
- **Skills produced:** `housekeeping`
- **Templates produced:** `housekeeping-template.md`

### Tier 2 — Quality & Governance

#### WF-07: Review (Multi-Lens)
- **Purpose:** Code review with specialized lenses — not just "looks good"
- **Speckit alignment:** Enhanced version of `/speckit.analyze` + `/speckit.checklist` applied to code
- **When to use:** Before merge, after implementation, periodic audit
- **Prompt techniques:** Dual-Agent Simulation (implementor vs reviewer), Self-Consistency (multiple review perspectives), Multi-Step (security → perf → style → correctness, each step feeds next)
- **Harness Engineering:** Contract (check against spec/plan), Feedback (actionable findings with severity), Dual-Agent (different reviewer perspectives)
- **Skills produced:** `review` (rebuild existing)
- **Templates produced:** `code-review-template.md` (update existing)

#### WF-08: Code Standards Update
- **Purpose:** Extract patterns from codebase → update project/workspace/global standards
- **Speckit alignment:** None — meta-governance workflow
- **When to use:** After significant implementation, periodic standards review
- **Prompt techniques:** Few-Shot (pattern recognition), Reflexion (compare current vs desired), Self-Consistency (check against multiple codebases in workspace)
- **Harness Engineering:** Feedback (patterns observed in reality), Memory (standards as durable memory), Contract (standards as constraints)
- **Skills produced:** `extract-standards` (updated — workspace-aware)
- **Templates produced:** `standard.md` (update existing)

#### WF-09: Memory Update
- **Purpose:** Update project/workspace memory from completed work (ledger, state, learnings)
- **Speckit alignment:** None — post-workflow housekeeping
- **When to use:** After any completed task, feature, bugfix, or housekeeping run
- **Prompt techniques:** Structured Output (ledger format), Few-Shot (memory entry patterns)
- **Harness Engineering:** Memory & State (beating amnesia — durable record of what happened), Feedback (lessons learned)
- **Skills produced:** `update-memory` (new, replaces/extends `memory-write`)
- **Templates produced:** `ledger-template.md`, `last-known-state-template.md`

#### WF-10: Constitution Check
- **Purpose:** Verify implementation against project constitution (principles, constraints, quality gates)
- **Speckit alignment:** Enhanced version of speckit's constitution check phase
- **When to use:** Gate before any plan execution, post-implementation compliance check
- **Prompt techniques:** Dual-Agent (plan vs constitution), Structured Output (pass/fail per principle)
- **Harness Engineering:** Contract (constitution as governing contract), Feedback (violations detected)
- **Skills produced:** Embedded in `speckit-plan` and `speckit-implement` — no standalone skill

### Tier 3 — Meta

#### WF-11: Self-Improvement
- **Purpose:** AI reviews its own output across sessions, identifies improvement patterns, updates its own guidance
- **Speckit alignment:** None — meta-cognitive workflow
- **When to use:** Periodic (weekly), after significant work, when patterns emerge
- **Prompt techniques:** Reflexion (self-evaluate outputs), ART — Automatic Reasoning and Tool-use (analyze own traces)
- **Harness Engineering:** Feedback (objective self-assessment), Memory (persist learnings), Anti-Slope (prevent regression)
- **Skills produced:** `self-improve`
- **Templates produced:** None — updates memory files directly

#### WF-12: Process Audit
- **Purpose:** Reviews workflow adherence, suggests process improvements
- **Speckit alignment:** None — meta-process workflow
- **When to use:** Periodic, after project milestones, when teams report friction
- **Prompt techniques:** Self-Consistency (check across multiple sessions), Structured Output (audit report)
- **Harness Engineering:** Feedback & Sensors (measure process effectiveness), Memory (process metrics)
- **Skills produced:** `process-audit`
- **Templates produced:** `audit-template.md`

---

## Prompt Engineering Techniques — How They're Embedded

Each skill will declare which techniques it uses in its YAML frontmatter:

```yaml
---
name: speckit-specify
description: Create a feature specification from a user description
argument-hint: "Build a photo album organizer that..."
techniques:
  - chain-of-thought    # Step-by-step reasoning for user stories
  - few-shot             # Examples of well-formed specs
  - feed-forward         # Blueprint before implementation
---
```

The skill body then instructs the AI to use these techniques:

```markdown
## Reasoning Protocol

Use **Chain-of-Thought** reasoning when decomposing user stories:
1. First, identify the primary user and their goal
2. Then, enumerate the steps the user takes
3. Finally, define the acceptance criteria for each step

Each user story MUST be independently testable — if you implement just one,
you should have a viable MVP.

## Harness Engineering: Feed Forward

Before writing the spec, generate a **Blueprint** section:
- What is the scope boundary? (what's IN and what's OUT)
- What are the key entities?
- What are the constraints from the constitution?

This blueprint drives the spec — do not start writing user stories until
the blueprint is complete.
```

---

## Template Alignment

| Current ai-setup template | → | New template | Changes |
|---|---|---|---|
| `templates/spec-template.md` | → | `templates/spec-template.md` | Full rewrite to speckit format (User Scenarios, FR-*, Success Criteria, Edge Cases, Assumptions) |
| `templates/plan-template.md` | → | `templates/plan-template.md` | Full rewrite to speckit format (Technical Context, Constitution Check, Project Structure, Complexity Tracking) |
| `templates/task.md` | → | `templates/tasks-template.md` | Full rewrite to speckit format (Phases, User Story grouping, [P] parallel markers, Dependencies) |
| `templates/checklist-template.md` | → | `templates/checklist-template.md` | Align with speckit quality gate format |
| `templates/adr.md` | → | Keep, minor update | Add speckit constitution reference |
| `templates/bugfix-rca-template.md` | → | Update | Add Harness Engineering Contract section |
| `templates/standard.md` | → | Update | Workspace-aware standards format |
| None | → | `templates/spike-template.md` | New |
| None | → | `templates/poc-template.md` | New |
| None | → | `templates/housekeeping-template.md` | New |
| None | → | `templates/ledger-template.md` | New (currently generated inline) |
| None | → | `templates/audit-template.md` | New |

---

## Agent Updates

| Agent | Changes |
|---|---|
| `planner` | Add speckit awareness — reads constitution, spec, generates speckit-compatible plans |
| `builder` | Add workspace awareness — respects per-repo permissions, updates ledgers |
| `reviewer` | Add multi-lens capability — security, perf, style, correctness lenses |
| `scout` | Add speckit-aware research — searches existing specs for related work |
| `red-team` | Add Dual-Agent Contract mode — adversarial spec/code verification |
| `orchestrator` | Add workflow chaining — can sequence skills in dependency order |

---

## Fragment Updates

| Current fragment | → | New fragment | Changes |
|---|---|---|---|
| `rpi-workflow.md` | → | `rpi-workflow.md` | Rebuild with Harness Engineering phases |
| `reasoning-protocol.md` | → | Update | Add CoT, ReAct, ToT technique references |
| `decision-protocol.md` | → | Keep | Already good |
| `context-discipline.md` | → | Keep | Already good |
| `quality-gates.xml` | → | Update | Add speckit constitution check |
| None | → | `harness-protocol.md` | New — Feed Forward, Contract, Feedback, Memory, Anti-Slope |
| None | → | `workspace-protocol.md` | New — multi-repo awareness, ledger updates, standards propagation |

---

## Acceptance Criteria

1. **All 12 workflows have corresponding skills** in `library/skills/`
2. **All speckit-aligned templates** exist in `library/templates/` with matching format
3. **Each skill** declares prompt techniques in frontmatter
4. **Each workflow** embeds Harness Engineering concepts at the appropriate phase
5. **All agents** are updated for speckit + workspace awareness
6. **RPI fragment** is rebuilt with Feed Forward → Research → Plan → Implement → Feedback phases
7. **Workspace integration** — skills reference planning repo, per-repo ledgers, standards propagation
8. **Tool-native output** — `ai-setup compile` generates correct skill files for Claude Code (`.claude/commands/`), OpenCode (`.opencode/skills/`), Codex (`.codex/config.toml`), etc.
9. **Existing tests continue to pass** — no regression in scaffold, compile, or adapter logic
10. **New tests** cover: skill frontmatter validation, template format validation, workspace-aware skill behavior
