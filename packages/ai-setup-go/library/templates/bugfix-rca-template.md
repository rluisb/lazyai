<template>
  <name>Bugfix RCA — Root Cause Analysis</name>
  <output>specs/bugfixes/NNN-bug-name/techspec.md</output>
  <input>Bug report + Scout research + Codebase context</input>
  <phase>Plan — Step 2 of Bugfix flow (replaces heavy TechSpec for bugs)</phase>
</template>

# Bugfix RCA: [Bug Name]

**Bug:** NNN-bug-name
**Date:** YYYY-MM-DD
**Severity:** P0 (production down) | P1 (major degradation) | P2 (minor impact) | P3 (cosmetic)
**Status:** Draft | Approved | Implemented
**Ticket:** [link]
**Research:** [link to research.md]

---

## Reproduction Steps

<!-- Exact steps to reproduce. If you cannot reproduce it, stop — do not fix it.
     A bug you cannot reproduce is a bug you do not understand. -->

1. [Step 1]
2. [Step 2]
3. [Step 3]

**Expected:** [what should happen]
**Actual:** [what actually happens]
**Environment:** [local | staging | production — version/branch]

---

## Harness Contract — Expected vs Actual

The bug exists where reality (Actual) deviates from the contract (Expected). State both before reasoning about the fix — this is the verifier surface for review.

| Aspect | Expected (per spec / API / docs) | Actual (per reproduction) |
|---|---|---|
| Inputs | [...] | [...] |
| Outputs | [...] | [...] |
| Side effects | [...] | [...] |
| Error behavior | [...] | [...] |

**Contract source:** [link to spec.md / API doc / standard that defines Expected].

> If the Expected column is empty or guessed, **stop**: write/locate the contract first. Without an Expected, the "fix" is just a preference change.

---

## Root Cause

<!-- One clear sentence: What is the actual cause?
     Not "the code was wrong" — be specific.
     Example: "Function X assumes Y is never null, but path Z can produce null when W." -->

**Root cause:** [specific technical statement]

### Why It Happened

<!-- 5-why or short causal chain. How did we get here?
     Stop when you reach a process/standard gap, not just a code gap. -->

- Why 1: [immediate cause]
- Why 2: [deeper cause]
- Why 3: [root of the root — process/standard gap if applicable]

---

## Blast Radius

<!-- What else could be affected by this same bug or by the fix? -->

| Area | Affected? | How |
|------|-----------|-----|
| [module/feature] | Yes / No / Maybe | [description] |
| [module/feature] | Yes / No / Maybe | [description] |

**Confirmed affected paths:** [list — or "None beyond reported case"]

---

## Fix Scope

<!-- What exactly will be changed? Be precise.
     Rule: fix the bug, nothing else. No drive-by refactors.
     If you see something that should be refactored: flag it separately. -->

### Files to Change

| File | Change | Why |
|------|--------|-----|
| [path] | [what changes] | [why this fixes the root cause] |

### What Will NOT Change

<!-- Explicitly state what you are intentionally NOT touching. -->

- [file/module] — out of scope for this fix

---

## Regression Test

<!-- MANDATORY. Every bugfix needs a regression test.
     The test must fail on the current code and pass after the fix.
     No regression test = fix is not complete. -->

**Test location:** [path]
**Test description:** [what the test verifies — one sentence]
**Test type:** Unit | Integration | E2E

---

## Simplicity Gate

| Question | Answer |
|----------|--------|
| Is this the minimal change that fixes the root cause? | [YES / NO — justify if NO] |
| Does the fix stay within the affected module? | [YES / NO — justify if NO] |
| Does the fix introduce new dependencies? | [YES → justify / NO] |
| Is there an existing pattern to follow? | [YES → cite file / NO → document new pattern] |

---

## Conformance Check

- [ ] Fix addresses root cause (not just symptom)
- [ ] Blast radius reviewed — no collateral damage
- [ ] Regression test added
- [ ] No new dependencies without justification
- [ ] Fix scope is minimal — no drive-by changes

---

<!-- PRINCIPLES CHECK before approving:
- [ ] Bug was reproduced before fix was designed
- [ ] Root cause is specific (not "bad code")
- [ ] Fix addresses root cause, not just symptom
- [ ] Blast radius reviewed
- [ ] Regression test specified
- [ ] Simplicity gate passed
-->
