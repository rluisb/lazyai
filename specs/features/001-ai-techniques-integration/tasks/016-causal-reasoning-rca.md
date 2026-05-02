# Task 016: Causal Reasoning RCA Guidance [P]

**Phase:** W2.A — Context + causal content  
**User Story:** [US2] Context-aware reasoning  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T017, T019, T020, T022, T023

---

## Objective

Enhance bugfix/RCA guidance with a formal but lightweight causal method using 5-Whys or a short fault-tree/causal chain.

## Spec References

- FR-W2-014 through FR-W2-016.
- AC-D13-001, AC-D13-002.
- `plan-wave2.md` W2.A.

## Files to Change/Create

- `packages/ai-setup-go/library/templates/bugfix-rca-template.md`
- `packages/ai-setup-go/library/skills/bugfix.md`
- Existing template/skill snapshot or frontmatter tests.

## Files NOT to Touch

- Orchestrator runtime files.
- Feature chain JSON.
- RAG/evaluation/learning artifacts.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing template snapshot assertions for causal-chain/5-Whys, evidence, confidence, and counterfactual fields.
2. Add failing skill guidance assertions requiring causal analysis before non-trivial bugfix planning.
3. Add negative assertions that RCA does not require external evaluation infrastructure or learning loops.
4. Update only the bugfix skill/template content.

## Done When

- [ ] RCA template includes causal method, evidence, confidence, and counterfactual checks.
- [ ] Bugfix skill requires causal analysis before fix planning for non-trivial bugs.
- [ ] Guidance pushes toward process/standard/test gaps, not only proximate code symptoms.

## Risks

- **Overly heavy RCA for trivial bugs:** mitigate by applying formal causal analysis to non-trivial bugs.
- **False certainty:** mitigate by requiring confidence and counterfactual checks.

## Constitution Check

- **Article I:** Reuse existing bugfix template and skill.
- **Article II:** Snapshot tests precede content edits.
- **Article III:** Causal fields trace to `spec-wave2.md`.
- **Article IV:** No learning/evaluation system.
- **Article V:** 5-Whys/causal chain is simpler than a new RCA framework.
- **Article VI:** No abstractions or tool integrations.
