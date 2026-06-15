# Adversarial Review: LazyAI Runtime Refactor — Adversarial Synthesis

Date: 2026-06-13
Reviewer: DeepSeek V4 Pro
Target Artifact: `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md`
Method: 3-round adversarial debate (Advocate vs. Skeptical), grounded in live codebase audit.

---

## Evidence Gathered (Live Audit)

Before the debate, the following was confirmed via direct codebase inspection:

### Schema Inventory (26 tables, not "20+")
`schema.go` defines: `schema_migrations`, `sessions`, `dispatches`, `decisions`, `artifacts`, `memories`, `token_log`, `parallel_tasks`, `messages`, `barriers`, `locks`, `teams`, `workflows`, `workflow_instances`, `workflow_steps`, `model_calls`, `tool_calls`, `gate_results`, `ledger_refs`, `cost_snapshots`, `checkpoints`, `eval_runs`, `eval_results`, `task_queue`, `task_claims`, `task_dlq`, `task_messages`.

### Hard CLI Dependencies (not audited in the report)
| CLI Command | Imports | Impact if removed |
|---|---|---|
| `cmd/task.go` | `runtime/taskqueue` | **Breaks** — direct import |
| `cmd/workflow.go` | `runtime/dispatch`, `runtime/workflow` | **Breaks** — direct import |
| `cmd/orchestration.go` (15.3KB) | `internal/orchestrator` | **Breaks** — full command surface |
| `cmd/mcp_setup.go` | References orchestrator MCP server | **Breaks** — scaffolds orchestrator |
| `cmd/config.go` | `DefaultAgent: "orchestrator"` | **Drift** — defaults to removed concept |
| `cmd/helpers.go` | Sets `FortniteMode` flag | **Dead path** if Fortnite removed |
| `cmd/doctor*.go` | Checks `lazyai-orchestrator` binary | **Stale checks** |
| `cmd/add.go` | Offers "Orchestrator"/"Orchestrate" options | **Stale options** |
| `cmd/message.go` | Defaults `fromAgent = "orchestrator"` | **Drift** |
| `cmd/validate_input.go` | Lists "orchestrator" as valid agent | **Drift** |
| `cmd/server.go` | Checks `.ai/orchestration/chains/` | **Stale check** |
| `cmd/list.go` | Lists `.ai/orchestration/` directory | **Stale listing** |
| `cmd/update.go` | References `.opencode/agents/orchestrator.md` | **Stale reference** |

### Fortnite Coupling in Runtime Core
- `schema.go:20`: `agent TEXT NOT NULL DEFAULT 'loop-driver'`
- `session/session.go:64`: `agent = "loop-driver"` (default fallback)
- `workflow/parser_test.go:80`: reads from `library/fortnite/workflows`

### Adapter Fortnite Surface
- `opencode.go:77-79`: `FortniteMode` → `defaultAgent = "loop-driver"`, instructions = `["AGENTS.md", "STARTUP.md"]`
- `opencode.go:102-240`: Full Fortnite installation branch (agents, scripts, workflows, STARTUP.md, AGENTS.md managed block)
- `opencode.go:241-303`: Legacy mode installs generic agents + orchestrator agent
- `adapter_test.go:1096-1263`: ~170 lines asserting Fortnite behavior (8 agents, loop-driver, STARTUP.md, scripts, workflows)
- `adapter_test.go:1382-1394`: Plain-mode assertions that loop-driver must NOT exist

### Session Handoff Gap
- `session.go`: `Start()`, `End()`, `Get()`, `List()`, `UpdateSummary()`, `AddTags()` — no `WriteHandoff()`
- No handoff markdown schema exists anywhere in the repository

---

## 🥊 Round 1: Dependency Surface vs. Removal Ambition

