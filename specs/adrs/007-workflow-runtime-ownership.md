# ADR-007: Workflow Runtime Ownership — LazyAI Generates and Validates, Host Tools Execute

**Date:** 2026-06-23  
**Status:** Proposed  
**Deciders:** LazyAI maintainers  
**Constitution:** [`.specify/memory/constitution.md`](../../.specify/memory/constitution.md)

> **Purpose.** Decide where executable workflow runtime belongs before LazyAI implements workflow emission beyond the current docs-only catalog. This ADR draws a bright line between what LazyAI may generate/validate and what must live in a host tool, plugin, or separate runtime package.

---

## Context

LazyAI embeds a workflow catalog under `packages/cli/library/workflows/` with nine workflow markdown files (adversarial-review, bugfix, code-review, documentation, feature, refactor, spike, verified-research, plus templates). Every entry is marked `adapter_targets: [none]` in `curation.yaml` — intentionally docs-only.

PR #322 captured the workflow delivery matrix and plugin/extension engine planning, but did not implement a runtime. That was deliberate: LazyAI is an asset manager/compiler/validation layer, not a workflow orchestrator. The retired `task`, `workflow`, `orchestrator`, and `eval` command surfaces were removed in prior cleanup (issues #229, #304).

Three ownership models are now on the table as LazyAI considers moving beyond docs-only workflow assets:

1. **LazyAI core** — embed a workflow execution engine in the `lazyai-cli` binary
2. **External/plugin runtime** — generate thin host-tool plugins/extensions that delegate execution to the host's native runtime
3. **Docs-only/catalog assets** — keep the current model: workflows are markdown reference material, never executable

