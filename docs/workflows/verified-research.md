---
name: verified-research
description: Disciplined investigation methodology with verification gates and append-only contribution back. Nine-phase pipeline that plans, parallelizes, verifies citations, cross-references authoritative team docs, re-baselines when needed, runs adversarial review, and contributes findings without overwriting others' work.
status: draft
---

# Verified Research

## Trigger

Use when:
- Investigating a system end-to-end (how does X work today?)
- Comparing two systems (how does X differ from Y?)
- Proposing a migration or change with traceable evidence
- Producing research that must hold up under engineering or product review
- Contributing findings back to existing team documentation without disrupting it

Do not use when:
- Implementation tasks (use `build-mode`)
- Pre-implementation specs from scratch (use `storm-scout`)
- One-off Q&A or ad-hoc lookups
- Code review of a specific PR (use `zero-point` review mode)

## Inputs

- The core question or target system to investigate.

## Purpose

Investigation that survives cross-examination. Most research fails because it (1) duplicates existing work, (2) accepts unverified citations, (3) ignores authoritative team docs, or (4) overwrites the work of others when contributing back. This workflow makes each of those a phase gate.

## Four Points

- WHAT: A verified, synthesis report on a system, optionally contributing to canonical docs.
- HOW: A 9-phase methodology orchestrating researchers, verifiers, and adversarial reviews.
- DON'T WANT: Duplicated work, hallucinated citations, overwritten team docs, patches glued onto someone else's work.
- VALIDATE: Run the pre-flight validation checklist before declaring "done".

## Behavior Scenario

- Given a complex technical question or proposal
- When the `/verified-research` workflow is triggered
- Then a 9-phase methodology is executed resulting in verified findings, alignment with existing team docs, and append-only contributions.

## TDD Mode

n/a

## Steps

1. **Phase 0 — Plan**: Plan approved by user; existing-work search complete. **(Must-pass Gate)**
2. **Phase 1 — Dispatch**: All tracks dispatched in parallel with explicit budgets and anchors.
3. **Phase 2 — Handle failures**: All track findings captured (file written or content compiled).
4. **Phase 3 — Synthesize**: Unified report exists with TL;DR, viewpoints, mapping, recommendation, open questions.
5. **Phase 4 — Verify (drift-scope)**: Every code claim has line-level citation; drift marked and corrected in place. **(Must-pass Gate)**
6. **Phase 5 — Cross-reference authoritative docs**: Team-authored docs fetched and compared; alignment file written if divergence found. **(Must-pass Gate)**
7. **Phase 6 — Re-baseline (if needed)**: New primary doc written; original preserved; primary marker explicit.
8. **Phase 7 — Adversarial review**: 3+ rounds of Advocate vs Skeptic; resolutions captured; unresolved tensions in open questions.
9. **Phase 8 — Contribute back (if applicable)**: Drop-in content matches source tone; append-only; rationale separate; review path provided.

## Adapters

- Claude Code: Coordinator agent orchestrates Researcher and Verifier sub-agents.
- OpenCode: Skill / plugin mapping
- OMP/Pi: Markdown instructions provided to the agent.

## Exit Gate

Stop only when:
1. The 9-phase pipeline is complete and the validation checklist at `.agents/workflows/verified-research/templates/checklist.md` is passed.
2. Verified findings are produced and drift-scope verification has passed.
3. Out-of-scope discoveries are called out in open questions instead of silently absorbed.

## Failure

Stop and report evidence when:
- Cannot locate authoritative source material.
- Drift verification finds that claims do not match the source files, and corrections cannot be successfully applied.
- The coordinator fails to resolve adversarial tensions (capture the failure as open questions).

---

## Templates

| File | Purpose |
|---|---|
| `templates/prompt-template.md` | Drop into a new session; fill `{{placeholders}}` to run the methodology |
| `templates/checklist.md` | Pre-flight validation checklist; every box must be ticked before "done" |
| `templates/artifact-set.md` | Standard artifact set produced by the methodology |

Templates live in `.agents/workflows/verified-research/templates/`.

---

## Lessons That Shaped This Methodology

1. **Search existing work BEFORE dispatching.** Independent research often duplicates work the team has already done. → **Phase 0**
2. **Read-only sub-agents cannot write files.** Grant explicit one-file write authorization OR have the coordinator do the write. → **Phase 2**
3. **Step and token limits are real.** Plan with smaller, focused dispatches and seed-context re-dispatch when needed. → **Phase 2**
4. **Drift verification finds real drift.** Wrong file references, wrong call chains, wrong test locations are routine in unverified research. → **Phase 4**
5. **External authoritative docs expose material gaps.** Independent research often misses architectural details that team-authored docs make explicit. → **Phase 5**
6. **Re-baseline when a better source appears.** Write a new primary doc rather than patching the old one. Preserves history; signals which doc is canonical. → **Phase 6**
7. **Adversarial review surfaces silent risks.** Latency cascades, fragility of human-managed inputs, silent state divergence rarely emerge without structured opposition. → **Phase 7**
8. **Match tone explicitly when contributing.** Voice mismatches make contributions feel like patches glued onto someone else's work. → **Phase 8**
