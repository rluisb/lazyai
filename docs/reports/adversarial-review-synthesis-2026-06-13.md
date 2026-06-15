# Adversarial Review Synthesis — LazyAI Runtime Refactor

**Date:** 2026-06-13
**Review Method:** 3-round adversarial synthesis (Advocate vs Skeptic subagents, cross-examined)
**Target Artifact:** `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md`
**Prior Reviews Considered:** Nemotron-3-Ultra, Gemini-3.1, DeepSeek-V4-Pro

---

## Executive Summary

The LazyAI Runtime Refactor plan is **strategically correct but not executable as written**. Both Advocate and Skeptic agree on the core direction — `lazyai` as runtime, `vibe-lab` as principles layer — but diverge sharply on executability. The Skeptic's live audit of the repository revealed dependency surface far larger than the plan acknowledges, and both profiles independently identified the same missing designs (V2 schema, handoff schema, adapter test contract, token-rent tooling).

**Verdict: Do not approve Phase 0 until 8 evidence-gated prerequisites are satisfied.**

---

## Round 1 — Position Assessment

### Advocate (71% confidence)

**Strengths confirmed:**
- Correct strategic reframe: `lazyai` is the runtime, `vibe-lab` is the principles layer
- Fortnite/orchestrator bloat is real and measurable (46K lines, 26 tables)
- Schema-first migration discipline prevents data loss
- Adapter layer identified as reusable core
- Token-rent governance replaces arbitrary skill counts
- Human approval gate for destructive work

**Weaknesses acknowledged:**
- Dependency audit is qualitative, not quantitative
- Timelines are directional, not evidence-based
- Rollback acceptance criteria under-specified
- Confidence drops from synthesis's 75% to Advocate's own 71%

### Skeptic (55% confidence)

**Critical gaps identified:**
1. **Dependency surface unquantified** — 13 CLI files reference orchestrator concepts; 3 commands have direct import breaks (`cmd/task.go` → `runtime/taskqueue`, `cmd/workflow.go` → `runtime/dispatch`/`runtime/workflow`, `cmd/orchestration.go` → `internal/orchestrator`)
2. **V2 runtime schema undefined** — "5 core tables" is a target, not a design; current `schema.go` has 412 lines defining 26 tables
3. **Handoff markdown schema does not exist** — `WriteHandoff(path string) error` assumes a design not written
4. **Adapter test contract undefined** — 1,494 lines of `adapter_test.go`, ~170 lines asserting Fortnite behavior, no definition of "neutral canonical behavior"
5. **Orchestrator paradox** — non-Fortnite adapter installs an orchestrator agent, but orchestrator packages are scheduled for removal
6. **Token-rent enforcement tool missing** — no tool, threshold, or integration point

---

## Round 2 — Cross-Examination & Gap Analysis

### Confidence Score Comparison

| Dimension | Advocate | Skeptic | Gap | Resolution |
|-----------|----------|---------|-----|------------|
| Strategic direction | 90% | 85% | 5% | **Agreement.** Both confirm the reframe is correct. |
| Plan completeness | 80% (scope clarity) | 45% | 35% | **Skeptic wins.** 5 critical designs absent; Advocate's "scope clarity" score conflates bounded phases with complete designs. |
| Evidence base | 55% | 50% (risk mitigation) | 5% | **Agreement.** Both acknowledge the dependency audit is missing. |
| Executability / estimates | 60% | 40% | 20% | **Skeptic wins.** Live audit found 13 CLI files with orchestrator references; estimates assume prerequisites are free. |
| Gate enforceability | 75% (risk mitigation) | 65% | 10% | **Partial agreement.** Phase 0 gate is structurally good, but checklist items lack acceptance criteria. |

### Composite Confidence: **63%**

Weighted average across both profiles, with Skeptic's evidence-weighted scores given higher weight on executability dimensions:

