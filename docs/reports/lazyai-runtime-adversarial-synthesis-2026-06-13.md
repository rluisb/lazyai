# LazyAI Runtime Refactor — Adversarial Synthesis

Date: 2026-06-13  
Status: Research/review artifact only. No implementation approval granted.  
Scope: Refactor and reposition `lazyai` as the simple but powerful Go runtime for agentic CLI tools, aligned to vibe-lab principles and adapter expectations.

## Executive Decision

Proceed only after a human-approved implementation plan. The refactor direction is sound, but the safe plan is narrower than the earlier reports implied.

Consensus confidence: **75%**

- Advocate confidence: **82%**
- Skeptic confidence: **68%**
- Revised execution estimate: **17–23 working days**

Core correction: **vibe-lab is not the runtime. `lazyai` is the runtime.** vibe-lab is the baseline for principles, philosophies, and adapter behavior.

## Non-Negotiables

1. `lazyai` owns the runtime.
2. vibe-lab supplies baseline principles, philosophy, and adapter expectations.
3. Do not port vibe-lab bash scripts into Go as the runtime.
4. Preserve the canonical-source → adapter/compiler pattern inside `lazyai`.
5. Remove Fortnite as the foundation, but archive before deletion.
6. Survey actual Fortnite/OpenCode usage before destructive removal.
7. Define migration, backup, and restore before shrinking runtime schema.
8. Define handoff schema before adding `WriteHandoff()`.
9. Enforce any canonical-library size budget in CI or pre-commit; advisory limits will regress.

## Evidence Reviewed

Observed repository metrics from the prior review session:

- `packages/cli`: **64,268** Go lines
- `packages/cli/internal/adapter/`: **10,059** Go lines
- `packages/cli/internal/runtime/`: **4,295** Go lines
- Runtime core subset: **1,331** lines
- Runtime orchestration subset: **2,964** lines
- `packages/cli/internal/orchestrator/`: **1,261** Go lines
- `packages/orchestrator/`: **18,175** Go lines
- `packages/cli/library/fortnite/`: **146** files, **28,608** lines
- Fortnite bash scripts: **27** files, **7,184** lines
- Canonical agents: **8** files
- Canonical skills: **33** files
- Fortnite Go references in `packages/cli/internal`: **41**
- `loop-driver` references: **17**
- CLI orchestration command lines: **1,836**

Important files and symbols:

- `packages/cli/library/fortnite/` — OpenCode/Fortnite runtime to remove or archive.
- `packages/orchestrator/` — separate Go orchestrator module to remove or archive.
- `packages/cli/internal/orchestrator/` — internal orchestration/catalog integration to audit.
- `packages/cli/internal/adapter/opencode.go` — contains Fortnite coupling and `loop-driver` default behavior.
- `packages/cli/internal/adapter/shared.go` — contains `InstallToolContextFiles()`.
- `packages/cli/internal/adapter/adapter_test.go` — 1,494 lines; heavily asserts Fortnite behavior.
- `packages/cli/internal/runtime/schema.go` — 412-line SQL schema with 20+ tables.
- `packages/cli/internal/runtime/session/session.go` — has `Start()`, `End()`, `Get()` but no `WriteHandoff()`.
- `packages/cli/internal/runtime/session/dispatch.go`
- `packages/cli/internal/runtime/session/parallel.go`
- `packages/cli/internal/runtime/session/barrier.go`
- `packages/cli/internal/runtime/session/lock.go`
- `packages/cli/internal/runtime/session/message.go`
- `packages/cli/internal/runtime/workflow/`
- `packages/cli/internal/runtime/taskqueue/`
- `packages/cli/internal/runtime/dispatch/`

## Five-Round Adversarial Review Synthesis

This section preserves the synthesis, not a verbatim transcript.

### Round 1 — Runtime Framing

Advocate:

- `lazyai` should become the focused Go runtime for agentic CLI tools.
- vibe-lab is valuable as a behavioral baseline: small surface area, clear adapters, canonical-source compilation, and tool-specific output.
- The current LazyAI shape is too broad because OpenCode/Fortnite orchestration has become runtime foundation instead of optional adapter content.

Skeptic:

