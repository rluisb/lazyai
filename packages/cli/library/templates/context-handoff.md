<template>
  <name>Context Handoff</name>
  <output>specs/memory/handoffs/YYYY-MM-DD-topic.md</output>
  <input>session state + decisions + open questions + evidence</input>
  <phase>Handoff — end of session, agent transfer, or phase boundary</phase>
</template>

# Context Handoff: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**From:** [agent id or human name]
**To:** [agent id or human name]
**Status:** Pending | Consumed | Expired

> **Purpose.** Transfer session context to another agent or a future session.
> Every claim MUST be backed by evidence — test output, command exit code,
> file path, or artifact reference. Unsupported claims are noise.

---

## 1. Current State

| Field | Value |
|---|---|
| Task | [one-line description] |
| Phase | research / plan / implement / verify / cleanup |
| Branch | [branch name] |
| Files changed | [paths, one per line] |
| Verification status | [what passes, what fails, what is untested] |

**Evidence:**
```
[command output, test result, or artifact reference supporting the above]
```

---

## 2. Decisions Made

| Decision | Rationale | Alternatives Rejected | Evidence |
|---|---|---|---|
| [decision] | [why this path] | [what was considered and rejected] | [link to test, diff, or artifact] |

---

## 3. Open Questions

| Question | Importance | Who Can Answer | Evidence Needed |
|---|---|---|---|
| [question] | high / medium / low | [name or role] | [what would resolve it] |

---

## 4. Next Steps

Ordered by dependency. Each step MUST have a clear exit criterion.

| Step | Action | Exit Criterion | Owner |
|---|---|---|---|
| 1 | [action] | [observable condition] | [name or agent] |
| 2 | [action] | [observable condition] | [name or agent] |

---

## 5. Trace Evidence

Minimal-form trace entries from the session. Use the [trace taxonomy](../templates/trace-taxonomy.md) categories and tags.

```
[category:tag1,tag2] summary — evidence: <path or ref>
[category:tag1,tag2] summary — evidence: <path or ref>
```

---

## 6. Constraints & Non-Goals

| Constraint | Source | Still Active? |
|---|---|---|
| [constraint] | [spec, ADR, human instruction] | yes / no / superseded by [X] |

---

## 7. Risks & Watchouts

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| [risk] | L / M / H | L / M / H | [mitigation] |

---

## 8. Required Evidence (before next agent acts)

The receiving agent MUST verify these before proceeding:

- [ ] [evidence item 1] — [how to verify]
- [ ] [evidence item 2] — [how to verify]

---

## 9. Handoff Closure

```
Status: PENDING / CONSUMED / EXPIRED
From: [agent id or human name]
To: [agent id or human name]
Date: YYYY-MM-DD
```
