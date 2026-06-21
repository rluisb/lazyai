# ADR-005: Core-First LazyAI with Optional Runtime Modules

**Date:** 2026-06-15  
**Status:** Accepted — maintainer approved 2026-06-21
**Deciders:** LazyAI maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Record how LazyAI should maximize vibe-lab philosophical alignment without deleting useful LazyAI-specific capabilities: keep setup-core as the default product, and reclassify runtime-adjacent commands as optional modules with a documented future enablement path.

---

## Context

The alignment cleanup and report closed part of the gap, but the remaining philosophy mismatch is not mainly about missing files. It is about product framing. LazyAI still presents itself broadly at the top level while vibe-lab is intentionally narrower. The current repo already contains the seam needed to narrow the default story: `docs/concepts/product-boundaries.md` explicitly separates `setup-core` from `ops-runtime-extra`, and the command inventory already classifies runtime-adjacent commands separately from the setup engine.

The user clarified the desired target state: LazyAI core should be as vibe-lab-like as practical, but LazyAI should still own an engine for install/update/maintenance and preserve the broader runtime-adjacent capabilities as opt-in modules. Those extras should be documented well enough now that they can later be implemented behind a proper module lifecycle instead of staying mixed into the product headline.

This means the decision is not whether runtime-adjacent commands may exist. The decision is whether they remain part of the default product identity. The alignment target now needs a second explicit rule: setup-core is the default product boundary; runtime-adjacent features are retained, but demoted to optional-module status until a later implementation phase gives them explicit enable/disable mechanics.

**Related artifacts:**
- Spec: [`specs/refactors/026-vibe-lab-alignment/spec.md`](../refactors/026-vibe-lab-alignment/spec.md)
- Plan: [`specs/refactors/026-vibe-lab-alignment/plan.md`](../refactors/026-vibe-lab-alignment/plan.md)
- Roadmap: [`specs/refactors/026-vibe-lab-alignment/core-module-roadmap.md`](../refactors/026-vibe-lab-alignment/core-module-roadmap.md)
- Prior ADRs: [`003-lazyai-runtime-boundary.md`](./003-lazyai-runtime-boundary.md), [`004-vibe-lab-alignment-contract.md`](./004-vibe-lab-alignment-contract.md)
- Product boundaries: [`docs/concepts/product-boundaries.md`](../../docs/concepts/product-boundaries.md)

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| Article I — Library-First | bears | The setup-core engine remains the center of the product and continues to reuse the existing embedded library and adapter system. |
| Article III — Docs as Source of Truth | bears | The core/module split must be documented before extraction work starts; otherwise philosophy drift will keep recurring. |
| Article IV — Anti-Speculation | bears | This decision documents the split now, but does not force immediate command deletion or a speculative module framework. |
| Article V — Simplicity | bears | One default story is simpler than presenting setup-core and runtime extras as one undifferentiated product. |
| Article VI — Anti-Overengineering | bears | The first phase is documentation and contract cleanup only; module lifecycle mechanics are deferred until there is explicit follow-up work. |

This ADR does not amend the constitution.

---

## Options Considered (Tree of Thoughts)

### Option A — Keep the broad LazyAI product story
- **Summary:** Preserve the current top-level framing of LazyAI as a broad AI operating system and document vibe-lab differences as intentional.
- **Complexity:** Low
- **Reversibility:** High
- **Performance impact:** Neutral
- **Team familiarity:** High
- **Constitution fit:** Acceptable operationally, but weaker fit with Articles III and V because the broad default story keeps philosophy drift alive.

### Option B — Make setup-core the default and runtime extras optional modules
- **Summary:** Reframe LazyAI around the vibe-lab-like setup engine, while keeping runtime-adjacent commands available but explicitly outside the default product philosophy.
- **Complexity:** Medium
- **Reversibility:** High when done as docs/ADR/roadmap first
- **Performance impact:** Neutral immediately; future module lifecycle work can preserve current behavior behind explicit enablement
- **Team familiarity:** High
- **Constitution fit:** Best fit for Articles III, IV, V, and VI because it narrows the default contract without speculative extraction work.