**Related artifacts:**
- Prior ADRs: [`003-lazyai-runtime-boundary.md`](./003-lazyai-runtime-boundary.md), [`005-core-vs-optional-modules.md`](./005-core-vs-optional-modules.md), [`006-manifest-driven-compile-and-seven-target-contract.md`](./006-manifest-driven-compile-and-seven-target-contract.md)
- Product boundaries: [`docs/concepts/product-boundaries.md`](../../docs/concepts/product-boundaries.md)
- Workflow delivery matrix: [`docs/concepts/tools.md`](../../docs/concepts/tools.md)
- Prior issues: [#229](https://github.com/rluisb/lazyai/issues/229), [#304](https://github.com/rluisb/lazyai/issues/304), [#317](https://github.com/rluisb/lazyai/issues/317), [#318](https://github.com/rluisb/lazyai/issues/318), [#319](https://github.com/rluisb/lazyai/issues/319)
- PR: [#322](https://github.com/rluisb/lazyai/pull/322)

---

## Constitution Alignment

| Article | Bearing | Note |
|---|---|---|
| Article I — Library-First | Supports | Workflow catalog remains canonical library content; execution is delegated to host-native surfaces. |
| Article III — Docs as Source of Truth | Supports | This ADR documents the boundary before implementation work begins. |
| Article IV — Anti-Speculation | Supports | The decision prevents speculative runtime work in the wrong layer. |
| Article V — Simplicity Over Abstraction | Supports | One canonical catalog, per-host execution delegation, no hidden orchestration. |
| Article VI — Anti-Overengineering | Supports | No workflow daemon, queue, or scheduler in LazyAI core. |

This ADR does not amend the constitution.

---

## Options Considered

### Option A — LazyAI core workflow engine

Embed a workflow execution engine in the `lazyai-cli` binary: parse workflow markdown, resolve step dependencies, execute subagent calls, manage state, and report results.

- **Complexity:** High. Requires a state machine, subprocess/agent lifecycle, error recovery, and a new command surface.
- **Consistency:** Low. Conflicts with ADR-003 (LazyAI owns runtime, but runtime = setup/compile/validate, not orchestration) and ADR-005 (setup-core is the default product; runtime extras are optional modules).
- **Reversibility:** Low. A new engine would be hard to remove without breaking users who depend on it.
- **Performance impact:** Negative. Adds binary size, startup cost, and a new attack surface.
- **Team familiarity:** Low. The retired Fortnite/orchestrator engine was removed for a reason.
- **Constitution fit:** Weakens Articles IV, V, and VI. Speculative orchestration in the wrong layer.

**Why rejected:** Reintroduces the retired orchestration surface under a new name. Conflicts with every prior boundary decision. LazyAI's job is asset management and compilation, not subagent orchestration.

### Option B — External/plugin runtime (host-native delegation)

Generate thin host-tool plugins/extensions that expose workflow helpers (list, show, start) through each host's native extension mechanism. The plugin delegates execution to the host's own runtime — LazyAI never runs a subagent or manages workflow state.

- **Complexity:** Medium. Requires per-host plugin/extension codegen, but no state machine or daemon.
- **Consistency:** High. Matches ADR-003 (LazyAI generates, host executes), ADR-005 (plugin generation is setup-core), and the workflow delivery matrix in tools.md.
- **Reversibility:** High. Plugins are emitted files; remove them from the catalog and they stop being generated.
- **Performance impact:** Neutral. No new binary surface; plugins are small JS/TS files.
- **Team familiarity:** Medium. Plugin generation already exists (vibe-lab-hooks.js, Pi extensions).
- **Constitution fit:** Supports all relevant articles. Delegates execution to the host that already has a runtime.

**Why chosen:** Preserves the LazyAI boundary while enabling workflow execution through the host tool's own runtime. The plugin is a generated asset, not a new LazyAI command or daemon.

### Option C — Docs-only/catalog assets (current state)

Keep the current model: workflow markdown files are reference material, never executable. Users read them and follow steps manually.

- **Complexity:** Low. No new code.
- **Consistency:** High. Matches current `adapter_targets: [none]` and the `create-workflow` skill's explicit docs-only contract.
- **Reversibility:** High. Nothing changes.
- **Performance impact:** Neutral.
- **Team familiarity:** High.
- **Constitution fit:** Acceptable, but leaves workflow execution value unrealized.

**Why rejected as sole model:** Workflow catalog assets are valuable, but keeping them permanently docs-only leaves execution value on the table. The plugin/extension path (Option B) can deliver that value without violating the LazyAI boundary. Option C is the safe fallback for targets without a verified plugin/extension mechanism.

---

## Decision

Adopt **Option B — external/plugin runtime** as the primary workflow delivery model, with **Option C — docs-only** as the fallback for targets without a verified native extension mechanism.

Concretely:

1. **LazyAI generates and validates workflow catalog assets** — the canonical markdown files under `packages/cli/library/workflows/` remain the source of truth. LazyAI may validate workflow frontmatter, step structure, and exit gates.

2. **LazyAI generates thin host-tool plugins/extensions** that expose workflow helpers (list, show, start) through each host's native extension mechanism. These are emitted files, not a LazyAI runtime daemon.

3. **LazyAI never executes a workflow.** Execution is always delegated to the host tool's own runtime. LazyAI does not manage subagent lifecycle, workflow state, queues, or scheduling.

4. **Claude Code native `.claude/workflows/` emission** remains gated by a future capability decision (per #319, Option A was chosen — docs-only for now). If added later, it must be a distinct capability, not a universal `AssetKindWorkflows`.

5. **Plugin/extension workflow helpers** are implemented per-target only after the host's extension format is source-verified. No fake `workflows/` directories are emitted for any target.

6. **The retired runtime boundary is preserved.** No `task`, `workflow`, `orchestrator`, or `eval` command surfaces are reintroduced. No workflow daemon, queue, scheduler, judge, or RAG core enters LazyAI.

---

## Rationale

- **Prior decisions converge on this boundary.** ADR-003 says LazyAI owns runtime = setup/compile/validate, not orchestration. ADR-005 says setup-core is the default product. ADR-006 freezes the seven-target contract. Issue #317 documents the per-target delivery matrix. Issue #319 chose docs-only for Claude workflows. Issue #318 scoped plugin/extension work to verified host mechanisms. This ADR codifies the consistent thread through all of them.

- **Plugin/extension delegation is already a proven pattern.** LazyAI already generates `vibe-lab-hooks.js` for OpenCode, `.pi/extensions/*.ts` for Pi, and `.gemini/hooks/` for Antigravity. Extending this to workflow helpers is a natural evolution, not a new capability.

- **The bright line prevents scope creep.** "LazyAI generates and validates, host executes" is a simple, testable rule. Any code that would run a subagent, manage workflow state, or schedule steps belongs outside LazyAI.

- **Docs-only is the safe default.** For targets without a verified plugin/extension mechanism (Copilot, Kiro today), workflows remain docs-only. This prevents emitting unsupported directories.

**Why the rejected options were rejected:**

- **Option A (LazyAI core engine):** Reintroduces the retired orchestration surface. Conflicts with every prior boundary decision. Would make LazyAI a subagent orchestrator, which is explicitly outside its product scope.

- **Option C (permanent docs-only):** Safe but leaves value unrealized. Plugin/extension delegation (Option B) can deliver execution without violating the boundary.

---

## Consequences

**Positive:**

- Clear, testable boundary: LazyAI generates and validates, host tools execute.
- Plugin/extension workflow helpers are emitted files, not a new LazyAI runtime surface.
- Docs-only fallback prevents unsupported workflow directories.
- Prior decisions (ADRs 003, 005, 006; issues 317, 318, 319) are unified under one ownership model.

**Negative / accepted trade-offs:**

- Plugin/extension implementation requires per-host source verification before each target is enabled.
- Claude Code native `.claude/workflows/` support remains gated on a future capability decision.
- Some targets (Copilot, Kiro) have no verified plugin/extension path today and remain docs-only.
- Plugin/extension workflow helpers are not a universal solution — each host's API differs.

**Neutral:**

- The workflow catalog under `packages/cli/library/workflows/` remains the canonical source.
- `adapter_targets: [none]` in curation.yaml stays correct until a specific target's plugin/extension is implemented.
- No new `AssetKind` or `Capability` field is added until a concrete plugin/extension is implemented.

---

## What LazyAI May Generate and Validate (Without Becoming an Orchestrator)

| Activity | Allowed | Rationale |
|---|---|---|
| Canonical workflow markdown catalog | Yes | `packages/cli/library/workflows/*.md` — source of truth |
| Workflow frontmatter validation | Yes | Validate name, description, steps, exit gates — same as agent/skill validation |
| Workflow step structure validation | Yes | Check that each step names a concrete owner (skill, agent, hook, command, human gate) |
| Workflow metadata in curation.yaml | Yes | `kind: workflow, adapter_targets: [none]` — catalog management |
| Skills that describe workflow processes | Yes | `create-workflow.md`, `rpi.md`, etc. — procedural guidance |
| Hook scripts for objective gates | Yes | Already emitted: `objective-workflow-gate.sh` for Claude Code, Copilot, Antigravity |
| Plugin/extension workflow helpers | Yes | Generated files that delegate to host runtime — no LazyAI execution |
| Workflow delivery matrix documentation | Yes | `docs/concepts/tools.md` — per-target delivery model |
| Claude Code `.claude/workflows/` emission | Gated | Requires future capability decision per #319 |
| Workflow execution engine | **No** | Belongs in host tool, plugin, or separate runtime package |
| Subagent lifecycle management | **No** | LazyAI does not run subagents |
| Workflow state machine | **No** | LazyAI does not track workflow progress |
| Workflow queue or scheduler | **No** | Retired surface — not reintroduced |
| Workflow daemon | **No** | No hidden orchestration process |
| Task/workflow/orchestrator/eval commands | **No** | Retired — not reintroduced |
| Mandatory judge/scoring engine | **No** | Outside LazyAI product boundary |
| RAG core | **No** | Outside LazyAI product boundary |
| Trace daemon | **No** | Outside LazyAI product boundary |
| LangChain/LangGraph/CrewAI dependency | **No** | Outside LazyAI product boundary |
| Codex adapter | **No** | Not a supported compile target per ADR-006 |

---

## Per-Target Workflow Delivery Model

| Target | Delivery Model | Mechanism | Status |
|---|---|---|---|
| OpenCode | Plugin helper | `.opencode/plugins/lazyai-workflows.ts` | Planned per #318; requires source verification |
| Claude Code | Native `.claude/workflows/` or plugin | Gated by future capability decision | Docs-only today per #319 |
| Copilot | Docs-only | No verified plugin/extension mechanism | Docs-only |
| Pi | Extension helper | `.pi/extensions/lazyai-workflows/index.ts` | Planned per #318; requires source verification |
| OMP | Extension helper | `.omp/extensions/lazyai-workflows/...` | Planned per #318; requires source verification |
| Kiro | Docs-only | Map to `.kiro/steering` and `.kiro/specs` if verified | Docs-only today |
| Antigravity | Plugin helper | `.gemini/antigravity-cli/plugins/lazyai-workflows/` | Planned per #318; requires source verification |

---

## Reversal Conditions

Re-open this ADR if any of the following becomes true:

- A host tool's native workflow format becomes the dominant execution surface and LazyAI must embed a runtime to remain useful.
- The plugin/extension delegation model proves infeasible across all targets, forcing a core engine.
- Human review decides LazyAI should own workflow execution as a first-class product capability.
- A future ADR explicitly supersedes this boundary with a different ownership model.

---

## Implementation Pointer

- Workflow catalog: `packages/cli/library/workflows/`
- Curation metadata: `packages/cli/library/manifests/curation.yaml` (workflow entries at lines 1165–1252)
- Workflow delivery matrix: `docs/concepts/tools.md` (lines 78–88)
- Product boundaries: `docs/concepts/product-boundaries.md`
- Plugin/extension planning: `specs/issues/318-workflow-plugin-engine/plan.md`
- Claude workflow decision: `specs/issues/319-claude-workflows/decision.md`
- Capability struct (no Workflows field yet): `packages/cli/internal/adapter/capabilities.go`
- AssetKind constants (no Workflows kind yet): `packages/cli/internal/adapter/output_mapping.go`
- Create command valid types (no workflow yet): `packages/cli/cmd/create.go`

---

## Memory Update

- [ ] If this ADR is accepted, update `specs/KNOWLEDGE_MAP.md` to reference the workflow ownership decision.
- [ ] Cross-link this ADR from `docs/concepts/tools.md` workflow delivery matrix section.
- [ ] If a future implementation adds plugin/extension workflow helpers, update the per-target delivery model table above.
