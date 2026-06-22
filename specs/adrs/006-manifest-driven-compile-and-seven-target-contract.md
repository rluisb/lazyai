# ADR-006: Manifest-Driven Compile and Seven-Target Contract

**Date:** 2026-06-22  
**Status:** Accepted — implemented in spec 029 cleanup phase  
**Deciders:** LazyAI maintainers

> **Purpose.** Record the V2 compile contract: `.ai/lazyai.json` and `.ai/lock.json` are the authoritative project-level inputs/outputs for setup-core compilation, the supported target set is exactly seven, Codex is not a compile target, and the binary name remains `lazyai-cli`.

---

## Context

Before spec 029, compile behavior still depended heavily on legacy setup-store state and repo-local conventions. The new product spec pack instead defines a canonical `.ai/` source model: manifest-driven target selection, an explicit lockfile, consolidated validation, migration/eject, and tool-bundle outputs. The pack also makes two boundary decisions explicit:

- the supported setup-core targets are exactly `opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, and `kiro`;
- the binary name stays `lazyai-cli`, even where older pack drafts used `lazyai` in command examples.

Historical Codex references still exist in the repository, but they are no longer part of the supported target contract. Some remain legitimate outside compile targeting, for example negative tests, provider/model names, or legacy migration context. The decision needed here is not “remove every Codex string from the repo”; it is “what contract does setup-core expose now?”

**Related artifacts:**
- Spec: [`specs/029-lazyai-v2/spec.md`](../029-lazyai-v2/spec.md)
- Plan: [`specs/029-lazyai-v2/plan.md`](../029-lazyai-v2/plan.md)
- Tasks: [`specs/029-lazyai-v2/tasks.md`](../029-lazyai-v2/tasks.md)
- Prior ADRs: [`004-vibe-lab-alignment-contract.md`](004-vibe-lab-alignment-contract.md), [`005-core-vs-optional-modules.md`](005-core-vs-optional-modules.md)

---

## Decision

Adopt the V2 compile contract defined by spec 029:

- `.ai/lazyai.json` is the authoritative compile manifest for target selection and safety/profile defaults;
- `.ai/lock.json` is the authoritative generated-output lockfile for idempotency, drift-aware writes, adapter metadata, and output hashes;
- canonical `.ai/` content is compiled into native tool surfaces through the existing adapter pipeline;
- the supported compile target set is frozen at seven: `opencode`, `claude`, `copilot`, `pi`, `omp`, `antigravity`, `kiro`;
- Codex is not part of the supported target contract and must be rejected when presented as a V2 manifest target;
- the binary name remains `lazyai-cli`; docs/examples that used bare `lazyai` as the executable are normalized to `lazyai-cli`.

---

## Options Considered

### Option A — Keep compile primarily store-driven, with manifest as an overlay
- **Summary:** Preserve `.ai-setup.json` as the practical source of truth and let `.ai/lazyai.json` override only some decisions.
- **Complexity:** Medium
- **Reversibility:** Medium
- **Performance impact:** Neutral
- **Team familiarity:** High
- **Why rejected:** It keeps two competing contracts alive. That weakens idempotency, migration clarity, and the pack’s canonical `.ai/`-first model.

### Option B — Make manifest + lockfile the explicit V2 compile contract
- **Summary:** Use `.ai/lazyai.json` for target selection and `.ai/lock.json` for generated-output state, while keeping adapters as the emitter seam.
- **Complexity:** Medium
- **Reversibility:** High
- **Performance impact:** Positive for repeat compiles through skip/lock behavior
- **Team familiarity:** High
- **Why chosen:** It matches the approved spec pack, reuses existing adapter/writer seams, and gives one durable contract to test and document.

### Option C — Broaden targets again or rename the binary to match older pack wording
- **Summary:** Treat older `lazyai` examples or historical Codex support as reasons to expand the supported contract.
- **Complexity:** High
- **Reversibility:** Medium
- **Performance impact:** Neutral
- **Team familiarity:** Low-medium
- **Why rejected:** It reopens already-approved scope. The human-approved clarifications for spec 029 explicitly kept the binary name `lazyai-cli` and dropped Codex from the supported target set.

---

## Consequences

**Positive:**
- Compile behavior has one authoritative project contract instead of split manifest/store ownership.
- The setup-core surface is easier to document, validate, and migrate because targets/profile/lock metadata live under `.ai/`.
- The exact supported target set is durable and testable.

**Negative / accepted trade-offs:**
- Legacy `.ai-setup.json` still exists for scoped install state, so maintainers must distinguish runtime/setup state from the new compile contract.
- Historical Codex references remain in non-target contexts; reviewers must not mistake those for active target support.
- MCP schema enforcement remains lighter than manifest enforcement today; the canonical scaffold emits a frozen v1 catalog, but validators still prioritize practical structure checks over full schema execution.

**Neutral:**
- Existing adapter implementations remain the emission seam; the ADR changes the source-of-truth contract, not the per-tool write strategy.

---

## Implementation Pointer

- Manifest: `packages/cli/internal/aimanifest/`
- Lockfile: `packages/cli/internal/lockfile/`
- Writer: `packages/cli/internal/writer/`
- Compile entrypoint: `packages/cli/cmd/compile.go`
- Validation/security: `packages/cli/internal/validate/`, `packages/cli/cmd/doctor_security.go`
- Migration/eject: `packages/cli/internal/migration/`, `packages/cli/internal/eject/`
- Bundles/evals: `packages/cli/internal/plugin/`, `packages/cli/internal/evals/`

---

## Follow-up Watchouts

- If future work adds or removes supported targets, supersede this ADR instead of silently widening the contract.
- If MCP validation moves to full schema execution, update this ADR’s note about lighter enforcement.
- If scoped install state is ever merged into the canonical `.ai/` contract, revisit the `.ai-setup.json` boundary explicitly.
