# Task 024: Wave 2 Integration Validation + Scope Audit

**Phase:** W2.D — Integration validation  
**User Story:** All Wave 2 stories  
**Status:** TODO  
**Depends on:** T015-T023 approved/completed subset  
**Parallel with:** none

---

## Objective

Validate that completed Wave 2 tasks satisfy `spec-wave2.md`, remain within approved scope, and introduce no Wave 3/4 or unsupported runtime/chain constructs.

## Spec References

- All `spec-wave2.md` acceptance criteria.
- `tasks-wave2.md` Acceptance Trace.
- `plan-wave2.md` Runtime/Deferred Capability Ledger.

## Files to Change/Create

- Existing integration/snapshot tests that validate installed library content.
- Existing chain-shape tests if T018 was approved.
- Existing orchestrator runtime tests if T021 was approved.
- A Wave 2 acceptance trace note only if the repo convention uses evidence notes.

## Files NOT to Touch

- Any unapproved deferred runtime files for D5/D8.
- Wave 3/4 implementation files.
- MCP server configuration.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing acceptance-trace checks mapping each Wave 2 AC to at least one test/snapshot/evidence item.
2. Add failing scope-audit assertions that no RAG/model-routing/eval/learning/debate/MCP artifacts were introduced.
3. Add failing chain-shape assertions that no chain contains runtime conditionals, template markers, or parallel blocks.
4. If T021 was approved, add regression tests that gate feedback remains backward compatible.
5. Run package-specific test suites for changed surfaces and record commands/evidence.

## Done When

- [ ] Every Wave 2 AC has test/snapshot/evidence trace.
- [ ] Approved runtime/chain changes are explicitly identified; unapproved D5/D8 runtime changes are absent.
- [ ] No unsupported chain constructs or Wave 3/4 scope is present.
- [ ] Test evidence is recorded for all changed packages.
- [ ] Human decisions from T014 are reflected in the final scope audit.

## Risks

- **Partial Wave 2 appears complete:** mitigate with explicit AC trace.
- **Unapproved runtime creep:** mitigate with negative scope assertions.

## Constitution Check

- **Article I:** Reuse existing integration/snapshot tests.
- **Article II:** Scope/trace tests precede final glue changes.
- **Article III:** `spec-wave2.md` is the acceptance source.
- **Article IV:** Audit rejects Wave 3/4 and unapproved runtime work.
- **Article V:** A trace audit is simpler than new governance tooling.
- **Article VI:** No new frameworks; validation only.