**Advocate:**
The report's strategic diagnosis is correct. The runtime carries 26 tables of orchestration infrastructure, 46K lines of Fortnite library content, and a separate `packages/orchestrator/` module — all of which should be runtime-agnostic adapter content, not baked into the core. The non-negotiables are well-scoped: archive before delete, survey usage, define migration before shrinking. The `loop-driver` default baked into `schema.go:20` and `session.go:64` is exactly the kind of Fortnite-foundation leakage the report correctly identifies. Removing it forces the runtime to be tool-agnostic.

**Skeptical:**
The report lists "high-confidence removals" before the dependency audit is done. My audit found **13 CLI files** referencing orchestrator concepts and **2 CLI commands** (`task.go`, `workflow.go`) with direct imports of the runtime packages marked for removal. The report's Phase 1 estimate of 4–5 days to excise 46K lines assumes these are dead code — they are not. `cmd/orchestration.go` alone is 15.3KB of live command surface. You cannot remove `runtime/taskqueue` without breaking `lazyai-cli task`. You cannot remove `internal/orchestrator` without breaking `lazyai-cli orchestration`. The "high-confidence" label is premature by at least a full dependency audit.

**Decision:**
The removal targets are directionally correct, but the confidence label is unsupported. The dependency audit must move from the pre-implementation checklist into a **blocking prerequisite** before any timeline is committed. The report's Phase 1 estimate (4–5 days) is only valid if the audit reveals zero hard dependencies — which it won't.

---

## 🥊 Round 2: Schema Migration — 26 → ~5 Tables

**Advocate:**
The schema shrink is the highest-value part of this refactor. 26 tables with foreign key cascades, partial unique indexes, and cross-table references create a maintenance burden disproportionate to their runtime value. The report correctly mandates: define V2 schema first, then migration with backup/restore, then deletion. Keeping sessions, ledger, and a few primitives is the right target. The `sessions.agent DEFAULT 'loop-driver'` is a concrete example of Fortnite leakage that migration fixes.

**Skeptical:**
The report allocates 2–3 days for Phase 3. This is dangerously insufficient. Consider what a safe migration requires:
- Dump/backup path for 26 tables with foreign key integrity
- V2 schema DDL that preserves session history, ledger integrity, and artifact references
- Migration SQL that handles the `agent DEFAULT 'loop-driver'` column — every existing session row has this value
- Restore path that can reconstruct V1 from backup if migration fails
- Integration tests covering: fresh install → V2, V1-with-data → migrate → V2, V1-with-data → migrate → rollback → V1
- The `sessions` table has FK relationships with `dispatches`, `decisions`, `artifacts`, `memories`, `token_log`, `parallel_tasks`, `messages`, `barriers`, `locks`, `checkpoints` — dropping any of these child tables requires cascade handling

A realistic estimate for this work with adequate testing is **8–12 days**, not 2–3. The report's estimate assumes the migration is trivial — it is not.

**Decision:**
Phase 3 estimate must be revised to 8–12 days. The V2 schema draft and migration SQL must be reviewed before any table is dropped. The `sessions.agent` default migration is a concrete, testable acceptance criterion.

---

## 🥊 Round 3: Adapter Decoupling and the Orchestrator Paradox

**Advocate:**
The adapter layer is the reusable innovation. Decoupling `OpenCodeAdapter.Install()` from Fortnite content preserves the canonical-source → adapter/compiler pattern while removing the Fortnite foundation assumption. The report correctly mandates: remove Fortnite and `loop-driver` defaults, rewrite tests around canonical behavior, keep `InstallToolContextFiles()` if independent. The 8 Fortnite agents, STARTUP.md, scripts, and workflows are adapter content — not runtime foundation.

**Skeptical:**
There is a paradox the report does not address. The non-Fortnite ("legacy") adapter path (lines 241–303) installs an **orchestrator agent** as the primary entry point. If Phase 1 removes `packages/orchestrator/` and `internal/orchestrator/`, what does the non-Fortnite adapter install? The report says "preserve adapters" and "remove orchestrator" — these conflict. Either:
- The orchestrator agent concept survives as a lightweight agent definition (not the heavy package), or
- The non-Fortnite adapter is redesigned to install a different primary agent