| Metric | Weight | Score | Rationale |
|--------|--------|-------|-----------|
| Strategic direction | 20% | 87% | Both profiles agree; strongest dimension |
| Plan completeness | 25% | 50% | 5 designs missing; Skeptic's 45% is evidence-backed |
| Risk mitigation | 20% | 55% | Archive ≠ rollback; usage survey absent; dependency ordering inverted |
| Estimate realism | 20% | 45% | Skeptic's live audit proves estimates 40-70% optimistic |
| Gate enforceability | 15% | 67% | Phase 0 gate good; checklist lacks evidence requirements |

**Composite: 63%** — below the synthesis document's claimed 75%, above Nemotron's 64%, and between Advocate's 71% and Skeptic's 55%.

---

## Round 3 — Point Gaps & Resolution

### Gap 1: Dependency Audit — Critical Path, Not Checklist Item

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Audit is framed as Phase 0 prerequisite; plan names the gate | Audit is not performed; 13 CLI files found with orchestrator references; 3 commands have direct import breaks |
| Evidence | Synthesis §Phase 0 lists "audit all cmd/*.go command dependencies" | Live audit: `cmd/task.go`, `cmd/workflow.go`, `cmd/orchestration.go` (15.3KB), `cmd/mcp_setup.go`, `cmd/config.go`, `cmd/helpers.go`, `cmd/doctor*.go`, `cmd/add.go`, `cmd/message.go`, `cmd/validate_input.go`, `cmd/server.go`, `cmd/list.go`, `cmd/update.go` |

**Resolution: Skeptic wins.** The audit is not a checklist item — it is the critical path that determines whether Phase 1 is 4-5 days or 8-10 days. The plan lists it but has not performed it. The Skeptic's live audit proves the dependency surface is larger than acknowledged.

### Gap 2: Schema Design — Prerequisite vs In-Phase Work

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Plan mandates "define V2 schema before deleting tables" — structure supports revising at Phase 0 gate | "Define V2 runtime schema" is listed as Phase 3 work, not a prerequisite; 26→5 table migration with FK cascades needs 8-12 days, not 2-3 |
| Evidence | Synthesis §Phase 3, Non-negotiable #4 | Current `schema.go`: 412 lines, 26 tables, FK cascades, partial unique indexes |

**Resolution: Skeptic wins.** Schema design must be a Phase 0 prerequisite, not Phase 3 work. You cannot estimate migration work without knowing the target schema. The plan allocates zero days to schema design.

### Gap 3: Estimate Realism

| Phase | Synthesis Claim | Advocate | Skeptic | Consensus |
|-------|-----------------|----------|---------|-----------|
| Phase 0 (prerequisites) | Implicit/0 days | Implicit | 5-7 days | **5-7 days** |
| Phase 1 (excision) | 4-5 days | Accepts with revision | 8-10 days | **8-10 days** |
| Phase 2 (adapter surgery) | 5-7 days | Accepts | 10-14 days | **10-14 days** |
| Phase 3 (migration) | 2-3 days | 8-12 days | 8-12 days | **8-12 days** |
| Phase 4 (handoff) | 3-4 days | 3-4 days (if schema ready) | 3-4 days (if schema ready) | **3-4 days** (conditional) |
| Phase 5 (library) | 3-4 days | 3-4 days (if budget design ready) | 3-4 days (if budget design ready) | **3-4 days** (conditional) |
| **Total** | **17-23 days** | **~30 days** | **29-39 days** | **30-40 days** |

**Resolution: Skeptic wins on total.** Both profiles independently converge on ~30-40 days when prerequisites are counted. The synthesis's 17-23 days assumes zero-cost design work.

### Gap 4: Phase Ordering

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Plan's sequential ordering is reasonable with Phase 0 gate | Phase sequence is inverted: adapter tests depend on packages scheduled for removal; CLI commands import packages Phase 2 must first decouple |

**Resolution: Skeptic wins.** Corrected order:
```
Audit → Adapter rewrite → Test rewrite → Excision → Schema migration → Library curation
```
The plan's Phase 1→2→3→4→5 ordering cannot work because Phase 1 removes packages that Phase 2's adapter tests and CLI commands depend on.