- Earlier reports made a category error by treating vibe-lab as the runtime.
- If that framing survives, the implementation will drift toward cloning vibe-lab bash scripts in Go.
- The correct target is not “vibe-lab runtime”; it is “LazyAI runtime aligned to vibe-lab principles/adapters.”

Decision:

- Reframe every future artifact around `lazyai` as runtime.
- Use vibe-lab only as baseline/specification layer.
- Do not implement vibe-lab bash behavior as runtime internals.

### Round 2 — Removal Scope

Advocate:

- Remove Fortnite as foundation.
- Remove or archive `packages/orchestrator/`.
- Remove workflow/taskqueue/dispatch runtime packages.
- Remove complex session coordination: barriers, locks, parallel execution, message bus, task queue, workflow engine.
- Keep the useful runtime core: sessions, ledger, database/migration primitives.

Skeptic:

- This is destructive and can break existing users.
- CLI commands may import orchestration packages directly.
- The OpenCode adapter and tests currently assert Fortnite behavior.
- Deleting before archiving loses rollback/reference material.

Decision:

- Removal targets are high-confidence.
- Implementation confidence is medium until command/import usage is audited.
- Archive Fortnite/orchestrator before deletion.
- Audit all `packages/cli/cmd/*.go` commands before removing runtime packages.

### Round 3 — Adapter Strategy

Advocate:

- `packages/cli/internal/adapter/` is the core reusable innovation.
- Preserve canonical-source → adapter/compiler flow.
- Keep tool-specific outputs for OpenCode, Claude Code, Copilot, MCP, and shared context files.
- Decouple `OpenCodeAdapter.Install()` from bundled Fortnite content.

Skeptic:

- `adapter_test.go` heavily encodes Fortnite assumptions.
- `OpenCodeAdapter` currently carries `loop-driver` defaults.
- Removing Fortnite without rewriting adapter tests can either fail the suite or preserve accidental behavior.
- Avoid replacing real tool-specific adapter behavior with a lowest-common-denominator compiler.

Decision:

- Preserve adapters.
- Remove Fortnite and `loop-driver` defaults from adapter output.
- Rewrite tests to assert neutral canonical behavior and tool-specific adapter contracts, not Fortnite history.
- Keep `InstallToolContextFiles()` if it remains a simple shared context installer.

### Round 4 — Runtime Data Model

Advocate:

- Shrink runtime schema from orchestration-heavy state to simple durable runtime state.
- Keep sessions and ledger.
- Add simple agent handoff capability.
- Target roughly five core tables instead of 20+ orchestration tables.

Skeptic:

- Schema deletion without migration can destroy user data.
- `WriteHandoff()` cannot be safely added until the handoff markdown schema and path semantics are defined.
- Removing dispatch/workflow/taskqueue must account for all CLI commands and integration tests.

Decision:

- Define V2 runtime schema before deleting tables.
- Define migration, backup, and restore contract before shrinking schema.
- Define handoff markdown schema before adding `WriteHandoff(path string) error`.
- Keep simple handoff capability; remove complex coordination infrastructure.

### Round 5 — Library Curation and Enforcement

Advocate:

- Replace arbitrary “number of skills” limits with token-rent enforcement.
- Prefer a byte budget such as ≤50KB total canonical agents+skills+hooks.
- Remove heavyweight bundled Fortnite library content.

Skeptic:

- A size limit without automated enforcement is cosmetic.
- Curation based only on aesthetics can remove workflows users actually need.
- The right source of truth is supported LazyAI workflows, not the current Fortnite inventory.

Decision:

- Use a token-rent gate instead of arbitrary skill count.
- Enforce the budget in CI or pre-commit.
- Curate against actual supported CLI-agent workflows after usage survey.

## Narrowed Refactor Plan

No implementation should start until this plan is expanded into an explicit, human-approved implementation plan.

### Phase 0 — Approval Gate

Deliverables:

- Final implementation plan with affected commands, imports, files, tests, migration steps, rollback steps, and acceptance criteria.
- Human approval before code changes.

Gate:

- Required because this is a destructive refactor.

### Phase 1 — Excision and Archival

Estimate: **4–5 days**

Work:

