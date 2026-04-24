# Worked Example — Commit Message Pattern

## Scenario
Feature task adds retry logic and tests for outbound webhooks.

## Message Drafting (Example)

**Why-focused summary (1–2 sentences):**
Retries reduce transient delivery failures and improve webhook reliability during short provider outages. The retry budget is intentionally capped to avoid queue amplification.

**Conventional commit example:**
`feat(webhooks): add bounded exponential retries for transient delivery failures`
