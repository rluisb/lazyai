---
name: evidence-verifier
description: Verify claims against source evidence. Classify each claim as Supported, Contradicted, or Inconclusive with specific citations.
trigger: /evidence-verify
phase: verify
techniques: [chain-of-thought]
output_schema:
  sections:
    - Claims Received
    - Evidence Trace (per claim)
    - Classification Summary
consumes:
  - claim or set of claims
  - source documents or code files
produces_for:
  - research (verified findings)
  - review (evidence-backed assessment)
---

# Evidence Verifier Skill

## Purpose

Evaluate claims against available source evidence and classify each claim with a verdict backed by specific citations.

## Instructions

1. Receive a claim and one or more source documents or code files.
2. For each claim, trace it to specific evidence in the source material.
3. Classify each claim as:
   - **Supported**: The source material directly confirms the claim.
   - **Contradicted**: The source material directly contradicts the claim.
   - **Inconclusive**: The source material does not contain enough information to confirm or deny the claim.
4. Cite the specific file, line, or passage that supports the classification.
5. If a claim is ambiguous, state what additional evidence would resolve it.

## Output Format

For each claim:

```
Claim: [verbatim claim text]
Verdict: Supported | Contradicted | Inconclusive
Evidence: [file:line or passage quote]
Notes: [optional — conflict flag, ambiguity note, or additional evidence needed]
```

## Constraints

- Never fabricate evidence. If the source does not contain the information, classify as inconclusive.
- Never infer intent beyond what is explicitly stated.
- Prefer direct quotes over paraphrase when citing evidence.
- If multiple sources conflict, flag the conflict and classify as inconclusive unless one source is authoritative.