- Archive `packages/cli/library/fortnite/` for rollback/reference.
- Archive or remove `packages/orchestrator/` after confirming no required callers remain.
- Audit `packages/cli/cmd/*.go` for dependencies on orchestration packages.
- Remove workflow/taskqueue/dispatch/barrier/lock/message infrastructure only after callsites are migrated or deleted.

High-confidence removals:

- Fortnite runtime foundation.
- Separate orchestrator package.
- Workflow/taskqueue/dispatch runtime packages.
- Complex session coordination infrastructure.

Risks:

- Hidden command dependencies.
- Existing OpenCode users relying on bundled Fortnite defaults.
- Tests preserving historical behavior unintentionally.

### Phase 2 — Adapter Surgery

Estimate: **5–7 days**

Work:

- Decouple `OpenCodeAdapter.Install()` from Fortnite content.
- Remove `loop-driver` default behavior from runtime/adapter paths.
- Preserve adapter-specific output conventions.
- Rewrite adapter tests around canonical behavior and each target tool’s expected files.
- Keep `InstallToolContextFiles()` only if it remains independent of Fortnite assumptions.

Risks:

- Breaking adapter output for OpenCode.
- Over-generalizing adapters and losing per-tool behavior.
- Tests turning into snapshots of accidental defaults.

### Phase 3 — Data Migration and Schema Shrink

Estimate: **2–3 days**

Work:

- Define V2 runtime schema.
- Keep only durable primitives needed for simple runtime behavior.
- Add migration with backup and restore path.
- Remove orchestration tables only through explicit migration.

Risks:

- Data loss.
- Backward-incompatible state changes without diagnostics.
- Integration tests missing migration edge cases.

### Phase 4 — Core Enhancement

Estimate: **3–4 days**

Work:

- Add simple handoff capability to sessions.
- Define handoff markdown schema first.
- Implement `WriteHandoff(path string) error` or equivalent only after schema agreement.
- Keep session lifecycle boring: start, get, end, handoff.

Risks:

- Reintroducing coordination complexity under a new name.
- Ambiguous handoff ownership/path semantics.
- Coupling handoff format to one adapter.

### Phase 5 — Library Curation and Enforcement

Estimate: **3–4 days**

Work:

- Replace Fortnite library with small canonical agent/skill/hook set.
- Enforce total canonical-library byte budget, proposed ≤50KB.
- Add CI/pre-commit size check.
- Keep content tied to supported LazyAI CLI-agent workflows.

Risks:

- Advisory-only budget regression.
- Removing useful workflows without usage evidence.
- Keeping too much “just in case” content and preserving token bloat.

## Final Recommendation

Proceed with the refactor only as a gated destructive cleanup. The strategic direction is correct: `lazyai` should be a compact Go runtime for agentic CLI tooling, and vibe-lab should inform principles and adapter expectations.

Do not proceed with an implementation that says or implies “vibe-lab is the runtime.” That framing is the main failure mode.

The safe plan is:

1. Reframe around `lazyai` runtime ownership.
2. Audit command/import usage.
3. Archive Fortnite/orchestrator.
4. Decouple adapters.
5. Define schema migration and handoff contract.
6. Remove orchestration-heavy runtime code.
7. Enforce canonical-library size.

## Pre-Implementation Checklist

- [ ] Replace “vibe-lab runtime” wording with “LazyAI runtime aligned to vibe-lab principles/adapters.”
- [ ] Survey actual Fortnite/OpenCode usage.
- [ ] Audit all `packages/cli/cmd/*.go` command dependencies.
- [ ] Audit imports of `packages/cli/internal/runtime/workflow`.
- [ ] Audit imports of `packages/cli/internal/runtime/taskqueue`.
- [ ] Audit imports of `packages/cli/internal/runtime/dispatch`.
- [ ] Audit imports of `packages/cli/internal/orchestrator`.
- [ ] Audit imports of `packages/orchestrator`.
- [ ] Define V2 runtime schema.
- [ ] Define migration, backup, and restore path.
- [ ] Define handoff markdown schema.
- [ ] Define tests for adapter output, migration, session handoff, and removed command behavior.
- [ ] Add canonical-library size enforcement.
- [ ] Get explicit human approval before implementation.
