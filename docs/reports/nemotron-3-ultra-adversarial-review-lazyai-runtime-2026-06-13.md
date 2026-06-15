# Nemotron-3-Ultra Adversarial Review — LazyAI Runtime Refactor

**Date:** 2026-06-13  
**Source Document:** `docs/reports/lazyai-runtime-adversarial-synthesis-2026-06-13.md`  
**Review Method:** 3-round adversarial synthesis (Advocate vs Skeptic subagents)  
**Document Claimed Confidence:** 75%  
**Actual Assessed Confidence:** 64%

---

## Executive Summary

The synthesis document correctly identifies the strategic direction: `lazyai` should be the Go runtime for agentic CLI tools, with vibe-lab providing principles and adapter expectations — not runtime implementation. However, the implementation plan has critical gaps that reduce true executability confidence to **64%**, well below the documented 75%.

**Verdict: Do not approve Phase 0 until 8 prerequisites are satisfied with evidence.**

---

## Round 1 — Position Assessment

### Advocate (82% confidence contribution)

**Strengths confirmed:**
- Correct framing correction: vibe-lab ≠ runtime (Non-negotiable #3)
- Concrete, enforceable non-negotiables (9 items)
- Explicit Phase 0 human approval gate
- Archive-before-delete mandate
- Adapter preservation with Fortnite decoupling
- Migration-first schema approach
- Handoff schema before implementation
- CI-enforced size budget (not advisory)
- 14-item pre-implementation checklist

### Skeptic (68% confidence contribution)

**Critical gaps:**
1. Zero usage survey evidence for "Survey actual Fortnite/OpenCode usage"
2. Command audit listed but no results (1,836 orchestration command lines)
3. Adapter test rewrite underestimated: 1,494 lines of Fortnite-coupled tests
4. V2 schema undefined — "5 core tables" is a target, not a design
5. Handoff schema undefined — `WriteHandoff(path string)` signature assumes nonexistent design
6. No migration path for existing users
7. Orchestrator package removal conditional on unaudited "no required callers"
8. Canonical library content unspecified
9. Token-rent enforcement: no tool, threshold, or integration point
10. Archive ≠ rollback procedure

---

## Round 2 — Cross-Examination & Gap Analysis

### Assumption Stress Tests

| Assumption | Evidence For | Evidence Against | Verdict |
|------------|--------------|------------------|---------|
| Fortnite removable foundation | 41 Go refs, 27 bash scripts, 28K lines | Adapter tests assert Fortnite behavior; 1,836 CLI command lines | Removable **but** adapter rewrite must precede excision |
| vibe-lab principles map to runtime | Small surface, clear adapters, canonical compilation | vibe-lab is bash; mapping bash→Go is exactly the trap warned against | Principles **yes**, implementation patterns **no** |
| 5 core tables sufficient | Current: 20+ tables | Missing: agent/skill/hook registry, tool context, session metadata | **Underspecified** — design must precede Phase 3 |
| Adapter tests rewrite in 5-7 days | — | 1,494 lines; needs new fixtures, contract tests, migration tests | **Likely 10-14 days** (2x estimate) |
| CLI audit straightforward | — | Unknown import graph across 5+ package paths | **Must complete before Phase 0** |

### Position Gap Matrix

| Dimension | Advocate | Skeptic | Gap |
|-----------|----------|---------|-----|
| Removal confidence | High (metrics) | Medium (usage unknown) | **Usage survey missing** |
| Adapter decoupling | Preserved, tests rewritten | Tests Fortnite-coupled, rewrite 2x | **Test scope 2x larger** |
| Schema migration | Define V2, migrate | No V2 design, no tooling, no rollback | **Design absent** |
| Handoff capability | Simple add after schema | Schema undefined, ownership ambiguous | **Design absent** |
| Library curation | ≤50KB, CI enforced | Which content? Usage-based? Tool missing | **Content + tooling absent** |
| Phase ordering | Sequential 0→5 | Tests block excision; schema blocks Phase 3 | **Dependencies inverted** |
| Rollback | Archive = safety | Archive ≠ rollback procedure | **Procedure missing** |

**Structural Issue**: Plan treats Phase 1 (excision) as independent, but adapter tests and CLI commands **depend on packages being excised**. Correct dependency order:
```
Audit → Adapter rewrite → Test rewrite → Excision → Schema migration → Library curation
```
Not the current sequential 1→2→3→4→5.

---

## Round 3 — Synthesis & Confidence Scoring

### Confidence Breakdown

| Metric | Weight | Score | Rationale |
|--------|--------|-------|-----------|
| Strategic direction | 20% | 85% | Correct framing, clear ownership, sound principles |
| Plan completeness | 25% | 55% | 5 critical designs missing (schema, handoff, migration, library, token-rent) |
| Risk mitigation | 20% | 60% | Archive ≠ rollback; usage survey absent; dependency ordering inverted |
| Estimate realism | 20% | 50% | Adapter rewrite 2x underestimated; schema/handoff design 0 days allocated |
| Gate enforceability | 15% | 70% | Phase 0 gate good; checklist items lack completion evidence |

**Composite: 64%** (document claims 75% — optimism from treating prerequisites as done)

### Estimate Correction

| | Document | Realistic |
|---|----------|-----------|
| Total | 17–23 days | **25–35 days** |
| Adapter test rewrite | 5–7 days | 10–14 days |
| Schema/handoff design | 0 days | 5–7 days |
| Prerequisites (1-8) | 0 days | 5–7 days |

---

## 8 Prerequisites Before Phase 0 Approval

| # | Requirement | Evidence Required |
|---|-------------|-------------------|
| 1 | **CLI command import audit** | Per-file matrix: every `packages/cli/cmd/*.go` → imports from `runtime/workflow`, `runtime/taskqueue`, `runtime/dispatch`, `internal/orchestrator`, `packages/orchestrator`, `library/fortnite` |
| 2 | **Fortnite/OpenCode usage survey** | Active user count, workflow usage breakdown, migration blockers |
| 3 | **V2 runtime schema design** | ER diagram + SQL (5-7 core tables) + migration up/down scripts |
| 4 | **Handoff markdown schema** | Frontmatter, sections, path semantics, ownership model |
| 5 | **Canonical library specification** | Agent/skill/hook list with usage justification per item |
| 6 | **Token-rent CI enforcement design** | Tool choice, byte threshold, integration point (pre-commit/CI), failure mode |
| 7 | **Corrected phase dependency order** | Revised plan: Audit → Adapter rewrite → Test rewrite → Excision → Schema → Library |
| 8 | **Rollback procedure** | Tagged releases, migration reversal steps, user notification template, data restore verification |

---

## 90% Confidence Additions (Recommended for Destructive Refactor)

| # | Validation | Evidence |
|---|------------|----------|
| 9 | Prototype adapter rewrite (1 target) | Working `ClaudeCodeAdapter` with canonical source → correct output + contract tests |
| 10 | Prototype V2 schema migration | Migration up/down clean; session/ledger/handoff round-trips on test DB |
| 11 | Zero breaking changes validated | Test matrix: OpenCode (±Fortnite), Claude Code, Copilot, MCP |
| 12 | Token-rent CI load test | Pipeline adds ≤30s; catches budget overflow reliably |

---

## Conclusion

**Strategic direction is correct** — `lazyai` as runtime, vibe-lab as principles layer.

**Plan is not executable** — 5 designs missing, dependency ordering inverted, 8 evidence-gated prerequisites incomplete.

**Action**: Complete all 8 prerequisites with documented evidence, then re-assess. The refactor should proceed, but only on a plan that reflects actual dependency order and design completeness.