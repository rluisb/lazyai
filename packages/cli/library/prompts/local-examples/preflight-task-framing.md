# Worked Example — Pre-Flight Task Framing

## Scenario
Add retry logic to outbound webhook delivery without changing public API behavior.

## Pre-Flight Output (Example)

**Goal (1 sentence):** Add bounded webhook retries with exponential backoff while preserving existing request/response contracts.

**Files I expect to touch:**
- `src/webhooks/deliver.ts`
- `src/webhooks/retry-policy.ts`
- `src/__tests__/webhooks/deliver.test.ts`

**I will NOT touch:**
- `src/api/**`
- `specs/adrs/**`

**Assumptions:**
- Existing queue supports delayed jobs (**verified**)
- Retry budget should be max 3 attempts (**unverified**)

**Uncertainty:** medium — biggest unknown is provider-specific timeout behavior.
