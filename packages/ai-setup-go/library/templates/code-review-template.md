<template>
  <name>Code Review — External PR Review</name>
  <output>specs/code-reviews/NNN-pr-name/review.md</output>
  <input>PR diff + original ticket + relevant standards + relevant ADRs</input>
  <phase>Review — used for teammate PR reviews, not self-review</phase>
</template>

# Code Review: [PR Name / Ticket]

**PR:** [link or #number]
**Author:** [teammate name]
**Ticket:** [JIRA/ticket link — or "N/A"]
**Reviewer:** [your name]
**Date:** YYYY-MM-DD
**Status:** Draft | Submitted | Resolved

---

## Context Loaded

<!-- Document what you read before reviewing.
     This ensures reproducibility — if a finding is questioned, the context is here.
     NOTE: In external reviews there is no spec file. Intent must be inferred from
     PR description + ticket. If neither exists, ask the author before proceeding. -->

- **PR diff:** [link — or "pasted below"]
- **Original ticket:** [link — or "N/A, inferred from PR description"]
- **Related ADRs:** [list — or "None"]
- **Relevant standards:** [specs/standards/... — or "None"]
- **Prior art:** [related PRs or features — or "None"]

---

## Intent Summary

<!-- 2–3 sentences: What is this PR trying to do?
     Infer from PR description + ticket. This is your north star for all findings.
     If you cannot state the intent clearly — stop and ask the author before proceeding.
     Do NOT assume intent from code alone. -->

[INTENT_SUMMARY]

---

## Findings

<!-- Severity guide:
     🔴 Critical — blocks merge (correctness, security, data integrity, crash)
     🟡 Major    — should fix before merge (logic gaps, missing tests, perf regression)
     🟢 Minor    — nice to have (style, naming, small improvements, cosmetic)

     Rules:
     - Every Critical/Major finding MUST include file + line reference
     - Every finding MUST include a concrete suggestion (not just "fix this")
     - Limit to findings within the PR scope — no drive-by feedback on unrelated code -->

### 🔴 Critical (must fix before merge)

- [ ] `[file:line]` — [issue description] — **Suggestion:** [concrete fix]

### 🟡 Major (should fix before merge)

- [ ] `[file:line]` — [issue description] — **Suggestion:** [concrete fix]

### 🟢 Minor (nice to have)

- [ ] `[file:line]` — [issue description] — **Suggestion:** [concrete fix]

---

## Coverage Check

<!-- Verify test coverage for the changed code. Check each item. -->

- [ ] New logic has unit tests
- [ ] Edge cases are tested
- [ ] Error paths are tested
- [ ] Integration tests updated (if applicable)
- [ ] No tests deleted to make the suite pass

---

## Standards Compliance

- [ ] Code style follows `specs/rules/code-style.md`
- [ ] Error handling follows project patterns
- [ ] No security violations (auth, input validation, secrets exposure)
- [ ] No new dependencies without justification
- [ ] No scope creep beyond ticket intent

---

## Verdict

```
APPROVE / REQUEST_CHANGES / COMMENT

Blocking issues: [N critical, N major — or "None"]
Summary: [1–2 sentences on overall quality and confidence]
```

---

<!-- PRINCIPLES CHECK before submitting:
- [ ] Intent was stated before findings (not inferred from code alone)
- [ ] Every Critical/Major finding references file + line number
- [ ] Every finding includes a concrete suggestion (not just "fix this")
- [ ] Coverage check completed
- [ ] Verdict is explicit — APPROVE / REQUEST_CHANGES / COMMENT
- [ ] No feedback on code outside the PR scope
-->
