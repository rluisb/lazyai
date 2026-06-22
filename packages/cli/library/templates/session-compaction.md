<template>
  <name>Session Compaction</name>
  <output>inline — used within the same session or as a compacted working summary</output>
  <input>full session context before compaction</input>
  <phase>Any — triggered by compact-after rules in the context-compaction fragment</phase>
</template>

# Session Compaction: [YYYY-MM-DD] — [Short Topic]

**Date:** YYYY-MM-DD
**Agent:** [agent id]
**Phase:** [current phase]
**Trigger:** [compact-after reason — phase boundary / handoff / drift / window full / retry / human gate]

> **Purpose.** Reduce session context to essential signal while preserving
> every never-compact-away item. This is a working summary for the same agent
> mid-session, not a handoff to another agent.

---

## 1. Task

```
[task description — verbatim from the original assignment]
```

**Acceptance criteria:**
- [ ] [criterion 1]
- [ ] [criterion 2]

---

## 2. Active Constraints

| Constraint | Source | Applies To |
|---|---|---|
| [constraint] | [spec, ADR, human instruction] | [current phase / all phases] |

---

## 3. Decisions Made

| Decision | Rationale | Evidence |
|---|---|---|
| [decision] | [why] | [test output, file path, artifact ref] |

---

## 4. Open Questions

- [question] — [importance: high/med/low]

---

## 5. Verification Status

| Check | Result | Evidence |
|---|---|---|
| [test / lint / build] | PASS / FAIL / SKIPPED | [command output or artifact ref] |

---

## 6. Files Touched

- [path] — [created / modified / deleted] — [purpose]

---

## 7. Next Immediate Action

[one sentence — what to do next]

---

## 8. Dropped (compacted away)

- [item dropped] — [why it is safe to drop]
- [item dropped] — [why it is safe to drop]

---

## Compaction Rules Applied

- [ ] Compact-after trigger identified and recorded.
- [ ] All never-compact-away items preserved.
- [ ] Execution trace dropped (tool calls, intermediate outputs, dead ends).
- [ ] Verbose tool output replaced with summaries or excerpts.
- [ ] Resolved questions dropped.
- [ ] Evidence preserved for every claim in sections 1–7.
