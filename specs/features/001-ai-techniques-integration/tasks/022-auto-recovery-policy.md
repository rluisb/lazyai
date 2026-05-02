# Task 022: Safe Auto-Recovery Policy Guidance [P]

**Phase:** W2.C — Feedback/recovery/state guidance  
**User Story:** [US3] Actionable feedback/recovery/state  
**Status:** TODO  
**Depends on:** T014  
**Parallel with:** T015, T016, T017, T019, T020, T023

---

## Objective

Add a static safe-recovery policy that distinguishes low-risk retries from human-gated recovery, while explicitly deferring autonomous runtime recovery.

## Spec References

- FR-W2-017 through FR-W2-019.
- AC-D5-001 through AC-D5-003.
- `plan-wave2.md` Decision 4.

## Files to Change/Create

- Create `packages/ai-setup-go/library/rules/auto-recovery.md`.
- `packages/ai-setup-go/library/skills/orchestrate.md`
- `packages/ai-setup-go/library/agents/orchestrator.md` if installed orchestrator persona guidance needs the policy.
- Existing rule/skill/agent snapshot tests.

## Files NOT to Touch

- `packages/orchestrator/src/chain-machine.ts`, `workflow-machine.ts`, retry logic, or recovery tools.
- Feature chain JSON.
- Any code that automatically edits files after failures.
- Existing W1 task files `001`–`013`.

## Test-First Order

1. Add failing snapshots proving the policy allowlist includes only low-risk actions from `spec-wave2.md`.
2. Add failing assertions that code edits, dependency changes, destructive commands, migrations, secrets/config, and ambiguous failures require human confirmation.
3. Add failing negative assertions that the policy does not claim runtime autonomous recovery support.
4. Update rule/skill/agent markdown only.

## Done When

- [ ] Auto-recovery policy exists and is validated by tests/snapshots.
- [ ] Orchestrator guidance uses the allowlist and human-gate list.
- [ ] Runtime autonomous recovery remains explicitly deferred.
- [ ] No orchestrator runtime behavior changes are made.

## Risks

- **Unsafe automatic actions:** mitigate by conservative allowlist and explicit human-confirm list.
- **Agents confuse policy with runtime feature:** state that Wave 2 policy is guidance only.

## Constitution Check

- **Article I:** Reuse rules/orchestrator guidance.
- **Article II:** Content tests precede edits.
- **Article III:** Allowlist comes from `spec-wave2.md`.
- **Article IV:** No runtime classifier or autonomous edit loop.
- **Article V:** Static policy is simpler and safer than automation.
- **Article VI:** No recovery framework abstraction.
