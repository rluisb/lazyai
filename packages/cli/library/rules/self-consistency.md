# Self-Consistency Verification

## Rule
For high-stakes decisions (architecture, security, data model changes), use independent verification rounds to reduce single-path reasoning errors.

## Rationale
A single reasoning pass can miss edge cases or follow a locally-optimal but globally-suboptimal path. Independent re-checks from different starting points catch errors that self-review misses.

## Pattern

### Lightweight (1 round) — Simple changes
- Re-read requirements after completing the change
- Verify output matches input specification

### Standard (2 rounds) — Moderate changes
- Round 1: Complete the task normally
- Round 2: Re-derive the solution starting from requirements (not from Round 1 output)
- Compare: if rounds diverge, investigate before proceeding

### Thorough (3 rounds) — Complex/high-risk changes
- Round 1: Complete the task
- Round 2: Independent re-derivation from requirements
- Round 3: Cross-model verification (Reviewer agent on different model)
- Consensus required: 2 of 3 rounds must agree

## Cross-Model Verification
When available, use a different model for verification rounds. The Builder (Sonnet) produces; the Reviewer (Opus) verifies. This structural separation maximizes reasoning path diversity.

## Enforcement
- Reviewer agent applies standard (2-round) verification by default
- Architecture decisions require thorough (3-round) verification
- Trivial changes (typos, formatting) use lightweight (1-round)
