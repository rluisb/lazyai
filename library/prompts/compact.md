# Compact Prompt

**Topic:** [Compaction Topic]
**Goal:** [Compaction Goal]

---

## Instructions

1. **Preserve Objective & Scope:** Capture the current objective, scope boundaries, and constraints.
2. **Preserve Decisions:** Capture key decisions made so far and rationale.
3. **Preserve Assumptions & Confidence:** Capture active assumptions/unknowns and current confidence level (high/medium/low).
4. **Preserve Progress:** Capture current progress and the next immediate action.
5. **Drop Noise:** Remove redundant narrative and stale exploration details.
6. **Produce a Structured Handoff:** Output a concise compaction report for seamless continuation.

## Output Format

```
## Compaction: [Topic]

### Redundancy
- [file:line] — [description] — [proposal]

### Objective, Scope, Constraints
- Objective: [current objective]
- In Scope: [explicit in-scope items]
- Out of Scope: [explicit exclusions]
- Constraints: [key constraints]

### Decisions & Rationale
- [decision] — [rationale]

### Assumptions, Unknowns, Confidence
- Assumptions: [verified/unverified items]
- Unknowns: [open questions]
- Confidence: [high|medium|low] — [why]

### Progress & Next Action
- Completed: [what is done]
- Next Action: [immediate next step]

### Dropped Context
- [redundant/stale detail removed]
```