Neither option is in the report. Additionally, `cmd/helpers.go` still toggles `FortniteMode` — removing the adapter's Fortnite branch without removing the CLI flag creates a dead code path that sets a mode with no effect.

On testing: `adapter_test.go` has ~200 lines of Fortnite-coupled assertions. The report says "rewrite tests to assert neutral canonical behavior" but never defines what that behavior is. Without a concrete test contract, the rewrite risks either:
- Preserving Fortnite assumptions under new names, or
- Dropping real behavioral assertions and leaving adapter output untested

**Decision:**
Resolve the orchestrator paradox before Phase 2: either keep a lightweight orchestrator agent definition or redesign the non-Fortnite adapter path. Remove `FortniteMode` from `cmd/helpers.go` and `AdapterContext` when the Fortnite adapter branch is removed — no dead flags. Define the adapter test contract concretely: what files exist, what agents are installed, what `default_agent` is set, for each adapter mode.

---

## 📊 Confidence Assessment

| Profile | Confidence | Rationale |
|---|---|---|
| **Advocate** | **78%** | Strategic direction is correct. The bloat is real (26 tables, 46K Fortnite lines, `loop-driver` baked into schema). Non-negotiables are well-scoped. But the dependency surface is larger than the report acknowledges — 13 CLI files reference orchestrator, 2 commands directly import removal targets. |
| **Skeptical** | **48%** | Estimates are untethered from reality. Phase 1 (4–5 days) assumes zero hard dependencies — audit found 2 direct import breaks. Phase 3 (2–3 days) for 26→5 table migration with backup/restore is off by 4x. The orchestrator paradox (remove package but adapter installs orchestrator agent) is unresolved. No test contract exists for adapter rewrite. |
| **Consensus** | **63%** | Direction: sound. Execution plan: under-scoped. The gap between strategic confidence and implementation readiness is the largest risk. |

### Confidence Delta: 30 points (Advocate 78% − Skeptic 48%)

This is a **wide gap** driven by the Skeptic's finding that the dependency audit — listed as a pre-implementation checklist item — is actually the critical path. Until it's done, every phase estimate is speculative.

---

## 🕳️ Point Gaps (Unresolved Disagreements)

1. **Dependency surface unquantified.** `cmd/task.go` imports `runtime/taskqueue`. `cmd/workflow.go` imports `runtime/dispatch` and `runtime/workflow`. `cmd/orchestration.go` (15.3KB) imports `internal/orchestrator`. 10 additional CLI files reference orchestrator concepts. The report's Phase 1 estimate assumes these are removable without breakage — they are not.

2. **Schema migration complexity underestimated.** 26 tables (not "20+") with FK cascades, partial unique indexes, and the `sessions.agent DEFAULT 'loop-driver'` baked-in default. The report's 2–3 day estimate is off by ~4x. A safe migration with backup/restore/testing is 8–12 days.

3. **Orchestrator paradox unresolved.** Phase 1 removes `packages/orchestrator/` and `internal/orchestrator/`. But the non-Fortnite adapter path (the one that survives) installs an orchestrator agent as primary entry point. These conflict. No resolution is proposed.

4. **"Neutral canonical behavior" undefined.** The adapter test rewrite (Phase 2) requires a concrete contract: what files, agents, and config each adapter mode produces. Without this, the rewrite either preserves Fortnite assumptions or drops real assertions.

5. **FortniteMode dead path.** `cmd/helpers.go` sets `FortniteMode` on `AdapterContext`. If the Fortnite adapter branch is removed, this flag becomes a no-op. Must be removed in the same phase.

6. **`loop-driver` baked into schema and session defaults.** `schema.go:20` and `session.go:64` both default to `loop-driver`. Migration must handle this — it's not just a table drop, it's a column default change on the surviving `sessions` table.

7. **No handoff schema draft exists.** Phase 4 allocates 3–4 days to implement `WriteHandoff()` but the handoff markdown schema is undefined. Implementation cannot start without it.

---