### Gap 5: Orchestrator Paradox

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Can resolve by distinguishing lightweight agent definition from heavy orchestrator package | Non-Fortnite adapter (`opencode.go:241-303`) installs orchestrator agent; removing orchestrator packages while preserving adapters is contradictory |

**Resolution: Unresolved — requires explicit design decision.** The plan must state whether the non-Fortnite adapter installs a lightweight `orchestrator.md` agent definition or follows a redesigned path. `FortniteMode` in `cmd/helpers.go` and `DefaultAgent: "orchestrator"` in `cmd/config.go` must be addressed.

### Gap 6: Token-Rent Enforcement

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Budget is a default governance mechanism; can add `.lazyai` local override | No tool, threshold, or integration point exists; budget is currently arbitrary and unenforceable |

**Resolution: Skeptic wins on tooling gap; Advocate's override suggestion is correct.** Both are needed: (a) CI/pre-commit enforcement tool with byte threshold, and (b) `.lazyai` configuration override for documented exceptions.

### Gap 7: Rollback Procedure

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Archive-before-delete limits blast radius; rollback criteria should be added per phase | "Archive" is not a rollback procedure; no tagged releases, migration reversal steps, or data restore verification |

**Resolution: Skeptic wins.** Archive is necessary but not sufficient. Rollback requires: tagged releases, migration reversal scripts, user notification template, and data restore verification.

### Gap 8: Library Budget Override

| Position | Advocate | Skeptic |
|----------|----------|---------|
| Claim | Add `.lazyai` local configuration option to exceed 50KB default | ≤50KB limit is arbitrary; lacks opt-out for enterprise users |

**Resolution: Agreement.** Both profiles recommend an override mechanism. The default should be ≤50KB with CI enforcement; a `.lazyai` config should allow documented exceptions.

---

## Point Gap Summary

| # | Gap | Advocate Position | Skeptic Position | Resolution | Severity |
|---|-----|-------------------|------------------|------------|----------|
| 1 | Dependency audit | Phase 0 prerequisite (named, not done) | Critical path; 13 CLI files found | **Skeptic wins** — audit must be completed before estimation | **Critical** |
| 2 | Schema design | Structure supports revision at gate | Prerequisite, not in-phase work | **Skeptic wins** — design must precede Phase 3 | **Critical** |
| 3 | Estimate realism | Directional, needs revision | 40-70% optimistic | **Skeptic wins** — 30-40 days realistic | **High** |
| 4 | Phase ordering | Sequential with gate | Inverted dependencies | **Skeptic wins** — corrected order required | **Critical** |
| 5 | Orchestrator paradox | Distinguish agent def from package | Contradictory goals | **Unresolved** — explicit design decision needed | **High** |
| 6 | Token-rent tooling | Add override mechanism | No tool exists | **Both correct** — tool + override needed | **High** |
| 7 | Rollback procedure | Add per-phase criteria | Archive ≠ rollback | **Skeptic wins** — full procedure needed | **High** |
| 8 | Library budget override | Add `.lazyai` config option | Arbitrary, lacks opt-out | **Agreement** — override needed | **Medium** |

---

## How to Improve Confidence

To raise confidence from **63% to ≥85%**, complete the following before Phase 0 approval:

### 8 Evidence-Gated Prerequisites

| # | Prerequisite | Evidence Required | Raises Confidence To |
|---|--------------|-------------------|---------------------|
| 1 | **CLI command import audit** | Per-file matrix: every `packages/cli/cmd/*.go` → imports from `runtime/workflow`, `runtime/taskqueue`, `runtime/dispatch`, `internal/orchestrator`, `packages/orchestrator`, `library/fortnite` | 68% |
| 2 | **Fortnite/OpenCode usage survey** | Active user count, workflow usage breakdown, migration blockers | 71% |
| 3 | **V2 runtime schema design** | ER diagram + SQL DDL (5-7 core tables) + migration up/down scripts | 75% |
| 4 | **Handoff markdown schema** | Frontmatter spec, sections, path conventions, ownership model | 78% |
| 5 | **Canonical library specification** | Agent/skill/hook list with usage justification per item | 80% |
| 6 | **Token-rent CI enforcement design** | Tool choice, byte threshold, integration point (pre-commit/CI), failure mode | 82% |
| 7 | **Corrected phase dependency order** | Revised plan: Audit → Adapter rewrite → Test rewrite → Excision → Schema → Library | 84% |
| 8 | **Rollback procedure** | Tagged releases, migration reversal steps, user notification template, data restore verification | 85% |

