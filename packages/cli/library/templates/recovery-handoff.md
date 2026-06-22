<template>
  <name>Recovery Handoff</name>
  <output>specs/memory/handoffs/YYYY-MM-DD-topic-recovery.md</output>
  <input>failure context + attempted fixes + remaining state + evidence</input>
  <phase>Recovery — after a failed attempt, before retrying or escalating</phase>
</template>

# Recovery Handoff: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**From:** [agent id or human name]
**To:** [agent id or human name]
**Status:** Pending Recovery | In Progress | Resolved | Escalated

> **Purpose.** Transfer recovery context after a failure. The receiving agent
> MUST understand what went wrong, what was tried, and what state remains
> before attempting a fix. Every claim MUST be backed by evidence.

---

## 1. Failure Summary

| Field | Value |
|---|---|
| Task | [one-line description of what was being done] |
| Failure mode | [build error / test failure / runtime crash / unexpected behavior / timeout] |
| Root cause (known or suspected) | [what went wrong] |
| Confidence | high / medium / low — [evidence supporting the root-cause claim] |

**Evidence:**
```
[error output, stack trace, log excerpt, or artifact reference]
```

---

## 2. What Was Tried

| Attempt | Action | Result | Evidence |
|---|---|---|---|
| 1 | [what was done] | [success / partial / failed] | [command output or artifact ref] |
| 2 | [what was done] | [success / partial / failed] | [command output or artifact ref] |

---

## 3. Remaining State

| Aspect | Current State | Evidence |
|---|---|---|
| Working tree | [clean / dirty — list uncommitted changes] | `git status` output |
| Branch | [branch name] | |
| Files modified | [paths] | |
| Tests passing | [which tests pass] | [test output] |
| Tests failing | [which tests fail] | [test output] |
| Build status | [pass / fail] | [build output] |

---

## 4. Recovery Path

| Step | Action | Exit Criterion | Owner |
|---|---|---|---|
| 1 | [action] | [observable condition] | [name or agent] |
| 2 | [action] | [observable condition] | [name or agent] |

---

## 5. Risks of Continuing

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| [risk] | L / M / H | L / M / H | [mitigation] |

---

## 6. Required Evidence (before recovery attempt)

The receiving agent MUST verify these before acting:

- [ ] [evidence item 1] — [how to verify]
- [ ] [evidence item 2] — [how to verify]

---

## 7. Escalation Criteria

Escalate to a human if:

- [condition 1]
- [condition 2]

---

## 8. Recovery Closure

```
Status: PENDING_RECOVERY / IN_PROGRESS / RESOLVED / ESCALATED
From: [agent id or human name]
To: [agent id or human name]
Date: YYYY-MM-DD
Resolution: [if resolved, what fixed it]
```
