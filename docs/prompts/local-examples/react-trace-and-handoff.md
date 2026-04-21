# Worked Example — ReAct Trace + Session Handoff

## Scenario
Refactor payment retry flow across service and worker modules.

## ReAct Trace (Complex Task)

1. **Thought:** Need to confirm retry ownership boundary before edits.
2. **Action:** Review retry logic in `src/payments/service.ts` and worker consumer.
3. **Observation:** Both layers retry, creating duplicate attempts.
4. **Decision:** Centralize retries in worker; keep service single-attempt.

## Handoff Note (Example)

**Path:** `docs/memory/handoffs/2026-03-31-payment-retry-refactor.md`

```md
# Payment Retry Refactor — Handoff

- Status: in-progress
- Objective: remove duplicate retries while preserving API contract
- Decisions: worker owns retry policy (better reversibility/observability)
- Open questions: should max attempts be 3 or 5?
- Next actions:
  1. add failing tests for duplicate retry regression
  2. update worker retry policy and rerun quality gates
- Risks: queue backlog if retry delays are misconfigured
```