## 🛠️ How to Improve Confidence (Target: ≥85% Consensus)

### 1. Pre-Flight Dependency Audit (BLOCKING)
Execute before any timeline commitment:
- Map every import chain from `cmd/*.go` → `internal/runtime/*` and `internal/orchestrator`
- Map every import chain from `internal/adapter/*` → `library/fortnite/*`
- Identify which CLI commands must be removed, which must be rewritten, which can stay
- **Deliverable:** Dependency matrix with breakage/rewrite/keep classification for every file

### 2. Resolve the Orchestrator Paradox
Choose one:
- **Option A:** Keep a lightweight `orchestrator.md` agent definition (not the heavy package). Remove `packages/orchestrator/` and `internal/orchestrator/`. The non-Fortnite adapter installs this lightweight agent.
- **Option B:** Redesign the non-Fortnite adapter to use a different primary agent (e.g., a generic "driver" agent). Remove all orchestrator references from CLI and adapter.
- **Deliverable:** Decision record with affected files enumerated.

### 3. Draft V2 Schema + Migration SQL
Before Phase 3 starts:
- Write the V2 schema DDL (target: sessions, ledger, artifacts, memories, token_log — ~5–7 tables)
- Write the migration SQL with backup/restore
- Test on a copy of a real user database (if available) or a synthetic one with FK-saturated data
- **Deliverable:** Reviewed schema draft + migration SQL + test plan

### 4. Draft Handoff Markdown Schema
Before Phase 4 starts:
- Define the handoff file format: sections, frontmatter, path conventions
- Agree on ownership: is the handoff per-session, per-worktree, per-project?
- **Deliverable:** Schema document with examples

### 5. Define Adapter Test Contract
Before Phase 2 starts:
- For each adapter mode, specify: which files are created, which agents are installed, what `default_agent` is set, what instructions array contains
- Write the contract as a table, not prose
- **Deliverable:** Test contract document that can be directly translated into assertions

### 6. Revise Phase Estimates
Based on audit findings:
| Phase | Report Estimate | Revised Estimate | Rationale |
|---|---|---|---|
| Phase 1 (Excision) | 4–5 days | **8–10 days** | 2 CLI commands need rewrite, 10+ files need orchestrator reference removal |
| Phase 2 (Adapter) | 5–7 days | **7–9 days** | Plus orchestrator paradox resolution, FortniteMode flag removal |
| Phase 3 (Migration) | 2–3 days | **8–12 days** | 26→~6 table migration with backup/restore/testing |
| Phase 4 (Handoff) | 3–4 days | **3–4 days** | Unchanged IF schema is drafted first |
| Phase 5 (Library) | 3–4 days | **3–4 days** | Unchanged IF budget enforcement is CI-only |
| **Total** | **17–23 days** | **29–39 days** | |

### 7. Add Rollback Acceptance Criteria
Every phase must include a rollback test:
- Phase 1: Archive is restorable → `lazyai-cli` builds and passes tests from archive
- Phase 3: V1 DB → migrate → V2 → rollback → V1 → verify data integrity
- Phase 2: Old adapter output vs. new adapter output diff is reviewed and approved

---

## Final Recommendation

The strategic direction is sound: `lazyai` should be a compact Go runtime, not a Fortnite orchestration platform. The non-negotiables are well-chosen.

**Do not proceed with the current timeline.** The dependency audit is not a checklist item — it is the critical path. Until `cmd/task.go`, `cmd/workflow.go`, and `cmd/orchestration.go` have concrete rewrite/removal plans, every phase estimate is speculative.

The safe sequence is:
1. Complete dependency audit → revise Phase 1 estimate
2. Resolve orchestrator paradox → revise Phase 2 scope
3. Draft V2 schema + migration SQL → revise Phase 3 estimate
4. Draft handoff schema → unlock Phase 4
5. Define adapter test contract → unlock Phase 2 test rewrite
6. Only then commit to a timeline and begin implementation

**Consensus confidence with these prerequisites met: projected 85–88%.**