### 90% Confidence Additions (Recommended for Destructive Refactor)

| # | Validation | Evidence |
|---|------------|----------|
| 9 | Prototype adapter rewrite (1 target) | Working `ClaudeCodeAdapter` with canonical source → correct output + contract tests |
| 10 | Prototype V2 schema migration | Migration up/down clean; session/ledger/handoff round-trips on test DB |
| 11 | Zero breaking changes validated | Test matrix: OpenCode (±Fortnite), Claude Code, Copilot, MCP |
| 12 | Token-rent CI load test | Pipeline adds ≤30s; catches budget overflow reliably |

---

## Corrected Phase Plan

```
Phase 0: Prerequisites (5-7 days)
├── 1. CLI command import audit
├── 2. Fortnite/OpenCode usage survey
├── 3. V2 runtime schema design (ER + DDL + migration scripts)
├── 4. Handoff markdown schema draft
├── 5. Canonical library specification
├── 6. Token-rent CI enforcement design
├── 7. Corrected dependency order plan
├── 8. Rollback procedure
└── ⛔ Human approval gate

Phase 1: Adapter Decouple & Test Rewrite (10-14 days)
├── Define adapter test contract ("neutral canonical behavior")
├── Rewrite adapter_test.go (1,494 lines → canonical assertions)
├── Resolve orchestrator paradox (agent def vs package removal)
├── Remove FortniteMode from cmd/helpers.go
├── Update DefaultAgent in cmd/config.go
└── ⛔ Checkpoint

Phase 2: CLI Command Audit & Excision (8-10 days)
├── Rewrite/remove affected CLI commands (task, workflow, orchestration)
├── Remove packages/orchestrator/, internal/orchestrator/
├── Remove runtime/taskqueue/, runtime/workflow/, runtime/dispatch/
├── Archive packages/cli/library/fortnite/
├── Clean stale references in cmd/doctor*, cmd/add, cmd/message, etc.
└── ⛔ Checkpoint

Phase 3: Schema Migration (8-12 days)
├── Execute V2 schema migration (up)
├── Backup/restore verification
├── Migration reversal testing (down)
├── Session/ledger/handoff round-trip tests
└── ⛔ Checkpoint

Phase 4: Handoff Implementation (3-4 days)
├── Implement WriteHandoff(path string) per handoff schema
├── Session integration
└── ⛔ Checkpoint

Phase 5: Library Curation & Enforcement (3-4 days)
├── Curate canonical agent/skill/hook set (≤50KB)
├── Implement CI/pre-commit byte-budget check
├── Add .lazyai config override for documented exceptions
└── ⛔ Final approval
```

**Total realistic estimate: 30-40 working days** (was 17-23 in synthesis)

---

## Conclusion

**Strategic direction is correct.** `lazyai` should be the Go runtime for agentic CLI tools; `vibe-lab` should supply principles and adapter expectations — not runtime implementation.

**Plan is not executable as written.** Five critical designs are absent (V2 schema, handoff schema, adapter test contract, canonical library spec, token-rent tooling). Phase ordering inverts actual dependencies. Estimates are 40-70% optimistic. The dependency surface is larger than acknowledged.

**Action:** Complete all 8 evidence-gated prerequisites, then re-assess. The refactor should proceed, but only on a plan that reflects actual dependency order, design completeness, and realistic estimates.

**Confidence trajectory:**
- Synthesis document claim: 75%
- Nemotron assessed: 64%
- This review (Advocate+Skeptic cross-examined): **63%**
- After 8 prerequisites: **85%**
- After 12 validations: **90%+**