### Option C — Extract or remove runtime extras immediately
- **Summary:** Delete or hard-separate runtime-adjacent commands now so LazyAI matches vibe-lab philosophy by force.
- **Complexity:** High
- **Reversibility:** Low-medium
- **Performance impact:** Neutral to positive, but with high migration risk
- **Team familiarity:** Medium
- **Constitution fit:** Weakens Article IV because it would turn a documentation-boundary decision into a broad unapproved code migration.

---

## Decision

Choose **Option B — make setup-core the default and runtime extras optional modules**.

Concretely:
- the default LazyAI product story becomes the setup-core engine: initialize, compile, update, validate, explain, and maintain tool-native AI setup surfaces;
- runtime-adjacent commands currently classified as `ops-runtime-extra` remain shipped for now, but are no longer part of the default philosophy or headline alignment claim;
- those extras must be documented as optional modules with clear purpose, current state, boundaries, and future enable/disable expectations;
- this refactor phase documents and stages the split first; it does **not** require immediate deletion, extraction, or a new module framework.

---

## Rationale

- The user explicitly asked for LazyAI core to be as vibe-lab-like as possible while keeping the install/update/maintain engine and preserving extras as opt-in modules. Option B matches that target exactly.
- The current codebase already has the right seam. `setup-core` versus `ops-runtime-extra` is already documented; the missing piece is making that split the default product contract instead of a buried inventory detail.
- Option A would preserve useful capabilities, but it would cap philosophy alignment because the broad runtime story would remain the first thing users see.
- Option C would maximize purity on paper, but it would force a risky product and migration change before the command/module boundaries, docs, and enablement mechanics are fully designed.

**Why the rejected options were rejected:**
- **Option A:** Rejected because it keeps the core philosophical mismatch in place even after the concrete contract defects are fixed.
- **Option C:** Rejected because immediate extraction/removal is larger than the approved scope and would conflate boundary definition with code retirement.

---

## Consequences

**Positive:**
- LazyAI can align much more closely with vibe-lab at the product-core level without throwing away existing capabilities.
- The default story becomes cleaner: setup engine first, optional local modules second.
- Future extraction or enablement work gains a checked-in contract instead of improvising from the current mixed product surface.

**Negative / accepted trade-offs:**
- Documentation burden increases in the short term because every retained extra now needs explicit module-facing explanation.
- The repository will temporarily ship commands that still exist physically in the main CLI even though the docs treat them as optional modules; that transitional state must be described clearly.
- A later follow-up will still need to design real module lifecycle semantics if the team wants more than documentation-only opt-in.

**Neutral:**
- Existing runtime-adjacent commands remain available until a later implementation phase changes their installation or activation mechanics.

---

## Reversal Conditions

Re-open this ADR if any of the following becomes true:

- Human review decides LazyAI should remain intentionally broader at the top level and should not pursue vibe-lab-like core positioning.
- Core flows (`init`, `compile`, `update`, `doctor`, adapter generation) become dependent on runtime-extra packages in a way that makes the optional-module boundary false.
- Future module-lifecycle design proves that some current extras are actually inseparable from setup-core and should stay in the default boundary.
- User or maintainer evidence shows that the default/core versus optional/module distinction creates more confusion than the current broad product story.

---

## Implementation Pointer

- Plan: [`specs/refactors/026-vibe-lab-alignment/plan.md`](../refactors/026-vibe-lab-alignment/plan.md)
- Roadmap: [`specs/refactors/026-vibe-lab-alignment/core-module-roadmap.md`](../refactors/026-vibe-lab-alignment/core-module-roadmap.md)
- Follow-up implementation: pending human approval of the updated Phase 6 plan
- Standards updated: pending implementation phase; expected updates include `README.md`, `docs/concepts/how-it-works.md`, `docs/concepts/product-boundaries.md`, and CLI reference docs

---

## Memory Update

- [ ] Update `specs/KNOWLEDGE_MAP.md` if this ADR is accepted.
- [ ] Cross-link the final core/module docs from `docs/concepts/product-boundaries.md`.
- [ ] If a later implementation introduces real module lifecycle semantics, update or supersede this ADR with the implementation decision.
