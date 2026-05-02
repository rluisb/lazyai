# Task 015: Generated Knowledge + Environment-Aware Planning Guidance [P]

**Phase:** W2.A — Context + causal content  
**User Story:** [US2] Context-aware reasoning  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T016, T017, T019, T020, T022, T023

---

## Objective

Add bounded Knowledge Surface and Environment Snapshot guidance to planning/reasoning prompts so non-trivial plans capture facts, assumptions, unknowns, and environment constraints before commitments.

## Spec References

- FR-W2-008 through FR-W2-010.
- AC-N2-001, AC-D14-001, AC-D14-002.
- `plan-wave2.md` W2.A.

## Files to Change/Create

- `packages/ai-setup-go/library/fragments/reasoning-protocol.md`
- `packages/ai-setup-go/library/skills/plan.md`
- Existing prompt/skill snapshot or frontmatter tests for these files.

## Files NOT to Touch

- `packages/orchestrator/src/*`
- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- RAG/knowledge retrieval tools, MCP configs, model routing rules.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing snapshot/assertion tests proving the reasoning protocol includes a concise Knowledge Surface format.
2. Add failing tests proving planner guidance includes Environment Snapshot fields: toolchain, CI/check latency, platform, budget/token constraints, network/secrets constraints, verified/unverified assumptions.
3. Add failing negative assertions proving the content does not promise RAG, model routing, provider billing, or automatic retrieval.
4. Edit only the targeted markdown guidance.

## Done When

- [ ] Knowledge Surface guidance exists and is bounded to facts/constraints/assumptions/unknowns/evidence.
- [ ] Environment Snapshot guidance exists and requires verified/unverified labels.
- [ ] Tests/snapshots prevent Wave 3 scope leakage.
- [ ] No runtime or MCP configuration changes are made.

## Risks

- **Prompt bloat:** mitigate by requiring a concise format and only for non-trivial tasks.
- **Speculative knowledge:** mitigate by requiring evidence/citation or marking unknown.

## Constitution Check

- **Article I:** Reuse existing reasoning/planning files.
- **Article II:** Snapshot/content tests fail before markdown edits.
- **Article III:** Content maps to Wave 2 ACs.
- **Article IV:** No retrieval/model-routing automation.
- **Article V:** One bounded section is simpler than a new planning subsystem.
- **Article VI:** No new skill unless existing files cannot hold the guidance.
