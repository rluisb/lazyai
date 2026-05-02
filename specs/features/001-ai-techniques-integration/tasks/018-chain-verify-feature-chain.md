# Task 018: Optional Sequential `chain-verify` Feature-Chain Integration

**Phase:** W2.B — Verification + completion  
**User Story:** [US1] Verification + completion  
**Status:** TODO — requires human approval after T017  
**Depends on:** T017 and explicit approval to change the feature chain  
**Parallel with:** none

---

## Objective

If approved, insert `chain-verify` as a simple sequential feature-chain step after implementation review and before documentation/completion, preserving all T8/W1.B runtime constraints.

## Spec References

- FR-W2-004.
- AC-D4-004.
- `plan-wave2.md` Decision 2 and Human Decisions Needed.
- T8 evidence constraints.

## Files to Change/Create

- `packages/ai-setup-go/library/orchestration/chains/feature.json`
- Existing chain-shape/scaffold install tests in Go/TS.
- Orchestrator chain-machine tests only if needed to validate existing sequential transitions.

## Files NOT to Touch

- Runtime conditional semantics, template chain rendering, parallel blocks, generic workflow engine files.
- `packages/orchestrator/src/chain-machine.ts` unless an existing sequential transition bug is exposed and approved.
- D9/D5/D8 runtime files.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Confirm the human approval decision is recorded by T014.
2. Add failing chain-shape tests expecting a sequential `chain-verify` step in the approved location.
3. Add failing assertions that no chain contains `{{#if`, `condition`, `optionalByFeature`, or parallel blocks.
4. Add failing transition tests proving `review pass → chain-verify → document` and `chain-verify fail/warn` behavior reaches the intended human/review path without automatic unsupported loops.
5. Add scaffold/install tests proving installed `feature.json` matches the source chain.
6. Edit only the chain JSON and minimal test fixtures.

## Done When

- [ ] Approval for default chain integration is documented.
- [ ] Feature chain remains a sequential `steps` array.
- [ ] `chain-verify` runs after implementation review and before document/done.
- [ ] No runtime conditionals, templates, or parallel blocks are introduced.
- [ ] Installed chain output is covered by tests.

## Risks

- **Active-run incompatibility:** mitigate by release notes/rollback policy and sequencing after W1 bake check.
- **Automatic fail loop surprises:** mitigate by explicit transition tests and human-approved behavior.

## Constitution Check

- **Article I:** Reuse existing chain JSON and scaffold tests.
- **Article II:** Chain-shape and transition tests precede edits.
- **Article III:** Integration is allowed only by `spec-wave2.md`/approval.
- **Article IV:** No unsupported runtime semantics.
- **Article V:** One sequential step is the simplest enforceable integration.
- **Article VI:** No engine rewrite or generic chain abstraction.
