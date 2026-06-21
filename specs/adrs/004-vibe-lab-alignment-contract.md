# ADR-004: Capability-First Vibe-Lab Alignment

**Date:** 2026-06-15  
**Status:** Accepted — maintainer approved 2026-06-21
**Deciders:** LazyAI maintainers
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Record how LazyAI should close the remaining vibe-lab parity gaps without regressing verified tool-native contracts or reviving retired runtime surfaces.

---

## Context

The 2026-06-15 parity research showed that the remaining gaps are concentrated in shipped setup-core behavior, not in missing reference notes. The biggest defects are concrete: the embedded library omits `packages/cli/library/skills/`, active emitted skills remain limited to four canonical micro-skills, runtime hook adapters are absent, Pi and Antigravity are not supported product adapters, `create` and validation disagree on supported artifact types, and docs/manifests drift from the code.

The same research also showed that literal file-for-file mirroring of the current local vibe-lab checkout would be unsafe. The baseline still carries repo-local or stale shapes in at least two places: Copilot surfaces use `.agent.md` there while current LazyAI code and prior Copilot research target `.agent.yaml`; baseline OpenCode guidance still centers a root `opencode.json` attachment model while current LazyAI intentionally standardized on `.opencode/opencode.jsonc`.

The refactor therefore needs an explicit rule for what “alignment” means: close capability gaps where LazyAI is missing baseline behavior, but do not blindly copy baseline file shapes when LazyAI already has a verified, supported native contract.

**Related artifacts:**
- Research: [`specs/refactors/026-vibe-lab-alignment/research.md`](../refactors/026-vibe-lab-alignment/research.md)
- Spec: [`specs/refactors/026-vibe-lab-alignment/spec.md`](../refactors/026-vibe-lab-alignment/spec.md)
- Prior ADR: [`specs/adrs/003-lazyai-runtime-boundary.md`](./003-lazyai-runtime-boundary.md)
- Companion ADR (product-framing rule): [`specs/adrs/005-core-vs-optional-modules.md`](./005-core-vs-optional-modules.md) — establishes setup-core as the default product boundary and demotes runtime-adjacent commands to optional modules; this ADR (004) only governs the capability-parity contract, not the product-headline framing.
- Product boundaries: [`docs/concepts/product-boundaries.md`](../../docs/concepts/product-boundaries.md)

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| Article I — Library-First | bears | The decision should reuse the existing library/adapters/scaffold system instead of adding a new external generation framework. |
| Article III — Docs as Source of Truth | bears | The chosen parity contract must be written down before implementation because current docs and code already disagree. |
| Article IV — Anti-Speculation | bears | The work must stop at identified parity gaps and not reintroduce retired runtime orchestration surfaces. |
| Article V — Simplicity | bears | Literal baseline mirroring is simpler only on paper; capability-first parity is simpler in operation because it avoids dual contracts. |
| Article VI — Anti-Overengineering | bears | The chosen path should add only the missing tool surfaces and runtime adapters needed to close the observed gaps. |

This ADR does not amend the constitution.

---

## Options Considered (Tree of Thoughts)

### Option A — Literal baseline mirroring
- **Summary:** Copy the current local vibe-lab checkout as closely as possible, including file names, root config placement, and tool-surface shapes.
- **Complexity:** High
- **Reversibility:** Medium
- **Performance impact:** Neutral
- **Team familiarity:** Medium
- **Constitution fit:** Weakens Articles V and VI because it can preserve stale shapes and force dual-path behavior.

### Option B — Capability-first parity with verified native shapes
- **Summary:** Close the missing capability gaps against vibe-lab, but keep currently verified native LazyAI/tool contracts where the baseline checkout is stale or repo-local.
- **Complexity:** Medium-high
- **Reversibility:** High when implemented in phases with fixture tests per tool
- **Performance impact:** Neutral
- **Team familiarity:** High
- **Constitution fit:** Best fit for Articles III, IV, V, and VI because it names the contract once and implements only the missing surfaces.

