# Compact Prompt

**Topic:** [Compaction Topic]
**Goal:** [Compaction Goal]

---

## Instructions

0. **Think Step-by-Step (CoT):** Privately reason step-by-step to decide what to keep/drop, but output only the final compressed handoff.
1. **Preserve Objective & Scope:** Capture the current objective, scope boundaries, and constraints.
2. **Preserve Decisions:** Capture key decisions made so far and rationale.
3. **Preserve Assumptions & Confidence:** Capture active assumptions/unknowns and current confidence level (high/medium/low).
4. **Preserve Progress:** Capture current progress and the next immediate action.
5. **Drop Noise:** Remove redundant narrative and stale exploration details.
6. **Produce a Structured Handoff:** Output a concise compaction report for seamless continuation.

## Few-Shot Mini Example (Generic)

Use this pattern as a guide:

```
Input (summary): Multiple threads discussing API auth, retries, and unrelated UI tweaks.
Output (shape):
- Objective/Scope: auth + retries only
- Decisions: token refresh interval = 10m
- Next Action: implement retry backoff test
- Dropped Context: unrelated UI brainstorming
```

## Common Mistakes to Avoid
- ❌ Losing critical file paths or decision rationale during compaction
- ❌ Summarizing too aggressively — preserve exact values and specifics
- ❌ Dropping open questions or unresolved items

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
