# Research: LazyAI Runtime Refactor from Final Reviews

**Date:** 2026-06-13  
**Status:** Complete — research supports spec/plan drafting; implementation gate pending
**Workflow type:** Refactor — Full RPI, ADR required  
**Primary artifact:** `specs/025-lazyai-runtime-refactor/spec.md`  
**ADR:** `specs/adrs/003-lazyai-runtime-boundary.md`

---

## Inputs

- `docs/reports/final-review-deepseek-v4-pro-2026-06-13.md`
- `docs/reports/final-review-gemini-3.1-pro-2026-06-13.md`
- `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md`
- `.specify/memory/constitution.md`

---

## Request Classification

This is a destructive architecture refactor, not a feature or bugfix.

- It changes runtime boundaries: `lazyai` as runtime, `vibe-lab` as principles/adapter guidance.
- It removes or archives large existing runtime surfaces: Fortnite library content, orchestrator packages, workflow/taskqueue/dispatch runtime packages.
- It changes persistence expectations: 26-table runtime state must shrink to a small V2 runtime schema without data loss.
- It changes adapter behavior: adapter output must be defined by a neutral canonical contract, not Fortnite assumptions.

Scope is larger than 100 LOC and contains more than three architecture decisions. Full RPI is required, with human gates after research/specification, plan, and each implementation batch.

---

## Consensus Findings

| Finding | Evidence | Spec impact |
|---|---|---|
| Strategic direction is correct | Both final reviews agree `lazyai` should be the runtime and `vibe-lab` should remain a principles layer. | Encode this as an invariant and ADR decision. |
| Existing plan is not executable as written | DeepSeek verdict says Phase 0 must not be approved until evidence-gated prerequisites are satisfied; Gemini agrees approval must be denied until unknowns are designed. | Planning must be gated by prerequisite artifacts. |
| Current confidence is too low for destructive work | DeepSeek assesses 63%; Gemini assesses ~63.5%; both project 85–88% only after mitigation. | Implementation may not begin until evidence raises confidence to at least 85%. |
| Dependency surface is undercounted | DeepSeek identifies 13 CLI files with orchestrator references and direct import breaks in task/workflow/orchestration commands. | Phase 0 must include a per-file import audit matrix. |
| Schema work is prerequisite design, not in-phase cleanup | Both reviews call the 26-table to 5-table migration underestimated and unsafe without V2 DDL, migration SQL, rollback, and synthetic data tests. | V2 schema and rollback acceptance criteria become mandatory pre-plan deliverables. |
| Adapter tests lack a neutral contract | Both reviews warn current adapter tests preserve Fortnite behavior unless the canonical behavior is specified first. | Adapter test contract is a prerequisite before excision. |
| Orchestrator paradox is unresolved | Non-Fortnite adapter installs an orchestrator agent while orchestrator packages are scheduled for removal. | Clarification gate must choose lightweight agent definition vs redesigned primary agent. |
| Token-rent governance is not enforceable yet | Reports agree CI/pre-commit enforcement and a documented override are both needed. | Budget tooling and override contract become requirements. |
| Archive-before-delete is insufficient rollback | Reports require tags, migration reversal, user notification, and restore verification. | Rollback procedure becomes a phase gate, not a courtesy. |

---

## Evidence-Gated Prerequisites

These prerequisites are merged from the DeepSeek 8-item list and Gemini 7-item list. They are mandatory before plan approval.

| ID | Prerequisite | Evidence required |
|---|---|---|
| P0-1 | CLI command import audit | Matrix for every `packages/cli/cmd/*.go` file, categorizing `runtime/workflow`, `runtime/taskqueue`, `runtime/dispatch`, `internal/orchestrator`, `packages/orchestrator`, and `library/fortnite` usage as breakage/rewrite/keep/remove. |
| P0-2 | Fortnite/OpenCode usage survey | Active user count, workflow usage breakdown, migration blockers, and notification needs. |
| P0-3 | V2 runtime schema design | ER diagram, DDL, migration up/down SQL, `loop-driver` default handling, and test plan for FK-saturated data. |
| P0-4 | Handoff markdown schema | Frontmatter contract, required sections, path conventions, ownership model, and round-trip expectations. |
| P0-5 | Canonical library specification | Agent/skill/hook inventory with usage justification and ownership per item. |
| P0-6 | Adapter test contract | Concrete table of files, agents, defaults, and expected outputs per adapter mode. |
| P0-7 | Token-rent enforcement design | Budget threshold, byte-counting rule, CI/pre-commit integration point, failure mode, and `.lazyai` override semantics. |
| P0-8 | Corrected phase plan and rollback procedure | Revised order, estimates, tagged release policy, migration reversal, data restore verification, and user notification template. |

---

## Corrected Execution Order

The reports converge on this order:

1. Audit dependencies and design missing contracts.
2. Rewrite adapters against a neutral contract.
3. Rewrite adapter tests against that contract.
4. Excise or archive Fortnite/orchestrator/runtime workflow surfaces.
5. Execute V2 schema migration with backup/restore/down tests.
6. Implement handoff capability using the approved handoff schema.
7. Curate canonical library and enforce token-rent budgets.

This order replaces the rejected sequence that removed packages before adapter tests and CLI commands were decoupled.

---

## Research Verdict

**Verdict:** GO for a specification and pre-approval plan draft; HOLD for implementation.

Reason: the reviews provide enough evidence to write an RPI refactor contract and Phase 0 plan, but they explicitly reject destructive implementation until prerequisite artifacts are complete, human-verifiable, and the confidence floor reaches 85%.

**Human gate required:** approve ADR-003, Spec 025, and the revised Phase 0 plan with tracked human authorship before moving to implementation.