### Option C — Minimal cleanup only
- **Summary:** Fix the embedded-skills bug and doc drift, then document Pi/Antigravity and hook-runtime gaps as intentional exclusions.
- **Complexity:** Low
- **Reversibility:** High
- **Performance impact:** Neutral
- **Team familiarity:** High
- **Constitution fit:** Fits Articles V and VI, but fails the user-approved alignment goal and leaves known parity gaps open.

---

## Decision

Choose **Option B — capability-first parity with verified native shapes**.

Alignment for this refactor means:
- add missing capabilities where LazyAI currently lacks the baseline behavior;
- preserve verified native contracts unless newer implementation evidence proves they are wrong; current evidence supersedes legacy Copilot `.agent.yaml` skill output with Agent Skills directories, while OpenCode MCP placement remains under active review;
- keep Pi and Antigravity limited to the smallest supported surfaces the baseline actually uses;
- keep retired workflow/task/orchestration/eval runtime surfaces retired.

---

## Rationale

- The user explicitly asked to cover the remaining gaps and align LazyAI as much as possible with the baseline. Option C does not satisfy that ask.
- Literal mirroring would treat the baseline checkout as infallible even where current LazyAI evidence already indicates otherwise. That would reopen contract drift instead of closing it.
- Capability-first parity lets LazyAI add the missing value: shipped skills, runtime hooks, Pi skills-only support, Antigravity support, workflow cataloging, and provenance cleanup.
- The phased implementation model keeps the work reversible: correctness and doc drift first, then supported-tool parity, then hook runtime, then new tool adapters, then final manifest/docs closure.

**Why the rejected options were rejected:**
- **Option A:** Rejected because it would knowingly copy stale or repo-local shapes such as baseline Copilot `.agent.md` files and the older OpenCode attachment model, creating new divergence inside LazyAI.
- **Option C:** Rejected because it preserves the biggest active gaps — no hook runtime emission and no Pi/Antigravity support — after the user explicitly asked to close them.

---

## Consequences

**Positive:**
- LazyAI gets a single explicit definition of what “aligned” means.
- Missing capabilities are closed without reviving retired runtime surfaces.
- Docs, manifests, and adapter tests can validate one consistent contract.

**Negative / accepted trade-offs:**
- The refactor is broader than a simple doc update because it adds tool adapters and hook-runtime generation.
- The resulting LazyAI surface will still not be byte-for-byte identical to the baseline checkout in places where native contracts differ.
- Pi and Antigravity add maintenance cost; re-evaluate if their emitted fixtures prove unstable across releases.

**Neutral:**
- Existing LazyAI runtime extras remain allowed by ADR-003 and are unaffected by this decision.

---

## Reversal Conditions

Re-open this ADR if any of the following becomes true:

- Upstream or implementation evidence proves another preserved native contract is wrong, as happened for the legacy Copilot `.agent.yaml` skill surface.
- Pi or Antigravity fixture coverage shows their generated surfaces cannot be kept stable without disproportionate maintenance.
- The team decides LazyAI should support only Claude/OpenCode/Copilot and explicitly no longer pursue baseline breadth.
- Human review rejects capability-first parity and asks for literal baseline mirroring instead.

---

## Implementation Pointer

- Plan: [`specs/refactors/026-vibe-lab-alignment/plan.md`](../refactors/026-vibe-lab-alignment/plan.md)
- Spec: [`specs/refactors/026-vibe-lab-alignment/spec.md`](../refactors/026-vibe-lab-alignment/spec.md)
- PRs: pending
- Standards updated: pending

---

## Memory Update

- [ ] Update `specs/KNOWLEDGE_MAP.md` if this ADR is accepted.
- [ ] Mark any superseding decision if future upstream tool-contract research overturns the Copilot or OpenCode contract assumptions.
