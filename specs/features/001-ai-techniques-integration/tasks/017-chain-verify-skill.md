# Task 017: `chain-verify` Skill + Report Contract [P]

**Phase:** W2.B — Verification + completion  
**User Story:** [US1] Verification + completion  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T016, T019, T020, T022, T023

---

## Objective

Create a bounded `chain-verify` skill and `ChainVerificationReport` schema for cross-artifact consistency checks without changing the feature chain yet.

## Spec References

- FR-W2-001 through FR-W2-003.
- AC-D4-001 through AC-D4-003.
- `spec-wave2.md` §5 Data / Report Contracts.

## Files to Change/Create

- Create `packages/ai-setup-go/library/skills/chain-verify.md`.
- Existing skill/frontmatter tests.
- Report-contract tests/fixtures under the appropriate existing package test location.

## Files NOT to Touch

- `packages/ai-setup-go/library/orchestration/chains/feature.json` — T018 owns optional chain insertion.
- Orchestrator runtime files.
- `packages/ai-setup-go/library/skills/plan.md` and `red-team-plan.md` except if tests only reference them as fixtures.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing frontmatter/metadata tests for `chain-verify`.
2. Add failing schema tests for `ChainVerificationReport` exactly as defined in `spec-wave2.md`.
3. Add failing fixture tests for pass/warn/fail cases: full trace, missing optional artifacts, missing requirement evidence.
4. Add failing test proving malformed/ambiguous artifact parsing yields `warn` findings, not crash or parser-driven fail.
5. Create the skill content only after tests fail.

## Done When

- [ ] `chain-verify` skill exists and validates.
- [ ] Report contract tests cover schema, traceability, findings, locations, and checked artifacts.
- [ ] Missing optional artifacts warn rather than halt.
- [ ] No chain integration or runtime code changes are made in this task.

## Risks

- **False hard failures:** mitigate by warning on missing/ambiguous optional inputs.
- **Validator framework creep:** keep this as one read-only skill/report, not a parser library.

## Constitution Check

- **Article I:** Reuse skill markdown and existing test harnesses.
- **Article II:** Schema/fixture tests come first.
- **Article III:** Report contract comes from `spec-wave2.md`.
- **Article IV:** No runtime verifier engine or Wave 3 retrieval.
- **Article V:** One skill is simpler than a general verification platform.
- **Article VI:** Avoid helper extraction and one-caller frameworks.
