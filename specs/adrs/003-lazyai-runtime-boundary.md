# ADR-003: LazyAI Owns Runtime, Vibe-Lab Supplies Principles

**Date:** 2026-06-13  
**Status:** Accepted — maintainer approved 2026-06-21
**Deciders:** LazyAI maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Capture the runtime-boundary decision required before the LazyAI runtime refactor can be planned or implemented.

---

## Context

The final adversarial reviews agree that the strategic direction is correct but the previous runtime refactor plan is unsafe as written. The central correction is that `lazyai` must be the compact Go runtime for agentic CLI tooling, while `vibe-lab` supplies principles and adapter expectations. Treating `vibe-lab` as the runtime is a category error.

The reviews identify a large dependency and data-safety surface: 13 CLI files with orchestrator references, direct imports from task/workflow/orchestration commands, a 26-table runtime schema, Fortnite-heavy library content, adapter tests that encode Fortnite behavior, missing handoff and V2 schema designs, missing token-rent enforcement, and rollback gaps.

This ADR records the proposed architectural boundary only. It does not approve package deletion, schema migration, or adapter rewrites. Those remain gated by `specs/025-lazyai-runtime-refactor/spec.md` and a future human-approved plan.

**Related artifacts:**
- Spec: [`specs/025-lazyai-runtime-refactor/spec.md`](../025-lazyai-runtime-refactor/spec.md)
- Research: [`specs/025-lazyai-runtime-refactor/research.md`](../025-lazyai-runtime-refactor/research.md)
- DeepSeek final review: [`docs/reports/final-review-deepseek-v4-pro-2026-06-13.md`](../../docs/reports/final-review-deepseek-v4-pro-2026-06-13.md)
- Gemini final review: [`docs/reports/final-review-gemini-3.1-pro-2026-06-13.md`](../../docs/reports/final-review-gemini-3.1-pro-2026-06-13.md)

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| Article II — Test-First Imperative | Supports | Adapter contracts, schema migrations, and rollback behavior must be tested before production changes. |
| Article III — Docs as Source of Truth | Supports | This ADR and Spec 025 must be approved before plan/tasks/code. |
| Article IV — Anti-Speculation | Supports | Refactor scope is limited to removing demonstrated runtime bloat and designing missing contracts. |
| Article V — Simplicity Over Abstraction | Supports | The chosen boundary keeps one runtime owner instead of spreading runtime semantics across principles libraries. |
| Article VI — Anti-Overengineering | Supports | The refactor removes mega-orchestration surfaces and enforces small-library token rent. |
| Gate 5 — Observability Readiness | Supports | Rollback and restore verification are mandatory before destructive phases. |

This ADR does not amend the constitution.

---

## Options Considered

### Option A — LazyAI runtime, vibe-lab principles/adapters *(proposed)*

- **Summary:** `lazyai` owns runtime behavior; `vibe-lab` informs principles and adapter expectations; Fortnite/orchestrator runtime bloat is removed only after evidence gates pass.
- **Complexity:** Medium. Requires audit, schema design, adapter contract, rollback, and budget tooling.
- **Consistency:** High. Matches the final-review consensus and existing Go runtime specs.
- **Reversibility:** Medium. Destructive steps require tagged releases, archived assets, migrations, and restore tests.
- **Performance impact:** Positive if bloat is removed; risk remains until measured.
- **Team familiarity:** High for LazyAI Go runtime work; medium for migration/rollback discipline.

### Option B — vibe-lab as runtime, LazyAI as wrapper

- **Summary:** Move runtime responsibility to vibe-lab and make LazyAI consume it.
- **Complexity:** High. Cross-repo runtime ownership and release coordination become mandatory.
- **Consistency:** Low. Final reviews reject this as the core category error.
- **Reversibility:** Low. Runtime semantics would be split before the current dependency graph is understood.
- **Performance impact:** Unknown; additional boundary may add indirection.
- **Team familiarity:** Low. No reviewed implementation contract exists.

### Option C — Preserve Fortnite/orchestrator runtime status quo

- **Summary:** Keep current runtime/library/orchestrator surfaces and avoid destructive refactor.
- **Complexity:** Low short-term, high long-term.
- **Consistency:** Low. Conflicts with the reviews' consensus that runtime bloat is real and measurable.
- **Reversibility:** High short-term because nothing changes.
- **Performance impact:** Negative/unchanged. Token-rent and schema bloat remain.
- **Team familiarity:** High, but at the cost of preserving the problem.

---

## Decision

Propose Option A: `lazyai` owns runtime behavior; `vibe-lab` supplies principles and adapter expectations; Fortnite/orchestrator runtime bloat is eligible for removal only after Spec 025 evidence gates are satisfied. Acceptance is pending a human-authored tracked approval commit.

---

## Rationale

- Both final reviews agree the strategic framing is correct: `lazyai` is the runtime and `vibe-lab` is not.
- Option A preserves the reusable adapter layer while forcing it through a neutral contract before excision.
- Option A supports the constitution's simplicity and anti-overengineering articles by shrinking runtime scope instead of hiding orchestration under a new name.
- Option A keeps destructive work gated by prerequisite evidence, rollback tests, and human approval.

**Why the rejected options were rejected:**

- Option B repeats the category error called out by the synthesis and would expand cross-repo coupling before the dependency audit exists.
- Option C avoids immediate risk but preserves 26-table/orchestrator/Fortnite bloat and leaves token-rent governance unenforced.

---

## Consequences

**Positive:**

- Runtime ownership is unambiguous.
- Adapter work can target neutral behavior instead of Fortnite fallback behavior.
- Schema, handoff, budget, and rollback design become explicit gates instead of hidden assumptions.

**Negative / accepted trade-offs:**

- Planning cannot proceed quickly; prerequisite evidence is mandatory.
- Destructive work requires 30–40 working days, not the prior 17–23 day estimate.
- Existing users of Fortnite/OpenCode workflows may require migration notes or retained lightweight definitions.

**Neutral:**

- Orchestrator paradox resolved (2026-06-13 Clarify): Option B — redesigned primary-agent path. The orchestrator concept is removed entirely; non-Fortnite adapters wire to a redesigned primary agent. No `orchestrator.md` shim.

---

## Reversal Conditions

Re-open this ADR if any of these become true:

- The dependency audit proves Fortnite/orchestrator packages are required for the majority of active, supported LazyAI workflows.
- The V2 schema cannot preserve or safely restore existing runtime data.
- Adapter contracts cannot express current supported behavior without retaining heavy orchestration packages.
- Human review rejects the LazyAI-runtime / vibe-lab-principles boundary.

---

## Implementation Pointer

- Spec: [`specs/025-lazyai-runtime-refactor/spec.md`](../025-lazyai-runtime-refactor/spec.md)
- Plan: [`specs/025-lazyai-runtime-refactor/plan.md`](../025-lazyai-runtime-refactor/plan.md) (created 2026-06-13; corrected per two adversarial reviews)
- Tasks: not yet created; blocked on plan approval.
- Code changes: none authorized by this ADR.

---

## Memory Update

- [ ] Append ADR acceptance to `.specify/memory/repos/lazyai/rpi-execution.md` if this ADR is approved.
